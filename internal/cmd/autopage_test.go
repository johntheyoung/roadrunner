package cmd

import (
	"errors"
	"testing"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func TestResolveAutoPageLimit(t *testing.T) {
	tests := []struct {
		name     string
		all      bool
		maxItems int
		want     int
		wantErr  bool
	}{
		{name: "disabled no max", all: false, maxItems: 0, want: 0},
		{name: "disabled with max", all: false, maxItems: 10, wantErr: true},
		{name: "enabled default", all: true, maxItems: 0, want: defaultAutoPageMaxItems},
		{name: "enabled explicit", all: true, maxItems: 250, want: 250},
		{name: "enabled invalid low", all: true, maxItems: -1, wantErr: true},
		{name: "enabled invalid high", all: true, maxItems: maxAutoPageItems + 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAutoPageLimit(tt.all, tt.maxItems)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveAutoPageLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				var exitErr *errfmt.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("error = %T, want *errfmt.ExitError", err)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("resolveAutoPageLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNextSearchCursor(t *testing.T) {
	if got := nextSearchCursor("", "old", "new"); got != "old" {
		t.Fatalf("before cursor = %q, want %q", got, "old")
	}
	if got := nextSearchCursor("after", "old", "new"); got != "new" {
		t.Fatalf("after cursor = %q, want %q", got, "new")
	}
	if got := nextSearchCursor("after", "old", ""); got != "old" {
		t.Fatalf("after fallback cursor = %q, want %q", got, "old")
	}
}

func TestValidateMaxParticipantCount(t *testing.T) {
	valid := []int{-1, 0, 42, 500}
	for _, value := range valid {
		if err := validateMaxParticipantCount(value); err != nil {
			t.Fatalf("validateMaxParticipantCount(%d) unexpected error: %v", value, err)
		}
	}

	invalid := []int{-2, 501}
	for _, value := range invalid {
		err := validateMaxParticipantCount(value)
		if err == nil {
			t.Fatalf("validateMaxParticipantCount(%d) expected error", value)
		}
		var exitErr *errfmt.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("error = %T, want *errfmt.ExitError", err)
		}
	}
}

func TestValidateAttachmentSize(t *testing.T) {
	width := 100.0
	height := 50.0
	if err := validateAttachmentSize(&width, &height); err != nil {
		t.Fatalf("validateAttachmentSize(valid) unexpected error: %v", err)
	}
	if err := validateAttachmentSize(nil, nil); err != nil {
		t.Fatalf("validateAttachmentSize(nil,nil) unexpected error: %v", err)
	}

	if err := validateAttachmentSize(&width, nil); err == nil {
		t.Fatal("validateAttachmentSize(width-only) expected error")
	}

	zero := 0.0
	if err := validateAttachmentSize(&zero, &height); err == nil {
		t.Fatal("validateAttachmentSize(zero-width) expected error")
	}
}
