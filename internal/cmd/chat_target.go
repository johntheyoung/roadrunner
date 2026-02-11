package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

func resolveChatTargetInput(chatIDArg, chatQuery string) (string, string, error) {
	chatID := normalizeChatID(chatIDArg)
	query := strings.TrimSpace(chatQuery)

	if chatID != "" && query != "" {
		return "", "", errfmt.UsageError("cannot use chatID argument with --chat")
	}
	if chatID == "" && query == "" {
		return "", "", errfmt.UsageError("chat ID argument or --chat is required")
	}

	return chatID, query, nil
}

func resolveChatIDByQuery(ctx context.Context, client *beeperapi.Client, query string, accountIDs []string) (string, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", errfmt.UsageError("query is required")
	}

	if looksLikeChatID(q) {
		return normalizeChatID(q), nil
	}

	cursor := ""
	var matchID string
	for {
		resp, err := client.Chats().Search(ctx, beeperapi.ChatSearchParams{
			Query:      q,
			AccountIDs: accountIDs,
			Limit:      200,
			Cursor:     cursor,
			Direction:  "before",
		})
		if err != nil {
			return "", err
		}

		for _, item := range resp.Items {
			if chatExactMatch(item, q) {
				if matchID != "" && matchID != item.ID {
					return "", errfmt.WithCode(fmt.Errorf("multiple chats matched %q", q), errfmt.ExitFailure)
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
		return "", errfmt.WithCode(fmt.Errorf("no chat matched %q", q), errfmt.ExitFailure)
	}

	return matchID, nil
}
