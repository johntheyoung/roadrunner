package errfmt

import (
	"errors"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/config"
)

func TestHint_NoToken(t *testing.T) {
	hint := Hint(config.ErrNoToken)
	if hint == "" {
		t.Fatal("expected hint for missing token")
	}
}

func TestHint_RestrictionErrors(t *testing.T) {
	t.Run("enable_commands", func(t *testing.T) {
		err := NewEnableCommandsError("messages send", []string{"chats", "messages"})
		hint := Hint(err)
		if hint == "" {
			t.Fatal("expected hint for enable-commands restriction")
		}

		var restrictionErr *RestrictionError
		if !errors.As(err, &restrictionErr) {
			t.Fatal("expected RestrictionError")
		}
		if restrictionErr.Kind != RestrictionEnableCommands {
			t.Fatalf("kind = %q, want %q", restrictionErr.Kind, RestrictionEnableCommands)
		}
	})

	t.Run("readonly", func(t *testing.T) {
		err := NewReadonlyError("messages send")
		hint := Hint(err)
		if hint == "" {
			t.Fatal("expected hint for readonly restriction")
		}

		var restrictionErr *RestrictionError
		if !errors.As(err, &restrictionErr) {
			t.Fatal("expected RestrictionError")
		}
		if restrictionErr.Kind != RestrictionReadonly {
			t.Fatalf("kind = %q, want %q", restrictionErr.Kind, RestrictionReadonly)
		}
	})
}

func TestHint_KnownUsagePatterns(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "chat ambiguity",
			err:  errors.New(`multiple chats matched "Alice"`),
		},
		{
			name: "chat no match",
			err:  errors.New(`no chat matched "Alice"`),
		},
		{
			name: "attachment override requires upload id",
			err:  errors.New("attachment overrides require --attachment-upload-id"),
		},
		{
			name: "max-items requires all",
			err:  errors.New("--max-items requires --all"),
		},
		{
			name: "connection error",
			err:  errors.New("dial tcp 127.0.0.1:23373: connect: connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if hint := Hint(tt.err); hint == "" {
				t.Fatalf("expected hint for %q", tt.name)
			}
		})
	}
}
