package beeperapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAssetsServeStreamsBytesAndMetadata(t *testing.T) {
	const body = "served-bytes"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/assets/serve" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("url"); got != "mxc://beeper.local/abc" {
			http.Error(w, "bad query url", http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("Accept"); got != "*/*" {
			http.Error(w, "bad accept header", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client, err := NewClient("token", server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var dst bytes.Buffer
	res, err := client.Assets().Serve(context.Background(), "mxc://beeper.local/abc", &dst)
	if err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	if got := dst.String(); got != body {
		t.Fatalf("dst = %q, want %q", got, body)
	}
	if res.ContentType != "application/octet-stream" {
		t.Fatalf("content type = %q, want %q", res.ContentType, "application/octet-stream")
	}
	if res.BytesWritten != int64(len(body)) {
		t.Fatalf("bytes written = %d, want %d", res.BytesWritten, len(body))
	}
}
