package beeperapi

import (
	"context"
	"time"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

// SearchResult is the response from global search.
type SearchResult struct {
	Chats    []SearchChat    `json:"chats"`
	InGroups []SearchChat    `json:"in_groups"`
	Messages SearchMessages  `json:"messages"`
}

// SearchChat represents a chat in search results.
type SearchChat struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Network     string `json:"network"`
	AccountID   string `json:"account_id"`
	UnreadCount int64  `json:"unread_count"`
}

// SearchMessages represents message results from global search.
type SearchMessages struct {
	Items        []MessageItem `json:"items"`
	HasMore      bool          `json:"has_more"`
	OldestCursor string        `json:"oldest_cursor,omitempty"`
	NewestCursor string        `json:"newest_cursor,omitempty"`
}

// Search performs a global search across chats and messages.
func (c *Client) Search(ctx context.Context, query string) (SearchResult, error) {
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	sdkParams := beeperdesktopapi.SearchParams{
		Query: query,
	}

	resp, err := c.SDK.Search(ctx, sdkParams)
	if err != nil {
		return SearchResult{}, err
	}

	result := SearchResult{
		Chats:    make([]SearchChat, 0, len(resp.Results.Chats)),
		InGroups: make([]SearchChat, 0, len(resp.Results.InGroups)),
		Messages: SearchMessages{
			Items:        make([]MessageItem, 0, len(resp.Results.Messages.Items)),
			HasMore:      resp.Results.Messages.HasMore,
			OldestCursor: resp.Results.Messages.OldestCursor,
			NewestCursor: resp.Results.Messages.NewestCursor,
		},
	}

	// Map chats
	for _, chat := range resp.Results.Chats {
		result.Chats = append(result.Chats, SearchChat{
			ID:          chat.ID,
			Title:       chat.Title,
			Type:        string(chat.Type),
			Network:     chat.Network,
			AccountID:   chat.AccountID,
			UnreadCount: chat.UnreadCount,
		})
	}

	// Map in_groups
	for _, chat := range resp.Results.InGroups {
		result.InGroups = append(result.InGroups, SearchChat{
			ID:          chat.ID,
			Title:       chat.Title,
			Type:        string(chat.Type),
			Network:     chat.Network,
			AccountID:   chat.AccountID,
			UnreadCount: chat.UnreadCount,
		})
	}

	// Map messages
	for _, msg := range resp.Results.Messages.Items {
		item := MessageItem{
			ID:       msg.ID,
			ChatID:   msg.ChatID,
			SenderID: msg.SenderID,
			Text:     msg.Text,
		}
		if msg.SenderName != "" {
			item.SenderName = msg.SenderName
		} else {
			item.SenderName = msg.SenderID
		}
		if !msg.Timestamp.IsZero() {
			item.Timestamp = msg.Timestamp.Format(time.RFC3339)
		}
		result.Messages.Items = append(result.Messages.Items, item)
	}

	return result, nil
}
