package cmd

import (
	"strings"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestExecute_UnknownFlag(t *testing.T) {
	withArgs(t, []string{"rr", "--definitely-nope"}, func() {
		var code int
		_, errText := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitUsageError {
			t.Fatalf("exit code = %d, want %d", code, errfmt.ExitUsageError)
		}
		if !strings.Contains(errText, "unknown flag") {
			t.Fatalf("expected unknown flag error, got %q", errText)
		}
	})
}

func TestExecute_JSONPlainConflict(t *testing.T) {
	withArgs(t, []string{"rr", "--json", "--plain", "version"}, func() {
		var code int
		_, errText := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitUsageError {
			t.Fatalf("exit code = %d, want %d", code, errfmt.ExitUsageError)
		}
		if !strings.Contains(errText, "cannot use both --json and --plain") {
			t.Fatalf("expected flag conflict error, got %q", errText)
		}
	})
}
