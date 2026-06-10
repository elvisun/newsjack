---
name: eval-graphics
scope: eval-internal
not_a_product_skill: true
description: "Turn an eval study's numbers into on-brand, publish-ready figures using the Newsjack chart room (the eval design system), then validate them with Playwright. For producing the charts in a published eval/data study."
when_to_use: "You have an eval study's results (a results.json, an aggregate table, head-to-head numbers, per-dimension deltas) and need branded figures to publish it — a flagship comparison, a win-rate donut, headline stat callouts, a heatmap, a trend line."
---

# Eval Graphics — the chart room

> **Internal tooling, not a product skill.** This lives in `eval/design-system/`
> and is used by maintainers to publish eval studies. It is **not** a Newsjack
> user skill, must **never** be installed into `skills/`, and is never loaded at
> product runtime. Do not confuse it with the main skills folder.

You produce **figures** for an eval study: standalone HTML that renders the
Newsjack house chart style, validated by Playwright and screenshotted to PNGs you
drop into a writeup. One grammar — newsprint paper, ink, a single vermilion mark
— across every figure, so a study reads like one publication.

## Files you work with

- `assets/colors_and_type.css` — design tokens (palette, type, spacing). Never
  edit; always link.
- `charts.css` — the chart primitives (`.bar-primary`, `.line-base`, `.fig`,
  `.stat`, heatmap classes, masthead/section/colophon scaffold). Never edit;
  always link.
- `chart-room.html` — the **specimen gallery** of all 9 figure types with
  placeholder data. This is your copy-paste source: find the figure that fits
  your data, lift its block, swap the geometry and labels.
- `scripts/validate.mjs` — Playwright validator + screenshotter.

## The grammar (non-negotiable brand rules)

These come straight from the design system. Breaking one means the figure is
off-brand:

- **One chroma.** Vermilion `#E05A47` is the only colour. The highlighted /
  winning series is accent; every other series is quiet grey (`--c-base`,
  `#E4E0D8`) or an ink wash. **If you reach for a second colour, you've gone
  wrong.** No gradients. No emoji. Ever.
- **Paper, not white.** Background is `--nj-page` `#F9F8F6`. Borders are the
  hairline `rgba(26,26,26,0.10)`. Cards are 0-radius with a faint editorial
  shadow.
- **Type roles.** Figure titles: Newsreader **italic**. Axis ticks, value
  labels, legends, eyebrows: IBM Plex Mono, ALL CAPS, ≥0.12em tracking.
  Descriptive captions: DM Sans.
- **Value labels float above the bar** (mono, centered over the column).
- **Headline figures are one consistent colour** — the whole numeral *and* its
  symbol in the accent (e.g. `+38%`, `4.2×` fully vermilion), via
  `<span class="accent">`.
- **Section pattern:** top hairline → mono number + lowercase-italic title (left)
  + mono subtitle (right) → content.

### Accent strategy — set on `<body>`

| `data-accent` | Use |
|---|---|
| `winner` (default) | best/highlighted series = accent, others grey. The standard comparison look. |
| `single` | comparisons go all-grey; accent is reserved for one hero mark (a donut, one bar). |
| `mono` | everything ink, no chroma — for a sober, neutral data study. |

Also `data-grid="on|off"` (gridlines) and `data-barstyle="solid|outline"`.

## Pick the figure to fit the data

| Data shape | Figure (block in `chart-room.html`) |
|---|---|
| Two series across a few benchmarks | **FIG.01 grouped bars** — the flagship |
| One series, ranked by category | **FIG.02 single-series bars** |
| Before → after on sparse metrics | **FIG.03 dumbbell / lollipop** |
| A value over time / scale / versions | **FIG.04 line / scaling curve** |
| Two-axis tradeoff, one point highlighted | **FIG.05 scatter / quadrant** |
| Composition / share across rows | **FIG.06 100% stacked bars** |
| One number that deserves the frame | **FIG.07 donut** |
| Model × task (or any) matrix | **FIG.08 heatmap** (data-driven JS) |
| 2–4 punchy single numbers, no axes | **FIG.09 big-stat callouts** |

## How to compute geometry (filling an SVG template)

The SVGs use plain coordinates inside a `viewBox`. The mapping math you need:

**Vertical bars (FIG.01/02).** Choose a baseline `yBase` (value 0) and a top
`yTop` (max value). `scale = (yBase - yTop) / maxVal`. For a value `v`:
`barHeight = v * scale`, `barY = yBase - barHeight`, value label at
`y = barY - 12`. (Specimen FIG.01: `yBase=500`, `100→y=120`, so `scale=3.8`.)

**Dumbbell (FIG.03).** Horizontal axis from `xMin` (value 0) to `xMax` (value
100). `x(v) = xMin + (v/ (maxVal)) * (xMax - xMin)`. Draw a `.stem` line from
`x(before)` to `x(after)`, a `.dot-base` at before, a `.dot-primary` at after.

**Line (FIG.04).** `y(v) = yBase - v*scale`; evenly space x across the points;
`.line-primary` for focal, `.line-base` (dashed) for baseline; optional
`.area-primary` polygon closes down to `yBase`.

**Donut (FIG.07).** `C = 2 * π * r` (specimen `r=104` → `C≈653.45`). For percent
`p`: `dash = (p/100) * C`; set `stroke-dasharray="{dash} {C}"` on the accent ring,
`transform="rotate(-90 cx cy)"` so it starts at 12 o'clock.

**Heatmap (FIG.08).** Don't hand-place cells — edit the `data`, `tasks`, `models`
arrays in the inline `<script>`. Cell alpha = `0.12 + clamp((v-min)/(max-min)) *
0.88`; text flips to white above alpha 0.55. Already handles any grid size.

**Big-stat (FIG.09).** Pure markup: `<span class="fig-num"><span class="accent">
+38</span>%</span>` — wrap the whole figure (or the part that should be coloured)
in `.accent`. Keep the numeral and its symbol the same colour.

## Build process

1. **Read the study's numbers** (e.g. an `aggregate.py` printout or
   `results.json`). Decide which 1–4 figures tell the story; don't over-chart.
2. **Scaffold an HTML file** in the study's run folder, e.g.
   `eval/<study>/runs/<run>/figures/<name>.html`. Link the design system by
   relative path:
   `<link rel="stylesheet" href="../../../../design-system/assets/colors_and_type.css">`
   and `.../design-system/charts.css`. (Count the `../` from the figure file to
   `eval/design-system/`.) Set `<body data-accent="…">`.
3. **Lift the matching figure block(s)** from `chart-room.html`. Recompute every
   coordinate from the real data using the math above. Replace category labels,
   value labels, series names, captions, the masthead/section text. Delete the
   figures you don't use.
4. **Keep it honest.** Label sample sizes (`n=`), say what the score is, don't
   round a tie into a win, and don't invent a series the data doesn't have. The
   `angle-generator`/`meanest-editor` anti-slop ethos applies to charts too.
5. **Validate with Playwright** (required):
   ```bash
   cd eval/design-system/scripts
   node validate.mjs ../../<study>/runs/<run>/figures/<name>.html --out ../../<study>/runs/<run>/figures/png
   ```
   All checks must pass (no JS errors, accent painted, serif applied, every
   figure has geometry, no empty SVGs). It writes a full-page PNG plus one crop
   per figure — those are your embeddable assets.
6. **Eyeball the screenshot** before publishing. Read the full-page PNG; confirm
   bars/labels line up and nothing overflows.

## Worked example

For the Fable-5-vs-Opus-4.8 study (`eval/fable-vs-opus/`), the headline numbers
(overall dim mean 4.60 vs 4.36; 67/33 head-to-head; 24 vs 7 robust wins; per-
dimension deltas) map cleanly to: a **FIG.01 grouped-bars** of the 7 dimensions
(Fable accent, Opus grey), a **FIG.07 donut** for the win rate, and a row of
**FIG.09 big-stat callouts** (overall mean, robust wins, `publishable` rate). See
`eval/fable-vs-opus/runs/2026-06-09-full/figures/` for the built, validated set.

## Guardrails

- Never edit `assets/colors_and_type.css` or `charts.css` to force a one-off look
  — if a figure needs something the system lacks, that's a design-system change,
  raise it, don't hack the study.
- Never add a second colour, a gradient, a drop shadow on a chart, or emoji.
- Never ship a figure that hasn't passed `validate.mjs`.
- Placeholder data stays clearly labelled until real numbers replace it.
