package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

func TestMessagesSend_DryRunSkipsAPI(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		http.NotFound(w, r)
	}))
	defer server.Close()

	testUI, err := ui.New(ui.Options{Color: "never"})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}

	ctx := context.Background()
	ctx = ui.WithUI(ctx, testUI)
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})

	cmd := MessagesSendCmd{ChatID: "!room:beeper.local", Text: "hello"}
	var runErr error
	output, errText := captureOutput(t, func() {
		runErr = cmd.Run(ctx, &RootFlags{DryRun: true, BaseURL: server.URL, Timeout: 5})
	})
	err = runErr
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(errText) != "" {
		t.Fatalf("stderr not empty: %q", errText)
	}
	if requestCount != 0 {
		t.Fatalf("request count = %d, want 0", requestCount)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, output)
	}
	if payload["dry_run"] != true {
		t.Fatalf("dry_run = %#v, want true", payload["dry_run"])
	}
	if payload["command"] != "messages send" {
		t.Fatalf("command = %#v, want %q", payload["command"], "messages send")
	}
}

func TestChatsArchive_DryRunSkipsAPIAndConfirmation(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		http.NotFound(w, r)
	}))
	defer server.Close()

	testUI, err := ui.New(ui.Options{Color: "never"})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}

	ctx := context.Background()
	ctx = ui.WithUI(ctx, testUI)
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})

	cmd := ChatsArchiveCmd{ChatID: "!room:beeper.local"}
	var runErr error
	output, errText := captureOutput(t, func() {
		runErr = cmd.Run(ctx, &RootFlags{DryRun: true, BaseURL: server.URL, Timeout: 5})
	})
	err = runErr
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(errText) != "" {
		t.Fatalf("stderr not empty: %q", errText)
	}
	if requestCount != 0 {
		t.Fatalf("request count = %d, want 0", requestCount)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, output)
	}
	if payload["dry_run"] != true {
		t.Fatalf("dry_run = %#v, want true", payload["dry_run"])
	}
	if payload["command"] != "chats archive" {
		t.Fatalf("command = %#v, want %q", payload["command"], "chats archive")
	}
}
