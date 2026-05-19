# journalist-fit-check

Gate one pitch against one journalist before sending. Returns `fit`, `soft-fit`, `no-fit`, or `unknown` with recent byline anchors, decay checks, and concrete edits when a soft-fit can be rescued.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source operating system for agentic PR.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then provide one journalist and one pitch.

### CLI (coming)

```bash
newsjack fit-check "Casey Newton, Platformer" pitch.md
newsjack fit-check "https://www.platformer.news/..." < pitch.txt
```

## Natural-language invocations

All of these should trigger the skill:

- "Does this pitch fit Casey Newton?"
- "Fit-check this for Maxwell Zeff at TechCrunch"
- "Should I pitch this journalist?"
- "Is this reporter a fit or am I stretching?"
- "Check this journalist before I send"
- "Can I add this person to the list?"
- "Would this Substacker care about this pitch?"

## What you get back

A JSON-shaped verdict covering:

- **Verdict**: `fit`, `soft-fit`, `no-fit`, or `unknown`
- **Confidence**: calibrated to recency, anchor strength, and angle overlap
- **Reasoning**: 2-3 sentences anchored to named recent work
- **Anchor pieces**: title, URL, date, age, and relevance note
- **Suggested changes**: only for soft-fits that can be fixed
- **Refusal block**: stale data, unresolved journalist, slop tells, missing time, or uncertainty
- **Decay block**: last verified byline date and warning when older than 60 days
- **Retrieval notes**: what the skill checked

## Files in this skill

| File | Purpose |
|------|---------|
| `SKILL.md` | Entry point. Agent persona, workflow, refusal rules, output format. |
| `rubric.md` | Hard gates, scored checks, confidence calibration, slop patterns. |
| `examples.md` | Five worked before/after examples. |
| `README.md` | This invocation guide and file index. |

## License

MIT
