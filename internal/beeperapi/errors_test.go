package beeperapi

import (
	"net/http"
	"testing"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

func TestIsUnsupportedRoute(t *testing.T) {
	req, err := http.NewRequest(http.MethodPut, "http://localhost:23373/v1/chats/abc/messages/def", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	apiErr := &beeperdesktopapi.Error{
		StatusCode: http.StatusNotFound,
		Request:    req,
	}

	if !IsUnsupportedRoute(apiErr, http.MethodPut, "/messages/") {
		t.Fatal("expected IsUnsupportedRoute to return true")
	}
	if IsUnsupportedRoute(apiErr, http.MethodPost, "/messages/") {
		t.Fatal("expected method mismatch to return false")
	}
	if IsUnsupportedRoute(apiErr, http.MethodPut, "/assets/") {
		t.Fatal("expected path mismatch to return false")
	}
}
