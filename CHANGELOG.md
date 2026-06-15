# Changelog

All notable Newsjack changes are documented here. **The release workflow
refuses to publish a tag that has no entry in this file** — add your notes
before pushing the tag.

How it works:

- Land user-visible changes with a bullet under `## Unreleased`.
- **Stable tag (`vX.Y.Z`)**: rename `Unreleased` to `## vX.Y.Z — YYYY-MM-DD`
  (and start a fresh empty `Unreleased` above it) before tagging. The
  workflow fails the release if the tag has no section.
- **Prerelease tag (`vX.Y.Z-rc.N`)**: no extra ceremony — the workflow uses
  the non-empty `Unreleased` section (or an exact tag section if you write
  one).

The published GitHub Release appends the auto-generated list of merged PRs
below these notes.

## Unreleased

### Changed

- **Medialyst now runs through the `newsjack` CLI's public-API wrapper instead
  of MCP.** The CLI calls Medialyst's public API directly; `find-journalists`
  and `news-search` use that path, with a local-artifact fallback when
  Medialyst is not configured. (#47)
- The Claude plugin manifests (`.claude-plugin/plugin.json` and
  `marketplace.json`) now track the release version; a release-workflow gate and
  a unit test keep them in sync with the tag.

### Removed

- The Medialyst MCP integration and its install-time setup, including the
  `NEWSJACK_INSTALL_MCP` option — no Node/MCP bridge is registered anymore.
  (#47)

### Docs

- README: added a skills × harness compatibility matrix, reorganized setup by
  agent platform (local agents / Claude.ai & Cowork / ChatGPT), and clarified
  ChatGPT support (Skills beta for Business and Enterprise). (#46)

## v0.1.10 — 2026-06-12

### Added

- **Windows support (GA).** One PowerShell line installs Newsjack on a stock
  Windows 11 machine with no prerequisites — no install script, no git, no
  Node. The bare `newsjack.exe` bootstraps its own release bundle (checksum
  verified), installs skills, registers MCP, adds itself to the user PATH,
  and self-updates natively from then on.
- **Go-native Medialyst MCP bridge.** `newsjack mcp-bridge` speaks
  streamable HTTP to Medialyst directly, replacing `npx mcp-remote` — the
  Node dependency is gone on every platform. API keys still load at runtime
  and never land in harness config files.
- `headline-generator` skill: headline and subject-line candidates from a
  story's raw facts.
- Windows CI: unit tests, cross-compile gates, a full bootstrap battery,
  and a Windows job in the post-release smoke.

### Fixed

- All 13 findings from the first real-machine Windows install test,
  including: bootstrap now pins the bundle to the binary's own version;
  Claude Code is detected on Windows even when not on PATH;
  `install --source` adopts prebuilt bundles into the managed install;
  `mcp setup` repairs stale Medialyst registrations instead of skipping
  them; the documented install one-liner works on stock PowerShell 5.1.
- npm release verification no longer reports false failures from registry
  propagation lag.

### Changed

- New site logo (the `N_` mark).

## v0.1.10-rc.1 — 2026-06-12

Prerelease of v0.1.10 for Windows verification; superseded by the stable
release above.

## v0.1.9 and earlier

Released before this changelog existed — see the
[release list](https://github.com/elvisun/newsjack/releases) and git
history.
