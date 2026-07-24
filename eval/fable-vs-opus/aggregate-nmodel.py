#!/usr/bin/env python3
"""Aggregate an N-model round-robin angle study.

Generalizes aggregate.py to any number of models. Reads a results.json (one
record per judgment, each carrying the A/B->model mapping and the GPT-5.5
verdict) and re-anchors every judgment to model identity so position bias
cancels. Reports and writes a machine-readable summary.json for the figures:

  - per-model per-dimension mean (each model scored across every judgment it
    appears in) + overall mean, ranked
  - full pairwise win matrix (row model's decisive win-rate vs column model)
  - robust wins: per unordered {brand,pair}, who won BOTH orderings
  - meanest-editor verdict distribution per model
  - position-bias diagnostic (slot-A decisive win-rate; 0.50 = unbiased)

Usage:
    python3 aggregate-nmodel.py runs/<run>/results.json [--summary out.json]
"""
import json
import sys
from collections import Counter, defaultdict
from itertools import combinations

DIMS = ["news_value", "distinctness", "journalist_shape", "grounding",
        "anti_slop", "proof_rigor", "usefulness"]

LABELS = {"o5": "Opus 5", "opus": "Opus 4.8", "fable": "Fable 5",
          "s5": "Sonnet 5", "s46": "Sonnet 4.6"}


def label(m):
    return LABELS.get(m, m)


def main(path, summary_path=None):
    with open(path) as f:
        data = json.load(f)
    records = data["records"]

    models = sorted({r["A_model"] for r in records} |
                    {r["B_model"] for r in records})

    dim_scores = {m: {d: [] for d in DIMS} for m in models}       # model->dim->[..]
    verdicts = {m: Counter() for m in models}                     # model->verdict->n
    # pair win tallies keyed by ordered (winner, loser); ties per unordered pair
    win = defaultdict(int)                                        # (a,b)->a beat b
    ties = defaultdict(int)                                       # frozenset({a,b})->n
    appear = defaultdict(int)                                     # frozenset->judgments
    # winners per (brand, unordered pair) across orderings -> robust wins
    by_bp = defaultdict(list)                                     # (brand,fset)->[winner_model|'tie']
    slot_a_wins = 0
    decisive = 0

    for r in records:
        a, b = r["A_model"], r["B_model"]
        v = r["verdict"]
        for d in DIMS:
            dim_scores[a][d].append(v["scores"]["A"][d])
            dim_scores[b][d].append(v["scores"]["B"][d])
        verdicts[a][v["verdict_A"]] += 1
        verdicts[b][v["verdict_B"]] += 1
        fset = frozenset((a, b))
        appear[fset] += 1
        w = v["winner"]
        bp = (r.get("brand_id"), fset)
        if w == "tie":
            ties[fset] += 1
            by_bp[bp].append("tie")
        else:
            decisive += 1
            winner = a if w == "A" else b
            loser = b if w == "A" else a
            win[(winner, loser)] += 1
            by_bp[bp].append(winner)
            if w == "A":
                slot_a_wins += 1

    def overall_mean(m):
        alls = [s for d in DIMS for s in dim_scores[m][d]]
        return sum(alls) / len(alls) if alls else float("nan")

    def dim_mean(m, d):
        xs = dim_scores[m][d]
        return sum(xs) / len(xs) if xs else float("nan")

    ranked = sorted(models, key=overall_mean, reverse=True)

    # ---- human-readable report ----
    print(f"\n=== {data.get('study','angle study')} ({data.get('run','?')}) ===")
    print(f"judge: {data.get('judge_model','?')}  |  models: "
          f"{', '.join(label(m) for m in ranked)}  |  judgments: {len(records)}\n")

    print("Per-model mean (1-5), ranked by overall:")
    hdr = "  " + f"{'model':<14}" + "".join(f"{d[:9]:>10}" for d in DIMS) + f"{'OVERALL':>10}"
    print(hdr)
    for m in ranked:
        row = "  " + f"{label(m):<14}"
        row += "".join(f"{dim_mean(m,d):>10.2f}" for d in DIMS)
        row += f"{overall_mean(m):>10.2f}"
        print(row)
    print()

    print("Pairwise win matrix (row's decisive win-rate vs column; wins in parens):")
    print("  " + f"{'':<14}" + "".join(f"{label(c):>13}" for c in ranked))
    for rmod in ranked:
        row = "  " + f"{label(rmod):<14}"
        for cmod in ranked:
            if rmod == cmod:
                row += f"{'—':>13}"
                continue
            rw, cw = win[(rmod, cmod)], win[(cmod, rmod)]
            tot = rw + cw
            rate = rw / tot if tot else float("nan")
            row += f"{rate:>7.2f}({rw:>2})"
        print(row)
    print()

    print("Robust wins (won BOTH orderings of a brand) per unordered pair:")
    for x, y in combinations(ranked, 2):
        fset = frozenset((x, y))
        rob = Counter()
        split = 0
        for (bid, fs), winners in by_bp.items():
            if fs != fset:
                continue
            uniq = set(winners)
            if len(uniq) == 1 and "tie" not in uniq:
                rob[next(iter(uniq))] += 1
            else:
                split += 1
        n_brands = sum(1 for (bid, fs) in by_bp if fs == fset)
        print(f"  {label(x)} vs {label(y)}: "
              f"{label(x)} {rob[x]}, {label(y)} {rob[y]}, split {split}  (of {n_brands} brands)")
    print()

    print("Meanest-editor verdict distribution:")
    for m in ranked:
        dist = ", ".join(f"{k}={v}" for k, v in sorted(verdicts[m].items()))
        print(f"  {label(m):<14}: {dist or '(none)'}")
    print()

    print("Position-bias diagnostic:")
    if decisive:
        print(f"  slot-A won {slot_a_wins}/{decisive} decisive judgments = "
              f"{slot_a_wins/decisive:.3f}  (0.50 = unbiased)")
    print()

    # ---- machine-readable summary for the figures ----
    summary = {
        "run": data.get("run"),
        "judge_model": data.get("judge_model"),
        "n_judgments": len(records),
        "dims": DIMS,
        "models": [
            {"slug": m, "label": label(m),
             "dim_means": {d: round(dim_mean(m, d), 4) for d in DIMS},
             "overall": round(overall_mean(m), 4),
             "verdicts": dict(verdicts[m])}
            for m in ranked
        ],
        "win_matrix": [
            {"row": r, "col": c,
             "row_wins": win[(r, c)], "col_wins": win[(c, r)],
             "ties": ties[frozenset((r, c))] if r != c else 0,
             "n": appear[frozenset((r, c))] if r != c else 0,
             "row_winrate_decisive": (
                 round(win[(r, c)] / (win[(r, c)] + win[(c, r)]), 4)
                 if (win[(r, c)] + win[(c, r)]) else None)}
            for r in ranked for c in ranked if r != c
        ],
        "position_bias_slotA": round(slot_a_wins / decisive, 4) if decisive else None,
    }
    if summary_path:
        with open(summary_path, "w") as f:
            json.dump(summary, f, indent=2)
        print(f"Wrote {summary_path}")
    return summary


if __name__ == "__main__":
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    if not args:
        print(__doc__)
        sys.exit(1)
    sp = None
    if "--summary" in sys.argv:
        i = sys.argv.index("--summary")
        sp = sys.argv[i + 1]
    main(args[0], sp)
