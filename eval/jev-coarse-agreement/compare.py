#!/usr/bin/env python3
"""Compare Jev coarse-filter decisions with LLM worker decisions.

usage: compare.py <runs/<timestamp>> [--json]

Reads the chunk/worker/jev triples staged by run.sh and prints agreement
rates plus every disagreement that matters for recall.
"""
from __future__ import annotations

import json
import sys
from collections import Counter
from pathlib import Path

SURVIVE = {"keep", "monitor_only"}


def load(path: Path):
    with path.open() as f:
        return json.load(f)


def decisions_by_id(payload) -> dict:
    items = payload.get("decisions", payload) if isinstance(payload, dict) else payload
    return {d["signal_id"]: d for d in items if isinstance(d, dict) and d.get("signal_id")}


def titles_by_id(chunk) -> dict:
    out = {}
    for s in chunk.get("signals", []):
        sid = s.get("signal_id") or s.get("id")
        if sid:
            out[sid] = s.get("title") or ""
    return out


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print(__doc__)
        return 2
    run_dir = Path(argv[1])
    as_json = "--json" in argv
    pairs = []
    for jev_path in sorted(run_dir.glob("jev.*.json")):
        key = jev_path.name[len("jev.") : -len(".json")]
        worker_path = run_dir / f"worker.{key}.json"
        chunk_path = run_dir / f"chunk.{key}.json"
        if not worker_path.exists() or not chunk_path.exists():
            continue
        pairs.append((key, load(chunk_path), load(worker_path), load(jev_path)))
    if not pairs:
        print(f"no jev/worker pairs found in {run_dir}")
        return 1

    n = 0
    exact = 0
    survive_agree = 0
    reason_agree = 0
    reason_n = 0
    recall_misses = []
    high_conf_misses = 0
    junk_through = []
    jev_counts: Counter = Counter()
    worker_counts: Counter = Counter()
    post_rules: Counter = Counter()
    cost = 0.0
    calls = 0
    failures = 0
    elapsed = 0

    for key, chunk, worker, jev in pairs:
        titles = titles_by_id(chunk)
        w = decisions_by_id(worker)
        j = decisions_by_id(jev)
        engine = jev.get("engine", {}) if isinstance(jev, dict) else {}
        cost += float(engine.get("est_cost_usd") or 0)
        calls += int(engine.get("calls") or 0)
        failures += int(engine.get("failures") or 0)
        elapsed += int(engine.get("elapsed_ms") or 0)
        for sid, wd in w.items():
            jd = j.get(sid)
            if not jd:
                continue
            n += 1
            wdec, jdec = wd.get("decision"), jd.get("decision")
            worker_counts[wdec] += 1
            jev_counts[jdec] += 1
            for r in jd.get("post_rules") or []:
                post_rules[r] += 1
            if wdec == jdec:
                exact += 1
                reason_n += 1
                if wd.get("reason") == jd.get("reason"):
                    reason_agree += 1
            if (wdec in SURVIVE) == (jdec in SURVIVE):
                survive_agree += 1
            elif wdec in SURVIVE and jdec == "reject":
                miss = {
                    "signal_id": sid,
                    "chunk": key,
                    "title": titles.get(sid, ""),
                    "worker": f"{wdec}/{wd.get('reason')} ({wd.get('confidence')})",
                    "worker_rationale": wd.get("rationale", ""),
                    "jev": f"{jdec}/{jd.get('reason')} ({jd.get('confidence')})",
                    "jev_rationale": jd.get("rationale", ""),
                }
                recall_misses.append(miss)
                if wdec == "keep" and wd.get("confidence") == "high":
                    high_conf_misses += 1
            else:
                junk_through.append(
                    {
                        "signal_id": sid,
                        "chunk": key,
                        "title": titles.get(sid, ""),
                        "worker": f"{wdec}/{wd.get('reason')}",
                        "jev": f"{jdec}/{jd.get('reason')} ({jd.get('confidence')})",
                    }
                )

    survive_rate = survive_agree / n if n else 0.0
    summary = {
        "signals": n,
        "decision_agreement": round(exact / n, 3) if n else None,
        "survive_vs_drop_agreement": round(survive_rate, 3) if n else None,
        "reason_agreement_when_decision_agrees": round(reason_agree / reason_n, 3) if reason_n else None,
        "recall_misses": len(recall_misses),
        "recall_misses_on_high_confidence_keeps": high_conf_misses,
        "junk_let_through": len(junk_through),
        "worker_decision_counts": dict(worker_counts),
        "jev_decision_counts": dict(jev_counts),
        "jev_post_rules": dict(post_rules),
        "jev_calls": calls,
        "jev_failures": failures,
        "jev_est_cost_usd": round(cost, 4),
        "jev_elapsed_ms": elapsed,
        "acceptance": {
            "zero_high_confidence_recall_misses": high_conf_misses == 0,
            "survive_agreement_at_least_0_85": survive_rate >= 0.85,
        },
    }
    if as_json:
        print(json.dumps({"summary": summary, "recall_misses": recall_misses, "junk_let_through": junk_through}, indent=2))
        return 0

    print("Jev vs worker coarse-filter agreement")
    for k, v in summary.items():
        print(f"  {k}: {v}")
    if recall_misses:
        print("\nRecall misses (worker kept, Jev rejected):")
        for m in recall_misses:
            print(f"  - [{m['chunk']}] {m['signal_id']}  {m['title'][:90]}")
            print(f"      worker: {m['worker']}  {m['worker_rationale'][:160]}")
            print(f"      jev:    {m['jev']}  {m['jev_rationale'][:160]}")
    if junk_through:
        print("\nJunk let through (worker rejected, Jev kept):")
        for m in junk_through:
            print(f"  - [{m['chunk']}] {m['signal_id']}  {m['title'][:90]}  worker={m['worker']}  jev={m['jev']}")
    ok = summary["acceptance"]
    print("\nAcceptance:", "PASS" if all(ok.values()) else "FAIL", ok)
    return 0 if all(ok.values()) else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
