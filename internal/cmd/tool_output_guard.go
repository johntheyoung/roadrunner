package cmd

import (
	"strings"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

// guardAgainstPastedToolOutput blocks obvious cases where an agent (or human)
// accidentally pastes rr JSON output into a message. This is a defense-in-depth
// safety check to reduce privacy leaks in automated reply flows.
func guardAgainstPastedToolOutput(text string, allow bool) error {
	if allow {
		return nil
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	// Common rr list/search outputs.
	if strings.Contains(trimmed, "\"items\"") &&
		strings.Contains(trimmed, "\"has_more\"") &&
		strings.Contains(trimmed, "\"oldest_cursor\"") &&
		strings.Contains(trimmed, "\"newest_cursor\"") {
		return errfmt.UsageError("refusing to send message text that looks like rr list output (possible privacy leak); remove pasted command output or pass --allow-tool-output")
	}

	// Common rr envelope output shape.
	if strings.Contains(trimmed, "\"success\"") &&
		strings.Contains(trimmed, "\"data\"") &&
		strings.Contains(trimmed, "\"error\"") &&
		strings.Contains(trimmed, "\"metadata\"") {
		return errfmt.UsageError("refusing to send message text that looks like rr envelope output (possible privacy leak); remove pasted command output or pass --allow-tool-output")
	}

	// rr doctor output (JSON).
	if strings.Contains(trimmed, "\"config_path\"") &&
		strings.Contains(trimmed, "\"token_source\"") &&
		strings.Contains(trimmed, "\"api_url\"") {
		return errfmt.UsageError("refusing to send message text that looks like rr doctor output (possible privacy leak); remove pasted command output or pass --allow-tool-output")
	}

	return nil
}
