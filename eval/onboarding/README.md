# Onboarding Eval

Behavioral eval for the first-run experience of an agent that just had newsjack
installed (or had the repo handed to it). It checks that the agent follows the
slow-start front door — `README.md` → [`docs/getting-started.md`](../../docs/getting-started.md)
— instead of the three failure modes a real beta tester reported:

1. **Capability dump** — reciting the whole ~16-skill menu on first contact.
2. **Premature key complaint** — volunteering up front that a Medialyst/X key is
   missing, before any step needs it.
3. **Hard Medialyst dependency** — stalling or demanding a key for news search
   instead of falling back to host web search.

This eval **grades the behavior, it does not change the docs.** If a case fails,
the fix is a better case or a logged finding — not softening an assertion to make
it pass.

## Design

Six scenarios across two classes:

- **Vague first contact** (`1`, `2`) — where the slow start must fire: short
  orientation, 3-4 on-ramps, one step at a time, no key complaint.
- **Known intent** (`3`-`6`) — where the agent should skip the menu and route to
  the right skill, and where any news lookup must use the host-search fallback
  and proceed without a key:
  - `3` find journalists → `find-journalists` (local artifact + host byline search)
  - `4` is this newsworthy → `newsworthiness-check` (anti-inflation + host pickup check)
  - `5` set up monitoring → `newsjack-monitor-setup` (X token optional, raised late)
  - `6` search the news → `news-search` (host web search, freshness best-effort)

Format follows the repo's `newsworthiness-check` eval (`evals.json` with
`id`/`eval_name`/`prompt`/`expected_output`/`assertions`). Unlike that one,
onboarding output is **conversational**, so assertions are `type: "judgment"` and
grading is an **LLM judge against the per-case criteria**, not a strict JSON
parser. There is intentionally no `grade.py`.

## How to run

1. **Executor (blind to assertions):** an agent reads `README.md`,
   `docs/getting-started.md`, and the front-matter + opening of the relevant
   skill, then responds to each `prompt` *as the freshly-installed agent*. Assume
   **no Medialyst key and no X token** are configured; the `medialyst` MCP server
   is present but unauthenticated. Save each response under
   `runs/<date>-<label>/<id>-<eval_name>.md`.
2. **Judge:** a second agent scores each saved response against that case's
   `assertions`, emitting `{id, passed, evidence}` per assertion and a per-case
   PASS/FAIL.

The two roles must be separate agents so the executor cannot see the answer key.

## Result (2026-06-04 validation)

First run, executed against the on-disk docs on branch `feat/agent-onboarding`.
Full transcript in [`runs/2026-06-04-validation/verdict.md`](runs/2026-06-04-validation/verdict.md).

| Scenario | Verdict | Reason |
|---|---|---|
| 1 what-can-you-do | PASS | 4 on-ramps, no dump, no key complaint |
| 2 just-installed | PASS | Slow-start menu, routes to getting-started not the setup skill |
| 3 find journalists | PASS | find-journalists, local artifact + host byline search |
| 4 newsworthiness | PASS | newsworthiness-check, host pickup check, no key complaint |
| 5 set up monitoring | PASS | newsjack-monitor-setup, X token raised optional + late |
| 6 search the news | PASS | news-search, host web fallback w/ freshness caveat |

All six pass against the on-disk docs. The run also surfaced two minor doc
tightenings that were applied in the same change (a "don't recite this on first
contact" note on the README capability section, and rewording `.env.example`'s
Medialyst line from "Required" to "Optional"), plus one *environmental*
non-defect: the executor's own machine still had the **pre-rename installed
skills** (`newsjack-setup`), which resolves on reinstall/release and is not a
repo issue.
