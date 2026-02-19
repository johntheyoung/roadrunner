package beeperapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessagesReactPayload(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/v1/chats/chat-1/messages/msg-1/reactions" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("reactionKey"); got != "👍" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if err := client.Messages().React(context.Background(), "chat-1", "msg-1", "👍"); err != nil {
		t.Fatalf("React() error = %v", err)
	}

	if captured["reactionKey"] != "👍" {
		t.Fatalf("reactionKey payload = %#v, want %q", captured["reactionKey"], "👍")
	}
}

func TestMessagesUnreactPayload(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/v1/chats/chat-1/messages/msg-1/reactions" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("reactionKey"); got != ":thumbsup:" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if err := client.Messages().Unreact(context.Background(), "chat-1", "msg-1", ":thumbsup:"); err != nil {
		t.Fatalf("Unreact() error = %v", err)
	}

	if captured["reactionKey"] != ":thumbsup:" {
		t.Fatalf("reactionKey payload = %#v, want %q", captured["reactionKey"], ":thumbsup:")
	}
}
