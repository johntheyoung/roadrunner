package cmd

import (
	"context"
	"sort"

	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// CapabilitiesCmd shows CLI capabilities for agent discovery.
type CapabilitiesCmd struct{}

// CapabilitiesResponse is the JSON output structure.
type CapabilitiesResponse struct {
	Version     string            `json:"version"`
	Features    []string          `json:"features"`
	Defaults    CapDefaults       `json:"defaults"`
	OutputModes []string          `json:"output_modes"`
	Safety      CapSafety         `json:"safety"`
	Commands    CapCommands       `json:"commands"`
	Flags       map[string]string `json:"flags"`
}

// CapDefaults shows default values for key settings.
type CapDefaults struct {
	Timeout int    `json:"timeout"`
	BaseURL string `json:"base_url"`
}

// CapSafety describes the safety-related flags.
type CapSafety struct {
	EnableCommandsDesc string `json:"enable_commands_desc"`
	ReadonlyDesc       string `json:"readonly_desc"`
	AgentDesc          string `json:"agent_desc"`
}

// CapCommands categorizes commands by type.
type CapCommands struct {
	Read   []string `json:"read"`
	Write  []string `json:"write"`
	Exempt []string `json:"exempt"`
}

// readCommands returns a list of read-only commands.
func readCommands() []string {
	return []string{
		"accounts list",
		"accounts alias list",
		"assets download",
		"assets serve",
		"auth status",
		"chats get",
		"chats list",
		"chats resolve",
		"chats search",
		"contacts resolve",
		"contacts search",
		"doctor",
		"messages context",
		"messages list",
		"messages search",
		"messages tail",
		"messages wait",
		"search",
		"status",
		"unread",
		"version",
		"capabilities",
	}
}

// Run executes the capabilities command.
func (c *CapabilitiesCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	writeList := DataWriteCommandsList()
	sort.Strings(writeList)

	exemptList := ExemptCommandsList()
	sort.Strings(exemptList)

	readList := readCommands()
	sort.Strings(readList)

	resp := CapabilitiesResponse{
		Version:  Version,
		Features: []string{"enable-commands", "readonly", "envelope", "agent-mode"},
		Defaults: CapDefaults{
			Timeout: flags.Timeout,
			BaseURL: flags.BaseURL,
		},
		OutputModes: []string{"human", "json", "plain"},
		Safety: CapSafety{
			EnableCommandsDesc: "Comma-separated allowlist of top-level commands",
			ReadonlyDesc:       "Block data write operations",
			AgentDesc:          "Agent profile: forces JSON, envelope, no-input, readonly; requires --enable-commands",
		},
		Commands: CapCommands{
			Read:   readList,
			Write:  writeList,
			Exempt: exemptList,
		},
		Flags: map[string]string{
			"--json":            "Output JSON to stdout",
			"--plain":           "Output stable TSV to stdout",
			"--envelope":        "Wrap JSON in {success,data,error,metadata}",
			"--no-input":        "Never prompt; fail instead",
			"--readonly":        "Block data write operations",
			"--enable-commands": "Allowlist of commands",
			"--agent":           "Agent profile mode",
			"--account":         "Default account ID for commands",
			"--timeout":         "API timeout in seconds",
			"--force":           "Skip confirmations",
		},
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "capabilities")
	}

	// Human-readable output
	u.Out().Printf("CLI Capabilities (v%s)", Version)
	u.Out().Println("")

	u.Out().Printf("Features:")
	for _, f := range resp.Features {
		u.Out().Printf("  - %s", f)
	}
	u.Out().Println("")

	u.Out().Printf("Output modes: %v", resp.OutputModes)
	u.Out().Println("")

	u.Out().Printf("Safety flags:")
	u.Out().Printf("  --enable-commands: %s", resp.Safety.EnableCommandsDesc)
	u.Out().Printf("  --readonly:        %s", resp.Safety.ReadonlyDesc)
	u.Out().Printf("  --agent:           %s", resp.Safety.AgentDesc)
	u.Out().Println("")

	u.Out().Printf("Commands:")
	u.Out().Printf("  Read-only (%d): %v", len(readList), readList[:5])
	u.Out().Dim("    ... use --json for full list")
	u.Out().Printf("  Write (%d): %v", len(writeList), writeList)
	u.Out().Printf("  Exempt (%d): %v", len(exemptList), exemptList)

	return nil
}
