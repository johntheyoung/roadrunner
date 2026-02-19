package beeperapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/beeper/desktop-api-go/option"
)

// ConnectService handles connect discovery and token introspection endpoints.
type ConnectService struct {
	client *Client
}

// ConnectInfo is metadata returned by GET /v1/info.
type ConnectInfo struct {
	Name      string            `json:"name,omitempty"`
	Version   string            `json:"version,omitempty"`
	Runtime   string            `json:"runtime,omitempty"`
	Endpoints map[string]string `json:"endpoints,omitempty"`
	Raw       map[string]any    `json:"raw,omitempty"`
}

// TokenIntrospection is metadata returned by POST /oauth/introspect.
type TokenIntrospection struct {
	Active    bool           `json:"active"`
	Scope     string         `json:"scope,omitempty"`
	ClientID  string         `json:"client_id,omitempty"`
	Subject   string         `json:"subject,omitempty"`
	Username  string         `json:"username,omitempty"`
	TokenType string         `json:"token_type,omitempty"`
	Issuer    string         `json:"issuer,omitempty"`
	ExpiresAt int64          `json:"expires_at,omitempty"`
	IssuedAt  int64          `json:"issued_at,omitempty"`
	NotBefore int64          `json:"not_before,omitempty"`
	Raw       map[string]any `json:"raw,omitempty"`
}

// Info retrieves connect metadata for the running Beeper Desktop API server.
func (s *ConnectService) Info(ctx context.Context) (ConnectInfo, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	var raw map[string]any
	if err := s.client.SDK.Get(ctx, "v1/info", nil, &raw); err != nil {
		return ConnectInfo{}, err
	}
	if raw == nil {
		raw = map[string]any{}
	}

	info := ConnectInfo{
		Name:      firstString(raw, "name", "appName", "title"),
		Version:   firstString(raw, "version", "appVersion"),
		Runtime:   firstString(raw, "runtime", "environment"),
		Endpoints: map[string]string{},
		Raw:       raw,
	}

	if info.Name == "" {
		if app, ok := asMap(raw["app"]); ok {
			info.Name = firstString(app, "name", "title")
			if info.Version == "" {
				info.Version = firstString(app, "version")
			}
		}
	}

	if endpoints, ok := asMap(raw["endpoints"]); ok {
		for k, v := range endpoints {
			if str, ok := asString(v); ok {
				info.Endpoints[k] = str
			}
		}
	}

	// Some servers expose URL fields at the top-level instead of under endpoints.
	for k, v := range raw {
		if str, ok := asString(v); ok && strings.Contains(strings.ToLower(k), "url") {
			if _, exists := info.Endpoints[k]; !exists {
				info.Endpoints[k] = str
			}
		}
	}

	if len(info.Endpoints) == 0 {
		info.Endpoints = nil
	}

	return info, nil
}

// Introspect checks token activity and metadata via OAuth introspection.
func (s *ConnectService) Introspect(ctx context.Context, token string) (TokenIntrospection, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	token = strings.TrimSpace(token)
	if token == "" {
		return TokenIntrospection{}, fmt.Errorf("token is required for introspection")
	}

	form := url.Values{}
	form.Set("token", token)
	form.Set("token_type_hint", "access_token")

	var raw map[string]any
	if err := s.client.SDK.Post(
		ctx,
		"oauth/introspect",
		[]byte(form.Encode()),
		&raw,
		option.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	); err != nil {
		return TokenIntrospection{}, err
	}
	if raw == nil {
		raw = map[string]any{}
	}

	result := TokenIntrospection{
		Active:    boolValue(raw["active"]),
		Scope:     firstString(raw, "scope"),
		ClientID:  firstString(raw, "client_id"),
		Subject:   firstString(raw, "sub", "subject"),
		Username:  firstString(raw, "username"),
		TokenType: firstString(raw, "token_type"),
		Issuer:    firstString(raw, "iss", "issuer"),
		ExpiresAt: int64Value(raw["exp"]),
		IssuedAt:  int64Value(raw["iat"]),
		NotBefore: int64Value(raw["nbf"]),
		Raw:       raw,
	}

	return result, nil
}

func asMap(value any) (map[string]any, bool) {
	m, ok := value.(map[string]any)
	return m, ok
}

func asString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", false
		}
		return v, true
	default:
		return "", false
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := asString(m[key]); ok {
			return value
		}
	}
	return ""
}

func boolValue(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		return lower == "true" || lower == "1" || lower == "yes"
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	default:
		return false
	}
}

func int64Value(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case int32:
		return int64(v)
	case json.Number:
		n, _ := strconv.ParseInt(v.String(), 10, 64)
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}
