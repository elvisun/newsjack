# News Desk Dealer

Today's Google News headlines judged side by side by Jev and Claude Opus 5, animated, on one screen. Set A stamps every headline with a desk, a story type, and six 0–4 scores that draw its radar: magnitude, velocity, novelty, window, heat, risk. Those weights give a newsworthiness score out of 10, per the newsworthiness-check skill. Set B scores each story per company on standing (none → direct) and journalist shape, then tiers it; a story can land on several desks or none, and you watch the cards fly into the piles. When Jev finishes, the Opus run is stopped so no more tokens are spent.

## Status

Mock engine by default. The UI, runner, question sets, counters, and abort logic are real; the judgments come from keyword heuristics plus a seeded PRNG in `src/engine/mock.ts`, with simulated latency and token usage per model. No API keys are used in this mode.

A live engine is wired but off. It runs Jev against TypeSafe's endpoint (`src/engine/typesafe.ts`) and Claude Opus 5 through OpenRouter with a strict JSON schema built from the same question set (`src/engine/openrouter.ts`). Keys stay on the dev server. Calls ride Vite's HMR WebSocket to a handler in `vite.config.ts` that adds the `Authorization` header from `.env` and forwards upstream, so nothing secret reaches the bundle and the browser's six-connections-per-host cap never throttles the fast model. A plain HTTP proxy at `/api/openrouter` and `/api/typesafe` is kept for curl. To try it:

```bash
cp .env.example .env      # add OPENROUTER_API_KEY and TYPESAFE_API_KEY
pnpm dev                  # then open http://localhost:5180/?engine=live
```

or set `VITE_ENGINE=live` in `.env` to make it the default. Lower the headline count first; a full Opus pass over 398 headlines is on the order of $25, though the run stops when Jev finishes. Failed calls (rate limits, malformed answers) are retried with backoff, then counted in the column header instead of ending the run. Measured live on 2026-09-18: Jev set A about 190 ms and set B (90 questions, ~10k input tokens) about 320 ms per call, roughly 1,200 judgments per second with eight workers; Opus 5 about 4 to 6 s for set A and longer for set B. TypeSafe score answers are probability-weighted floats and are rounded to whole levels.

Press cards show a page screenshot when one exists in `public/shots/`, otherwise a typeset stand-in with the outlet's logo. Screenshots are fetched with ScrapingBee:

```bash
SCRAPINGBEE_API_KEY=... node scripts/screenshot.mjs --limit 20   # try a few first
SCRAPINGBEE_API_KEY=... node scripts/screenshot.mjs              # all of them
```

Each shot costs ScrapingBee credits (JS rendering is on so Google News redirect pages resolve). Thumbnails are 360×270 JPEGs and the manifest is `data/shots.json`.

## Run it

```bash
cd demos/news-desk-dealer
pnpm install
pnpm ingest      # optional: refresh data/headlines.json from Google News
pnpm dev         # http://localhost:5180
```

The two columns sit on their own tinted paper, pink for Jev and peach for Opus. Each header carries a tachometer for judgments per second and a stack of poker chips for spend, a quarter per chip, so the cost gap is visible before you read a number. Headlines are dealt onto the wire with a snap; a "Sound" toggle in the masthead adds synthesised card snaps, chip clinks, and pile thuds (Web Audio, no files, off by default).

Designed for a 1440×900 viewport or larger; nothing scrolls. Speed 1× uses realistic per-call latencies (Jev 120–520 ms; Opus 2.8–5.5 s for set A and 6–12 s for the 90-question set-B fan-out), eight workers each; 4× is a good walkthrough.

## Layout

- `data/headlines.json` — today's pull with outlet domains, committed so the demo runs offline.
- `scripts/ingest.mjs` — dependency-free RSS pull from six Google News topic feeds plus Techmeme.
- `scripts/screenshot.mjs` — ScrapingBee screenshots to `public/shots/`.
- `src/questions.ts` — sets A and B in Jev's Choice / Score / Noul shape. The six radar axes and the two standing axes are `score` questions with five ordered levels each; criteria are lifted from the public Newsjack skills.
- `src/clients.ts` — fifteen public companies, one per industry, used as illustrative desks. Not clients.
- `src/engine/` — adapter interface, mock adapter, live adapters (`typesafe.ts`, `openrouter.ts`), runner (concurrency pool, abort, post-rules, per-call error tolerance), model registry with list prices and the mock/live switch.
- `src/components/` — `Card` (press card), `Radar` (static radar), `ModelColumn` (folio, wire strip, sorting slot with the story radar, piles, fly animation).

