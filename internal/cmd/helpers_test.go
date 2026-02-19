package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateTokenUsesIntrospection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/info":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"Beeper Desktop API","version":"4.2.999","runtime":"desktop"}`))
		case "/oauth/introspect":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"active":true,"sub":"user-1"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result := ValidateToken(context.Background(), "test-token", server.URL, 5)
	if !result.Valid {
		t.Fatalf("valid = false, want true (error: %s)", result.Error)
	}
	if result.ValidationMethod != "introspection" {
		t.Fatalf("validation method = %q, want %q", result.ValidationMethod, "introspection")
	}
	if !result.IntrospectionAvailable {
		t.Fatal("introspection_available = false, want true")
	}
	if !result.ConnectInfoAvailable {
		t.Fatal("connect_info_available = false, want true")
	}
	if result.Subject != "user-1" {
		t.Fatalf("subject = %q, want %q", result.Subject, "user-1")
	}
}

func TestValidateTokenFallbackToAccountsWhenIntrospectionUnsupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/info":
			http.NotFound(w, r)
		case "/oauth/introspect":
			http.NotFound(w, r)
		case "/v1/accounts":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"accountID":"acc1","user":{"id":"u1"}}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result := ValidateToken(context.Background(), "test-token", server.URL, 5)
	if !result.Valid {
		t.Fatalf("valid = false, want true (error: %s)", result.Error)
	}
	if result.ValidationMethod != "accounts" {
		t.Fatalf("validation method = %q, want %q", result.ValidationMethod, "accounts")
	}
	if result.Account != "acc1" {
		t.Fatalf("account = %q, want %q", result.Account, "acc1")
	}
}

func TestValidateTokenInactiveViaIntrospection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/info":
			http.NotFound(w, r)
		case "/oauth/introspect":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"active":false}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result := ValidateToken(context.Background(), "test-token", server.URL, 5)
	if result.Valid {
		t.Fatal("valid = true, want false")
	}
	if result.ValidationMethod != "introspection" {
		t.Fatalf("validation method = %q, want %q", result.ValidationMethod, "introspection")
	}
	if result.Error == "" {
		t.Fatal("expected validation error for inactive token")
	}
}
