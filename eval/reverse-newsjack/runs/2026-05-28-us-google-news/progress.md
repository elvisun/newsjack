# Reverse Newsjack Test Run - 2026-05-28

Single-target test. Resumable: if `runs/<id>/candidates.json` exists and parses, treat target as complete.

## Targets

| target_id | company | story |
|---|---|---|
| rev-20260528-001 | Specright | European Commission fines Temu $232M for illegal or unsafe product risks |
| rev-20260528-002 | Form Health | CVS restores coverage for Lilly's Zepbound and adds Foundayo |
| rev-20260528-003 | Orgvue | Wix lays off about 20% of staff, citing AI and exchange-rate pressure |
| rev-20260528-004 | TRM Labs | Google engineer charged with using confidential data for Polymarket bets |
| rev-20260528-005 | Profound | OpenAI prepares ChatGPT ads around conversational intent |
| rev-20260528-006 | CloudZero | Snowflake expands $6B AWS partnership tied to enterprise AI demand |
| rev-20260528-007 | Chainguard | IBM and Red Hat commit $5B to open-source security / AI initiative |

## Progress

- 2026-05-28: created folder structure, drafted profile `profiles/specright.json`.
- 2026-05-28: ran detector for rev-20260528-001; exit 0, 4 emitted, 114 all_scored. Verdict: ranking_miss. Wrote run.md. No `summarize-run` subcommand on installed CLI, so summary.json skipped.
- 2026-05-28T15:11Z: drafted profile `profiles/form-health.json` (obesity medicine / GLP-1 access; no headline terms). Ran detector for rev-20260528-002; exit 0, 10 emitted, 127 all_scored. Verdict: pass_top_3 — positions 1-6 are variants of the CVS/Zepbound story; canonical CNBC+Reuters cluster at position 6 with story_size band "high" (63.6). summarize-run succeeded → `runs/rev-20260528-002/summary/{run.md,summary.json}`. Note: earlier remark that "no summarize-run subcommand on installed CLI" was wrong — the command exists and takes positional INPUT plus `--output` and `--markdown` flags.
- 2026-05-28T15:35Z: updated eval methodology to separate primary recall from ranking quality. Primary detector runs now use `--limit 0`; top-3/top-5 are secondary rank buckets.
- 2026-05-28T15:45Z: drafted profiles and ran rev-20260528-003 through rev-20260528-007 in parallel with `--limit 0`. All five exited 0, all five primary recall passed. Rank buckets: four `top_3`, one `below_10` (`rev-20260528-005`, OpenAI/ChatGPT ads). Wrote `results/overlap_matrix_003_007.csv` and `results/summary_003_007.md`.
- 2026-05-28T16:05Z: ran full-pipeline smoke on 8 Orgvue candidates under `full-pipeline-smoke/orgvue-mixed/`: coarse relevance, `filter-apply`, story-origin smoke findings, `origin-apply`, final rubric report, and `summarize-run`. Outcome: 0 `pitch_now`, 4 `develop_angle`, 3 `monitor`, 1 coarse reject.

## Ten-profile expansion batch

Started: 2026-05-29T00:35:34Z
Profiles: Validic, Z2Data, Brave, Credo AI, Sardine, Constructor, TollBit, Recurrent, FairNow, Qmerit.
- 2026-05-29T00:36:12Z rev-20260528-008-validic complete
- 2026-05-29T00:36:14Z rev-20260528-012-sardine complete
- 2026-05-29T00:36:14Z rev-20260528-009-z2data complete
- 2026-05-29T00:36:15Z rev-20260528-010-brave complete
- 2026-05-29T00:36:15Z rev-20260528-011-credo-ai complete
- 2026-05-29T00:36:51Z rev-20260528-011-fairnow complete
- 2026-05-29T00:36:51Z rev-20260528-015-qmerit complete
- 2026-05-29T00:36:51Z rev-20260528-014-tollbit complete
- 2026-05-29T00:36:53Z rev-20260528-015-recurrent complete
- 2026-05-29T00:36:54Z rev-20260528-013-constructor complete
