package beeperapi

import (
	"context"
	"fmt"
	"net/url"
	"time"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

// MessagesService handles message operations.
type MessagesService struct {
	client *Client
}

// MessageListParams configures message list queries.
type MessageListParams struct {
	Cursor    string
	Direction string // before|after
}

// MessageSearchParams configures message search queries.
type MessageSearchParams struct {
	Query              string
	AccountIDs         []string
	ChatIDs            []string
	ChatType           string // group|single
	Sender             string // me|others|<user-id>
	MediaTypes         []string
	DateAfter          *time.Time
	DateBefore         *time.Time
	IncludeMuted       *bool
	ExcludeLowPriority *bool
	Cursor             string
	Direction          string // before|after
	Limit              int
}

// MessageListResult is the list response with pagination info.
type MessageListResult struct {
	Items      []MessageItem `json:"items"`
	HasMore    bool          `json:"has_more"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// MessageSearchResult is the search response with pagination info.
type MessageSearchResult struct {
	Items        []MessageItem `json:"items"`
	HasMore      bool          `json:"has_more"`
	OldestCursor string        `json:"oldest_cursor,omitempty"`
	NewestCursor string        `json:"newest_cursor,omitempty"`
}

// MessageItem represents a message in list/search output.
type MessageItem struct {
	ID                    string              `json:"id"`
	AccountID             string              `json:"account_id,omitempty"`
	ChatID                string              `json:"chat_id"`
	SenderID              string              `json:"sender_id,omitempty"`
	SenderName            string              `json:"sender_name,omitempty"`
	Text                  string              `json:"text,omitempty"`
	MessageType           string              `json:"message_type,omitempty"`
	Timestamp             string              `json:"timestamp,omitempty"`
	SortKey               string              `json:"sort_key,omitempty"`
	LinkedMessageID       string              `json:"linked_message_id,omitempty"`
	IsSender              bool                `json:"is_sender,omitempty"`
	IsUnread              bool                `json:"is_unread,omitempty"`
	HasMedia              bool                `json:"has_media,omitempty"`
	Attachments           []MessageAttachment `json:"attachments,omitempty"`
	Reactions             []MessageReaction   `json:"reactions,omitempty"`
	ReactionKeys          []string            `json:"reaction_keys,omitempty"`
	DownloadedAttachments []string            `json:"downloaded_attachments,omitempty"`
}

// MessageAttachment represents a message attachment.
type MessageAttachment struct {
	Type        string  `json:"type,omitempty"`
	FileName    string  `json:"file_name,omitempty"`
	FileSize    int64   `json:"file_size,omitempty"`
	MimeType    string  `json:"mime_type,omitempty"`
	SrcURL      string  `json:"src_url,omitempty"`
	Duration    float64 `json:"duration,omitempty"`
	IsGif       bool    `json:"is_gif,omitempty"`
	IsSticker   bool    `json:"is_sticker,omitempty"`
	IsVoiceNote bool    `json:"is_voice_note,omitempty"`
	PosterImg   string  `json:"poster_img,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
}

// MessageReaction represents a reaction on a message.
type MessageReaction struct {
	ID            string `json:"id,omitempty"`
	ParticipantID string `json:"participant_id,omitempty"`
	ReactionKey   string `json:"reaction_key,omitempty"`
	Emoji         bool   `json:"emoji,omitempty"`
	ImgURL        string `json:"img_url,omitempty"`
}

// List retrieves messages for a chat with cursor-based pagination.
func (s *MessagesService) List(ctx context.Context, chatID string, params MessageListParams) (MessageListResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	sdkParams := beeperdesktopapi.MessageListParams{}
	if params.Cursor != "" {
		sdkParams.Cursor = beeperdesktopapi.String(params.Cursor)
	}
	switch params.Direction {
	case "before":
		sdkParams.Direction = beeperdesktopapi.MessageListParamsDirectionBefore
	case "after":
		sdkParams.Direction = beeperdesktopapi.MessageListParamsDirectionAfter
	}

	page, err := s.client.SDK.Messages.List(ctx, chatID, sdkParams)
	if err != nil {
		return MessageListResult{}, err
	}

	result := MessageListResult{
		Items:   make([]MessageItem, 0, len(page.Items)),
		HasMore: page.HasMore,
	}

	for _, msg := range page.Items {
		item := MessageItem{
			ID:              msg.ID,
			AccountID:       msg.AccountID,
			ChatID:          msg.ChatID,
			SenderID:        msg.SenderID,
			Text:            msg.Text,
			MessageType:     string(msg.Type),
			SortKey:         msg.SortKey,
			LinkedMessageID: msg.LinkedMessageID,
			IsSender:        msg.IsSender,
			IsUnread:        msg.IsUnread,
			HasMedia:        len(msg.Attachments) > 0,
		}
		if msg.SenderName != "" {
			item.SenderName = msg.SenderName
		} else {
			item.SenderName = msg.SenderID
		}
		if !msg.Timestamp.IsZero() {
			item.Timestamp = msg.Timestamp.Format(time.RFC3339)
		}
		if len(msg.Attachments) > 0 {
			item.Attachments = make([]MessageAttachment, 0, len(msg.Attachments))
			for _, att := range msg.Attachments {
				item.Attachments = append(item.Attachments, MessageAttachment{
					Type:        string(att.Type),
					FileName:    att.FileName,
					FileSize:    int64(att.FileSize),
					MimeType:    att.MimeType,
					SrcURL:      att.SrcURL,
					Duration:    att.Duration,
					IsGif:       att.IsGif,
					IsSticker:   att.IsSticker,
					IsVoiceNote: att.IsVoiceNote,
					PosterImg:   att.PosterImg,
					Width:       int(att.Size.Width),
					Height:      int(att.Size.Height),
				})
			}
		}
		if len(msg.Reactions) > 0 {
			item.Reactions = make([]MessageReaction, 0, len(msg.Reactions))
			item.ReactionKeys = make([]string, 0, len(msg.Reactions))
			for _, r := range msg.Reactions {
				item.Reactions = append(item.Reactions, MessageReaction{
					ID:            r.ID,
					ParticipantID: r.ParticipantID,
					ReactionKey:   r.ReactionKey,
					Emoji:         r.Emoji,
					ImgURL:        r.ImgURL,
				})
				item.ReactionKeys = append(item.ReactionKeys, r.ReactionKey)
			}
		}
		result.Items = append(result.Items, item)
	}

	if result.HasMore && len(result.Items) > 0 {
		last := result.Items[len(result.Items)-1]
		if last.SortKey != "" {
			result.NextCursor = last.SortKey
		}
	}

	return result, nil
}

// SendParams configures message send requests.
type SendParams struct {
	Text             string
	ReplyToMessageID string
	Attachment       *SendAttachmentParams
}

// SendAttachmentParams configures attachment payload values for message send.
type SendAttachmentParams struct {
	UploadID string
	FileName string
	MimeType string
	Type     string
	Duration *float64
	Width    *float64
	Height   *float64
}

// EditParams configures message edit requests.
type EditParams struct {
	Text string
}

// SendResult is the response from sending a message.
type SendResult struct {
	ChatID           string `json:"chat_id"`
	PendingMessageID string `json:"pending_message_id"`
}

// EditResult is the response from editing a message.
type EditResult struct {
	ChatID    string `json:"chat_id"`
	MessageID string `json:"message_id"`
	Success   bool   `json:"success"`
}

type messageReactionBody struct {
	ReactionKey string `json:"reactionKey"`
}

func messageReactionPath(chatID, messageID, reactionKey string) string {
	return fmt.Sprintf("v1/chats/%s/messages/%s/reactions?reactionKey=%s", chatID, messageID, url.QueryEscape(reactionKey))
}

// Send sends a text message to a chat.
func (s *MessagesService) Send(ctx context.Context, chatID string, params SendParams) (SendResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	sdkParams := beeperdesktopapi.MessageSendParams{}
	if params.Text != "" {
		sdkParams.Text = beeperdesktopapi.String(params.Text)
	}
	if params.ReplyToMessageID != "" {
		sdkParams.ReplyToMessageID = beeperdesktopapi.String(params.ReplyToMessageID)
	}
	if params.Attachment != nil {
		attachment := beeperdesktopapi.MessageSendParamsAttachment{
			UploadID: params.Attachment.UploadID,
		}
		if params.Attachment.FileName != "" {
			attachment.FileName = beeperdesktopapi.String(params.Attachment.FileName)
		}
		if params.Attachment.MimeType != "" {
			attachment.MimeType = beeperdesktopapi.String(params.Attachment.MimeType)
		}
		if params.Attachment.Type != "" {
			attachment.Type = params.Attachment.Type
		}
		if params.Attachment.Duration != nil {
			attachment.Duration = beeperdesktopapi.Float(*params.Attachment.Duration)
		}
		if params.Attachment.Width != nil && params.Attachment.Height != nil {
			attachment.Size = beeperdesktopapi.MessageSendParamsAttachmentSize{
				Width:  *params.Attachment.Width,
				Height: *params.Attachment.Height,
			}
		}
		sdkParams.Attachment = attachment
	}

	resp, err := s.client.SDK.Messages.Send(ctx, chatID, sdkParams)
	if err != nil {
		return SendResult{}, err
	}

	return SendResult{
		ChatID:           resp.ChatID,
		PendingMessageID: resp.PendingMessageID,
	}, nil
}

// Edit updates the text content of an existing message.
func (s *MessagesService) Edit(ctx context.Context, chatID, messageID string, params EditParams) (EditResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	resp, err := s.client.SDK.Messages.Update(ctx, messageID, beeperdesktopapi.MessageUpdateParams{
		ChatID: chatID,
		Text:   params.Text,
	})
	if err != nil {
		return EditResult{}, err
	}

	return EditResult{
		ChatID:    resp.ChatID,
		MessageID: resp.MessageID,
		Success:   resp.Success,
	}, nil
}

// React adds a reaction to a message.
func (s *MessagesService) React(ctx context.Context, chatID, messageID, reactionKey string) error {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	path := messageReactionPath(chatID, messageID, reactionKey)
	return s.client.SDK.Post(ctx, path, messageReactionBody{ReactionKey: reactionKey}, nil)
}

// Unreact removes a reaction from a message.
func (s *MessagesService) Unreact(ctx context.Context, chatID, messageID, reactionKey string) error {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	path := messageReactionPath(chatID, messageID, reactionKey)
	return s.client.SDK.Delete(ctx, path, messageReactionBody{ReactionKey: reactionKey}, nil)
}

// Search retrieves messages matching a query.
func (s *MessagesService) Search(ctx context.Context, params MessageSearchParams) (MessageSearchResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	sdkParams := beeperdesktopapi.MessageSearchParams{}
	if params.Query != "" {
		sdkParams.Query = beeperdesktopapi.String(params.Query)
	}
	if len(params.AccountIDs) > 0 {
		sdkParams.AccountIDs = params.AccountIDs
	}
	if len(params.ChatIDs) > 0 {
		sdkParams.ChatIDs = params.ChatIDs
	}
	switch params.ChatType {
	case "group":
		sdkParams.ChatType = beeperdesktopapi.MessageSearchParamsChatTypeGroup
	case "single":
		sdkParams.ChatType = beeperdesktopapi.MessageSearchParamsChatTypeSingle
	}
	if params.Sender != "" {
		sdkParams.Sender = beeperdesktopapi.MessageSearchParamsSender(params.Sender)
	}
	if len(params.MediaTypes) > 0 {
		sdkParams.MediaTypes = params.MediaTypes
	}
	if params.DateAfter != nil {
		sdkParams.DateAfter = beeperdesktopapi.Time(*params.DateAfter)
	}
	if params.DateBefore != nil {
		sdkParams.DateBefore = beeperdesktopapi.Time(*params.DateBefore)
	}
	if params.IncludeMuted != nil {
		sdkParams.IncludeMuted = beeperdesktopapi.Bool(*params.IncludeMuted)
	}
	if params.ExcludeLowPriority != nil {
		sdkParams.ExcludeLowPriority = beeperdesktopapi.Bool(*params.ExcludeLowPriority)
	}
	if params.Cursor != "" {
		sdkParams.Cursor = beeperdesktopapi.String(params.Cursor)
	}
	switch params.Direction {
	case "before":
		sdkParams.Direction = beeperdesktopapi.MessageSearchParamsDirectionBefore
	case "after":
		sdkParams.Direction = beeperdesktopapi.MessageSearchParamsDirectionAfter
	}
	if params.Limit > 0 {
		sdkParams.Limit = beeperdesktopapi.Int(int64(params.Limit))
	}

	page, err := s.client.SDK.Messages.Search(ctx, sdkParams)
	if err != nil {
		return MessageSearchResult{}, err
	}

	result := MessageSearchResult{
		Items:        make([]MessageItem, 0, len(page.Items)),
		HasMore:      page.HasMore,
		OldestCursor: page.OldestCursor,
		NewestCursor: page.NewestCursor,
	}

	for _, msg := range page.Items {
		item := MessageItem{
			ID:              msg.ID,
			AccountID:       msg.AccountID,
			ChatID:          msg.ChatID,
			SenderID:        msg.SenderID,
			Text:            msg.Text,
			MessageType:     string(msg.Type),
			LinkedMessageID: msg.LinkedMessageID,
			IsSender:        msg.IsSender,
			IsUnread:        msg.IsUnread,
		}
		if msg.SenderName != "" {
			item.SenderName = msg.SenderName
		} else {
			item.SenderName = msg.SenderID
		}
		if !msg.Timestamp.IsZero() {
			item.Timestamp = msg.Timestamp.Format(time.RFC3339)
		}
		if len(msg.Attachments) > 0 {
			item.Attachments = make([]MessageAttachment, 0, len(msg.Attachments))
			for _, att := range msg.Attachments {
				item.Attachments = append(item.Attachments, MessageAttachment{
					Type:        string(att.Type),
					FileName:    att.FileName,
					FileSize:    int64(att.FileSize),
					MimeType:    att.MimeType,
					SrcURL:      att.SrcURL,
					Duration:    att.Duration,
					IsGif:       att.IsGif,
					IsSticker:   att.IsSticker,
					IsVoiceNote: att.IsVoiceNote,
					PosterImg:   att.PosterImg,
					Width:       int(att.Size.Width),
					Height:      int(att.Size.Height),
				})
			}
		}
		if len(msg.Reactions) > 0 {
			item.Reactions = make([]MessageReaction, 0, len(msg.Reactions))
			item.ReactionKeys = make([]string, 0, len(msg.Reactions))
			for _, r := range msg.Reactions {
				item.Reactions = append(item.Reactions, MessageReaction{
					ID:            r.ID,
					ParticipantID: r.ParticipantID,
					ReactionKey:   r.ReactionKey,
					Emoji:         r.Emoji,
					ImgURL:        r.ImgURL,
				})
				item.ReactionKeys = append(item.ReactionKeys, r.ReactionKey)
			}
		}
		result.Items = append(result.Items, item)
	}

	return result, nil
}
