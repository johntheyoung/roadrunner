package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

func TestAssetsDownloadDestination(t *testing.T) {
	temp := t.TempDir()
	dirPath := filepath.Join(temp, "outdir")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cases := []struct {
		name    string
		srcURL  string
		dest    string
		want    string
		wantErr bool
	}{
		{
			name:   "file-url-to-dir",
			srcURL: "file:///tmp/beeper/hello.png",
			dest:   dirPath,
			want:   filepath.Join(dirPath, "hello.png"),
		},
		{
			name:   "file-url-to-file",
			srcURL: "file:///tmp/beeper/hello.png",
			dest:   filepath.Join(temp, "out.png"),
			want:   filepath.Join(temp, "out.png"),
		},
		{
			name:   "path-to-dir",
			srcURL: "/tmp/beeper/hello.png",
			dest:   dirPath,
			want:   filepath.Join(dirPath, "hello.png"),
		},
		{
			name:   "path-to-file",
			srcURL: "/tmp/beeper/hello.png",
			dest:   filepath.Join(temp, "out.png"),
			want:   filepath.Join(temp, "out.png"),
		},
		{
			name:    "bad-url",
			srcURL:  "file://",
			dest:    temp,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := assetsDownloadDestination(tc.srcURL, tc.dest)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("assetsDownloadDestination error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("assetsDownloadDestination = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAssetLocalPath(t *testing.T) {
	got, err := assetLocalPath("file:///tmp/beeper/hello%20world.png")
	if err != nil {
		t.Fatalf("assetLocalPath error: %v", err)
	}
	want := filepath.FromSlash("/tmp/beeper/hello world.png")
	if got != want {
		t.Fatalf("assetLocalPath = %q, want %q", got, want)
	}
}

func TestAssetsUploadRequiresFilePath(t *testing.T) {
	cmd := AssetsUploadCmd{}
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

func TestAssetsUploadBase64RequiresContent(t *testing.T) {
	cmd := AssetsUploadBase64Cmd{}
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

func TestAssetsUploadBase64SourceConflict(t *testing.T) {
	cmd := AssetsUploadBase64Cmd{
		Content:     "YQ==",
		ContentFile: "payload.b64",
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

func TestAssetsServeRequiresURL(t *testing.T) {
	cmd := AssetsServeCmd{}
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

func TestAssetsServeRejectsDestWithStdout(t *testing.T) {
	cmd := AssetsServeCmd{
		URL:    "mxc://beeper.local/abc",
		Dest:   "out.bin",
		Stdout: true,
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

func TestAssetsServeJSONRequiresDest(t *testing.T) {
	testUI, err := ui.New(ui.Options{Color: "never"})
	if err != nil {
		t.Fatalf("ui.New() error = %v", err)
	}
	ctx := ui.WithUI(context.Background(), testUI)
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})

	cmd := AssetsServeCmd{
		URL: "mxc://beeper.local/abc",
	}
	err = cmd.Run(ctx, &RootFlags{})
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

func TestAssetsServeStreamsToStdoutByDefault(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	const body = "streamed-binary"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/assets/serve" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("url"); got != "mxc://beeper.local/abc" {
			http.Error(w, "bad url query", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	out, _ := captureOutput(t, func() {
		testUI, err := ui.New(ui.Options{Color: "never"})
		if err != nil {
			t.Fatalf("ui.New() error = %v", err)
		}
		ctx := ui.WithUI(context.Background(), testUI)
		cmd := AssetsServeCmd{
			URL: "mxc://beeper.local/abc",
		}
		if err := cmd.Run(ctx, &RootFlags{
			BaseURL: server.URL,
			Timeout: 5,
		}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	if out != body {
		t.Fatalf("stdout = %q, want %q", out, body)
	}
}

func TestAssetsServeWritesDestAndJSONMetadata(t *testing.T) {
	t.Setenv("BEEPER_TOKEN", "test-token")
	t.Setenv("BEEPER_ACCESS_TOKEN", "")

	payload := []byte{0x00, 0x01, 0x02, 0x03}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/assets/serve" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "asset.bin")
	out, _ := captureOutput(t, func() {
		testUI, err := ui.New(ui.Options{Color: "never"})
		if err != nil {
			t.Fatalf("ui.New() error = %v", err)
		}
		ctx := ui.WithUI(context.Background(), testUI)
		ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})
		cmd := AssetsServeCmd{
			URL:  "mxc://beeper.local/abc",
			Dest: dest,
		}
		if err := cmd.Run(ctx, &RootFlags{
			BaseURL: server.URL,
			Timeout: 5,
		}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	gotFile, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest file: %v", err)
	}
	if !bytes.Equal(gotFile, payload) {
		t.Fatalf("dest bytes = %v, want %v", gotFile, payload)
	}

	var envelope map[string]any
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, out)
	}
	if envelope["dest"] != dest {
		t.Fatalf("json dest = %#v, want %q", envelope["dest"], dest)
	}
	if envelope["bytes_written"] != float64(len(payload)) {
		t.Fatalf("json bytes_written = %#v, want %d", envelope["bytes_written"], len(payload))
	}
}
