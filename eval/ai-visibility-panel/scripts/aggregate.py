#!/usr/bin/env python3
"""Aggregate deterministic grades and blind Opus 5 judge verdicts.

Usage:
    python3 aggregate.py RUN_DIR [--grading FILE] [--out FILE]

The script never re-grades artifacts. It summarizes the immutable grading.json
and per-case judgments, preserving deterministic and semantic results
separately so a strong judge score cannot hide a contract failure.
"""

from __future__ import annotations

import argparse
import collections
import json
import statistics
import sys
from pathlib import Path
from typing import Any

sys.path.insert(0, str(Path(__file__).resolve().parent))
import _common as common  # noqa: E402


DIMENSIONS = (
    "evidence_quality",
    "icp_market_model",
    "job_intent_architecture",
    "prompt_coverage_traceability",
    "prompt_realism_cell_discipline",
    "contamination_qa",
    "panel_measurement_design",
    "artifact_usability_integrity",
)


def mean(values: list[float]) -> float | None:
    return round(statistics.fmean(values), 3) if values else None


def load_verdict(run_dir: Path, slug: str, embedded: Any) -> dict[str, Any] | None:
    path = run_dir / "judgments" / slug / "verdict.json"
    if path.is_file():
        value = common.load_json(path)
        return value if isinstance(value, dict) else None
    return embedded if isinstance(embedded, dict) else None


def summarize(records: list[dict[str, Any]]) -> dict[str, Any]:
    deterministic_pass = sum(bool(record.get("deterministic_passed")) for record in records)
    verdicts = collections.Counter(
        str(record["judge"].get("verdict"))
        for record in records
        if isinstance(record.get("judge"), dict)
    )
    dims = {
        dimension: [
            float(record["judge"]["scores"][dimension])
            for record in records
            if isinstance(record.get("judge"), dict)
            and isinstance(record["judge"].get("scores"), dict)
            and isinstance(record["judge"]["scores"].get(dimension), (int, float))
        ]
        for dimension in DIMENSIONS
    }
    overall = [
        float(record["judge"]["overall_score"])
        for record in records
        if isinstance(record.get("judge"), dict)
        and isinstance(record["judge"].get("overall_score"), (int, float))
    ]
    gate_counts: dict[str, collections.Counter[str]] = collections.defaultdict(collections.Counter)
    must_haves = collections.Counter()
    anti_patterns = collections.Counter()
    critical_failures: list[dict[str, str]] = []
    recommended = collections.Counter()
    for record in records:
        judge = record.get("judge")
        if not isinstance(judge, dict):
            continue
        for gate, result in (judge.get("hard_gates") or {}).items():
            if isinstance(result, dict):
                gate_counts[gate][str(result.get("status"))] += 1
        for result in judge.get("must_have_results") or []:
            if isinstance(result, dict):
                must_haves[str(result.get("status"))] += 1
        for result in judge.get("anti_pattern_results") or []:
            if isinstance(result, dict):
                anti_patterns["triggered" if result.get("triggered") else "not_triggered"] += 1
        for failure in judge.get("critical_failures") or []:
            critical_failures.append({"case_id": record["case_id"], "failure": str(failure)})
        for change in judge.get("recommended_changes") or []:
            if isinstance(change, dict):
                recommended[(str(change.get("owner")), str(change.get("change")))] += 1

    judge_count = sum(verdicts.values())
    joint_pass = sum(
        bool(record.get("deterministic_passed"))
        and isinstance(record.get("judge"), dict)
        and record["judge"].get("verdict") == "pass"
        for record in records
    )
    return {
        "cases": len(records),
        "deterministic": {
            "passed": deterministic_pass,
            "failed": len(records) - deterministic_pass,
            "pass_rate": round(deterministic_pass / len(records), 3) if records else None,
        },
        "judge": {
            "completed": judge_count,
            "verdicts": dict(verdicts),
            "pass_rate": round(verdicts["pass"] / judge_count, 3) if judge_count else None,
            "overall_mean": mean(overall),
            "dimension_means": {key: mean(values) for key, values in dims.items()},
            "hard_gates": {key: dict(value) for key, value in sorted(gate_counts.items())},
            "must_have_statuses": dict(must_haves),
            "anti_patterns": dict(anti_patterns),
            "critical_failures": critical_failures,
        },
        "joint_pass": {
            "count": joint_pass,
            "rate": round(joint_pass / len(records), 3) if records else None,
            "definition": "deterministic contract pass AND semantic judge verdict pass",
        },
        "top_recommended_changes": [
            {"owner": owner, "change": change, "count": count}
            for (owner, change), count in recommended.most_common(20)
        ],
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("run_dir")
    parser.add_argument("--grading")
    parser.add_argument("--out")
    args = parser.parse_args()

    run_dir = Path(args.run_dir).resolve()
    grading_path = Path(args.grading).resolve() if args.grading else run_dir / "grading.json"
    out_path = Path(args.out).resolve() if args.out else run_dir / "summary.json"
    grading = common.load_json(grading_path)
    if not isinstance(grading, dict) or not isinstance(grading.get("cases"), list):
        raise SystemExit(f"{grading_path}: expected grading object with cases array")

    records: list[dict[str, Any]] = []
    for result in grading["cases"]:
        if not isinstance(result, dict):
            continue
        records.append({
            "case_id": result.get("case_id"),
            "case_name": result.get("case_name"),
            "case_slug": result.get("case_slug"),
            "split": result.get("split"),
            "tags": result.get("tags", []),
            "deterministic_passed": bool(result.get("passed")),
            "failed_checks": [
                check.get("id")
                for check in result.get("checks", [])
                if isinstance(check, dict)
                and not check.get("passed")
                and check.get("severity") in ("error", "critical")
            ],
            "counts": result.get("counts", {}),
            "judge": load_verdict(
                run_dir,
                str(result.get("case_slug")),
                result.get("judge_verdict"),
            ),
        })

    slices: dict[str, list[dict[str, Any]]] = {"all": records}
    for record in records:
        if record.get("split"):
            slices.setdefault(f"split={record['split']}", []).append(record)
        for tag in record.get("tags") or []:
            slices.setdefault(f"tag={tag}", []).append(record)

    failed_checks = collections.Counter(
        ident for record in records for ident in record["failed_checks"]
    )
    output = {
        "schema_version": "1.0.0",
        "aggregated_at": common.utc_now(),
        "run": run_dir.name,
        "grading_sha256": common.sha256_file(grading_path),
        "summary": summarize(records),
        "slices": {
            name: summarize(slice_records)
            for name, slice_records in sorted(slices.items())
            if name != "all"
        },
        "top_deterministic_failures": failed_checks.most_common(20),
        "records": records,
    }
    common.atomic_write_json(out_path, output)

    summary = output["summary"]
    print("=== AI visibility panel eval aggregate ===")
    print(
        f"cases: {summary['cases']}  deterministic: "
        f"{summary['deterministic']['passed']}/{summary['cases']} pass"
    )
    print(
        f"judge verdicts: {summary['judge']['verdicts']}  "
        f"overall mean: {summary['judge']['overall_mean']}"
    )
    print(
        f"joint pass: {summary['joint_pass']['count']}/{summary['cases']} "
        f"({summary['joint_pass']['rate']})"
    )
    for dimension, value in summary["judge"]["dimension_means"].items():
        print(f"  {dimension:<34} {value}")
    print(f"wrote {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
