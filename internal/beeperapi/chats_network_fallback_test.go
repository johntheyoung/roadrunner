package beeperapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatsSearchNetworkFallbackUnknown(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/accounts":
			w.Header().Set("Content-Type", "application/json")
			// Intentionally omit network to reflect latest account schema.
			_, _ = w.Write([]byte(`[{"accountID":"acc1","user":{"id":"u1","fullName":"User"}}]`))
		case "/v1/chats/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"items":[
					{
						"id":"!chat:beeper.local",
						"accountID":"acc1",
						"title":"Chat",
						"type":"group",
						"participants":{"hasMore":false,"items":[],"total":0},
						"unreadCount":0
					}
				],
				"hasMore": false
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Chats().Search(context.Background(), ChatSearchParams{Limit: 20})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].Network != unknownNetwork {
		t.Fatalf("network = %q, want %q", resp.Items[0].Network, unknownNetwork)
	}
}
