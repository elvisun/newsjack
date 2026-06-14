# Cowork Full Mode: Runtime Research and Shipping Decision

Research note on whether Newsjack's Full Mode can run inside Claude Cowork.
It started as a feasibility study and ends as a decision record. Supersedes the
current launch stance that treats Cowork as Limited Mode only (see README
runtime table and `docs/2026-06-01-open-source-installation-distribution-spec.md`).

**Decision (TL;DR).** A full scheduled/persistent path inside Cowork is
*technically* possible — it is documented below as evidence — but it is a **P2:
deprioritized on bandwidth, not killed.** The risk and positioning notes below
explain why it is low-ROI to build now, not why it is wrong. In the meantime
Cowork inherits the same **on-demand** experience as claude.ai chat: every skill
runs in-session at full quality, nothing to install. Persistence and scheduling
already have good homes — **Claude Code (or any local agent)** and the
**cloud-hosted Medialyst experience** — so a Cowork-specific build is not on the
critical path. The one thing that *is* worth prioritizing now is the **Medialyst
MCP OAuth connector (P0)** — see
["P0" below](#p0-the-medialyst-mcp-oauth-connector). It is small, standards-based,
and improves every surface and every MCP client, not just Cowork.

The sections through "The Medialyst credential problem" are the evidence behind
that call. The reasoning for the call itself is in
["Why shipping the Cowork scheduled path is risky"](#why-shipping-the-cowork-scheduled-path-is-risky)
and the sections after it.

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

## Feasibility: the path that *would* enable Full Mode in Cowork

This section documents that the scheduled/persistent path is achievable, as the
evidence behind the decision above. It is **not** a build plan — see
["Why shipping the Cowork scheduled path is risky"](#why-shipping-the-cowork-scheduled-path-is-risky).

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

## Why shipping the Cowork scheduled path is risky

The feasibility path demos well, but standing it up as a *supported* product
surface carries cost and risk that the on-demand framing avoids entirely. This is
the case for leaving it at **P2**:

- **External egress bugs we don't control.** The desktop egress proxy has open
  reports of not enforcing the allowlist correctly even on "All domains"
  (anthropics/claude-code#30861, #51400). Direct RSS ingestion rides on that
  proxy, so a core data source depends on someone else's bug tracker. We can't
  commit to reliability we don't own.
- **App-must-be-open scheduling.** Cowork scheduled tasks only fire while the
  machine is awake and the desktop app is open. A monitor that silently skips
  runs is worse than no monitor — it erodes trust in freshness.
- **Working-folder discipline.** The whole path hinges on the user attaching the
  same persistent folder to every conversation *and* the scheduled task. Miss it
  once and that session has no state. That is a support-ticket generator.
- **VirtioFS SQLite is unverified.** Seen-state is SQLite; file locking over a
  shared VirtioFS mount can be flaky. Low risk for the single-writer case, but
  unproven, and corruption of seen-state shows up as missed or duplicate alerts.
- **Per-session binary re-bootstrap.** The VM home is ephemeral, so the path
  leans on caching the binary to the mount and re-resolving it each session.
  More moving parts to keep working as Cowork's VM internals change underneath us.
- **X token still needs manual placement.** X is a direct API call, not a
  connector, so even with the Medialyst OAuth win the user still hand-places a
  token on the mount/env for the X trend source.
- **Positioning cost (the big one).** A free, local, scheduled monitor inside
  Cowork competes directly with the cloud-hosted Medialyst experience — the tier
  we actually want engaged users to graduate into — and it does so on Anthropic's
  least-reliable surface. Building it spends engineering effort to blur a clean
  product story.

None of these make the path *wrong* — they make it **low-ROI right now**, which
is why it sits at P2 rather than in the backlog as a committed build.

## Where each need is served

The clean three-surface story makes Cowork's role obvious without a special build:

| If you want… | Use… | Why |
| --- | --- | --- |
| Basic, on-demand, no setup | claude.ai chat **or** Cowork | Both are container-backed and run every skill in-session at full quality through the Medialyst OAuth connector. |
| Scheduled runs + persistence, self-hosted | Claude Code or your own agent | Real local state (`~/.newsjack`) plus routines — the canonical Full Mode. |
| Full feature, zero setup, always-on | The cloud-hosted Medialyst experience | Persistence and scheduling live server-side; nothing to install, reachable from any surface. |

Cowork does not get its own tier — it **inherits the on-demand tier**. The middle
and bottom rows are already covered by Claude Code and Medialyst, which is exactly
why the Cowork-specific scheduled path is P2: it would duplicate, on a riskier
surface, capability two other surfaces already deliver.

## P0: the Medialyst MCP OAuth connector

The highest-leverage, lowest-cost finding in this whole investigation has nothing
to do with Cowork's plumbing: **finish the standard MCP OAuth flow on the
Medialyst MCP server.** It already half-advertises it — the `WWW-Authenticate`
header points at a `resource_metadata` URL that currently 404s — so the remaining
work is serving the protected-resource and authorization-server metadata so any
client can complete the handshake.

Why this is P0:

- **It's the standards-compliant thing to do.** The MCP authorization spec
  (RFC 9728 / OAuth 2.0 protected resources) is how remote MCP servers are meant
  to authenticate. Today Medialyst is a static-Bearer endpoint with a cosmetic
  OAuth header; finishing the flow makes it a well-behaved MCP citizen.
- **It helps the whole ecosystem, not just Newsjack.** A standards-based
  connector works in *any* MCP client — Claude chat, Cowork, Claude Code, and
  third-party agents — with no proprietary `newsjack mcp-bridge` shim in the loop.
  The open-source skills stay decoupled from a private key-injection helper.
- **It's a security win for users.** No copy-pasting long-lived `mlst_` API keys
  into `.env` files or routine secret stores. OAuth tokens are scoped, revocable,
  and never sit in plaintext config.
- **It removes setup friction everywhere.** A one-click "connect Medialyst" in
  the user's account replaces "get a key, run `newsjack login`, configure MCP."
  This is the difference between on-demand chat/Cowork being a real experience
  and a degraded one — it is what turns the news-search, story-origin,
  fact-check, journalist-fit, and media-list skills from best-effort fallback
  into the full thing.
- **It is the only path to no-setup persistence for chat.** Connector traffic
  routes through the user's account, so a future cloud-hosted detector becomes
  reachable from a plain browser tab with nothing installed.

This work lives in the Medialyst codebase (a commercial cloud substrate), not in
this open-source repo — `apps/site` here is the `newsjack.sh` install site, not
the Medialyst backend. But it is the single change that most improves the
open-source skills' out-of-box experience, so it belongs at the top of the
cross-repo priority list. Medialyst stays **optional cloud substrate, not a
signup wall** (consistent with `docs/getting-started.md`); the connector just
makes the optional upgrade frictionless and standards-based.

## Repo changes this implies

**Now (low cost, supports the on-demand tier):**

- Update skill runtime-mode guidance: the detector/setup skills currently
  hard-code "Cowork → Limited Mode, never install." Reframe Cowork (and chat) as
  the **on-demand tier** — every skill runs in-session; the news/list skills
  light up once Medialyst is connected via OAuth. Point users who want scheduled,
  persistent monitoring to Claude Code or the cloud-hosted Medialyst experience.
- Update the README runtime table: Cowork is **on-demand (same as chat)**, not
  "Limited only" and not a Full Mode scheduled surface.

**Deferred (P2 — only if bandwidth and the positioning case change):**

- The workspace-store bootstrap (set `NEWSJACK_HOME` on the mount, npm install
  once, cache the binary) for a Cowork-local scheduled monitor, plus the two
  empirical checks (RSS egress on desktop, SQLite over VirtioFS). Documented in
  "Feasibility" above; not on the critical path.

Any future build must stay consistent with the Minimal Memory doctrine in
`AGENTS.md`: the only durable state introduced is the already-sanctioned set
(monitor profiles, credentials, schedules, seen/suppression data), relocated onto
a user-owned, inspectable, deletable working folder rather than a hidden VM home.

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
