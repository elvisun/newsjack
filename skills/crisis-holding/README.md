# crisis-holding

Draft holding statements for brewing or live incidents without inventing facts, overpromising, or talking past legal exposure.

Part of [newsjack](https://github.com/elvisun/newsjack), the open-source operating system for agentic PR.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then paste the structured incident intake.

### CLI (coming)

```bash
newsjack crisis
newsjack crisis-holding
```

## Natural-language invocations

All of these will trigger the skill:

- "I have a brewing crisis. Walk me through a holding statement."
- "Write a holding statement for this incident."
- "We have inbound press on a customer-impacting outage."
- "Our social post is getting backlash and reporters are asking."
- "Draft language for counsel to review."
- "What should we not say about this incident?"

## What you get back

- **Three statements** - short, medium, and cautious-legal-pass.
- **Journalist Q&A scaffold** - likely questions, posture, rationale, and holding lines.
- **What-not-to-say list** - exact risky phrases and why they fail.
- **Legal-counsel flag** - binary gate with the trigger named.
- **Decay tag** - issued time, valid-until time, and refresh trigger.

The skill refuses publishable drafts when counsel is required and not yet engaged. Use `--counsel-review-mode` only for language that will be reviewed before publication.

## Files in this skill

| File | Purpose |
|---|---|
| `SKILL.md` | Entry point. Persona, intake, legal gate, workflow, and output format. |
| `rubric.md` | Hard gates, scored criteria, banned phrases, and legal trigger checks. |
| `examples.md` | Four worked before/after crisis examples. |
| `README.md` | This invocation guide and file index. |

## License

MIT
