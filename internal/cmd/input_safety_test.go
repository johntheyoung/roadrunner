package cmd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestMessagesSendRejectsPollutedChatID(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := MessagesSendCmd{
		ChatID: "!room:beeper.local?fields=id",
		Text:   "hello",
	}
	err := cmd.Run(context.Background(), &RootFlags{BaseURL: server.URL, Timeout: 1})
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
	if !strings.Contains(err.Error(), "chatID") {
		t.Fatalf("error = %q, want mention chatID", err.Error())
	}
}

func TestMessagesEditRejectsPollutedMessageID(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := MessagesEditCmd{
		ChatID:    "!room:beeper.local",
		MessageID: "msg-1#frag",
		Text:      "updated",
	}
	err := cmd.Run(context.Background(), &RootFlags{BaseURL: server.URL, Timeout: 1})
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
	if !strings.Contains(err.Error(), "messageID") {
		t.Fatalf("error = %q, want mention messageID", err.Error())
	}
}

func TestChatsArchiveRejectsEncodedTraversalChatID(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := ChatsArchiveCmd{ChatID: "%2e%2e/%2e%2e/secret"}
	err := cmd.Run(context.Background(), &RootFlags{DryRun: true, BaseURL: server.URL, Timeout: 1})
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
	if !strings.Contains(err.Error(), "chatID") {
		t.Fatalf("error = %q, want mention chatID", err.Error())
	}
}
