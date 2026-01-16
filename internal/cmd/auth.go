package cmd

import (
	"context"
	"os"

	"github.com/johntheyoung/roadrunner/internal/config"
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
	Token string `arg:"" help:"API token to store"`
}

// Run executes the auth set command.
func (c *AuthSetCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	if err := config.SetToken(c.Token); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"success": true,
			"message": "Token saved",
		})
	}

	u.Out().Success("Token saved to config file")
	return nil
}

// AuthStatusCmd shows authentication status.
type AuthStatusCmd struct {
	Check bool `help:"Validate token by making API call" short:"c"`
}

// Run executes the auth status command.
func (c *AuthStatusCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	token, source, err := config.GetToken()
	if err != nil {
		if outfmt.IsJSON(ctx) {
			return outfmt.WriteJSON(os.Stdout, map[string]any{
				"authenticated": false,
				"source":        "none",
				"error":         err.Error(),
			})
		}
		u.Out().Warn("Not authenticated")
		u.Out().Dim("Run: rr auth set <token>")
		return nil
	}

	configPath, _ := config.FilePath()

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

		if c.Check {
			// TODO: Validate token with API call when SDK adapter is ready
			result["valid"] = "unknown (API check not implemented)"
		}

		return outfmt.WriteJSON(os.Stdout, result)
	}

	u.Out().Printf("Authenticated: yes")
	u.Out().Printf("Token source:  %s", source)
	u.Out().Printf("Config path:   %s", configPath)

	// Show masked token preview
	if len(token) > 8 {
		u.Out().Printf("Token:         %s...%s", token[:4], token[len(token)-4:])
	}

	if c.Check {
		// TODO: Validate token with API call when SDK adapter is ready
		u.Out().Warn("API validation not yet implemented")
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
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"success": true,
			"message": "Token cleared",
		})
	}

	u.Out().Success("Token cleared from config file")
	return nil
}
