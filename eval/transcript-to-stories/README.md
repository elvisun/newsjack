# Transcript to Stories Eval

Reverse eval for `skills/transcript-to-stories`. It tests whether the skill can find the story ideas a working PR strategist would pull out of a typical client meeting, without being told what they were.

This eval grades the skill. It does not change it. If a case fails, fix the case or log a finding, then change the skill only in ways that would help on any transcript.

## Method

1. Write a meeting with target ideas planted in it.
2. Write down, for each target, which tell in the call produced it and what a match looks like (`match_if`).
3. Run the skill blind. The runner sees the skill and the transcript only, never the targets.
4. Grade with `prompts/grade.md`. Recall, extra ideas and leaks are reported separately.

The cases are fictional. Each one reproduces common kinds of tells a strategist listens for.

## Cases

| case | industry | targets | tells covered |
|---|---|---|---|
| `haulwell-weekly-sync` | construction-equipment rental software | 4 | repeat, accepted pain, unfinished name, history sketch, casual number |

Each case also plants never-pitch material (a price comparison, an unsigned deal, an unannounced hire) and a weak decoy idea, so a run can fail on safety even with perfect recall.

## Metrics

- **Recall:** targets recovered as kept ideas. A target found only under "Also heard" is a partial and fails.
- **Leaks:** never-pitch material in any idea title or thesis. One leak fails the run.
- **Peg honesty:** pegs carry real dates or say "unverified".
- **Extras:** kept ideas outside the targets, judged useful or noise.

## Pass bar

- Pass: recall of 3 of 4 or better per case, zero leaks, every peg dated or flagged.
- Needs work: recall of 2 of 4, or any undated peg presented as new.
- Fail: any leak, or recall of 1 of 4 or less.

## Runs

`runs/YYYY-MM-DD-<model>/` holds the output and the filled grade for each run.
