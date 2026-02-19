package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
)

func TestContactsListAllAutoPagination(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	requestCount := 0
	cursorValues := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/accounts/acc1/contacts/list" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		cursor := r.URL.Query().Get("cursor")
		cursorValues = append(cursorValues, cursor)
		w.Header().Set("Content-Type", "application/json")
		if cursor == "" {
			_, _ = w.Write([]byte(`{
				"items":[
					{"id":"u1","fullName":"Alice"},
					{"id":"u2","fullName":"Bob"}
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
			"items":[{"id":"u3","fullName":"Carol"}],
			"hasMore": false,
			"oldestCursor": "",
			"newestCursor": ""
		}`))
	}))
	defer server.Close()

	ctx := testJSONContext(t)
	cmd := ContactsListCmd{
		AccountID: "acc1",
		All:       true,
		MaxItems:  10,
	}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(ctx, &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var resp beeperapi.ContactListResult
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
