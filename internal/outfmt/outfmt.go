package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Mode represents the output mode.
type Mode struct {
	JSON  bool
	JSONL bool
	Plain bool
}

// ParseError is returned for invalid flag combinations.
type ParseError struct {
	Msg string
}

func (e *ParseError) Error() string { return e.Msg }

// FromFlags creates a Mode from flag values.
// Returns error if more than one output mode is set.
func FromFlags(jsonOut bool, jsonlOut bool, plainOut bool) (Mode, error) {
	count := 0
	if jsonOut {
		count++
	}
	if jsonlOut {
		count++
	}
	if plainOut {
		count++
	}
	if count > 1 {
		return Mode{}, &ParseError{Msg: "cannot combine output modes; use only one of --json, --jsonl, or --plain"}
	}

	return Mode{JSON: jsonOut, JSONL: jsonlOut, Plain: plainOut}, nil
}

type ctxKey struct{}
type requestIDCtxKey struct{}

// WithMode attaches the output mode to a context.
func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, ctxKey{}, mode)
}

// FromContext retrieves the output mode from context.
func FromContext(ctx context.Context) Mode {
	if v := ctx.Value(ctxKey{}); v != nil {
		if m, ok := v.(Mode); ok {
			return m
		}
	}

	return Mode{}
}

// IsJSON returns true if JSON output mode is enabled.
func IsJSON(ctx context.Context) bool {
	m := FromContext(ctx)
	return m.JSON || m.JSONL
}

// IsJSONL returns true if JSONL output mode is enabled.
func IsJSONL(ctx context.Context) bool {
	return FromContext(ctx).JSONL
}

// IsPlain returns true if plain output mode is enabled.
func IsPlain(ctx context.Context) bool {
	return FromContext(ctx).Plain
}

// IsHuman returns true if human-readable output (default) is enabled.
func IsHuman(ctx context.Context) bool {
	m := FromContext(ctx)
	return !m.JSON && !m.JSONL && !m.Plain
}

// WithRequestID attaches an optional request ID to context for output metadata.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDCtxKey{}, requestID)
}

// RequestIDFromContext retrieves the optional request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDCtxKey{}); v != nil {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}

// WriteJSON writes v as indented JSON to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// WriteJSONLine writes v as single-line JSON to w.
func WriteJSONLine(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json line: %w", err)
	}

	return nil
}
