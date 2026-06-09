#!/usr/bin/env python3
"""Aggregate blind pairwise-judge records for the pr-strategist eval.

Input: a results.json that is a list of judgment records, each:
  {
    "case_id": int, "case_name": str, "split": "train|holdout",
    "model": str,            # model that generated the CANDIDATE
    "ordering": "cand_first|gold_first",
    "candidate_label": "A|B",  # which label the candidate was shown as
    "judge": { ...the judge JSON... }
  }

Everything is re-anchored to the candidate (our skill) vs the gold (expert).
Reports, overall and split/model-sliced:
  - candidate win / tie / loss rate vs gold (the headline parity metric)
  - mean dimension scores cand vs gold + deltas (where we're systematically weak)
  - judge identification accuracy ("which is AI") vs 50% chance — the
    "can the judge tell us apart" metric the goal targets
  - position-bias sanity (A-wins vs B-wins across orderings)
  - the most common candidate gaps (actionable iteration signal)

Usage: python3 aggregate.py results.json [--out summary.json]
"""
import argparse, json, sys
from collections import Counter

DIMS = ["audience_goal", "positioning", "news_peg", "channel_cadence",
        "tactics_quality", "judgment_refusals", "fit_actionability"]


def other(lbl):
    return "B" if lbl == "A" else "A"


def summarize(records):
    n = len(records)
    cand_win = cand_tie = cand_loss = 0
    id_correct = id_wrong = id_unsure = 0
    a_wins = b_wins = ties = 0
    cand_dim = {d: [] for d in DIMS}
    gold_dim = {d: [] for d in DIMS}
    gaps = Counter()

    for r in records:
        j = r["judge"]
        cl = r["candidate_label"]
        gl = other(cl)
        w = j.get("winner")
        if w == "tie":
            cand_tie += 1; ties += 1
        elif w == cl:
            cand_win += 1
        elif w == gl:
            cand_loss += 1
        if w == "A":
            a_wins += 1
        elif w == "B":
            b_wins += 1

        wai = j.get("which_is_ai")
        if wai == "unsure":
            id_unsure += 1
        elif wai == cl:
            id_correct += 1
        elif wai == gl:
            id_wrong += 1

        sc = j.get("scores", {})
        for d in DIMS:
            if cl in sc and d in sc[cl]:
                cand_dim[d].append(sc[cl][d])
            if gl in sc and d in sc[gl]:
                gold_dim[d].append(sc[gl][d])

        for g in j.get("gaps_in_" + cl, []) or []:
            gaps[g.strip()] += 1

    def mean(xs):
        return round(sum(xs) / len(xs), 3) if xs else None

    cand_means = {d: mean(cand_dim[d]) for d in DIMS}
    gold_means = {d: mean(gold_dim[d]) for d in DIMS}
    deltas = {d: (round(cand_means[d] - gold_means[d], 3)
                  if cand_means[d] is not None and gold_means[d] is not None else None)
              for d in DIMS}
    id_decided = id_correct + id_wrong
    return {
        "n": n,
        "candidate_vs_gold": {
            "win": cand_win, "tie": cand_tie, "loss": cand_loss,
            "win_rate": round(cand_win / n, 3) if n else None,
            "tie_rate": round(cand_tie / n, 3) if n else None,
            "loss_rate": round(cand_loss / n, 3) if n else None,
            "tie_or_better_rate": round((cand_win + cand_tie) / n, 3) if n else None,
        },
        "dim_means_candidate": cand_means,
        "dim_means_gold": gold_means,
        "dim_deltas_cand_minus_gold": deltas,
        "mean_total_candidate": mean([v for v in cand_means.values() if v is not None]) if cand_means else None,
        "mean_total_gold": mean([v for v in gold_means.values() if v is not None]) if gold_means else None,
        "judge_identification": {
            "correct": id_correct, "wrong": id_wrong, "unsure": id_unsure,
            "accuracy_of_decided": round(id_correct / id_decided, 3) if id_decided else None,
            "note": "0.5 = judge cannot tell candidate from expert (goal). >0.5 = candidate is identifiable as the AI.",
        },
        "position_bias": {"A_wins": a_wins, "B_wins": b_wins, "ties": ties},
        "top_candidate_gaps": gaps.most_common(15),
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("results")
    ap.add_argument("--out", default=None)
    args = ap.parse_args()
    records = json.load(open(args.results))
    if isinstance(records, dict) and "records" in records:
        records = records["records"]

    slices = {"ALL": records}
    for r in records:
        slices.setdefault("split=" + r.get("split", "?"), []).append(r)
        slices.setdefault("model=" + r.get("model", "?"), []).append(r)

    out = {name: summarize(recs) for name, recs in slices.items()}
    out_path = args.out or args.results.replace(".json", ".summary.json")
    json.dump(out, open(out_path, "w"), indent=2)

    a = out["ALL"]
    cg = a["candidate_vs_gold"]
    print("\n=== pr-strategist pairwise eval ===")
    print(f"records: {a['n']}")
    print(f"candidate vs gold:  win {cg['win']}  tie {cg['tie']}  loss {cg['loss']}  "
          f"(tie-or-better {cg['tie_or_better_rate']})")
    print(f"judge identification accuracy (0.5=indistinguishable): "
          f"{a['judge_identification']['accuracy_of_decided']} "
          f"(unsure={a['judge_identification']['unsure']})")
    print(f"mean total  candidate {a['mean_total_candidate']}  vs  gold {a['mean_total_gold']}")
    print("dim deltas (cand - gold), negative = we're weaker:")
    for d in DIMS:
        print(f"   {d:<20} {a['dim_deltas_cand_minus_gold'][d]}")
    for name in sorted(slices):
        if name == "ALL":
            continue
        s = out[name]; c = s["candidate_vs_gold"]
        print(f"  [{name}]  win {c['win']} tie {c['tie']} loss {c['loss']}  "
              f"tie+={c['tie_or_better_rate']}  id-acc={s['judge_identification']['accuracy_of_decided']}")
    print("\ntop candidate gaps:")
    for g, k in a["top_candidate_gaps"]:
        print(f"   {k:>2}x  {g}")
    print(f"\nwrote {out_path}")


if __name__ == "__main__":
    sys.exit(main())
