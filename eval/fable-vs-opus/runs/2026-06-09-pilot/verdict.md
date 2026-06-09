# Pilot verdict — 2026-06-09

**Goal of the pilot:** prove the full pipeline works end-to-end before scaling to
50 brands — clean-context subagent generators (Opus 4.8, Fable 5), a blind
`codex exec` GPT-5.5 judge grounded in `meanest-editor`, counterbalanced
orderings, and aggregation. **It works.**

**Brand:** Notion (offline-first + customer-managed encryption keys launch).
**Judgments:** 2 (one brand × both orderings).

## Numbers (apparatus-sizing, not a conclusion at n=2)

| | Fable 5 | Opus 4.8 |
|---|---|---|
| Head-to-head wins | 1 | 1 |
| news_value | 4.50 | 4.50 |
| distinctness | **5.00** | 3.50 |
| journalist_shape | 5.00 | 4.50 |
| grounding | 3.50 | 3.50 |
| anti_slop | 4.50 | 4.50 |
| proof_rigor | 4.50 | **5.00** |
| usefulness | 4.50 | 4.50 |
| **overall** | **4.50** | 4.29 |

**Position bias: slot A won 2/2 (1.00 vs 0.50 unbiased).** The judge preferred
whichever set was shown first, in both runs. De-biased, this brand is a 1–1 tie.

## What the judge actually said

- **For Fable:** its fourth angle — *"Notion ships offline editing, the feature
  cloud-native software spent a decade treating as optional"* — is "a real second
  story," giving the founder four distinct press lanes instead of three jackets
  on the same bank-pilot story.
- **Against Fable:** *"What it took for a 100M-user productivity app to clear a
  bank's security review"* overstates the record — the update says *pilots*, not
  cleared review. "approve and deploy" overclaims (a pilot is not deployment).
- **For Opus:** stronger proof discipline — it makes "what does 'pilot' actually
  mean" the point, "exactly the question a real reporter will ask."
- **Against Opus:** refusing the standalone offline angle as "just a feature
  note" was the judge's flagged miss; and its first/third angles "orbit the same
  bank-pilot fact."

## Takeaways for the 50-brand run

1. **Both orderings are mandatory** — position bias here was total; one ordering
   would have manufactured a winner.
2. The interesting, repeatable signal is **per-dimension** (distinctness vs
   proof_rigor), not the win/loss flip that position bias drives.
3. Need many brands (the plan is 50) before the position-averaged head-to-head
   and dimension deltas stabilize. `scripts/run.js` is ready to do that.
