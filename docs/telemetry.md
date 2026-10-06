# Telemetry Design

Status: proposal
Date: 2026-06-11

This is a design-only proposal for privacy-preserving telemetry across
newsjack.sh, the `newsjack` CLI, the MCP bridge, and the installer. The goal is
to learn where the product is working or failing without collecting the
public-relations work itself: no queries, no articles, no profiles, no prompts,
no credentials.

## Goals

- Measure the install funnel: site visit, installer request, install completion,
  setup completion, and first successful command.
- Understand command usage at the product-shape level: which command families
  are used, whether they complete, how long they take, and where users hit
  errors.
- Understand MCP adoption: which runtimes are configured, whether the bridge
  starts, and, if we later add an explicit proxy, tool-call counts by tool
  family.
- Improve reliability: command error classes, detector source error classes,
  update failures, doctor findings, and install/setup failure stages.
- Estimate retention for users who explicitly opt into a stable community
  installation id.
- Keep all telemetry inspectable, documented, removable, and easy to disable.

## Non-goals and Privacy Commitments

Telemetry must not become a second copy of the user's news research.

- No query content by default.
- No article titles, URLs, snippets, outlet names, authors, or source result
  payloads.
- No monitor profile content: no company names, competitors, product names,
  topics, feeds, proof assets, spokesperson names, or monitor slugs.
- No prompt text, model output, draft pitches, notes, comments, or journalist
  names.
- No credentials, tokens, API keys, env var values, or credential file paths.
- No local paths, cwd, home directory, username, hostname, repo slug, branch name,
  or git remote.
- No raw IP address storage. Web request dedupe may use a day-salted hash with a
  short raw-event retention window; CLI and MCP telemetry should not use IP as an
  identifier.
- No raw CLI argv. Only command family, subcommand, and allowlisted enum flags.
- No session replay, browser autocapture, command transcript capture, or hidden
  SDK behavior.
- Unknown event names and unknown properties are rejected server-side.
- When telemetry cannot prove a value is safe, it drops the property or no-ops.

## Inspiration From gstack and gbrain

I found the public repos at `garrytan/gstack` and `garrytan/gbrain`.

### gstack

gstack does not appear to use PostHog, Plausible, OTLP, or another general
analytics SDK for its CLI telemetry. It uses a small custom telemetry path:
local JSONL first, then an opt-in Supabase edge-function sync.

What is worth copying:

- The README states telemetry is opt-in, default-off, prompts on first run, sends
  command name/duration/status/skill/runtime info, does not send prompts or
  outputs, and can be disabled with config commands
  ([README.md lines 440-452](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/README.md#L440-L452)).
- The skill prompt defines a first-run consent flow with `off`, `anonymous`, and
  `community` modes, stores a `.telemetry-prompted` marker, and avoids prompting
  repeatedly
  ([SKILL.md lines 174-198](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/SKILL.md#L174-L198)).
- The config helper defaults telemetry to `off` and only accepts
  `off|anonymous|community`
  ([bin/gstack-config lines 33-38](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-config#L33-L38),
  [lines 111-119](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-config#L111-L119)).
- The event logger exits immediately when telemetry is off, treats invalid
  config as off, and creates a stable random UUID only for community mode rather
  than deriving identity from host/user/repo
  ([bin/gstack-telemetry-log lines 71-85](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-log#L71-L85),
  [lines 129-153](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-log#L129-L153)).
- Repo slug and branch are kept local-only and the sync step strips local-only
  fields before upload
  ([bin/gstack-telemetry-log lines 155-162](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-log#L155-L162),
  [bin/gstack-telemetry-sync lines 81-90](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-sync#L81-L90)).
- Sync is background/rate-limited, strips local-only data, and advances its
  cursor only after server success
  ([bin/gstack-telemetry-sync lines 1-6](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-sync#L1-L6),
  [lines 36-45](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-sync#L36-L45),
  [lines 117-138](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-telemetry-sync#L117-L138)).
- The local analytics command reads local JSONL directly, so users can inspect
  usage without remote analytics
  ([bin/gstack-analytics lines 1-32](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/bin/gstack-analytics#L1-L32)).
- The Supabase ingest path uses an explicit schema, payload limits, event
  allowlists, and length caps
  ([supabase/migrations/001_telemetry.sql lines 4-22](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/supabase/migrations/001_telemetry.sql#L4-L22),
  [supabase/functions/telemetry-ingest/index.ts lines 24-25](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/supabase/functions/telemetry-ingest/index.ts#L24-L25),
  [lines 70-89](https://github.com/garrytan/gstack/blob/a5833c413f98b13f105beac96262e8098b628461/supabase/functions/telemetry-ingest/index.ts#L70-L89)).

I did not find a gstack command that deletes already-uploaded remote telemetry.
That is a gap Newsjack should close if it stores a community installation id.

### gbrain

gbrain also does not appear to use a hosted product analytics provider. Its
notable pattern is local, explicitly enabled eval capture and audit logging
rather than general telemetry.

What is worth copying:

- Eval capture hooks into the operation layer, so MCP/CLI/subagent search and
  query paths can be observed consistently, but the capture is fire-and-forget
  and failures are logged instead of breaking product flow
  ([docs/eval-capture.md lines 11-40](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/docs/eval-capture.md#L11-L40),
  [src/core/eval-capture.ts lines 1-14](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/eval-capture.ts#L1-L14)).
- Capture is default-off from v0.25, env/config opt-in, and config `false`
  overrides env. PII scrubbing defaults on
  ([docs/eval-capture.md lines 123-156](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/docs/eval-capture.md#L123-L156),
  [src/core/eval-capture.ts lines 232-267](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/eval-capture.ts#L232-L267)).
- The scrubber targets emails, phones, SSNs, credit cards, JWTs, bearer tokens,
  and common secret shapes before local capture
  ([src/core/eval-capture-scrub.ts lines 1-23](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/eval-capture-scrub.ts#L1-L23),
  [lines 77-103](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/eval-capture-scrub.ts#L77-L103)).
- Query/search operation call sites explicitly decide what to capture
  ([src/core/operations.ts lines 538-563](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/operations.ts#L538-L563),
  [lines 1624-1645](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/operations.ts#L1624-L1645)).
- The schema caps captured query length and has a failure table instead of
  failing silently
  ([src/schema.sql lines 1051-1099](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/schema.sql#L1051-L1099)).
- Public eval baselines are synthetic and real captures stay local
  ([docs/eval-bench.md lines 42-49](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/docs/eval-bench.md#L42-L49)).
- Doctor surfaces capture health, and there is an explicit prune path for old
  eval rows
  ([src/commands/doctor.ts lines 6365-6379](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/commands/doctor.ts#L6365-L6379),
  [src/commands/eval-prune.ts lines 66-113](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/commands/eval-prune.ts#L66-L113)).
- Local audit logs store operational metadata, hash task text, and avoid paths
  or content where not needed
  ([src/core/audit/self-upgrade-audit.ts lines 1-8](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/audit/self-upgrade-audit.ts#L1-L8),
  [src/core/skillopt/audit.ts lines 1-10](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/skillopt/audit.ts#L1-L10),
  [lines 73-75](https://github.com/garrytan/gbrain/blob/ecd6ae87722accaa399c8594bb8ed562f967130d/src/core/skillopt/audit.ts#L73-L75)).

The strongest lesson is architectural: put telemetry at stable product
boundaries, make it default-off for sensitive workflows, and make local
inspection/deletion a first-class feature.

## Current Newsjack State

Newsjack already has limited website request telemetry. It does not have CLI or
MCP product telemetry.

Existing site telemetry:

- `apps/site/lib/install-telemetry.ts` defines `site_visit` and
  `install_request`, classifies client kind from user agent, and writes rows to
  `install_events`.
- `apps/site/proxy.ts` records `install_request` for curl/wget-style installer
  traffic and `site_visit` for browser-style traffic before routing.
- `apps/site/db/migrations/0001_install_events.sql` creates the
  `install_events` table and event-type constraint.
- `apps/site/docs/telemetry.md` says site request events are recorded and the
  shell installer does not send telemetry.
- `apps/site/README.md` documents the migration and funnel test flow.
- `scripts/funnel-stats.mjs` aggregates rows by event type, client kind,
  country, and daily unique IP hashes.

Privacy issue to fix before broadening telemetry: current site rows can include
raw user agent, full query params, referrer, and metadata. That is acceptable for
an internal early funnel experiment only if retention is short and documented,
but the long-term telemetry contract should store user-agent family, referrer
domain, and an allowlist of UTM params instead.

CLI and MCP surfaces with no telemetry today:

- `apps/cli/cmd/newsjack/main.go` dispatches all command families and is the
  natural place to wrap command timing and outcome recording.
- `apps/cli/cmd/newsjack/install.go` installs skills and configures supported
  MCP runtimes.
- `apps/cli/cmd/newsjack/mcp.go` configures Codex, Claude Code, OpenClaw, and
  Hermes MCP entries.
- `apps/cli/cmd/newsjack/mcp_bridge.go` loads the saved Medialyst key and execs
  `npx -y mcp-remote`; it can record bridge start/failure, but not successful
  tool calls without a future JSON-RPC proxy.
- `apps/cli/cmd/newsjack/detector_command.go` and
  `apps/cli/cmd/newsjack/detector_run.go` own detector options, source
  collection, diagnostics, and aggregate result counts. These files have enough
  aggregate data to record useful telemetry without logging query/profile/feed
  content.
- `apps/cli/cmd/newsjack/coverage.go` owns coverage tracker setup and counts.
  It should only expose keyword and result counts, not the keywords themselves.
- `install.sh` fetches the release, installs the CLI, writes
  `~/.newsjack/install.json`, and launches setup. It should not send a completion
  callback before consent.
- `bin/newsjack` is the source-checkout shim. Telemetry should live in the Go
  CLI, not the shim.

Local product state that is not telemetry:

- The monitor store under the Newsjack data dir keeps monitor profiles, signal
  snapshots, coverage articles, and coverage decisions. That state is
  user-owned product data and must not be copied into telemetry.

## Proposed Event Taxonomy

All events use a versioned allowlist. Unknown event names and unknown property
keys are rejected by the client before enqueue and by the server before insert.

Common properties:

| Property | Notes |
| --- | --- |
| `schema_version` | Integer, starts at `1`. |
| `event_name` | Allowlisted string. |
| `occurred_at` | Client UTC timestamp. Server also stores `received_at`. |
| `newsjack_version` | CLI/site version or `dev` for source checkout. |
| `os`, `arch` | Go runtime values, coarse only. |
| `install_channel` | `homebrew`, `curl`, `source`, `binary`, `unknown`. |
| `telemetry_mode` | `anonymous` or `community`. Never send events in `off`. |
| `session_id` | Per-process random UUID, not persisted. |
| `installation_id` | Stable random UUID only in `community` mode. |
| `duration_ms` | Wall-clock duration when applicable. |
| `outcome` | `success`, `failure`, `cancelled`, `skipped`. |
| `exit_code` | Integer for CLI command completion. |
| `error_class` | Allowlisted class, not raw error text. |

Web events:

| Event | Properties |
| --- | --- |
| `site_visit` | `client_kind`, `path_family`, `country`, `region`, `referrer_domain`, allowlisted UTM params, deployment id. |
| `install_request` | Same as `site_visit`, plus `installer_client` and `accept_language_family`. |

Recommended changes to current web telemetry:

- Store user-agent family or installer client, not raw user agent.
- Store referrer domain, not full referrer URL.
- Store only `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`,
  `utm_content`, and known distribution params.
- Keep day-salted IP hash only for same-day dedupe, not identity.

Installer and setup events:

| Event | Properties |
| --- | --- |
| `install_completed` | `platform`, `arch`, `install_channel`, `release_version`, `skills_mode`, `mcp_requested`, `runtime_count`, `setup_launched`. |
| `install_failed` | `stage`, `platform`, `arch`, `install_channel`, `error_class`. |
| `setup_started` | `interactive`, `runtime_count`, `medialyst_key_source_kind`. |
| `setup_completed` | `runtime_count`, `mcp_configured_count`, `monitor_created`, `schedule_created`, `outcome`, `error_class`. |

Do not emit installer completion until the user has opted in. The web
`install_request` event is enough for raw top-of-funnel volume; completion rate
will be measured among opted-in users.

CLI events:

| Event | Properties |
| --- | --- |
| `cli_command_completed` | `command`, `subcommand`, `duration_ms`, `outcome`, `exit_code`, `error_class`. |
| `auto_update_checked` | `current_version_known`, `target_version_known`, `update_available`, `skipped_reason`. |
| `auto_update_completed` | `from_version_known`, `to_version_known`, `outcome`, `error_class`. |
| `doctor_completed` | Counts of `ok`, `warning`, `error`, plus coarse finding classes. |
| `auth_configured` | `provider`, `key_source_kind`, `outcome`. No key value or path. |
| `skills_install_completed` | `runtime`, `skill_count`, `outcome`, `error_class`. |

Detector and monitor events:

| Event | Properties |
| --- | --- |
| `detector_run_completed` | `depth`, `mock`, `source_count`, `sources_used`, `feed_count_bucket`, `query_count_bucket`, `lookback_bucket`, `max_age_bucket`, `new_only`, `saved`, result-count buckets, lane counts, source status counts, `duration_ms`, `outcome`, `error_class`. |
| `monitor_init_completed` | `runtime`, `schedule_requested`, `profile_source_kind`, `outcome`, `error_class`. |
| `monitor_schedule_created` | `runtime`, `cadence`, `jitter_minute_bucket`, `outcome`. |
| `monitor_run_completed` | `source_count`, result-count buckets, `saved_artifact_count`, `outcome`, `error_class`. |

Detector telemetry must not include topic strings, generated queries, feed URLs,
profile JSON, monitor names, signal titles, outlet names, article URLs, or skill
judgments. Counts and enum source names are enough.

Coverage tracker events:

| Event | Properties |
| --- | --- |
| `coverage_tracker_initialized` | `keyword_count_bucket`, `outcome`, `error_class`. |
| `coverage_check_completed` | `keyword_count_bucket`, `candidate_count_bucket`, `new_real_coverage_count_bucket`, `alert_count_bucket`, `duration_ms`, `outcome`, `error_class`. |
| `coverage_record_completed` | `verdict`, `alert`, `outcome`, `error_class`. |

Do not send keyword text, article URLs, article titles, outlet names, rationale,
or tracker slug.

MCP events:

| Event | Properties |
| --- | --- |
| `mcp_configured` | `runtime`, `outcome`, `error_class`. |
| `mcp_bridge_started` | `bridge_kind`, `npx_available`, `key_source_kind`. |
| `mcp_bridge_failed` | `bridge_kind`, `error_class`, `npx_available`, `key_source_kind`. |
| `mcp_tool_call_completed` | Future only if Newsjack owns a JSON-RPC proxy: `tool_family`, `duration_ms`, `outcome`, `error_class`, `result_count_bucket`. |

Do not log MCP JSON-RPC arguments or results. In v1, do not claim tool-call
visibility from the current `mcp-remote` exec bridge.

## Consent Model

Recommendation: CLI, MCP, and install telemetry should be opt-in, not opt-out.

Reasoning:

- Newsjack users are searching live news, companies, competitors, journalists,
  and client-sensitive topics. Even "query metadata" can reveal a client brief.
- The product is open source. Trust is more valuable than a larger dataset.
- The existing site funnel already gives anonymous top-of-funnel signal.
- gstack and gbrain both lean default-off for sensitive developer workflows.

Modes:

| Mode | Meaning |
| --- | --- |
| `off` | No local event file and no network sends. Invalid config falls back to this. |
| `anonymous` | Local JSONL and remote aggregate events with no stable installation id. Good for command/error-rate metrics. |
| `community` | Same as anonymous plus a stable random installation id for retention cohorts and deletion requests. |

First-run UX:

- Prompt once on the first interactive `newsjack setup` or first interactive
  non-help command.
- Do not prompt for `help`, `version`, shell completion, or machine-facing
  `mcp-bridge`.
- Non-interactive commands default to `off` unless env or config opts in.
- Store a prompt marker so users are not asked repeatedly.
- Show the exact commands for status, inspect, disable, and delete.

Environment:

- `NEWSJACK_TELEMETRY=0|off|false|no` disables telemetry.
- `NEWSJACK_TELEMETRY=anonymous` enables anonymous telemetry.
- `NEWSJACK_TELEMETRY=1|on|community` enables community telemetry.
- Honor `DO_NOT_TRACK=1` as `off` unless the user explicitly sets a Newsjack
  config value.

Config:

```json
{
  "telemetry": {
    "mode": "off",
    "prompted_at": "2026-06-11T00:00:00Z"
  }
}
```

Suggested files:

- `~/.newsjack/config.json` for the mode and prompt marker.
- `~/.newsjack/telemetry/events.jsonl` for local queued events.
- `~/.newsjack/telemetry/cursor.json` for upload cursor state.
- `~/.newsjack/telemetry/installation-id` for community mode only.
- `~/.newsjack/telemetry/delete-token` for authenticated remote deletion in
  community mode.

Planned commands:

- `newsjack telemetry status`
- `newsjack telemetry on`
- `newsjack telemetry anonymous`
- `newsjack telemetry off`
- `newsjack telemetry inspect`
- `newsjack telemetry flush`
- `newsjack telemetry delete --local`
- `newsjack telemetry delete --remote`

## Backend Recommendation

Use a first-party collector in `apps/site` backed by the existing Postgres/Neon
database. Do not ship PostHog, Plausible, or OpenTelemetry SDKs in the CLI for
v1.

Why first-party:

- The current website already has a small telemetry substrate.
- The schema can be public, reviewed in Git, and constrained by migrations.
- The CLI can send plain JSON to one endpoint with no hidden SDK behavior.
- Server validation can enforce event names, property names, type caps, payload
  size caps, and rate limits.
- It keeps user trust higher than forwarding OSS CLI usage into a third-party
  analytics product on day one.

Tradeoffs:

| Option | Fit |
| --- | --- |
| First-party collector on `newsjack.sh` | Best v1. More schema work, less dashboard convenience, strongest privacy posture. |
| PostHog cloud | Strong funnels and retention, but too easy to overcollect. Only consider later for sanitized aggregate mirrors. Disable autocapture/session replay. |
| PostHog self-host | Better control than cloud, but operationally heavier than the current need. |
| Plausible | Good for website page analytics, weak for CLI/MCP command taxonomy and deletion workflows. |
| OTLP to our own collector | Good for traces and service observability, too infra-shaped for product usage and privacy allowlists. |

Retention:

- Raw telemetry events: 90 days.
- Daily aggregate tables/views: 13 months.
- Local JSONL queue: size capped and prunable by command.
- Anonymous remote events cannot be individually deleted after upload because
  they have no stable id; they expire by retention policy.
- Community-mode events can be deleted by installation id plus delete token.

## Implementation Sketch

CLI:

- Add `apps/cli/cmd/newsjack/telemetry.go` with a small `Recorder` interface:
  `Record(event Event)`, `Flush(ctx)`, and `Close()`.
- `NewRecorder()` returns a no-op recorder unless consent resolves to
  `anonymous` or `community`.
- Do not import a third-party telemetry SDK into the CLI. The disabled path
  should instantiate only the no-op recorder and should not initialize network,
  queue, or background-flush state.
- Parse env/config before command dispatch. Invalid values resolve to `off`.
- Wrap `runCLIWithIO` in `main.go` to time top-level command execution and
  record `cli_command_completed` after the command returns.
- Add command-specific event builders near existing command code when aggregate
  data is already available. Do not thread raw options into telemetry.
- Classify errors into allowlisted classes at the boundary:
  `network`, `auth_missing`, `auth_invalid`, `runtime_missing`,
  `config_invalid`, `rate_limited`, `source_unavailable`, `permission`,
  `timeout`, `unknown`.
- Never include raw error strings in events.

Local queue:

- Append JSONL to `~/.newsjack/telemetry/events.jsonl`.
- Hold a lock while appending/flushing.
- Keep a cursor with byte offset or event id.
- Cap file size and age. If the cap is hit, drop oldest unsent events and record
  a local-only diagnostic.
- Flush opportunistically at process end with a short timeout, around
  500-750 ms.
- Flush in the background for long-running commands, never blocking command
  success.
- Offline behavior is normal: keep local events until the cap, then drop.

Server:

- Add `apps/site/app/api/telemetry/route.ts` or equivalent endpoint.
- Accept batches, not single-event chatty calls.
- Require `schema_version`.
- Reject unknown event names, unknown properties, too-large strings, too-large
  batches, and nested arbitrary JSON.
- Store `received_at`, request country/region if available, and coarse transport
  diagnostics, but not raw IP.
- For community mode, store the stable random `installation_id` and a hash of a
  local delete token. For anonymous mode, store no stable id.
- Build aggregate views for funnel, command usage, error rates, and retention.

Installer:

- Keep `install.sh` telemetry-free before consent.
- Continue relying on website `install_request` for raw installer demand.
- After the CLI is installed and consent exists, emit `install_completed` or
  `install_failed` from the Go CLI path using `~/.newsjack/install.json`.

MCP:

- In v1, record only `mcp_configured`, `mcp_bridge_started`, and
  `mcp_bridge_failed`.
- To measure tool-call patterns later, replace the direct `mcp-remote` exec path
  with an explicit Newsjack JSON-RPC proxy. That proxy may count tool families
  and outcomes, but must not log arguments or results.

Site:

- Keep telemetry server-side. Do not add a client analytics bundle to the public
  site for this product telemetry path.
- Tighten `apps/site/lib/install-telemetry.ts` before using it as the general
  collector:
  - replace raw user agent with user-agent family;
  - replace full referrer with referrer domain;
  - allowlist query params;
  - reject arbitrary metadata;
  - set raw-event retention.

Tests:

- Go unit tests for consent parsing, env precedence, config precedence,
  allowlisted event building, redaction, JSONL queue behavior, and command
  wrapper outcomes.
- Site tests for endpoint validation, payload caps, event allowlists, unknown
  property rejection, and delete-token behavior.
- Fixture tests for detector telemetry summaries that prove topics, queries,
  profile fields, feeds, titles, URLs, and raw errors are absent.

## User-facing Transparency

Ship the feature with public docs and local commands, not only a privacy-policy
paragraph.

Users should be able to:

- Run `newsjack telemetry status` to see mode, local queue path, last flush time,
  last flush status, and whether a community installation id exists.
- Run `newsjack telemetry inspect` to print the local JSONL queue and recent
  sent cursor state.
- Run `newsjack telemetry off` to disable future local capture and remote sends.
- Run `newsjack telemetry delete --local` to remove local telemetry files.
- Run `newsjack telemetry delete --remote` in community mode to delete remote
  rows for that installation id.
- Run `newsjack doctor` and see telemetry health only after telemetry is enabled.
- Read `docs/telemetry.md` and the website privacy page for the exact schema,
  retention, and deletion model.

Remote deletion design:

- Community mode creates a random `installation_id` and random delete token.
- The server stores the installation id and a hash of the delete token.
- `newsjack telemetry delete --remote` sends both values over HTTPS.
- The server deletes raw rows and removes the installation id from aggregate
  joins where feasible.
- Anonymous events have no stable id, so deletion is by retention only.

## Open Questions

- Is a stable random `community` installation id acceptable for retention
  analysis, or should Newsjack only ever ship anonymous aggregate telemetry?
- Should the existing website telemetry stop storing raw user agent, full
  referrer, and full query params before any CLI telemetry work starts?
- Is 90 days for raw events and 13 months for daily aggregates the right
  retention policy?
- Should `DO_NOT_TRACK=1` and browser Global Privacy Control be treated as hard
  off for all telemetry, including website request telemetry?
- Should consent be asked during installer-driven setup only, or also on the
  first interactive non-help command?
- Who owns the privacy contact and remote deletion endpoint operationally?
- Should the telemetry database live in the current Newsjack site database or a
  separate project with stricter access?
- Do we want a JSON-RPC MCP proxy for tool-call patterns, or is bridge
  start/failure enough for v1?
- Should users be allowed to opt into a separate local eval-capture mode that
  stores scrubbed queries locally for quality work, modeled more like gbrain,
  without ever uploading those queries?
