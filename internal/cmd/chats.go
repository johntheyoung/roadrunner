package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
	"github.com/beeper/desktop-api-go/option"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// ChatsCmd is the parent command for chat subcommands.
type ChatsCmd struct {
	List   ChatsListCmd   `cmd:"" help:"List chats"`
	Search ChatsSearchCmd `cmd:"" help:"Search chats"`
	Get    ChatsGetCmd    `cmd:"" help:"Get chat details"`
}

// ChatsListCmd lists chats.
type ChatsListCmd struct {
	AccountIDs []string `help:"Filter by account IDs" name:"account-ids"`
	Cursor     string   `help:"Pagination cursor"`
	Direction  string   `help:"Pagination direction: before|after" enum:"before,after,"`
}

// ChatsListResponse is the JSON output structure.
type ChatsListResponse struct {
	Items      []ChatListItem `json:"items"`
	HasMore    bool           `json:"has_more"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// ChatListItem represents a chat in list output.
type ChatListItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	AccountID    string `json:"account_id"`
	LastActivity string `json:"last_activity,omitempty"`
	Preview      string `json:"preview,omitempty"`
}

// Run executes the chats list command.
func (c *ChatsListCmd) Run(ctx context.Context, flags *RootFlags) error {
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
	params := beeperdesktopapi.ChatListParams{}
	if len(c.AccountIDs) > 0 {
		params.AccountIDs = c.AccountIDs
	}
	if c.Cursor != "" {
		params.Cursor = beeperdesktopapi.String(c.Cursor)
	}
	if c.Direction == "before" {
		params.Direction = beeperdesktopapi.ChatListParamsDirectionBefore
	} else if c.Direction == "after" {
		params.Direction = beeperdesktopapi.ChatListParamsDirectionAfter
	}

	// Create context with timeout
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	page, err := client.Chats.List(ctx, params)
	if err != nil {
		return err
	}

	// Build response
	resp := ChatsListResponse{
		Items:   make([]ChatListItem, 0, len(page.Items)),
		HasMore: len(page.Items) > 0, // Approximate - SDK doesn't expose cursor clearly
	}

	for _, chat := range page.Items {
		item := ChatListItem{
			ID:        chat.ID,
			Title:     chat.Title,
			AccountID: chat.AccountID,
		}
		if !chat.LastActivity.IsZero() {
			item.LastActivity = chat.LastActivity.Format(time.RFC3339)
		}
		if chat.JSON.Preview.Valid() {
			item.Preview = chat.Preview.Text
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
			u.Out().Printf("%s\t%s\t%s", item.ID, item.Title, item.AccountID)
		}
		return nil
	}

	// Human-readable output
	if len(resp.Items) == 0 {
		u.Out().Warn("No chats found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	u.Out().Printf("Chats (%d):\n", len(resp.Items))
	for _, item := range resp.Items {
		title := truncate(item.Title, 40)
		_, _ = w.Write([]byte(fmt.Sprintf("  %s\t%s\n", title, item.ID)))
	}
	w.Flush()

	return nil
}

// ChatsSearchCmd searches for chats.
type ChatsSearchCmd struct {
	Query      string `arg:"" optional:"" help:"Search query"`
	Inbox      string `help:"Filter by inbox: primary" enum:"primary,"`
	UnreadOnly bool   `help:"Only show unread chats" name:"unread-only"`
	Type       string `help:"Filter by type: direct|group" enum:"direct,group,"`
	Limit      int    `help:"Max results (1-200)" default:"50"`
}

// ChatsSearchResponse is the JSON output structure.
type ChatsSearchResponse struct {
	Items      []ChatSearchItem `json:"items"`
	HasMore    bool             `json:"has_more"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

// ChatSearchItem represents a chat in search output.
type ChatSearchItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Network     string `json:"network"`
	UnreadCount int64  `json:"unread_count"`
	IsArchived  bool   `json:"is_archived"`
	IsMuted     bool   `json:"is_muted"`
}

// Run executes the chats search command.
func (c *ChatsSearchCmd) Run(ctx context.Context, flags *RootFlags) error {
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
	params := beeperdesktopapi.ChatSearchParams{}
	if c.Query != "" {
		params.Query = beeperdesktopapi.String(c.Query)
	}
	if c.Inbox == "primary" {
		params.Inbox = beeperdesktopapi.ChatSearchParamsInboxPrimary
	}
	// Note: "other" inbox filter not yet supported by SDK
	if c.UnreadOnly {
		params.UnreadOnly = beeperdesktopapi.Bool(true)
	}
	if c.Type == "direct" {
		params.Type = beeperdesktopapi.ChatSearchParamsTypeSingle
	} else if c.Type == "group" {
		params.Type = beeperdesktopapi.ChatSearchParamsTypeGroup
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

	page, err := client.Chats.Search(ctx, params)
	if err != nil {
		return err
	}

	// Build response
	resp := ChatsSearchResponse{
		Items:   make([]ChatSearchItem, 0, len(page.Items)),
		HasMore: len(page.Items) > 0,
	}

	for _, chat := range page.Items {
		resp.Items = append(resp.Items, ChatSearchItem{
			ID:          chat.ID,
			Title:       chat.Title,
			Type:        string(chat.Type),
			Network:     chat.Network,
			UnreadCount: chat.UnreadCount,
			IsArchived:  chat.IsArchived,
			IsMuted:     chat.IsMuted,
		})
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, resp)
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		for _, item := range resp.Items {
			u.Out().Printf("%s\t%s\t%s\t%d", item.ID, item.Title, item.Type, item.UnreadCount)
		}
		return nil
	}

	// Human-readable output
	if len(resp.Items) == 0 {
		u.Out().Warn("No chats found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	u.Out().Printf("Found %d chats:\n", len(resp.Items))
	for _, item := range resp.Items {
		title := truncate(item.Title, 35)
		unread := ""
		if item.UnreadCount > 0 {
			unread = fmt.Sprintf("(%d)", item.UnreadCount)
		}
		_, _ = w.Write([]byte(fmt.Sprintf("  %s\t%s\t%s\t%s\n", title, item.Type, unread, item.ID)))
	}
	w.Flush()

	return nil
}

// ChatsGetCmd gets a single chat.
type ChatsGetCmd struct {
	ChatID string `arg:"" name:"chatID" help:"Chat ID to retrieve"`
}

// Run executes the chats get command.
func (c *ChatsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
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

	// Create context with timeout
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	chat, err := client.Chats.Get(ctx, c.ChatID, beeperdesktopapi.ChatGetParams{})
	if err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, chat)
	}

	// Plain output
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s\t%s\t%s", chat.ID, chat.Title, chat.AccountID)
		return nil
	}

	// Human-readable output
	u.Out().Printf("Chat: %s", chat.Title)
	u.Out().Printf("ID:      %s", chat.ID)
	u.Out().Printf("Account: %s", chat.AccountID)
	u.Out().Printf("Type:    %s", chat.Type)
	u.Out().Printf("Network: %s", chat.Network)
	if chat.UnreadCount > 0 {
		u.Out().Printf("Unread:  %d", chat.UnreadCount)
	}

	return nil
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
