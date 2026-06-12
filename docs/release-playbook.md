# Release Playbook

This playbook covers stable, beta, and test releases for the GitHub
Release-backed Newsjack installer and npm distribution.

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

They are also published to npm as:

```text
newsjack
newsjack-linux-arm64
newsjack-linux-x64
newsjack-darwin-arm64
newsjack-darwin-x64
```

The npm path is the fallback install path for Full Mode agent harnesses where
GitHub Release assets are blocked:

```bash
npm i -g newsjack
newsjack install
```

Claude.ai, ChatGPT chat, and Claude Cowork are Limited Mode surfaces for launch;
do not treat them as CLI install targets in release smoke tests.

Beta and test releases are GitHub prereleases. They do not become the default
`latest` install target. Install them by pinning `NEWSJACK_VERSION`.

## Before Any Release

Required gates:

- `main` is green in GitHub Actions.
- The repo has been audited for secrets/private artifacts.
- The repo is public before validating unauthenticated release downloads.
- You have permission to push tags and create GitHub Releases.
- GitHub Actions can dispatch workflows with the repository `GITHUB_TOKEN`.
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
NEWSJACK_VERSION=v0.1.0-test NEWSJACK_NPM_DIST=.tmp/newsjack-npm-test \
  node scripts/build-npm-packages.mjs
node scripts/verify-npm-packages.mjs .tmp/newsjack-npm-test v0.1.0-test
```

## Required npm Setup

The npm workflow uses trusted publishing only. Do not add an `NPM_TOKEN` for
normal releases.

Trusted publishing must be configured on npm for:

```text
newsjack
newsjack-linux-arm64
newsjack-linux-x64
newsjack-darwin-arm64
newsjack-darwin-x64
```

Each package should trust:

```text
Publisher: GitHub Actions
Owner: elvisun
Repository: newsjack
Workflow filename: npm-release.yml
Allowed action: npm publish
Environment: unset
```

If npm returns `ENEEDAUTH`, check the trusted publisher fields first. The
workflow filename must be exactly `npm-release.yml`, and GitHub-hosted runners
must be used so npm can validate the OIDC token.

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

3. Watch the release workflow. It builds GitHub Release assets, creates the
   release, then queues `npm-release.yml` for the same tag:

```bash
gh run list --workflow release.yml --limit 5
gh run watch
gh run list --workflow npm-release.yml --limit 5
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

The npm workflow publishes the five npm packages after checking that the exact
version is not already present on npm, running CLI tests, building packages,
running `npm pack --dry-run`, and smoking the generated `newsjack` wrapper.

For a future tag that somehow was not published to npm, run the npm workflow
manually from that same tag:

```bash
gh workflow run npm-release.yml --ref v0.1.5 -f version=v0.1.5 -f source_ref=v0.1.5
gh run watch
```

For a historical tag that predates this npm workflow or its verification
scripts, run from `main` only if you intentionally accept that the package
provenance points at `main` instead of the original tag:

```bash
gh workflow run npm-release.yml --ref main \
  -f version=v0.1.5 \
  -f source_ref=main \
  -f allow_ref_mismatch=true
gh run watch
```

Verify npm:

```bash
npm view newsjack version
npm view newsjack-linux-arm64 version
npm i -g newsjack@0.1.5
newsjack version
newsjack skills list | grep -F newsjack-detector
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

## Windows

Windows ships as a bare `newsjack_windows_amd64.exe` release asset plus the
`newsjack_windows_amd64.tar.gz` bundle. There is no install script: users
download the exe and run `newsjack setup`, which bootstraps the bundle,
installs the binary to `%USERPROFILE%\.newsjack\bin`, and installs skills.
Self-update is native Go (`NEWSJACK_NATIVE_UPDATE` controls it on other
platforms; Windows always uses it).

Automated coverage:

- `agent-harness-ci.yml` runs Go tests on `windows-latest`, cross-compile
  gates on every PR, and the full bootstrap battery
  (`harness/scripts/run-ci-installer.ps1`): bootstrap, doctor, skills, mock
  detector, monitor lifecycle, no-Node bridge smoke, no-git leg,
  spaces-in-profile leg, and the auto-update exe swap.
- `post-release-smoke.yml` has a `windows-install` job: downloads the
  released exe, verifies its checksum, bootstraps, and runs a mock detector.
- `agent-harness-integration.yml` with `os: windows` installs real Claude
  Code via its native installer and verifies the Medialyst MCP registration;
  with the `MEDIALYST_API_KEY` secret it also runs a live bridge handshake.

Manual checklist per release (CI cannot cover these):

- Interactive `newsjack setup` prompt flow in Windows Terminal: selection
  UX, colors, and the launch handoff into Claude Code.
- SmartScreen / Mark-of-the-Web on a browser-downloaded exe: verify the
  warning path and document the click-through until code signing lands.
- One fresh Windows 11 VM run: download the exe from the release page, full
  setup with Claude Code, one real detector run.

Troubleshooting:

- If bootstrap fails, check `checksums.txt` includes the windows artifacts
  and that the release is not a prerelease.
- If auto-update leaves a `newsjack.exe.old` behind, any later `newsjack`
  command removes it once the old process has exited.
- Reinstall command on Windows is `newsjack setup` (re-runs bootstrap).
