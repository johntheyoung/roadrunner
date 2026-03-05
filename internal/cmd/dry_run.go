package cmd

import (
	"context"
	"encoding/json"

	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

func handleDryRunWrite(ctx context.Context, flags *RootFlags, command string, plan any) (bool, error) {
	if flags == nil || !flags.DryRun {
		return false, nil
	}

	if outfmt.IsJSON(ctx) {
		return true, writeJSON(ctx, map[string]any{
			"dry_run": true,
			"command": command,
			"plan":    plan,
		}, command)
	}

	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		b, _ := json.Marshal(plan)
		u.Out().Printf("%s\t%s", command, string(b))
		return true, nil
	}

	u.Out().Warn("Dry run: no API request sent")
	u.Out().Printf("Command: %s", command)
	if b, err := json.MarshalIndent(plan, "", "  "); err == nil {
		u.Out().Printf("Plan:\n%s", string(b))
	}

	return true, nil
}
