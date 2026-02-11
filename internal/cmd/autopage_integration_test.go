package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

func testJSONContext(t *testing.T) context.Context {
	t.Helper()

	testUI, err := ui.New(ui.Options{Color: "never"})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}
	ctx := ui.WithUI(context.Background(), testUI)
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})
	return ctx
}

func TestChatsListAllAutoPagination(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	requestCount := 0
	cursorValues := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chats" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		cursor := r.URL.Query().Get("cursor")
		cursorValues = append(cursorValues, cursor)

		w.Header().Set("Content-Type", "application/json")
		if cursor == "" {
			_, _ = w.Write([]byte(`{
				"items": [
					{"id":"!chat1:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 1","type":"group","unreadCount":0},
					{"id":"!chat2:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 2","type":"group","unreadCount":0}
				],
				"hasMore": true,
				"oldestCursor": "c1",
				"newestCursor": "n1"
			}`))
			return
		}
		if cursor != "c1" {
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{
			"items": [
				{"id":"!chat3:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 3","type":"group","unreadCount":0}
			],
			"hasMore": false,
			"oldestCursor": "",
			"newestCursor": ""
		}`))
	}))
	defer server.Close()

	ctx := testJSONContext(t)
	cmd := ChatsListCmd{
		All:      true,
		MaxItems: 10,
	}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var resp beeperapi.ChatListResult
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}

	if len(resp.Items) != 3 {
		t.Fatalf("items len = %d, want 3", len(resp.Items))
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want 2", requestCount)
	}
	if len(cursorValues) != 2 || cursorValues[0] != "" || cursorValues[1] != "c1" {
		t.Fatalf("unexpected cursor sequence: %#v", cursorValues)
	}
}

func TestChatsSearchAllAutoPagination(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	searchRequestCount := 0
	accountsRequestCount := 0
	cursorValues := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/accounts":
			accountsRequestCount++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"accountID":"acc1","network":"WhatsApp","user":{"id":"u1"}}]`))
		case "/v1/chats/search":
			searchRequestCount++
			cursor := r.URL.Query().Get("cursor")
			cursorValues = append(cursorValues, cursor)
			w.Header().Set("Content-Type", "application/json")
			if cursor == "" {
				_, _ = w.Write([]byte(`{
					"items": [
						{"id":"!chat1:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 1","type":"group","unreadCount":0},
						{"id":"!chat2:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 2","type":"group","unreadCount":0}
					],
					"hasMore": true,
					"oldestCursor": "c1",
					"newestCursor": "n1"
				}`))
				return
			}
			if cursor != "c1" {
				http.Error(w, "bad cursor", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{
				"items": [
					{"id":"!chat3:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 3","type":"group","unreadCount":0}
				],
				"hasMore": false,
				"oldestCursor": "",
				"newestCursor": ""
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ctx := testJSONContext(t)
	cmd := ChatsSearchCmd{
		Query:    "chat",
		Limit:    50,
		All:      true,
		MaxItems: 10,
	}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var resp beeperapi.ChatSearchResult
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}

	if len(resp.Items) != 3 {
		t.Fatalf("items len = %d, want 3", len(resp.Items))
	}
	if searchRequestCount != 2 {
		t.Fatalf("search request count = %d, want 2", searchRequestCount)
	}
	if accountsRequestCount != 1 {
		t.Fatalf("accounts request count = %d, want 1", accountsRequestCount)
	}
	if len(cursorValues) != 2 || cursorValues[0] != "" || cursorValues[1] != "c1" {
		t.Fatalf("unexpected cursor sequence: %#v", cursorValues)
	}
	if resp.Items[0].Network != "WhatsApp" {
		t.Fatalf("network = %q, want %q", resp.Items[0].Network, "WhatsApp")
	}
}

func TestMessagesListAllAutoPagination(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	requestCount := 0
	cursorValues := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chats/!room:beeper.local/messages" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		cursor := r.URL.Query().Get("cursor")
		cursorValues = append(cursorValues, cursor)

		w.Header().Set("Content-Type", "application/json")
		if cursor == "" {
			_, _ = w.Write([]byte(`{
				"items": [
					{"id":"msg1","accountID":"acc1","chatID":"!room:beeper.local","senderID":"u1","sortKey":"s1","timestamp":"2026-02-11T00:00:00Z","text":"one"},
					{"id":"msg2","accountID":"acc1","chatID":"!room:beeper.local","senderID":"u1","sortKey":"s2","timestamp":"2026-02-11T00:01:00Z","text":"two"}
				],
				"hasMore": true
			}`))
			return
		}
		if cursor != "s2" {
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{
			"items": [
				{"id":"msg3","accountID":"acc1","chatID":"!room:beeper.local","senderID":"u1","sortKey":"s3","timestamp":"2026-02-11T00:02:00Z","text":"three"}
			],
			"hasMore": false
		}`))
	}))
	defer server.Close()

	ctx := testJSONContext(t)
	cmd := MessagesListCmd{
		ChatID:   "!room:beeper.local",
		All:      true,
		MaxItems: 10,
	}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var resp beeperapi.MessageListResult
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}
	if len(resp.Items) != 3 {
		t.Fatalf("items len = %d, want 3", len(resp.Items))
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want 2", requestCount)
	}
	if len(cursorValues) != 2 || cursorValues[0] != "" || cursorValues[1] != "s2" {
		t.Fatalf("unexpected cursor sequence: %#v", cursorValues)
	}
}

func TestMessagesSearchAllAutoPaginationWithCap(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages/search" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{"id":"msg1","accountID":"acc1","chatID":"!room:beeper.local","senderID":"u1","sortKey":"s1","timestamp":"2026-02-11T00:00:00Z","text":"one"},
				{"id":"msg2","accountID":"acc1","chatID":"!room:beeper.local","senderID":"u1","sortKey":"s2","timestamp":"2026-02-11T00:01:00Z","text":"two"}
			],
			"hasMore": true,
			"oldestCursor": "c1",
			"newestCursor": "n1"
		}`))
	}))
	defer server.Close()

	ctx := testJSONContext(t)
	cmd := MessagesSearchCmd{
		Query:    "msg",
		Limit:    20,
		All:      true,
		MaxItems: 2,
	}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var resp beeperapi.MessageSearchResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &resp); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items len = %d, want 2", len(resp.Items))
	}
	if !resp.HasMore {
		t.Fatalf("has_more = %t, want true due to cap", resp.HasMore)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1 because --max-items capped first page", requestCount)
	}
}
