package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// ContactsCmd is the parent command for contacts subcommands.
type ContactsCmd struct {
	Search  ContactsSearchCmd  `cmd:"" help:"Search contacts on an account"`
	Resolve ContactsResolveCmd `cmd:"" help:"Resolve a contact by exact match"`
}

// ContactsSearchCmd searches contacts within an account.
type ContactsSearchCmd struct {
	AccountID   string   `arg:"" name:"accountID" help:"Account ID to search"`
	Query       string   `arg:"" help:"Search query"`
	Fields      []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// ContactsResolveCmd resolves a contact by exact match.
type ContactsResolveCmd struct {
	AccountID string   `arg:"" name:"accountID" help:"Account ID to search"`
	Query     string   `arg:"" help:"Exact contact name, username, email, phone, or ID"`
	Fields    []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

// Run executes the contacts search command.
func (c *ContactsSearchCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.AccountID == "" || c.Query == "" {
		return errfmt.UsageError("account ID and query are required")
	}

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	resp, err := client.Accounts().SearchContacts(ctx, c.AccountID, c.Query)
	if err != nil {
		return err
	}

	if err := failIfEmpty(c.FailIfEmpty, len(resp), "contacts"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, resp)
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "full_name", "username", "phone_number", "cannot_message"})
		if err != nil {
			return err
		}
		for _, item := range resp {
			writePlainFields(u, fields, map[string]string{
				"id":             item.ID,
				"full_name":      item.FullName,
				"username":       item.Username,
				"phone_number":   item.PhoneNumber,
				"cannot_message": formatBool(item.CannotMessage),
			})
		}
		return nil
	}

	// Human-readable output
	if len(resp) == 0 {
		u.Out().Warn("No contacts found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	u.Out().Printf("Contacts (%d):\n", len(resp))
	for _, item := range resp {
		name := item.FullName
		if name == "" {
			name = item.Username
		}
		if name == "" {
			name = item.ID
		}
		status := ""
		if item.CannotMessage {
			status = "cannot-message"
		}
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", name, item.ID, item.Username, status); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

// Run executes the contacts resolve command.
func (c *ContactsResolveCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.AccountID == "" || c.Query == "" {
		return errfmt.UsageError("account ID and query are required")
	}

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	resp, err := client.Accounts().SearchContacts(ctx, c.AccountID, c.Query)
	if err != nil {
		return err
	}

	matches := make([]beeperapi.Contact, 0, 1)
	for _, item := range resp {
		if contactExactMatch(item, c.Query) {
			matches = append(matches, item)
		}
	}

	if len(matches) == 0 {
		return errfmt.WithCode(fmt.Errorf("no contact matched %q", c.Query), errfmt.ExitFailure)
	}
	if len(matches) > 1 {
		return errfmt.WithCode(fmt.Errorf("multiple contacts matched %q", c.Query), errfmt.ExitFailure)
	}
	contact := matches[0]

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, contact)
	}

	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "full_name", "username", "phone_number", "email", "cannot_message"})
		if err != nil {
			return err
		}
		writePlainFields(u, fields, map[string]string{
			"id":             contact.ID,
			"full_name":      contact.FullName,
			"username":       contact.Username,
			"phone_number":   contact.PhoneNumber,
			"email":          contact.Email,
			"cannot_message": formatBool(contact.CannotMessage),
		})
		return nil
	}

	name := contact.FullName
	if name == "" {
		name = contact.Username
	}
	if name == "" {
		name = contact.ID
	}
	u.Out().Printf("Contact: %s", name)
	u.Out().Printf("ID:      %s", contact.ID)
	if contact.Username != "" {
		u.Out().Printf("Username: %s", contact.Username)
	}
	if contact.PhoneNumber != "" {
		u.Out().Printf("Phone:   %s", contact.PhoneNumber)
	}
	if contact.Email != "" {
		u.Out().Printf("Email:   %s", contact.Email)
	}
	if contact.CannotMessage {
		u.Out().Warn("Cannot message this contact")
	}

	return nil
}

func contactExactMatch(contact beeperapi.Contact, query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return false
	}
	if strings.EqualFold(contact.ID, q) {
		return true
	}
	if strings.EqualFold(contact.FullName, q) {
		return true
	}
	if strings.EqualFold(contact.Username, q) {
		return true
	}
	if strings.EqualFold(contact.Email, q) {
		return true
	}
	if strings.EqualFold(contact.PhoneNumber, q) {
		return true
	}
	return false
}
