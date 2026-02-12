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
	if env.Metadata.Pagination != nil {
		t.Error("expected pagination metadata to be nil by default")
	}
	if env.Metadata.RequestID != "" {
		t.Errorf("expected request_id empty, got %q", env.Metadata.RequestID)
	}
}

func TestWriteEnvelopeWithPagination(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"foo": "bar"}
	pagination := &EnvelopePagination{
		HasMore:      true,
		Direction:    "before",
		OldestCursor: "old",
		NewestCursor: "new",
		AutoPaged:    true,
		Capped:       true,
		MaxItems:     50,
	}

	err := WriteEnvelopeWithPagination(&buf, data, "1.0.0", "chats list", pagination)
	if err != nil {
		t.Fatalf("WriteEnvelopeWithPagination() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if env.Metadata == nil {
		t.Fatal("expected metadata to be set")
	}
	if env.Metadata.Pagination == nil {
		t.Fatal("expected pagination metadata to be set")
	}
	if !env.Metadata.Pagination.HasMore {
		t.Error("expected has_more=true")
	}
	if env.Metadata.Pagination.Direction != "before" {
		t.Errorf("expected direction=before, got %q", env.Metadata.Pagination.Direction)
	}
	if env.Metadata.Pagination.OldestCursor != "old" {
		t.Errorf("expected oldest_cursor=old, got %q", env.Metadata.Pagination.OldestCursor)
	}
	if env.Metadata.Pagination.NewestCursor != "new" {
		t.Errorf("expected newest_cursor=new, got %q", env.Metadata.Pagination.NewestCursor)
	}
	if !env.Metadata.Pagination.AutoPaged {
		t.Error("expected auto_paged=true")
	}
	if !env.Metadata.Pagination.Capped {
		t.Error("expected capped=true")
	}
	if env.Metadata.Pagination.MaxItems != 50 {
		t.Errorf("expected max_items=50, got %d", env.Metadata.Pagination.MaxItems)
	}
	if env.Metadata.RequestID != "" {
		t.Errorf("expected request_id empty, got %q", env.Metadata.RequestID)
	}
}

func TestWriteEnvelopeWithMetadataRequestID(t *testing.T) {
	var buf bytes.Buffer

	err := WriteEnvelopeWithMetadata(&buf, map[string]any{"ok": true}, "1.0.0", "version", nil, "req-abc")
	if err != nil {
		t.Fatalf("WriteEnvelopeWithMetadata() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}
	if env.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if env.Metadata.RequestID != "req-abc" {
		t.Fatalf("request_id = %q, want %q", env.Metadata.RequestID, "req-abc")
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
	if env.Error.Hint != "" {
		t.Errorf("expected hint empty, got %q", env.Error.Hint)
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
	if env.Metadata.RequestID != "" {
		t.Errorf("expected request_id empty, got %q", env.Metadata.RequestID)
	}
}

func TestWriteEnvelopeErrorWithHint(t *testing.T) {
	var buf bytes.Buffer

	err := WriteEnvelopeErrorWithHint(&buf, ErrCodeValidation, "bad input", "use --help", "1.0.0", "messages send")
	if err != nil {
		t.Fatalf("WriteEnvelopeErrorWithHint() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if env.Error == nil {
		t.Fatal("expected error to be set")
	}
	if env.Error.Hint != "use --help" {
		t.Fatalf("hint = %q, want %q", env.Error.Hint, "use --help")
	}
	if env.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if env.Metadata.RequestID != "" {
		t.Fatalf("request_id = %q, want empty", env.Metadata.RequestID)
	}
}

func TestWriteEnvelopeErrorWithMetadataRequestID(t *testing.T) {
	var buf bytes.Buffer

	err := WriteEnvelopeErrorWithMetadata(&buf, ErrCodeValidation, "bad input", "use --help", "1.0.0", "messages send", "req-xyz")
	if err != nil {
		t.Fatalf("WriteEnvelopeErrorWithMetadata() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}
	if env.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if env.Metadata.RequestID != "req-xyz" {
		t.Fatalf("request_id = %q, want %q", env.Metadata.RequestID, "req-xyz")
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
