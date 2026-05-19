# reactive-comment

Triage inbound journalist source queries and draft a response only when the
user is a genuine fit. Kills weak fits, asks for missing proof, never
auto-sends.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source
operating system for agentic PR.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then paste one journalist query
plus the user's expertise profile.

### CLI (coming)

```bash
newsjack reactive query.md
newsjack reactive < clipboard
```

## Natural-language invocations

All of these should trigger the skill:

- "Should I respond to this HARO query?"
- "Draft a Source of Sources reply, but only if I fit"
- "React to this Qwoted request"
- "Is this Featured.com query worth answering?"
- "Kill or draft this JournoRequest"
- "Can I credibly answer this reporter request?"
- "Write the response if it is actually on-topic"

## What you get back

A single YAML verdict:

- **`draft`** - fit score, rationale, response draft, provenance, slop check,
  cap stamp, and manual-send next action
- **`kill`** - fit score, kill reason, "why this is not your fight," optional
  better move, and cap stamp
- **`ask`** - exact missing facts needed before the skill can draft or kill

## Files in this skill

| File | Purpose |
|------|---------|
| `SKILL.md` | Entry point. Agent persona, triage flow, refusal rules, output format. |
| `rubric.md` | Weighted fit scoring and hard refusal gates. |
| `examples.md` | Worked before/after examples for draft, kill, ask, and cap refusal. |
| `README.md` | This file. |

## License

MIT
