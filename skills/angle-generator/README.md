# angle-generator

Turn a company update into 3-7 structurally distinct, journalist-shaped story angles. Refuses duplicate rephrasings, invented facts, named-journalist guesses, and AI-marketing slop.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source operating system for agentic PR.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then paste the company update, supporting facts, company profile, and current timestamp.

### CLI (coming)

```bash
newsjack angles update.json
newsjack angles "We raised a seed round" --company company.json
```

## Natural-language invocations

All of these should trigger the skill:

- "What are the angles for this funding announcement?"
- "We have a product launch. What stories can we pitch?"
- "Find the journalist-shaped angles in this update."
- "This pitch got rejected because the angle is weak. Give me better ones."
- "I have a newsjack signal. What are the angles?"
- "Give me angles, but don't let me send spam."
- "How many real angles are in this partnership announcement?"

## What you get back

A JSON object covering:

- **Angles** - 3-7 only when they honestly exist.
- **Journalist shape** - beat, outlet archetype, evidence they care, and who not to target.
- **Why now + decay** - honest timing, including `30min`, `4hr`, `24hr`, `week`, `month`, or `evergreen`.
- **Required proof** - the data, quote, source, named customer, or document needed before pitching.
- **Refused angles** - duplicates, slop, hallucinated facts, off-beat ideas, and other killed candidates.
- **Uncomfortable questions** - the missing facts that decide whether the pitch is real.
- **Next skill** - usually `journalist-fit-check`, `meanest-editor`, `newsjack-detector`, or `null`.

## Files in this skill

| File | Purpose |
|------|---------|
| `SKILL.md` | Entry point. Agent persona, rules, workflow, output format. |
| `rubric.md` | Scored checks mapped to the source design doc. |
| `examples.md` | 4 realistic before/after examples. |
| `README.md` | This file. |

## License

MIT
