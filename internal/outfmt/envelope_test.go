package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestWithEnvelope(t *testing.T) {
	ctx := context.Background()

	// Default should be false
	if IsEnvelope(ctx) {
		t.Error("IsEnvelope() should be false by default")
	}

	// Enabled
	ctx = WithEnvelope(ctx, true)
	if !IsEnvelope(ctx) {
		t.Error("IsEnvelope() should be true after WithEnvelope(ctx, true)")
	}

	// Disabled
	ctx = WithEnvelope(ctx, false)
	if IsEnvelope(ctx) {
		t.Error("IsEnvelope() should be false after WithEnvelope(ctx, false)")
	}
}

func TestWriteEnvelope(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"foo": "bar"}

	err := WriteEnvelope(&buf, data, "1.0.0", "test cmd")
	if err != nil {
		t.Fatalf("WriteEnvelope() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if !env.Success {
		t.Error("expected success=true")
	}
	if env.Error != nil {
		t.Error("expected error=nil")
	}
	if env.Data == nil {
		t.Error("expected data to be set")
	}
	if env.Metadata == nil {
		t.Fatal("expected metadata to be set")
	}
	if env.Metadata.Version != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %s", env.Metadata.Version)
	}
	if env.Metadata.Command != "test cmd" {
		t.Errorf("expected command='test cmd', got %s", env.Metadata.Command)
	}
	if env.Metadata.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestWriteEnvelopeError(t *testing.T) {
	var buf bytes.Buffer

	err := WriteEnvelopeError(&buf, ErrCodeNotFound, "chat not found", "1.0.0", "chats get")
	if err != nil {
		t.Fatalf("WriteEnvelopeError() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if env.Success {
		t.Error("expected success=false")
	}
	if env.Data != nil {
		t.Error("expected data=nil")
	}
	if env.Error == nil {
		t.Fatal("expected error to be set")
	}
	if env.Error.Code != ErrCodeNotFound {
		t.Errorf("expected code=%s, got %s", ErrCodeNotFound, env.Error.Code)
	}
	if env.Error.Message != "chat not found" {
		t.Errorf("expected message='chat not found', got %s", env.Error.Message)
	}
	if env.Metadata == nil {
		t.Fatal("expected metadata to be set")
	}
	if env.Metadata.Version != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %s", env.Metadata.Version)
	}
	if env.Metadata.Command != "chats get" {
		t.Errorf("expected command='chats get', got %s", env.Metadata.Command)
	}
}

func TestEnvelopeJSONStructure(t *testing.T) {
	tests := []struct {
		name     string
		envelope Envelope
		wantKeys []string
	}{
		{
			name: "success envelope has success, data, metadata",
			envelope: Envelope{
				Success:  true,
				Data:     map[string]string{"foo": "bar"},
				Metadata: &EnvelopeMeta{Timestamp: "2024-01-01T00:00:00Z"},
			},
			wantKeys: []string{"success", "data", "metadata"},
		},
		{
			name: "error envelope has success, error, metadata",
			envelope: Envelope{
				Success:  false,
				Error:    &EnvelopeError{Code: "TEST", Message: "test"},
				Metadata: &EnvelopeMeta{Timestamp: "2024-01-01T00:00:00Z"},
			},
			wantKeys: []string{"success", "error", "metadata"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.envelope)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			for _, key := range tt.wantKeys {
				if _, ok := m[key]; !ok {
					t.Errorf("expected key %q in JSON", key)
				}
			}
		})
	}
}
