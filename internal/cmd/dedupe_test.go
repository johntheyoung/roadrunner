package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
)

func TestCheckAndRememberNonIdempotentDuplicate_BlocksWithinWindow(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	ctx := outfmt.WithRequestID(context.Background(), "req-1")
	flags := &RootFlags{DedupeWindow: 10 * time.Minute}
	payload := struct {
		Value string `json:"value"`
	}{Value: "hello"}

	if err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "messages send", payload); err != nil {
		t.Fatalf("first checkAndRememberNonIdempotentDuplicate() error = %v", err)
	}

	err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "messages send", payload)
	if err == nil {
		t.Fatal("expected duplicate-blocking error")
	}

	var exitErr *errfmt.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T, want *errfmt.ExitError", err)
	}
	if exitErr.Code != errfmt.ExitUsageError {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, errfmt.ExitUsageError)
	}
}

func TestCheckAndRememberNonIdempotentDuplicate_ForceBypass(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	ctx := outfmt.WithRequestID(context.Background(), "req-2")
	payload := struct {
		Value string `json:"value"`
	}{Value: "hello"}

	flags := &RootFlags{DedupeWindow: 10 * time.Minute}
	if err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "messages send", payload); err != nil {
		t.Fatalf("first checkAndRememberNonIdempotentDuplicate() error = %v", err)
	}

	forceFlags := &RootFlags{DedupeWindow: 10 * time.Minute, Force: true}
	if err := checkAndRememberNonIdempotentDuplicate(ctx, forceFlags, "messages send", payload); err != nil {
		t.Fatalf("force checkAndRememberNonIdempotentDuplicate() error = %v", err)
	}
}

func TestCheckAndRememberNonIdempotentDuplicate_SkipsWithoutRequestID(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	ctx := context.Background()
	flags := &RootFlags{DedupeWindow: 10 * time.Minute}
	payload := struct {
		Value string `json:"value"`
	}{Value: "hello"}

	if err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "messages send", payload); err != nil {
		t.Fatalf("checkAndRememberNonIdempotentDuplicate() error = %v", err)
	}
	if err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "messages send", payload); err != nil {
		t.Fatalf("second checkAndRememberNonIdempotentDuplicate() error = %v", err)
	}
}
