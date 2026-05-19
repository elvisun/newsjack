# voice-extractor

Capture a user's actual writing voice from real samples, store it locally as a `voice.yaml` fingerprint, and enforce it on newsjack drafts so they stop reading like generic AI copy.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source operating system for agentic PR.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then ask to set up, refresh, or check a voice fingerprint.

### CLI (coming)

```bash
newsjack voice init
newsjack voice refresh
newsjack voice check draft.md
newsjack voice use <profile_id>
```

## Natural-language invocations

All of these will trigger the skill:

- "Set up my voice"
- "Build my voice fingerprint"
- "Make this sound like me"
- "Check this against my voice"
- "This draft sounds like AI"
- "Refresh my voice profile"
- "Use my voice before drafting this pitch"

## What you get back

A local voice fingerprint covering:

- **Cadence** - sentence length, paragraph rhythm, short-burst versus flowing style
- **Mechanics** - contractions, em-dashes, Oxford comma, punctuation rates, capitalization quirks
- **Openers and closers** - phrases the user actually uses and stock phrases to ban
- **Idiom set** - signature phrases, signature words, real hedges, banned hedges
- **Anti-slop rules** - global and user-specific banned words and structures
- **Check output** - pass/fail, rule ids, spans, severities, fix hints, drift stats

## Files in this skill

| File | Purpose |
|------|---------|
| `SKILL.md` | Entry point. Agent instructions, modes, refusals, and output formats. |
| `rubric.md` | Scored extraction/check/enforce criteria mapped to the source design doc. |
| `examples.md` | 4 worked before/after examples across init, enforce, check, and mixed-register cases. |
| `README.md` | This file. |

## License

MIT
