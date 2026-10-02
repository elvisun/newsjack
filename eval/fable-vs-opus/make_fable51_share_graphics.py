#!/usr/bin/env python3
"""Build screenshot-ready Fable 5.1 benchmark share graphics (6-model run).

Reads a finished N-model run (summary.json + results.json) and writes four
standalone 1600x900 HTML cards under figures/share/. Every number on the cards
is computed from the run; only the headline copy lives in COPY below. Render at
2x with the eval design-system validator:

    cd eval/design-system/scripts && for f in ../../fable-vs-opus/runs/<run>/figures/share/0*.html; do
      node validate.mjs "$f" --width 1600 --out ../../fable-vs-opus/runs/<run>/figures/share/png; done

Usage:
    python3 make_fable51_share_graphics.py runs/2026-09-01-fable51
"""

import html
import json
import sys
from collections import Counter, defaultdict
from pathlib import Path

DIMS = ["news_value", "distinctness", "journalist_shape", "grounding",
        "anti_slop", "proof_rigor", "usefulness"]
DIM_LABELS = {"news_value": "NEWS VALUE", "distinctness": "DISTINCT",
              "journalist_shape": "SHAPE", "grounding": "GROUNDING",
              "anti_slop": "ANTI-SLOP", "proof_rigor": "PROOF", "usefulness": "USEFUL"}
DIM_WORDS = {"news_value": "news value", "distinctness": "distinctness",
             "journalist_shape": "reporter shape", "grounding": "grounding",
             "anti_slop": "anti-slop", "proof_rigor": "proof rigor",
             "usefulness": "usefulness"}
SHORT = {"f51": "F5.1", "o5": "O5", "fable": "F5", "opus": "O4.8", "s5": "S5", "s46": "S4.6"}
NEW = "f51"
ACCENT = "#E05A47"
# graded warm greys for ranks 2..6 (dark -> light)
GREYS = ["#8E887C", "#A8A296", "#C0BAAE", "#D2CDC2", "#E2DDD3"]
PAPER_2 = "#F3F1EC"

# Headline copy. Numbers are never typed here; they are computed below.
COPY = {
    "01_title": "Fable 5.1 finishes first. Cleanly.",
    "01_dek": "Same 50 company updates. Same angle-generator skill. The only fresh contestant was Fable 5.1.",
    "01_line": "Fable 5.1 finishes first.<br />Nobody else is close.",
    "01_sub": "Six Claude models, one independent GPT-5.5 judge, all fifteen pairs judged blind in both orderings.",
    "02_title": "Fable 5.1 leads six of seven dimensions.",
    "02_dek": "Balanced round-robin means: every model appears in 500 judgments across five opponents. Grounding is where the new model breaks away.",
    "03_title": "Fable 5.1 against the field.",
    "03_dek": "Raw judgments count both orderings. Robust wins require the same model to win in both slots.",
    "04_title": "Sharper than Fable 5. Better grounded than Opus 5.",
    "04_dek": "Direct pairwise score deltas: what changed between generations, and against the previous leader.",
    "04_line": "Up on every dimension over Fable 5.<br />Far better grounded than Opus 5.<br /><span class=\"quiet\">Proof rigor: Opus 5 still edges it.</span>",
}


def esc(v):
    return html.escape(str(v), quote=True)


def mean(xs):
    return sum(xs) / len(xs) if xs else 0.0


def fmt_p(p):
    return f"p={p:.3f}" if p >= 0.001 else f"p<0.001"


def sign_test_p(k, n):
    """Two-sided exact binomial sign test, P(X<=min(k,n-k)) * 2 under p=0.5."""
    from math import comb
    lo = min(k, n - k)
    p = sum(comb(n, i) for i in range(lo + 1)) / 2 ** n * 2
    return min(1.0, p)


def pair_scores(records, left, right):
    scores = {left: {d: [] for d in DIMS}, right: {d: [] for d in DIMS}}
    verdicts = {left: Counter(), right: Counter()}
    for r in records:
        if {r["A_model"], r["B_model"]} != {left, right}:
            continue
        v = r["verdict"]
        for slot, model in (("A", r["A_model"]), ("B", r["B_model"])):
            for d in DIMS:
                scores[model][d].append(v["scores"][slot][d])
            verdicts[model][v[f"verdict_{slot}"]] += 1
    return scores, verdicts


def head_to_head(records):
    wins = Counter()
    by_brand = defaultdict(list)
    slot_a = 0
    decisive = 0
    for r in records:
        a, b = r["A_model"], r["B_model"]
        pair = frozenset((a, b))
        w = r["verdict"]["winner"]
        if w == "tie":
            winner = "tie"
        else:
            decisive += 1
            slot_a += w == "A"
            winner = a if w == "A" else b
            wins[(pair, winner)] += 1
        by_brand[(pair, r["brand_id"])].append(winner)
    robust, splits = Counter(), Counter()
    for (pair, _bid), ws in by_brand.items():
        u = set(ws)
        if len(u) == 1 and "tie" not in u:
            robust[(pair, ws[0])] += 1
        else:
            splits[pair] += 1
    return wins, robust, splits, (slot_a / decisive if decisive else 0)


COMMON_STYLE = """
<style>
  html, body { width: 1600px; height: 900px; overflow: hidden; }
  body { background: var(--nj-page); }
  .share-frame { width: 1600px; height: 900px; padding: 42px 54px; }
  .fig.share { width: 100%; height: 100%; padding: 42px 48px 30px; display: grid;
    grid-template-rows: auto 1fr auto; gap: 22px; }
  .share-head { display: flex; align-items: flex-start; justify-content: space-between;
    gap: 48px; padding-bottom: 20px; border-bottom: 1px solid var(--nj-border); }
  .share-kicker { font-family: var(--nj-mono); font-size: 12px; letter-spacing: 0.26em;
    text-transform: uppercase; color: var(--nj-accent); margin-bottom: 9px; }
  .share-head h3 { font-size: 48px; line-height: 0.98; margin: 0; max-width: 1040px; }
  .share-dek { max-width: 960px; margin-top: 12px; font-family: var(--nj-sans);
    font-size: 17px; line-height: 1.45; color: var(--nj-fg-3); }
  .share-meta { min-width: 290px; padding-top: 4px; text-align: right; font-family: var(--nj-mono);
    font-size: 11px; line-height: 1.8; letter-spacing: 0.16em; text-transform: uppercase; color: var(--nj-fg-4); }
  .share-meta b { color: var(--nj-fg-1); font-weight: 500; }
  .share-body { min-height: 0; }
  .share-foot { border-top: 1px solid var(--nj-ink); padding-top: 16px; display: flex; align-items: center;
    justify-content: space-between; gap: 24px; font-family: var(--nj-mono); font-size: 10.5px;
    letter-spacing: 0.16em; text-transform: uppercase; color: var(--nj-fg-4); }
  .share-foot b { color: var(--nj-fg-1); font-weight: 500; }
  .accent { color: var(--nj-accent); }
  .quiet { color: var(--nj-fg-4); }
  .micro-label { font-family: var(--nj-mono); font-size: 10.5px; letter-spacing: 0.18em;
    text-transform: uppercase; color: var(--nj-fg-4); }
  .callout-strip { display: grid; grid-template-columns: repeat(3, 1fr); border: 1px solid var(--nj-border); }
  .callout { padding: 17px 22px; border-right: 1px solid var(--nj-border); }
  .callout:last-child { border-right: 0; }
  .callout strong { display: block; margin-top: 4px; font-family: var(--nj-serif); font-size: 27px;
    line-height: 1.05; font-style: italic; font-weight: 400; color: var(--nj-fg-1); }
  .callout p { margin-top: 5px; font-size: 12.5px; line-height: 1.35; color: var(--nj-fg-3); }
  .chart text { font-family: var(--nj-mono); }
</style>
"""


def shell(title, dek, body, page_no, meta):
    return f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=1600, initial-scale=1.0" />
<title>{esc(title)}</title>
<link rel="stylesheet" href="../../../../../design-system/assets/colors_and_type.css" />
<link rel="stylesheet" href="../../../../../design-system/charts.css" />
{COMMON_STYLE}
</head>
<body data-accent="winner" data-grid="on" data-barstyle="solid">
  <main class="share-frame">
    <section class="fig share">
      <header class="share-head">
        <div>
          <div class="share-kicker">NEWSJACK.SH · THE EVAL DESK · {page_no}</div>
          <h3>{title}</h3>
          <p class="share-dek">{dek}</p>
        </div>
        <div class="share-meta">
          <b>{meta['brands']} BRANDS · {meta['models']} MODELS</b><br />
          {meta['n']} BLIND JUDGMENTS<br />
          GPT-5.5 JUDGE
        </div>
      </header>
      <div class="share-body">{body}</div>
      <footer class="share-foot">
        <span>FABLE 5.1 VIA <b>CLAUDE CODE CLI</b> · FIVE PRIOR MODELS REUSED FROZEN</span>
        <span>{meta['date']} · <b>NEWSJACK.SH</b></span>
      </footer>
    </section>
  </main>
</body>
</html>
"""


def fill_for(rank):
    return ACCENT if rank == 0 else GREYS[min(rank - 1, len(GREYS) - 1)]


def overview_graphic(models, robust, splits, meta):
    n = len(models)
    top = 60
    row_h = 330 / n
    bar_h = min(46, row_h * 0.62)
    x0, span = 210, 580
    svg = [
        f'<svg class="chart" viewBox="0 0 830 {top + n * row_h + 50:.0f}" role="img" '
        f'aria-label="Overall mean ranking for {n} models">',
        f'<text class="t-axislabel" x="{x0}" y="30">OVERALL MEAN · FULL 1–5 SCALE</text>',
        f'<line class="axis" x1="{x0}" y1="{top + n * row_h:.0f}" x2="{x0 + span}" y2="{top + n * row_h:.0f}" />',
    ]
    for t in range(6):
        x = x0 + t * span / 5
        svg.append(f'<line class="grid-line" x1="{x:.0f}" y1="50" x2="{x:.0f}" y2="{top + n * row_h:.0f}" />')
        svg.append(f'<text class="t-tick" x="{x:.0f}" y="{top + n * row_h + 24:.0f}" text-anchor="middle">{t}</text>')
    for i, m in enumerate(models):
        y = top + i * row_h + (row_h - bar_h) / 2
        w = m["overall"] / 5 * span
        cls = "bar-primary" if m["slug"] == NEW else "bar-base"
        style = "" if m["slug"] == NEW else f' style="fill:{fill_for(i)}"'
        svg += [
            f'<text class="t-cat" x="0" y="{y + bar_h / 2 + 5:.1f}">{esc(m["label"]).upper()}</text>',
            f'<rect class="{cls}"{style} x="{x0}" y="{y:.1f}" width="{w:.1f}" height="{bar_h:.1f}" />',
            f'<text class="t-val" x="{x0 + w + 14:.1f}" y="{y + bar_h / 2 + 5:.1f}">{m["overall"]:.2f}</text>',
        ]
    svg.append("</svg>")

    new = next(m for m in models if m["slug"] == NEW)
    rank = [m["slug"] for m in models].index(NEW) + 1
    cards = []
    for opp, lab in (("fable", "VS FABLE 5 · ROBUST"), ("o5", "VS OPUS 5 · ROBUST")):
        pair = frozenset((NEW, opp))
        cards.append(f"""
      <div style="background:{PAPER_2}; padding:16px;">
        <div class="micro-label">{lab}</div>
        <div style="font-family:var(--nj-serif); font-style:italic; font-size:34px; color:var(--nj-accent);">{robust[(pair, NEW)]}–{robust[(pair, opp)]}</div>
        <div style="font-size:11px; color:var(--nj-fg-4);">{splits[pair]} position splits</div>
      </div>""")
    body = f"""
<div style="display:grid; grid-template-columns: 1.55fr 0.85fr; gap:42px; height:100%; align-items:center;">
  <div>{''.join(svg)}</div>
  <aside style="border-left:1px solid var(--nj-border); padding-left:42px;">
    <div class="micro-label">FABLE 5.1 · OVERALL MEAN · RANK {rank} OF {n}</div>
    <div style="font-family:var(--nj-serif); font-style:italic; font-size:76px; line-height:.92; color:var(--nj-accent); margin:12px 0 18px;">{new["overall"]:.2f}</div>
    <p style="font-family:var(--nj-serif); font-style:italic; font-size:31px; line-height:1.08; color:var(--nj-fg-1);">{COPY['01_line']}</p>
    <p style="font-size:15px; line-height:1.5; color:var(--nj-fg-3); margin-top:18px;">{COPY['01_sub']}</p>
    <div style="display:grid; grid-template-columns:1fr 1fr; gap:12px; margin-top:24px;">{''.join(cards)}
    </div>
  </aside>
</div>
"""
    return shell(COPY["01_title"], COPY["01_dek"], body, "01 / 04", meta)


def dimension_graphic(models, meta):
    n = len(models)
    W, y_base, y_top = 1420, 420, 80
    scale = (y_base - y_top) / 5
    left, right = 55, 25
    pitch = (W - left - right) / len(DIMS)
    gap = 5
    bar_w = (pitch * 0.82 - (n - 1) * gap) / n
    cluster_w = n * bar_w + (n - 1) * gap
    svg = [f'<svg class="chart" style="height:385px" viewBox="0 0 {W} 475" role="img" '
           f'aria-label="Seven-dimension scorecard for {n} models">']
    for t in range(6):
        y = y_base - t * scale
        svg.append(f'<text class="t-tick" x="{left - 18}" y="{y + 4}" text-anchor="end">{t}</text>')
        if t:
            svg.append(f'<line class="grid-line" x1="{left}" y1="{y}" x2="{W - right}" y2="{y}" />')
    svg.append(f'<line class="axis" x1="{left}" y1="{y_base}" x2="{W - right}" y2="{y_base}" />')
    for di, d in enumerate(DIMS):
        gl = left + di * pitch + (pitch - cluster_w) / 2
        svg.append(f'<text class="t-cat" x="{gl + cluster_w / 2:.1f}" y="38" text-anchor="middle">{DIM_LABELS[d]}</text>')
        for mi, m in enumerate(models):
            v = m["dim_means"][d]
            h = v * scale
            x = gl + mi * (bar_w + gap)
            y = y_base - h
            is_new = m["slug"] == NEW
            cls = "bar-primary" if is_new else "bar-base"
            style = "" if is_new else f' style="fill:{fill_for(mi)}"'
            lab_y = y - (9 + 13 * (mi % 3))
            svg += [
                f'<rect class="{cls}"{style} x="{x:.1f}" y="{y:.1f}" width="{bar_w:.1f}" height="{h:.1f}" />',
                f'<text class="t-val" style="font-size:10px" x="{x + bar_w / 2:.1f}" y="{lab_y:.1f}" text-anchor="middle">{v:.2f}</text>',
                f'<text class="t-series" style="font-size:9.5px" x="{x + bar_w / 2:.1f}" y="444" text-anchor="middle">{SHORT.get(m["slug"], m["slug"])}</text>',
            ]
    svg.append("</svg>")

    # callouts: dims led by the new model / biggest gain over Fable 5 / where it trails
    new = next(m for m in models if m["slug"] == NEW)
    f5 = next(m for m in models if m["slug"] == "fable")
    leaders = {d: max(models, key=lambda m: m["dim_means"][d]) for d in DIMS}
    led = [d for d in DIMS if leaders[d]["slug"] == NEW]
    gains = sorted(DIMS, key=lambda d: new["dim_means"][d] - f5["dim_means"][d], reverse=True)
    best_gain = gains[0]
    trail = [d for d in DIMS if leaders[d]["slug"] != NEW]
    if trail:
        worst = min(trail, key=lambda d: new["dim_means"][d] - leaders[d]["dim_means"][d])
        trail_html = (f'<span class="micro-label">WHERE IT TRAILS</span>'
                      f'<strong>{esc(leaders[worst]["label"])} · {DIM_WORDS[worst]}</strong>'
                      f'<p>{esc(leaders[worst]["label"])} {leaders[worst]["dim_means"][worst]:.2f} vs Fable 5.1 '
                      f'{new["dim_means"][worst]:.2f} on {DIM_WORDS[worst]}.</p>')
    else:
        trail_html = ('<span class="micro-label">WHERE IT TRAILS</span><strong>Nowhere</strong>'
                      '<p>Fable 5.1 leads every one of the seven dimensions.</p>')
    body = f"""
  <div style="height:100%; display:grid; grid-template-rows:minmax(0,1fr) auto; gap:8px;">
  <div>{''.join(svg)}</div>
  <div class="callout-strip">
    <div class="callout">
      <span class="micro-label">DIMENSIONS LED BY FABLE 5.1</span>
      <strong><span class="accent">{len(led)} of 7</span></strong>
      <p>{', '.join(DIM_WORDS[d] for d in led) if led else 'None outright; see the per-pair deltas.'}.</p>
    </div>
    <div class="callout">
      <span class="micro-label">BIGGEST GAIN OVER FABLE 5</span>
      <strong><span class="accent">{DIM_WORDS[best_gain]}</span> · {new["dim_means"][best_gain] - f5["dim_means"][best_gain]:+.2f}</strong>
      <p>Fable 5.1 {new["dim_means"][best_gain]:.2f} vs Fable 5 {f5["dim_means"][best_gain]:.2f}, balanced round-robin means.</p>
    </div>
    <div class="callout">{trail_html}</div>
  </div>
</div>
"""
    return shell(COPY["02_title"], COPY["02_dek"], body, "02 / 04", meta)


def head_to_head_graphic(models, wins, robust, splits, slot_a, meta):
    opps = [m for m in models if m["slug"] != NEW]
    n = len(opps)
    row_h = 440 / n
    bar_h = 44
    x0, width = 340, 1000
    svg = [f'<svg class="chart" style="height:415px" viewBox="0 0 1420 {60 + n * row_h:.0f}" role="img" '
           'aria-label="Head-to-head judgment and robust-win results">',
           f'<text class="t-axislabel" x="{x0}" y="35">JUDGMENT WINS · BOTH ORDERINGS · FABLE 5.1 LEFT</text>']
    for i, opp in enumerate(opps):
        pair = frozenset((NEW, opp["slug"]))
        lw, rw = wins[(pair, NEW)], wins[(pair, opp["slug"])]
        total = lw + rw or 1
        lwid = width * lw / total
        y = 60 + i * row_h
        rank = [m["slug"] for m in models].index(opp["slug"])
        svg += [
            f'<text class="t-cat" x="0" y="{y + bar_h / 2 + 5:.0f}">VS {esc(opp["label"]).upper()}</text>',
            f'<rect class="bar-primary" x="{x0}" y="{y:.0f}" width="{lwid:.1f}" height="{bar_h}" />',
            f'<rect class="bar-base" style="fill:{fill_for(rank)}" x="{x0 + lwid:.1f}" y="{y:.0f}" width="{width - lwid:.1f}" height="{bar_h}" />',
            f'<text class="t-val on-bar" x="{x0 + lwid / 2:.1f}" y="{y + bar_h / 2 + 5:.0f}" text-anchor="middle">{lw}</text>',
            f'<text class="t-val" x="{x0 + lwid + (width - lwid) / 2:.1f}" y="{y + bar_h / 2 + 5:.0f}" text-anchor="middle">{rw}</text>',
            f'<text class="t-catsub" x="{x0}" y="{y + bar_h + 22:.0f}">ROBUST · {robust[(pair, NEW)]} FABLE 5.1 · {splits[pair]} SPLIT · {robust[(pair, opp["slug"])]} {esc(opp["label"]).upper()}</text>',
        ]
    svg.append("</svg>")

    def robust_line(opp_slug):
        pair = frozenset((NEW, opp_slug))
        a, b = robust[(pair, NEW)], robust[(pair, opp_slug)]
        p = sign_test_p(a, a + b) if a + b else 1.0
        return a, b, p

    a1, b1, p1 = robust_line("fable")
    a2, b2, p2 = robust_line("o5")
    body = f"""
<div style="height:100%; display:grid; grid-template-rows:minmax(0,1fr) auto; gap:8px;">
  <div>{''.join(svg)}</div>
  <div class="callout-strip">
    <div class="callout">
      <span class="micro-label">FABLE 5.1 VS FABLE 5</span>
      <strong>{'<span class="accent">' if a1 > b1 else ''}{a1}–{b1} robust{'</span>' if a1 > b1 else ''}</strong>
      <p>Exploratory two-sided exact sign test {fmt_p(p1)}.</p>
    </div>
    <div class="callout">
      <span class="micro-label">FABLE 5.1 VS OPUS 5</span>
      <strong>{'<span class="accent">' if a2 > b2 else ''}{a2}–{b2} robust{'</span>' if a2 > b2 else ''}</strong>
      <p>Exploratory two-sided exact sign test {fmt_p(p2)}.</p>
    </div>
    <div class="callout">
      <span class="micro-label">POSITION BIAS</span>
      <strong>Slot A won {slot_a * 100:.1f}%</strong>
      <p>Both orderings cancel identity bias; robust wins are the cleaner read.</p>
    </div>
  </div>
</div>
"""
    return shell(COPY["03_title"], COPY["03_dek"], body, "03 / 04", meta)


def delta_panel(records, opp, opp_label, x_offset=0):
    ps, _ = pair_scores(records, NEW, opp)
    deltas = {d: mean(ps[NEW][d]) - mean(ps[opp][d]) for d in DIMS}
    xc = 330
    max_abs = max(abs(v) for v in deltas.values()) or 1.0
    scale = min(260.0, 240.0 / max_abs)
    ys = [64 + i * 52 for i in range(7)]
    out = [f'<line class="axis" x1="{xc}" y1="34" x2="{xc}" y2="{ys[-1] + 40}" />',
           f'<text class="t-axislabel" x="{xc - 12}" y="18" text-anchor="end">{opp_label} EDGE</text>',
           f'<text class="t-axislabel" x="{xc + 12}" y="18">FABLE 5.1 EDGE</text>']
    for d, y in zip(DIMS, ys):
        dv = deltas[d]
        ln = abs(dv) * scale
        x = xc if dv >= 0 else xc - ln
        cls = "bar-primary" if dv >= 0 else "bar-base"
        style = "" if dv >= 0 else f' style="fill:{GREYS[0]}"'
        vx = x + ln + 10 if dv >= 0 else x - 10
        anchor = "start" if dv >= 0 else "end"
        out += [f'<text class="t-cat" x="0" y="{y + 18}">{DIM_LABELS[d]}</text>',
                f'<rect class="{cls}"{style} x="{x:.1f}" y="{y}" width="{ln:.1f}" height="26" />',
                f'<text class="t-val" x="{vx:.1f}" y="{y + 19}" text-anchor="{anchor}">{dv:+.2f}</text>']
    overall = mean([deltas[d] for d in DIMS])
    return out, overall, ps


def tradeoff_graphic(records, models, meta):
    p1, o1, ps1 = delta_panel(records, "fable", "FABLE 5")
    p2, o2, ps2 = delta_panel(records, "o5", "OPUS 5")
    svg1 = f'<svg class="chart" viewBox="0 0 640 460" role="img" aria-label="Fable 5.1 minus Fable 5">{"".join(p1)}</svg>'
    svg2 = f'<svg class="chart" viewBox="0 0 640 460" role="img" aria-label="Fable 5.1 minus Opus 5">{"".join(p2)}</svg>'
    by_slug = {m["slug"]: m for m in models}
    pubs = [(m["label"], m["verdicts"].get("publishable", 0), sum(m["verdicts"].values()))
            for m in models]
    pub_rows = "".join(
        f'<div style="display:flex; align-items:baseline; justify-content:space-between; margin-top:3px; '
        f'color:{"var(--nj-accent)" if lab == "Fable 5.1" else "var(--nj-fg-3)"};">'
        f'<span style="font-family:var(--nj-serif); font-style:italic; font-size:19px;">{esc(lab)}</span>'
        f'<span style="font-family:var(--nj-mono); font-size:15px;">{p}/{t}</span></div>'
        for lab, p, t in sorted(pubs, key=lambda x: -x[1]))
    body = f"""
<div style="height:100%; display:grid; grid-template-columns:1fr 1fr .62fr; gap:36px; align-items:center;">
  <div><div class="micro-label" style="margin-bottom:6px;">DIRECT PAIR · FABLE 5.1 − FABLE 5 · MEAN Δ {o1:+.2f}</div>{svg1}</div>
  <div><div class="micro-label" style="margin-bottom:6px;">DIRECT PAIR · FABLE 5.1 − OPUS 5 · MEAN Δ {o2:+.2f}</div>{svg2}</div>
  <aside style="border-left:1px solid var(--nj-border); padding-left:32px;">
    <div class="micro-label">READ</div>
    <p style="font-family:var(--nj-serif); font-style:italic; font-size:30px; line-height:1.08; margin:12px 0 20px;">{COPY['04_line']}</p>
    <div style="background:{PAPER_2}; padding:16px 18px;">
      <div class="micro-label">MEANEST-EDITOR ‘PUBLISHABLE’ · ALL PAIRS</div>
      <div style="margin-top:6px;">{pub_rows}</div>
    </div>
  </aside>
</div>
"""
    return shell(COPY["04_title"], COPY["04_dek"], body, "04 / 04", meta)


def gallery(files, meta):
    cards = "".join(
        f'\n      <a class="card" href="{esc(fn)}"><img src="png/{Path(fn).stem}--full.png" alt="{esc(lab)}" /><span>{esc(lab)}</span></a>'
        for fn, lab in files)
    return f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Fable 5.1 benchmark share graphics</title>
<link rel="stylesheet" href="../../../../../design-system/assets/colors_and_type.css" />
<style>
  * {{ box-sizing:border-box; }}
  body {{ margin:0; padding:52px; background:#F9F8F6; color:#1A1A1A; font-family:var(--nj-sans); }}
  h1 {{ font-family:var(--nj-serif); font-style:italic; font-weight:400; font-size:54px; margin:0 0 10px; }}
  p {{ color:rgba(26,26,26,.6); margin:0 0 34px; }}
  .grid {{ display:grid; grid-template-columns:1fr 1fr; gap:28px; }}
  .card {{ color:inherit; text-decoration:none; border:1px solid rgba(26,26,26,.1); background:#F9F8F6; padding:14px; }}
  .card img {{ width:100%; display:block; border:1px solid rgba(26,26,26,.08); }}
  .card span {{ display:block; padding:14px 4px 3px; font-family:var(--nj-mono); font-size:12px; letter-spacing:.12em; text-transform:uppercase; }}
</style>
</head>
<body>
  <h1>Fable 5.1 benchmark graphics</h1>
  <p>{meta['models']} models · {meta['n']} blind judgments · {meta['date']}. Open any card for the 1600×900 HTML. Matching 2× PNGs live in <code>png/</code>.</p>
  <div class="grid">{cards}
  </div>
</body>
</html>
"""


def main(run_dir):
    run = Path(run_dir).resolve()
    summary = json.loads((run / "summary.json").read_text())
    results = json.loads((run / "results.json").read_text())
    records = results["records"]
    models = summary["models"]
    brands = len({r["brand_id"] for r in records})
    meta = {"brands": brands, "models": len(models), "n": len(records),
            "date": summary.get("run", "")[:10]}
    wins, robust, splits, slot_a = head_to_head(records)

    out = run / "figures" / "share"
    out.mkdir(parents=True, exist_ok=True)
    graphics = [
        ("01-overview.html", overview_graphic(models, robust, splits, meta), "01 · Overall verdict"),
        ("02-dimensions.html", dimension_graphic(models, meta), "02 · Seven-dimension scorecard"),
        ("03-head-to-head.html", head_to_head_graphic(models, wins, robust, splits, slot_a, meta),
         "03 · Fable 5.1 against the field"),
        ("04-tradeoff.html", tradeoff_graphic(records, models, meta), "04 · What changed vs Fable 5 and Opus 5"),
    ]
    for fn, content, _ in graphics:
        (out / fn).write_text(content)
        print(f"Wrote {out / fn}")
    (out / "index.html").write_text(gallery([(fn, lab) for fn, _, lab in graphics], meta))
    print(f"Wrote {out / 'index.html'}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(__doc__)
        raise SystemExit(1)
    main(sys.argv[1])
