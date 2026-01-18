package cmd

import (
	"context"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// StatusCmd summarizes unread counts and chat state.
type StatusCmd struct {
	ByAccount bool `help:"Group unread counts by account" name:"by-account"`
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
			u.Out().Printf("account_id\tdisplay_name\tnetwork\tchats\tunread_chats\tunread_messages\tmuted_chats\tarchived_chats")
			for _, acct := range summary.AccountsSummary {
				u.Out().Printf("%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d",
					acct.AccountID,
					acct.DisplayName,
					acct.Network,
					acct.Chats,
					acct.UnreadChats,
					acct.UnreadMessages,
					acct.MutedChats,
					acct.ArchivedChats,
				)
			}
			return nil
		}
		u.Out().Printf("accounts\t%d", summary.Accounts)
		u.Out().Printf("chats\t%d", summary.Chats)
		u.Out().Printf("unread_chats\t%d", summary.UnreadChats)
		u.Out().Printf("unread_messages\t%d", summary.UnreadMessages)
		u.Out().Printf("muted_chats\t%d", summary.MutedChats)
		u.Out().Printf("archived_chats\t%d", summary.ArchivedChats)
		u.Out().Printf("reminders_supported\t%v", summary.RemindersSupported)
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
