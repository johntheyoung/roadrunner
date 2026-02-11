package cmd

import (
	"errors"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestResolveChatTargetInput(t *testing.T) {
	tests := []struct {
		name      string
		chatIDArg string
		chatQuery string
		wantID    string
		wantQuery string
		wantErr   bool
	}{
		{
			name:      "chat id only",
			chatIDArg: "!room:beeper.local",
			wantID:    "!room:beeper.local",
		},
		{
			name:      "chat query only",
			chatQuery: "Alice",
			wantQuery: "Alice",
		},
		{
			name:      "chat id normalized",
			chatIDArg: "\\!room:beeper.local",
			wantID:    "!room:beeper.local",
		},
		{
			name:      "both provided",
			chatIDArg: "!room:beeper.local",
			chatQuery: "Alice",
			wantErr:   true,
		},
		{
			name:    "neither provided",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotQuery, err := resolveChatTargetInput(tt.chatIDArg, tt.chatQuery)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveChatTargetInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				var exitErr *errfmt.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("error = %T, want *errfmt.ExitError", err)
				}
				return
			}
			if gotID != tt.wantID {
				t.Fatalf("chat id = %q, want %q", gotID, tt.wantID)
			}
			if gotQuery != tt.wantQuery {
				t.Fatalf("chat query = %q, want %q", gotQuery, tt.wantQuery)
			}
		})
	}
}
