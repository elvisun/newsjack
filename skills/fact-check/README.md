# fact-check

Extract factual claims from PR copy, verify each one independently, attach
citations, and warn when certainty is low.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source
operating system for agentic PR.

## Naming

The source research used the name `fact-guard` because the behavior is a
pre-send safety gate. This repo uses `fact-check` for the folder and skill
name because it matches the public command shape users will reach for:
`/newsjack fact-check`. The guardrail posture remains the product doctrine.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then paste the draft, source URLs
if any, and a current timestamp.

### CLI (coming)

```bash
newsjack fact-check pitch.md
newsjack fact-check --text "Draft text" --current-time 2026-05-18T14:00:00Z
```

## Natural-language invocations

All of these should trigger the skill:

- "Fact-check this pitch."
- "Verify every claim and add citations."
- "Is this safe to send?"
- "Which claims in this draft are unsupported?"
- "Check whether these statistics and bylines are real."
- "Find hallucinated experts or stale titles in this response."

## What you get back

A Markdown fact-check report covering:

- **Fact-check verdict** - short summary of whether the draft is safe, risky,
  or blocked by evidence gaps.
- **Facts & Citations** - numbered claim list with status, citation URLs, and
  notes on ambiguity or staleness.
- **Warning** - residual risk, unresolved claims, stale-source risk, and human
  review items.

Statuses are:

- `Verified`
- `Disputed`
- `Unverifiable`
- `Missing source`

For machine-readable callers, those map to `verified`, `disputed`,
`unverifiable`, and `missing-source`.

## Ideal Runtime Shape

This is an atomic skill, but it is designed for multi-agent use:

1. Claim extraction agent.
2. Verification and source lookup agent.
3. Final adjudication and consistency pass.

A single-agent runtime can still use the skill by running the same stages in
sequence.

## Files in this skill

| File | Purpose |
|------|---------|
| `SKILL.md` | Entry point. Doctrine, workflow, status ladder, output format. |
| `rubric.md` | Scored checks for extraction, citations, uncertainty, and warnings. |
| `examples.md` | Worked examples, including mixed-result and failure cases. |
| `README.md` | This file. |

## License

MIT
