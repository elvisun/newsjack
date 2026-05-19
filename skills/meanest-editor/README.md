# meanest-editor

Roast a pitch or press release with the eye of a veteran PR director. Honest, sharp, constructive — never cruel for its own sake.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source operating system for agentic PR.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then paste your pitch.

### CLI (coming)

```bash
newsjack roast draft.md
newsjack roast < clipboard
```

## Natural-language invocations

All of these will trigger the skill:

- "Roast this pitch"
- "What's wrong with this press release?"
- "Is this any good?"
- "Review my draft"
- "Tear this apart"
- "Give me honest feedback on this pitch"
- "Would a journalist read past the subject line?"

## What you get back

A structured critique covering:

- **Score** (1–10) with a one-word verdict: publishable, workshopable, or start over
- **Top 3 offenses** — quoted directly from your draft
- **Line-by-line critique** — specific, quoted, no padding
- **Suggested lede rewrite** — an actual draft, not a suggestion
- **What to do next** — 2–3 concrete moves

## Files in this skill

| File | Purpose |
|------|---------|
| `SKILL.md` | Entry point. Agent instructions, persona, output format. |
| `rubric.md` | The 13-criterion scoring rubric. |
| `examples.md` | 3 worked before/after examples. |
| `README.md` | This file. |

## License

MIT
