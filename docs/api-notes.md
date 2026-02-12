# API Notes

- Desktop API docs: https://developers.beeper.com/desktop-api
- Base URL: http://localhost:23373 (Beeper Desktop must be running).

## Search behavior

- Message search is literal word match (not semantic).
- Global search uses `/v1/search` for chats/groups.
- Message pagination for global search uses `/v1/messages/search` when `--messages-*` flags are set.
- `rr chats search` supports `--scope=participants` to search by participant names.
- Action commands using `--chat` resolve by exact title/display-name/ID via chat search, then perform the mutation by chat ID.
- CLI derives `display_name` for single chats from participants (fullName → username → id).
- Contacts search uses the account-specific contacts endpoint; use results to create new chats.
- Asset downloads use `mxc://` or `localmxc://` URLs and return a local `file://` URL.

## Message and asset mutations

- Attachment sending uses a two-step flow:
  1) upload via `/v1/assets/upload` or `/v1/assets/upload/base64`
  2) send via `/v1/chats/{chatID}/messages/send` with `attachment.upload_id`.
- `rr messages send-file` performs upload + send in one command.
- `rr messages edit` calls `/v1/chats/{chatID}/messages/edit`.
- Attachment override fields (`file_name`, `mime_type`, `type`, `duration`, `size`) are sent only when an `upload_id` is provided.
- `rr assets serve` streams `/v1/assets/serve` bytes for `mxc://`, `localmxc://`, or `file://` URLs.

## Envelope errors

- In `--json --envelope` mode, error responses may include `error.hint` for actionable remediation.
- Hints are deterministic guidance for common failure modes (allowlist/readonly restrictions, missing chat disambiguation, missing upload ID, connectivity checks).
- When `--request-id` (or `BEEPER_REQUEST_ID`) is set, envelopes include `metadata.request_id` for attempt correlation.

## Idempotency and retries

- Read/list/search/get/resolve commands are safe to retry.
- `messages send`, `messages send-file`, `chats create`, and asset uploads are non-idempotent and may duplicate side effects on replay.
- `messages edit`, `chats archive`/`--unarchive`, reminders set/clear, and account alias set/unset are state-convergent for identical inputs.
- For machine clients:
  - retry `CONNECTION_ERROR` with backoff,
  - fix inputs/config before retrying `VALIDATION_ERROR`/`AUTH_ERROR`,
  - refresh IDs before retrying `NOT_FOUND`.
- `--dedupe-window` can block duplicate non-idempotent writes when the same `--request-id` + payload repeats within the configured window (bypass with `--force`).

## Pagination

- Chats list returns `newestCursor` and `oldestCursor`.
- Messages list uses the last item's `sortKey` as the cursor.
- Message search returns `oldestCursor`/`newestCursor` for paging.
- `--all` auto-fetches pages client-side with a safety cap (default 500, max 5000 via `--max-items`).
- `rr search --messages-all` auto-pages global search message results using `--messages-cursor` semantics with a separate cap via `--messages-max-items`.
- In `--json --envelope` mode, cursor-based commands include normalized machine metadata at
  `metadata.pagination` with fields: `has_more`, `direction`, `next_cursor`, `oldest_cursor`,
  `newest_cursor`, `auto_paged`, `capped`, and `max_items` (when auto-paging is enabled).
