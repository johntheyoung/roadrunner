package beeperapi

import "strings"

const unknownNetwork = "unknown"

func normalizeNetwork(network string) string {
	value := strings.TrimSpace(network)
	if value == "" {
		return unknownNetwork
	}
	return value
}

func networkForAccount(accountNetworks map[string]string, accountID string) string {
	if accountNetworks == nil {
		return unknownNetwork
	}
	return normalizeNetwork(accountNetworks[accountID])
}
