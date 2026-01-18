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
}

// EnvelopeMeta contains metadata about the response.
type EnvelopeMeta struct {
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
	Command   string `json:"command,omitempty"`
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
	env := Envelope{
		Success: true,
		Data:    data,
		Metadata: &EnvelopeMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   version,
			Command:   command,
		},
	}
	return WriteJSON(w, env)
}

// WriteEnvelopeError writes an error envelope to w.
func WriteEnvelopeError(w io.Writer, code, message, version, command string) error {
	env := Envelope{
		Success: false,
		Error: &EnvelopeError{
			Code:    code,
			Message: message,
		},
		Metadata: &EnvelopeMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   version,
			Command:   command,
		},
	}
	return WriteJSON(w, env)
}
