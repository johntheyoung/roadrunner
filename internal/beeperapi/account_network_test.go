package beeperapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

func TestAccountNetworkMapNil(t *testing.T) {
	if got := accountNetworkMap(nil); got != nil {
		t.Fatalf("accountNetworkMap(nil) = %#v, want nil", got)
	}
}

func TestAccountNetworkMapFiltersInvalidEntries(t *testing.T) {
	accounts := []beeperdesktopapi.Account{
		{AccountID: "acc-whatsapp", Network: "WhatsApp"},
		{AccountID: "", Network: "Telegram"},
		{AccountID: "acc-empty-network", Network: ""},
		{AccountID: "acc-telegram", Network: "Telegram"},
	}

	got := accountNetworkMap(&accounts)
	if got == nil {
		t.Fatal("accountNetworkMap returned nil, want populated map")
	}

	if got["acc-whatsapp"] != "WhatsApp" {
		t.Fatalf("network for acc-whatsapp = %q, want %q", got["acc-whatsapp"], "WhatsApp")
	}
	if got["acc-telegram"] != "Telegram" {
		t.Fatalf("network for acc-telegram = %q, want %q", got["acc-telegram"], "Telegram")
	}
	if _, ok := got["acc-empty-network"]; ok {
		t.Fatalf("unexpected mapping for acc-empty-network: %q", got["acc-empty-network"])
	}
	if _, ok := got[""]; ok {
		t.Fatalf("unexpected mapping for empty account ID: %q", got[""])
	}
}

func TestAccountNetworksByIDCachedAfterSuccess(t *testing.T) {
	t.Parallel()

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/accounts" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"accountID":"acc1","network":"WhatsApp","user":{"id":"u1"}}]`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	first := client.accountNetworksByID(context.Background())
	second := client.accountNetworksByID(context.Background())

	if requestCount != 1 {
		t.Fatalf("account lookup count = %d, want 1", requestCount)
	}
	if first["acc1"] != "WhatsApp" {
		t.Fatalf("first lookup network = %q, want %q", first["acc1"], "WhatsApp")
	}
	if second["acc1"] != "WhatsApp" {
		t.Fatalf("second lookup network = %q, want %q", second["acc1"], "WhatsApp")
	}
}

func TestAccountNetworksByIDRetriesAfterFailure(t *testing.T) {
	t.Parallel()

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/accounts" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		// SDK retries failed requests by default (max 2 retries),
		// so fail the first call attempt and its retries.
		if requestCount <= 3 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"accountID":"acc2","network":"Telegram","user":{"id":"u2"}}]`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	first := client.accountNetworksByID(context.Background())
	second := client.accountNetworksByID(context.Background())

	if first != nil {
		t.Fatalf("first lookup = %#v, want nil on HTTP failure", first)
	}
	if second["acc2"] != "Telegram" {
		t.Fatalf("second lookup network = %q, want %q", second["acc2"], "Telegram")
	}
	if requestCount != 4 {
		t.Fatalf("account lookup count = %d, want 4", requestCount)
	}
}
