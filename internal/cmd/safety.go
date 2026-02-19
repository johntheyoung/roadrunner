package cmd

import (
	"sort"
	"strings"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

// dataWriteCommands are commands that modify data and should be blocked in --readonly mode.
var dataWriteCommands = map[string]bool{
	"messages send":        true,
	"messages send-file":   true,
	"messages edit":        true,
	"messages react":       true,
	"messages unreact":     true,
	"chats create":         true,
	"chats start":          true,
	"chats archive":        true,
	"reminders set":        true,
	"reminders clear":      true,
	"assets upload":        true,
	"assets upload-base64": true,
	"accounts alias set":   true,
	"accounts alias unset": true,
}

// exemptCommands are commands that bypass --readonly restrictions (local-only operations).
var exemptCommands = map[string]bool{
	"auth set":   true,
	"auth clear": true,
	"focus":      true,
}

// DataWriteCommandsList returns a sorted list of data write commands.
func DataWriteCommandsList() []string {
	list := make([]string, 0, len(dataWriteCommands))
	for cmd := range dataWriteCommands {
		list = append(list, cmd)
	}
	sort.Strings(list)
	return list
}

// ExemptCommandsList returns a sorted list of exempt commands.
func ExemptCommandsList() []string {
	list := make([]string, 0, len(exemptCommands))
	for cmd := range exemptCommands {
		list = append(list, cmd)
	}
	sort.Strings(list)
	return list
}

// checkEnableCommands validates that the command is in the allowlist.
// If EnableCommands is empty, all commands are allowed.
// Allowlist entries can be top-level commands (e.g., "messages") which allow all subcommands,
// or full paths (e.g., "messages send") for specific commands.
func checkEnableCommands(flags *RootFlags, command string) error {
	if len(flags.EnableCommands) == 0 {
		return nil
	}

	// Check if command or its parent is in the allowlist
	for _, allowed := range flags.EnableCommands {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		// Exact match
		if command == allowed {
			return nil
		}

		// Parent match: "messages" allows "messages send", "messages list", etc.
		if strings.HasPrefix(command, allowed+" ") {
			return nil
		}

		// Check if allowed is a prefix of the top-level command
		// e.g., allowed="messages" matches command="messages send"
		topLevel := strings.Split(command, " ")[0]
		if topLevel == allowed {
			return nil
		}
	}

	return errfmt.NewEnableCommandsError(command, flags.EnableCommands)
}

// checkReadonly blocks data write operations when --readonly is set.
// Exemptions: auth set/clear and focus are always allowed.
func checkReadonly(flags *RootFlags, command string) error {
	if !flags.Readonly {
		return nil
	}

	// Exempt local-only commands
	if exemptCommands[command] {
		return nil
	}

	// Block data write commands
	if dataWriteCommands[command] {
		return errfmt.NewReadonlyError(command)
	}

	return nil
}

// normalizeCommand extracts a clean command path from Kong's command string.
// Kong returns paths like "messages send <chatID> <text>" and we want "messages send".
func normalizeCommand(kongCmd string) string {
	// Kong command format: "command subcommand <arg> <arg>"
	// We want just the command words, not the arguments

	parts := strings.Fields(kongCmd)
	var cmdParts []string

	for _, part := range parts {
		// Stop when we hit an argument (starts with < or --)
		if strings.HasPrefix(part, "<") || strings.HasPrefix(part, "--") {
			break
		}
		cmdParts = append(cmdParts, part)
	}

	return strings.Join(cmdParts, " ")
}
