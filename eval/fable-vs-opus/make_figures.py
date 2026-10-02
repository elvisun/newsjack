#!/usr/bin/env python3
"""Generate N-model study figures from summary.json (aggregate-nmodel.py).

Emits one standalone HTML using the eval design system
("the chart room"): grouped bars (per-dimension mean, winner in vermilion + the
other models in graded warm-grey by overall rank), a head-to-head win matrix,
a model x dimension heatmap, and big-stat callouts. Geometry is computed from the
data so the figure is reproducible. Validate/screenshot with design-system
scripts/validate.mjs afterward.

Usage:
    python3 make_figures.py runs/<run>/summary.json [--out runs/<run>/figures/four-model.html]
"""
import json
import sys
import os

DIM_LABEL = {
    "news_value": "NEWS VALUE", "distinctness": "DISTINCT",
    "journalist_shape": "SHAPE", "grounding": "GROUNDING",
    "anti_slop": "ANTI-SLOP", "proof_rigor": "PROOF", "usefulness": "USEFUL",
}
SHORT = {"f51": "F5.1", "o5": "O5", "opus": "OP", "fable": "FB", "s5": "S5", "s46": "S46"}
# warm-grey shades for overall ranks 2,3,4 (dark -> light); rank 1 is accent
GREY = ["#8E887C", "#A8A296", "#C0BAAE", "#D2CDC2", "#E2DDD3"]
ACCENT = "#E05A47"


def esc(s):
    return str(s).replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def grouped_bars(summary):
    dims = summary["dims"]
    models = summary["models"]            # already ranked by overall desc
    n_models = len(models)
    n = summary["n_judgments"]
    # geometry
    ML, MR = 96, 40
    W = 1650
    yBase, yTop = 520.0, 170.0            # value 0 .. value 5 (headroom for labels)
    scale = (yBase - yTop) / 5.0
    plotW = W - ML - MR
    pitch = plotW / len(dims)
    ig = 8
    clusterW = min(0.80 * pitch, n_models * 34 + (n_models - 1) * ig)
    levels = 3 if n_models >= 5 else 2   # label stagger levels; 3 keeps six labels apart
    val_font = ' style="font-size:10px"' if n_models >= 5 else ""
    bw = (clusterW - (n_models - 1) * ig) / n_models

    def fill_for(rank):
        return ACCENT if rank == 0 else GREY[rank - 1]

    svg = []
    svg.append(f'<svg class="chart" viewBox="0 0 {W} 600" role="img" '
               f'aria-label="Grouped bars of per-dimension mean by model">')
    # y ticks + gridlines (0..5)
    for t in range(6):
        y = yBase - t * scale
        svg.append(f'<text class="t-tick" x="{ML-18}" y="{y+4:.1f}" text-anchor="end">{t}</text>')
        if t > 0:
            svg.append(f'<line class="grid-line" x1="{ML}" y1="{y:.1f}" x2="{W-MR}" y2="{y:.1f}" />')
    svg.append(f'<line class="axis" x1="{ML}" y1="{yBase:.1f}" x2="{W-MR}" y2="{yBase:.1f}" />')

    for gi, d in enumerate(dims):
        gLeft = ML + gi * pitch + (pitch - clusterW) / 2.0
        svg.append(f'<text class="t-cat" x="{gLeft + clusterW/2:.1f}" y="96" '
                   f'text-anchor="middle">{DIM_LABEL.get(d, d.upper())}</text>')
        for rank, m in enumerate(models):
            v = m["dim_means"][d]
            h = v * scale
            x = gLeft + rank * (bw + ig)
            y = yBase - h
            cls = "bar-primary" if rank == 0 else "bar-base"
            style = "" if rank == 0 else f' style="fill:{fill_for(rank)}"'
            svg.append(f'<rect class="{cls}"{style} x="{x:.1f}" y="{y:.1f}" '
                       f'width="{bw:.1f}" height="{h:.1f}" />')
            # stagger adjacent value labels vertically so they never collide
            # even when two bars are near-equal height (odd ranks sit higher)
            lab_y = y - (9 + 14 * (rank % levels))
            svg.append(f'<text class="t-val"{val_font} x="{x+bw/2:.1f}" y="{lab_y:.1f}" '
                       f'text-anchor="middle">{v:.2f}</text>')
            svg.append(f'<text class="t-series" x="{x+bw/2:.1f}" y="538" '
                       f'text-anchor="middle">{SHORT.get(m["slug"], m["slug"])}</text>')
    svg.append('</svg>')

    legend = ['<div class="legend">']
    for rank, m in enumerate(models):
        sw = ACCENT if rank == 0 else fill_for(rank)
        legend.append(
            f'<span class="item"><span class="key" style="background:{sw}"></span>'
            f'{esc(m["label"])} · {m["overall"]:.2f}</span>')
    legend.append(f'<span class="item" style="margin-left:auto;color:var(--nj-fg-4);">'
                  f'SCALE 1–5 · n={n} JUDGMENTS</span></div>')

    winner = models[0]["label"]
    lead = models[0]["overall"] - models[1]["overall"]
    cap = (f'Per-dimension mean on the judge’s 1–5 scale, both orderings '
           f'averaged (n={n}). {esc(winner)} carries the accent; the field trails '
           f'in graded grey, darkest = runner-up. Lead over 2nd on overall mean: '
           f'+{lead:.2f}.')
    return f'''  <div class="fig">
    <div class="fig-head"><span class="fig-tag">NEWSJACK.SH</span><span class="fig-kind">GROUPED BARS</span></div>
    <h3>{n_models} models, seven dimensions</h3>
    <p class="cap">{cap}</p>
    <div class="plot">
{os.linesep.join("      " + s for s in svg)}
    </div>
{legend[0]}
    {os.linesep.join("    " + s for s in legend[1:])}
  </div>'''


def js_array(rows):
    return "[\n    " + ",\n    ".join("[" + ", ".join(f"{v}" for v in r) + "]" for r in rows) + "\n  ]"


def heatmaps_script(summary):
    models = summary["models"]
    dims = summary["dims"]
    mlabels = [m["label"] for m in models]
    dlabels = [DIM_LABEL.get(d, d.upper()) for d in dims]
    dim_matrix = [[round(m["dim_means"][d], 2) for d in dims] for m in models]

    # win matrix: row's decisive win-rate vs col, as % (diagonal = null)
    wm = {(w["row"], w["col"]): w for w in summary["win_matrix"]}
    slugs = [m["slug"] for m in models]
    win_matrix = []
    for r in slugs:
        row = []
        for c in slugs:
            if r == c:
                row.append("null")
            else:
                wr = wm[(r, c)]["row_winrate_decisive"]
                row.append("null" if wr is None else round(wr * 100))
        win_matrix.append(row)

    js_dim_rows = json.dumps(mlabels)
    js_dim_cols = json.dumps(dlabels)
    dim_data = js_array(dim_matrix)
    js_win_rows = json.dumps(mlabels)
    js_win_cols = json.dumps(mlabels)
    win_data = js_array(win_matrix)
    return f'''<script>
(function () {{
  function draw(svgId, rowLabels, colLabels, data, vmin, vmax, suffix) {{
    var svg = document.getElementById(svgId);
    var NS = 'http://www.w3.org/2000/svg';
    var nCols = colLabels.length, nRows = rowLabels.length;
    var x0 = 150, y0 = 20, gap = 6;
    var cw = (960 - x0) / nCols, ch = 54;
    function el(tag, attrs, txt) {{
      var e = document.createElementNS(NS, tag);
      for (var k in attrs) e.setAttribute(k, attrs[k]);
      if (txt != null) e.textContent = txt;
      svg.appendChild(e); return e;
    }}
    for (var c = 0; c < nCols; c++)
      el('text', {{x: x0 + c*cw + (cw-gap)/2, y: y0, 'text-anchor':'middle', 'class':'hm-label'}}, colLabels[c]);
    for (var r = 0; r < nRows; r++) {{
      var ry = y0 + 20 + r*(ch+gap);
      el('text', {{x: x0-14, y: ry+ch/2+4, 'text-anchor':'end', 'class':'hm-label'}}, rowLabels[r]);
      for (var cI = 0; cI < nCols; cI++) {{
        var v = data[r][cI];
        var cx = x0 + cI*cw;
        if (v === null) {{
          el('rect', {{x: cx, y: ry, width: cw-gap, height: ch, fill:'rgba(26,26,26,0.04)', 'class':'hm-cell'}});
          el('text', {{x: cx+(cw-gap)/2, y: ry+ch/2+5, 'text-anchor':'middle', 'class':'hm-val', fill:'#B7B2A8'}}, '—');
          continue;
        }}
        var a = 0.12 + Math.max(0, Math.min(1, (v-vmin)/(vmax-vmin))) * 0.88;
        el('rect', {{x: cx, y: ry, width: cw-gap, height: ch, fill:'rgba(224,90,71,'+a.toFixed(3)+')', 'class':'hm-cell'}});
        el('text', {{x: cx+(cw-gap)/2, y: ry+ch/2+5, 'text-anchor':'middle', 'class':'hm-val', fill: a>0.55?'#fff':'#1A1A1A'}}, v + (suffix||''));
      }}
    }}
  }}
  draw('dimheat', {js_dim_rows}, {js_dim_cols}, {dim_data}, 3.4, 5.0, '');
  draw('winmatrix', {js_win_rows}, {js_win_cols}, {win_data}, 0, 100, '%');
}})();
</script>'''


def big_stats(summary):
    m = summary["models"]
    top, second = m[0], m[1]
    pb = summary.get("position_bias_slotA")
    pub_model = max(m, key=lambda x: x["verdicts"].get("publishable", 0))
    pub_count = pub_model["verdicts"].get("publishable", 0)
    pub_total = sum(pub_model["verdicts"].values())
    cards = [
        ("TOP OVERALL MEAN", f'<span class="accent">{top["overall"]:.2f}</span>',
         f'{esc(top["label"])} on the judge’s 1–5 scale, both orderings averaged.'),
        ("LEAD OVER 2ND", f'<span class="accent">+{top["overall"]-second["overall"]:.2f}</span>',
         f'{esc(top["label"])} vs {esc(second["label"])} on overall mean.'),
        ("MOST ‘PUBLISHABLE’", f'{pub_count}<span class="accent">/{pub_total}</span>',
         f'Meanest-editor {esc(pub_model["label"])} sets rated publishable.'),
        ("POSITION BIAS", f'{pb:.2f}' if pb is not None else "—",
         'Slot-A decisive win-rate across all pairs (0.50 = unbiased); cancelled by both-orderings.'),
    ]
    out = ['<div class="stat-grid">']
    for lab, num, note in cards:
        out.append(f'''  <div class="stat">
    <span class="lab">{lab}</span>
    <span class="fig-num">{num}</span>
    <span class="rule"></span>
    <span class="note">{note}</span>
  </div>''')
    out.append('</div>')
    return "\n".join(out)


def main(summary_path, out_path=None):
    with open(summary_path) as f:
        summary = json.load(f)
    run = summary.get("run", "run")
    if not out_path:
        out_path = os.path.join(os.path.dirname(summary_path), "figures", "four-model.html")
    os.makedirs(os.path.dirname(out_path), exist_ok=True)

    models = summary["models"]
    n_models = len(models)
    hm_h = 40 + n_models * 60 + 10          # rows of 54px + 6px gap, plus header
    model_word = {3: "Three", 4: "Four", 5: "Five", 6: "Six"}.get(n_models, str(n_models))
    order = " › ".join(m["label"] for m in models)
    html = f'''<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>{model_word}-model angle study — {run}</title>
<!--
  Built with eval/design-system ("the chart room") by make_figures.py.
  Data: eval/fable-vs-opus/runs/{run}/summary.json (aggregate-nmodel.py).
  Validate: cd eval/design-system/scripts && node validate.mjs \\
    ../../fable-vs-opus/runs/{run}/figures/four-model.html --out ../../fable-vs-opus/runs/{run}/figures/png
-->
<link rel="stylesheet" href="../../../../design-system/assets/colors_and_type.css" />
<link rel="stylesheet" href="../../../../design-system/charts.css" />
</head>
<body data-accent="winner" data-grid="on" data-barstyle="solid">
<div class="wrap" style="padding-top:48px; padding-bottom:48px;">

  <section class="sec">
    <div class="sec-head">
      <div class="lhs"><span class="num">01</span><h2>dimension scorecard</h2></div>
      <span class="sub">{esc(order)} · GPT-5.5 JUDGE · n={summary["n_judgments"]}</span>
    </div>
{grouped_bars(summary)}
  </section>

  <section class="sec">
    <div class="sec-head">
      <div class="lhs"><span class="num">02</span><h2>head-to-head</h2></div>
      <span class="sub">ROW’S DECISIVE WIN-RATE vs COLUMN</span>
    </div>
    <div class="fig">
      <div class="fig-head"><span class="fig-tag">NEWSJACK.SH</span><span class="fig-kind">WIN MATRIX</span></div>
      <h3>Who beats whom</h3>
      <p class="cap">Each cell is the row model’s share of decisive judgments against the column model (both orderings, 50 brands per pair). Darker = stronger. The diagonal is blank.</p>
      <div class="plot"><svg id="winmatrix" class="chart" viewBox="0 0 980 {hm_h}" role="img" aria-label="Win-rate matrix"></svg></div>
    </div>
    <div class="fig">
      <div class="fig-head"><span class="fig-tag">NEWSJACK.SH</span><span class="fig-kind">HEATMAP</span></div>
      <h3>Mean score, model × dimension</h3>
      <p class="cap">The full evaluation surface. Cell density encodes the 1–5 mean (darker = stronger).</p>
      <div class="plot"><svg id="dimheat" class="chart" viewBox="0 0 980 {hm_h}" role="img" aria-label="Model by dimension heatmap"></svg></div>
    </div>
  </section>

  <section class="sec">
    <div class="sec-head">
      <div class="lhs"><span class="num">03</span><h2>headline figures</h2></div>
      <span class="sub">BIG-STAT CALLOUTS</span>
    </div>
{big_stats(summary)}
  </section>

  <footer class="colophon">
    <span class="c">NEWSJACK.SH — <b>THE EVAL DESK</b></span>
    <span class="c">{n_models}-MODEL ANGLE STUDY · <b>{run}</b></span>
    <span class="c">$ <b>curl newsjack.sh | sh</b></span>
  </footer>
</div>
{heatmaps_script(summary)}
</body>
</html>'''
    with open(out_path, "w") as f:
        f.write(html)
    print(f"Wrote {out_path}")


if __name__ == "__main__":
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    if not args:
        print(__doc__)
        sys.exit(1)
    out = None
    if "--out" in sys.argv:
        out = sys.argv[sys.argv.index("--out") + 1]
    main(args[0], out)
