package outfmt

import (
	"context"
	"io"
	"time"
)

// Envelope wraps JSON output in a standardized structure.
type Envelope struct {
	Success  bool           `json:"success"`
	Data     any            `json:"data,omitempty"`
	Error    *EnvelopeError `json:"error,omitempty"`
	Metadata *EnvelopeMeta  `json:"metadata,omitempty"`
}

// EnvelopeError contains error details.
type EnvelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// EnvelopeMeta contains metadata about the response.
type EnvelopeMeta struct {
	Timestamp  string              `json:"timestamp"`
	Version    string              `json:"version,omitempty"`
	Command    string              `json:"command,omitempty"`
	Pagination *EnvelopePagination `json:"pagination,omitempty"`
}

// EnvelopePagination contains normalized pagination metadata for machine consumers.
// It is included in metadata for cursor-based commands when --envelope is enabled.
type EnvelopePagination struct {
	HasMore      bool   `json:"has_more"`
	Direction    string `json:"direction,omitempty"`
	NextCursor   string `json:"next_cursor,omitempty"`
	OldestCursor string `json:"oldest_cursor,omitempty"`
	NewestCursor string `json:"newest_cursor,omitempty"`
	AutoPaged    bool   `json:"auto_paged"`
	Capped       bool   `json:"capped"`
	MaxItems     int    `json:"max_items,omitempty"`
}

// Error codes
const (
	ErrCodeAuth       = "AUTH_ERROR"
	ErrCodeNotFound   = "NOT_FOUND"
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeConnection = "CONNECTION_ERROR"
	ErrCodeInternal   = "INTERNAL_ERROR"
)

type envelopeCtxKey struct{}

// WithEnvelope attaches envelope mode to a context.
func WithEnvelope(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, envelopeCtxKey{}, enabled)
}

// IsEnvelope returns true if envelope mode is enabled.
func IsEnvelope(ctx context.Context) bool {
	if v := ctx.Value(envelopeCtxKey{}); v != nil {
		if enabled, ok := v.(bool); ok {
			return enabled
		}
	}
	return false
}

// WriteEnvelope writes a success envelope to w.
func WriteEnvelope(w io.Writer, data any, version, command string) error {
	return WriteEnvelopeWithPagination(w, data, version, command, nil)
}

// WriteEnvelopeWithPagination writes a success envelope with optional pagination metadata.
func WriteEnvelopeWithPagination(w io.Writer, data any, version, command string, pagination *EnvelopePagination) error {
	env := Envelope{
		Success: true,
		Data:    data,
		Metadata: &EnvelopeMeta{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Version:    version,
			Command:    command,
			Pagination: pagination,
		},
	}
	return WriteJSON(w, env)
}

// WriteEnvelopeError writes an error envelope to w.
func WriteEnvelopeError(w io.Writer, code, message, version, command string) error {
	return WriteEnvelopeErrorWithHint(w, code, message, "", version, command)
}

// WriteEnvelopeErrorWithHint writes an error envelope with an optional hint.
func WriteEnvelopeErrorWithHint(w io.Writer, code, message, hint, version, command string) error {
	env := Envelope{
		Success: false,
		Error: &EnvelopeError{
			Code:    code,
			Message: message,
			Hint:    hint,
		},
		Metadata: &EnvelopeMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   version,
			Command:   command,
		},
	}
	return WriteJSON(w, env)
}
