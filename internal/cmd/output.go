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
		return outfmt.WriteEnvelopeWithMetadata(os.Stdout, data, Version, command, nil, outfmt.RequestIDFromContext(ctx))
	}
	return outfmt.WriteJSON(os.Stdout, data)
}

// writeJSONWithPagination writes data as JSON with optional normalized pagination
// metadata when envelope mode is enabled.
func writeJSONWithPagination(ctx context.Context, data any, command string, pagination *outfmt.EnvelopePagination) error {
	if outfmt.IsEnvelope(ctx) {
		return outfmt.WriteEnvelopeWithMetadata(os.Stdout, data, Version, command, pagination, outfmt.RequestIDFromContext(ctx))
	}
	return outfmt.WriteJSON(os.Stdout, data)
}
