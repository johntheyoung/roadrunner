package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// VersionCmd shows version information.
type VersionCmd struct {
	Fields []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

// VersionString returns a human-readable version string.
func VersionString() string {
	v := strings.TrimSpace(Version)
	if v == "" {
		v = "dev"
	}

	commit := strings.TrimSpace(Commit)
	date := strings.TrimSpace(Date)
	if commit == "" && date == "" {
		return v
	}
	if commit == "" {
		return fmt.Sprintf("%s (%s)", v, date)
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", v, commit)
	}
	return fmt.Sprintf("%s (%s %s)", v, commit, date)
}

// Run executes the version command.
func (c *VersionCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"version":  strings.TrimSpace(Version),
			"commit":   strings.TrimSpace(Commit),
			"date":     strings.TrimSpace(Date),
			"features": []string{"enable-commands", "readonly", "envelope", "agent-mode"},
		}, "version")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"version", "commit", "date"})
		if err != nil {
			return err
		}
		writePlainFields(u, fields, map[string]string{
			"version": strings.TrimSpace(Version),
			"commit":  strings.TrimSpace(Commit),
			"date":    strings.TrimSpace(Date),
		})
		return nil
	}

	u.Out().Println(VersionString())
	return nil
}
