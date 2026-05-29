# Reverse Eval Summary: 10 More Profiles

Run window: 2026-05-29 00:35-00:37 UTC, using the 2026-05-28 US Google News target set.

Detector mode: high-recall, `--limit 0`, `--include-all-scored`, `news_search,x`, local `go run ./cmd/newsjack`.

## Results

| Run | Profile | Target | Primary Recall | Rank Bucket | Best Match |
|---|---|---|---:|---|---|
| rev-20260528-008-validic | Validic | Oura Ring 5 wearable health signals | pass | top_5 | `Blood pressure tracking & GLP-1 tools central to new Oura Ring 5` |
| rev-20260528-009-z2data | Z2Data | Steam Deck price jump from component costs | pass | top_3 | `Valve hikes Steam Deck prices following global hardware component crises` |
| rev-20260528-010-brave | Brave | Websites infer visitor behavior from SSD activity | pass | top_3 | `New SSD-based web tracking technique can infer activity after single site visit` |
| rev-20260528-011-credo-ai | Credo AI | AI hiring algorithm racial disparities | pass | top_5 | `AI hiring tools can be biased, new study finds` |
| rev-20260528-012-sardine | Sardine | Robinhood enables AI agents to trade and buy | pass | top_3 | `Robinhood Is About to Let AI Agents Trade Stocks and Spend Money on Your Behalf` |
| rev-20260528-013-constructor | Constructor | AWS retail AI shopping assistants | pass | top_3 | `AWS launches agentic shopping assistant for retailers` |
| rev-20260528-014-tollbit | TollBit | CNN sues Perplexity over AI copyright | pass | top_3 | `CNN sues Perplexity for copyright infringement over AI-generated content` |
| rev-20260528-015-recurrent | Recurrent | Rivian R2 delivery/order invites | pass | top_5 | `Rivian R2 Order Invitations and First Deliveries Begin June 9` |
| rev-20260528-011-fairnow | FairNow | AI hiring algorithm racial disparities | pass | below_10 | `AI hiring tools face fresh scrutiny after study finds racial bias` |
| rev-20260528-015-qmerit | Qmerit | Rivian R2 delivery/order invites | pass | below_10 | `Rivian R2 Order Invitations and First Deliveries Begin June 9` |

## Metrics

- Profiles run: 10
- Primary recall rate: 10/10
- Top-3 ranking rate: 5/10
- Top-5 ranking rate: 8/10
- Below-10 but emitted: 2/10
- Ranking misses: 0
- Source misses: 0

## Drill-In Notes

- The cleanest wins are Z2Data, Brave, Sardine, Constructor, and TollBit: all placed direct target-story matches in the top 3.
- Validic, Credo AI, and Recurrent are useful top-5 passes.
- FairNow and Qmerit prove candidate-pool recall but expose ranking sensitivity: broader AI-governance and EV-charging stories outranked the exact target.
- The next useful drill is not retrieval recall; it is final-report quality on the top-5 passes and ranking behavior on the two below-10 alternates.
