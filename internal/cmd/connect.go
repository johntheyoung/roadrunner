package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// ConnectCmd is the parent command for connect/discovery subcommands.
type ConnectCmd struct {
	Info ConnectInfoCmd `cmd:"" help:"Show Connect server metadata and discovered endpoints"`
}

// ConnectInfoCmd retrieves server metadata from GET /v1/info.
type ConnectInfoCmd struct {
	Fields []string `help:"Comma-separated list of fields for --plain output" name:"fields" sep:","`
}

// Run executes the connect info command.
func (c *ConnectInfoCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	info, err := client.Connect().Info(ctx)
	if err != nil {
		if beeperapi.IsUnsupportedRoute(err, "GET", "/v1/info") {
			return errUnsupportedConnectInfo()
		}
		return err
	}

	introspectionURL := endpointByHint(info.Endpoints, "oauth_introspect", "introspect")
	wsURL := endpointByHint(info.Endpoints, "ws", "websocket")

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, info, "connect info")
	}

	if outfmt.IsPlain(ctx) {
		fields, err := resolveFields(c.Fields, []string{"name", "version", "runtime", "oauth_introspect_url", "ws_url"})
		if err != nil {
			return err
		}
		writePlainFields(u, fields, map[string]string{
			"name":                 info.Name,
			"version":              info.Version,
			"runtime":              info.Runtime,
			"oauth_introspect_url": introspectionURL,
			"ws_url":               wsURL,
		})
		return nil
	}

	u.Out().Printf("Connect Metadata")
	if info.Name != "" {
		u.Out().Printf("Name:    %s", info.Name)
	}
	if info.Version != "" {
		u.Out().Printf("Version: %s", info.Version)
	}
	if info.Runtime != "" {
		u.Out().Printf("Runtime: %s", info.Runtime)
	}
	if len(info.Endpoints) == 0 {
		u.Out().Dim("No endpoint metadata exposed by this server.")
		return nil
	}

	keys := make([]string, 0, len(info.Endpoints))
	for key := range info.Endpoints {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	u.Out().Println("")
	u.Out().Printf("Endpoints:")
	for _, key := range keys {
		u.Out().Printf("  %s: %s", key, info.Endpoints[key])
	}

	return nil
}

func endpointByHint(endpoints map[string]string, preferred string, contains string) string {
	if len(endpoints) == 0 {
		return ""
	}
	if preferred != "" {
		if value := strings.TrimSpace(endpoints[preferred]); value != "" {
			return value
		}
	}
	if contains == "" {
		return ""
	}
	needle := strings.ToLower(contains)
	for key, value := range endpoints {
		if strings.Contains(strings.ToLower(key), needle) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func errUnsupportedConnectInfo() error {
	return fmt.Errorf("connect info is not supported by this Beeper Desktop API version (requires a newer Beeper Desktop build)")
}
