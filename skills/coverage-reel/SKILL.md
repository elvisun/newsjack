---
name: coverage-reel
description: "Turn a user-owned coverage CSV and reviewed press-clip captures into an honest, source-linked earned-media dashboard and a motion-designed highlight reel (MP4) that scrolls each real article to the client's sentence and highlights it. Use for coverage reports, earned-media dashboards, campaign recaps, or coverage videos; requires local files and composes press-clip instead of scraping articles itself."
when_to_use: "User has a spreadsheet or CSV of earned coverage and wants a client-ready report, dashboard, visual recap, or short highlight video built from the real coverage. Use only in a local agent that can write files and run Node; captures and video also need press-clip's browser setup (Chrome or Edge with playwright-core)."
metadata:
  category: Act
---

# Coverage Reel

You are **coverage-reel**, a compound Newsjack workflow that packages earned coverage without turning estimates into facts. It makes two things from the same reviewed evidence:

- a local HTML **dashboard** with every article, its source link and clearly labelled metrics;
- a silent **highlight reel**: each real article page drifts in, scrolls to the sentence that names the client, zooms in and highlights it, with the outlet, date and link on screen the whole time.

This skill inherits the ethical floor from `skills/ETHICS.md`. If local instructions conflict with that doctrine, `skills/ETHICS.md` wins.

## Boundary

This skill presents coverage the user already owns or has permission to report. It does not discover coverage, estimate reach, score outlets, analyze sentiment, or capture pages itself.

Every article image comes from the **press-clip** skill. Read and follow `press-clip/SKILL.md` in full. Run its `clip.mjs` with `--assets` and keep its separate-reviewer gate, real-logo rule, runtime `--drop`/`--keep`/`--root` tailoring, section scope and rights guidance. Never copy its browser or page-cleanup logic into this skill.

## Inputs

Ask for three things:

1. A UTF-8 CSV with one row per article. `source_url` is the only required column.
2. A small brand JSON file with confirmed display facts.
3. Which video shapes they want, if any (see Formats).

CSV columns:

| Column | Rule |
| --- | --- |
| `source_url` | Required unique `http` or `https` article URL. |
| `outlet` | Optional outlet name. Do not derive a rank or tier from it. |
| `headline` | Optional headline. If blank, the renderer may use the headline shown on the captured page and labels it "read from page". |
| `published_at` | Optional strict `YYYY-MM-DD` date. Never inferred or filled from the page (see Dates). |
| `byline` | Optional byline. If blank, the page's byline may be used and is labelled. |
| `note` | Optional user-written takeaway, up to 280 characters. Do not paste article text into it. |
| `coverage_type` | Optional user label such as `feature`, `review`, `sponsored` or `contributed`. Shown as written. |
| `sentiment` | Optional user label: `positive`, `neutral`, `negative` or `mixed`. Never run or imply automatic sentiment analysis. |
| `reach` | Optional whole number, accepted only with `reach_source_url`. |
| `reach_source_url` | Required `http` or `https` source for any reach figure. |
| `clip_section` | Optional roundup heading, passed to press-clip's `--section`. |
| `include_in_reel` | Optional `true` or `false`. Use it when there are more rows than the reel holds. |

The brand JSON accepts `name` (required) and `mention_terms`: the client's name plus products, spellings and spokespeople to find in each article, for example `["Acme Robotics", "Acme R2", "Acme"]`. Put the longest, most specific names first. Matching is case-sensitive, so list each spelling. Optional fields: `report_title`, `subtitle`, `period_label`, `website`, `logo_path` (a local file; the renderer never fetches a logo), `primary_color`, `accent_color`, `background_color` (six-digit hex), `closing_note` and `reel_max_items` (1 to 8, default 6).

Do not fill blank cells from memory or guesses. A hostname may be shown as the source host, but it is not an outlet name.

## Workflow

### 1. Capture each article with press-clip

Make a working folder the user owns. For each CSV row, run the installed press-clip script (in a source checkout, `skills/press-clip/clip.mjs`; when installed, the sibling `press-clip/clip.mjs`):

```bash
node "<press-clip-dir>/clip.mjs" --url "<source_url>" \
  --out "<work>/clips/<slug>.pdf" --preview "<work>/clips/<slug>.png" \
  --assets "<work>/clips/<slug>" --mention "<term 1>,<term 2>,<term 3>"
```

Use the brand's `mention_terms` for `--mention`. Add `--section "<clip_section>"` for a roundup. Add `--drop`, `--keep`, `--root` or `--logo` only after inspecting that live page, as press-clip directs. If an article spells the product differently (for example "4A" instead of "(4a)"), add that spelling with another `--mention` and tell the user.

Read the console and the `warnings` in `clip.json`:

- **Exit code 3 / blocked.** The site served a bot check instead of the article. Do not try to get around it. Ask the user for their own screenshot (PNG or JPEG) of the article. If they have none, mark the row omitted (step 3); it stays in the dashboard and sources list but not in the reel.
- **`no_mention_found`.** The client is not named in the article body. Stop and tell the user; do not clip a stretch.
- **`paywall_suspected`.** Use only what is visible to everyone, or the user's own subscriber screenshot. Never work around a paywall.
- **`layout_shifted`.** The page moved during capture. Run the clip again.

### 2. Review every capture with a separate agent (required)

Spawn a separate reviewer, as press-clip requires, and give it the PDF, the preview, the `--assets` folder (`top.png`, `strip-*.png`, `mention-1.png`, `masthead.png`, `clip.json`), the source URL and the client's terms. The reviewer must confirm:

- the clip is clean and carries the outlet's real logo (press-clip's own checks, including the rendered PDF pages);
- the captures are clean too: no cookie banners, video players, ads, deal widgets, sticky bars or repeated content anywhere in `top.png` or the strips;
- the **chosen mention** (`"chosen": true` in `clip.json`) is a sentence in the article body that really names the client, not a link, caption, related-story list or navigation;
- the headline and byline in `clip.json` match what the page shows.

Fix problems with runtime flags (`--drop` selectors verified against the live page, `--section`, `--pick <n>` to capture a different ranked mention) and capture again until the reviewer says clean.

### 3. Record the review in clips.json

Create `clips.json` beside the CSV. It is the renderer's machine handoff, not proof by itself: record only reviews that actually happened. Every CSV row needs exactly one entry.

```json
{
  "version": 2,
  "clips": [
    {
      "source_url": "https://example.com/article",
      "sidecar_path": "clips/example/clip.json",
      "pdf_path": "clips/example.pdf",
      "review": {
        "verdict": "clean",
        "logo_verified": true,
        "pdf_reviewed": true,
        "mention_verified": true,
        "captures_clean": true,
        "scope": "whole",
        "reviewed_at": "2026-10-10T14:30:00Z",
        "notes": "Optional factual review note"
      }
    },
    {
      "source_url": "https://example.com/blocked-article",
      "user_capture": { "image_path": "clips/users-own-screenshot.png" },
      "review": { "verdict": "clean", "reviewed_at": "2026-10-10T14:40:00Z", "notes": "User's own screenshot" }
    },
    {
      "source_url": "https://example.com/another",
      "omitted": { "reason": "blocked", "detail": "Bot check in the automated browser" }
    }
  ]
}
```

Use `"scope": "section"` when `--section` was used. `omitted.reason` is one of `blocked`, `paywall`, `capture_failed`, `no_mention` or `user_excluded`. The renderer rejects missing rows, duplicate URLs, any unchecked review flag, a scope that does not match the capture, captures that do not match their sidecar, and sidecars with unresolved capture warnings. A user's own screenshot appears as a still card with no highlight.

### 4. Render

Run the renderer from this skill's folder:

```bash
node render.mjs --csv "<coverage.csv>" --brand "<brand.json>" --clips "<clips.json>" \
  --out "<output-dir>" --format landscape --video
```

It writes, in the output folder:

- `index.html`: the dashboard. `manifest.json`: every value with its source, the review record, the reel selection and timing. `sources.txt`: numbered sources with full links.
- `reel.html`: the reel as an offline player. Open it in a browser to watch it in real time and scrub through it.
- With `--preview` or `--video`: `dashboard.png` and `storyboard.png` (a contact sheet of key moments).
- With `--video`: `highlight.mp4`.

The renderer makes no network requests. Previews and video use Chrome or Edge through `playwright-core`; the video is encoded inside Chrome, so no other tool is needed. If the browser cannot encode H.264 video (some open-source Chromium builds), it stops and says so: install Google Chrome and pass `--chrome "<path>"` or set `COVERAGE_REEL_CHROME`. Use `--frames-at "3.2,8.5"` to save full-size frames for a close look, and `--fps 60` only when asked.

Before exporting, the renderer checks the motion against its own rules and refuses to export if any fail: the camera never freezes, article shots never run the same length, only the named easing curves are used, nothing overshoots, at most two transition types, no capture is enlarged past its real pixels, every piece of text stays up long enough to read (at most 20 characters per second), key text stays inside the format's safe area, article shots stay under about 6.5 seconds, and the whole reel stays within 40 seconds for up to six articles. If the length check fails, feature fewer articles or capture a shorter mention sentence with press-clip's `--pick`.

If the CSV has more rows than `reel_max_items`, ask the user which to feature and set `include_in_reel`. Do not call the first or biggest outlets "top" coverage. CSV order is reel order.

### 5. Review the reel yourself

Open `storyboard.png`, then pull full-size frames at a few moments of each article (`--frames-at`). Check: readable text, no clipped words, no soft or stretched captures, the highlight sits on the client's sentence, the right order, and outlet, date and link visible in every article shot. The renderer prints its own MP4 check (size, frame rate, frame count, length); make sure it matches the format asked for.

## Formats

| `--format` | Size | Best for |
| --- | --- | --- |
| `landscape` (default) | 1920×1080 | Decks, YouTube, LinkedIn desktop, websites |
| `portrait` | 1080×1350 (4:5) | LinkedIn and Instagram feeds |
| `vertical` | 1080×1920 (9:16) | Reels, Shorts, Stories; key text avoids the top 14% and bottom 35% where apps put their buttons |
| `square` | 1080×1080 | X and older placements |

Run once per format, each into its own output folder.

## What the reel shows

An opening card (the report title, revealed word by word, over the real article tops, with the calculated facts line: pieces of coverage, outlets, supplied date span), then the outlets' mastheads exactly as captured. Then one shot per article: the page enters, drifts on the real headline, scrolls to the client's sentence, zooms in while the rest of the page softens, and a highlighter sweeps the key part of the sentence. A small "Highlights added" label shows whenever a highlight is on screen. Then a wall of the coverage with counters for calculated figures (and reach only with a source), and an end card listing every source with its date and link, plus a plain line that the coverage belongs to the publishers and their mastheads are shown to credit them, not as endorsement. The reel is silent; do not add music unless the user supplies a track they have licensed.

## Dates and other details from the page

CSV values always win. A blank headline, outlet or byline may be filled from the captured page; the dashboard and manifest label it "read from page". A blank **publish date is never filled**: pages often carry a first-published date years older than the version you captured. The renderer prints the date the page gives; show it to the user, ask them to confirm, and add it to the CSV yourself only once they do. Until then the reel shows "date not supplied".

## Metric rules

- Every metric is labelled `calculated`, `sourced` or `unavailable`.
- Coverage count, outlet count, supplied date range and shares of user-supplied sentiment are calculations, not third-party facts.
- Reach appears only for rows with a source link. A partial sum says how many rows it covers and is never called total reach.
- Missing means unavailable. Never invent reach, rank, sentiment, ratios, quotes, dates or bylines.

## Rights, briefly

This is practical guidance, not legal advice.

- Internal and client reporting is the normal use. Keep the outlet, date, link and byline on screen; the reel does this by default.
- The highlight is emphasis drawn over the real page. Never retype, shorten or "clean up" the article's words, and never stage a mention that is not there.
- Do not label sponsored, contributed or press-release pieces as earned "featured in" coverage. Set `coverage_type` so the label shows.
- For public posts or paid ads, remind the user that publishers sell licences for headlines, logos and reprints (often through agents such as PARS or Wright's Media; look for the outlet's "Reprints & Permissions" page). In the UK, Ireland and Australia, sharing article text or screenshots publicly usually needs a collective licence (NLA media access, Newspaper Licensing Ireland, Copyright Agency) or the publisher's permission.
- Photos inside captures belong to their photographers and are the most actively enforced part. For a public post, the section scope or the user's own judgement on photos is safer.
- Comply promptly if a publisher asks for a clip to come down: re-render without that row.

## Handoff

Tell the user, in plain language: which articles made the reel and which were left out and why; what was calculated, supplied, read from the page or unavailable; any publish dates waiting for their confirmation; which runtime flags you used on which sites; and where the files are. Do not publish or upload anything without a separate request.

Ethics gates: anti-hallucination and anti-slop apply. Decay-aware language applies only if the user calls the recap current; check the dates before using that word. Anti-spray and human-send do not apply because this workflow creates no journalist list or outreach.
