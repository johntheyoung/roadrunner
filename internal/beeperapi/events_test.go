package beeperapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestEventsConnectSubscribeAndRead(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ws" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		var sub map[string]any
		if err := conn.ReadJSON(&sub); err != nil {
			t.Logf("read sub: %v", err)
			return
		}
		if sub["type"] != "subscriptions.set" {
			t.Errorf("subscription type = %#v, want subscriptions.set", sub["type"])
		}
		if sub["requestID"] != "r1" {
			t.Errorf("requestID = %#v, want r1", sub["requestID"])
		}

		if err := conn.WriteJSON(map[string]any{
			"type":    "ready",
			"version": 1,
			"chatIDs": []string{"*"},
		}); err != nil {
			t.Logf("write ready: %v", err)
			return
		}
		_ = conn.WriteJSON(map[string]any{
			"type":   "message.upserted",
			"seq":    42,
			"ts":     1739320000000,
			"chatID": "chat_a",
			"ids":    []string{"m1"},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ws, err := client.Events().Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = ws.Close() }()

	if err := ws.SetSubscriptions(context.Background(), "r1", []string{"*"}); err != nil {
		t.Fatalf("SetSubscriptions() error = %v", err)
	}

	first, err := ws.ReadEvent(context.Background())
	if err != nil {
		t.Fatalf("ReadEvent(1) error = %v", err)
	}
	if first.Type != "ready" {
		t.Fatalf("first type = %q, want ready", first.Type)
	}

	second, err := ws.ReadEvent(context.Background())
	if err != nil {
		t.Fatalf("ReadEvent(2) error = %v", err)
	}
	if second.Type != "message.upserted" {
		t.Fatalf("second type = %q, want message.upserted", second.Type)
	}
	if second.Seq != 42 {
		t.Fatalf("seq = %d, want 42", second.Seq)
	}
}

func TestEventsConnectUnsupportedRoute(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Events().Connect(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsEventsUnsupported(err) {
		t.Fatalf("IsEventsUnsupported(%v) = false, want true", err)
	}
}

func TestSetSubscriptionsRejectsMixedWildcard(t *testing.T) {
	t.Parallel()

	// Use a nil connection wrapper to only validate inputs without dialing.
	ws := &EventsConnection{}

	err := ws.SetSubscriptions(context.Background(), "r1", []string{"*", "chat_a"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEventDecodeWithEntries(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"type":"message.upserted","seq":7,"ts":123,"chatID":"c1","ids":["m1"],"entries":[{"id":"m1","text":"hello"}]}`)
	var evt Event
	if err := json.Unmarshal(raw, &evt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(evt.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(evt.Entries))
	}
	if evt.Type != "message.upserted" {
		t.Fatalf("type = %q, want message.upserted", evt.Type)
	}
}
