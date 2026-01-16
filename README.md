# roadrunner

CLI for Beeper Desktop. Beep beep! 🏃

## Install

```bash
go install github.com/johntheyoung/roadrunner/cmd/rr@latest
```

## Usage

```bash
# Check setup
rr doctor

# List accounts
rr accounts list --json

# Search chats
rr chats search "John" --json

# Send message
rr messages send "!chatid:beeper.com" "Hello!"
```

## Requirements

- [Beeper Desktop](https://www.beeper.com/) running locally
- API token from Beeper Desktop settings

## Setup

```bash
# Set your token
rr auth set <token>

# Verify setup
rr doctor
```

## License

MIT
