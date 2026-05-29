# Reverse Eval Run: 2026-05-28 US Google News

Seed surface: US Google News Business and Technology, collected on 2026-05-28.

This folder is the durable, committed copy of the reverse-eval run. The original scratch workspace was `.tmp/reverse-newsjack/2026-05-28-test/`, which is ignored by Git.

## Contents

- `targets.md`: seed stories and expected entities.
- `company-candidates.md`: reverse-profile candidate matrix.
- `profiles/`: temporary company profiles used for detector runs.
- `detector-runs/`: per-profile detector artifacts, including `command.txt`, `profile.json`, `candidates.json`, `stderr.log`, and summary artifacts when generated.
- `full-pipeline-smoke/`: end-to-end Orgvue smoke test with coarse relevance, origin gate, final report, and rendered `run.md`.
- `results/`: recall/ranking scorecards.
- `progress.md`: chronological run notes.

## Key Results

- First scored batch, rev-20260528-003 to rev-20260528-007: 5/5 primary recall, 4/5 top-3.
- Ten-profile expansion batch: 10/10 primary recall, 5/10 top-3, 8/10 top-5.
- Useful ranking drill-ins: `rev-20260528-011-fairnow` and `rev-20260528-015-qmerit`.

Start with:

- `results/summary_003_007.md`
- `results/summary_10_more_profiles.md`
- `full-pipeline-smoke/orgvue-mixed/run.md`
