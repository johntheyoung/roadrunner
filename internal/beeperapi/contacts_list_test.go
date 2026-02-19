package beeperapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListContacts(t *testing.T) {
	t.Parallel()

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/accounts/acc1/contacts/list" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		if got := r.URL.Query().Get("cursor"); got != "c1" {
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		if got := r.URL.Query().Get("direction"); got != "before" {
			http.Error(w, "bad direction", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items":[
				{
					"id":"u1",
					"fullName":"Alice Example",
					"username":"alice",
					"email":"alice@example.com",
					"phoneNumber":"+155555501",
					"cannotMessage":false,
					"imgURL":"https://example.com/a.png"
				}
			],
			"hasMore": true,
			"oldestCursor": "c2",
			"newestCursor": "n1"
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Accounts().ListContacts(context.Background(), "acc1", ContactListParams{
		Cursor:    "c1",
		Direction: "before",
	})
	if err != nil {
		t.Fatalf("ListContacts() error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	if !resp.HasMore {
		t.Fatal("has_more = false, want true")
	}
	if resp.OldestCursor != "c2" {
		t.Fatalf("oldest_cursor = %q, want %q", resp.OldestCursor, "c2")
	}
	if resp.NewestCursor != "n1" {
		t.Fatalf("newest_cursor = %q, want %q", resp.NewestCursor, "n1")
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].ID != "u1" {
		t.Fatalf("item.id = %q, want %q", resp.Items[0].ID, "u1")
	}
	if resp.Items[0].FullName != "Alice Example" {
		t.Fatalf("item.full_name = %q, want %q", resp.Items[0].FullName, "Alice Example")
	}
}
