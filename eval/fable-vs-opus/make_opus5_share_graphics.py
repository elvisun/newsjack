#!/usr/bin/env python3
"""Build screenshot-ready Opus 5 benchmark graphics.

Reads the finished three-model run and writes four standalone 1600x900 HTML
cards under figures/share/. The eval design-system validator can render each
card at 2x resolution for publication.

Usage:
    python3 make_opus5_share_graphics.py runs/2026-07-24-opus5
"""

import html
import json
import sys
from collections import Counter, defaultdict
from pathlib import Path


DIMS = [
    "news_value",
    "distinctness",
    "journalist_shape",
    "grounding",
    "anti_slop",
    "proof_rigor",
    "usefulness",
]

DIM_LABELS = {
    "news_value": "NEWS VALUE",
    "distinctness": "DISTINCT",
    "journalist_shape": "SHAPE",
    "grounding": "GROUNDING",
    "anti_slop": "ANTI-SLOP",
    "proof_rigor": "PROOF",
    "usefulness": "USEFUL",
}

SHORT = {"o5": "O5", "fable": "FB", "opus": "O4.8"}
ACCENT = "#E05A47"
GREY_DARK = "#9E988C"
GREY_LIGHT = "#D8D3C9"
PAPER_2 = "#F3F1EC"


def esc(value):
    return html.escape(str(value), quote=True)


def mean(values):
    return sum(values) / len(values) if values else 0.0


def model_scores_for_pair(records, left, right):
    scores = {left: {d: [] for d in DIMS}, right: {d: [] for d in DIMS}}
    verdicts = {left: Counter(), right: Counter()}
    for record in records:
        if {record["A_model"], record["B_model"]} != {left, right}:
            continue
        verdict = record["verdict"]
        for slot, model in (("A", record["A_model"]), ("B", record["B_model"])):
            for dim in DIMS:
                scores[model][dim].append(verdict["scores"][slot][dim])
            verdicts[model][verdict[f"verdict_{slot}"]] += 1
    return scores, verdicts


def head_to_head(records):
    wins = Counter()
    by_brand = defaultdict(list)
    for record in records:
        a, b = record["A_model"], record["B_model"]
        pair = frozenset((a, b))
        verdict = record["verdict"]
        if verdict["winner"] == "tie":
            winner = "tie"
        else:
            winner = a if verdict["winner"] == "A" else b
            wins[(pair, winner)] += 1
        by_brand[(pair, record["brand_id"])].append(winner)

    robust = Counter()
    splits = Counter()
    for (pair, _brand_id), winners in by_brand.items():
        unique = set(winners)
        if len(unique) == 1 and "tie" not in unique:
            robust[(pair, winners[0])] += 1
        else:
            splits[pair] += 1
    return wins, robust, splits


COMMON_STYLE = """
<style>
  html, body {
    width: 1600px;
    height: 900px;
    overflow: hidden;
  }
  body { background: var(--nj-page); }
  .share-frame {
    width: 1600px;
    height: 900px;
    padding: 42px 54px;
  }
  .fig.share {
    width: 100%;
    height: 100%;
    padding: 42px 48px 30px;
    display: grid;
    grid-template-rows: auto 1fr auto;
    gap: 22px;
  }
  .share-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 48px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--nj-border);
  }
  .share-kicker {
    font-family: var(--nj-mono);
    font-size: 12px;
    letter-spacing: 0.26em;
    text-transform: uppercase;
    color: var(--nj-accent);
    margin-bottom: 9px;
  }
  .share-head h3 {
    font-size: 48px;
    line-height: 0.98;
    margin: 0;
    max-width: 980px;
  }
  .share-dek {
    max-width: 920px;
    margin-top: 12px;
    font-family: var(--nj-sans);
    font-size: 17px;
    line-height: 1.45;
    color: var(--nj-fg-3);
  }
  .share-meta {
    min-width: 290px;
    padding-top: 4px;
    text-align: right;
    font-family: var(--nj-mono);
    font-size: 11px;
    line-height: 1.8;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--nj-fg-4);
  }
  .share-meta b { color: var(--nj-fg-1); font-weight: 500; }
  .share-body { min-height: 0; }
  .share-foot {
    border-top: 1px solid var(--nj-ink);
    padding-top: 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    font-family: var(--nj-mono);
    font-size: 10.5px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--nj-fg-4);
  }
  .share-foot b { color: var(--nj-fg-1); font-weight: 500; }
  .accent { color: var(--nj-accent); }
  .ink { color: var(--nj-fg-1); }
  .quiet { color: var(--nj-fg-4); }
  .micro-label {
    font-family: var(--nj-mono);
    font-size: 10.5px;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--nj-fg-4);
  }
  .callout-strip {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    border: 1px solid var(--nj-border);
  }
  .callout {
    padding: 17px 22px;
    border-right: 1px solid var(--nj-border);
  }
  .callout:last-child { border-right: 0; }
  .callout strong {
    display: block;
    margin-top: 4px;
    font-family: var(--nj-serif);
    font-size: 27px;
    line-height: 1.05;
    font-style: italic;
    font-weight: 400;
    color: var(--nj-fg-1);
  }
  .callout p {
    margin-top: 5px;
    font-size: 12.5px;
    line-height: 1.35;
    color: var(--nj-fg-3);
  }
  .chart text { font-family: var(--nj-mono); }
</style>
"""


def shell(title, dek, body, page_no):
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
          <b>50 BRANDS</b><br />
          300 BLIND JUDGMENTS<br />
          GPT-5.5 JUDGE
        </div>
      </header>
      <div class="share-body">{body}</div>
      <footer class="share-foot">
        <span>OPUS 5 VIA <b>CLAUDE CODE CLI</b> · FABLE 5 + OPUS 4.8 REUSED</span>
        <span>2026-07-24 · <b>NEWSJACK.SH</b></span>
      </footer>
    </section>
  </main>
</body>
</html>
"""


def overview_graphic(models, robust, splits):
    by_slug = {m["slug"]: m for m in models}
    rows = [by_slug[s] for s in ("o5", "fable", "opus")]
    fills = [ACCENT, GREY_DARK, GREY_LIGHT]
    y_positions = [90, 205, 320]
    svg = [
        '<svg class="chart" viewBox="0 0 830 430" role="img" '
        'aria-label="Overall mean ranking for Opus 5, Fable 5, and Opus 4.8">',
        '<text class="t-axislabel" x="210" y="30">OVERALL MEAN · FULL 1–5 SCALE</text>',
        '<line class="axis" x1="210" y1="390" x2="790" y2="390" />',
    ]
    for tick in range(6):
        x = 210 + tick * 116
        svg.append(f'<line class="grid-line" x1="{x}" y1="50" x2="{x}" y2="390" />')
        svg.append(
            f'<text class="t-tick" x="{x}" y="414" text-anchor="middle">{tick}</text>'
        )
    for i, (model, y, fill) in enumerate(zip(rows, y_positions, fills)):
        width = model["overall"] / 5 * 580
        cls = "bar-primary" if i == 0 else "bar-base"
        style = "" if i == 0 else f' style="fill:{fill}"'
        svg.extend(
            [
                f'<text class="t-cat" x="0" y="{y + 23}">{esc(model["label"]).upper()}</text>',
                f'<rect class="{cls}"{style} x="210" y="{y}" width="{width:.1f}" height="46" />',
                f'<text class="t-val" x="{210 + width + 16:.1f}" y="{y + 30}">{model["overall"]:.2f}</text>',
            ]
        )
    svg.append("</svg>")

    o5_fable = frozenset(("o5", "fable"))
    o5_opus = frozenset(("o5", "opus"))
    body = f"""
<div style="display:grid; grid-template-columns: 1.55fr 0.85fr; gap:42px; height:100%; align-items:center;">
  <div>{''.join(svg)}</div>
  <aside style="border-left:1px solid var(--nj-border); padding-left:42px;">
    <div class="micro-label">THE RESULT</div>
    <div style="font-family:var(--nj-serif); font-style:italic; font-size:76px; line-height:.92; color:var(--nj-accent); margin:12px 0 18px;">4.52</div>
    <p style="font-family:var(--nj-serif); font-style:italic; font-size:31px; line-height:1.08; color:var(--nj-fg-1);">Opus 5 finishes first.<br />Fable keeps it honest.</p>
    <p style="font-size:15px; line-height:1.5; color:var(--nj-fg-3); margin-top:18px;">Clear generational improvement over Opus 4.8. Narrow, non-decisive edge over Fable 5.</p>
    <div style="display:grid; grid-template-columns:1fr 1fr; gap:12px; margin-top:24px;">
      <div style="background:{PAPER_2}; padding:16px;">
        <div class="micro-label">VS FABLE · ROBUST</div>
        <div style="font-family:var(--nj-serif); font-style:italic; font-size:34px; color:var(--nj-accent);">15–9</div>
        <div style="font-size:11px; color:var(--nj-fg-4);">{splits[o5_fable]} position splits</div>
      </div>
      <div style="background:{PAPER_2}; padding:16px;">
        <div class="micro-label">VS OPUS 4.8 · ROBUST</div>
        <div style="font-family:var(--nj-serif); font-style:italic; font-size:34px; color:var(--nj-accent);">20–4</div>
        <div style="font-size:11px; color:var(--nj-fg-4);">{splits[o5_opus]} position splits</div>
      </div>
    </div>
  </aside>
</div>
"""
    return shell(
        "Opus 5 finishes first. Fable keeps it honest.",
        "Same 50 company updates. Same angle-generator skill. The only fresh contestant was Opus 5.",
        body,
        "01 / 04",
    )


def dimension_graphic(models):
    by_slug = {m["slug"]: m for m in models}
    ordered = [by_slug[s] for s in ("o5", "fable", "opus")]
    fills = [ACCENT, GREY_DARK, GREY_LIGHT]
    W, y_base, y_top = 1420, 420, 80
    scale = (y_base - y_top) / 5
    left, right = 55, 25
    pitch = (W - left - right) / len(DIMS)
    bar_w, gap = 34, 9
    cluster_w = 3 * bar_w + 2 * gap
    svg = [
        f'<svg class="chart" style="height:385px" viewBox="0 0 {W} 475" role="img" '
        'aria-label="Seven-dimension scorecard for the three models">'
    ]
    for tick in range(6):
        y = y_base - tick * scale
        svg.append(
            f'<text class="t-tick" x="{left - 18}" y="{y + 4}" text-anchor="end">{tick}</text>'
        )
        if tick:
            svg.append(
                f'<line class="grid-line" x1="{left}" y1="{y}" x2="{W - right}" y2="{y}" />'
            )
    svg.append(
        f'<line class="axis" x1="{left}" y1="{y_base}" x2="{W - right}" y2="{y_base}" />'
    )
    for dim_index, dim in enumerate(DIMS):
        group_left = left + dim_index * pitch + (pitch - cluster_w) / 2
        svg.append(
            f'<text class="t-cat" x="{group_left + cluster_w / 2:.1f}" y="38" '
            f'text-anchor="middle">{DIM_LABELS[dim]}</text>'
        )
        for model_index, (model, fill) in enumerate(zip(ordered, fills)):
            value = model["dim_means"][dim]
            height = value * scale
            x = group_left + model_index * (bar_w + gap)
            y = y_base - height
            cls = "bar-primary" if model_index == 0 else "bar-base"
            style = "" if model_index == 0 else f' style="fill:{fill}"'
            svg.extend(
                [
                    f'<rect class="{cls}"{style} x="{x:.1f}" y="{y:.1f}" width="{bar_w}" height="{height:.1f}" />',
                    f'<text class="t-val" x="{x + bar_w / 2:.1f}" y="{y - 10:.1f}" text-anchor="middle">{value:.2f}</text>',
                    f'<text class="t-series" x="{x + bar_w / 2:.1f}" y="444" text-anchor="middle">{SHORT[model["slug"]]}</text>',
                ]
            )
    svg.append("</svg>")

    body = f"""
  <div style="height:100%; display:grid; grid-template-rows:minmax(0,1fr) auto; gap:8px;">
  <div>{''.join(svg)}</div>
  <div class="callout-strip">
    <div class="callout">
      <span class="micro-label">SHARPEST REPORTER SHAPE</span>
      <strong><span class="accent">Opus 5</span> · 4.92</strong>
      <p>Best at naming the real sub-beat, outlet type, and reporting path.</p>
    </div>
    <div class="callout">
      <span class="micro-label">HARDEST PROOF STANDARD</span>
      <strong><span class="accent">Opus 5</span> · 4.97</strong>
      <p>Turns missing evidence into questions a journalist can actually test.</p>
    </div>
    <div class="callout">
      <span class="micro-label">SAFEST ON FACTS</span>
      <strong>Fable 5 · 3.89</strong>
      <p>Leads grounding by 0.47 and anti-slop by 0.15 over Opus 5.</p>
    </div>
  </div>
</div>
"""
    return shell(
        "Opus 5 wins five dimensions. Fable wins the safety checks.",
        "Balanced round-robin means: every model appears in 200 judgments across both opponents.",
        body,
        "02 / 04",
    )


def head_to_head_graphic(wins, robust, splits):
    pairs = [
        ("o5", "fable", "OPUS 5", "FABLE 5"),
        ("o5", "opus", "OPUS 5", "OPUS 4.8"),
        ("fable", "opus", "FABLE 5", "OPUS 4.8"),
    ]
    colors = {
        "o5": ACCENT,
        "fable": GREY_DARK,
        "opus": GREY_LIGHT,
    }
    pair_y = [95, 235, 375]
    x0, width = 315, 1010
    svg = [
        '<svg class="chart" style="height:360px" viewBox="0 0 1420 500" role="img" '
        'aria-label="Head-to-head judgment and robust-win results">'
    ]
    for index, ((left, right, left_label, right_label), y) in enumerate(
        zip(pairs, pair_y)
    ):
        pair = frozenset((left, right))
        left_wins = wins[(pair, left)]
        right_wins = wins[(pair, right)]
        total = left_wins + right_wins
        left_width = width * left_wins / total
        right_width = width - left_width
        left_cls = "bar-primary" if left == "o5" else "bar-base"
        left_style = "" if left == "o5" else f' style="fill:{colors[left]}"'
        right_value_class = "t-val on-bar" if right == "fable" else "t-val"
        svg.extend(
            [
                f'<text class="t-cat" x="0" y="{y + 25}">{left_label} VS {right_label}</text>',
                f'<rect class="{left_cls}"{left_style} x="{x0}" y="{y}" width="{left_width:.1f}" height="48" />',
                f'<rect class="bar-base" style="fill:{colors[right]}" x="{x0 + left_width:.1f}" y="{y}" width="{right_width:.1f}" height="48" />',
                f'<text class="t-val on-bar" x="{x0 + left_width / 2:.1f}" y="{y + 31}" text-anchor="middle">{left_wins}</text>',
                f'<text class="{right_value_class}" x="{x0 + left_width + right_width / 2:.1f}" y="{y + 31}" text-anchor="middle">{right_wins}</text>',
                f'<text class="t-catsub" x="{x0}" y="{y + 77}">ROBUST · {robust[(pair, left)]} {left_label} · {splits[pair]} SPLIT · {robust[(pair, right)]} {right_label}</text>',
            ]
        )
        if index == 0:
            svg.append(
                f'<text class="t-axislabel" x="{x0}" y="35">JUDGMENT WINS · BOTH ORDERINGS</text>'
            )
    svg.append("</svg>")

    body = f"""
<div style="height:100%; display:grid; grid-template-rows:minmax(0,1fr) auto; gap:8px;">
  <div>{''.join(svg)}</div>
  <div class="callout-strip">
    <div class="callout">
      <span class="micro-label">OPUS 5 VS FABLE</span>
      <strong>Close, not decisive</strong>
      <p>15–9 robust; exploratory exact sign test p=0.307.</p>
    </div>
    <div class="callout">
      <span class="micro-label">OPUS 5 VS OPUS 4.8</span>
      <strong><span class="accent">Clear generation jump</span></strong>
      <p>20–4 robust; exploratory exact sign test p=0.0015.</p>
    </div>
    <div class="callout">
      <span class="micro-label">POSITION BIAS</span>
      <strong>Slot A won 73.7%</strong>
      <p>Both orderings cancel identity bias; robust wins are the cleaner read.</p>
    </div>
  </div>
</div>
"""
    return shell(
        "Opus 5 clears 4.8. Fable remains a real contest.",
        "Raw judgments count both orderings. Robust wins require the same model to win in both slots.",
        body,
        "03 / 04",
    )


def tradeoff_graphic(pair_scores, models):
    o5 = {dim: mean(pair_scores["o5"][dim]) for dim in DIMS}
    fable = {dim: mean(pair_scores["fable"][dim]) for dim in DIMS}
    deltas = {dim: o5[dim] - fable[dim] for dim in DIMS}
    x_center, scale = 765, 650
    y_positions = [64, 120, 176, 232, 288, 344, 400]
    svg = [
        '<svg class="chart" viewBox="0 0 1120 465" role="img" '
        'aria-label="Direct Opus 5 minus Fable 5 dimension deltas">',
        f'<line class="axis" x1="{x_center}" y1="34" x2="{x_center}" y2="432" />',
        f'<text class="t-axislabel" x="{x_center - 14}" y="18" text-anchor="end">FABLE EDGE</text>',
        f'<text class="t-axislabel" x="{x_center + 14}" y="18">OPUS 5 EDGE</text>',
    ]
    for dim, y in zip(DIMS, y_positions):
        delta = deltas[dim]
        length = abs(delta) * scale
        x = x_center if delta >= 0 else x_center - length
        cls = "bar-primary" if delta >= 0 else "bar-base"
        style = "" if delta >= 0 else f' style="fill:{GREY_DARK}"'
        value_x = x + length + 12 if delta >= 0 else x - 12
        anchor = "start" if delta >= 0 else "end"
        svg.extend(
            [
                f'<text class="t-cat" x="0" y="{y + 18}">{DIM_LABELS[dim]}</text>',
                f'<rect class="{cls}"{style} x="{x:.1f}" y="{y}" width="{length:.1f}" height="26" />',
                f'<text class="t-val" x="{value_x:.1f}" y="{y + 19}" text-anchor="{anchor}">{delta:+.2f}</text>',
            ]
        )
    svg.append("</svg>")

    by_slug = {m["slug"]: m for m in models}
    fable_pub = by_slug["fable"]["verdicts"].get("publishable", 0)
    o5_pub = by_slug["o5"]["verdicts"].get("publishable", 0)
    body = f"""
<div style="height:100%; display:grid; grid-template-columns:1.55fr .72fr; gap:44px; align-items:center;">
  <div>{''.join(svg)}</div>
  <aside style="border-left:1px solid var(--nj-border); padding-left:40px;">
    <div class="micro-label">DIRECT O5 − FABLE SCORE DELTA</div>
    <p style="font-family:var(--nj-serif); font-style:italic; font-size:34px; line-height:1.08; margin:14px 0 24px;">Sharper angles.<br />Harder proof.<br /><span class="quiet">Looser facts.</span></p>
    <div style="background:{PAPER_2}; padding:20px 22px; margin-bottom:14px;">
      <div class="micro-label">MOST PUBLISHABLE · FULL ROUND ROBIN</div>
      <div style="display:flex; align-items:baseline; justify-content:space-between; margin-top:8px;">
        <span style="font-family:var(--nj-serif); font-style:italic; font-size:28px;">Fable 5</span>
        <span style="font-family:var(--nj-serif); font-style:italic; font-size:40px;">{fable_pub}/200</span>
      </div>
      <div style="display:flex; align-items:baseline; justify-content:space-between; margin-top:5px; color:var(--nj-fg-3);">
        <span style="font-family:var(--nj-serif); font-style:italic; font-size:24px;">Opus 5</span>
        <span style="font-family:var(--nj-serif); font-style:italic; font-size:32px;">{o5_pub}/200</span>
      </div>
    </div>
    <p style="font-size:14px; line-height:1.45; color:var(--nj-fg-3);">Opus 5 wins more pairwise decisions. Fable needs fewer factual repairs before a journalist sees the work.</p>
  </aside>
</div>
"""
    return shell(
        "Opus 5 is sharper. Fable is safer.",
        "Direct Opus-5-vs-Fable scoring exposes the trade: ambition and proof pressure versus factual restraint.",
        body,
        "04 / 04",
    )


def gallery(files):
    cards = []
    for filename, label in files:
        stem = Path(filename).stem
        png = f"png/{stem}--full.png"
        cards.append(
            f"""
      <a class="card" href="{esc(filename)}">
        <img src="{esc(png)}" alt="{esc(label)}" />
        <span>{esc(label)}</span>
      </a>"""
        )
    return f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Opus 5 benchmark share graphics</title>
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
  <h1>Opus 5 benchmark graphics</h1>
  <p>Open any card for the 1600×900 HTML. Matching 2× PNGs live in <code>png/</code>.</p>
  <div class="grid">{''.join(cards)}
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
    wins, robust, splits = head_to_head(records)
    pair_scores, _pair_verdicts = model_scores_for_pair(records, "o5", "fable")

    out = run / "figures" / "share"
    out.mkdir(parents=True, exist_ok=True)
    graphics = [
        (
            "01-overview.html",
            overview_graphic(models, robust, splits),
            "01 · Overall verdict",
        ),
        (
            "02-dimensions.html",
            dimension_graphic(models),
            "02 · Seven-dimension scorecard",
        ),
        (
            "03-head-to-head.html",
            head_to_head_graphic(wins, robust, splits),
            "03 · Head-to-head and robust wins",
        ),
        (
            "04-tradeoff.html",
            tradeoff_graphic(pair_scores, models),
            "04 · Opus 5 versus Fable tradeoff",
        ),
    ]
    for filename, content, _label in graphics:
        (out / filename).write_text(content)
        print(f"Wrote {out / filename}")
    (out / "index.html").write_text(
        gallery([(filename, label) for filename, _content, label in graphics])
    )
    print(f"Wrote {out / 'index.html'}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(__doc__)
        raise SystemExit(1)
    main(sys.argv[1])
