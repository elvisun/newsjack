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

## v0.1.13 — 2026-06-24

### Fixed

- `press-clip` now treats the outlet logo as a required trust signal: it tries
  the article masthead, then the outlet home page, then logo metadata, and
  reports where the logo came from so reviewers can catch text-only fallbacks.

## v0.1.12 — 2026-06-24

### Added

- **`press-clip` skill:** turn a live article URL into a branded press-clip
  PDF, preserving the outlet's logo and layout while stripping ads and
  clutter. This is local-agent only because it drives a real browser.
- **Medialyst OAuth device login:** `newsjack login` now guides local agents
  through a browser-based Medialyst authorization flow and stores refreshable
  credentials for news search and journalist enrichment.

### Changed

- README and getting-started docs now explain the AI-prompt setup path and
  show `press-clip` in the skills compatibility matrix.
- `find-journalists` and `news-search` copy now points agents toward the
  refreshed Medialyst public-API path.

### Fixed

- `press-clip` now closes its browser process even when clipping fails.

## v0.1.11 — 2026-06-15

### Changed

- **Medialyst now runs through the `newsjack` CLI's public-API wrapper instead
  of MCP.** The CLI calls Medialyst's public API directly; `find-journalists`
  and `news-search` use that path, with a local-artifact fallback when
  Medialyst is not configured. (#47)
- **`find-journalists` now treats `newsjack` as a news and enrichment data
  layer, not a hosted media-list manager.** Agents own list organization and
  final fit judgment; the final outreach list stays small and relevant, while
  larger candidate enrichment is allowed when screening multiple regions,
  angles, beats, or ambiguous bylines. (#48)
- The Claude plugin manifests (`.claude-plugin/plugin.json` and
  `marketplace.json`) now track the release version; a release-workflow gate and
  a unit test keep them in sync with the tag.

### Removed

- The Medialyst MCP integration and its install-time setup, including the
  `NEWSJACK_INSTALL_MCP` option — no Node/MCP bridge is registered anymore.
  (#47)
- The `newsjack media-lists ...` command family. The CLI no longer creates,
  inspects, updates, shares, stores, or manages hosted media lists. (#48)

### Docs

- README: added a skills × harness compatibility matrix, reorganized setup by
  agent platform (local agents / Claude.ai & Cowork / ChatGPT), and clarified
  ChatGPT support (Skills beta for Business and Enterprise). (#46)
- Added Medialyst agent-native media-list and enrichment API handoff notes.
  (#48)

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
