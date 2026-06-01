# Release Playbook

This playbook covers stable, beta, and test releases for the GitHub
Release-backed Newsjack installer.

## Release Types

Newsjack uses semver-style Git tags:

```text
v0.1.0             stable release
v0.1.0-beta.1      beta prerelease
v0.1.0-test.1      test prerelease
```

Stable releases become the default install target for:

```bash
curl -fsSL newsjack.sh | bash
```

Beta and test releases are GitHub prereleases. They do not become the default
`latest` install target. Install them by pinning `NEWSJACK_VERSION`.

## Before Any Release

Required gates:

- `main` is green in GitHub Actions.
- The repo has been audited for secrets/private artifacts.
- The repo is public before validating unauthenticated release downloads.
- You have permission to push tags and create GitHub Releases.
- The local tree is clean:

```bash
git status --short
```

Recommended local checks:

```bash
cd apps/cli && go test ./...
cd ../site && pnpm lint && pnpm test && pnpm build
cd ../..
NEWSJACK_VERSION=v0.1.0-test NEWSJACK_RELEASE_DIST=.tmp/newsjack-release-test \
  node scripts/build-release-dist.mjs
```

## Local Release Smoke

Use this before pushing a tag if you changed installer, release, or packaging
code.

```bash
NEWSJACK_VERSION=v0.1.0-test NEWSJACK_RELEASE_DIST=.tmp/newsjack-release-test \
  node scripts/build-release-dist.mjs

python3 -m http.server 8765 --directory .tmp/newsjack-release-test
```

In another terminal:

```bash
export HOME=/tmp/newsjack-release-smoke
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL http://127.0.0.1:8765/install.sh | \
  NEWSJACK_RELEASE_BASE=http://127.0.0.1:8765 \
  NEWSJACK_RUNTIMES=codex \
  NEWSJACK_INSTALL_MCP=0 \
  bash

newsjack version
newsjack skills list
newsjack doctor --json | jq .
```

Marketplace bootstrap smoke:

```bash
export HOME=/tmp/newsjack-marketplace-smoke
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL http://127.0.0.1:8765/install.sh | \
  NEWSJACK_RELEASE_BASE=http://127.0.0.1:8765 \
  NEWSJACK_INSTALL_SKILLS=0 \
  NEWSJACK_INSTALL_MCP=0 \
  bash

newsjack skills status --json | jq -e '.skills_mode == "external"'
test ! -e "$HOME/.agents/skills/newsjack-detector/SKILL.md"
```

Also verify POSIX shell compatibility:

```bash
export HOME=/tmp/newsjack-sh-smoke
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL http://127.0.0.1:8765/install.sh | \
  NEWSJACK_RELEASE_BASE=http://127.0.0.1:8765 \
  NEWSJACK_INSTALL_SKILLS=0 \
  NEWSJACK_INSTALL_MCP=0 \
  sh

newsjack version
```

## Stable Release

Use stable releases for public default installs.

1. Update or confirm release notes.
2. Tag the exact commit on `main`:

```bash
git checkout main
git pull --ff-only
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

3. Watch the release workflow:

```bash
gh run list --workflow release.yml --limit 5
gh run watch
```

4. Verify release assets:

```bash
gh release view v0.1.0 --json tagName,isPrerelease,assets
```

Required assets:

```text
install.sh
manifest.json
skills-manifest.json
checksums.txt
newsjack_darwin_amd64.tar.gz
newsjack_darwin_arm64.tar.gz
newsjack_linux_amd64.tar.gz
newsjack_linux_arm64.tar.gz
```

5. Run post-release smoke:

```bash
gh workflow run post-release-smoke.yml -f version=v0.1.0
gh run watch
```

The post-release smoke workflow verifies both direct GitHub Release assets and
the public `newsjack.sh` path in Docker. Failures use GitHub Actions' default
failed-workflow notifications.

6. To run a true public install smoke locally without touching the host home,
   use Docker:

```bash
docker run --rm -i --network bridge debian:bookworm-slim bash -s <<'EOF'
set -euo pipefail
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
  ca-certificates curl gzip jq tar
rm -rf /var/lib/apt/lists/*

export HOME=/tmp/newsjack-public-smoke
curl -fsSL newsjack.sh | NEWSJACK_VERSION=v0.1.0 bash
export PATH="$HOME/.newsjack/bin:$PATH"
newsjack version
newsjack skills status --json | jq .
EOF
```

## Beta Release

Use beta releases for external testers. They are GitHub prereleases and should
not become the default `latest` release.

```bash
git checkout main
git pull --ff-only
git tag -a v0.1.0-beta.1 -m "v0.1.0-beta.1"
git push origin v0.1.0-beta.1
```

Watch and smoke:

```bash
gh run list --workflow release.yml --limit 5
gh run watch
gh workflow run post-release-smoke.yml -f version=v0.1.0-beta.1
gh run watch
```

Install pinned beta:

```bash
curl -fsSL newsjack.sh | NEWSJACK_VERSION=v0.1.0-beta.1 bash
```

Marketplace-style beta bootstrap:

```bash
curl -fsSL newsjack.sh | \
  NEWSJACK_VERSION=v0.1.0-beta.1 \
  NEWSJACK_INSTALL_SKILLS=0 \
  NEWSJACK_INSTALL_MCP=0 \
  bash
```

## Test Release

Use test releases to validate release plumbing before a beta or stable cut.
Prefer `vX.Y.Z-test.N`.

```bash
git checkout main
git pull --ff-only
git tag -a v0.1.0-test.1 -m "v0.1.0-test.1"
git push origin v0.1.0-test.1
```

Install pinned test release directly from GitHub assets:

```bash
base=https://github.com/elvisun/newsjack/releases/download/v0.1.0-test.1

curl -fsSL "$base/install.sh" | \
  NEWSJACK_RELEASE_BASE="$base" \
  NEWSJACK_INSTALL_SKILLS=0 \
  NEWSJACK_INSTALL_MCP=0 \
  bash
```

Run the manual post-release workflow:

```bash
gh workflow run post-release-smoke.yml -f version=v0.1.0-test.1
gh run watch
```

If the test release is bad and has no users, delete it:

```bash
gh release delete v0.1.0-test.1 --cleanup-tag
```

Do not delete a stable release after users may have installed it. Publish a new
patch release instead.

## Rollback

There is no mutable rollback for stable users. Publish a new patch release:

```bash
git tag -a v0.1.1 -m "v0.1.1"
git push origin v0.1.1
```

Auto-update will move managed installs from `v0.1.0` to `v0.1.1` on the next
normal user-facing `newsjack` command.

## Troubleshooting

If `curl -fsSL newsjack.sh | bash` fails:

- Check that `newsjack.sh` serves `/install.sh` to curl and redirects browsers.
- Check that the latest stable GitHub Release exists and is not a prerelease.
- Check that all required assets exist.
- Check that `checksums.txt` contains the platform artifact.
- Run a pinned install with `NEWSJACK_VERSION=<tag>`.

If marketplace bootstrap overwrites skills:

- Confirm the command used `NEWSJACK_INSTALL_SKILLS=0`.
- Check `~/.newsjack/install.json` for `"skills_mode": "external"`.
- Run `newsjack skills status --json`.

If auto-update does not run:

- Confirm the executable is `~/.newsjack/bin/newsjack`.
- Confirm `NEWSJACK_AUTO_UPDATE` is not `0`.
- Confirm `NEWSJACK_NO_AUTO_UPDATE` is not `1`.
- Run `newsjack doctor --json` and inspect `.install`.
