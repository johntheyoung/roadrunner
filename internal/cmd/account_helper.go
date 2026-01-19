package cmd

import "github.com/johntheyoung/roadrunner/internal/config"

// applyAccountDefault returns the account IDs to use, applying the default account
// if the provided list is empty and a default is set.
func applyAccountDefault(accountIDs []string, defaultAccount string) []string {
	if len(accountIDs) > 0 {
		// Resolve aliases in the provided list
		resolved := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			resolved[i] = config.ResolveAccountAlias(id)
		}
		return resolved
	}
	if defaultAccount != "" {
		return []string{config.ResolveAccountAlias(defaultAccount)}
	}
	return accountIDs
}

// resolveAccount resolves an account ID or alias to the actual account ID.
// If accountID is empty, uses the default account.
func resolveAccount(accountID string, defaultAccount string) string {
	if accountID == "" {
		if defaultAccount != "" {
			return config.ResolveAccountAlias(defaultAccount)
		}
		return ""
	}
	return config.ResolveAccountAlias(accountID)
}
