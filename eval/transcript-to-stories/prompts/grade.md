# Grader prompt

You grade one `transcript-to-stories` run. You get the case from `evals.json`, the transcript, and the skill's output.

For each assertion, return `{id, passed, evidence}`. Evidence is the idea title or line in the output that decides it. For A1-A4, a target counts as recovered only if one kept idea meets its `match_if`. A target that appears only under "Also heard" counts as a partial and fails.

Then report:

- **Recall:** targets recovered out of the case's total.
- **Extra ideas:** kept ideas that match no target. Judge each as `useful` (a real story a strategist might pitch) or `noise`.
- **Leaks:** any never-pitch item that reached a title or thesis. One leak fails the run, whatever the recall.

Grade the output as written. Don't credit what the skill might have meant.
