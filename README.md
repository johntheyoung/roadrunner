# 🐦💨 roadrunner — Beeper Desktop CLI

## Features

- **Chats** — list, search, resolve, get, create, archive conversations
- **Contacts** — search and resolve contacts on an account
- **Messages** — list, search, send, reply, tail (polling), wait, and context
- **Search** — global search across all chats and messages
- **Unread** — roll up unread chats across accounts
- **Status** — unread and chat summary (optional per-account)
- **Reminders** — set and clear chat reminders
- **Focus** — focus app window, pre-fill drafts with text or attachments
- **Scripting** — stdin/text-file input, `--fail-if-empty`, and `--fields` for plain output
- **Output** — JSON, plain (TSV), or human-readable formats

## Install

### Homebrew (macOS)

```bash
brew install johntheyoung/tap/roadrunner
```

### Go install

```bash
go install github.com/johntheyoung/roadrunner/cmd/rr@latest
```

### Binary download

Download from the [releases page](https://github.com/johntheyoung/roadrunner/releases).

## Requirements

- [Beeper Desktop](https://www.beeper.com/) **v4.1.169 or later** running locally

## Setup

1. Open Beeper Desktop → **Settings** → **Developers**
2. Toggle **Beeper Desktop API** to enable it (server starts at `localhost:23373`)
3. Click **+** next to "Approved connections" to create a token
4. Configure the CLI:

```bash
rr auth set <your-token>
rr doctor  # verify setup
```

Token is stored in `~/.config/beeper/config.json`. `BEEPER_TOKEN` env var overrides the config file.

## Chats

```bash
# List your chats
rr chats list

# Search chats by name
rr chats search "Alice"

# Resolve a chat by exact title or ID
rr chats resolve "Alice"

# Search by participant name (useful when chat title shows Matrix ID)
rr chats search "Alice" --scope=participants

# Filter by inbox and unread status
rr chats search --inbox=primary --unread-only

# Filter by activity date
rr chats search --last-activity-after="2024-07-01T00:00:00Z"

# Get chat details
rr chats get '!roomid:beeper.local'

# Create a new chat (single)
rr chats create "<account-id>" --participant "<user-id>"

# Create a group chat
rr chats create "<account-id>" \
  --participant "<user-a>" \
  --participant "<user-b>" \
  --type group \
  --title "Project Team" \
  --message "Welcome!"

# Archive a chat
rr chats archive '!roomid:beeper.local'

# Unarchive
rr chats archive '!roomid:beeper.local' --unarchive
```

## Contacts

```bash
# Search contacts on an account (positional or flag)
rr contacts search "<account-id>" "Alice"
rr contacts search "Alice" --account-id="<account-id>"

# Using default account (from --account or BEEPER_ACCOUNT)
rr contacts search "Alice"

# Resolve a contact by exact match
rr contacts resolve "<account-id>" "Alice"
rr contacts resolve "Alice" --account-id="<account-id>"

# If a name is ambiguous, resolve by ID
rr contacts search "Michael Johnson" --account-id="<account-id>" --json
rr contacts resolve "<contact-id>" --account-id="<account-id>" --json
```

## Messages

```bash
# List messages in a chat
rr messages list '!roomid:beeper.local'

# Search messages globally
rr messages search "meeting notes"

# Filter by sender
rr messages search --sender=me

# Filter by date range
rr messages search --date-after="2024-07-01T00:00:00Z" --date-before="2024-08-01T00:00:00Z"

# Filter by media type
rr messages search --media-types=image

# Send a message
rr messages send '!roomid:beeper.local' "Hello!"

# Reply to a specific message
rr messages send '!roomid:beeper.local' "Thanks!" --reply-to "<message-id>"

# Send message from a file
rr messages send '!roomid:beeper.local' --text-file ./message.txt

# Send message from stdin
cat message.txt | rr messages send '!roomid:beeper.local' --stdin

# Tail new messages (polling)
rr messages tail '!roomid:beeper.local' --interval 2s --stop-after 30s

# Tail with filters
rr messages tail '!roomid:beeper.local' --contains "deploy" --sender "Alice" --from "2024-07-01" --interval 5s

# Wait for a matching message
rr messages wait --chat-id='!roomid:beeper.local' --contains "deploy" --wait-timeout 2m

# Context around a message (by sortKey)
rr messages context '!roomid:beeper.local' '<sortKey>' --before 5 --after 2

# Download attachments from listed messages
rr messages list '!roomid:beeper.local' --download-media --download-dir ./media
```

## Search

```bash
# Global search across chats and messages
rr search "dinner plans"

# Paginate message results (max 20 per page)
rr search "project" --messages-limit=20
rr search "project" --messages-cursor="<cursor>" --messages-direction=before
```

Global search returns matching chats, messages, and an "In Groups" section for participant name matches.

## Status

```bash
# Summary of unread, muted, and archived chats
rr status

# Per-account summary
rr status --by-account
```

## Unread

```bash
# Roll up unread chats across all accounts
rr unread --json
```

## Reminders

```bash
# Set a reminder (relative time)
rr reminders set '!roomid:beeper.local' "2h"

# Set a reminder (absolute time)
rr reminders set '!roomid:beeper.local' "2024-12-25T09:00:00Z"

# Clear a reminder
rr reminders clear '!roomid:beeper.local'
```

## Focus & Drafts

```bash
# Focus Beeper Desktop window
rr focus

# Focus and open a specific chat
rr focus --chat-id='!roomid:beeper.local'

# Jump to a specific message
rr focus --chat-id='!roomid:beeper.local' --message-id="<message-id>"

# Pre-fill a draft (message not sent)
rr focus --chat-id='!roomid:beeper.local' --draft-text="Hello!"

# Pre-fill a draft from a file
rr focus --chat-id='!roomid:beeper.local' --draft-text-file ./draft.txt

# Pre-fill a draft with attachment
rr focus --chat-id='!roomid:beeper.local' --draft-attachment="/path/to/file.jpg"

# Combine draft text and attachment
rr focus --chat-id='!roomid:beeper.local' --draft-text="Check this out!" --draft-attachment="/path/to/file.jpg"
```

## Scripting

```bash
# Find a chat and send a message
CHAT_ID=$(rr chats search "Alice" --json | jq -r '.items[0].id')
rr messages send "$CHAT_ID" "Hey!"

# Resolve a chat by exact match
CHAT_ID=$(rr chats resolve "Alice" --json | jq -r '.id')
rr messages send "$CHAT_ID" "Hello!"

# Exit with code 1 if no results
rr chats search "Alice" --json --fail-if-empty

# List unread chats
rr chats search --inbox=primary --unread-only --json

# Set a reminder for 2 hours from now
rr reminders set "$CHAT_ID" "2h"

# Focus a chat and pre-fill a draft
rr focus --chat-id="$CHAT_ID" --draft-text="Hello!"

# Send a multi-line draft via stdin
cat draft.txt | rr messages send "$CHAT_ID" --stdin

# Download an attachment
rr assets download "mxc://beeper.local/abc123" --dest "./attachment.jpg"

# Search contacts on an account
rr contacts search "Alice" --account-id="<account-id>" --json
```

## Output Modes

### Human-readable (default)

```bash
$ rr chats list
CHAT                              LAST MESSAGE           TIME
Alice                             See you tomorrow!      2h ago
Project Team                      Meeting at 3pm         5h ago
```

### JSON (for scripting)

```bash
$ rr chats list --json
{
  "items": [
    {
      "id": "!abc123:beeper.local",
      "display_name": "Alice",
      "last_message": "See you tomorrow!",
      ...
    }
  ]
}
```

Message JSON includes `is_sender`, `is_unread`, `attachments`, and `reactions`.
`downloaded_attachments` is only populated when `--download-media` is used.

### Plain (TSV)

```bash
$ rr chats list --plain
!abc123:beeper.local	Alice	See you tomorrow!
!def456:beeper.local	Project Team	Meeting at 3pm

# Select fields for plain output
rr chats list --plain --fields id,title
```

JSON and plain output go to stdout. Errors and hints go to stderr.

Commands supporting `--plain --fields`:
- `accounts list`, `chats list/search/resolve`, `messages list/search`
- `contacts search/resolve`, `search`, `unread`
- `status`, `doctor`, `auth status`, `version`

### Envelope Mode

Wrap JSON output in a consistent structure for easier error handling:

```bash
$ rr chats list --json --envelope
{
  "success": true,
  "data": { "items": [...] },
  "metadata": {
    "timestamp": "2026-01-18T12:00:00Z",
    "version": "0.8.0",
    "command": "chats list"
  }
}

$ rr chats get "invalid" --json --envelope
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "API error (404): Chat not found"
  },
  "metadata": { ... }
}
```

Error codes: `AUTH_ERROR`, `NOT_FOUND`, `VALIDATION_ERROR`, `CONNECTION_ERROR`, `INTERNAL_ERROR`.

## Agent Safety

Restrict CLI capabilities when used by AI agents or in sandboxed environments:

```bash
# Only allow specific commands
rr --enable-commands=chats,messages,status chats list

# Block data writes (send, create, archive, reminders)
rr --readonly messages list '!roomid:beeper.local'

# Combine for read-only access to specific commands
rr --enable-commands=chats,messages --readonly chats search "Alice"
```

Write commands blocked by `--readonly`: `messages send`, `chats create`, `chats archive`, `reminders set`, `reminders clear`.

Exemptions: `auth set`, `auth clear`, and `focus` are always allowed (local-only operations).

### Agent Profile Mode

For AI agent integrations, use `--agent` to enable a hardened profile:

```bash
# Agent mode forces JSON+envelope, no-input, readonly
# and requires --enable-commands for safety
rr --agent --enable-commands=chats,messages,status chats list
```

Agent mode automatically sets: `--json`, `--envelope`, `--no-input`, `--readonly`.

The `--enable-commands` flag is **required** in agent mode to ensure agents only access explicitly allowed commands.

### Capability Discovery

```bash
$ rr version --json
{
  "version": "0.9.0",
  "features": ["enable-commands", "readonly", "envelope", "agent-mode"]
}
```

For detailed capability discovery:

```bash
$ rr capabilities --json
{
  "version": "0.9.0",
  "features": ["enable-commands", "readonly", "envelope", "agent-mode"],
  "defaults": { "timeout": 30, "base_url": "http://localhost:23373" },
  "output_modes": ["human", "json", "plain"],
  "safety": {
    "enable_commands_desc": "Comma-separated allowlist of top-level commands",
    "readonly_desc": "Block data write operations",
    "agent_desc": "Agent profile: forces JSON, envelope, no-input, readonly; requires --enable-commands"
  },
  "commands": {
    "read": ["accounts list", "chats list", "messages list", ...],
    "write": ["messages send", "chats create", "chats archive", ...],
    "exempt": ["auth set", "auth clear", "focus"]
  },
  "flags": { ... }
}
```

Agents can check `features` to detect supported safety flags before use.

## Multi-Account Usage

By default, commands search **all accounts**. Use `--account` to focus on one.

```bash
# All accounts (default behavior)
rr chats list
rr messages search "dinner"

# Single account (optional)
rr --account="imessage:+1234567890" chats list

# Or via environment variable
export BEEPER_ACCOUNT="imessage:+1234567890"
rr chats list            # now defaults to imessage
rr chats list --account-ids=telegram,whatsapp  # explicit overrides default
```

For `contacts` commands (which require an account), `--account` provides a default:

```bash
rr contacts search "Alice"                      # uses BEEPER_ACCOUNT
rr contacts search "Alice" --account-id=slack   # explicit override
```

### Account Aliases

Create short aliases for frequently used accounts:

```bash
# Set an alias
rr accounts alias set work "slack:T12345678"
rr accounts alias set personal "imessage:+1234567890"

# List aliases
rr accounts alias list

# Remove an alias
rr accounts alias unset work
```

Aliases can be used anywhere an account ID is expected:

```bash
rr --account=work chats list
rr contacts search "Alice" --account-id=work
rr chats create work --participant "<user-id>"
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
| `BEEPER_URL` | API base URL (default: `http://localhost:23373`) |
| `BEEPER_TIMEOUT` | API timeout in seconds (0 disables) |
| `BEEPER_COLOR` | Color mode: `auto` \| `always` \| `never` |
| `BEEPER_JSON` | Default to JSON output |
| `BEEPER_PLAIN` | Default to plain output |
| `BEEPER_NO_INPUT` | Never prompt, fail instead |
| `BEEPER_HELP` | Set to `full` for expanded help |
| `BEEPER_ENABLE_COMMANDS` | Comma-separated allowlist of commands |
| `BEEPER_READONLY` | Block data write operations |
| `BEEPER_ENVELOPE` | Wrap JSON in envelope structure |
| `BEEPER_AGENT` | Enable agent profile mode |
| `BEEPER_ACCOUNT` | Default account ID for commands |
| `NO_COLOR` | Disable colored output |

## Shell Notes

In bash/zsh, `!` triggers history expansion. If you see `\!` in text or chat IDs, it came from your shell/runner.

**Solutions:**
- Use single quotes: `rr chats get '!roomid:beeper.local'`
- Disable history expansion: `set +H` (bash) or `setopt NO_HIST_EXPAND` (zsh)

## Non-interactive Mode

Destructive commands require confirmation. In non-interactive environments (no TTY, or `--no-input` / `BEEPER_NO_INPUT`), commands fail unless `--force` is provided.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error |

## AI Agent Skill

The ClawdHub skill definition lives in `./skill/SKILL.md`. This is the single source published to ClawdHub.

## Links

- [Desktop API Documentation](https://developers.beeper.com/desktop-api)
- [Troubleshooting Guide](docs/troubleshooting.md)
- [API Notes](docs/api-notes.md)

## License

MIT
