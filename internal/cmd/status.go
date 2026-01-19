package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// StatusCmd summarizes unread counts and chat state.
type StatusCmd struct {
	ByAccount bool     `help:"Group unread counts by account" name:"by-account"`
	Fields    []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

type statusSummary struct {
	Accounts           int                    `json:"accounts"`
	Chats              int                    `json:"chats"`
	UnreadChats        int                    `json:"unread_chats"`
	UnreadMessages     int64                  `json:"unread_messages"`
	MutedChats         int                    `json:"muted_chats"`
	ArchivedChats      int                    `json:"archived_chats"`
	RemindersSupported bool                   `json:"reminders_supported"`
	AccountsSummary    []statusAccountSummary `json:"accounts_summary,omitempty"`
}

type statusAccountSummary struct {
	AccountID      string `json:"account_id"`
	DisplayName    string `json:"display_name,omitempty"`
	Network        string `json:"network,omitempty"`
	Chats          int    `json:"chats"`
	UnreadChats    int    `json:"unread_chats"`
	UnreadMessages int64  `json:"unread_messages"`
	MutedChats     int    `json:"muted_chats"`
	ArchivedChats  int    `json:"archived_chats"`
}

// Run executes the status command.
func (c *StatusCmd) Run(ctx context.Context, flags *RootFlags) error {
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

	accounts, err := client.Accounts().List(ctx)
	if err != nil {
		return err
	}

	summary := statusSummary{
		Accounts:           len(accounts),
		RemindersSupported: false,
	}

	accountIndex := make(map[string]statusAccountSummary, len(accounts))
	for _, acct := range accounts {
		accountIndex[acct.ID] = statusAccountSummary{
			AccountID:   acct.ID,
			DisplayName: acct.DisplayName,
			Network:     acct.Network,
		}
	}

	cursor := ""
	for {
		resp, err := client.Chats().Search(ctx, beeperapi.ChatSearchParams{
			Limit:     200,
			Cursor:    cursor,
			Direction: "before",
		})
		if err != nil {
			return err
		}

		for _, chat := range resp.Items {
			summary.Chats++
			acct := accountIndex[chat.AccountID]
			acct.AccountID = chat.AccountID
			acct.Chats++
			if chat.UnreadCount > 0 {
				summary.UnreadChats++
				summary.UnreadMessages += chat.UnreadCount
				acct.UnreadChats++
				acct.UnreadMessages += chat.UnreadCount
			}
			if chat.IsMuted {
				summary.MutedChats++
				acct.MutedChats++
			}
			if chat.IsArchived {
				summary.ArchivedChats++
				acct.ArchivedChats++
			}
			accountIndex[chat.AccountID] = acct
		}

		if !resp.HasMore || resp.OldestCursor == "" {
			break
		}
		cursor = resp.OldestCursor
	}

	if c.ByAccount {
		summary.AccountsSummary = make([]statusAccountSummary, 0, len(accountIndex))
		for _, acct := range accountIndex {
			if acct.AccountID == "" {
				continue
			}
			summary.AccountsSummary = append(summary.AccountsSummary, acct)
		}
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, summary, "status")
	}

	if outfmt.IsPlain(ctx) {
		if c.ByAccount {
			fields, err := resolveFields(c.Fields, []string{"account_id", "display_name", "network", "chats", "unread_chats", "unread_messages", "muted_chats", "archived_chats"})
			if err != nil {
				return err
			}
			for _, acct := range summary.AccountsSummary {
				writePlainFields(u, fields, map[string]string{
					"account_id":      acct.AccountID,
					"display_name":    acct.DisplayName,
					"network":         acct.Network,
					"chats":           fmt.Sprintf("%d", acct.Chats),
					"unread_chats":    fmt.Sprintf("%d", acct.UnreadChats),
					"unread_messages": fmt.Sprintf("%d", acct.UnreadMessages),
					"muted_chats":     fmt.Sprintf("%d", acct.MutedChats),
					"archived_chats":  fmt.Sprintf("%d", acct.ArchivedChats),
				})
			}
			return nil
		}
		fields, err := resolveFields(c.Fields, []string{"accounts", "chats", "unread_chats", "unread_messages", "muted_chats", "archived_chats", "reminders_supported"})
		if err != nil {
			return err
		}
		values := map[string]string{
			"accounts":            fmt.Sprintf("%d", summary.Accounts),
			"chats":               fmt.Sprintf("%d", summary.Chats),
			"unread_chats":        fmt.Sprintf("%d", summary.UnreadChats),
			"unread_messages":     fmt.Sprintf("%d", summary.UnreadMessages),
			"muted_chats":         fmt.Sprintf("%d", summary.MutedChats),
			"archived_chats":      fmt.Sprintf("%d", summary.ArchivedChats),
			"reminders_supported": formatBool(summary.RemindersSupported),
		}
		writePlainFields(u, fields, values)
		return nil
	}

	u.Out().Printf("Accounts:        %d", summary.Accounts)
	u.Out().Printf("Chats:           %d", summary.Chats)
	u.Out().Printf("Unread chats:    %d", summary.UnreadChats)
	u.Out().Printf("Unread messages: %d", summary.UnreadMessages)
	u.Out().Printf("Muted chats:     %d", summary.MutedChats)
	u.Out().Printf("Archived chats:  %d", summary.ArchivedChats)
	u.Out().Dim("Reminders:       n/a (API does not list reminders)")

	if c.ByAccount {
		u.Out().Println("")
		u.Out().Printf("By account:")
		for _, acct := range summary.AccountsSummary {
			name := acct.DisplayName
			if name == "" {
				name = acct.AccountID
			}
			u.Out().Printf("  %s: unread %d chats (%d msgs)", name, acct.UnreadChats, acct.UnreadMessages)
		}
	}

	return nil
}
