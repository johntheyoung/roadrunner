package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
)

func TestEventsTailJSONOutputsDomainEvents(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ws" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		var sub map[string]any
		if err := conn.ReadJSON(&sub); err != nil {
			return
		}
		_ = conn.WriteJSON(map[string]any{
			"type":   "message.upserted",
			"seq":    1,
			"ts":     1739320000000,
			"chatID": "chat_a",
			"ids":    []string{"m1"},
		})
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := EventsTailCmd{
		All:       true,
		StopAfter: 200 * time.Millisecond,
		Reconnect: false,
	}

	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	if !strings.Contains(out, `"type":"message.upserted"`) {
		t.Fatalf("expected message.upserted in output, got: %s", out)
	}
}

func TestEventsTailUnsupportedRouteMessage(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := EventsTailCmd{
		All:       true,
		Reconnect: false,
		StopAfter: 100 * time.Millisecond,
	}
	err := cmd.Run(testJSONContext(t), &RootFlags{BaseURL: server.URL, Timeout: 5})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "websocket events are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
