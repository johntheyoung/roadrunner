package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func TestEventsTailSkipsControlMessagesByDefault(t *testing.T) {
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
			"type":    "ready",
			"version": 1,
			"chatIDs": []string{"*"},
		})
		_ = conn.WriteJSON(map[string]any{
			"type":   "message.upserted",
			"seq":    2,
			"ts":     1739320000001,
			"chatID": "chat_a",
			"ids":    []string{"m2"},
		})
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := EventsTailCmd{
		All:       true,
		Reconnect: false,
		StopAfter: 250 * time.Millisecond,
	}

	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	if strings.Contains(out, `"type":"ready"`) {
		t.Fatalf("expected ready control message to be filtered, got: %s", out)
	}
	if !strings.Contains(out, `"type":"message.upserted"`) {
		t.Fatalf("expected message.upserted in output, got: %s", out)
	}
}

func TestEventsTailIncludesControlMessagesWhenFlagSet(t *testing.T) {
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
			"type":    "ready",
			"version": 1,
			"chatIDs": []string{"*"},
		})
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := EventsTailCmd{
		All:            true,
		IncludeControl: true,
		Reconnect:      false,
		StopAfter:      250 * time.Millisecond,
	}

	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	if !strings.Contains(out, `"type":"ready"`) {
		t.Fatalf("expected ready control message in output, got: %s", out)
	}
}

func TestEventsTailReconnectsAndResubscribesAfterDisconnect(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	var connectCount atomic.Int32
	var subscriptionCount atomic.Int32

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

		attempt := connectCount.Add(1)

		var sub map[string]any
		if err := conn.ReadJSON(&sub); err != nil {
			return
		}
		subscriptionCount.Add(1)

		if attempt == 1 {
			// Force reconnect by closing before any domain event is emitted.
			return
		}

		_ = conn.WriteJSON(map[string]any{
			"type":   "message.upserted",
			"seq":    3,
			"ts":     1739320000002,
			"chatID": "chat_b",
			"ids":    []string{"m3"},
		})
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := EventsTailCmd{
		All:            true,
		Reconnect:      true,
		ReconnectDelay: 10 * time.Millisecond,
		StopAfter:      300 * time.Millisecond,
	}

	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	if !strings.Contains(out, `"type":"message.upserted"`) {
		t.Fatalf("expected message.upserted in output, got: %s", out)
	}
	if got := connectCount.Load(); got < 2 {
		t.Fatalf("expected reconnect attempts >= 2, got %d", got)
	}
	if got := subscriptionCount.Load(); got < 2 {
		t.Fatalf("expected subscriptions to be re-sent after reconnect, got %d", got)
	}
}

func TestEventsTailReconnectsAfterTemporaryHandshakeFailure(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ws" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		attempt := requestCount.Add(1)
		if attempt == 1 {
			http.Error(w, "temporary failure", http.StatusInternalServerError)
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
			"seq":    4,
			"ts":     1739320000003,
			"chatID": "chat_c",
			"ids":    []string{"m4"},
		})
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := EventsTailCmd{
		All:            true,
		Reconnect:      true,
		ReconnectDelay: 10 * time.Millisecond,
		StopAfter:      350 * time.Millisecond,
	}

	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	if !strings.Contains(out, `"type":"message.upserted"`) {
		t.Fatalf("expected message.upserted in output, got: %s", out)
	}
	if got := requestCount.Load(); got < 2 {
		t.Fatalf("expected at least two connection attempts, got %d", got)
	}
}
