package beeperapi

import (
	"context"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

// accountNetworksByID returns accountID->network mappings.
// It is best-effort: lookup failures return nil so callers can continue.
func (c *Client) accountNetworksByID(ctx context.Context) map[string]string {
	accounts, err := c.SDK.Accounts.List(ctx)
	if err != nil {
		return nil
	}
	return accountNetworkMap(accounts)
}

func accountNetworkMap(accounts *[]beeperdesktopapi.Account) map[string]string {
	if accounts == nil || len(*accounts) == 0 {
		return nil
	}

	networks := make(map[string]string, len(*accounts))
	for _, account := range *accounts {
		if account.AccountID == "" || account.Network == "" {
			continue
		}
		networks[account.AccountID] = account.Network
	}

	if len(networks) == 0 {
		return nil
	}
	return networks
}
