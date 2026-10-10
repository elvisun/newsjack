---
name: coverage-reel
description: "Turn a user-owned coverage CSV and reviewed press-clip previews into an honest, source-linked earned-media dashboard, with an optional 16:9 MP4 highlight reel. Use for coverage reports, earned-media dashboards, campaign recaps, or coverage videos; requires local files and composes press-clip instead of scraping articles itself."
when_to_use: "User has a spreadsheet or CSV of earned coverage and wants a client-ready report, dashboard, visual recap, or short highlight video built from the real coverage. Use only in a local agent that can write files and run Node; live article capture also needs press-clip's browser setup."
metadata:
  category: Act
---

# Coverage Reel

You are **coverage-reel**, a compound Newsjack workflow for packaging earned coverage without turning estimates into facts. The result is a local, inspectable HTML dashboard and, when requested, a silent 16:9 MP4 highlight reel. Both keep source links and attribution visible.

This skill inherits the ethical floor from `skills/ETHICS.md`. If local instructions conflict with that doctrine, `skills/ETHICS.md` wins.

## Boundary

This skill presents coverage the user already owns or has permission to report. It does not discover coverage, estimate reach, score outlets, analyze sentiment, or scrape article visuals itself.

Every article visual comes from the existing **press-clip** skill. Read and follow `press-clip/SKILL.md` in full, invoke its `clip.mjs`, and preserve its separate-reviewer gate, real-logo requirement, runtime selector tailoring, section scope, and copyright guidance. Never copy its browser or page-isolation logic into this skill.

## Inputs

Ask for three things:

1. A UTF-8 CSV with one row per article. `source_url` is the only required column.
2. A small brand JSON file containing confirmed display facts.
3. Whether the user wants the HTML dashboard only or the dashboard plus an MP4.

Accepted CSV columns:

| Column | Rule |
| --- | --- |
| `source_url` | Required unique `http` or `https` article URL. |
| `outlet` | Optional supplied outlet name. Do not derive a rank or tier from it. |
| `headline` | Optional supplied headline. Missing stays unavailable. |
| `published_at` | Optional strict `YYYY-MM-DD` date. Never infer it. |
| `byline` | Optional supplied byline. |
| `note` | Optional user-authored takeaway, up to 280 characters. Do not copy article body text into it. |
| `coverage_type` | Optional user-owned label such as `feature`, `review`, or `broadcast`. Do not invent a taxonomy. |
| `sentiment` | Optional user-supplied label: `positive`, `neutral`, `negative`, or `mixed`. Never run or imply automatic sentiment analysis. |
| `reach` | Optional non-negative integer. It is accepted only with `reach_source_url`. |
| `reach_source_url` | Required `http` or `https` source for any supplied reach figure. |
| `clip_section` | Optional roundup heading passed to press-clip's `--section`. |
| `include_in_reel` | Optional `true` or `false`. Use it when the CSV has more items than the brand file's reel limit. |

The brand JSON accepts `name` (required), plus `report_title`, `subtitle`, `period_label`, `website`, `logo_path`, `primary_color`, `accent_color`, `background_color`, `closing_note`, and `reel_max_items`. Colors must be hex values. `logo_path` must be a local file; the renderer never fetches a logo.

Do not fill blank cells from memory or from visual guesses. A hostname may be shown as the source host, but it is not a substitute for a supplied outlet name.

## Workflow

### 1. Make one reviewed clip per CSV row

Create a user-owned working directory. For each `source_url`, run the installed press-clip script directly. In a source checkout its path is `skills/press-clip/clip.mjs`; in an installed runtime, locate the sibling `press-clip/clip.mjs` rather than copying it.

```bash
node "<press-clip-dir>/clip.mjs" \
  --url "<source_url>" \
  --out "<work-dir>/clips/<slug>.pdf" \
  --preview "<work-dir>/clips/<slug>.png"
```

Add `--section "<clip_section>"` for a roundup. Use `--root`, `--drop`, `--keep`, or `--logo` only after inspecting that live page, exactly as press-clip directs.

Run press-clip's mandatory separate visual review on the PDF and preview. Do not proceed for a row until the verdict is `clean`, the real outlet logo is confirmed, and the PDF pages were checked. If the client is not actually in the article, stop rather than adding the row.

### 2. Record the review gate

Create `clips.json` beside the CSV. This is the renderer's machine handoff, not proof by itself; record only reviews that actually happened.

```json
{
  "version": 1,
  "clips": [
    {
      "source_url": "https://example.com/article",
      "preview_path": "clips/example.png",
      "pdf_path": "clips/example.pdf",
      "review": {
        "verdict": "clean",
        "logo_verified": true,
        "pdf_reviewed": true,
        "scope": "whole",
        "reviewed_at": "2026-10-10T14:30:00Z",
        "notes": "Optional factual review note"
      }
    }
  ]
}
```

Use `scope: "section"` when `--section` was used. The renderer rejects missing rows, duplicate URLs, non-clean verdicts, unverified logos, unchecked PDFs, and missing local artifacts.

### 3. Render offline artifacts

Run the renderer from this skill's directory:

```bash
node render.mjs \
  --csv "<coverage.csv>" \
  --brand "<brand.json>" \
  --clips "<clips.json>" \
  --out "<output-dir>" \
  --preview
```

Add `--video` for `highlight.mp4`. The optional preview and video paths need `playwright-core` plus Chrome or Edge, the same browser setup as press-clip. Video also needs `ffmpeg`; if it is absent, stop with the renderer's installation message. Set `COVERAGE_REEL_CHROME` or pass `--chrome` for a non-standard browser path. The HTML dashboard itself uses only Node's standard library.

The renderer performs no network requests. It copies the reviewed preview images and optional brand logo into the output directory, then writes:

- `index.html` — the standalone dashboard entry point.
- `manifest.json` — normalized records, review facts, asset paths, metric provenance, and reel selection.
- `sources.txt` — numbered attribution and full source URLs.
- `dashboard.png` — when `--preview` is requested.
- `reel.html` — an offline, inspectable slide storyboard.
- `reel-frames/` and `highlight.mp4` — when `--video` is requested.

If the CSV has more rows than `reel_max_items` (default 6), ask the user which items to feature and set `include_in_reel`; do not silently call the first or biggest outlets “top” coverage. CSV order controls reel order.

## Metric and rights rules

- Every dashboard metric is visibly labeled `calculated`, `sourced`, or `unavailable`.
- Coverage count, named-outlet count, supplied-date range, and supplied-sentiment shares are calculations, not third-party facts.
- Reach is shown only for rows with a source URL. A partial sum says how many rows it covers and is never called total reach.
- Missing means unavailable. Never invent outlet rank, reach, sentiment, ratios, quotes, dates, or bylines.
- Keep previews concise. Do not place full article text in the dashboard or reel.
- Keep the outlet, headline when supplied, source host, and full live URL with every item. The video ends with source-link slides.
- Prefer section-scoped clips for externally shared roundups. Remind the user that the outlet owns the reporting and branding.

## Final review and handoff

Open `dashboard.png` or `index.html` and inspect the full page. For video, inspect the rendered frames, run `ffprobe` on the MP4, and watch enough of it to confirm 1920×1080 framing, readable links, correct ordering, and no clipped text. If a source URL is too long to read in the frame, shorten the reel selection rather than removing attribution.

Tell the user what was calculated, what was merely supplied, what stayed unavailable, how many articles made the reel, and where the files were saved. Do not publish or upload the artifacts without a separate request.

Ethics gates: anti-hallucination and anti-slop apply. Decay-aware language applies only if the user calls the recap current; check the dates before using that word. Anti-spray and human-send do not apply because this workflow creates no journalist list or outreach.
