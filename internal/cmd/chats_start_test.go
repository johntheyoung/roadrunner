package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
)

func TestChatsStartRequiresUserIdentifier(t *testing.T) {
	cmd := ChatsStartCmd{
		AccountID: "acc",
	}
	err := cmd.Run(context.Background(), &RootFlags{})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *errfmt.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T, want *errfmt.ExitError", err)
	}
	if exitErr.Code != errfmt.ExitUsageError {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, errfmt.ExitUsageError)
	}
}

func TestChatsStartRequiresAccount(t *testing.T) {
	cmd := ChatsStartCmd{
		Email: "alice@example.com",
	}
	err := cmd.Run(context.Background(), &RootFlags{})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *errfmt.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T, want *errfmt.ExitError", err)
	}
	if exitErr.Code != errfmt.ExitUsageError {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, errfmt.ExitUsageError)
	}
}

func TestChatsStartSendsStartModePayload(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	var captured map[string]any
	startCalled := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chats" {
			http.NotFound(w, r)
			return
		}
		startCalled++
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chatID":"chat-1","status":"created"}`))
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := ChatsStartCmd{
		AccountID: "acc-1",
		Email:     "alice@example.com",
		FullName:  "Alice",
		Message:   "hello",
	}
	err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if startCalled != 1 {
		t.Fatalf("start endpoint calls = %d, want 1", startCalled)
	}
	if captured["mode"] != "start" {
		t.Fatalf("mode payload = %#v, want %q", captured["mode"], "start")
	}
	if captured["accountID"] != "acc-1" {
		t.Fatalf("accountID payload = %#v, want %q", captured["accountID"], "acc-1")
	}
	if captured["messageText"] != "hello" {
		t.Fatalf("messageText payload = %#v, want %q", captured["messageText"], "hello")
	}
	user, ok := captured["user"].(map[string]any)
	if !ok {
		t.Fatalf("user payload type = %T, want object", captured["user"])
	}
	if user["email"] != "alice@example.com" {
		t.Fatalf("user.email payload = %#v, want %q", user["email"], "alice@example.com")
	}
}
