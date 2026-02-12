package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDirUsesXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	wantDir := filepath.Join(tmp, "beeper")
	if dir != wantDir {
		t.Fatalf("Dir() = %q, want %q", dir, wantDir)
	}

	path, err := FilePath()
	if err != nil {
		t.Fatalf("FilePath() error: %v", err)
	}
	wantPath := filepath.Join(tmp, "beeper", "config.json")
	if path != wantPath {
		t.Fatalf("FilePath() = %q, want %q", path, wantPath)
	}
}

func TestGetTokenPrecedence(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if _, _, err := GetToken(); !errors.Is(err, ErrNoToken) {
		t.Fatalf("GetToken() err = %v, want ErrNoToken", err)
	}

	if err := SetToken("cfg-token"); err != nil {
		t.Fatalf("SetToken error: %v", err)
	}

	token, source, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if token != "cfg-token" || source != TokenSourceConfig {
		t.Fatalf("GetToken() = %q/%v, want cfg-token/config", token, source)
	}

	t.Setenv("BEEPER_ACCESS_TOKEN", "sdk-token")
	token, source, err = GetToken()
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if token != "sdk-token" || source != TokenSourceEnvSDK {
		t.Fatalf("GetToken() = %q/%v, want sdk-token/env-sdk", token, source)
	}

	t.Setenv("BEEPER_TOKEN", "cli-token")
	token, source, err = GetToken()
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if token != "cli-token" || source != TokenSourceEnv {
		t.Fatalf("GetToken() = %q/%v, want cli-token/env", token, source)
	}
}

func TestSavePreservesUnknownFields(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	path, err := FilePath()
	if err != nil {
		t.Fatalf("FilePath() error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Simulate Beeper Desktop (or other tooling) storing unrelated keys here.
	orig := map[string]any{
		"unrelated": "keep-me",
		"nested": map[string]any{
			"x": float64(1),
		},
		"token": "old",
	}
	data, _ := json.MarshalIndent(orig, "", "  ")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	if err := SetToken("new"); err != nil {
		t.Fatalf("SetToken error: %v", err)
	}

	gotBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	got := map[string]any{}
	if err := json.Unmarshal(gotBytes, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["unrelated"] != "keep-me" {
		t.Fatalf("unrelated key was not preserved: %#v", got["unrelated"])
	}
	if nested, ok := got["nested"].(map[string]any); !ok || nested["x"] != float64(1) {
		t.Fatalf("nested key was not preserved: %#v", got["nested"])
	}
	if got["token"] != "new" {
		t.Fatalf("token = %#v, want %q", got["token"], "new")
	}
}
