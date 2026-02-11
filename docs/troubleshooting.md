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
