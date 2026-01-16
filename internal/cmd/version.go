package cmd

import (
	"context"
	"os"

	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// VersionCmd shows version information.
type VersionCmd struct{}

// Run executes the version command.
func (c *VersionCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"version": Version,
		})
	}

	u.Out().Printf("rr version %s", Version)
	return nil
}
