package cmd

import (
	"context"
	"errors"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// Build info set at build time via ldflags.
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

func init() {
	// If Version wasn't set by ldflags (goreleaser), try to get it from
	// Go module info embedded by "go install". This ensures users who
	// install via "go install ...@latest" see the correct version.
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			Version = info.Main.Version
		}
	}
}

// RootFlags contains global flags available to all commands.
type RootFlags struct {
	Color          string           `help:"Color output: auto|always|never" default:"auto" env:"BEEPER_COLOR"`
	JSON           bool             `help:"Output JSON to stdout (best for scripting)" env:"BEEPER_JSON"`
	Plain          bool             `help:"Output stable TSV to stdout (no colors)" env:"BEEPER_PLAIN"`
	Verbose        bool             `help:"Enable debug logging" short:"v"`
	NoInput        bool             `help:"Never prompt; fail instead (useful for CI)" env:"BEEPER_NO_INPUT"`
	Force          bool             `help:"Skip confirmations for destructive commands" short:"f"`
	Timeout        int              `help:"Timeout for API calls in seconds (0=none)" default:"30" env:"BEEPER_TIMEOUT"`
	BaseURL        string           `help:"API base URL" default:"http://localhost:23373" env:"BEEPER_URL"`
	Version        kong.VersionFlag `help:"Show version and exit"`
	EnableCommands []string         `help:"Comma-separated allowlist of top-level commands" env:"BEEPER_ENABLE_COMMANDS" sep:","`
	Readonly       bool             `help:"Block data write operations" env:"BEEPER_READONLY"`
	Envelope       bool             `help:"Wrap JSON output in {success,data,error,metadata} envelope" env:"BEEPER_ENVELOPE"`
	Agent          bool             `help:"Agent profile: forces JSON, envelope, no-input, readonly" env:"BEEPER_AGENT"`
	Account        string           `help:"Default account ID for commands" env:"BEEPER_ACCOUNT"`
}

// CLI is the root command structure.
type CLI struct {
	RootFlags

	Auth         AuthCmd         `cmd:"" help:"Manage authentication"`
	Accounts     AccountsCmd     `cmd:"" help:"Manage messaging accounts"`
	Contacts     ContactsCmd     `cmd:"" help:"Search contacts"`
	Assets       AssetsCmd       `cmd:"" help:"Manage assets"`
	Chats        ChatsCmd        `cmd:"" help:"Manage chats"`
	Messages     MessagesCmd     `cmd:"" help:"Manage messages"`
	Reminders    RemindersCmd    `cmd:"" help:"Manage chat reminders"`
	Search       SearchCmd       `cmd:"" help:"Global search across chats and messages"`
	Status       StatusCmd       `cmd:"" help:"Show chat and unread summary"`
	Unread       UnreadCmd       `cmd:"" help:"List unread chats"`
	Focus        FocusCmd        `cmd:"" help:"Focus Beeper Desktop app"`
	Doctor       DoctorCmd       `cmd:"" help:"Diagnose configuration and connectivity"`
	Version      VersionCmd      `cmd:"" help:"Show version information"`
	Capabilities CapabilitiesCmd `cmd:"" help:"Show CLI capabilities for agent discovery"`
	Completion   CompletionCmd   `cmd:"" help:"Generate shell completions"`
}

// Execute runs the CLI and returns an exit code.
func Execute() int {
	cli := &CLI{}

	// Check for expanded help mode
	helpCompact := os.Getenv("BEEPER_HELP") != "full"

	// Pre-parse to get flags for UI setup
	parser, err := kong.New(cli,
		kong.Name("rr"),
		kong.Description("CLI for Beeper Desktop. Beep beep!"),
		kong.UsageOnError(),
		kong.Vars{"version": VersionString()},
		kong.ConfigureHelp(kong.HelpOptions{Compact: helpCompact}),
	)
	if err != nil {
		_, _ = os.Stderr.WriteString("error: " + err.Error() + "\n")
		return errfmt.ExitFailure
	}

	kongCtx, err := parser.Parse(os.Args[1:])
	if err != nil {
		// Handle parse errors with our custom exit codes
		// Kong's FatalIfErrorf calls os.Exit directly, bypassing our handling
		_, _ = os.Stderr.WriteString("error: " + err.Error() + "\n")
		_, _ = os.Stderr.WriteString("Run with --help to see available commands and flags\n")
		return errfmt.ExitUsageError
	}

	// Apply agent mode: force JSON, Envelope, NoInput, Readonly
	if cli.Agent {
		cli.JSON = true
		cli.Envelope = true
		cli.NoInput = true
		cli.Readonly = true
		cli.Plain = false // JSON takes precedence

		// Agent mode requires --enable-commands for safety
		if len(cli.EnableCommands) == 0 {
			command := normalizeCommand(kongCtx.Command())
			_ = outfmt.WriteEnvelopeError(os.Stdout, errfmt.ErrCodeValidation,
				"agent mode requires --enable-commands to specify allowed commands", Version, command)
			return errfmt.ExitUsageError
		}
	}

	// Validate flag combinations
	mode, err := outfmt.FromFlags(cli.JSON, cli.Plain)
	if err != nil {
		// Can't use envelope here - conflicting flags mean envelope state is ambiguous
		_, _ = os.Stderr.WriteString("error: " + errfmt.Format(err) + "\n")
		return errfmt.ExitUsageError
	}

	// Create UI (respects --color and NO_COLOR)
	// Disable colors for JSON/Plain output
	colorMode := cli.Color
	if cli.JSON || cli.Plain {
		colorMode = "never"
	}

	u, err := ui.New(ui.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Color:  colorMode,
	})
	if err != nil {
		_, _ = os.Stderr.WriteString("error: " + errfmt.Format(err) + "\n")
		return errfmt.ExitUsageError
	}

	// Build context with UI and output mode
	ctx := context.Background()
	ctx = ui.WithUI(ctx, u)
	ctx = outfmt.WithMode(ctx, mode)

	// Validate command allowlist and readonly mode
	command := normalizeCommand(kongCtx.Command())
	if err := checkEnableCommands(&cli.RootFlags, command); err != nil {
		if cli.JSON && cli.Envelope {
			_ = outfmt.WriteEnvelopeError(os.Stdout, errfmt.ErrCodeValidation, errfmt.Format(err), Version, command)
		} else {
			_, _ = os.Stderr.WriteString("error: " + errfmt.Format(err) + "\n")
		}
		return errfmt.ExitUsageError
	}
	if err := checkReadonly(&cli.RootFlags, command); err != nil {
		if cli.JSON && cli.Envelope {
			_ = outfmt.WriteEnvelopeError(os.Stdout, errfmt.ErrCodeValidation, errfmt.Format(err), Version, command)
		} else {
			_, _ = os.Stderr.WriteString("error: " + errfmt.Format(err) + "\n")
		}
		return errfmt.ExitUsageError
	}

	// Add envelope context if enabled
	ctx = outfmt.WithEnvelope(ctx, cli.Envelope && cli.JSON)

	// Bind context and flags for command execution
	kongCtx.BindTo(ctx, (*context.Context)(nil))
	kongCtx.Bind(&cli.RootFlags)

	// Run the command
	if err := kongCtx.Run(); err != nil {
		// Handle envelope mode errors to stdout
		if cli.Envelope && cli.JSON {
			code := errfmt.ErrorCode(err)
			_ = outfmt.WriteEnvelopeError(os.Stdout, code, errfmt.Format(err), Version, command)
			var exitErr *errfmt.ExitError
			if errors.As(err, &exitErr) {
				return exitErr.Code
			}
			return errfmt.ExitFailure
		}

		// Check for ExitError with specific code
		var exitErr *errfmt.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Err != nil {
				u.Err().Error("error: " + errfmt.Format(exitErr.Err))
			}
			return exitErr.Code
		}

		// Default error handling
		u.Err().Error("error: " + errfmt.Format(err))
		return errfmt.ExitFailure
	}

	return errfmt.ExitSuccess
}
