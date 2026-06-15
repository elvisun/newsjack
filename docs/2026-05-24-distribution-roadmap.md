# Distribution Roadmap

This roadmap captures the pre-open-source distribution plan. For the public
open-source release model, including `newsjack.sh`, GitHub Releases, skills.sh,
and Claude marketplace distribution, see
[`2026-06-01-open-source-installation-distribution-spec.md`](./2026-06-01-open-source-installation-distribution-spec.md).

Newsjack v1 ships with the curl installer as the primary distribution path:

```bash
curl -fsSL newsjack.sh | sh
```

The installer tracks the latest production deployment from `main`, installs the local skill layer, and configures supported agent runtimes. This keeps the beta loop fast: push to this repo, deploy the site, and new installs pick up the Vercel-bundled installer, binaries, and skills without a manual package release.

## Current Channel

### `curl -fsSL newsjack.sh | sh`

Status: v1 channel.

Owns:

- installing a managed checkout at `~/.newsjack/newsjack`
- installing the prebuilt Go `newsjack` binary at `~/.newsjack/bin/newsjack`
- detecting Codex, Claude Code, OpenClaw, Hermes, or combinations of them
- generating Limited Mode skills into runtime-specific skill directories
- preparing Medialyst REST API commands once the user logs in
- updating from the latest Vercel production deployment
- auto-updating installed binaries before normal user-facing commands

Use this channel for:

- beta users
- fast iteration
- latest skill updates
- dogfooding runtime detection and REST-backed skill workflows

Do not promise immutable installs on this channel. It follows the latest main deployment by design.

Auto-update is default-on for installed binaries. The CLI checks the hosted channel commit against `~/.newsjack/newsjack/VERSION`; when it differs, the CLI runs the hosted installer and then re-runs the original command on the new binary. It skips installer/update internals and auth commands, and `NEWSJACK_AUTO_UPDATE=0` disables the behavior.

The public curl path must never require Go on the user's machine. It installs a compiled binary from the hosted artifact. Local source installs may still pass `NEWSJACK_SOURCE_DIR` and `NEWSJACK_CLI_BINARY` while iterating on the installer.

## Later Channel

### `npm install -g newsjack`

Status: planned, not v1.

Purpose:

- stable versioned CLI installs
- normal `npm update -g newsjack` and `npm uninstall -g newsjack` flows
- npm package identity, registry metadata, and provenance
- friendlier security review for teams that dislike `curl | sh`

Open decision:

- whether npm ships the full CLI and bundled skills, or acts as a bootstrapper for the hosted installer.

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
newsjack media-lists create ...
newsjack news search ...
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
- Medialyst auth/API failures degrade to local mode and do not block install
- README install instructions match the actual hosted behavior

Before npm:

- keep the real `apps/cli` package as the npm entrypoint or ship platform binary artifacts
- decide package name: `newsjack` vs scoped package
- add `npm pack --dry-run` to CI
- add a tag/release-based publish workflow
- prefer npm trusted publishing from GitHub Actions over long-lived npm tokens
