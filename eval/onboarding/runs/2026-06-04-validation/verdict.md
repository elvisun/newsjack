# Onboarding eval — 2026-06-04 validation run

Branch `feat/agent-onboarding`. Executor role-played a freshly-installed agent
against the six scenarios, reading only `README.md`, `docs/getting-started.md`,
`.mcp.json`, `.env.example`, and the relevant skill front-matter/openings. No
Medialyst key, no X token; `medialyst` MCP present but unauthenticated.

## Verdict

| Scenario | PASS/FAIL | Reason |
| --- | --- | --- |
| S1 what is this / what can you do | PASS | Short orientation + 4 on-ramps, no capability dump, no key complaint. |
| S2 just installed, help me start | PASS | Same slow-start menu; routes to getting-started, not the setup skill. |
| S3 find me journalists | PASS | Routes to media-list-manager; local-artifact + host byline search; Medialyst framed optional. |
| S4 is hitting 10k users newsworthy | PASS | Routes to newsworthiness-check (pitch mode), anti-inflation; host pickup check; no key complaint. |
| S5 set up monitoring | PASS | Routes to newsjack-monitor-setup; X token raised only at its step, optional. |
| S6 search the news | PASS | Routes to news-search; host web fallback w/ freshness caveat; no stall. |

All six pass as driven by the on-disk docs. The README + getting-started.md +
each skill's fallback language actively suppress the three old failure modes
(capability dump, premature key complaint, hard Medialyst dependency).

## Findings

1. **Applied — README capability section was a latent dump.** `README.md`'s
   "What your agent can do" enumerates ~12 capabilities; an agent paraphrasing it
   instead of following getting-started could reproduce the dump. Fix: added a
   blockquote telling agents not to recite it on first contact.
2. **Applied — `.env.example` said Medialyst was "Required".** Contradicted the
   "optional, fine without" framing. Fix: reworded to "Optional … fine to leave
   blank."
3. **Environmental non-defect — stale installed skills.** The executor's machine
   still had the pre-rename `newsjack-setup` skill registered (under
   `~/.claude/skills`), so a runtime routing off the live registry (not the repo
   file) could still treat "setup" as the front door. This is the on-disk rename
   not yet propagated to an existing install; it resolves on reinstall/release.
   Not a repo defect — the on-disk `newsjack-monitor-setup/SKILL.md` carries the
   correct name and the "defer first-contact to getting-started" deferral.
