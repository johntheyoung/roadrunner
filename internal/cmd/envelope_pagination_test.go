package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
)

func testEnvelopeJSONContext(t *testing.T) context.Context {
	t.Helper()
	ctx := testJSONContext(t)
	return outfmt.WithEnvelope(ctx, true)
}

func TestChatsListEnvelopePaginationMetadata(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chats" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{"id":"!chat1:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 1","type":"group","unreadCount":0}
			],
			"hasMore": true,
			"oldestCursor": "c1",
			"newestCursor": "n1"
		}`))
	}))
	defer server.Close()

	cmd := ChatsListCmd{}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(testEnvelopeJSONContext(t), &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var env struct {
		Success  bool                     `json:"success"`
		Data     beeperapi.ChatListResult `json:"data"`
		Metadata struct {
			Pagination *outfmt.EnvelopePagination `json:"pagination"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}

	if !env.Success {
		t.Fatal("expected success=true")
	}
	if env.Metadata.Pagination == nil {
		t.Fatal("expected pagination metadata")
	}
	if !env.Metadata.Pagination.HasMore {
		t.Fatal("expected has_more=true")
	}
	if env.Metadata.Pagination.OldestCursor != "c1" || env.Metadata.Pagination.NewestCursor != "n1" {
		t.Fatalf("unexpected cursors: %+v", env.Metadata.Pagination)
	}
	if env.Metadata.Pagination.AutoPaged {
		t.Fatal("expected auto_paged=false")
	}
	if env.Metadata.Pagination.Capped {
		t.Fatal("expected capped=false")
	}
}

func TestMessagesListEnvelopePaginationMetadata(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chats/!room:beeper.local/messages" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{"id":"msg1","accountID":"acc1","chatID":"!room:beeper.local","senderID":"u1","sortKey":"s1","timestamp":"2026-02-11T00:00:00Z","text":"one"}
			],
			"hasMore": true
		}`))
	}))
	defer server.Close()

	cmd := MessagesListCmd{ChatID: "!room:beeper.local"}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(testEnvelopeJSONContext(t), &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var env struct {
		Success  bool                        `json:"success"`
		Data     beeperapi.MessageListResult `json:"data"`
		Metadata struct {
			Pagination *outfmt.EnvelopePagination `json:"pagination"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}

	if env.Metadata.Pagination == nil {
		t.Fatal("expected pagination metadata")
	}
	if env.Metadata.Pagination.NextCursor != "s1" {
		t.Fatalf("next_cursor = %q, want %q", env.Metadata.Pagination.NextCursor, "s1")
	}
	if !env.Metadata.Pagination.HasMore {
		t.Fatal("expected has_more=true")
	}
}

func TestSearchEnvelopePaginationMetadataForMessagesAllCap(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	searchRequestCount := 0
	messageSearchRequestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/accounts":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"accountID":"acc1","network":"WhatsApp","user":{"id":"u1"}}]`))
		case "/v1/search":
			searchRequestCount++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"results": {
					"chats": [],
					"in_groups": [],
					"messages": {"chats": {}, "hasMore": true, "items": [], "oldestCursor": "c1", "newestCursor": "n1"}
				}
			}`))
		case "/v1/messages/search":
			messageSearchRequestCount++
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
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cmd := SearchCmd{
		Query:            "msg",
		MessagesAll:      true,
		MessagesMaxItems: 2,
	}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(testEnvelopeJSONContext(t), &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var env struct {
		Success  bool                   `json:"success"`
		Data     beeperapi.SearchResult `json:"data"`
		Metadata struct {
			Pagination *outfmt.EnvelopePagination `json:"pagination"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}

	if searchRequestCount != 1 {
		t.Fatalf("search request count = %d, want 1", searchRequestCount)
	}
	if messageSearchRequestCount != 1 {
		t.Fatalf("message search request count = %d, want 1", messageSearchRequestCount)
	}
	if env.Metadata.Pagination == nil {
		t.Fatal("expected pagination metadata")
	}
	if !env.Metadata.Pagination.AutoPaged {
		t.Fatal("expected auto_paged=true")
	}
	if !env.Metadata.Pagination.Capped {
		t.Fatal("expected capped=true")
	}
	if env.Metadata.Pagination.MaxItems != 2 {
		t.Fatalf("max_items = %d, want 2", env.Metadata.Pagination.MaxItems)
	}
	if !env.Metadata.Pagination.HasMore {
		t.Fatal("expected has_more=true")
	}
}

func TestUnreadEnvelopePaginationMetadata(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/accounts":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"accountID":"acc1","network":"WhatsApp","user":{"id":"u1"}}]`))
		case "/v1/chats/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"items": [
					{"id":"!chat1:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 1","type":"group","unreadCount":2}
				],
				"hasMore": true,
				"oldestCursor": "c1",
				"newestCursor": "n1"
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cmd := UnreadCmd{Limit: 200}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(testEnvelopeJSONContext(t), &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var env struct {
		Success  bool                       `json:"success"`
		Data     beeperapi.ChatSearchResult `json:"data"`
		Metadata struct {
			Pagination *outfmt.EnvelopePagination `json:"pagination"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}

	if env.Metadata.Pagination == nil {
		t.Fatal("expected pagination metadata")
	}
	if !env.Metadata.Pagination.HasMore {
		t.Fatal("expected has_more=true")
	}
	if env.Metadata.Pagination.AutoPaged || env.Metadata.Pagination.Capped {
		t.Fatalf("unexpected pagination flags: %+v", env.Metadata.Pagination)
	}
}
