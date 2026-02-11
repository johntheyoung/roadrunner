package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestMessagesEditRequiresText(t *testing.T) {
	cmd := MessagesEditCmd{
		ChatID:    "!room:beeper.local",
		MessageID: "$msg",
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

func TestMessagesEditInputSourceConflict(t *testing.T) {
	cmd := MessagesEditCmd{
		ChatID:    "!room:beeper.local",
		MessageID: "$msg",
		Text:      "inline",
		TextFile:  "edit.txt",
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

func TestMessagesEditChatTargetConflict(t *testing.T) {
	cmd := MessagesEditCmd{
		ChatID:    "!room:beeper.local",
		Chat:      "Alice",
		MessageID: "$msg",
		Text:      "inline",
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
