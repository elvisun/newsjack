# Eval design system — "the chart room"

The house visual system for **publishing Newsjack eval studies**: a set of
reusable, on-brand figure types plus the tooling to fill them with real study
data and validate them. Editorial-meets-terminal — newsprint paper, ink, a
single vermilion mark — so every study reads like one publication.

> **This is eval-internal tooling, NOT a product skill.** It lives under `eval/`
> on purpose. `SKILL.md` here is for maintainers publishing studies; it must
> never be copied into the top-level `skills/` folder or loaded at product
> runtime. Do not confuse `eval/design-system/SKILL.md` with the main skills.

Implemented from the Claude Design handoff *"Newsjack Eval Graphs.html"*
(claude.ai/design), which itself derives from the `newsjack.sh` marketing-site
design system.

## What's here

```
eval/design-system/
├── README.md                  ← this file
├── SKILL.md                   ← eval-graphics skill: how to build a study's figures
├── assets/colors_and_type.css ← design tokens (palette, type, spacing, motion)
├── charts.css                 ← chart primitives + page scaffold
├── chart-room.html            ← specimen gallery: all 9 figure types (copy-paste source)
├── scripts/
│   ├── validate.mjs           ← Playwright validator + screenshotter
│   └── package.json           ← playwright dependency
└── screenshots/               ← generated PNGs (gitignored)
```

## The figure library (9 types)

`chart-room.html` is a single specimen page showing every figure with
placeholder data:

1. **Grouped bars** — the flagship two-series comparison (the "Gemini layout").
2. **Single-series bars** — one series ranked by category.
3. **Dumbbell / lollipop** — before→after on sparse metrics.
4. **Line / scaling curve** — a value over time, scale, or releases.
5. **Scatter / quadrant** — two-axis tradeoff, one point highlighted.
6. **100% stacked bars** — composition / share across rows.
7. **Donut** — one headline percentage.
8. **Heatmap** — a model × task (or any) matrix, data-driven.
9. **Big-stat callouts** — 2–4 punchy single figures, no axes.

Variants are driven entirely by CSS via `<body>` attributes — no framework:
`data-accent="winner|single|mono"`, `data-grid="on|off"`,
`data-barstyle="solid|outline"`.

## Quick start

```bash
# one-time: install Playwright + Chromium
cd eval/design-system/scripts
npm install && npx playwright install chromium

# validate the specimen gallery and write screenshots
node validate.mjs                      # -> ../screenshots/chart-room--*.png
```

`validate.mjs` checks the system actually rendered (no JS errors, accent token
painted, Newsreader applied, every figure has real geometry, no empty SVGs) and
writes a full-page PNG plus one crop per figure.

## Building a study's figures

See **`SKILL.md`** for the full process: pick the figures that fit your data,
copy their blocks out of `chart-room.html`, recompute the SVG geometry from the
real numbers (the geometry math is documented), link the tokens by relative
path, then validate with `validate.mjs`. Built figures live alongside the study,
e.g. `eval/fable-vs-opus/runs/<run>/figures/`.

## Brand rules (the short version)

One chroma (vermilion `#E05A47`), paper not white, hairline borders, 0-radius
cards, Newsreader-italic titles, mono-caps labels, value labels above bars,
single-colour headline figures. No second colour, no gradient, no emoji. The
full system is documented in `assets/colors_and_type.css` and the design-system
README that shipped with the handoff.

## Provenance

Design source: Claude Design handoff bundle `newsjack-evaluation-graphs`
(`Newsjack Eval Graphs.html` + `newsjack-sh-design-system`). The tokens
(`colors_and_type.css`) and chart primitives (`charts.css`) are ported verbatim;
`chart-room.html` is the faithful static implementation with the design tool's
React tweak-panel harness removed (the `<body>` data-attributes are the controls
instead).
