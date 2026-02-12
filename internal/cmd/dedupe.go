package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
)

type dedupeEntry struct {
	Command     string `json:"command"`
	RequestID   string `json:"request_id"`
	PayloadHash string `json:"payload_hash"`
	SeenUnix    int64  `json:"seen_unix"`
}

type dedupeLedger struct {
	Entries []dedupeEntry `json:"entries"`
}

func dedupeFilePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", fmt.Errorf("get config dir: %w", err)
	}
	return filepath.Join(dir, "dedupe.json"), nil
}

func loadDedupeLedger(path string) (dedupeLedger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return dedupeLedger{}, nil
		}
		return dedupeLedger{}, err
	}
	if len(data) == 0 {
		return dedupeLedger{}, nil
	}
	var ledger dedupeLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return dedupeLedger{}, err
	}
	return ledger, nil
}

func saveDedupeLedger(path string, ledger dedupeLedger) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func payloadHash(payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func checkAndRememberNonIdempotentDuplicate(ctx context.Context, flags *RootFlags, command string, payload any) error {
	if flags.DedupeWindow <= 0 {
		return nil
	}

	requestID := strings.TrimSpace(outfmt.RequestIDFromContext(ctx))
	if requestID == "" {
		return nil
	}

	hash, err := payloadHash(payload)
	if err != nil {
		return fmt.Errorf("encode dedupe payload: %w", err)
	}

	path, err := dedupeFilePath()
	if err != nil {
		return err
	}

	ledger, err := loadDedupeLedger(path)
	if err != nil {
		return fmt.Errorf("load dedupe ledger: %w", err)
	}

	now := time.Now().UTC()
	cutoff := now.Add(-flags.DedupeWindow).Unix()
	filtered := make([]dedupeEntry, 0, len(ledger.Entries)+1)

	var matched *dedupeEntry
	for i := range ledger.Entries {
		entry := ledger.Entries[i]
		if entry.SeenUnix < cutoff {
			continue
		}
		if entry.Command == command && entry.RequestID == requestID && entry.PayloadHash == hash {
			e := entry
			matched = &e
		}
		filtered = append(filtered, entry)
	}

	if matched != nil && !flags.Force {
		seenAt := time.Unix(matched.SeenUnix, 0).UTC().Format(time.RFC3339)
		return errfmt.UsageError("duplicate non-idempotent request blocked (command=%q request_id=%q seen_at=%s window=%s); rerun with --force to bypass", command, requestID, seenAt, flags.DedupeWindow)
	}

	filtered = append(filtered, dedupeEntry{
		Command:     command,
		RequestID:   requestID,
		PayloadHash: hash,
		SeenUnix:    now.Unix(),
	})

	if err := saveDedupeLedger(path, dedupeLedger{Entries: filtered}); err != nil {
		return fmt.Errorf("save dedupe ledger: %w", err)
	}

	return nil
}
