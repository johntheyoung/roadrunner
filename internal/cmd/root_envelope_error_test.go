package cmd

import (
	"encoding/json"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestExecute_EnvelopeErrorIncludesHint_EnableCommands(t *testing.T) {
	withArgs(t, []string{"rr", "--json", "--envelope", "--request-id=req-err-1", "--enable-commands=chats", "messages", "list", "!room:beeper.local"}, func() {
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
		meta, _ := env["metadata"].(map[string]any)
		if meta == nil {
			t.Fatalf("expected metadata in envelope")
		}
		if requestID, _ := meta["request_id"].(string); requestID != "req-err-1" {
			t.Fatalf("request_id = %q, want %q", requestID, "req-err-1")
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

func TestExecute_EnvelopeSuccessIncludesRequestID(t *testing.T) {
	withArgs(t, []string{"rr", "--json", "--envelope", "--request-id=req-ok-1", "version"}, func() {
		var code int
		out, _ := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitSuccess {
			t.Fatalf("exit code = %d, want %d", code, errfmt.ExitSuccess)
		}

		var env map[string]any
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
		}
		if success, _ := env["success"].(bool); !success {
			t.Fatalf("expected success=true, got false")
		}
		meta, _ := env["metadata"].(map[string]any)
		if meta == nil {
			t.Fatalf("expected metadata in envelope")
		}
		if requestID, _ := meta["request_id"].(string); requestID != "req-ok-1" {
			t.Fatalf("request_id = %q, want %q", requestID, "req-ok-1")
		}
	})
}
