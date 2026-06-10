#!/usr/bin/env python3
"""Aggregate Fable-5-vs-Opus-4.8 angle-study judgments.

Reads a results.json (one record per judgment, each carrying the A/B->model
mapping and the GPT-5.5 verdict) and re-anchors every judgment to model identity
so position bias cancels. Reports:

  - head-to-head win / tie / loss for Fable vs Opus
  - per-dimension mean scores per model and the Fable-minus-Opus delta
  - meanest-editor verdict distribution per model
  - a position-bias diagnostic: how often slot A won regardless of model

Usage:
    python3 aggregate.py runs/<run>/results.json
"""
import json
import sys
from collections import Counter, defaultdict

DIMS = ["news_value", "distinctness", "journalist_shape", "grounding",
        "anti_slop", "proof_rigor", "usefulness"]


def main(path):
    with open(path) as f:
        data = json.load(f)
    records = data["records"]

    # Head-to-head, re-anchored to model identity.
    wins = Counter()          # model -> wins
    ties = 0
    dim_scores = defaultdict(lambda: defaultdict(list))   # model -> dim -> [scores]
    verdicts = defaultdict(Counter)                       # model -> verdict -> n
    slot_a_wins = 0
    decisive = 0              # non-tie judgments
    n = len(records)

    for r in records:
        a_model, b_model = r["A_model"], r["B_model"]
        v = r["verdict"]
        # scores -> models
        for dim in DIMS:
            dim_scores[a_model][dim].append(v["scores"]["A"][dim])
            dim_scores[b_model][dim].append(v["scores"]["B"][dim])
        verdicts[a_model][v["verdict_A"]] += 1
        verdicts[b_model][v["verdict_B"]] += 1
        # winner -> model
        w = v["winner"]
        if w == "tie":
            ties += 1
        else:
            decisive += 1
            winner_model = a_model if w == "A" else b_model
            wins[winner_model] += 1
            if w == "A":
                slot_a_wins += 1

    print(f"\n=== Fable 5 vs Opus 4.8 — angle study ({data.get('run','?')}) ===")
    print(f"judge: {data.get('judge_model','?')}  |  judgments: {n}  "
          f"(brands x 2 orderings)\n")

    print("Head-to-head (re-anchored to model identity):")
    print(f"  Fable 5 wins : {wins['fable']}")
    print(f"  Opus 4.8 wins: {wins['opus']}")
    print(f"  Ties         : {ties}")
    fable_wt = wins['fable'] + ties
    if n:
        print(f"  Fable win-or-tie rate: {fable_wt}/{n} = {fable_wt/n:.3f}")
    print()

    print("Per-dimension mean (1-5) and Fable - Opus delta:")
    print(f"  {'dimension':<18} {'fable':>7} {'opus':>7} {'delta':>7}")
    for dim in DIMS:
        fs = dim_scores['fable'][dim]
        os_ = dim_scores['opus'][dim]
        fm = sum(fs) / len(fs) if fs else float('nan')
        om = sum(os_) / len(os_) if os_ else float('nan')
        print(f"  {dim:<18} {fm:>7.2f} {om:>7.2f} {fm-om:>+7.2f}")
    # overall mean
    fall = [s for d in DIMS for s in dim_scores['fable'][d]]
    oall = [s for d in DIMS for s in dim_scores['opus'][d]]
    if fall and oall:
        fm, om = sum(fall)/len(fall), sum(oall)/len(oall)
        print(f"  {'OVERALL':<18} {fm:>7.2f} {om:>7.2f} {fm-om:>+7.2f}")
    print()

    print("Meanest-editor verdict distribution:")
    for model in ("fable", "opus"):
        dist = ", ".join(f"{k}={v}" for k, v in sorted(verdicts[model].items()))
        print(f"  {model:<6}: {dist or '(none)'}")
    print()

    print("Position-bias diagnostic (the reason we run both orderings):")
    if decisive:
        print(f"  slot-A won {slot_a_wins}/{decisive} decisive judgments = "
              f"{slot_a_wins/decisive:.3f}  (0.50 = unbiased)")
    else:
        print("  no decisive judgments")
    print()


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(1)
    main(sys.argv[1])
