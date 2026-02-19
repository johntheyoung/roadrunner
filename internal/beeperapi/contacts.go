package beeperapi

import (
	"context"
	"fmt"
	"net/url"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

// Contact represents a contact search result.
type Contact struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name,omitempty"`
	Username      string `json:"username,omitempty"`
	Email         string `json:"email,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	CannotMessage bool   `json:"cannot_message,omitempty"`
	ImgURL        string `json:"img_url,omitempty"`
}

// ContactListParams configures cursor-based contacts list queries.
type ContactListParams struct {
	Cursor    string
	Direction string // before|after
}

// ContactListResult is the contacts list response with pagination info.
type ContactListResult struct {
	Items        []Contact `json:"items"`
	HasMore      bool      `json:"has_more"`
	OldestCursor string    `json:"oldest_cursor,omitempty"`
	NewestCursor string    `json:"newest_cursor,omitempty"`
}

type contactsListQuery struct {
	Cursor    string
	Direction string
}

func (q contactsListQuery) URLQuery() (url.Values, error) {
	values := url.Values{}
	if q.Cursor != "" {
		values.Set("cursor", q.Cursor)
	}
	if q.Direction != "" {
		values.Set("direction", q.Direction)
	}
	return values, nil
}

type contactsListUser struct {
	ID            string `json:"id"`
	FullName      string `json:"fullName"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phoneNumber"`
	CannotMessage bool   `json:"cannotMessage"`
	ImgURL        string `json:"imgURL"`
}

type contactsListResponse struct {
	Items        []contactsListUser `json:"items"`
	HasMore      bool               `json:"hasMore"`
	OldestCursor string             `json:"oldestCursor"`
	NewestCursor string             `json:"newestCursor"`
	NextCursor   string             `json:"nextCursor"`
}

// SearchContacts finds contacts on a specific account.
func (s *AccountsService) SearchContacts(ctx context.Context, accountID string, query string) ([]Contact, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	resp, err := s.client.SDK.Accounts.Contacts.Search(ctx, accountID, beeperdesktopapi.AccountContactSearchParams{
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]Contact, 0, len(resp.Items))
	for _, c := range resp.Items {
		contacts = append(contacts, Contact{
			ID:            c.ID,
			FullName:      c.FullName,
			Username:      c.Username,
			Email:         c.Email,
			PhoneNumber:   c.PhoneNumber,
			CannotMessage: c.CannotMessage,
			ImgURL:        c.ImgURL,
		})
	}

	return contacts, nil
}

// ListContacts lists contacts on a specific account with cursor pagination.
func (s *AccountsService) ListContacts(ctx context.Context, accountID string, params ContactListParams) (ContactListResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	var resp contactsListResponse
	if err := s.client.SDK.Get(ctx, fmt.Sprintf("v1/accounts/%s/contacts/list", accountID), contactsListQuery{
		Cursor:    params.Cursor,
		Direction: params.Direction,
	}, &resp); err != nil {
		return ContactListResult{}, err
	}

	result := ContactListResult{
		Items:        make([]Contact, 0, len(resp.Items)),
		HasMore:      resp.HasMore,
		OldestCursor: resp.OldestCursor,
		NewestCursor: resp.NewestCursor,
	}
	if result.OldestCursor == "" && resp.NextCursor != "" {
		result.OldestCursor = resp.NextCursor
	}
	if result.NewestCursor == "" && params.Direction == "after" && resp.NextCursor != "" {
		result.NewestCursor = resp.NextCursor
	}

	for _, c := range resp.Items {
		result.Items = append(result.Items, Contact{
			ID:            c.ID,
			FullName:      c.FullName,
			Username:      c.Username,
			Email:         c.Email,
			PhoneNumber:   c.PhoneNumber,
			CannotMessage: c.CannotMessage,
			ImgURL:        c.ImgURL,
		})
	}

	return result, nil
}
