# Cowork Full Mode: Runtime Research and Minimal Path

Research note on whether Newsjack's Full Mode can run inside Claude Cowork, and
the minimal path to get there. Supersedes the current launch stance that treats
Cowork as Limited Mode only (see README runtime table and
`docs/2026-06-01-open-source-installation-distribution-spec.md`).

Status: research + proposal. Two empirical checks (marked below) must pass
before we change the README runtime table or skill runtime-mode guidance.

## Background: the original "can't persist" conclusion

We had classified Cowork as Limited Mode because the detector pipeline depends
on durable local state (monitor profiles, credentials, seen-state SQLite) and a
scheduler, and we assumed an ephemeral Cowork runtime could host none of it. The
spec and README accordingly tell skills: in Cowork, do not attempt a CLI
install; fall back to Limited Mode.

That conclusion turns out to be only half right. The persistence problem is real
but narrower than we thought, and a viable Full Mode path exists.

## What the Cowork runtime actually provides

### Scheduling — two flavors, both with CLI access

- **Cowork scheduled tasks** (`/schedule` in Cowork, or the Scheduled tasks
  page) run **locally in the same Linux VM** as regular Cowork tasks, with the
  same capabilities: bash, the Claude Code CLI inside the guest, npm/pip, and
  arbitrary runnable binaries. Limitation: they only fire while the machine is
  awake and the desktop app is open; missed runs are skipped and re-run when the
  app reopens.
- **Cloud routines** (`claude.ai/code/routines`, research preview) are **full
  Claude Code cloud sessions** on Anthropic infra: shell commands, a cached
  setup script for installing tools, env vars for secrets, configurable network
  (custom allowlist or Full), MCP connectors, and schedule/API/GitHub triggers.
  These keep running with the laptop closed, but clone a fresh repo each run and
  keep no local disk between runs.

So routines *do* have CLI access, contradicting our earlier assumption.

### Persistence — the working folder is the escape hatch

Each Cowork conversation gets a **fresh per-session VM home**
(`/sessions/<random-id>/...`). Even `${CLAUDE_PLUGIN_DATA}`, documented as
persistent, does not survive across conversations today (known bug,
anthropics/claude-code#51398). So home-dir-based state is genuinely ephemeral —
that part of our original read was correct.

**But the user-selected working folder is the real host disk, mounted into the
VM via VirtioFS.** Anything written there persists across conversations
indefinitely. This is the hook that makes Full Mode possible: put Newsjack's
state and binary on the mount, not in the VM home.

### Network — configurable, not hard-locked

Cowork has an "Allow network egress" setting (Package managers only / All
domains, plus custom domain entries). Caveat: several open bugs report the
allowlist not being enforced correctly on desktop
(anthropics/claude-code#30861, #51400), so **"All domains" is the only reliable
setting today** for reaching arbitrary RSS feeds and the X API.

### A lucky break from our own packaging

`npm i -g newsjack` bundles the Go binaries inside platform packages
(`newsjack-linux-x64` etc. as `optionalDependencies`, per
`scripts/build-npm-packages.mjs`) — it does **not** download from GitHub
Releases at install time. So the npm install path needs only
`registry.npmjs.org`, which is reachable even in the most restrictive
"Package managers only" network mode. No GitHub Releases egress required.

## The minimal path to Full Mode in Cowork

1. **Dedicated persistent working folder** (e.g. `~/Newsjack`) attached to every
   Cowork conversation and to the scheduled task.
2. **State on the mount:** set `NEWSJACK_HOME` (knob already exists in the Go
   CLI) to `<working folder>/.newsjack`. Credentials, monitor profiles, and
   seen-state SQLite then live on host disk and survive sessions. No Go changes.
3. **Binary on the mount:** first run does `npm i -g newsjack` (registry-only,
   works in the VM), then cache the linux binary into
   `<working folder>/.newsjack/bin/`. Later sessions run it straight from the
   mount — no network install at all. Use `NEWSJACK_INSTALL_SKILLS=0` since
   skills come from the plugin, not the CLI installer.
4. **Network egress → "All domains"** so RSS feeds and the X API are reachable
   for source ingestion.
5. **A Cowork scheduled task** pointed at that working folder, prompting the
   newsjack-detector skill — that is the recurring monitor.

This reuses existing env knobs (`NEWSJACK_HOME`, `NEWSJACK_INSTALL_SKILLS`) and
the existing npm packaging. No Go or distribution changes are required for the
scheduled-task path.

### Residual gap and the always-on variant

The one gap vs. native Claude Code is that Cowork scheduled tasks need the app
open. For true always-on, a **cloud routine** closes it: the cached setup script
installs newsjack with Full network access, and seen-state is persisted by
pushing to a small private state repo (or the Medialyst substrate), since cloud
runs clone fresh and keep no local disk between runs.

## Empirical checks required before changing product docs

1. **RSS egress on desktop.** Confirm the macOS egress proxy bug
   (anthropics/claude-code#30861) does not block RSS fetching even on
   "All domains". If it does, source ingestion is degraded regardless of the
   rest of the path.
2. **SQLite over VirtioFS.** Confirm the seen-state SQLite store behaves over the
   VirtioFS mount — file locking can be flaky on shared filesystems. Low risk for
   the single-writer scheduled-task case, but verify before documenting.

## The Medialyst credential problem

The persistence path above moves Newsjack's *own* state (profiles, seen-state,
the binary) onto the host mount. But the live news-search source, hosted media
lists, and enrichment depend on a **Medialyst credential**, and that credential
has its own placement problem that the working-folder trick does not fully
solve for the cloud case.

### How Medialyst auth works today

- Medialyst auth is a **static per-user API key** (`mlst_…`) used as a Bearer
  token — not an OAuth connector. Each user gets their own key from
  `medialyst.ai/agents` (300 free credits, then paid), so usage is billed per
  account.
- The key is stored locally in `~/.newsjack/credentials.json` (or
  `MEDIALYST_API_KEY` env / `.env`), per `apps/cli/cmd/newsjack/auth.go`.
- The MCP server is **remote** (`https://medialyst.ai/api/mcp`), but the key is
  **injected locally** by the `newsjack` CLI, in one of two modes
  (`.mcp.json` + `apps/cli/cmd/newsjack/mcp_bridge.go`):
  1. **stdio bridge** (`newsjack mcp-bridge`) — loads the key, attaches
     `Authorization: Bearer mlst_…` to each HTTP call.
  2. **headersHelper** (`newsjack auth headers`) — the harness hits the remote
     URL directly but calls the local CLI for the auth header.
- Both modes require the `newsjack` CLI present in the runtime to inject the
  key. Per `docs/getting-started.md`: "MCP configs do not store the key; they
  point the runtime at `newsjack mcp-bridge`."

### Can Medialyst be a native claude.ai connector? Not today.

Probing the live endpoint, it *advertises* OAuth:

```
www-authenticate: Bearer error="invalid_token",
  resource_metadata="https://medialyst.ai/.well-known/oauth-protected-resource"
```

…but that advertised discovery URL **returns 404** (as do the
authorization-server metadata variants). So the OAuth challenge header is
cosmetic — the actual authorization-server flow a claude.ai remote connector
needs is not served. Medialyst is a static-token resource, full stop.

### What this means per surface

| Surface | Works? | Where the key lives |
| --- | --- | --- |
| **Cowork scheduled task** (local VM) | Yes | `newsjack` CLI is present (per the path above) -> bridge loads the key from `credentials.json` on the persistent working folder (`NEWSJACK_HOME` on the mount). No extra mechanism. |
| **Cloud routine** (Anthropic cloud) | Only via env-var secret | No local key store survives between runs and Medialyst isn't a connector, so `MEDIALYST_API_KEY` goes in the routine's environment-variable secret store; the setup script installs `newsjack` and the bridge picks it up. Manual, per-user paste. |
| **Native claude.ai connector** (cleanest) | No | Requires Medialyst to ship a working OAuth flow (serve the resource-metadata + authorization-server docs). This is a Medialyst-side change, not a Newsjack one — and not in this repo (`apps/site` here is the `newsjack.sh` install site, not the Medialyst backend). |

So the working-folder path closes the **Cowork scheduled-task** case cleanly: the
Medialyst key rides along on the same persistent mount as the rest of
`NEWSJACK_HOME`. The residual gap is the **always-on cloud-routine** case, where
the key must be hand-entered as a routine env secret until Medialyst implements
OAuth.

Highest-leverage fix: if Medialyst served the OAuth protected-resource flow it
already half-advertises, it becomes a one-click claude.ai connector and the
entire Cowork/cloud-routine key problem disappears — connector traffic routes
through Anthropic and no local key store is needed. That work lives in the
Medialyst codebase, not here.

## Repo changes this implies (once checks pass)

- Update skill runtime-mode guidance: the detector/setup skills currently
  hard-code "Cowork → Limited Mode, never install". Add a workspace-store
  bootstrap path for the persistent-working-folder case.
- Add a "workspace store" bootstrap snippet (set `NEWSJACK_HOME` on the mount,
  npm install once, cache the binary) to the detector and setup skills.
- Flip the README runtime table: Cowork moves from "Limited only" to Full Mode
  via the persistent-working-folder path, with the app-open caveat noted and the
  cloud-routine variant for always-on.

This must stay consistent with the Minimal Memory doctrine in `AGENTS.md`: the
only durable state introduced is the already-sanctioned set (monitor profiles,
credentials, schedules, seen/suppression data), now relocated onto a
user-owned, inspectable, deletable working folder rather than a hidden VM home.

## Sources

- Cowork scheduled tasks help:
  https://support.claude.com/en/articles/13854387-schedule-recurring-tasks-in-claude-cowork
- Claude Code routines docs: https://code.claude.com/docs/en/routines
- Cowork virtualization deep-dive:
  https://blog.jimmyvo.com/posts/claude-coworks-virtualization/
- Plugin-data persistence bug: anthropics/claude-code#51398
- Egress allowlist bugs: anthropics/claude-code#30861, #51400
- Scheduled tasks vs routines comparison:
  https://buildtolaunch.substack.com/p/claude-cowork-scheduled-tasks-vs-routines-vs-loop
