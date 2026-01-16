# Roadrunner (rr)

Command-line interface for Beeper messaging. Beep beep!

## Setup

Requires Beeper Desktop running. Set your token:
```bash
rr auth set <token>
```

Or use environment variable: `export BEEPER_TOKEN=<token>`

## Commands

### List accounts
```bash
rr accounts list --json
```

### List chats (with pagination)
```bash
# First page
rr chats list --json

# Next page using cursor from previous response
rr chats list --cursor="<oldestCursor>" --direction=before --json
```

### Search chats
```bash
# Find chats by name
rr chats search "John" --json

# Filter by inbox type
rr chats search --inbox=primary --unread-only --json

# Get specific chat
rr chats get "!chatid:beeper.com" --json
```

### Global search
Note: Message search is literal word match, not semantic.

```bash
# Search across chats and messages
rr search "dinner" --json

# Paginate message results (max 20)
rr search "dinner" --messages-limit=20 --json
rr search "dinner" --messages-cursor="<cursor>" --messages-direction=before --json
```

### Send message
```bash
# Send to chat by ID
rr messages send "!chatid:beeper.com" "Hello!"

# Returns pending_message_id (temporary ID during delivery)
# {"chat_id": "!chatid:beeper.com", "pending_message_id": "pending_msg_123"}

# Compose with search (find chat, then send)
rr messages send "$(rr chats search "John" --json | jq -r '.items[0].id')" "Hi John!"
```

### List messages (with pagination)
```bash
rr messages list "!chatid:beeper.com" --json

# Paginate through older messages (use sortKey from last item as cursor)
rr messages list "!chatid:beeper.com" --cursor="<sortKey>" --direction=before --json
```

### Search messages
**Note:** Search is literal word match, NOT semantic. All words must match exactly.

```bash
# Use single words, not phrases
rr messages search "dinner" --json

# Filter by chat
rr messages search "meeting" --chat-id="!chatid:beeper.com" --json

# Paginate results
rr messages search "project" --limit=20 --json
rr messages search "project" --cursor="<cursor>" --direction=before --json
```

### Reminders
```bash
# Set reminder (relative time)
rr reminders set "!chatid:beeper.com" "2h"

# Set reminder (absolute time)
rr reminders set "!chatid:beeper.com" "2024-01-20T15:00:00"

# Clear reminder
rr reminders clear "!chatid:beeper.com"
```

### Archive/unarchive chat
```bash
rr chats archive "!chatid:beeper.com"
rr chats archive "!chatid:beeper.com" --unarchive
```

### Focus Beeper Desktop
```bash
# Just focus the app
rr focus

# Focus on specific chat
rr focus --chat-id="!chatid:beeper.com"

# Focus and pre-fill draft
rr focus --chat-id="!chatid:beeper.com" --draft-text="Hello!"
```

## Output Modes

- Default: Human-readable colored output
- `--json`: Machine-readable JSON to stdout (best for scripting/AI)
- `--plain`: Tab-separated values to stdout (no colors)

## Composing Commands

Chat IDs are used directly. To find a chat ID, search first:

```bash
# Find chat ID for "John"
rr chats search "John" --json | jq -r '.items[0].id'

# One-liner: search and send
CHAT_ID=$(rr chats search "John" --json | jq -r '.items[0].id')
rr messages send "$CHAT_ID" "Hey!"
```

## Diagnostics

```bash
# Check configuration and connectivity
rr doctor

# Validate token is working
rr auth status --check
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (bad flags, missing args) |
