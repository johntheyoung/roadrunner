# Troubleshooting

## rr doctor fails to connect

- Ensure Beeper Desktop is running.
- Confirm the base URL (default is http://localhost:23373).
- If you changed the port, set `BEEPER_URL` or pass `--base-url`.

## Token errors

- If you see "no token configured", run `rr auth set <token>`.
- `BEEPER_TOKEN` overrides the config file.
- Use `rr auth status --check` to validate the token.

## Non-interactive failures

- Destructive commands require confirmation.
- In CI or scripts, pass `--force` explicitly.
- `--no-input`/`BEEPER_NO_INPUT` disables prompts and will fail without `--force`.

## Unsupported route errors for edit/upload commands

- If you see `message editing is not supported` or `asset upload is not supported`,
  your Beeper Desktop build is older than the required Desktop API routes.
- Update Beeper Desktop and retry `rr doctor`, then re-run the command.

## Attachment send validation errors

- `--attachment-upload-id` is required when using attachment metadata override flags.
- `--attachment-width` and `--attachment-height` must be provided together.
- `rr messages send` requires either message text or `--attachment-upload-id`.
- `rr search --messages-max-items` requires `--messages-all`.

## assets serve output mode errors

- `rr assets serve --json` and `rr assets serve --plain` require `--dest`.
- Without `--dest`, `assets serve` streams raw bytes to stdout.
- To intentionally stream to an interactive terminal, pass `--stdout`.

## Name-based chat targeting issues

- `--chat` is exact match only (chat title/display name/ID).
- If multiple chats match the same text, the command fails with `multiple chats matched`.
- Narrow with an exact chat ID or set `--account`/`BEEPER_ACCOUNT` to reduce ambiguity.

## Agent integration checks

- Run `make test-agent-smoke` to validate agent-mode safety and envelope contracts locally.
- In envelope mode, inspect `error.hint` for deterministic next-step remediation.
- Set `--request-id` to correlate repeated attempts in logs and envelope `metadata.request_id`.

## Duplicate write blocking

- If you see `duplicate non-idempotent request blocked`, the same `--request-id` and payload was replayed within `--dedupe-window`.
- Use a new `--request-id` for deliberate retries, or `--force` to bypass the dedupe guard.
