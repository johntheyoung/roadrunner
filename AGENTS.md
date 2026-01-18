# Repository Guidelines

## Project Structure & Module Organization
- CLI entrypoint: `cmd/rr/main.go`.
- Command implementations: `internal/cmd/`.
- Beeper Desktop API client: `internal/beeperapi/`.
- Config + paths: `internal/config/`.
- Output formatting: `internal/outfmt/`.
- UI helpers: `internal/ui/`.
- Error formatting: `internal/errfmt/`.
- Docs: `README.md`, `docs/`, `SKILL.md`, `CHANGELOG.md`.
- ClawdHub skill source: `SKILL.md` (synced to `skill/SKILL.md` via `make skill-sync`).

## Build, Test, and Development Commands
- Format: `make fmt` (or `make fmt-check` to verify formatting).
- Test: `make test`.
- Vet: `make vet`.
- Lint: `make lint` (golangci-lint).
- GoReleaser builds: `.goreleaser.yml` (tag-driven, CI).
- Publish skill: `make skill-publish VERSION=vX.Y.Z` (strips leading `v`).

## Coding Style & Naming Conventions
- Go formatting via `gofmt` (`make fmt`).
- Output hygiene: send machine-readable data to stdout; send hints/errors to stderr.

## Testing Guidelines
- Unit tests live alongside code (`*_test.go`).
- Default test command: `go test ./...` via `make test`.
- Lint config: `.golangci.yml` enables `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`.

## Commit & Pull Request Guidelines
- Commit messages: Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`) with action-oriented subjects.
- Keep changes scoped; avoid bundling unrelated refactors.
- PRs should summarize scope, note testing performed, and call out user-facing changes.

## Release Notes
- Release flow documented in `docs/release.md`.
- Update `CHANGELOG.md` before tagging releases.

## Security & Configuration Tips
- Requires Beeper Desktop running locally; token stored in `~/.config/beeper/config.json`.
- `BEEPER_TOKEN` overrides config; `BEEPER_URL` and `BEEPER_TIMEOUT` control API access.
- Destructive commands require confirmation unless `--force` or `--no-input` (see README).

## Links
- Desktop API: https://developers.beeper.com/desktop-api
- Troubleshooting: `docs/troubleshooting.md`
- API notes: `docs/api-notes.md`
