package beeperapi

import (
	"encoding/json"
	"errors"
	"fmt"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

// FormatError converts SDK errors to user-friendly messages.
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var apiErr *beeperdesktopapi.Error
	if errors.As(err, &apiErr) {
		return formatAPIError(apiErr)
	}

	return err.Error()
}

func formatAPIError(err *beeperdesktopapi.Error) string {
	// Try to extract message from response body
	respBody := err.DumpResponse(false)
	var body map[string]any
	if jsonErr := json.Unmarshal(respBody, &body); jsonErr == nil {
		if msg, ok := body["message"].(string); ok && msg != "" {
			return fmt.Sprintf("API error (%d): %s", err.StatusCode, msg)
		}
		if errMsg, ok := body["error"].(string); ok && errMsg != "" {
			return fmt.Sprintf("API error (%d): %s", err.StatusCode, errMsg)
		}
	}

	// Fallback to status code
	switch err.StatusCode {
	case 401:
		return "API error (401): Unauthorized - check your token"
	case 403:
		return "API error (403): Forbidden - insufficient permissions"
	case 404:
		return "API error (404): Not found"
	case 429:
		return "API error (429): Rate limited - try again later"
	default:
		return fmt.Sprintf("API error (%d)", err.StatusCode)
	}
}

// IsNotFound returns true if the error is a 404.
func IsNotFound(err error) bool {
	var apiErr *beeperdesktopapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsUnauthorized returns true if the error is a 401.
func IsUnauthorized(err error) bool {
	var apiErr *beeperdesktopapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}
