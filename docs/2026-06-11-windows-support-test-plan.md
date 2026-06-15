# Windows Support: Harness and Test Plan

Date: 2026-06-11
Status: proposed

## Scope

This plan covers verification for native Windows support built on the
no-install-script path:

1. Windows release artifacts (`newsjack.exe`, `newsjack_windows_amd64.tar.gz`).
2. Native bundle-apply in Go: fetch release, verify checksum, unpack, swap.
3. Self-bootstrap: a bare downloaded exe plus `newsjack setup` produces a full
   install with no shell script, Node, or git.
4. Go-native Medialyst REST client commands (shared across all platforms).
5. Windows-native self-update, including the running-exe swap.
6. Per-OS setup wizard behavior (runtime install commands, doctor checks).

Goals, in priority order: (a) do not break the existing macOS/Linux install
path, (b) prove the Windows path end to end, (c) keep both protected in CI
permanently.

## Principles

- Same doctrine as `harness/README.md`: the harness observes the real compiled
  `newsjack` binary. No second implementation of install logic in test
  scripts. The no-script design makes this easier on Windows: there is no
  installer script to mirror, the binary owns everything and the harness only
  asserts.
- Windows CI does not use Docker. A fresh `windows-latest` runner VM is the
  disposable clean environment; isolation comes from pointing `NEWSJACK_HOME`
  (and `USERPROFILE` where needed) at a throwaway directory, exactly as the
  Linux harness does with `HOME=/tmp/newsjack-home`.
- Windows assertions are written in pwsh, not Git Bash. Running the battery
  under bash on Windows would mask the PowerShell-environment bugs the tests
  exist to catch.
- Every assertion that exists in the Linux battery
  (`harness/scripts/run-ci-installer.sh`) has a Windows twin unless it is
  installer-script-specific. The JSON contracts (`setup --json`,
  `doctor --json`, detector/monitor outputs) are identical across platforms
  and are the shared source of truth.

## Layer 0: cross-compile gates (land before any feature work)

In `agent-harness-ci.yml` `cli-unit`:

- `GOOS=windows GOARCH=amd64 go build ./...`
- `GOOS=darwin GOARCH=arm64 go build ./...` (guards the reverse direction)

Cost: seconds. Catches compile-level platform regressions in both directions
from day one. The Windows build passes today; this keeps it that way.

## Layer 1: unit tests on both OSes

Add a `windows-latest` leg to `cli-unit` running `go test ./...`.

Prerequisite test fixes (pure test-infra, no product change):

- `withTempEnv` and tests that fake home via `HOME` must also set
  `USERPROFILE` so `os.UserHomeDir()` resolves on Windows.
- `monitor_test.go` writes `#!/bin/sh` scripts as fake runtime installers.
  These need per-OS doubles (`.cmd` shims) or explicit platform skips with a
  tracking comment.
- Audit golden fixtures for hardcoded `/` path separators in compared output.

New unit tests required with each feature:

### Bundle-apply (with native update/bootstrap work)

Serve a local release layout (`manifest.json`, `checksums.txt`, platform
tar.gz) from `httptest`. Assert, on both OSes:

- Happy path: download, checksum verify, unpack to `~/.newsjack/newsjack`,
  layout matches what `install.sh` produces today (VERSION, COMMIT,
  `skills-manifest.json`, `bin/` binary, prebuilt marker).
- Checksum mismatch: fails, and a pre-existing install is left byte-identical
  (no partial mutation).
- Truncated/interrupted download: no mutation.
- Missing platform artifact in manifest: clean, actionable error.

### Bootstrap (with bootstrap work)

- Empty `NEWSJACK_HOME` + `NEWSJACK_RELEASE_BASE` override: full layout
  appears, `install.json` fields match the install.sh-written equivalent
  (`skills_mode`, `runtimes_raw`, `channel`), binary copied to
  `~/.newsjack/bin/newsjack` (`.exe` on Windows).
- Re-run is idempotent and preserves the `.previous` rollback dir behavior.

### Self-update swap (with update work)

- Swap function uses rename-then-move (never in-place overwrite), leaves
  `newsjack.exe.old` on Windows, and the next invocation cleans it up.
- `runningInstalledBinary()` matches the per-OS binary name.

### Medialyst REST client (runs on all platforms)

Contract tests against in-process fake Medialyst REST endpoints (`httptest`):

- `Authorization: Bearer` header present on every upstream request; key
  loaded from `credentials.json` and from `MEDIALYST_API_KEY`, with the
  credentials-file path winning per current precedence.
- Commands forward request bodies exactly when `--json` or `--json-file` is
  used.
- Upstream 401: the CLI exits nonzero with the re-auth hint, not a panic.
- API errors include Medialyst's code, message, and request ID when available.

These replace the former bridge transport tests.

## Layer 2: Linux/macOS regression protection (existing harness, surgical changes)

The existing pipeline (`release-installer-smoke`, `no-token-installer` Docker
battery, `post-release-smoke`) stays as-is and keeps gating every PR. Changes
only where the shared REST client lands:

- `doctor --json` should not list Node or npx as required for Medialyst.
- REST smoke (new step in the container battery): start a mock Medialyst
  REST endpoint, point the CLI at it via `NEWSJACK_MEDIALYST_API_BASE`, run
  one news search and one media-list command, and assert request/response
  handling. This runs no-token by injecting a fake `MEDIALYST_API_KEY`.
- Live gate for REST commands: run a small credit-spending smoke against the
  real Medialyst endpoint before merge and once after release.
- Self-update: macOS/Linux stay on the hosted `curl | sh` update path
  initially (Windows-gated native apply), so the existing auto-update
  observation flow in `harness/README.md` remains valid unchanged. If/when
  unix switches to native apply, the same Docker auto-update smoke
  (stale VERSION, run doctor, assert rewrite) covers it with no new
  infrastructure.

## Layer 3: Windows CI battery (the harness twin)

New job(s) in `agent-harness-ci.yml` on `windows-latest`, plus
`harness/scripts/run-ci-installer.ps1` (pwsh) mirroring the container battery.

Build once, test native: the Linux job already builds the full release dist;
upload `.tmp/newsjack-release` as an artifact and have the Windows job
download it. This exercises the exact cross-compiled artifact a release
would ship instead of rebuilding on Windows.

Battery steps (each mirrors a named step in `run-ci-installer.sh`):

1. **Serve dist locally** — `python -m http.server` (preinstalled on
   runners) over the downloaded artifact dir.
2. **Bootstrap smoke** — copy the bare `newsjack.exe` to a scratch dir,
   set `NEWSJACK_HOME` to a fresh temp dir and
   `NEWSJACK_RELEASE_BASE=http://127.0.0.1:<port>`, run
   `newsjack setup --json`. Assert: bundle layout under
   `$env:NEWSJACK_HOME\newsjack`, exe copied to `...\bin\newsjack.exe`,
   `install.json` contents, `setup --json` schema (same jq assertions,
   via `ConvertFrom-Json`).
3. **Doctor** — `newsjack doctor --json`: `root_ok == true`, per-OS
   dependency list (no npx requirement), PATH-presence check result.
4. **Skills** — `newsjack-detector/SKILL.md` and
   `newsjack-monitor-setup/SKILL.md` exist under
   `$env:USERPROFILE\.claude\skills` (and `.agents\skills` for codex
   selection), no `scripts/` leakage — same exclusion asserts as Linux.
5. **Mock detector** — `newsjack detector run "AI search visibility" --mock
   --limit 1`, identical JSON assertions.
6. **Monitor lifecycle** — init/test/schedule/status with the same
   harness-coffee profile and the same field-level assertions, including
   the no-system-cron guarantees (`system_cron == false`, schedule markdown
   free of crontab/launchd/systemd references).
7. **Auto-update smoke** — write a stale `VERSION`, run `newsjack doctor`,
   assert: stderr shows auto-update, the running exe was swapped via the
   rename dance, `newsjack.exe.old` exists then is cleaned by the next
   invocation, stdout stayed valid JSON. This is the only place the real
   Windows file-locking behavior can be tested.
8. **REST smoke, no Node** — for this step strip every Node directory
   from `PATH` (runners preinstall Node; customers will not have it), run
   the same mock REST search/list workflow as the Linux battery. This leg
   is the permanent contract that the Windows path needs no Node.
9. **No-git leg** — strip git from `PATH`, re-run steps 2–6 in a second
   fresh `NEWSJACK_HOME`. Proves install/setup never silently shells to
   bash/git.
10. **Spaces-in-profile leg** — repeat bootstrap with
    `NEWSJACK_HOME=C:\t e m p\jane smith\.newsjack`-style path. Runner
    usernames never contain spaces; real customers' do.

Matrix note: steps run under pwsh 7. One spot-check leg invokes the
user-facing commands from Windows PowerShell 5.1, since with no install
script the user-typed surface is just `newsjack <cmd>` and the printed PATH
instruction — small but cheap to keep honest.

## Layer 4: manual / dispatch harness (agent-in-the-loop)

Extend `agent-harness-integration.yml` with an `os: windows` input (or a
sibling dispatch workflow):

- Installs Claude Code on the runner via its native Windows installer
  (`irm https://claude.ai/install.ps1 | iex`), pinned version arg like the
  Docker image pins `CLAUDE_CODE_VERSION`.
- Runs `newsjack install --runtimes claude`, then checks that the local
  skills can call `newsjack news search` and `newsjack media-lists`.
- With `--with-local-env`-equivalent secrets: one live `newsjack news search`
  through Claude Code on Windows. Manual, credit-spending, run before the
  Windows GA and after bridge changes.

## Layer 5: post-release smoke

Add a `windows-release-smoke` job to `post-release-smoke.yml`:

- Download the released bare exe (and checksums) from the GitHub Release,
  verify the checksum in pwsh, run bootstrap against the real release base
  (no `NEWSJACK_RELEASE_BASE` override), then doctor + skills list + mock
  detector. This is the closest CI gets to the real customer flow:
  "download one file, run setup."

## What stays manual (release checklist additions)

CI cannot cover these; keep them as a short checklist in
`docs/release-playbook.md` once Windows ships:

- Interactive `newsjack setup` prompt flow in Windows Terminal (TTY
  selection UX, colors under conhost vs Windows Terminal).
- SmartScreen / Mark-of-the-Web behavior on a browser-downloaded exe
  (known friction until code signing lands; verify the warning text and
  click-through path once per release).
- One fresh Windows 11 VM run per release: download exe from the site, full
  setup with Claude Code, one real detector run. This is the
  customer-shaped test no runner reproduces.

Local dev note: there is no Docker equivalent for Windows on macOS hosts.
The local loop is either a Windows 11 ARM VM or, preferably, driving the
Layer 3/4 workflows with `gh workflow run` and reading artifacts.

## Phasing (each lands with its feature, protection-first)

| Phase | Feature | Tests that land with it |
| --- | --- | --- |
| 0 | none (now) | Layer 0 gates; Layer 1 windows `go test` leg + test-infra fixes |
| 1 | Medialyst REST client | REST contract tests; harness mock-REST smoke (Linux); doctor no-Node assertion; live Medialyst gate before merge |
| 2 | Bundle-apply + bootstrap | Bundle/bootstrap unit tests (both OSes); Layer 3 battery steps 1–6, 8–10 |
| 3 | Windows self-update | Swap unit tests; Layer 3 step 7 |
| 4 | Release | Layer 5 post-release job; Layer 4 dispatch workflow; manual checklist in release playbook |

Phase 0 is pure protection and has no dependency on any Windows feature
work. Nothing in Phases 1–4 modifies `install.sh` or the existing Linux
battery except the single doctor-npx assertion that the bridge makes stale.
