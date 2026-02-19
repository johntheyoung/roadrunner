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

func TestMessagesReactRequiresReactionKey(t *testing.T) {
	cmd := MessagesReactCmd{
		ChatID:    "!room:beeper.local",
		MessageID: "msg-1",
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

func TestMessagesReactCallsEndpoint(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	reactCalled := 0
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chats/!room:beeper.local/messages/msg-1/reactions" {
			http.NotFound(w, r)
			return
		}
		reactCalled++
		if got := r.URL.Query().Get("reactionKey"); got != "👍" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := MessagesReactCmd{
		ChatID:      "!room:beeper.local",
		MessageID:   "msg-1",
		ReactionKey: "👍",
	}
	err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if reactCalled != 1 {
		t.Fatalf("react endpoint calls = %d, want 1", reactCalled)
	}
	if captured["reactionKey"] != "👍" {
		t.Fatalf("reactionKey payload = %#v, want %q", captured["reactionKey"], "👍")
	}
}

func TestMessagesUnreactCallsEndpoint(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	unreactCalled := 0
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/chats/!room:beeper.local/messages/msg-1/reactions" {
			http.NotFound(w, r)
			return
		}
		unreactCalled++
		if got := r.URL.Query().Get("reactionKey"); got != ":thumbsup:" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	cmd := MessagesUnreactCmd{
		ChatID:      "!room:beeper.local",
		MessageID:   "msg-1",
		ReactionKey: ":thumbsup:",
	}
	err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if unreactCalled != 1 {
		t.Fatalf("unreact endpoint calls = %d, want 1", unreactCalled)
	}
	if captured["reactionKey"] != ":thumbsup:" {
		t.Fatalf("reactionKey payload = %#v, want %q", captured["reactionKey"], ":thumbsup:")
	}
}
