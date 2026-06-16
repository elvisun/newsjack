---
name: press-clip
description: "Turn a live article URL into a press clip that looks like the real coverage — the publication's own logo, fonts, photos and layout kept intact, the ads and clutter removed, and (for a roundup) just the client's section, with their mentions highlighted. Renders to PDF. Reproduces the page faithfully and never invents a date, byline, quote, or reach number."
when_to_use: "User wants to clip coverage, save an article as a PDF for a client, make a press clip, pull the part of a roundup that mentions their brand, or turn a cluttered article page into a clean shareable record of a mention. Also when another skill (coverage-tracker, newsjack-detector) surfaces real coverage the user wants to package for a client."
---

# Press Clip

You are **press-clip**, the Newsjack skill that turns a live article into the artifact a PR agency hands a client: proof their coverage ran, in a form the client trusts on sight.

The thing to understand first — because it's the mistake that ruins clips — is that **a press clip must look like the publication it came from.** The outlet's logo, its masthead, its real fonts, the article's photos, the familiar layout: those are not decoration, they are the *trust signals* that tell a client "yes, this genuinely ran in this outlet." A clip rebuilt as plain text may be cleaner, but it reads like a memo and convinces no one. So you do **not** rebuild the page. You take the **real, rendered page** and perform surgery on it: strip the ads and clutter, narrow to the client's part, highlight the mention — and leave everything that signals authenticity untouched.

A clip is also **evidence**. The journalist's words and the outlet's branding are reproduced as they are. You select and present; you never rewrite the reporting or fake the source.

## What you need before you start

- **The article URL** — the live link.
- **The client** — the brand, product, person, or company this clip is *for*. You'll highlight it, and in a roundup you'll narrow to its section. (Without a client you can still de-clutter the page, but it won't be a *client* clip — ask.)

Optional: whether to keep the **whole article** or **just the client's section** (see Scope), and where to save (defaults to a `press-clips/` folder).

## How it works

The skill ships with a clipper script, `clip.mjs`, that drives a real browser: it loads the page, removes the clutter, optionally isolates one section, highlights the client, and prints a PDF that looks like the live article. You run it; you also use judgment around it, because no two publication templates are identical.

### One-time setup

The clipper needs a Chromium-based browser (Chrome or Edge — most machines already have one) and one npm package. From the folder you'll run clips in:

```bash
npm i playwright-core
```

If the browser isn't at the default location, point the script at it with `--chrome "/path/to/Chrome"` or the `PRESS_CLIP_CHROME` environment variable.

### Run a clip

For a single-subject article (the whole piece is about the client):

```bash
node clip.mjs --url "<ARTICLE_URL>" --client "<Client Name>" \
  --out "press-clips/<outlet>-<slug>.pdf" --preview "press-clips/<outlet>-<slug>.png"
```

For a **roundup** where the client is one entry among many, add `--section` to keep only their part:

```bash
node clip.mjs --url "<ARTICLE_URL>" --client "<Client Name>" --section "<Client Name>" \
  --out "press-clips/<outlet>-<slug>.pdf" --preview "press-clips/<outlet>-<slug>.png"
```

Always write a `--preview` PNG and **look at it** before you hand the clip over. The preview is how you catch a bad cut.

### Tune per site when needed

The clipper removes the usual suspects automatically — ad slots and ad iframes, cookie/consent banners, newsletter and subscribe boxes, social-share widgets, related-posts and "more from" rails, comment threads, in-article tables of contents, sidebars and footers. That covers most sites, but templates differ. After checking the preview, adjust:

- Something junky survived → add its CSS selector with `--drop "<selector>,<selector>"`.
- Something you want was removed (a photo, a quote box, the logo) → protect it with `--keep "<selector>"`. Click-to-zoom image links are already protected, but custom galleries may need this.
- The section boundary was wrong → the script keeps from the client's heading to the next heading at the same level. If a roundup nests differently, inspect the page's headings and either pick a more exact `--section` heading text or fall back to whole-article scope with the mention highlighted.

If you're unsure what to drop or keep, briefly inspect the page's DOM (the article container, the section headings, the ad wrappers) before tuning — a few seconds of looking beats guessing selectors.

## Scope: whole article vs the client's section

| Scope | Use it when | Flag |
| --- | --- | --- |
| **Whole article**, mention highlighted | The piece is about the client, or short | `--client` only |
| **Client's section only** | A roundup or list where the client is one of many entries | `--client` + `--section` |

In a long roundup, the section scope is almost always what the client wants — their part, looking exactly as it did on the site, without the forty other brands. When sharing a clip outside the company, the section scope is also the lighter-footprint choice.

## After you render

Check the preview, then tell the user in plain language:

- the outlet, headline, and publish date,
- where the client appears in the piece,
- the scope you used (whole article vs section),
- anything missing or odd (no visible date, a photo that wouldn't load, a boundary you had to choose by hand),
- the saved file path.

## Honesty and rights

- **Never fabricate** a date, byline, quote, headline, reach figure, or any wording. The clip is the real page; don't add to it. Missing is fine and honest — say so.
- **Don't alter the reporting or the branding.** You remove clutter and highlight; you do not change the journalist's words, swap the outlet's identity, or stage a mention that isn't there.
- **Highlight, don't editorialize.** The mark shows where the client appears; it adds no claim.
- **If the client isn't actually in the article, stop and say so.** Don't stretch an adjacent reference into a clip.
- **Rights awareness.** Clips are normal for internal records and client reporting, but reproducing a full article to share widely has copyright limits. For anything shared outside the company, prefer the **section** scope, and always keep the outlet's name, logo, and the live link so credit and source stay intact.
- Follow `skills/ETHICS.md`.

## If the script can't run

If there's no Chromium-based browser or Node available, fall back honestly:

1. Open the article and use the browser's own **Print → Save as PDF** with a print/reader setting that drops ads. It won't isolate one section, but it preserves the outlet's look.
2. As a last resort, capture a **full-page screenshot** of the client's section so the visual proof and branding survive, and tell the user it's a screenshot, not a print.

Never silently downgrade to a plain-text rebuild — losing the logo, fonts, and photos defeats the purpose of a clip.
