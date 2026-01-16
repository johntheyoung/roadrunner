package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
	"github.com/beeper/desktop-api-go/option"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// MessagesCmd is the parent command for message subcommands.
type MessagesCmd struct {
	List   MessagesListCmd   `cmd:"" help:"List messages in a chat"`
	Search MessagesSearchCmd `cmd:"" help:"Search messages"`
}

// MessagesListCmd lists messages in a chat.
type MessagesListCmd struct {
	ChatID    string `arg:"" name:"chatID" help:"Chat ID to list messages from"`
	Cursor    string `help:"Pagination cursor (use sortKey from previous results)"`
	Direction string `help:"Pagination direction: before|after" enum:"before,after," default:"before"`
}

// MessagesListResponse is the JSON output structure.
type MessagesListResponse struct {
	Items   []MessageListItem `json:"items"`
	HasMore bool              `json:"has_more"`
}

// MessageListItem represents a message in list output.
type MessageListItem struct {
	ID         string   `json:"id"`
	ChatID     string   `json:"chat_id"`
	SenderName string   `json:"sender_name"`
	Text       string   `json:"text"`
	Timestamp  string   `json:"timestamp"`
	SortKey    string   `json:"sort_key,omitempty"`
	HasMedia   bool     `json:"has_media,omitempty"`
	Reactions  []string `json:"reactions,omitempty"`
}

// Run executes the messages list command.
func (c *MessagesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	opts := []option.RequestOption{
		option.WithAccessToken(token),
	}
	if flags.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(flags.BaseURL))
	}
	client := beeperdesktopapi.NewClient(opts...)

	// Build params
	params := beeperdesktopapi.MessageListParams{}
	if c.Cursor != "" {
		params.Cursor = beeperdesktopapi.String(c.Cursor)
	}
	if c.Direction == "before" {
		params.Direction = beeperdesktopapi.MessageListParamsDirectionBefore
	} else if c.Direction == "after" {
		params.Direction = beeperdesktopapi.MessageListParamsDirectionAfter
	}

	// Create context with timeout
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	page, err := client.Messages.List(ctx, c.ChatID, params)
	if err != nil {
		return err
	}

	// Build response
	resp := MessagesListResponse{
		Items:   make([]MessageListItem, 0, len(page.Items)),
		HasMore: len(page.Items) > 0,
	}

	for _, msg := range page.Items {
		item := MessageListItem{
			ID:         msg.ID,
			ChatID:     msg.ChatID,
			SenderName: msg.SenderName,
			Text:       msg.Text,
			SortKey:    msg.SortKey,
			HasMedia:   len(msg.Attachments) > 0,
		}
		if !msg.Timestamp.IsZero() {
			item.Timestamp = msg.Timestamp.Format(time.RFC3339)
		}
		if len(msg.Reactions) > 0 {
			item.Reactions = make([]string, 0, len(msg.Reactions))
			for _, r := range msg.Reactions {
				item.Reactions = append(item.Reactions, r.ReactionKey)
			}
		}
		resp.Items = append(resp.Items, item)
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, resp)
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		for _, item := range resp.Items {
			u.Out().Printf("%s\t%s\t%s\t%s", item.ID, item.SenderName, item.Timestamp, truncate(item.Text, 50))
		}
		return nil
	}

	// Human-readable output
	if len(resp.Items) == 0 {
		u.Out().Warn("No messages found")
		return nil
	}

	u.Out().Printf("Messages (%d):\n", len(resp.Items))
	for _, item := range resp.Items {
		ts := ""
		if item.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, item.Timestamp); err == nil {
				ts = t.Format("Jan 2 15:04")
			}
		}
		text := truncate(item.Text, 60)
		u.Out().Printf("  [%s] %s: %s", ts, item.SenderName, text)
	}

	if resp.HasMore && len(resp.Items) > 0 {
		lastItem := resp.Items[len(resp.Items)-1]
		if lastItem.SortKey != "" {
			u.Out().Dim(fmt.Sprintf("\nMore messages available. Use --cursor=%q", lastItem.SortKey))
		}
	}

	return nil
}

// MessagesSearchCmd searches for messages.
type MessagesSearchCmd struct {
	Query     string `arg:"" help:"Search query (literal word match)"`
	ChatID    string `help:"Filter by chat ID" name:"chat-id"`
	Cursor    string `help:"Pagination cursor"`
	Direction string `help:"Pagination direction: before|after" enum:"before,after,"`
	Limit     int    `help:"Max results (1-20)" default:"20"`
}

// MessagesSearchResponse is the JSON output structure.
type MessagesSearchResponse struct {
	Items      []MessageSearchItem `json:"items"`
	HasMore    bool                `json:"has_more"`
	NextCursor string              `json:"next_cursor,omitempty"`
}

// MessageSearchItem represents a message in search output.
type MessageSearchItem struct {
	ID         string `json:"id"`
	ChatID     string `json:"chat_id"`
	SenderName string `json:"sender_name"`
	Text       string `json:"text"`
	Timestamp  string `json:"timestamp"`
}

// Run executes the messages search command.
func (c *MessagesSearchCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	opts := []option.RequestOption{
		option.WithAccessToken(token),
	}
	if flags.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(flags.BaseURL))
	}
	client := beeperdesktopapi.NewClient(opts...)

	// Build params
	params := beeperdesktopapi.MessageSearchParams{
		Query: beeperdesktopapi.String(c.Query),
	}
	if c.Cursor != "" {
		params.Cursor = beeperdesktopapi.String(c.Cursor)
	}
	if c.Direction == "before" {
		params.Direction = beeperdesktopapi.MessageSearchParamsDirectionBefore
	} else if c.Direction == "after" {
		params.Direction = beeperdesktopapi.MessageSearchParamsDirectionAfter
	}
	if c.Limit > 0 {
		params.Limit = beeperdesktopapi.Int(int64(c.Limit))
	}

	// Create context with timeout
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	page, err := client.Messages.Search(ctx, params)
	if err != nil {
		return err
	}

	// Build response
	resp := MessagesSearchResponse{
		Items:   make([]MessageSearchItem, 0, len(page.Items)),
		HasMore: len(page.Items) > 0,
	}

	for _, msg := range page.Items {
		item := MessageSearchItem{
			ID:         msg.ID,
			ChatID:     msg.ChatID,
			SenderName: msg.SenderName,
			Text:       msg.Text,
		}
		if !msg.Timestamp.IsZero() {
			item.Timestamp = msg.Timestamp.Format(time.RFC3339)
		}
		resp.Items = append(resp.Items, item)
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, resp)
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		for _, item := range resp.Items {
			u.Out().Printf("%s\t%s\t%s\t%s", item.ID, item.ChatID, item.SenderName, truncate(item.Text, 50))
		}
		return nil
	}

	// Human-readable output
	if len(resp.Items) == 0 {
		u.Out().Warn("No messages found")
		u.Out().Dim("Note: Search uses literal word match, not semantic search.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	u.Out().Printf("Found %d messages:\n", len(resp.Items))
	for _, item := range resp.Items {
		ts := ""
		if item.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, item.Timestamp); err == nil {
				ts = t.Format("Jan 2")
			}
		}
		text := truncate(item.Text, 50)
		_, _ = w.Write([]byte(fmt.Sprintf("  [%s]\t%s:\t%s\n", ts, item.SenderName, text)))
	}
	w.Flush()

	return nil
}
