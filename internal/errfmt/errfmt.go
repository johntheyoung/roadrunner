package errfmt

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// Format converts an error to a user-friendly message.
func Format(err error) string {
	if err == nil {
		return ""
	}

	// Handle Kong parse errors
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return formatParseError(parseErr)
	}

	// Handle outfmt parse errors (--json/--plain conflict)
	var outfmtErr *outfmt.ParseError
	if errors.As(err, &outfmtErr) {
		return outfmtErr.Msg + "\nRun with --help to see usage"
	}

	// Handle UI parse errors (--color)
	var uiErr *ui.ParseError
	if errors.As(err, &uiErr) {
		return uiErr.Msg + "\nRun with --help to see usage"
	}

	// Handle no token error
	if errors.Is(err, config.ErrNoToken) {
		return "No API token configured.\n\nSet your token:\n  rr auth set <token>\n\nOr use environment variable:\n  export BEEPER_TOKEN=<token>"
	}

	// Handle API errors
	if beeperapi.IsAPIError(err) {
		return beeperapi.FormatError(err)
	}

	return err.Error()
}

// formatParseError enhances Kong parse errors with helpful hints.
func formatParseError(err *kong.ParseError) string {
	msg := err.Error()

	// If Kong already provided a suggestion, use it as-is
	if strings.Contains(msg, "did you mean") {
		return msg
	}

	// For unknown flag errors without suggestions, add a help hint
	if strings.HasPrefix(msg, "unknown flag") {
		return msg + "\nRun with --help to see available flags"
	}

	// For missing required flags
	if strings.Contains(msg, "missing") || strings.Contains(msg, "required") {
		return msg + "\nRun with --help to see usage"
	}

	return msg
}

// ExitError wraps an error with an exit code.
type ExitError struct {
	Err  error
	Code int
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// Exit codes
const (
	ExitSuccess    = 0
	ExitFailure    = 1
	ExitUsageError = 2
)

// WithCode wraps an error with an exit code.
func WithCode(err error, code int) *ExitError {
	return &ExitError{Err: err, Code: code}
}

// UsageError creates an error with exit code 2.
func UsageError(format string, args ...any) *ExitError {
	return &ExitError{
		Err:  fmt.Errorf(format, args...),
		Code: ExitUsageError,
	}
}

// RestrictionKind identifies command safety restriction types.
type RestrictionKind string

const (
	// RestrictionEnableCommands indicates command allowlist blocking.
	RestrictionEnableCommands RestrictionKind = "enable_commands"
	// RestrictionReadonly indicates readonly mode blocking.
	RestrictionReadonly RestrictionKind = "readonly"
)

// RestrictionError reports command safety restriction violations.
type RestrictionError struct {
	Kind      RestrictionKind
	Command   string
	Allowlist []string
}

func (e *RestrictionError) Error() string {
	switch e.Kind {
	case RestrictionEnableCommands:
		return fmt.Sprintf("command %q not in --enable-commands allowlist: %v", e.Command, e.Allowlist)
	case RestrictionReadonly:
		return fmt.Sprintf("command %q blocked by --readonly mode", e.Command)
	default:
		return "command restricted"
	}
}

// NewEnableCommandsError creates an allowlist restriction error.
func NewEnableCommandsError(command string, allowlist []string) error {
	copied := slices.Clone(allowlist)
	return &RestrictionError{
		Kind:      RestrictionEnableCommands,
		Command:   command,
		Allowlist: copied,
	}
}

// NewReadonlyError creates a readonly restriction error.
func NewReadonlyError(command string) error {
	return &RestrictionError{
		Kind:    RestrictionReadonly,
		Command: command,
	}
}

// Error codes for envelope responses
const (
	ErrCodeAuth       = "AUTH_ERROR"
	ErrCodeNotFound   = "NOT_FOUND"
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeConnection = "CONNECTION_ERROR"
	ErrCodeInternal   = "INTERNAL_ERROR"
)

// ErrorCode maps an error to an error code string.
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}

	// Check for no token error
	if errors.Is(err, config.ErrNoToken) {
		return ErrCodeAuth
	}

	// Check for API errors
	if beeperapi.IsUnauthorized(err) {
		return ErrCodeAuth
	}
	if beeperapi.IsNotFound(err) {
		return ErrCodeNotFound
	}
	if beeperapi.IsAPIError(err) {
		// Could check for rate limiting, but SDK doesn't expose status code directly
		// For now, map unknown API errors to internal error
		return ErrCodeInternal
	}

	// Check for usage/validation errors
	var exitErr *ExitError
	if errors.As(err, &exitErr) && exitErr.Code == ExitUsageError {
		return ErrCodeValidation
	}

	// Check for Kong parse errors
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return ErrCodeValidation
	}

	// Check for outfmt parse errors
	var outfmtErr *outfmt.ParseError
	if errors.As(err, &outfmtErr) {
		return ErrCodeValidation
	}

	// Check for connection errors
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "network is unreachable") {
		return ErrCodeConnection
	}

	return ErrCodeInternal
}

// Hint returns an optional actionable hint for known error patterns.
func Hint(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, config.ErrNoToken) {
		return "Set a token with `rr auth set <token>` or export `BEEPER_TOKEN`."
	}

	var restrictionErr *RestrictionError
	if errors.As(err, &restrictionErr) {
		switch restrictionErr.Kind {
		case RestrictionEnableCommands:
			return "Include this command in `--enable-commands` (or use top-level allowlisting like `--enable-commands=messages,chats`)."
		case RestrictionReadonly:
			return "Remove `--readonly` (or `--agent`) for write operations, or use a read-only command."
		}
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "agent mode requires --enable-commands"):
		return "Pass `--enable-commands` with an explicit allowlist, e.g. `--enable-commands=chats,messages,status`."
	case strings.Contains(msg, "multiple chats matched"):
		return "Use `rr chats resolve <query> --json` or pass an explicit chat ID to disambiguate."
	case strings.Contains(msg, "no chat matched"):
		return "Try `rr chats search <query> --scope=participants --json` to discover the chat ID."
	case strings.Contains(msg, "attachment overrides require --attachment-upload-id"):
		return "Upload first via `rr assets upload <path> --json`, then pass `--attachment-upload-id`."
	case strings.Contains(msg, "message text or --attachment-upload-id is required"):
		return "Provide message text or an uploaded attachment ID."
	case strings.Contains(msg, "requires --all"):
		return "Add `--all` when using this max-items flag."
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "no such host"),
		strings.Contains(msg, "network is unreachable"):
		return "Run `rr doctor` to verify Desktop API connectivity and token validity."
	}

	return ""
}
