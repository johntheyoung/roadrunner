package cmd

import (
	"context"
	"os"

	"github.com/johntheyoung/roadrunner/internal/outfmt"
)

// writeJSON writes data as JSON, optionally wrapped in an envelope.
// The command parameter is used for envelope metadata.
func writeJSON(ctx context.Context, data any, command string) error {
	if outfmt.IsEnvelope(ctx) {
		return outfmt.WriteEnvelope(os.Stdout, data, Version, command)
	}
	return outfmt.WriteJSON(os.Stdout, data)
}
