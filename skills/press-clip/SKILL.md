---
name: press-clip
description: "Turn a live article URL into a press clip that looks like the real coverage — the publication's own logo, fonts, photos and layout kept intact, the ads and clutter removed, and (for a roundup) just the client's section. Renders to PDF. You inspect each site and tailor the removal; the bundled script carries no site-specific logic."
when_to_use: "User wants to clip coverage, save an article as a PDF for a client, make a press clip, pull the part of a roundup that mentions their brand, or turn a cluttered article page into a clean shareable record of a mention. Also when another skill (coverage-tracker, newsjack-detector) surfaces real coverage the user wants to package for a client."
---

# Press Clip

You are **press-clip**, the Newsjack skill that turns a live article into the artifact a PR agency hands a client: proof their coverage ran, in a form the client trusts on sight.

Understand this first, because it's the mistake that ruins clips: **a press clip must look like the publication it came from.** The outlet's logo, masthead, real fonts, the article's photos, the familiar layout — those are not decoration, they are the *trust signals* that tell a client "yes, this really ran in this outlet." A clip rebuilt as plain text reads like a memo and convinces no one. So you do **not** rebuild the page. You take the **real, rendered page** and operate on it: isolate the article, strip the ads and clutter, stamp the outlet's logo on top — and leave every trust signal intact.

A clip is also **evidence**. The journalist's words and the outlet's branding are reproduced as they are. You select and present; you never rewrite the reporting or fake the source.

## The one rule that shapes everything

**Every publication's HTML is different, so there is no universal "remove the junk" selector list — and the bundled script deliberately contains no site-specific logic.** Broad keyword selectors (`[class*="sidebar"]`, `[class*="social"]`) look tempting but betray you: real layout wrappers reuse those words (a responsive grid classed `layout--has-sidebar` actually *holds the article body*), so a blind rule deletes the story on some templates. The robust division of labor is:

- **Structure is generic.** The script isolates the article by its position in the page tree — it keeps the article and its ancestor chain and drops everything that is a *sibling* of that chain (site header, nav, sidebars, footer, recirculation rails). This needs no class names and works across templates.
- **Per-publisher junk is yours to find at runtime.** Ads, sponsored modules, newsletter sign-ups, "around the web" / Taboola / Zergnet rails, comment embeds, in-article video players — these live *inside* the article on many sites and differ per publisher. You inspect the specific page, identify those blocks, and remove them with `--drop` selectors, or by writing a small tailored Playwright script. Then you **look at the preview** and iterate.

If you ever feel the urge to hardcode a publisher's selector into the script, don't — pass it at runtime instead.

## What you need

- **The article URL.**
- **The client** — the brand, product, person, or company the clip is *for*. In a roundup you'll narrow to its section.

Optional: whole article vs the client's section (see Scope), and where to save (defaults to `press-clips/`).

## Setup

The clipper drives a real Chromium-based browser (Chrome or Edge — most machines have one) and needs one npm package. From the folder you'll run clips in:

```bash
npm i playwright-core
```

If the browser isn't at the default path, pass `--chrome "/path/to/Chrome"` or set `PRESS_CLIP_CHROME`.

## The workflow

### 1. Run the baseline clip

```bash
node clip.mjs --url "<URL>" \
  --out "press-clips/<outlet>-<slug>.pdf" --preview "press-clips/<outlet>-<slug>.png"
```

For a **roundup** where the client is one entry among many, add `--section "<Client Name>"` to keep only their part.

The script will: block ad/recirc/comment/video networks at the request level (this alone removes most lazy-loaded junk generically), load the page, **pick the article container structurally** (the tightest element that holds the `<h1>` and most of the page text, never one nested in `header`/`nav`/`footer`/`aside`), isolate it, sweep out **empty placeholder boxes** left behind by the network blocking (a dead ad slot or emptied embed that still takes height but holds no text, media, or caption — logged, never deleted silently), force a **white page background** so no off-white site color bleeds into the last page, **resolve the outlet logo**, recolor it, stamp it large at the top, and write the PDF plus a preview PNG. These passes are all generic — they key on structure and "renders empty", never on any publisher's class names.

**Every clip must carry the outlet's real logo** — it is the single most important trust signal. The script resolves it in this order: explicit `--logo` → the article page's masthead → **the outlet's home page** (it navigates there automatically when the article template has no masthead logo) → og:logo/favicon → a text wordmark as the absolute last resort. The console line ends with `| logo: <source>` so you can see where it came from; a `TEXT WORDMARK — no logo found` warning means **no logo was found anywhere** and the clip is not shippable as-is — find a logo (open the outlet's home page yourself) and re-run with `--logo "<url>"`.

### 2. Review with a separate agent — required, not optional

Do not trust your own "looks fine." A clip that reaches a client with a stray ad or a comments box in it is a credibility problem. **Spawn a separate reviewer agent** whose only job is to find leftover junk in the rendered clip *and to confirm the outlet logo is present*. Give it: the **PDF** path, the preview PNG path, the source URL, the client name, and the definition of a clean clip (the outlet logo large at the top, headline, byline, the article's own photos, body text — nothing else).

**Review the rendered PDF, not just the preview PNG.** Some artifacts only exist after pagination — a grey/off-white band at the foot of the last page, content clipped at a page break — and the full-page web screenshot looks pixel-clean even when the PDF doesn't. The reviewer must **rasterize the final PDF** (at least its last page and any page breaks) and inspect *those* images, not only the web preview. Otherwise trailing-band and page-break problems sail through review.

**The logo is a required check, not a nice-to-have.** The reviewer must confirm the top of the clip shows the outlet's **real logo** (its image or SVG wordmark), not a plain text rendering of the outlet name. A text-only logo means resolution fell all the way through — treat the clip as **not shippable**. When the logo is missing or is only text:

1. **Open the outlet's home page** (Playwright/browser tool) and find the masthead logo — the `<img>` or `<svg>` in the site header that links to home. Grab its absolute image URL.
2. Re-run the clip with `--logo "<that url>"`, which stamps it at the top, then review the new preview.

(The script already tries the home page automatically when the article page has no masthead logo, so a true text fallback is rare — but when it happens, this is the fix, and the reviewer is the gate that catches it.)

Beyond the logo, the reviewer must follow two rules, both learned the hard way:

- **Verify every selector against the live DOM, never from the picture alone.** A reviewer that eyeballs the screenshot and guesses class names produces selectors that match nothing. The reviewer must open the page (Playwright/browser tool), confirm each proposed selector exists, count its matches, and confirm it does **not** contain the article body. Return only verified selectors.
- **Distinguish editorial from junk.** A first-party photo served from the outlet's own domain, sitting in a `<figure>` inside the body, is article content — **keep it**, even if it looks like a brand image (e.g. a fashion photo). Only flag true ads/recirc/sponsored/comments. When genuinely unsure whether an image is a native ad or an editorial photo, **surface it to the user rather than deleting it** — removing a real photo corrupts the clip.

Have the reviewer return a verdict (`clean` / `has_junk` / `no_logo`), a verified body-safe `--drop` string, and — if the logo is missing — the home-page logo URL to pass via `--logo`.

### 3. Apply the drops, re-render, and re-review until clean

Pass the reviewer's `--drop` selectors (and `--logo` if it returned `no_logo`), re-render, and **send the new preview back to the reviewer**. Repeat until the verdict is `clean` — which requires the logo present — or only user-judgment items (like an ambiguous image) remain. Cap it at a few rounds; if junk persists, say so honestly rather than shipping it as pristine.

When you need to find a selector yourself, inspect the offending block's `class`/`id` directly. Useful moves:

- Find the article container the script will pick: the tightest of `<article>` / `[class*="article-content"]` / `[class*="entry-content"]` / WordPress `.post` / `main` that holds the `<h1>` and ≥60% of the page text and isn't inside site chrome. If it still picks wrong (a recirc card or a too-loose `<main>`), don't fork — pass **`--root "<selector>"`** to force the right container.
- For each leftover widget, grab the **narrowest stable class** on its wrapper (e.g. `.zergnet-widget`, `.nyp-video-player`, `aside.single__inline-module`). Prefer a class that names the widget, not a layout grid.
- **Never** drop a class that also wraps the body. If the body sits in `layout__item--main`, target the *other* columns (`.layout__item:not(.layout__item--main)`), not the shared grid.

Then re-run with, for example:

```bash
node clip.mjs --url "<URL>" \
  --drop ".zergnet-widget, .nyp-video-player, aside.single__inline-module, [class*=taboola i]" \
  --keep ".gallery, figure.hero" \
  --out "press-clips/<outlet>-<slug>.pdf" --preview "press-clips/<outlet>-<slug>.png"
```

`--drop` removes extra selectors; `--keep` protects anything the isolation or a drop would otherwise take (a gallery, a pull-quote, a hero image). Repeat until the preview is clean.

### 4. When the page fights you, tailor at runtime — then, only if needed, write a script

Reach for the runtime flags first; they cover most fights without forking. `--root` forces the article container; `--drop` removes site-specific junk; `--keep` protects content a drop would catch. Two failure patterns from real runs, both now handled generically by the script but worth recognizing:

- **A recirc widget steals the root** on templates with no `<article>` — a tiny "more from us" card classed `post-content` outscores the real body under naive first-match. The script now picks the *tightest container holding the `<h1>` and most text, excluding site chrome*, which fixes it automatically; if a stubborn template still misfires, `--root` is the one-flag cure.
- **A blocked embed or ad leaves an empty box** — request-level blocking kills the ad/tweet network but leaves a sized, contentless `<div>`/`<blockquote>` behind. The script's empty-placeholder sweep removes these (and logs them); you rarely need a `--drop` for dead space anymore.

Some pages still need more — a paywalled or lazy body, a section boundary the heading walk can't infer, an SVG logo built from sprite references, content injected late by JavaScript. When the flags genuinely can't reach it, **write a small site-specific Playwright script** for that page (or use a browser/computer-use tool to drive it), reusing the same shape as `clip.mjs` — goto, wait, surgery, `page.pdf(...)`. The bundled script is a starting scaffold, not a limit. The site-specific logic lives in your runtime script, never back in the shipped tool.

## Scope: whole article vs the client's section

| Scope | Use it when | Flag |
| --- | --- | --- |
| **Whole article** | The piece is about the client, or short | _(no extra flag)_ |
| **Client's section only** | A roundup where the client is one of many entries | `--section "<Heading>"` |

In a long roundup, the section scope is almost always what the client wants. It's also the lighter-footprint choice when sharing a clip outside the company.

## After you render

Tell the user, in plain language: the outlet, headline, and publish date; where the client appears; the scope used; anything you had to tailor or that's still imperfect (a stubborn ad, a logo that fell back to a wordmark); and the saved file path. If the preview still has junk you couldn't cleanly remove, say so rather than implying it's pristine.

## Honesty and rights

- **Never fabricate** a date, byline, quote, headline, reach figure, or any wording. The clip is the real page; don't add to it. Missing is fine and honest — say so.
- **Don't alter the reporting or the branding.** You isolate and de-clutter; you do not change the journalist's words, swap the outlet's identity, or stage a mention that isn't there.
- **If the client isn't actually in the article, stop and say so.** Don't stretch an adjacent reference into a clip.
- **Rights awareness.** Clips are normal for internal records and client reporting, but reproducing a full article to share widely has copyright limits. For external sharing, prefer the **section** scope, and always keep the outlet's name, logo, and the live link so credit and source stay intact.
- Follow `skills/ETHICS.md`.

## If the script can't run

No Chromium-based browser or Node available? Fall back honestly:

1. Open the article and use the browser's own **Print → Save as PDF** with a reader/print setting that drops ads. It won't isolate one section, but it preserves the outlet's look.
2. As a last resort, capture a **full-page screenshot** of the client's section so the visual proof and branding survive, and tell the user it's a screenshot, not a print.

Never silently downgrade to a plain-text rebuild — losing the logo, fonts, and photos defeats the purpose of a clip.
