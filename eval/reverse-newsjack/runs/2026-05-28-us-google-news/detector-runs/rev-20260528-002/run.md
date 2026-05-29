# rev-20260528-002 — Form Health × CVS restores Zepbound / adds Foundayo

- CLI: ~/.newsjack/bin/newsjack (version 0.2.0-go+9a07d9f317df)
- Sources: news_search,x  | lookback-days: 1 | depth: quick | limit: 120
- Flags: --include-all-scored --no-x-trends --emit json (no --save, no --new-only)
- Exit: 0; stderr.log empty
- candidates.json: 10 emitted signals; debug.all_scored_signals: 127
- Story seed: CVS Caremark restores coverage for Eli Lilly's Zepbound and adds Foundayo (oral obesity pill) to formularies
- summarize-run: succeeded → see summary/run.md and summary/summary.json

## Emitted signals (top 10, in queue order)

| # | queue | profile_match | story_size | source | title |
|---|---|---|---|---|---|
| 1 | 62.4 | 0.114 | 41.4 (moderate) | news_search | CVS Expands Access to Eli Lilly's Obesity Medicines |
| 2 | 62.2 | 0.103 | 43.6 (moderate) | news_search | CVS Caremark Delivers Affordability and Access to GLP-1 Weight Management Medications with Expanded Coverage Options |
| 3 | 61.9 | 0.085 | 40.2 (moderate) | news_search | CVS to expand access to GLP-1 weight-loss drugs Zepbound and Foundayo |
| 4 | 61.7 | 0.070 | unknown | news_search | CVS expands GLP-1 drug coverage, adding Zepbound and easing access to affordable weight management meds. |
| 5 | 61.7 | 0.071 | 41.8 (moderate) | news_search | CVS Caremark adds Zepbound and Foundayo to insurance coverage |
| 6 | 61.6 | 0.065 | 63.6 (high) | news_search | CVS to restore coverage of Zepbound, add Eli Lilly's obesity pill to drug plans (CNBC + Reuters cluster) |
| 7 | 61.6 | 0.063 | unknown | news_search | France 'first EU country to cover obesity drugs' |
| 8 | 61.5 | 0.063 | 48.4 (high) | news_search | Medicare will expand coverage for GLP-1 drugs. How this affects you |
| 9 | 58.7 | 0.057 | 40.6 (moderate) | news_search | Gala Health: High-Quality Compounded GLP-1 Weight Loss & Hormone Therapy HRT Support |
| 10 | 58.4 | 0.054 | unknown | x | Doctor Warns of Ozempic Muscle Loss and Bone Risks Amid Weight Loss Boom |

## Originating story coverage

- Emitted: 6 of the top 6 emitted signals (positions 1-6) are variants of the CVS-restores-Zepbound / adds-Foundayo story.
- Best/canonical match: position 6 — "CVS to restore coverage of Zepbound, add Eli Lilly's obesity pill to drug plans" — CNBC + Reuters cluster, story_size band **high** (score 63.6, coverage_spread 0.427, strongest_outlet 0.892).
- debug.all_scored_signals: ~58 CVS-related variants and ~83 broader GLP-1/Lilly variants were scored. The originating story is heavily represented across PR Newswire (Caremark's release), CNBC, Reuters, NBC News, Quartz, StocktTitan, StatNews, etc.

## Verdict

`pass_top_3` — the originating CVS/Zepbound story is surfaced at the top of the emitted set. Multiple cluster variants occupy positions 1-6, with the highest-story-size CNBC/Reuters cluster emitted at position 6. Form Health's profile (obesity medicine, GLP-1 access, PBM coverage, Zepbound/Wegovy terms) cleanly matches the news vocabulary, and the story's traffic-weighted size dominates the lookback window.

## Notes / failure modes

- No failures. stderr.log empty. summarize-run produced both summary.json and summary/run.md.
- diagnostics reports 1 source error (single-source; not affecting recall here).
- Profile_match values are all small in absolute terms (0.05–0.11) but the relative ordering is correct and traffic-weighted story_size carries position 6 to "high" band.
- search_terms include the drug names (Zepbound, Wegovy, tirzepatide) and "PBM formulary weight loss" / "GLP-1 coverage" but NOT the article headline phrasing ("CVS restores coverage", "Foundayo"), per the constraint.
