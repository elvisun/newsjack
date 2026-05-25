# Distribution Roadmap

Newsjack v1 ships with the curl installer as the primary distribution path:

```bash
curl newsjack.sh | sh
```

The installer tracks GitHub `main`, installs the local skill layer, and configures supported agent runtimes. This keeps the beta loop fast: push to this repo, deploy the site, and new installs pick up the current installer and skills without a manual package release.

## Current Channel

### `curl newsjack.sh | sh`

Status: v1 channel.

Owns:

- installing a managed checkout at `~/.newsjack/newsjack`
- building or installing the Go `newsjack` binary at `~/.newsjack/bin/newsjack`
- detecting Codex, Claude Code, OpenClaw, Hermes, or combinations of them
- generating instruction-only skills into runtime-specific skill directories
- configuring optional Medialyst MCP where a noninteractive setup path exists
- updating from GitHub `main`

Use this channel for:

- beta users
- fast iteration
- latest skill updates
- dogfooding runtime detection and MCP setup

Do not promise immutable installs on this channel. It follows `main` by design.

## Later Channel

### `npm install -g newsjack`

Status: planned, not v1.

Purpose:

- stable versioned CLI installs
- normal `npm update -g newsjack` and `npm uninstall -g newsjack` flows
- npm package identity, registry metadata, and provenance
- friendlier security review for teams that dislike `curl | sh`

Open decision:

- whether npm ships the full CLI and bundled skills, or acts as a bootstrapper for the GitHub-backed installer.

Default recommendation:

- keep `curl` as the latest-main channel
- make npm the stable release channel
- do not publish npm on every push to `main`
- publish npm from tags or GitHub releases after v1 behavior settles

## CLI Shape

The `newsjack` command should become the stable product interface. Skill docs should call:

```bash
newsjack login
newsjack skills
newsjack detector run ...
newsjack mcp-bridge
```

They should not rely on repo-relative implementation paths. Skill-local helper files can stay near the skills that need them, but they are implementation details. The CLI owns the public contract.

## Runtime Targets

Supported v1 targets:

- Codex
- Claude Code
- OpenClaw
- Hermes

The installer should remain additive: if multiple runtimes are present, configure all detected runtimes unless `NEWSJACK_RUNTIMES` narrows the target list.

## Release Gates

Before calling the curl channel live:

- DNS for `newsjack.sh` resolves
- `newsjack.sh` browser requests serve the marketing site
- curl/wget requests serve shell from `/install.sh`
- installer smoke test passes from a clean temporary `$HOME`
- runtime skill install paths are verified for Codex, Claude Code, OpenClaw, and Hermes
- MCP setup failures are warnings, not install blockers
- README install instructions match the actual hosted behavior

Before npm:

- keep the real `apps/cli` package as the npm entrypoint or ship platform binary artifacts
- decide package name: `newsjack` vs scoped package
- add `npm pack --dry-run` to CI
- add a tag/release-based publish workflow
- prefer npm trusted publishing from GitHub Actions over long-lived npm tokens
