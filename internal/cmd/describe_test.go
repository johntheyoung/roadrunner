package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestDescribeCommand_JSON(t *testing.T) {
	withArgs(t, []string{"rr", "describe", "messages", "send", "--json"}, func() {
		var code int
		out, errText := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitSuccess {
			t.Fatalf("exit code = %d, want %d\nstderr: %s", code, errfmt.ExitSuccess, errText)
		}
		if strings.TrimSpace(errText) != "" {
			t.Fatalf("stderr not empty: %q", errText)
		}

		var resp struct {
			Command string `json:"command"`
			Safety  struct {
				ReadonlyBlocked bool   `json:"readonly_blocked"`
				RetryClass      string `json:"retry_class"`
			} `json:"safety"`
			Positionals []struct {
				Name string `json:"name"`
			} `json:"positionals"`
			Flags []struct {
				Name string `json:"name"`
			} `json:"flags"`
		}
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
		}

		if resp.Command != "messages send" {
			t.Fatalf("command = %q, want %q", resp.Command, "messages send")
		}
		if !resp.Safety.ReadonlyBlocked {
			t.Fatal("readonly_blocked = false, want true")
		}
		if resp.Safety.RetryClass != "non-idempotent" {
			t.Fatalf("retry_class = %q, want %q", resp.Safety.RetryClass, "non-idempotent")
		}

		if len(resp.Positionals) == 0 {
			t.Fatal("positionals empty")
		}

		hasChatFlag := false
		for _, f := range resp.Flags {
			if f.Name == "chat" {
				hasChatFlag = true
				break
			}
		}
		if !hasChatFlag {
			t.Fatal("flags missing expected 'chat'")
		}
	})
}

func TestDescribeCommand_JSONEnvelope(t *testing.T) {
	withArgs(t, []string{"rr", "--json", "--envelope", "describe", "messages", "send"}, func() {
		var code int
		out, errText := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitSuccess {
			t.Fatalf("exit code = %d, want %d\nstderr: %s", code, errfmt.ExitSuccess, errText)
		}
		if strings.TrimSpace(errText) != "" {
			t.Fatalf("stderr not empty: %q", errText)
		}
		if !strings.Contains(out, `"success": true`) {
			t.Fatalf("expected success envelope, got: %s", out)
		}
		if !strings.Contains(out, `"command": "describe messages send"`) {
			t.Fatalf("expected command metadata in envelope, got: %s", out)
		}
	})
}

func TestDescribeCommand_Unknown(t *testing.T) {
	withArgs(t, []string{"rr", "describe", "does", "not", "exist", "--json"}, func() {
		var code int
		_, errText := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitUsageError {
			t.Fatalf("exit code = %d, want %d", code, errfmt.ExitUsageError)
		}
		if !strings.Contains(errText, "unknown command path") {
			t.Fatalf("expected unknown command path error, got: %s", errText)
		}
	})
}
