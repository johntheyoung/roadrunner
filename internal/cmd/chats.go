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

// ChatsCmd is the parent command for chat subcommands.
type ChatsCmd struct {
	List    ChatsListCmd    `cmd:"" help:"List chats"`
	Search  ChatsSearchCmd  `cmd:"" help:"Search chats"`
	Resolve ChatsResolveCmd `cmd:"" help:"Resolve a chat by exact match"`
	Get     ChatsGetCmd     `cmd:"" help:"Get chat details"`
	Create  ChatsCreateCmd  `cmd:"" help:"Create a new chat"`
	Archive ChatsArchiveCmd `cmd:"" help:"Archive or unarchive a chat"`
}

// ChatsListCmd lists chats.
type ChatsListCmd struct {
	AccountIDs  []string `help:"Filter by account IDs" name:"account-ids"`
	Cursor      string   `help:"Pagination cursor"`
	Direction   string   `help:"Pagination direction: before|after" enum:"before,after," default:""`
	Fields      []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// Run executes the chats list command.
func (c *ChatsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	accountIDs := applyAccountDefault(c.AccountIDs, flags.Account)
	resp, err := client.Chats().List(ctx, beeperapi.ChatListParams{
		AccountIDs: accountIDs,
		Cursor:     c.Cursor,
		Direction:  c.Direction,
	})
	if err != nil {
		return err
	}

	if err := failIfEmpty(c.FailIfEmpty, len(resp.Items), "chats"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "chats list")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "title", "account_id", "display_name", "last_activity", "preview"})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			writePlainFields(u, fields, map[string]string{
				"id":            item.ID,
				"title":         item.Title,
				"display_name":  item.DisplayName,
				"account_id":    item.AccountID,
				"last_activity": item.LastActivity,
				"preview":       item.Preview,
			})
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
		title := item.Title
		if item.DisplayName != "" {
			title = item.DisplayName
		}
		title = ui.Truncate(title, 40)
		if _, err := fmt.Fprintf(w, "  %s\t%s\n", title, item.ID); err != nil {
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

// ChatsSearchCmd searches for chats.
type ChatsSearchCmd struct {
	Query              string   `arg:"" optional:"" help:"Search query"`
	AccountIDs         []string `help:"Filter by account IDs" name:"account-ids"`
	Inbox              string   `help:"Filter by inbox: primary|low-priority|archive" enum:"primary,low-priority,archive," default:""`
	UnreadOnly         bool     `help:"Only show unread chats" name:"unread-only"`
	IncludeMuted       *bool    `help:"Include muted chats (default true)" name:"include-muted"`
	LastActivityAfter  string   `help:"Only include chats after time (RFC3339 or duration)" name:"last-activity-after"`
	LastActivityBefore string   `help:"Only include chats before time (RFC3339 or duration)" name:"last-activity-before"`
	Type               string   `help:"Filter by type: direct|group|any" enum:"direct,group,any," default:""`
	Scope              string   `help:"Search scope: titles|participants" enum:"titles,participants," default:""`
	Limit              int      `help:"Max results (1-200)" default:"50"`
	Cursor             string   `help:"Pagination cursor"`
	Direction          string   `help:"Pagination direction: before|after" enum:"before,after," default:""`
	Fields             []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty        bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// Run executes the chats search command.
func (c *ChatsSearchCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.Limit < 1 || c.Limit > 200 {
		return errfmt.UsageError("invalid --limit %d (expected 1-200)", c.Limit)
	}

	var lastAfter *time.Time
	if c.LastActivityAfter != "" {
		t, err := parseTime(c.LastActivityAfter)
		if err != nil {
			return errfmt.UsageError("invalid --last-activity-after %q (expected RFC3339 or duration)", c.LastActivityAfter)
		}
		lastAfter = &t
	}

	var lastBefore *time.Time
	if c.LastActivityBefore != "" {
		t, err := parseTime(c.LastActivityBefore)
		if err != nil {
			return errfmt.UsageError("invalid --last-activity-before %q (expected RFC3339 or duration)", c.LastActivityBefore)
		}
		lastBefore = &t
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

	accountIDs := applyAccountDefault(c.AccountIDs, flags.Account)
	resp, err := client.Chats().Search(ctx, beeperapi.ChatSearchParams{
		Query:              c.Query,
		AccountIDs:         accountIDs,
		Inbox:              c.Inbox,
		UnreadOnly:         c.UnreadOnly,
		IncludeMuted:       c.IncludeMuted,
		LastActivityAfter:  lastAfter,
		LastActivityBefore: lastBefore,
		Type:               c.Type,
		Scope:              c.Scope,
		Limit:              c.Limit,
		Cursor:             c.Cursor,
		Direction:          c.Direction,
	})
	if err != nil {
		return err
	}

	if err := failIfEmpty(c.FailIfEmpty, len(resp.Items), "chats"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "chats search")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "title", "type", "unread_count", "display_name", "network", "account_id", "is_archived", "is_muted"})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			writePlainFields(u, fields, map[string]string{
				"id":           item.ID,
				"title":        item.Title,
				"display_name": item.DisplayName,
				"type":         item.Type,
				"network":      item.Network,
				"account_id":   item.AccountID,
				"unread_count": fmt.Sprintf("%d", item.UnreadCount),
				"is_archived":  formatBool(item.IsArchived),
				"is_muted":     formatBool(item.IsMuted),
			})
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

// ChatsGetCmd gets a single chat.
type ChatsGetCmd struct {
	ChatID string `arg:"" name:"chatID" help:"Chat ID to retrieve"`
}

// ChatsResolveCmd resolves a chat by exact match.
type ChatsResolveCmd struct {
	Query      string   `arg:"" help:"Exact chat title, display name, or ID"`
	AccountIDs []string `help:"Filter by account IDs" name:"account-ids"`
	Fields     []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

// ChatsCreateCmd creates a new chat.
type ChatsCreateCmd struct {
	AccountID    string   `arg:"" name:"accountID" optional:"" help:"Account ID to create the chat on (uses --account default if omitted)"`
	Participants []string `help:"Participant IDs (repeatable)" name:"participant"`
	Type         string   `help:"Chat type: single|group" enum:"single,group," default:""`
	Title        string   `help:"Title for group chats"`
	Message      string   `help:"Optional first message content"`
}

// Run executes the chats get command.
func (c *ChatsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	chatID := normalizeChatID(c.ChatID)

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	chat, err := client.Chats().Get(ctx, chatID)
	if err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, chat, "chats get")
	}

	// Plain output
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s\t%s\t%s", chat.ID, chat.Title, chat.AccountID)
		return nil
	}

	// Human-readable output
	displayTitle := chat.Title
	if chat.DisplayName != "" {
		displayTitle = chat.DisplayName
	}
	u.Out().Printf("Chat: %s", displayTitle)
	if chat.DisplayName != "" && chat.DisplayName != chat.Title {
		u.Out().Printf("Title:  %s", chat.Title)
	}
	u.Out().Printf("ID:      %s", chat.ID)
	u.Out().Printf("Account: %s", chat.AccountID)
	u.Out().Printf("Type:    %s", chat.Type)
	u.Out().Printf("Network: %s", chat.Network)
	if chat.UnreadCount > 0 {
		u.Out().Printf("Unread:  %d", chat.UnreadCount)
	}
	if chat.LastActivity != "" {
		u.Out().Printf("Last:    %s", chat.LastActivity)
	}

	return nil
}

// Run executes the chats resolve command.
func (c *ChatsResolveCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if strings.TrimSpace(c.Query) == "" {
		return errfmt.UsageError("query is required")
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

	query := strings.TrimSpace(c.Query)
	if looksLikeChatID(query) {
		return writeResolvedChat(ctx, u, client, normalizeChatID(query), c.Fields)
	}

	cursor := ""
	var matchID string
	accountIDs := applyAccountDefault(c.AccountIDs, flags.Account)
	for {
		resp, err := client.Chats().Search(ctx, beeperapi.ChatSearchParams{
			Query:      query,
			AccountIDs: accountIDs,
			Limit:      200,
			Cursor:     cursor,
			Direction:  "before",
		})
		if err != nil {
			return err
		}

		for _, item := range resp.Items {
			if chatExactMatch(item, query) {
				if matchID != "" && matchID != item.ID {
					return errfmt.WithCode(fmt.Errorf("multiple chats matched %q", query), errfmt.ExitFailure)
				}
				matchID = item.ID
			}
		}

		if !resp.HasMore || resp.OldestCursor == "" {
			break
		}
		cursor = resp.OldestCursor
	}

	if matchID == "" {
		return errfmt.WithCode(fmt.Errorf("no chat matched %q", query), errfmt.ExitFailure)
	}

	return writeResolvedChat(ctx, u, client, matchID, c.Fields)
}

// Run executes the chats create command.
func (c *ChatsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	accountID := resolveAccount(c.AccountID, flags.Account)
	if accountID == "" {
		return errfmt.UsageError("account ID is required")
	}
	if len(c.Participants) == 0 {
		return errfmt.UsageError("at least one --participant is required")
	}

	chatType := c.Type
	if chatType == "" {
		if len(c.Participants) == 1 {
			chatType = "single"
		} else {
			chatType = "group"
		}
	}
	if chatType == "single" && len(c.Participants) != 1 {
		return errfmt.UsageError("single chats require exactly one --participant")
	}
	if chatType == "group" && len(c.Participants) < 2 {
		return errfmt.UsageError("group chats require at least two --participant values")
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

	resp, err := client.Chats().Create(ctx, beeperapi.ChatCreateParams{
		AccountID:      accountID,
		ParticipantIDs: c.Participants,
		Type:           chatType,
		Title:          c.Title,
		MessageText:    c.Message,
	})
	if err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "chats create")
	}

	// Plain output
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s", resp.ChatID)
		return nil
	}

	u.Out().Success("Chat created")
	u.Out().Printf("Chat ID: %s", resp.ChatID)
	return nil
}

func chatExactMatch(chat beeperapi.ChatSearchItem, query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return false
	}
	if strings.EqualFold(chat.ID, q) {
		return true
	}
	if strings.EqualFold(chat.Title, q) {
		return true
	}
	if strings.EqualFold(chat.DisplayName, q) {
		return true
	}
	return false
}

func looksLikeChatID(value string) bool {
	return strings.HasPrefix(value, "!") && strings.Contains(value, ":")
}

func writeResolvedChat(ctx context.Context, u *ui.UI, client *beeperapi.Client, chatID string, fields []string) error {
	chat, err := client.Chats().Get(ctx, chatID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, chat, "chats resolve")
	}

	if outfmt.IsPlain(ctx) {
		cols, err := resolveFields(fields, []string{"id", "title", "display_name", "account_id", "type", "network", "unread_count", "is_archived", "is_muted"})
		if err != nil {
			return err
		}
		writePlainFields(u, cols, map[string]string{
			"id":           chat.ID,
			"title":        chat.Title,
			"display_name": chat.DisplayName,
			"account_id":   chat.AccountID,
			"type":         chat.Type,
			"network":      chat.Network,
			"unread_count": fmt.Sprintf("%d", chat.UnreadCount),
			"is_archived":  formatBool(chat.IsArchived),
			"is_muted":     formatBool(chat.IsMuted),
		})
		return nil
	}

	displayTitle := chat.Title
	if chat.DisplayName != "" {
		displayTitle = chat.DisplayName
	}
	u.Out().Printf("Chat: %s", displayTitle)
	u.Out().Printf("ID:    %s", chat.ID)
	u.Out().Printf("Type:  %s", chat.Type)
	u.Out().Printf("Account: %s", chat.AccountID)
	if chat.UnreadCount > 0 {
		u.Out().Printf("Unread: %d", chat.UnreadCount)
	}

	return nil
}

// ChatsArchiveCmd archives or unarchives a chat.
type ChatsArchiveCmd struct {
	ChatID    string `arg:"" name:"chatID" help:"Chat ID to archive/unarchive"`
	Unarchive bool   `help:"Unarchive instead of archive" name:"unarchive"`
}

// Run executes the chats archive command.
func (c *ChatsArchiveCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	chatID := normalizeChatID(c.ChatID)

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	archived := !c.Unarchive
	action := "archive chat " + chatID
	if c.Unarchive {
		action = "unarchive chat " + chatID
	}
	if err := confirmDestructive(flags, action); err != nil {
		return err
	}
	if err := client.Chats().Archive(ctx, chatID, archived); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		result := map[string]any{
			"chat_id":  chatID,
			"archived": archived,
		}
		return writeJSON(ctx, result, "chats archive")
	}

	// Plain output
	if outfmt.IsPlain(ctx) {
		action := "archived"
		if c.Unarchive {
			action = "unarchived"
		}
		u.Out().Printf("%s\t%s", chatID, action)
		return nil
	}

	// Human-readable output
	if c.Unarchive {
		u.Out().Success("Chat unarchived")
	} else {
		u.Out().Success("Chat archived")
	}

	return nil
}
