# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added
- `rr messages edit` to update the text of existing messages.
- `rr assets upload` and `rr assets upload-base64` to create upload IDs for attachment workflows.
- Endpoint compatibility hints for newly added commands when Desktop API routes are unavailable.

### Changed
- Upgraded SDK dependency to `github.com/beeper/desktop-api-go v0.2.0`.
- Chat-level `network` fields in chat/search output are now blank when unavailable from the API.

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
