package beeperapi

import (
	"context"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

// accountNetworksByID returns accountID->network mappings.
// It is best-effort: lookup failures return nil so callers can continue.
// Results are cached per client instance after the first successful lookup.
func (c *Client) accountNetworksByID(ctx context.Context) map[string]string {
	c.accountNetworksMu.Lock()
	defer c.accountNetworksMu.Unlock()

	if c.accountNetworksLoaded {
		return copyStringMap(c.accountNetworks)
	}

	accounts, err := c.SDK.Accounts.List(ctx)
	if err != nil {
		return nil
	}
	c.accountNetworks = accountNetworkMap(accounts)
	c.accountNetworksLoaded = true
	return copyStringMap(c.accountNetworks)
}

func accountNetworkMap(accounts *[]beeperdesktopapi.Account) map[string]string {
	if accounts == nil || len(*accounts) == 0 {
		return nil
	}

	networks := make(map[string]string, len(*accounts))
	for _, account := range *accounts {
		if account.AccountID == "" {
			continue
		}
		networks[account.AccountID] = normalizeNetwork(account.Network)
	}

	if len(networks) == 0 {
		return nil
	}
	return networks
}

func copyStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
