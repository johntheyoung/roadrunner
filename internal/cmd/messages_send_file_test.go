package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

func TestMessagesSendFileInputSourceConflict(t *testing.T) {
	cmd := MessagesSendFileCmd{
		ChatID:   "!room:beeper.local",
		FilePath: "photo.jpg",
		Text:     "hello",
		TextFile: "message.txt",
	}
	err := cmd.Run(context.Background(), &RootFlags{})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *errfmt.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T, want *errfmt.ExitError", err)
	}
	if exitErr.Code != errfmt.ExitUsageError {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, errfmt.ExitUsageError)
	}
}

func TestMessagesSendFileUploadsThenSendsWithUploadID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")
	fileContents := []byte("fake-jpeg-content")
	if err := os.WriteFile(filePath, fileContents, 0600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	uploadCalled := 0
	sendCalled := 0
	uploadDone := false
	var sendPayload map[string]any
	var handlerErr error
	var handlerErrMu sync.Mutex
	setHandlerErr := func(format string, args ...any) {
		handlerErrMu.Lock()
		defer handlerErrMu.Unlock()
		if handlerErr == nil {
			handlerErr = fmt.Errorf(format, args...)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/assets/upload":
			uploadCalled++
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				setHandlerErr("parse multipart: %v", err)
				http.Error(w, "parse multipart", http.StatusBadRequest)
				return
			}
			if got := r.FormValue("fileName"); got != "renamed.jpg" {
				setHandlerErr("multipart fileName = %q, want %q", got, "renamed.jpg")
				http.Error(w, "invalid fileName", http.StatusBadRequest)
				return
			}
			if got := r.FormValue("mimeType"); got != "image/jpeg" {
				setHandlerErr("multipart mimeType = %q, want %q", got, "image/jpeg")
				http.Error(w, "invalid mimeType", http.StatusBadRequest)
				return
			}
			file, _, err := r.FormFile("file")
			if err != nil {
				setHandlerErr("multipart file: %v", err)
				http.Error(w, "missing file", http.StatusBadRequest)
				return
			}
			defer func() { _ = file.Close() }()
			gotBytes, err := io.ReadAll(file)
			if err != nil {
				setHandlerErr("read multipart file: %v", err)
				http.Error(w, "read file", http.StatusBadRequest)
				return
			}
			if !bytes.Equal(gotBytes, fileContents) {
				setHandlerErr("multipart file content = %q, want %q", string(gotBytes), string(fileContents))
				http.Error(w, "invalid file content", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"uploadID":"up_test_123","srcURL":"file:///tmp/uploaded.jpg","fileName":"renamed.jpg","mimeType":"image/jpeg","fileSize":17}`))
			uploadDone = true
		case r.Method == http.MethodPost && r.URL.Path == "/v1/chats/chat-1/messages":
			sendCalled++
			if !uploadDone {
				setHandlerErr("send was called before upload completed")
				http.Error(w, "upload not completed", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&sendPayload); err != nil {
				setHandlerErr("decode send payload: %v", err)
				http.Error(w, "decode payload", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"chatID":"chat-1","pendingMessageID":"pending-123"}`))
		default:
			setHandlerErr("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	testUI, err := ui.New(ui.Options{
		Stdout: outBuf,
		Stderr: errBuf,
		Color:  "never",
	})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}

	ctx := ui.WithUI(context.Background(), testUI)
	cmd := MessagesSendFileCmd{
		ChatID:   "chat-1",
		FilePath: filePath,
		Text:     "see attachment",
		FileName: "renamed.jpg",
		MimeType: "image/jpeg",
	}
	if err := cmd.Run(ctx, &RootFlags{
		BaseURL: server.URL,
		Timeout: 5,
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if handlerErr != nil {
		t.Fatalf("handler error: %v", handlerErr)
	}

	if uploadCalled != 1 {
		t.Fatalf("upload endpoint calls = %d, want 1", uploadCalled)
	}
	if sendCalled != 1 {
		t.Fatalf("send endpoint calls = %d, want 1", sendCalled)
	}
	if sendPayload["text"] != "see attachment" {
		t.Fatalf("send text payload = %#v, want %q", sendPayload["text"], "see attachment")
	}

	attachment, ok := sendPayload["attachment"].(map[string]any)
	if !ok {
		t.Fatalf("send attachment payload type = %T, want object", sendPayload["attachment"])
	}
	if attachment["uploadID"] != "up_test_123" {
		t.Fatalf("send attachment.uploadID payload = %#v, want %q", attachment["uploadID"], "up_test_123")
	}
}
