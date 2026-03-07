# Troubleshooting

## rr doctor fails to connect

- Ensure Beeper Desktop is running.
- Confirm the base URL (default is http://localhost:23373).
- If you changed the port, set `BEEPER_URL` or pass `--base-url`.

## Stable Beeper AppImage on Linux

- In this setup, the preferred stable install is a local AppImage at `/home/jmy/apps/beeper/stable/Beeper-latest.AppImage`.
- The local launcher override is `~/.local/share/applications/beeper.desktop`.
- If Beeper's in-app `Download update` button appears to do nothing, do not rely on it. Replace the local AppImage instead and relaunch.
- Verify the live build with `rr connect info --json` or `curl -fsS http://localhost:23373/v1/info`.
- If you previously mixed package-managed Beeper, stable AppImages, and Nightly AppImages, remove duplicate launcher entries before debugging API issues.

## Stable headless service

- The stable headless service is `beeper.service`.
- Start it with:

```bash
systemctl --user start beeper.service
```

- Stop it with:

```bash
systemctl --user stop beeper.service
```

- Verify the stable API is up on the default port:

```bash
ss -ltnp | rg 23373
curl -fsS http://localhost:23373/v1/info
rr doctor --json
```

- After `systemctl --user restart beeper.service`, give Beeper a few seconds to rebind the API socket before treating a failed `rr doctor` as a real problem.
- In this setup, stable Beeper stores Desktop API settings in the main profile at `~/.config/BeeperTexts`.
- After removing Nightly, stable Beeper should own `localhost:23373`.

## Beeper Nightly launcher or API conflicts

If Nightly appears to "not open", or `rr` connects to the wrong instance, you likely have multiple Beeper instances or launcher entries fighting over the same single-instance lock and API port.

- Keep one local launcher entry for Nightly (for example `~/.local/share/applications/beeper-nightly.desktop`).
- Disable duplicate Nightly entries such as `~/.local/share/applications/beepertexts.desktop` if both point to AppImage builds.
- Official and Nightly can coexist on disk, but only one active Beeper instance should own `localhost:23373` at a time.
- If you use a headless Nightly unit, stop it before opening GUI Nightly:

```bash
systemctl --user stop beeper-nightly.service
```

- Launch Nightly GUI explicitly:

```bash
gtk-launch beeper-nightly
```

- Verify expected instance and API endpoint:

```bash
rr connect info --json
ss -ltnp | rg 23373
```

- For stable Nightly testing, use a separate profile directory:
  `--user-data-dir=$HOME/.config/BeeperTexts-Nightly`

## Token errors

- If you see "no token configured", run `rr auth set '' --stdin` (recommended) or `rr auth set <token>`.
- `BEEPER_TOKEN` overrides the config file.
- Use `rr auth status --check` to validate the token.
- If `rr auth set --stdin` returns `expected "<token>"`, use `rr auth set '' --stdin` (current parser workaround).

Tip: avoid putting tokens in your shell history:

```bash
rr auth set '' --stdin
```

## Non-interactive failures

- Destructive commands require confirmation.
- In CI or scripts, pass `--force` explicitly.
- `--no-input`/`BEEPER_NO_INPUT` disables prompts and will fail without `--force`.

## Unsupported route errors for edit/upload commands

- If you see `message editing is not supported` or `asset upload is not supported`,
  your Beeper Desktop build is older than the required Desktop API routes.
- Update Beeper Desktop and retry `rr doctor`, then re-run the command.
- On stable Beeper `4.2.605`, send, react/unreact, and websocket events were validated, but `messages edit` still returned unsupported. Treat edit support as build-dependent and verify against your current Desktop version.

## `events tail` websocket issues

- If you see `websocket events are not supported`, your Beeper Desktop build does not expose `/v1/ws` yet. Use `rr messages tail` or upgrade Desktop.
- If you see `repeated read on failed websocket connection`, you are likely on `v0.16.1`; upgrade to a build that includes the websocket idle-timeout fix.
- If your stream reconnects repeatedly, verify local connectivity and auth first (`rr auth status --check`, `rr doctor`).
- `rr events tail` reconnects by default; disable with `--reconnect=false` if you want failures surfaced immediately.
- For deterministic CI/script runs, bound runtime with `--stop-after` and use `--include-control` when you need subscription/control diagnostics.

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

## Refusing to send pasted rr output

To prevent accidental privacy leaks, `rr messages send` / `send-file` / `edit` refuse message text that looks like pasted `rr --json` output.

Fix: remove the pasted output from the message text, or pass `--allow-tool-output` if you really intend to send it.

## Message text contains "$" (e.g. $100 becomes 00)

If you pass message text in double quotes, your shell may expand `$100` before `rr` sees it (e.g. `$100/month` can become `00/month`).

Fix: send message text via stdin or single quotes:

```bash
rr messages send "<chat-id>" --stdin <<'EOF'
Cost is $100/month
EOF
```

Or escape the dollar sign:

```bash
rr messages send "<chat-id>" "Cost is \\$100/month"
```
