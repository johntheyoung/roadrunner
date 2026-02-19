package cmd

import (
	"context"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
)

// ValidateTokenResult holds the result of token validation.
type ValidateTokenResult struct {
	Valid                  bool   `json:"valid"`
	Error                  string `json:"error,omitempty"`
	Account                string `json:"account,omitempty"` // First account ID if available
	ValidationMethod       string `json:"validation_method,omitempty"`
	IntrospectionAvailable bool   `json:"introspection_available,omitempty"`
	IntrospectionActive    bool   `json:"introspection_active,omitempty"`
	Subject                string `json:"subject,omitempty"`
	ConnectInfoAvailable   bool   `json:"connect_info_available,omitempty"`
	ConnectName            string `json:"connect_name,omitempty"`
	ConnectVersion         string `json:"connect_version,omitempty"`
	ConnectRuntime         string `json:"connect_runtime,omitempty"`
}

// ValidateToken checks if the token is valid by making a read-only API call.
// Returns account info if successful.
func ValidateToken(ctx context.Context, token, baseURL string, timeoutSec int) ValidateTokenResult {
	timeout := time.Duration(timeoutSec) * time.Second

	client, err := beeperapi.NewClient(token, baseURL, timeout)
	if err != nil {
		return ValidateTokenResult{
			Valid: false,
			Error: err.Error(),
		}
	}

	result := ValidateTokenResult{}
	info, infoErr := client.Connect().Info(ctx)
	if infoErr == nil {
		result.ConnectInfoAvailable = true
		result.ConnectName = info.Name
		result.ConnectVersion = info.Version
		result.ConnectRuntime = info.Runtime
	}

	introspection, introspectionErr := client.Connect().Introspect(ctx, token)
	if introspectionErr == nil {
		result.ValidationMethod = "introspection"
		result.IntrospectionAvailable = true
		result.IntrospectionActive = introspection.Active
		result.Subject = introspection.Subject
		result.Valid = introspection.Active
		if !introspection.Active {
			result.Error = "token is inactive"
		}
		return result
	}
	if beeperapi.IsUnauthorized(introspectionErr) {
		result.ValidationMethod = "introspection"
		result.IntrospectionAvailable = true
		result.Valid = false
		result.Error = "unauthorized - token is invalid or expired"
		return result
	}

	accounts, err := client.Accounts().List(ctx)
	if err != nil {
		// Check if it's an auth error
		if beeperapi.IsUnauthorized(err) {
			return ValidateTokenResult{
				Valid: false,
				Error: "unauthorized - token is invalid or expired",
			}
		}
		return ValidateTokenResult{
			Valid: false,
			Error: beeperapi.FormatError(err),
		}
	}

	result.Valid = true
	result.ValidationMethod = "accounts"
	if len(accounts) > 0 {
		result.Account = accounts[0].ID
	}

	return result
}
