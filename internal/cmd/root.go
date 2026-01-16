package cmd

import (
	"github.com/alecthomas/kong"
)

// Version is set at build time via ldflags.
var Version = "dev"

type RootFlags struct {
	Color   string            `help:"Color output: auto|always|never" default:"auto" env:"BEEPER_COLOR"`
	JSON    bool              `help:"Output JSON to stdout (best for scripting)" env:"BEEPER_JSON"`
	Plain   bool              `help:"Output stable TSV to stdout (no colors)" env:"BEEPER_PLAIN"`
	Verbose bool              `help:"Enable debug logging" short:"v"`
	NoInput bool              `help:"Never prompt; fail instead (useful for CI)" env:"BEEPER_NO_INPUT"`
	Force   bool              `help:"Skip confirmations for destructive commands" short:"f"`
	Timeout int               `help:"Timeout for API calls in seconds (0=none)" default:"30" env:"BEEPER_TIMEOUT"`
	BaseURL string            `help:"API base URL" default:"http://localhost:23373" env:"BEEPER_URL"`
	Version kong.VersionFlag  `help:"Show version and exit"`
}

type CLI struct {
	RootFlags

	Version VersionCmd `cmd:"" help:"Show version information"`
}

func Execute() int {
	cli := &CLI{}
	ctx := kong.Parse(cli,
		kong.Name("rr"),
		kong.Description("CLI for Beeper Desktop. Beep beep!"),
		kong.UsageOnError(),
		kong.Vars{"version": Version},
	)

	if err := ctx.Run(&cli.RootFlags); err != nil {
		return 1
	}

	return 0
}
