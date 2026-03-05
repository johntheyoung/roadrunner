package outfmt

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFromFlags(t *testing.T) {
	mode, err := FromFlags(false, false, false)
	if err != nil {
		t.Fatalf("FromFlags(false,false,false) error: %v", err)
	}
	if mode.JSON || mode.JSONL || mode.Plain {
		t.Fatalf("FromFlags(false,false,false) = %+v, want all false", mode)
	}

	mode, err = FromFlags(true, false, false)
	if err != nil {
		t.Fatalf("FromFlags(true,false,false) error: %v", err)
	}
	if !mode.JSON || mode.JSONL || mode.Plain {
		t.Fatalf("FromFlags(true,false,false) = %+v, want JSON only", mode)
	}

	mode, err = FromFlags(false, true, false)
	if err != nil {
		t.Fatalf("FromFlags(false,true,false) error: %v", err)
	}
	if mode.JSON || !mode.JSONL || mode.Plain {
		t.Fatalf("FromFlags(false,true,false) = %+v, want JSONL only", mode)
	}

	if _, err = FromFlags(true, false, true); err == nil {
		t.Fatal("FromFlags(true,false,true) expected error")
	}
	if _, err = FromFlags(true, true, false); err == nil {
		t.Fatal("FromFlags(true,true,false) expected error")
	}
}

func TestRequestIDContext(t *testing.T) {
	ctx := context.Background()
	if got := RequestIDFromContext(ctx); got != "" {
		t.Fatalf("RequestIDFromContext() = %q, want empty", got)
	}

	ctx = WithRequestID(ctx, "req-123")
	if got := RequestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("RequestIDFromContext() = %q, want %q", got, "req-123")
	}
}

func TestWriteJSONLine(t *testing.T) {
	buf := &bytes.Buffer{}
	payload := map[string]any{"a": 1, "b": "two"}
	if err := WriteJSONLine(buf, payload); err != nil {
		t.Fatalf("WriteJSONLine() error: %v", err)
	}

	out := strings.TrimSuffix(buf.String(), "\n")
	if strings.Contains(out, "\n") {
		t.Fatalf("WriteJSONLine() output contains newline: %q", out)
	}
	if !strings.HasPrefix(out, "{") || !strings.HasSuffix(out, "}") {
		t.Fatalf("WriteJSONLine() output not object json: %q", out)
	}
}
