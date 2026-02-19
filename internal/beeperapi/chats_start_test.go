package beeperapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatsStartPayload(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/v1/chats" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chatID":"chat-123","status":"existing"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	allowInvite := true
	resp, err := client.Chats().Start(context.Background(), ChatStartParams{
		AccountID: "acc-1",
		User: ChatStartUser{
			Email:    "alice@example.com",
			FullName: "Alice",
		},
		AllowInvite: &allowInvite,
		MessageText: "hello",
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if resp.ChatID != "chat-123" {
		t.Fatalf("ChatID = %q, want %q", resp.ChatID, "chat-123")
	}
	if resp.Status != "existing" {
		t.Fatalf("Status = %q, want %q", resp.Status, "existing")
	}

	if captured["accountID"] != "acc-1" {
		t.Fatalf("accountID payload = %#v, want %q", captured["accountID"], "acc-1")
	}
	if captured["mode"] != "start" {
		t.Fatalf("mode payload = %#v, want %q", captured["mode"], "start")
	}
	if captured["messageText"] != "hello" {
		t.Fatalf("messageText payload = %#v, want %q", captured["messageText"], "hello")
	}
	if captured["allowInvite"] != true {
		t.Fatalf("allowInvite payload = %#v, want true", captured["allowInvite"])
	}
	user, ok := captured["user"].(map[string]any)
	if !ok {
		t.Fatalf("user payload type = %T, want object", captured["user"])
	}
	if user["email"] != "alice@example.com" {
		t.Fatalf("user.email payload = %#v, want %q", user["email"], "alice@example.com")
	}
	if user["fullName"] != "Alice" {
		t.Fatalf("user.fullName payload = %#v, want %q", user["fullName"], "Alice")
	}
}

func TestChatsStartWithoutStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chats" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chatID":"chat-456"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Chats().Start(context.Background(), ChatStartParams{
		AccountID: "acc-1",
		User: ChatStartUser{
			Username: "alice",
		},
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if resp.ChatID != "chat-456" {
		t.Fatalf("ChatID = %q, want %q", resp.ChatID, "chat-456")
	}
	if resp.Status != "" {
		t.Fatalf("Status = %q, want empty", resp.Status)
	}
}
