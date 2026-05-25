# Newsjack Newsjack Brief

| key | value |
|---|---|
| Status | Detector preview only |
| Generated | 2026-05-25 03:24 UTC |
| Action queue | Needs editorial pass before sharing |
| Detector candidates | 2 |
| Cheap filter | pending |
| Targeted set | pending |

## Review Status

**Not ready to forward as recommendations yet.** The detector run is complete, but the editorial judgment pass has not produced `final_report.md`.

Next action: run the cheap filter, apply it, write `final_report.md`, then rerender this file.

## Candidate Preview

These are the highest-priority signals for review. They are not final pitch recommendations.

1. **Regulators open inquiry tied to AI customer support**
   - Why surfaced: profile relevance; queue 70.8, profile 0.4, major 0.
   - Query: AI customer support
   - Cheap filter: confidence=high, decision=keep, evidence_urls=['https://example.com/news/60f5725c'], rationale=Golden fixture keep, reason=relevant_news, signal_id=e32ebc6ac34ee9d2
   - news search: [Regulators open inquiry tied to AI customer support](https://example.com/news/60f5725c) (2026-05-25)
2. **Experts are reacting to AI customer support**
   - Why surfaced: x posts; queue 64, profile 0.4, major 0.
   - Query: AI customer support
   - Cheap filter: confidence=high, decision=keep, evidence_urls=['https://x.com/example/status/f7f584be'], rationale=Golden fixture keep, reason=relevant_news, signal_id=578832fabe7e6a64
   - x: [Experts are reacting to AI customer support](https://x.com/example/status/f7f584be) (2026-05-25)

## What Was Scanned

- **Profile:** (unknown)
- **Queries:** AI customer support
- **Lookback:** 1 day(s); max item age 24 hour(s); depth quick
- **Coverage:** News search: used (1 item); X News: used (0 items); X trends: not used; X posts: used (1 item); RSS feeds: not configured (0 feeds)

## Run Notes

- This report is a detector preview. Do not treat candidates as approved outreach hooks.

## Appendix: Provenance

### Pipeline

| key | value |
|---|---|
| detector | pending - candidates.json |
| cheap_filter | pending - filter_decisions.json |
| filter_apply | pending - targeted_candidates.json |
| final_report | pending - final_report.md |

### Run Context

- **Input:** `/tmp/newsjack-filtered.json`
- **Profile:** (unknown)
- **Queries:** AI customer support
- **Sources used:** `news_search, x_news, x`

### Detector Counts

| key | value |
|---|---|
| scored | 2 |
| selected | 2 |
| debug rows | 0 |
| debug selected rows | 0 |
| debug unselected rows | 0 |
| debug duplicate rows | 0 |
| source_errors | 0 |

### Selection

| key | value |
|---|---|
| limit | 20 |
| min_major_news | 0.55 |
| min_queue_priority | 40.0 |
| mode | mechanical_floor |

### Lanes

| key | value |
|---|---|
| scored.profile_relevance | 1 |
| scored.x_posts | 1 |
| emitted.profile_relevance | 1 |
| emitted.x_posts | 1 |

### Evidence Sources

| key | value |
|---|---|
| evidence.news_search | 1 |
| evidence.x | 1 |

### Candidate Queue

1. **Regulators open inquiry tied to AI customer support**
   - Why surfaced: profile relevance; queue 70.8, profile 0.4, major 0.
   - Signal ID: `e32ebc6ac34ee9d2`
   - Query: AI customer support
   - Cheap filter: confidence=high, decision=keep, evidence_urls=['https://example.com/news/60f5725c'], rationale=Golden fixture keep, reason=relevant_news, signal_id=e32ebc6ac34ee9d2
   - news search: [Regulators open inquiry tied to AI customer support](https://example.com/news/60f5725c) (2026-05-25)
2. **Experts are reacting to AI customer support**
   - Why surfaced: x posts; queue 64, profile 0.4, major 0.
   - Signal ID: `578832fabe7e6a64`
   - Query: AI customer support
   - Cheap filter: confidence=high, decision=keep, evidence_urls=['https://x.com/example/status/f7f584be'], rationale=Golden fixture keep, reason=relevant_news, signal_id=578832fabe7e6a64
   - x: [Experts are reacting to AI customer support](https://x.com/example/status/f7f584be) (2026-05-25)
