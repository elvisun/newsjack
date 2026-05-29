# rev-20260528-001 — Specright × Temu fine

- CLI: ~/.newsjack/bin/newsjack (version 0.2.0-go+9a07d9f317df)
- Sources: news_search,x  | lookback-days: 1 | depth: quick | limit: 120
- Flags: --include-all-scored --no-x-trends --emit json (no --save, no --new-only)
- Exit: 0; stderr.log empty
- candidates.json: 4 emitted signals; debug.all_scored_signals: 114
- Story seed: European Commission fines Temu $232M (~€200M) for illegal/unsafe product risks under the DSA

## Emitted signals (top 4)

| # | profile_match | story_size | source | title |
|---|---|---|---|---|
| 1 | 0.065 | 0.381 (moderate) | news_search | Carbonfact Acquires Fashion Sustainability Software Provider Vaayu |
| 2 | 0.072 | n/a (unknown) | news_search | Product Safety Consulting CEO Featured for In Compliance Magazine Cover Story |
| 3 | 0.060 | 0.359 (moderate) | news_search | New rail data requirement a 'win' for shippers, expert says |
| 4 | 0.138 | n/a (unknown) | x | How Supply Chain Converts Data Into Profits / Supply Chain Transparency Cheat Sheet |

## Temu story in `debug.all_scored_signals`

35 items whose title contains "Temu" were scored but not emitted. Best representative variants:

- "EU Tests Limits of Platform 'Risk Assessments' with €220 Million Temu Fine" — news_search, techpolicy.press, profile_match 0.014, story_size 0.375 (moderate)
- "EU Commission fines Temu €200m under DSA over illegal product risks"
- "Chinese online retailer Temu hit with $232 million fine over unsafe toys and electronics"
- "Temu Hit With Fine in E.U. Over Sales of Unsafe Goods"
- "EU Fines Temu €200 Million for Risk Assessment Failures Under DSA" — x_news cluster, profile_match 0.034, momentum 0.252

## Verdict

`ranking_miss` — the originating story was scored (35 variants present) but did not survive ranking into the emitted set. Failure mode is profile-match: Temu/DSA stories scored 0.014–0.034, while emitted items scored 0.06–0.138. Specright's standing/topics vocabulary (specification management, packaging traceability, GPSR) didn't overlap with the article terms (DSA, "risk assessments", marketplace, illegal listings).
