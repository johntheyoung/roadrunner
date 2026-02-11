package cmd

import (
	"encoding/json"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestExecute_EnvelopeErrorIncludesHint_EnableCommands(t *testing.T) {
	withArgs(t, []string{"rr", "--json", "--envelope", "--enable-commands=chats", "messages", "list", "!room:beeper.local"}, func() {
		var code int
		out, _ := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitUsageError {
			t.Fatalf("exit code = %d, want %d", code, errfmt.ExitUsageError)
		}

		var env map[string]any
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
		}
		if success, _ := env["success"].(bool); success {
			t.Fatalf("expected success=false, got true")
		}
		errObj, _ := env["error"].(map[string]any)
		if errObj == nil {
			t.Fatalf("expected error object in envelope")
		}
		if hint, _ := errObj["hint"].(string); hint == "" {
			t.Fatalf("expected non-empty error.hint, got: %#v", errObj["hint"])
		}
	})
}

func TestExecute_EnvelopeErrorIncludesHint_AgentMissingAllowlist(t *testing.T) {
	withArgs(t, []string{"rr", "--agent", "version"}, func() {
		var code int
		out, _ := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitUsageError {
			t.Fatalf("exit code = %d, want %d", code, errfmt.ExitUsageError)
		}

		var env map[string]any
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
		}
		errObj, _ := env["error"].(map[string]any)
		if errObj == nil {
			t.Fatalf("expected error object in envelope")
		}
		if hint, _ := errObj["hint"].(string); hint == "" {
			t.Fatalf("expected non-empty error.hint, got: %#v", errObj["hint"])
		}
	})
}
