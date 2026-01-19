package cmd

import (
	"context"
	"os"
	"text/tabwriter"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// AccountsCmd is the parent command for account subcommands.
type AccountsCmd struct {
	List  AccountsListCmd  `cmd:"" help:"List connected messaging accounts"`
	Alias AccountsAliasCmd `cmd:"" help:"Manage account aliases"`
}

// AccountsListCmd lists all connected accounts.
type AccountsListCmd struct {
	Fields      []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// AccountsListResponse is the JSON output structure.
type AccountsListResponse struct {
	Accounts []beeperapi.Account `json:"accounts"`
}

// Run executes the accounts list command.
func (c *AccountsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	// Get token
	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	// Create client
	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	// Fetch accounts
	accounts, err := client.Accounts().List(ctx)
	if err != nil {
		return err
	}

	if err := failIfEmpty(c.FailIfEmpty, len(accounts), "accounts"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, AccountsListResponse{
			Accounts: accounts,
		}, "accounts list")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "network", "display_name"})
		if err != nil {
			return err
		}
		for _, a := range accounts {
			writePlainFields(u, fields, map[string]string{
				"id":           a.ID,
				"network":      a.Network,
				"display_name": a.DisplayName,
			})
		}
		return nil
	}

	// Human-readable output
	if len(accounts) == 0 {
		u.Out().Warn("No accounts connected")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	u.Out().Printf("Connected accounts (%d):\n", len(accounts))
	for _, a := range accounts {
		_, _ = w.Write([]byte("  " + a.Network + "\t" + a.DisplayName + "\t" + a.ID + "\n"))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

// AccountsAliasCmd is the parent command for alias subcommands.
type AccountsAliasCmd struct {
	Set   AccountsAliasSetCmd   `cmd:"" help:"Create or update an account alias"`
	List  AccountsAliasListCmd  `cmd:"" help:"List account aliases"`
	Unset AccountsAliasUnsetCmd `cmd:"" help:"Remove an account alias"`
}

// AccountsAliasSetCmd creates or updates an account alias.
type AccountsAliasSetCmd struct {
	Alias     string `arg:"" help:"Alias name (e.g., 'work', 'personal')"`
	AccountID string `arg:"" help:"Account ID to map the alias to"`
}

// Run executes the accounts alias set command.
func (c *AccountsAliasSetCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	if err := config.SetAccountAlias(c.Alias, c.AccountID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"success":    true,
			"alias":      c.Alias,
			"account_id": c.AccountID,
		}, "accounts alias set")
	}

	u.Out().Success("Alias saved: " + c.Alias + " -> " + c.AccountID)
	return nil
}

// AccountsAliasListCmd lists all account aliases.
type AccountsAliasListCmd struct{}

// Run executes the accounts alias list command.
func (c *AccountsAliasListCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	aliases, err := config.GetAccountAliases()
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"aliases": aliases,
		}, "accounts alias list")
	}

	if outfmt.IsPlain(ctx) {
		for alias, accountID := range aliases {
			u.Out().Printf("%s\t%s", alias, accountID)
		}
		return nil
	}

	if len(aliases) == 0 {
		u.Out().Warn("No account aliases configured")
		u.Out().Dim("Use: rr accounts alias set <alias> <account-id>")
		return nil
	}

	u.Out().Printf("Account aliases:")
	for alias, accountID := range aliases {
		u.Out().Printf("  %s -> %s", alias, accountID)
	}

	return nil
}

// AccountsAliasUnsetCmd removes an account alias.
type AccountsAliasUnsetCmd struct {
	Alias string `arg:"" help:"Alias name to remove"`
}

// Run executes the accounts alias unset command.
func (c *AccountsAliasUnsetCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	if err := config.UnsetAccountAlias(c.Alias); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"success": true,
			"alias":   c.Alias,
		}, "accounts alias unset")
	}

	u.Out().Success("Alias removed: " + c.Alias)
	return nil
}
