# Media List Manager Rubric

Use this rubric before returning a list, share link, or management summary.

## Hard Gates

### Gate 1 - Current-time anchor

Fail when the workflow depends on recency and no current time is available.

Result: continue only for non-recency work and mark all recency-sensitive rows `research-needed`.

### Gate 2 - Standing missing

Fail when the client has no credible reason to comment on the angle.

Result: do not build the list. Send the user to `newsworthiness-check` or `angle-generator`.

### Gate 3 - No anchor evidence

Fail when a journalist row lacks a specific article, profile, newsletter issue, public query, or other dated evidence anchor.

Result: `research-needed` at best. It cannot be `fit`.

### Gate 4 - Spray pattern

Fail when the user asks for a large undifferentiated list, same-body blast list, or broad beat database.

Result: refuse the broad list and offer a smaller segmented first wave.

### Gate 5 - Fabrication

Fail when an anchor title, date, URL, journalist identity, outlet, email, or credential is guessed.

Result: cut or mark `research-needed`; never smooth over uncertainty.

## Scored Criteria

Score each list 0-2 on each criterion. Hard gates override the score.

| Criterion | 0 | 1 | 2 |
| --- | --- | --- | --- |
| Angle clarity | Generic pitch or unclear story | Usable but broad | Specific story with proof and decay window |
| Journalist shape | Outlet category only | Beat described but loose | Specific beat, format, and story type |
| Anchor evidence | Missing or stale | Present but indirect | Recent, dated, URL-pointed, relevant |
| Fit reasoning | Vibes or database tag | Plausible but thin | Specific bridge from anchor to angle |
| List size | Volume-first | Slightly broad | Small first wave with clear rationale |
| Segmentation | None | Basic beat buckets | Distinct segments with distinct angles |
| Anti-spam compliance | Same-body blast risk | Some weak rows remain | Weak rows cut or marked for research |
| MCP audit | No sync status | Partial status | Tools used, IDs captured, verification performed |
| Management hygiene | Columns/views chaotic | Some review fields | Clear columns, statuses, and review views |
| Next step | Vague | Plausible | Concrete review or sync action |

## Verdicts

- `ready-for-review`: 16-20 points, no hard gates, and all first-wave rows have anchors.
- `needs-research`: 10-15 points or several rows lack anchors.
- `not-list-ready`: under 10 points, standing missing, angle unclear, or spray pattern present.

## Row Status Rules

- `fit`: exact or near-exact recent anchor, clear beat overlap, and a pitch bridge the user can actually use.
- `soft-fit`: real adjacent anchor, but the pitch needs a specific edit or narrower angle.
- `research-needed`: journalist identity, current role, anchor, or date is unresolved.
- `cut`: wrong beat, stale, duplicate, weak evidence, unsafe hook, or obvious database filler.

Do not use `fit` for outlet-level relevance. The row belongs to a person, not a publication logo.
