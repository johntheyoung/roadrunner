package cmd

import (
	"context"
	"encoding/json"
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

// MessagesCmd is the parent command for message subcommands.
type MessagesCmd struct {
	List    MessagesListCmd    `cmd:"" help:"List messages in a chat"`
	Search  MessagesSearchCmd  `cmd:"" help:"Search messages"`
	Send    MessagesSendCmd    `cmd:"" help:"Send a message to a chat"`
	Tail    MessagesTailCmd    `cmd:"" help:"Follow messages in a chat"`
	Wait    MessagesWaitCmd    `cmd:"" help:"Wait for a matching message"`
	Context MessagesContextCmd `cmd:"" help:"Fetch context around a message"`
}

// MessagesListCmd lists messages in a chat.
type MessagesListCmd struct {
	ChatID        string   `arg:"" name:"chatID" help:"Chat ID to list messages from"`
	Cursor        string   `help:"Pagination cursor (use sortKey from previous results)"`
	Direction     string   `help:"Pagination direction: before|after" enum:"before,after," default:"before"`
	DownloadMedia bool     `help:"Download attachments for listed messages" name:"download-media"`
	DownloadDir   string   `help:"Directory to save downloaded attachments" name:"download-dir" default:"."`
	Fields        []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty   bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// Run executes the messages list command.
func (c *MessagesListCmd) Run(ctx context.Context, flags *RootFlags) error {
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

	resp, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
		Cursor:    c.Cursor,
		Direction: c.Direction,
	})
	if err != nil {
		return err
	}

	if c.DownloadMedia {
		if err := downloadMessageAttachments(ctx, client, resp.Items, c.DownloadDir); err != nil {
			return err
		}
	}

	if err := failIfEmpty(c.FailIfEmpty, len(resp.Items), "messages"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "messages list")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "sender_name", "timestamp", "text", "chat_id", "sort_key", "account_id", "is_sender", "is_unread", "attachments_count", "reaction_keys", "downloaded_attachments"})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			writePlainFields(u, fields, map[string]string{
				"id":                     item.ID,
				"account_id":             item.AccountID,
				"chat_id":                item.ChatID,
				"sender_name":            item.SenderName,
				"timestamp":              item.Timestamp,
				"text":                   ui.Truncate(item.Text, 50),
				"sort_key":               item.SortKey,
				"is_sender":              formatBool(item.IsSender),
				"is_unread":              formatBool(item.IsUnread),
				"attachments_count":      fmt.Sprintf("%d", len(item.Attachments)),
				"reaction_keys":          strings.Join(item.ReactionKeys, ","),
				"downloaded_attachments": strings.Join(item.DownloadedAttachments, ","),
			})
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
		text := ui.Truncate(item.Text, 60)
		u.Out().Printf("  [%s] %s: %s", ts, item.SenderName, text)
	}

	if resp.HasMore && resp.NextCursor != "" {
		u.Out().Dim(fmt.Sprintf("\nMore messages available. Use --cursor=%q", resp.NextCursor))
	}

	return nil
}

// MessagesSearchCmd searches for messages.
type MessagesSearchCmd struct {
	Query              string   `arg:"" optional:"" help:"Search query (literal word match)"`
	AccountIDs         []string `help:"Filter by account IDs" name:"account-ids"`
	ChatIDs            []string `help:"Filter by chat IDs" name:"chat-id"`
	ChatType           string   `help:"Filter by chat type: single|group" name:"chat-type" enum:"single,group," default:""`
	Sender             string   `help:"Filter by sender: me|others|<user-id>" name:"sender"`
	MediaTypes         []string `help:"Filter by media types: any|image|video|link|file" name:"media-types"`
	DateAfter          string   `help:"Only include messages after time (RFC3339 or duration)" name:"date-after"`
	DateBefore         string   `help:"Only include messages before time (RFC3339 or duration)" name:"date-before"`
	IncludeMuted       *bool    `help:"Include muted chats (default true)" name:"include-muted"`
	ExcludeLowPriority *bool    `help:"Exclude low priority messages (default true)" name:"exclude-low-priority"`
	Cursor             string   `help:"Pagination cursor"`
	Direction          string   `help:"Pagination direction: before|after" enum:"before,after," default:""`
	Limit              int      `help:"Max results (1-20)" default:"20"`
	Fields             []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
	FailIfEmpty        bool     `help:"Exit with code 1 if no results" name:"fail-if-empty"`
}

// MessagesTailCmd follows messages in a chat via polling.
type MessagesTailCmd struct {
	ChatID    string        `arg:"" name:"chatID" help:"Chat ID to follow"`
	Cursor    string        `help:"Start cursor (sortKey)"`
	Contains  string        `help:"Only include messages containing text (case-insensitive)"`
	Sender    string        `help:"Only include messages from sender ID or name"`
	From      string        `help:"Only include messages after time (RFC3339 or duration)" name:"from"`
	To        string        `help:"Only include messages before time (RFC3339 or duration)" name:"to"`
	Interval  time.Duration `help:"Polling interval" default:"2s"`
	StopAfter time.Duration `help:"Stop after duration (0=forever)" name:"stop-after" default:"0s"`
}

// MessagesContextCmd fetches messages around a sort key.
type MessagesContextCmd struct {
	ChatID  string `arg:"" name:"chatID" help:"Chat ID to fetch context from"`
	SortKey string `arg:"" name:"sortKey" help:"Sort key of the anchor message"`
	Before  int    `help:"Number of messages before the anchor" default:"10"`
	After   int    `help:"Number of messages after the anchor" default:"0"`
}

// MessagesWaitCmd waits for a message that matches filters.
type MessagesWaitCmd struct {
	ChatID      string        `help:"Limit to a chat ID" name:"chat-id"`
	Contains    string        `help:"Only match messages containing text (case-insensitive)"`
	Sender      string        `help:"Only match messages from sender ID or name"`
	Interval    time.Duration `help:"Polling interval" default:"2s"`
	WaitTimeout time.Duration `help:"Stop waiting after duration (0=forever)" name:"wait-timeout" default:"0s"`
}

// Run executes the messages search command.
func (c *MessagesSearchCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.Limit < 1 || c.Limit > 20 {
		return errfmt.UsageError("invalid --limit %d (expected 1-20)", c.Limit)
	}
	allowedMedia := map[string]struct{}{
		"any":   {},
		"image": {},
		"video": {},
		"link":  {},
		"file":  {},
	}
	for _, media := range c.MediaTypes {
		if _, ok := allowedMedia[media]; !ok {
			return errfmt.UsageError("invalid --media-types %q (expected any|image|video|link|file)", media)
		}
	}

	var dateAfter *time.Time
	if c.DateAfter != "" {
		t, err := parseTime(c.DateAfter)
		if err != nil {
			return errfmt.UsageError("invalid --date-after %q (expected RFC3339 or duration)", c.DateAfter)
		}
		dateAfter = &t
	}

	var dateBefore *time.Time
	if c.DateBefore != "" {
		t, err := parseTime(c.DateBefore)
		if err != nil {
			return errfmt.UsageError("invalid --date-before %q (expected RFC3339 or duration)", c.DateBefore)
		}
		dateBefore = &t
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

	resp, err := client.Messages().Search(ctx, beeperapi.MessageSearchParams{
		Query:              c.Query,
		AccountIDs:         c.AccountIDs,
		ChatIDs:            normalizeChatIDs(c.ChatIDs),
		ChatType:           c.ChatType,
		Sender:             c.Sender,
		MediaTypes:         c.MediaTypes,
		DateAfter:          dateAfter,
		DateBefore:         dateBefore,
		IncludeMuted:       c.IncludeMuted,
		ExcludeLowPriority: c.ExcludeLowPriority,
		Cursor:             c.Cursor,
		Direction:          c.Direction,
		Limit:              c.Limit,
	})
	if err != nil {
		return err
	}

	if err := failIfEmpty(c.FailIfEmpty, len(resp.Items), "messages"); err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "messages search")
	}

	// Plain output (TSV)
	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"id", "chat_id", "sender_name", "text", "timestamp", "sort_key", "account_id", "is_sender", "is_unread", "attachments_count", "reaction_keys", "downloaded_attachments"})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			writePlainFields(u, fields, map[string]string{
				"id":                     item.ID,
				"account_id":             item.AccountID,
				"chat_id":                item.ChatID,
				"sender_name":            item.SenderName,
				"timestamp":              item.Timestamp,
				"text":                   ui.Truncate(item.Text, 50),
				"sort_key":               item.SortKey,
				"is_sender":              formatBool(item.IsSender),
				"is_unread":              formatBool(item.IsUnread),
				"attachments_count":      fmt.Sprintf("%d", len(item.Attachments)),
				"reaction_keys":          strings.Join(item.ReactionKeys, ","),
				"downloaded_attachments": strings.Join(item.DownloadedAttachments, ","),
			})
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
		text := ui.Truncate(item.Text, 50)
		if _, err := fmt.Fprintf(w, "  [%s]\t%s:\t%s\n", ts, item.SenderName, text); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if resp.HasMore && resp.OldestCursor != "" {
		u.Out().Dim(fmt.Sprintf("\nMore results available. Use --cursor=%q --direction=before", resp.OldestCursor))
	}

	return nil
}

// Run executes the messages tail command.
func (c *MessagesTailCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	chatID := normalizeChatID(c.ChatID)

	if c.Interval <= 0 {
		return errfmt.UsageError("invalid --interval %s (must be > 0)", c.Interval)
	}
	if isSpecialSender(c.Sender) {
		return errfmt.UsageError("--sender=%s is only supported by messages search and global waits", c.Sender)
	}

	var fromTime *time.Time
	if c.From != "" {
		t, err := parseTime(c.From)
		if err != nil {
			return errfmt.UsageError("invalid --from %q (expected RFC3339 or duration)", c.From)
		}
		fromTime = &t
	}

	var toTime *time.Time
	if c.To != "" {
		t, err := parseTime(c.To)
		if err != nil {
			return errfmt.UsageError("invalid --to %q (expected RFC3339 or duration)", c.To)
		}
		toTime = &t
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

	cursor := c.Cursor
	if cursor == "" {
		seed, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
			Direction: "before",
		})
		if err != nil {
			return err
		}
		cursor = firstSortKey(seed.Items)
	}

	deadline := time.Time{}
	if c.StopAfter > 0 {
		deadline = time.Now().Add(c.StopAfter)
	}

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil
		}

		resp, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
			Cursor:    cursor,
			Direction: "after",
		})
		if err != nil {
			return err
		}

		if len(resp.Items) > 0 {
			for _, item := range resp.Items {
				if !messageMatches(item, c.Contains, c.Sender, fromTime, toTime) {
					continue
				}
				switch {
				case outfmt.IsJSON(ctx):
					if err := encoder.Encode(item); err != nil {
						return fmt.Errorf("encode json: %w", err)
					}
				case outfmt.IsPlain(ctx):
					u.Out().Printf("%s\t%s\t%s\t%s", item.ID, item.SenderName, item.Timestamp, ui.Truncate(item.Text, 50))
				default:
					ts := ""
					if item.Timestamp != "" {
						if t, err := time.Parse(time.RFC3339, item.Timestamp); err == nil {
							ts = t.Format("Jan 2 15:04")
						}
					}
					text := ui.Truncate(item.Text, 60)
					u.Out().Printf("[%s] %s: %s", ts, item.SenderName, text)
				}
			}

			if next := lastSortKey(resp.Items); next != "" {
				cursor = next
			}
		}

		<-ticker.C
	}
}

// Run executes the messages wait command.
func (c *MessagesWaitCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.Interval <= 0 {
		return errfmt.UsageError("invalid --interval %s (must be > 0)", c.Interval)
	}
	if c.ChatID != "" && isSpecialSender(c.Sender) {
		return errfmt.UsageError("--sender=%s is only supported by messages search and global waits", c.Sender)
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

	deadline := time.Time{}
	if c.WaitTimeout > 0 {
		deadline = time.Now().Add(c.WaitTimeout)
	}

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	if c.ChatID != "" {
		chatID := normalizeChatID(c.ChatID)
		cursor := ""
		seed, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
			Direction: "before",
		})
		if err != nil {
			return err
		}
		cursor = firstSortKey(seed.Items)

		for {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return errfmt.WithCode(fmt.Errorf("timed out waiting for message"), errfmt.ExitFailure)
			}

			resp, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
				Cursor:    cursor,
				Direction: "after",
			})
			if err != nil {
				return err
			}

			if len(resp.Items) > 0 {
				for _, item := range resp.Items {
					if !messageMatches(item, c.Contains, c.Sender, nil, nil) {
						continue
					}
					return writeWaitResult(ctx, u, item, c.ChatID == "")
				}

				if next := lastSortKey(resp.Items); next != "" {
					cursor = next
				}
			}

			<-ticker.C
		}
	}

	since := time.Now()
	cursor := ""
	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return errfmt.WithCode(fmt.Errorf("timed out waiting for message"), errfmt.ExitFailure)
		}

		params := beeperapi.MessageSearchParams{
			Sender: c.Sender,
			Limit:  20,
		}
		if cursor != "" {
			params.Cursor = cursor
			params.Direction = "after"
		} else {
			params.DateAfter = &since
		}

		resp, err := client.Messages().Search(ctx, params)
		if err != nil {
			return err
		}

		for _, item := range resp.Items {
			if !messageMatches(item, c.Contains, c.Sender, nil, nil) {
				continue
			}
			return writeWaitResult(ctx, u, item, true)
		}

		if cursor != "" && resp.NewestCursor != "" {
			cursor = resp.NewestCursor
		} else if cursor == "" {
			if t := maxTimestamp(resp.Items); !t.IsZero() && t.After(since) {
				since = t.Add(time.Nanosecond)
			}
		}

		<-ticker.C
	}
}

// Run executes the messages context command.
func (c *MessagesContextCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	chatID := normalizeChatID(c.ChatID)

	if c.Before < 0 || c.After < 0 {
		return errfmt.UsageError("--before/--after must be >= 0")
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

	beforeItems := []beeperapi.MessageItem{}
	afterItems := []beeperapi.MessageItem{}

	if c.Before > 0 {
		resp, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
			Cursor:    c.SortKey,
			Direction: "before",
		})
		if err != nil {
			return err
		}
		beforeItems = limitItems(resp.Items, c.Before)
	}

	if c.After > 0 {
		resp, err := client.Messages().List(ctx, chatID, beeperapi.MessageListParams{
			Cursor:    c.SortKey,
			Direction: "after",
		})
		if err != nil {
			return err
		}
		afterItems = limitItems(resp.Items, c.After)
	}

	result := map[string]any{
		"chat_id":  chatID,
		"sort_key": c.SortKey,
		"before":   beforeItems,
		"after":    afterItems,
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, result, "messages context")
	}

	if outfmt.IsPlain(ctx) {
		for _, item := range beforeItems {
			u.Out().Printf("before\t%s\t%s\t%s\t%s", item.ID, item.SenderName, item.Timestamp, ui.Truncate(item.Text, 50))
		}
		for _, item := range afterItems {
			u.Out().Printf("after\t%s\t%s\t%s\t%s", item.ID, item.SenderName, item.Timestamp, ui.Truncate(item.Text, 50))
		}
		return nil
	}

	u.Out().Printf("Context around %s:", c.SortKey)
	if len(beforeItems) == 0 {
		u.Out().Dim("No messages before")
	} else {
		u.Out().Printf("Before:")
		for _, item := range beforeItems {
			u.Out().Printf("  %s: %s", item.SenderName, ui.Truncate(item.Text, 60))
		}
	}
	if len(afterItems) == 0 {
		u.Out().Dim("No messages after")
	} else {
		u.Out().Printf("After:")
		for _, item := range afterItems {
			u.Out().Printf("  %s: %s", item.SenderName, ui.Truncate(item.Text, 60))
		}
	}

	return nil
}

// MessagesSendCmd sends a message to a chat.
type MessagesSendCmd struct {
	ChatID           string `arg:"" name:"chatID" help:"Chat ID to send message to"`
	Text             string `arg:"" optional:"" help:"Message text to send"`
	ReplyToMessageID string `help:"Message ID to reply to" name:"reply-to"`
	TextFile         string `help:"Read message text from file ('-' for stdin)" name:"text-file"`
	Stdin            bool   `help:"Read message text from stdin" name:"stdin"`
}

// Run executes the messages send command.
func (c *MessagesSendCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	chatID := normalizeChatID(c.ChatID)

	text, err := resolveTextInput(c.Text, c.TextFile, c.Stdin, true, "message text", "--text-file", "--stdin")
	if err != nil {
		return err
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

	resp, err := client.Messages().Send(ctx, chatID, beeperapi.SendParams{
		Text:             text,
		ReplyToMessageID: c.ReplyToMessageID,
	})
	if err != nil {
		return err
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "messages send")
	}

	// Plain output
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s\t%s", resp.ChatID, resp.PendingMessageID)
		return nil
	}

	// Human-readable output
	u.Out().Success("Message sent")
	u.Out().Printf("Chat ID:    %s", resp.ChatID)
	u.Out().Printf("Pending ID: %s", resp.PendingMessageID)

	return nil
}

func firstSortKey(items []beeperapi.MessageItem) string {
	for _, item := range items {
		if item.SortKey != "" {
			return item.SortKey
		}
	}
	return ""
}

func lastSortKey(items []beeperapi.MessageItem) string {
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].SortKey != "" {
			return items[i].SortKey
		}
	}
	return ""
}

func limitItems(items []beeperapi.MessageItem, limit int) []beeperapi.MessageItem {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func messageMatches(item beeperapi.MessageItem, contains, sender string, from, to *time.Time) bool {
	if contains != "" {
		if !strings.Contains(strings.ToLower(item.Text), strings.ToLower(contains)) {
			return false
		}
	}
	if sender != "" && !isSpecialSender(sender) {
		if !strings.EqualFold(item.SenderID, sender) && !strings.EqualFold(item.SenderName, sender) {
			return false
		}
	}
	if from == nil && to == nil {
		return true
	}
	if item.Timestamp == "" {
		return false
	}
	ts, err := time.Parse(time.RFC3339, item.Timestamp)
	if err != nil {
		return false
	}
	if from != nil && ts.Before(*from) {
		return false
	}
	if to != nil && ts.After(*to) {
		return false
	}
	return true
}

func isSpecialSender(sender string) bool {
	return strings.EqualFold(sender, "me") || strings.EqualFold(sender, "others")
}

func maxTimestamp(items []beeperapi.MessageItem) time.Time {
	var max time.Time
	for _, item := range items {
		if item.Timestamp == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, item.Timestamp)
		if err != nil {
			continue
		}
		if t.After(max) {
			max = t
		}
	}
	return max
}

func writeWaitResult(ctx context.Context, u *ui.UI, item beeperapi.MessageItem, includeChatID bool) error {
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, item, "messages wait")
	}
	if outfmt.IsPlain(ctx) {
		if includeChatID {
			u.Out().Printf("%s\t%s\t%s\t%s\t%s", item.ID, item.ChatID, item.SenderName, item.Timestamp, ui.Truncate(item.Text, 50))
			return nil
		}
		u.Out().Printf("%s\t%s\t%s\t%s", item.ID, item.SenderName, item.Timestamp, ui.Truncate(item.Text, 50))
		return nil
	}

	ts := ""
	if item.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, item.Timestamp); err == nil {
			ts = t.Format("Jan 2 15:04")
		}
	}
	text := ui.Truncate(item.Text, 60)
	if includeChatID {
		u.Out().Printf("[%s] %s: %s (%s)", ts, item.SenderName, text, item.ChatID)
		return nil
	}
	u.Out().Printf("[%s] %s: %s", ts, item.SenderName, text)
	return nil
}

func downloadMessageAttachments(ctx context.Context, client *beeperapi.Client, items []beeperapi.MessageItem, destDir string) error {
	for i := range items {
		item := &items[i]
		if len(item.Attachments) == 0 {
			continue
		}
		paths := make([]string, 0, len(item.Attachments))
		for _, att := range item.Attachments {
			if att.SrcURL == "" {
				continue
			}
			path, err := downloadAttachment(ctx, client, att.SrcURL, destDir)
			if err != nil {
				return err
			}
			if path != "" {
				paths = append(paths, path)
			}
		}
		if len(paths) > 0 {
			item.DownloadedAttachments = paths
		}
	}
	return nil
}

func downloadAttachment(ctx context.Context, client *beeperapi.Client, srcURL string, destDir string) (string, error) {
	url := srcURL
	if strings.HasPrefix(url, "mxc://") || strings.HasPrefix(url, "localmxc://") {
		localURL, err := client.Assets().Download(ctx, url)
		if err != nil {
			return "", err
		}
		url = localURL
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("unsupported attachment URL: %s", url)
	}
	dest, err := assetsDownloadDestination(url, destDir)
	if err != nil {
		return "", err
	}
	if err := copyAsset(url, dest); err != nil {
		return "", err
	}
	return dest, nil
}
