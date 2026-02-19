package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
)

func TestConnectInfoCommandJSON(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/info" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"Beeper Desktop API",
			"version":"4.2.999",
			"runtime":"desktop",
			"endpoints":{"oauth_introspect":"http://localhost:23373/oauth/introspect"}
		}`))
	}))
	defer server.Close()

	cmd := ConnectInfoCmd{}
	out, _ := captureOutput(t, func() {
		if err := cmd.Run(testJSONContext(t), &RootFlags{BaseURL: server.URL, Timeout: 5}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	var resp beeperapi.ConnectInfo
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}
	if resp.Name != "Beeper Desktop API" {
		t.Fatalf("name = %q, want %q", resp.Name, "Beeper Desktop API")
	}
	if resp.Version != "4.2.999" {
		t.Fatalf("version = %q, want %q", resp.Version, "4.2.999")
	}
	if resp.Endpoints["oauth_introspect"] == "" {
		t.Fatalf("oauth introspect endpoint missing: %+v", resp.Endpoints)
	}
}
