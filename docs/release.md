# Release

This project uses GoReleaser (see `.goreleaser.yml`) and GitHub Actions
(`.github/workflows/release.yml`). Tag pushes create GitHub releases and
update the Homebrew tap automatically.

## Versioning Policy

- Keep the CLI and Clawdhub skill in lockstep (`vX.Y.Z` tag for the repo, skill
  published as `X.Y.Z`).
- Cut a patch release for skill-only changes (SKILL metadata, install guidance,
  safety notes). Avoid "skill version drift" where the skill version is ahead of
  the latest CLI tag.
- If `SKILL.md` changes, keep the Go install pin (`module: ...@vX.Y.Z`) aligned
  with the release tag.

## Prereqs

- Clean working tree on `main`.
- CI is green for the commit you are tagging.
- `TAP_GITHUB_TOKEN` secret set on the repo (for Homebrew tap updates).

## Release Steps

1) Update `CHANGELOG.md` (add the new version section).

2) Tag and push:

```bash
git checkout main
git pull

git commit -am "release: vX.Y.Z"
git tag vX.Y.Z
git push origin main --tags
```

3) Verify GitHub release artifacts:

```bash
gh run list -L 5 --workflow release.yml

gh release view vX.Y.Z
```

4) Verify Homebrew tap updated:

```bash
gh api repos/johntheyoung/homebrew-tap/contents/Formula/roadrunner.rb?ref=main --jq '.sha'
```

5) Publish Clawdhub skill:

```bash
make skill-publish VERSION=vX.Y.Z
```

The Makefile strips the leading `v` for Clawdhub versions.

Only publish if `SKILL.md` changed since the last release.

6) Sanity check a release artifact:

```bash
curl -sL https://github.com/johntheyoung/roadrunner/releases/download/vX.Y.Z/roadrunner_X.Y.Z_linux_amd64.tar.gz | tar xz
./rr version
```

7) Verify Clawdhub shows the new version (if the CLI supports it):

```bash
clawdhub info roadrunner
```

## Homebrew Install Test (optional)

```bash
brew untap johntheyoung/tap || true
brew tap johntheyoung/tap
brew install johntheyoung/tap/roadrunner
rr version
```

## Rerun a Release

If the workflow needs a rerun:

```bash
gh workflow run release.yml -f tag=vX.Y.Z
```
