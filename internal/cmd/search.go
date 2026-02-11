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

// SearchCmd performs a global search.
type SearchCmd struct {
	Query             string   `arg:"" help:"Search query (literal word match)"`
	MessagesCursor    string   `help:"Cursor for message results pagination" name:"messages-cursor"`
	MessagesDirection string   `help:"Pagination direction for message results: before|after" name:"messages-direction" enum:"before,after," default:""`
	MessagesLimit     int      `help:"Max messages per page when paging (1-20)" name:"messages-limit" default:"0"`
	MessagesAll       bool     `help:"Fetch all message pages automatically" name:"messages-all"`
	MessagesMaxItems  int      `help:"Maximum message items to collect with --messages-all (default 500, max 5000)" name:"messages-max-items" default:"0"`
	FailIfEmpty       bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
	Fields            []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

// Run executes the search command.
func (c *SearchCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	if c.MessagesLimit < 0 || c.MessagesLimit > 20 {
		return errfmt.UsageError("invalid --messages-limit %d (expected 0-20)", c.MessagesLimit)
	}
	autoPageLimit, err := resolveAutoPageLimitNamed(c.MessagesAll, c.MessagesMaxItems, "--messages-all", "--messages-max-items")
	if err != nil {
		return err
	}
	effectiveMessagesLimit := c.MessagesLimit
	if c.MessagesAll && effectiveMessagesLimit == 0 {
		effectiveMessagesLimit = 20
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

	resp, err := client.Search(ctx, beeperapi.SearchParams{
		Query:             c.Query,
		MessagesCursor:    c.MessagesCursor,
		MessagesDirection: c.MessagesDirection,
		MessagesLimit:     effectiveMessagesLimit,
	})
	if err != nil {
		return err
	}
	capped := false
	if c.MessagesAll {
		items := make([]beeperapi.MessageItem, 0, len(resp.Messages.Items))
		items = append(items, resp.Messages.Items...)
		lastCursor := c.MessagesCursor
		for resp.Messages.HasMore {
			if limitReached(len(items), autoPageLimit) {
				capped = true
				break
			}

			nextCursor := nextSearchCursor(c.MessagesDirection, resp.Messages.OldestCursor, resp.Messages.NewestCursor)
			if nextCursor == "" || nextCursor == lastCursor {
				break
			}
			lastCursor = nextCursor

			page, err := client.Messages().Search(ctx, beeperapi.MessageSearchParams{
				Query:     c.Query,
				Cursor:    nextCursor,
				Direction: c.MessagesDirection,
				Limit:     effectiveMessagesLimit,
			})
			if err != nil {
				return err
			}

			items = append(items, page.Items...)
			resp.Messages.HasMore = page.HasMore
			resp.Messages.OldestCursor = page.OldestCursor
			resp.Messages.NewestCursor = page.NewestCursor
		}
		if limitReached(len(items), autoPageLimit) {
			if len(items) > autoPageLimit {
				items = items[:autoPageLimit]
			}
			capped = true
		}
		resp.Messages.Items = items
		if capped {
			resp.Messages.HasMore = true
		}
	}

	resultCount := len(resp.Chats) + len(resp.InGroups) + len(resp.Messages.Items)
	if err := failIfEmpty(c.FailIfEmpty, resultCount, "results"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "search")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"type", "id", "chat_id", "title", "text"})
		if err != nil {
			return err
		}
		for _, chat := range resp.Chats {
			writePlainFields(u, fields, map[string]string{
				"type":    "chat",
				"id":      chat.ID,
				"chat_id": chat.ID,
				"title":   chat.Title,
				"text":    "",
			})
		}
		for _, chat := range resp.InGroups {
			writePlainFields(u, fields, map[string]string{
				"type":    "group",
				"id":      chat.ID,
				"chat_id": chat.ID,
				"title":   chat.Title,
				"text":    "",
			})
		}
		for _, msg := range resp.Messages.Items {
			writePlainFields(u, fields, map[string]string{
				"type":    "message",
				"id":      msg.ID,
				"chat_id": msg.ChatID,
				"title":   "",
				"text":    ui.Truncate(msg.Text, 50),
			})
		}
		return nil
	}

	// Human-readable output
	hasResults := len(resp.Chats) > 0 || len(resp.InGroups) > 0 || len(resp.Messages.Items) > 0

	if !hasResults {
		u.Out().Warn("No results found")
		u.Out().Dim("Note: Search uses literal word match, not semantic search.")
		return nil
	}

	// Chats
	if len(resp.Chats) > 0 {
		u.Out().Printf("Chats (%d):", len(resp.Chats))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, chat := range resp.Chats {
			title := chat.Title
			if chat.DisplayName != "" {
				title = chat.DisplayName
			}
			title = ui.Truncate(title, 35)
			if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\n", title, chat.Type, chat.ID); err != nil {
				return err
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}
		u.Out().Println("")
	}

	// In groups (participant matches)
	if len(resp.InGroups) > 0 {
		u.Out().Printf("In Groups (%d):", len(resp.InGroups))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, chat := range resp.InGroups {
			title := chat.Title
			if chat.DisplayName != "" {
				title = chat.DisplayName
			}
			title = ui.Truncate(title, 35)
			if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\n", title, chat.Type, chat.ID); err != nil {
				return err
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}
		u.Out().Println("")
	}

	// Messages
	if len(resp.Messages.Items) > 0 {
		u.Out().Printf("Messages (%d):", len(resp.Messages.Items))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, msg := range resp.Messages.Items {
			ts := ""
			if msg.Timestamp != "" {
				if t, err := time.Parse(time.RFC3339, msg.Timestamp); err == nil {
					ts = t.Format("Jan 2")
				}
			}
			text := ui.Truncate(msg.Text, 40)
			if _, err := fmt.Fprintf(w, "  [%s]\t%s:\t%s\n", ts, msg.SenderName, text); err != nil {
				return err
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}

		if resp.Messages.HasMore {
			if resp.Messages.OldestCursor != "" {
				u.Out().Dim(fmt.Sprintf("\nMore message results available. Use --messages-cursor=%q --messages-direction=before.", resp.Messages.OldestCursor))
			} else {
				u.Out().Dim("\nMore message results available. Use 'rr messages search' for pagination.")
			}
		}
		if capped {
			u.Out().Dim(autoPageStoppedMessageNamed(autoPageLimit, "--messages-max-items"))
		}
	}

	return nil
}
