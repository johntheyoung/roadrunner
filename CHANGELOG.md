# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

## v0.16.0 - 2026-02-19

### Added
- Experimental websocket live event support via `rr events tail` (connects to `GET /v1/ws`, sends `subscriptions.set`, supports `--all` or repeated `--chat-id` selectors).

### Changed
- Shell completion metadata, root command registration, and capabilities retry/read lists now include `events tail`.

## v0.15.0 - 2026-02-19

### Added
- `rr chats start` to resolve/create direct chats from merged contact hints (`mode=start`).
- Message reaction commands: `rr messages react` and `rr messages unreact`.
- `rr contacts list` with cursor pagination (`--cursor`, `--direction`, `--all`, `--max-items`).
- Connect discovery command: `rr connect info` (reads `/v1/info` metadata/endpoints).

### Changed
- Upgraded SDK dependency to `github.com/beeper/desktop-api-go v0.3.0`.
- `auth status --check` and `doctor` now use OAuth introspection (`/oauth/introspect`) when available, with fallback to account-list validation for older API builds.
- Message list/search JSON/plain output now includes `message_type` and `linked_message_id`; reactions include image URL.
- Account network handling now degrades gracefully to `"unknown"` when newer API builds omit `Account.network`.

### Fixed
- Updated completion metadata and capabilities/read-only coverage for new commands (`connect info`, `contacts list`, `messages react`, `messages unreact`, `chats start`).
- Updated skill docs (`SKILL.md`, `skill/SKILL.md`) and README examples to match current command surface.

## v0.14.4 - 2026-02-12

### Fixed
- ClawdHub skill metadata: don't gate skill availability on `BEEPER_TOKEN` or config-file presence; Roadrunner reads token from `~/.config/beeper/config.json` and `BEEPER_TOKEN` is an override.

## v0.14.3 - 2026-02-12

### Added
- `rr auth set --stdin` and `rr auth set --from-env <VAR>` for secret-safe token input (avoid shell history).
- Safety guard: `rr messages send`/`send-file`/`edit` refuse message text that looks like pasted rr JSON output unless `--allow-tool-output` is set.

### Fixed
- `rr auth set` now preserves unrelated keys in `~/.config/beeper/config.json` (Roadrunner only manages its own keys).

## v0.14.2 - 2026-02-12

### Changed
- Linked the ClawHub skill page from the README for easier discovery/installation.
- Updated ClawHub skill metadata to use the current `metadata.clawdbot.requires`
  schema for required env/config.
- Documented safe shell quoting for message text containing `$` or `!` (e.g.
  `$100/month` becoming `00/month` due to shell expansion) and recommended using
  `--stdin <<'EOF' ... EOF` for literal message bodies.

## v0.14.0 - 2026-02-12

### Added
- `--dedupe-window` / `BEEPER_DEDUPE_WINDOW` to block duplicate non-idempotent
  writes when the same `--request-id` and payload replay within a configured window.
- Machine-readable retry classes in `rr capabilities --json` via `retry_classes`
  (`safe`, `state-convergent`, `non-idempotent`).
- macOS CI coverage for agent smoke checks.

### Changed
- `version --json` and `capabilities --json` now advertise `dedupe-guard` and
  `retry-classes` features.
- Envelope metadata supports `request_id` on both success and error paths for
  retry correlation.
- Hardened Clawdhub skill scope: read-first defaults, explicit mutation gating,
  minimized declared environment requirements, and pinned Go install version.

## v0.13.0 - 2026-02-12

### Added
- `--chat` exact-name targeting for action commands:
  `messages send`, `messages send-file`, `messages edit`,
  `reminders set`, and `reminders clear`.
- Normalized envelope pagination metadata (`metadata.pagination`) for cursor-based
  commands (`chats list/search`, `messages list/search`, `search`, `unread`) to
  provide a stable machine contract in `--json --envelope` mode.
- Optional `error.hint` in envelope errors for actionable next steps on common
  validation/safety/connectivity failures.
- `make test-agent-smoke` and `scripts/agent-smoke.sh` for end-to-end agent-mode
  contract checks (envelope success/error shape, safety restrictions, connectivity hints).
- `--request-id` / `BEEPER_REQUEST_ID` to attach attempt IDs to envelope metadata
  (`metadata.request_id`) for retry correlation.

### Changed
- Shared exact chat resolution logic across `chats resolve` and action commands to
  keep ambiguity handling consistent for agent flows.
- `version --json` and `capabilities --json` now advertise `error-hints` in
  the feature list.
- Documented idempotency/retry guidance for agent retries across write commands.
- Added retry helper examples (bash/Node.js) using envelope `error.code` and stable request IDs.

## v0.12.0 - 2026-02-11

### Added
- `rr search --messages-all --messages-max-items` to auto-page global search
  message results with a safety cap.
- `rr messages send-file` now supports `--reply-to` and attachment override flags
  (`--attachment-file-name`, `--attachment-mime-type`, `--attachment-type`,
  `--attachment-duration`, `--attachment-width`, `--attachment-height`).
- `rr assets serve` to stream raw asset bytes from `mxc://`, `localmxc://`,
  or `file://` URLs (stdout by default, `--dest` for file output).

### Changed
- Refreshed skill documentation (`SKILL.md` and `skill/SKILL.md`) for v0.10.0/v0.11.0 features:
  message edit/send-file, asset upload/upload-base64, attachment send overrides,
  and safe auto-pagination flags.
- Expanded skill metadata with Beeper config path and required environment variables.
- Updated API and troubleshooting docs for upload/edit route behavior and attachment validation rules.

## v0.11.0 - 2026-02-11

### Added
- Attachment metadata override flags on `rr messages send`:
  `--attachment-file-name`, `--attachment-mime-type`, `--attachment-type`,
  `--attachment-duration`, `--attachment-width`, `--attachment-height`.
- `rr chats get --max-participant-count` for bounded participant payloads.
- `--all` and `--max-items` for `chats list/search` and `messages list/search`
  to auto-fetch multiple pages with a safety cap.

### Changed
- `messages send` now supports richer attachment payload metadata beyond upload ID.
- Shell completion includes new attachment override and pagination flags.

### Fixed
- Added command-level integration tests for auto-pagination loops in
  `chats list/search` and `messages list/search`.

## v0.10.0 - 2026-02-11

### Added
- `rr messages edit` to update the text of existing messages.
- `rr assets upload` and `rr assets upload-base64` to create upload IDs for attachment workflows.
- `rr messages send --attachment-upload-id <upload-id>` to send attachments using uploaded assets.
- `rr messages send-file` to upload and send an attachment in one command.
- Shell completion updates for `assets`, `contacts`, `status`, `unread`, and new message/asset flags.
- Endpoint compatibility hints for newly added commands when Desktop API routes are unavailable.

### Changed
- Upgraded SDK dependency to `github.com/beeper/desktop-api-go v0.2.0`.
- Chat/search `network` fields are populated best-effort via account lookup.
- Account-to-network lookups are cached per client instance after first successful fetch.

### Fixed
- Added regression tests for attachment send payloads and upload-then-send flow.

## v0.9.0 - 2026-01-19

### Added
- `--agent` flag for hardened agent profile (forces JSON, envelope, no-input, readonly; requires `--enable-commands`).
- `rr capabilities` command for full CLI capability discovery (JSON and human output).
- `--account` flag and `BEEPER_ACCOUNT` env var for default account ID.
- Account aliases: `rr accounts alias set/list/unset` to create shortcuts for account IDs.
- `--plain --fields` support for `status`, `search`, `doctor`, `auth status`, and `version` commands.

### Changed
- `contacts search` and `contacts resolve` now support both positional (`<account> <query>`) and flag (`<query> --account-id=<account>`) syntax.
- Account aliases are resolved transparently in all commands accepting account IDs.
- `accounts alias set/unset` blocked by `--readonly` mode.
- Command lists in capabilities are sorted alphabetically.
- Capabilities defaults now reflect current flag/env values.

### Fixed
- `auth status --plain` now outputs correctly when unauthenticated.

## v0.8.0 - 2026-01-18

### Added
- `--enable-commands` flag to allowlist specific commands (agent safety).
- `--readonly` flag to block data write operations (agent safety).
- `--envelope` flag to wrap JSON output in `{success, data, error, metadata}` structure.
- `rr version --json` now includes `features` array for capability discovery.
- Envelope error output for `--enable-commands` and `--readonly` violations when `--json --envelope` is set.

### Changed
- All JSON outputs now route through envelope-aware helper for consistent wrapping.

## v0.7.3 - 2026-01-18

### Fixed
- Add display name to ClawdHub skill publish.

## v0.7.0 - 2026-01-18

### Added
- `rr unread` to list unread chats across accounts.
- `rr status --by-account` for per-account unread summaries.
- `rr messages context` for before/after message context.
- Attachment metadata and reactions in message list/search JSON.
- `rr messages list --download-media` to copy attachments locally.

## v0.6.2 - 2026-01-17

### Changed
- Clarify strict contact resolve behavior in docs.

## v0.6.1 - 2026-01-17

### Fixed
- Deduplicate contact resolve matches to avoid false "multiple matches" errors.

## v0.6.0 - 2026-01-17

### Added
- `rr messages wait` to block until a matching message arrives.
- Filters for `rr messages tail` (`--contains`, `--sender`, `--from`, `--to`).
- `rr status` summary (unread, muted, archived).
- `rr chats resolve` and `rr contacts resolve` for exact-match IDs.

## v0.5.1 - 2026-01-17

### Added
- ClawdHub skill guidance updates (install options, safety notes).

## v0.5.0 - 2026-01-17

### Added
- Message input from file/stdin (`--text-file`, `--stdin`) and draft text file support.
- `rr messages tail` with polling and `--stop-after`.
- `--fail-if-empty` for list/search commands.
- `--fields` for `--plain` output column selection.

## v0.4.1 - 2026-01-17

### Fixed
- Normalize chat IDs that were passed with shell-escaped `\!` before API calls.

## v0.3.2 - 2026-01-16

### Added
- Align ClawdHub skill description with README tagline.

## v0.3.0 - 2026-01-16

### Added
- Contacts search (`rr contacts search`) to find users on an account.
- Chat creation (`rr chats create`) with validation for single/group participants.
- Advanced message search filters (account IDs, chat type, sender, media types, date range).
- Advanced chat search filters (include muted, last activity range).
- Documented reply-to messages and focus draft options (`--reply-to`, `--draft-text`, `--draft-attachment`).

## v0.1.0 - 2026-01-16

### Added
- Auth, doctor, accounts, chats, messages, reminders, search, and focus commands.
- JSON/Plain output modes and human-friendly UI output.
- Non-interactive safety for destructive commands.
- Shell completions, README, SKILL, and release automation.
