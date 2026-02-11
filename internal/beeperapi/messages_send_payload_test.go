package beeperapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessagesSendAttachmentPayload(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/v1/chats/chat-1/messages" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chatID":"chat-1","pendingMessageID":"pending-1"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token", server.URL, 0)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Messages().Send(context.Background(), "chat-1", SendParams{
		Text:             "hello",
		ReplyToMessageID: "msg-1",
		Attachment: &SendAttachmentParams{
			UploadID: "upload-123",
		},
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if resp.ChatID != "chat-1" {
		t.Fatalf("ChatID = %q, want %q", resp.ChatID, "chat-1")
	}
	if resp.PendingMessageID != "pending-1" {
		t.Fatalf("PendingMessageID = %q, want %q", resp.PendingMessageID, "pending-1")
	}

	if captured["text"] != "hello" {
		t.Fatalf("text payload = %#v, want %q", captured["text"], "hello")
	}
	if captured["replyToMessageID"] != "msg-1" {
		t.Fatalf("replyToMessageID payload = %#v, want %q", captured["replyToMessageID"], "msg-1")
	}

	attachment, ok := captured["attachment"].(map[string]any)
	if !ok {
		t.Fatalf("attachment payload type = %T, want object", captured["attachment"])
	}
	if attachment["uploadID"] != "upload-123" {
		t.Fatalf("attachment.uploadID payload = %#v, want %q", attachment["uploadID"], "upload-123")
	}
}
