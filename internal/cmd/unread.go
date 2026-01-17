package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// UnreadCmd lists unread chats across all accounts.
type UnreadCmd struct {
	AccountIDs   []string `help:"Filter by account IDs" name:"account-ids"`
	Inbox        string   `help:"Filter by inbox: primary|low-priority|archive" enum:"primary,low-priority,archive," default:""`
	IncludeMuted *bool    `help:"Include muted chats (default true)" name:"include-muted"`
	Limit        int      `help:"Max results (1-200)" default:"200"`
	Cursor       string   `help:"Pagination cursor"`
	Direction    string   `help:"Pagination direction: before|after" enum:"before,after," default:""`
	Fields       []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty  bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// Run executes the unread command.
func (c *UnreadCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.Limit < 1 || c.Limit > 200 {
		return errfmt.UsageError("invalid --limit %d (expected 1-200)", c.Limit)
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

	resp, err := client.Chats().Search(ctx, beeperapi.ChatSearchParams{
		AccountIDs:   c.AccountIDs,
		Inbox:        c.Inbox,
		UnreadOnly:   true,
		IncludeMuted: c.IncludeMuted,
		Limit:        c.Limit,
		Cursor:       c.Cursor,
		Direction:    c.Direction,
	})
	if err != nil {
		return err
	}

	if err := failIfEmpty(c.FailIfEmpty, len(resp.Items), "unread chats"); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, resp)
	}

	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "title", "display_name", "account_id", "unread_count", "type", "network", "is_archived", "is_muted"})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			writePlainFields(u, fields, map[string]string{
				"id":           item.ID,
				"title":        item.Title,
				"display_name": item.DisplayName,
				"account_id":   item.AccountID,
				"type":         item.Type,
				"network":      item.Network,
				"unread_count": fmt.Sprintf("%d", item.UnreadCount),
				"is_archived":  formatBool(item.IsArchived),
				"is_muted":     formatBool(item.IsMuted),
			})
		}
		return nil
	}

	if len(resp.Items) == 0 {
		u.Out().Warn("No unread chats")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	u.Out().Printf("Unread chats (%d):\n", len(resp.Items))
	for _, item := range resp.Items {
		title := item.Title
		if item.DisplayName != "" {
			title = item.DisplayName
		}
		title = ui.Truncate(title, 35)
		unread := ""
		if item.UnreadCount > 0 {
			unread = fmt.Sprintf("(%d)", item.UnreadCount)
		}
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", title, item.Type, unread, item.ID); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if resp.HasMore && resp.OldestCursor != "" {
		u.Out().Dim(fmt.Sprintf("\nMore chats available. Use --cursor=%q --direction=before", resp.OldestCursor))
	}

	return nil
}
