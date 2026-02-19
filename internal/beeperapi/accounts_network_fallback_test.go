package beeperapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccountsListNetworkFallbackUnknown(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/accounts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// network omitted in latest API schema.
		_, _ = w.Write([]byte(`[{"accountID":"acc1","user":{"id":"u1","fullName":"Alice"}}]`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	accounts, err := client.Accounts().List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("len(accounts) = %d, want 1", len(accounts))
	}
	if accounts[0].Network != unknownNetwork {
		t.Fatalf("network = %q, want %q", accounts[0].Network, unknownNetwork)
	}
}
