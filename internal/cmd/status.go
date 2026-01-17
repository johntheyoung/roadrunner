package cmd

import (
	"context"
	"os"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// StatusCmd summarizes unread counts and chat state.
type StatusCmd struct{}

type statusSummary struct {
	Accounts           int   `json:"accounts"`
	Chats              int   `json:"chats"`
	UnreadChats        int   `json:"unread_chats"`
	UnreadMessages     int64 `json:"unread_messages"`
	MutedChats         int   `json:"muted_chats"`
	ArchivedChats      int   `json:"archived_chats"`
	RemindersSupported bool  `json:"reminders_supported"`
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
			if chat.UnreadCount > 0 {
				summary.UnreadChats++
				summary.UnreadMessages += chat.UnreadCount
			}
			if chat.IsMuted {
				summary.MutedChats++
			}
			if chat.IsArchived {
				summary.ArchivedChats++
			}
		}

		if !resp.HasMore || resp.OldestCursor == "" {
			break
		}
		cursor = resp.OldestCursor
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, summary)
	}

	if outfmt.IsPlain(ctx) {
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

	return nil
}
