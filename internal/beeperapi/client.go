package beeperapi

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
	"github.com/beeper/desktop-api-go/option"
)

// Client wraps the Beeper Desktop SDK.
type Client struct {
	SDK     *beeperdesktopapi.Client
	token   string
	baseURL string
	timeout time.Duration

	accountNetworksMu     sync.Mutex
	accountNetworks       map[string]string
	accountNetworksLoaded bool
}

// NewClient creates a new Beeper API client.
// Token precedence: BEEPER_TOKEN > BEEPER_ACCESS_TOKEN > provided token
// URL precedence: BEEPER_URL > BEEPER_DESKTOP_BASE_URL > provided baseURL
func NewClient(token string, baseURL string, timeout time.Duration) (*Client, error) {
	// Token precedence
	if t := os.Getenv("BEEPER_TOKEN"); t != "" {
		token = t
	} else if t := os.Getenv("BEEPER_ACCESS_TOKEN"); t != "" {
		token = t
	}

	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	// URL precedence
	if u := os.Getenv("BEEPER_URL"); u != "" {
		baseURL = u
	} else if u := os.Getenv("BEEPER_DESKTOP_BASE_URL"); u != "" {
		baseURL = u
	}

	opts := []option.RequestOption{
		option.WithAccessToken(token),
	}

	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	sdk := beeperdesktopapi.NewClient(opts...)

	return &Client{
		SDK:     &sdk,
		token:   token,
		baseURL: baseURL,
		timeout: timeout,
	}, nil
}

// contextWithTimeout returns a context with the client's timeout.
func (c *Client) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout == 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.timeout)
}

// Accounts returns the accounts service.
func (c *Client) Accounts() *AccountsService {
	return &AccountsService{client: c}
}

// Chats returns the chats service.
func (c *Client) Chats() *ChatsService {
	return &ChatsService{client: c}
}

// Messages returns the messages service.
func (c *Client) Messages() *MessagesService {
	return &MessagesService{client: c}
}

// Assets returns the assets service.
func (c *Client) Assets() *AssetsService {
	return &AssetsService{client: c}
}

// Connect returns the connect/discovery service.
func (c *Client) Connect() *ConnectService {
	return &ConnectService{client: c}
}

// Events returns the experimental WebSocket events service.
func (c *Client) Events() *EventsService {
	return &EventsService{client: c}
}
