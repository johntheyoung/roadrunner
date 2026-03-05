package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestChatsList_JSONL(t *testing.T) {
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
				{"id":"!chat1:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 1","type":"group","unreadCount":0},
				{"id":"!chat2:beeper.local","accountID":"acc1","participants":{"hasMore":false,"items":[],"total":0},"title":"Chat 2","type":"group","unreadCount":0}
			],
			"hasMore": false,
			"oldestCursor": "",
			"newestCursor": ""
		}`))
	}))
	defer server.Close()

	withArgs(t, []string{"rr", "--jsonl", "--base-url", server.URL, "--timeout", "5", "chats", "list"}, func() {
		var code int
		out, errText := captureOutput(t, func() {
			code = Execute()
		})
		if code != errfmt.ExitSuccess {
			t.Fatalf("exit code = %d, want %d\nstderr: %s", code, errfmt.ExitSuccess, errText)
		}
		if strings.TrimSpace(errText) != "" {
			t.Fatalf("stderr not empty: %q", errText)
		}

		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != 2 {
			t.Fatalf("jsonl lines = %d, want 2\noutput: %s", len(lines), out)
		}

		for i, line := range lines {
			var item map[string]any
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				t.Fatalf("line %d invalid json: %v\nline: %s", i, err, line)
			}
			if item["id"] == "" {
				t.Fatalf("line %d missing id: %s", i, line)
			}
		}
	})
}
