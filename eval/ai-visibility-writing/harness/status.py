#!/usr/bin/env python3
"""Report elapsed time, spend, capacity, score history, and projected burn."""

from __future__ import annotations

import json
import math
import sys
from datetime import datetime
from pathlib import Path


EVAL = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EVAL / "scripts"))
from pilotlib import PHASE_LIMITS, PRECALIBRATION_CEILINGS, observed_p95, read_jsonl, verify_ledger  # noqa: E402


def main() -> int:
    ledger = EVAL / "runs" / "cost-ledger.jsonl"
    total, phases = verify_ledger(ledger)
    invocations = read_jsonl(EVAL / "runs" / "model-invocations.jsonl")
    scores = read_jsonl(EVAL / "runs" / "score-history.jsonl")
    started = None
    if invocations:
        started = invocations[0].get("started_at")
    elif ledger.exists() and read_jsonl(ledger):
        started = read_jsonl(ledger)[0].get("recorded_at")
    elapsed = None
    if started:
        elapsed = (datetime.now().astimezone() - datetime.fromisoformat(started.replace("Z", "+00:00"))).total_seconds() / 3600
    platform = {}
    for name in ("google", "chatgpt"):
        p95 = observed_p95(ledger, name) or PRECALIBRATION_CEILINGS[name]
        platform[name] = {"p95_or_ceiling_usd": p95, "remaining_calls_at_total_cap": max(0, math.floor((10-total)/p95))}
    gains = []
    for previous, current in zip(scores, scores[1:]):
        gains.append(float(current.get("score", 0)) - float(previous.get("score", 0)))
    report = {
        "wall_clock_hours": elapsed, "wall_clock_limit": 10,
        "dataforseo_total_usd": total, "dataforseo_limit_usd": 10,
        "phases": {phase: {"actual_usd": phases[phase], "limit_usd": limit, "remaining_usd": limit-phases[phase]} for phase, limit in PHASE_LIMITS.items()},
        "platform_projection": platform,
        "model_invocations": len(invocations), "model_invocation_cap": 450,
        "invocations_by_executor": {name: sum(row.get("executor") == name for row in invocations) for name in ("codex", "claude")},
        "score_history": scores, "cycle_gains": gains,
    }
    print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
