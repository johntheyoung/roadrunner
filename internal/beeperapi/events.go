package beeperapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EventsService handles experimental WebSocket live events.
type EventsService struct {
	client *Client
}

// Event is a normalized payload for both control and domain messages.
type Event struct {
	Type      string           `json:"type"`
	Version   int64            `json:"version,omitempty"`
	RequestID string           `json:"requestID,omitempty"`
	Code      string           `json:"code,omitempty"`
	Message   string           `json:"message,omitempty"`
	ChatIDs   []string         `json:"chatIDs,omitempty"`
	Seq       int64            `json:"seq,omitempty"`
	TS        int64            `json:"ts,omitempty"`
	ChatID    string           `json:"chatID,omitempty"`
	IDs       []string         `json:"ids,omitempty"`
	Entries   []map[string]any `json:"entries,omitempty"`
	Raw       map[string]any   `json:"-"`
}

// IsControlMessage reports whether the event is one of the protocol control messages.
func (e Event) IsControlMessage() bool {
	switch e.Type {
	case "ready", "subscriptions.updated", "error":
		return true
	default:
		return false
	}
}

// EventsConnection wraps an active websocket connection.
type EventsConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// EventsHandshakeError is returned when opening /v1/ws fails at HTTP handshake time.
type EventsHandshakeError struct {
	StatusCode int
	Status     string
	Err        error
}

func (e *EventsHandshakeError) Error() string {
	if e == nil {
		return ""
	}
	if e.StatusCode > 0 {
		if e.Status != "" {
			return fmt.Sprintf("websocket handshake failed (%d %s): %v", e.StatusCode, e.Status, e.Err)
		}
		return fmt.Sprintf("websocket handshake failed (%d): %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("websocket handshake failed: %v", e.Err)
}

// Unwrap returns the underlying dial error.
func (e *EventsHandshakeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsEventsUnsupported returns true when /v1/ws is not available on this API build.
func IsEventsUnsupported(err error) bool {
	var hsErr *EventsHandshakeError
	if !errors.As(err, &hsErr) {
		return false
	}
	return hsErr.StatusCode == http.StatusNotFound || hsErr.StatusCode == http.StatusMethodNotAllowed
}

// IsEventsUnauthorized returns true when websocket auth fails.
func IsEventsUnauthorized(err error) bool {
	var hsErr *EventsHandshakeError
	if !errors.As(err, &hsErr) {
		return false
	}
	return hsErr.StatusCode == http.StatusUnauthorized
}

// Connect opens an authenticated websocket to /v1/ws.
func (s *EventsService) Connect(ctx context.Context) (*EventsConnection, error) {
	if strings.TrimSpace(s.client.token) == "" {
		return nil, fmt.Errorf("no token provided")
	}

	wsURL, err := eventsURLFromBase(s.client.baseURL)
	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: handshakeTimeout(s.client.timeout),
	}
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+s.client.token)

	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		statusCode := 0
		statusText := ""
		if resp != nil {
			statusCode = resp.StatusCode
			statusText = resp.Status
			_ = resp.Body.Close()
		}
		if statusCode > 0 {
			return nil, &EventsHandshakeError{
				StatusCode: statusCode,
				Status:     statusText,
				Err:        err,
			}
		}
		return nil, err
	}

	return &EventsConnection{conn: conn}, nil
}

// SetSubscriptions replaces current websocket subscriptions.
func (c *EventsConnection) SetSubscriptions(ctx context.Context, requestID string, chatIDs []string) error {
	if err := validateChatIDs(chatIDs); err != nil {
		return err
	}
	if c == nil || c.conn == nil {
		return fmt.Errorf("events connection is not initialized")
	}
	if strings.TrimSpace(requestID) == "" {
		requestID = fmt.Sprintf("sub-%d", time.Now().UnixNano())
	}

	payload := map[string]any{
		"type":      "subscriptions.set",
		"requestID": requestID,
		"chatIDs":   chatIDs,
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(deadline)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(payload)
}

// ReadEvent reads one websocket event payload.
func (c *EventsConnection) ReadEvent(ctx context.Context) (Event, error) {
	if c == nil || c.conn == nil {
		return Event{}, fmt.Errorf("events connection is not initialized")
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetReadDeadline(deadline)
	}

	var evt Event
	if err := c.conn.ReadJSON(&evt); err != nil {
		return Event{}, err
	}
	return evt, nil
}

// Close closes the websocket connection.
func (c *EventsConnection) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func handshakeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 15 * time.Second
	}
	return timeout
}

func eventsURLFromBase(baseURL string) (string, error) {
	raw := strings.TrimSpace(baseURL)
	if raw == "" {
		raw = "http://localhost:23373"
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse base URL %q: %w", raw, err)
	}

	switch strings.ToLower(u.Scheme) {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		// already websocket scheme
	default:
		return "", fmt.Errorf("unsupported base URL scheme %q for websocket events", u.Scheme)
	}

	basePath := strings.TrimRight(u.Path, "/")
	if basePath == "" {
		u.Path = "/v1/ws"
	} else {
		u.Path = basePath + "/v1/ws"
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func validateChatIDs(chatIDs []string) error {
	hasWildcard := false
	for _, chatID := range chatIDs {
		trimmed := strings.TrimSpace(chatID)
		if trimmed == "" {
			return fmt.Errorf("chatIDs cannot contain empty values")
		}
		if trimmed == "*" {
			hasWildcard = true
		}
	}
	if hasWildcard && len(chatIDs) > 1 {
		return fmt.Errorf("chatIDs cannot combine '*' with specific IDs")
	}
	return nil
}
