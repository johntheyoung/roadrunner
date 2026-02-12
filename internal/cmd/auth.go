package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// AuthCmd is the parent command for auth subcommands.
type AuthCmd struct {
	Set    AuthSetCmd    `cmd:"" help:"Store API token"`
	Status AuthStatusCmd `cmd:"" help:"Show authentication status"`
	Clear  AuthClearCmd  `cmd:"" help:"Remove stored token"`
}

// AuthSetCmd stores an API token.
type AuthSetCmd struct {
	Token   string `arg:"" help:"API token to store (avoid shell history; prefer --stdin)"`
	Stdin   bool   `help:"Read token from stdin (recommended to avoid shell history)"`
	FromEnv string `help:"Read token from an environment variable (e.g. BEEPER_TOKEN)" name:"from-env" placeholder:"VAR"`
}

// Run executes the auth set command.
func (c *AuthSetCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	sources := 0
	if strings.TrimSpace(c.Token) != "" {
		sources++
	}
	if c.Stdin {
		sources++
	}
	if strings.TrimSpace(c.FromEnv) != "" {
		sources++
	}
	if sources == 0 {
		return errfmt.UsageError("token is required (use `rr auth set --stdin` or `rr auth set --from-env BEEPER_TOKEN`)")
	}
	if sources > 1 {
		return errfmt.UsageError("use only one of <token>, --stdin, --from-env")
	}

	token := c.Token
	if c.Stdin {
		t, err := resolveTextInput("", "", true, true, "token", "", "--stdin")
		if err != nil {
			return err
		}
		token = t
	} else if strings.TrimSpace(c.FromEnv) != "" {
		val := os.Getenv(strings.TrimSpace(c.FromEnv))
		if strings.TrimSpace(val) == "" {
			return errfmt.UsageError("environment variable %q is empty", strings.TrimSpace(c.FromEnv))
		}
		token = val
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return errfmt.UsageError("token is required")
	}

	if err := config.SetToken(token); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"success": true,
			"message": "Token saved",
		}, "auth set")
	}

	u.Out().Success("Token saved to config file")
	return nil
}

// AuthStatusCmd shows authentication status.
type AuthStatusCmd struct {
	Check  bool     `help:"Validate token by making API call" short:"c"`
	Fields []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

// Run executes the auth status command.
func (c *AuthStatusCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	token, source, err := config.GetToken()
	if err != nil {
		if outfmt.IsJSON(ctx) {
			return writeJSON(ctx, map[string]any{
				"authenticated": false,
				"source":        "none",
				"error":         err.Error(),
			}, "auth status")
		}
		if outfmt.IsPlain(ctx) {
			fields, err := resolveFields(c.Fields, []string{"authenticated", "source", "config_path", "valid"})
			if err != nil {
				return err
			}
			writePlainFields(u, fields, map[string]string{
				"authenticated": "false",
				"source":        "none",
				"config_path":   "",
				"valid":         "",
			})
			return nil
		}
		u.Out().Warn("Not authenticated")
		u.Out().Dim("Run: rr auth set --stdin  # recommended (avoids shell history)")
		u.Out().Dim("Or:  rr auth set --from-env BEEPER_TOKEN")
		return nil
	}

	configPath, _ := config.FilePath()

	// Validate token if requested
	var validation *ValidateTokenResult
	if c.Check {
		result := ValidateToken(ctx, token, flags.BaseURL, flags.Timeout)
		validation = &result
	}

	if outfmt.IsJSON(ctx) {
		result := map[string]any{
			"authenticated": true,
			"source":        source.String(),
			"config_path":   configPath,
		}

		// Mask token for security
		if len(token) > 8 {
			result["token_preview"] = token[:4] + "..." + token[len(token)-4:]
		}

		if validation != nil {
			result["valid"] = validation.Valid
			if validation.Account != "" {
				result["account"] = validation.Account
			}
			if validation.Error != "" {
				result["validation_error"] = validation.Error
			}
		}

		return writeJSON(ctx, result, "auth status")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"authenticated", "source", "config_path", "valid"})
		if err != nil {
			return err
		}
		valid := ""
		if validation != nil {
			valid = formatBool(validation.Valid)
		}
		writePlainFields(u, fields, map[string]string{
			"authenticated": "true",
			"source":        source.String(),
			"config_path":   configPath,
			"valid":         valid,
		})
		return nil
	}

	u.Out().Printf("Authenticated: yes")
	u.Out().Printf("Token source:  %s", source)
	u.Out().Printf("Config path:   %s", configPath)

	// Show masked token preview
	if len(token) > 8 {
		u.Out().Printf("Token:         %s...%s", token[:4], token[len(token)-4:])
	}

	if validation != nil {
		if validation.Valid {
			msg := "Token valid:   yes"
			if validation.Account != "" {
				msg += " (account: " + validation.Account + ")"
			}
			u.Out().Success(msg)
		} else {
			u.Out().Error("Token valid:   no")
			if validation.Error != "" {
				u.Out().Dim("  " + validation.Error)
			}
		}
	}

	return nil
}

// AuthClearCmd removes the stored token.
type AuthClearCmd struct{}

// Run executes the auth clear command.
func (c *AuthClearCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	if err := config.ClearToken(); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"success": true,
			"message": "Token cleared",
		}, "auth clear")
	}

	u.Out().Success("Token cleared from config file")
	return nil
}
