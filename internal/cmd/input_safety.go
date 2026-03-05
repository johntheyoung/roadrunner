package cmd

import (
	"strings"
	"unicode/utf8"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func validateResourceID(value, label string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	if err := rejectControlChars(trimmed, label); err != nil {
		return err
	}
	if strings.ContainsAny(trimmed, "?#%") {
		return errfmt.UsageError("invalid %s %q (must not contain ?, #, or %%)", label, trimmed)
	}
	if !utf8.ValidString(trimmed) {
		return errfmt.UsageError("invalid %s (must be valid UTF-8)", label)
	}

	return nil
}

func rejectControlChars(value, label string) error {
	for _, r := range value {
		if (r >= 0x00 && r < 0x20) || r == 0x7f {
			return errfmt.UsageError("invalid %s %q (contains control characters)", label, value)
		}
	}
	return nil
}
