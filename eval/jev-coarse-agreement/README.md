# Jev vs LLM worker — coarse-filter agreement

Checks whether `newsjack coarse-filter --engine jev` makes the same survive-or-drop calls as the low-cost LLM worker path on real detector pools, before Jev is trusted in a scheduled run.

## What it uses

Local detector fixture runs already hold paired inputs and outputs for the coarse pass:

- `fixtures/newsjack-detector-agent/runs/<run>/coarse_chunk.N.json` — the `{profile_context, signals}` chunk a worker judged
- `fixtures/newsjack-detector-agent/runs/<run>/coarse_decisions.N.json` — the worker's decisions for that chunk

Those run folders are local artifacts and are not committed. The 2026-06-03 runs for simular, slite, and localfalcon give roughly 200 labeled signals. The labels are LLM judgments, not ground truth; the eval measures agreement and lists every disagreement so a human can read them.

## Run it

```bash
export TYPESAFE_API_KEY=...            # or: newsjack auth set-typesafe --key ...
eval/jev-coarse-agreement/run.sh fixtures/newsjack-detector-agent/runs/20260603T175505Z_simular \
                                fixtures/newsjack-detector-agent/runs/20260603T175505Z_slite \
                                fixtures/newsjack-detector-agent/runs/20260603T175505Z_localfalcon
python3 eval/jev-coarse-agreement/compare.py eval/jev-coarse-agreement/runs/<timestamp>
```

`run.sh` calls the CLI once per chunk and writes `jev.<run>.<N>.json` next to a copy of the worker decisions under `runs/<timestamp>/`. `compare.py` reports:

- decision agreement (exact) and survive-vs-drop agreement (keep or monitor_only vs reject)
- reason agreement on signals where both engines agreed on the decision
- recall misses: every signal the worker kept that Jev rejected, one per line with title and both rationales
- junk let through: every signal the worker rejected that Jev kept
- Jev cost and latency from the `engine` blocks

## Acceptance bar

- Zero recall misses on any signal the worker marked `keep` with `confidence: high`.
- Survive-vs-drop agreement at or above 85%.

Below that bar, tune the criteria wording in `apps/cli/cmd/newsjack/coarse_filter_questions.json` (pass a copy with `--questions` to iterate without a rebuild), not the post-rules. Record the run's numbers in the plan doc when they exist.

## Results so far

- `runs/20260918T162324Z/` — 176 signals across simular, slite, localfalcon; 0 failures; $0.013; 8.3 s. Zero recall misses on high-confidence keeps (pass). Survive-vs-drop agreement 76.7% (below the 85% bar), with every disagreement in the recall-safe direction: Jev monitors or keeps what the worker rejected. Full breakdown in `docs/2026-09-18-jev-coarse-filter-plan.md`.
