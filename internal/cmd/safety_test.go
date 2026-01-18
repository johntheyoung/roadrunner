package cmd

import (
	"testing"
)

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple command",
			input: "version",
			want:  "version",
		},
		{
			name:  "command with subcommand",
			input: "messages send",
			want:  "messages send",
		},
		{
			name:  "command with arguments",
			input: "messages send <chatID> <text>",
			want:  "messages send",
		},
		{
			name:  "command with flags",
			input: "messages list --json",
			want:  "messages list",
		},
		{
			name:  "complex command",
			input: "chats archive <chatID> --unarchive",
			want:  "chats archive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCommand(tt.input)
			if got != tt.want {
				t.Errorf("normalizeCommand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheckEnableCommands(t *testing.T) {
	tests := []struct {
		name           string
		enableCommands []string
		command        string
		wantErr        bool
	}{
		{
			name:           "empty allowlist allows all",
			enableCommands: nil,
			command:        "messages send",
			wantErr:        false,
		},
		{
			name:           "exact match allowed",
			enableCommands: []string{"messages send"},
			command:        "messages send",
			wantErr:        false,
		},
		{
			name:           "parent allows subcommands",
			enableCommands: []string{"messages"},
			command:        "messages send",
			wantErr:        false,
		},
		{
			name:           "parent allows all subcommands",
			enableCommands: []string{"messages"},
			command:        "messages list",
			wantErr:        false,
		},
		{
			name:           "command not in allowlist",
			enableCommands: []string{"chats"},
			command:        "messages send",
			wantErr:        true,
		},
		{
			name:           "multiple allowlist entries",
			enableCommands: []string{"chats", "status"},
			command:        "chats list",
			wantErr:        false,
		},
		{
			name:           "multiple allowlist entries - blocked",
			enableCommands: []string{"chats", "status"},
			command:        "messages send",
			wantErr:        true,
		},
		{
			name:           "whitespace in allowlist",
			enableCommands: []string{" messages "},
			command:        "messages send",
			wantErr:        false,
		},
		{
			name:           "empty string in allowlist ignored",
			enableCommands: []string{"", "messages"},
			command:        "messages send",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &RootFlags{EnableCommands: tt.enableCommands}
			err := checkEnableCommands(flags, tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkEnableCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckReadonly(t *testing.T) {
	tests := []struct {
		name     string
		readonly bool
		command  string
		wantErr  bool
	}{
		{
			name:     "readonly disabled allows write",
			readonly: false,
			command:  "messages send",
			wantErr:  false,
		},
		{
			name:     "readonly blocks messages send",
			readonly: true,
			command:  "messages send",
			wantErr:  true,
		},
		{
			name:     "readonly blocks chats create",
			readonly: true,
			command:  "chats create",
			wantErr:  true,
		},
		{
			name:     "readonly blocks chats archive",
			readonly: true,
			command:  "chats archive",
			wantErr:  true,
		},
		{
			name:     "readonly blocks reminders set",
			readonly: true,
			command:  "reminders set",
			wantErr:  true,
		},
		{
			name:     "readonly blocks reminders clear",
			readonly: true,
			command:  "reminders clear",
			wantErr:  true,
		},
		{
			name:     "readonly allows read commands",
			readonly: true,
			command:  "messages list",
			wantErr:  false,
		},
		{
			name:     "readonly allows chats list",
			readonly: true,
			command:  "chats list",
			wantErr:  false,
		},
		{
			name:     "readonly allows auth set (exempt)",
			readonly: true,
			command:  "auth set",
			wantErr:  false,
		},
		{
			name:     "readonly allows auth clear (exempt)",
			readonly: true,
			command:  "auth clear",
			wantErr:  false,
		},
		{
			name:     "readonly allows focus (exempt)",
			readonly: true,
			command:  "focus",
			wantErr:  false,
		},
		{
			name:     "readonly allows version",
			readonly: true,
			command:  "version",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &RootFlags{Readonly: tt.readonly}
			err := checkReadonly(flags, tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkReadonly() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
