package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/config"
)

func TestApplyAccountDefault(t *testing.T) {
	tests := []struct {
		name           string
		accountIDs     []string
		defaultAccount string
		wantLen        int
		wantFirst      string
	}{
		{
			name:           "empty with no default",
			accountIDs:     []string{},
			defaultAccount: "",
			wantLen:        0,
			wantFirst:      "",
		},
		{
			name:           "empty with default",
			accountIDs:     []string{},
			defaultAccount: "default-account",
			wantLen:        1,
			wantFirst:      "default-account",
		},
		{
			name:           "provided accounts override default",
			accountIDs:     []string{"account-1", "account-2"},
			defaultAccount: "default-account",
			wantLen:        2,
			wantFirst:      "account-1",
		},
		{
			name:           "nil accounts with default",
			accountIDs:     nil,
			defaultAccount: "default-account",
			wantLen:        1,
			wantFirst:      "default-account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyAccountDefault(tt.accountIDs, tt.defaultAccount)
			if len(result) != tt.wantLen {
				t.Errorf("applyAccountDefault() len = %d, want %d", len(result), tt.wantLen)
			}
			if tt.wantLen > 0 && result[0] != tt.wantFirst {
				t.Errorf("applyAccountDefault() first = %q, want %q", result[0], tt.wantFirst)
			}
		})
	}
}

func TestResolveAccount(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		defaultAccount string
		want           string
	}{
		{
			name:           "provided account",
			accountID:      "my-account",
			defaultAccount: "default-account",
			want:           "my-account",
		},
		{
			name:           "empty with default",
			accountID:      "",
			defaultAccount: "default-account",
			want:           "default-account",
		},
		{
			name:           "empty with no default",
			accountID:      "",
			defaultAccount: "",
			want:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveAccount(tt.accountID, tt.defaultAccount)
			if result != tt.want {
				t.Errorf("resolveAccount() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestAccountAliasResolution(t *testing.T) {
	// Create a temporary config directory
	tmpDir, err := os.MkdirTemp("", "roadrunner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to use the temp directory
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Create the beeper config directory
	configDir := filepath.Join(tmpDir, "beeper")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Test setting and resolving aliases
	if err := config.SetAccountAlias("work", "slack:T12345678"); err != nil {
		t.Fatalf("SetAccountAlias failed: %v", err)
	}
	if err := config.SetAccountAlias("personal", "imessage:+1234567890"); err != nil {
		t.Fatalf("SetAccountAlias failed: %v", err)
	}

	// Test ResolveAccountAlias
	resolved := config.ResolveAccountAlias("work")
	if resolved != "slack:T12345678" {
		t.Errorf("ResolveAccountAlias(work) = %q, want %q", resolved, "slack:T12345678")
	}

	resolved = config.ResolveAccountAlias("personal")
	if resolved != "imessage:+1234567890" {
		t.Errorf("ResolveAccountAlias(personal) = %q, want %q", resolved, "imessage:+1234567890")
	}

	// Unknown alias returns unchanged
	resolved = config.ResolveAccountAlias("unknown")
	if resolved != "unknown" {
		t.Errorf("ResolveAccountAlias(unknown) = %q, want %q", resolved, "unknown")
	}

	// Actual account ID returns unchanged
	resolved = config.ResolveAccountAlias("slack:T12345678")
	if resolved != "slack:T12345678" {
		t.Errorf("ResolveAccountAlias(slack:T12345678) = %q, want %q", resolved, "slack:T12345678")
	}

	// Test GetAccountAliases
	aliases, err := config.GetAccountAliases()
	if err != nil {
		t.Fatalf("GetAccountAliases failed: %v", err)
	}
	if len(aliases) != 2 {
		t.Errorf("GetAccountAliases() len = %d, want 2", len(aliases))
	}

	// Test UnsetAccountAlias
	if err := config.UnsetAccountAlias("work"); err != nil {
		t.Fatalf("UnsetAccountAlias failed: %v", err)
	}
	resolved = config.ResolveAccountAlias("work")
	if resolved != "work" {
		t.Errorf("ResolveAccountAlias(work) after unset = %q, want %q", resolved, "work")
	}
}
