package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

func TestMessagesSendRequiresTextOrAttachment(t *testing.T) {
	cmd := MessagesSendCmd{
		ChatID: "!room:beeper.local",
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

func TestMessagesSendChatTargetConflict(t *testing.T) {
	cmd := MessagesSendCmd{
		ChatID:             "!room:beeper.local",
		Chat:               "Alice",
		AttachmentUploadID: "up_123",
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

func TestMessagesSendInputSourceConflict(t *testing.T) {
	cmd := MessagesSendCmd{
		ChatID:             "!room:beeper.local",
		Text:               "hello",
		TextFile:           "message.txt",
		AttachmentUploadID: "up_123",
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

func TestMessagesSendAttachmentOverridesRequireUploadID(t *testing.T) {
	cmd := MessagesSendCmd{
		ChatID:             "!room:beeper.local",
		Text:               "hello",
		AttachmentFileName: "photo.jpg",
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

func TestMessagesSendAttachmentSizeRequiresBothDimensions(t *testing.T) {
	cmd := MessagesSendCmd{
		ChatID:             "!room:beeper.local",
		Text:               "hello",
		AttachmentUploadID: "up_123",
		AttachmentWidth:    "1280",
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

func TestMessagesSendResolvesChatQuery(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	searchCalled := 0
	sendCalled := 0
	var sendPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/chats/search":
			searchCalled++
			if got := r.URL.Query().Get("query"); got != "Alice" {
				http.Error(w, "bad query", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"items": [
					{"id":"chat-1","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Alice","type":"single","unreadCount":0}
				],
				"hasMore": false,
				"oldestCursor": "",
				"newestCursor": ""
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/chats/chat-1/messages":
			sendCalled++
			if err := json.NewDecoder(r.Body).Decode(&sendPayload); err != nil {
				http.Error(w, "bad payload", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"chatID":"chat-1","pendingMessageID":"pending-123"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	testUI, err := ui.New(ui.Options{
		Stdout: outBuf,
		Stderr: errBuf,
		Color:  "never",
	})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}

	ctx := ui.WithUI(context.Background(), testUI)
	cmd := MessagesSendCmd{
		Chat: "Alice",
		Text: "hello",
	}
	if err := cmd.Run(ctx, &RootFlags{
		BaseURL: server.URL,
		Timeout: 5,
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if searchCalled != 1 {
		t.Fatalf("search endpoint calls = %d, want 1", searchCalled)
	}
	if sendCalled != 1 {
		t.Fatalf("send endpoint calls = %d, want 1", sendCalled)
	}
	if sendPayload["text"] != "hello" {
		t.Fatalf("send text payload = %#v, want %q", sendPayload["text"], "hello")
	}
}

func TestMessagesSendChatQueryAmbiguous(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	sendCalled := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/chats/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"items": [
					{"id":"chat-1","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Alice","type":"single","unreadCount":0},
					{"id":"chat-2","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Alice","type":"single","unreadCount":0}
				],
				"hasMore": false,
				"oldestCursor": "",
				"newestCursor": ""
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/chats/chat-1/messages":
			sendCalled++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"chatID":"chat-1","pendingMessageID":"pending-123"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	testUI, err := ui.New(ui.Options{Color: "never"})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}
	ctx := ui.WithUI(context.Background(), testUI)
	cmd := MessagesSendCmd{
		Chat: "Alice",
		Text: "hello",
	}
	err = cmd.Run(ctx, &RootFlags{
		BaseURL: server.URL,
		Timeout: 5,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *errfmt.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T, want *errfmt.ExitError", err)
	}
	if exitErr.Code != errfmt.ExitFailure {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, errfmt.ExitFailure)
	}
	if sendCalled != 0 {
		t.Fatalf("send endpoint calls = %d, want 0", sendCalled)
	}
}
