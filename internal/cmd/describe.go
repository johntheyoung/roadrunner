package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// DescribeCmd describes command schema for runtime introspection.
type DescribeCmd struct {
	CommandPath []string `arg:"" optional:"" name:"command" help:"Command path to describe (example: messages send)"`
}

type DescribeResponse struct {
	Command     string               `json:"command"`
	Path        []string             `json:"path,omitempty"`
	Kind        string               `json:"kind"`
	Help        string               `json:"help,omitempty"`
	Detail      string               `json:"detail,omitempty"`
	Aliases     []string             `json:"aliases,omitempty"`
	Positionals []DescribePositional `json:"positionals,omitempty"`
	Flags       []DescribeFlag       `json:"flags,omitempty"`
	Subcommands []DescribeSubcommand `json:"subcommands,omitempty"`
	Safety      *DescribeSafety      `json:"safety,omitempty"`
}

type DescribePositional struct {
	Name     string   `json:"name"`
	Help     string   `json:"help,omitempty"`
	Required bool     `json:"required"`
	Default  string   `json:"default,omitempty"`
	Enum     []string `json:"enum,omitempty"`
	Type     string   `json:"type,omitempty"`
}

type DescribeFlag struct {
	Name      string   `json:"name"`
	Long      string   `json:"long"`
	Short     string   `json:"short,omitempty"`
	Help      string   `json:"help,omitempty"`
	Required  bool     `json:"required"`
	Default   string   `json:"default,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Envs      []string `json:"env,omitempty"`
	Type      string   `json:"type,omitempty"`
	Inherited bool     `json:"inherited,omitempty"`
}

type DescribeSubcommand struct {
	Name    string   `json:"name"`
	Help    string   `json:"help,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type DescribeSafety struct {
	ReadonlyBlocked bool   `json:"readonly_blocked"`
	ReadonlyExempt  bool   `json:"readonly_exempt"`
	RetryClass      string `json:"retry_class,omitempty"`
}

// Run executes the describe command.
func (c *DescribeCmd) Run(ctx context.Context) error {
	u := ui.FromContext(ctx)

	parser, err := newDescribeParser()
	if err != nil {
		return fmt.Errorf("build parser model: %w", err)
	}

	requested := normalizeDescribePath(c.CommandPath)
	node, resolved := resolveCommandPath(parser.Model.Node, requested)
	if node == nil {
		return errfmt.UsageError("unknown command path %q", strings.Join(requested, " "))
	}

	resp := buildDescribeResponse(node, resolved)
	command := "describe"
	if len(resolved) > 0 {
		command = command + " " + strings.Join(resolved, " ")
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, command)
	}

	if outfmt.IsPlain(ctx) {
		for _, sub := range resp.Subcommands {
			u.Out().Printf("%s\t%s", sub.Name, sub.Help)
		}
		return nil
	}

	u.Out().Printf("Command: %s", resp.Command)
	u.Out().Printf("Kind:    %s", resp.Kind)
	if resp.Help != "" {
		u.Out().Printf("Help:    %s", resp.Help)
	}
	if len(resp.Subcommands) > 0 {
		u.Out().Println("")
		u.Out().Printf("Subcommands:")
		for _, sub := range resp.Subcommands {
			u.Out().Printf("  - %s", sub.Name)
		}
	}
	if len(resp.Positionals) > 0 {
		u.Out().Println("")
		u.Out().Printf("Positionals:")
		for _, p := range resp.Positionals {
			u.Out().Printf("  - %s", p.Name)
		}
	}
	if len(resp.Flags) > 0 {
		u.Out().Println("")
		u.Out().Printf("Flags:")
		for _, f := range resp.Flags {
			u.Out().Printf("  - %s", f.Long)
		}
	}

	return nil
}

func newDescribeParser() (*kong.Kong, error) {
	model := &CLI{}
	helpCompact := os.Getenv("BEEPER_HELP") != "full"
	return kong.New(model,
		kong.Name("rr"),
		kong.Description("CLI for Beeper Desktop. Beep beep!"),
		kong.UsageOnError(),
		kong.Vars{"version": VersionString()},
		kong.ConfigureHelp(kong.HelpOptions{Compact: helpCompact}),
	)
}

func normalizeDescribePath(parts []string) []string {
	var out []string
	for _, part := range parts {
		for _, token := range strings.Fields(part) {
			out = append(out, token)
		}
	}
	return out
}

func resolveCommandPath(root *kong.Node, path []string) (*kong.Node, []string) {
	if root == nil {
		return nil, nil
	}
	if len(path) == 0 {
		return root, nil
	}

	node := root
	resolved := make([]string, 0, len(path))
	for _, part := range path {
		var next *kong.Node
		for _, child := range node.Children {
			if child.Name == part || hasAlias(child.Aliases, part) {
				next = child
				break
			}
		}
		if next == nil {
			return nil, resolved
		}
		node = next
		resolved = append(resolved, next.Name)
	}

	return node, resolved
}

func hasAlias(aliases []string, name string) bool {
	for _, alias := range aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func buildDescribeResponse(node *kong.Node, resolved []string) DescribeResponse {
	resp := DescribeResponse{
		Command:     "rr",
		Path:        append([]string{}, resolved...),
		Kind:        nodeKind(node),
		Help:        strings.TrimSpace(node.Help),
		Detail:      strings.TrimSpace(node.Detail),
		Aliases:     append([]string{}, node.Aliases...),
		Positionals: describePositionals(node),
		Flags:       describeFlags(node),
		Subcommands: describeSubcommands(node),
	}

	if len(resolved) > 0 {
		commandPath := strings.Join(resolved, " ")
		resp.Command = commandPath
		resp.Safety = describeSafety(commandPath)
	}

	return resp
}

func nodeKind(node *kong.Node) string {
	switch node.Type {
	case kong.ApplicationNode:
		return "application"
	case kong.CommandNode:
		return "command"
	case kong.ArgumentNode:
		return "argument"
	default:
		return "unknown"
	}
}

func describePositionals(node *kong.Node) []DescribePositional {
	out := make([]DescribePositional, 0, len(node.Positional))
	for _, p := range node.Positional {
		pos := DescribePositional{
			Name:     p.Name,
			Help:     strings.TrimSpace(p.Help),
			Required: p.Required,
			Default:  defaultValue(p.HasDefault, p.Default),
			Enum:     compactStrings(p.EnumSlice()),
		}
		if p.Target.IsValid() {
			pos.Type = p.Target.Type().String()
		}
		out = append(out, pos)
	}
	return out
}

func describeFlags(node *kong.Node) []DescribeFlag {
	chain := nodeChain(node)
	seen := map[string]bool{}
	out := make([]DescribeFlag, 0)

	for i, n := range chain {
		inherited := i < len(chain)-1
		for _, flag := range n.Flags {
			if flag.Hidden {
				continue
			}
			if seen[flag.Name] {
				continue
			}
			seen[flag.Name] = true

			desc := DescribeFlag{
				Name:      flag.Name,
				Long:      "--" + flag.Name,
				Help:      strings.TrimSpace(flag.Help),
				Required:  flag.Required,
				Default:   defaultValue(flag.HasDefault, flag.Default),
				Enum:      compactStrings(flag.EnumSlice()),
				Envs:      compactStrings(flag.Envs),
				Inherited: inherited,
			}
			if flag.Short != 0 {
				desc.Short = "-" + string(flag.Short)
			}
			if flag.Target.IsValid() {
				desc.Type = flag.Target.Type().String()
			}
			out = append(out, desc)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}

func nodeChain(node *kong.Node) []*kong.Node {
	var reversed []*kong.Node
	for n := node; n != nil; n = n.Parent {
		reversed = append(reversed, n)
	}
	out := make([]*kong.Node, 0, len(reversed))
	for i := len(reversed) - 1; i >= 0; i-- {
		out = append(out, reversed[i])
	}
	return out
}

func describeSubcommands(node *kong.Node) []DescribeSubcommand {
	out := make([]DescribeSubcommand, 0, len(node.Children))
	for _, child := range node.Children {
		if child.Hidden || child.Type != kong.CommandNode {
			continue
		}
		out = append(out, DescribeSubcommand{
			Name:    child.Name,
			Help:    strings.TrimSpace(child.Help),
			Aliases: append([]string{}, child.Aliases...),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func describeSafety(commandPath string) *DescribeSafety {
	retryClass := retryClasses()[commandPath]
	if retryClass == "" {
		retryClass = "safe"
	}

	return &DescribeSafety{
		ReadonlyBlocked: dataWriteCommands[commandPath],
		ReadonlyExempt:  exemptCommands[commandPath],
		RetryClass:      retryClass,
	}
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func defaultValue(hasDefault bool, value string) string {
	if !hasDefault {
		return ""
	}
	return strings.TrimSpace(value)
}
