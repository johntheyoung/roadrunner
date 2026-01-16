# roadrunner

CLI for Beeper Desktop. Beep beep!

## Install

```bash
go install github.com/johntheyoung/roadrunner/cmd/rr@latest
```

Or download a binary from the [releases page](https://github.com/johntheyoung/roadrunner/releases).

## Requirements

- [Beeper Desktop](https://www.beeper.com/) running locally
- API token from Beeper Desktop settings

## Quick Start

```bash
# Set your token (get it from Beeper Desktop settings)
rr auth set <token>

# Verify setup
rr doctor

# List your chats
rr chats list

# Search for a chat
rr chats search "John"

# Send a message
rr messages send "!chatid:beeper.com" "Hello!"
```

## Commands

| Command | Description |
|---------|-------------|
| `rr auth set/status/clear` | Manage authentication |
| `rr accounts list` | List messaging accounts |
| `rr chats list/search/get/archive` | Manage chats |
| `rr messages list/search/send` | Manage messages |
| `rr reminders set/clear` | Manage chat reminders |
| `rr search <query>` | Global search |
| `rr focus` | Focus Beeper Desktop |
| `rr doctor` | Diagnose configuration |
| `rr completion <shell>` | Generate shell completions |

Run `rr --help` or `rr <command> --help` for details.

## Output Modes

```bash
# Human-readable (default)
rr chats list

# JSON for scripting
rr chats list --json

# Plain TSV
rr chats list --plain
```

## Scripting Examples

```bash
# Find chat ID and send message
CHAT_ID=$(rr chats search "John" --json | jq -r '.items[0].id')
rr messages send "$CHAT_ID" "Hey!"

# List unread chats
rr chats search --inbox=primary --unread-only --json

# Set a reminder for 2 hours from now
rr reminders set "$CHAT_ID" "2h"
```

## Shell Completions

```bash
# Bash
rr completion bash >> ~/.bashrc

# Zsh
rr completion zsh >> ~/.zshrc

# Fish
rr completion fish > ~/.config/fish/completions/rr.fish
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `BEEPER_TOKEN` | API token (overrides config) |
| `BEEPER_URL` | API base URL (default: http://localhost:23373) |
| `BEEPER_JSON` | Default to JSON output |
| `BEEPER_PLAIN` | Default to plain output |
| `BEEPER_NO_INPUT` | Never prompt, fail instead |
| `NO_COLOR` | Disable colored output |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error |

## License

MIT
