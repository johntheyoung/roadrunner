package beeperapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConnectInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/info" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"Beeper Desktop API",
			"version":"4.2.999",
			"runtime":"desktop",
			"endpoints":{
				"oauth_introspect":"http://localhost:23373/oauth/introspect",
				"ws":"ws://localhost:23373/v1/ws"
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	info, err := client.Connect().Info(context.Background())
	if err != nil {
		t.Fatalf("Info() error = %v", err)
	}

	if info.Name != "Beeper Desktop API" {
		t.Fatalf("name = %q, want %q", info.Name, "Beeper Desktop API")
	}
	if info.Version != "4.2.999" {
		t.Fatalf("version = %q, want %q", info.Version, "4.2.999")
	}
	if info.Endpoints["oauth_introspect"] == "" {
		t.Fatalf("oauth_introspect endpoint missing in %+v", info.Endpoints)
	}
	if info.Raw == nil {
		t.Fatal("raw metadata is nil")
	}
}

func TestConnectIntrospectFormEncoded(t *testing.T) {
	t.Parallel()

	var contentType string
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/introspect" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		contentType = r.Header.Get("Content-Type")
		payload, _ := io.ReadAll(r.Body)
		body = string(payload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"active": true,
			"client_id": "rr-cli",
			"sub": "user-123",
			"scope": "desktop_api",
			"exp": 1900000000
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Connect().Introspect(context.Background(), "token-abc")
	if err != nil {
		t.Fatalf("Introspect() error = %v", err)
	}

	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		t.Fatalf("content-type = %q, want form-urlencoded", contentType)
	}
	if !strings.Contains(body, "token=token-abc") {
		t.Fatalf("body = %q, expected token form field", body)
	}
	if !strings.Contains(body, "token_type_hint=access_token") {
		t.Fatalf("body = %q, expected token_type_hint field", body)
	}
	if !resp.Active {
		t.Fatal("active = false, want true")
	}
	if resp.Subject != "user-123" {
		t.Fatalf("subject = %q, want %q", resp.Subject, "user-123")
	}
	if resp.ExpiresAt != 1900000000 {
		t.Fatalf("expires_at = %d, want %d", resp.ExpiresAt, 1900000000)
	}
}
