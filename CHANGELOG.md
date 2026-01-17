# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added
- (none)

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
