# API Notes

- Desktop API docs: https://developers.beeper.com/desktop-api
- Base URL: http://localhost:23373 (Beeper Desktop must be running).

## Search behavior

- Message search is literal word match (not semantic).
- Global search uses `/v1/search` for chats/groups.
- Message pagination for global search uses `/v1/messages/search` when `--messages-*` flags are set.
- `rr chats search` supports `--scope=participants` to search by participant names.
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

## Pagination

- Chats list returns `newestCursor` and `oldestCursor`.
- Messages list uses the last item's `sortKey` as the cursor.
- Message search returns `oldestCursor`/`newestCursor` for paging.
- `--all` auto-fetches pages client-side with a safety cap (default 500, max 5000 via `--max-items`).
- `rr search --messages-all` auto-pages global search message results using `--messages-cursor` semantics with a separate cap via `--messages-max-items`.
