#!/usr/bin/env python3
"""Compile quality, factor, citation, fidelity, and ten-lever simulation results."""

from __future__ import annotations

import argparse
import difflib
import json
import re
import subprocess
import sys
from collections import defaultdict
from pathlib import Path
from statistics import median

from model_runner import EVAL
from pilotlib import compare_fact_ledgers, composite, read_jsonl, verify_ledger, wilson


GENERATORS = ("codex", "claude")
FACTOR_PATTERNS = {
    "F1": r"\b(?:firsthand|original data|unique (?:evidence|information)|owned analysis|evidence page)\b",
    "F2": r"\b(?:direct|scoped|self-contained|opening) answer\b|\banswer.{0,35}\bheading\b",
    "F3": r"\b(?:attribution|primary source|evidence|method|population|source relationship)\b",
    "F4": r"\b(?:ambiguous entit|clarify (?:the )?(?:entity|acronym)|disambiguat|first reference)\b",
    "F5": r"\b(?:descriptive heading|coherent (?:chunk|section)|single-purpose|section structure)\b",
    "F6": r"\b(?:comparison table|step-by-step|intent-fit|structured format|real procedure)\b",
    "F7": r"\b(?:effective date|measurement date|publication date|update date|freshness|time-sensitive)\b",
    "F8": r"\b(?:promotional|superlative|absolute claim|qualified (?:prose|claim)|specificity)\b",
    "F9": r"\b(?:preserve (?:useful )?precision|intended register|clarify syntax|technical term)\b",
    "F10": r"\b(?:keyword repetition|checklist padding|redundant repetition|fake faq)\b",
}
LEVER_FACTOR = {"L1":"F2", "L2":"F3", "L3":"F1", "L4":"F5", "L5":"F8", "L6":"F10", "L7":"F4", "L8":"F6", "L9":"F7", "L10":"F9"}


def treatment_value(result: dict, order: str, treatment_position_in_ab: str = "B") -> float:
    treatment = treatment_position_in_ab if order == "AB" else ("A" if treatment_position_in_ab == "B" else "B")
    if result["winner"] == "tie":
        return 0.5
    return 1.0 if result["winner"] == treatment else 0.0


def treatment_field(result: dict, order: str, stem: str, treatment_position_in_ab: str = "B"):
    treatment = treatment_position_in_ab if order == "AB" else ("A" if treatment_position_in_ab == "B" else "B")
    return result[f"{stem}_{treatment}"]


def control_field(result: dict, order: str, stem: str, treatment_position_in_ab: str = "B"):
    treatment = treatment_position_in_ab if order == "AB" else ("A" if treatment_position_in_ab == "B" else "B")
    control = "B" if treatment == "A" else "A"
    return result[f"{stem}_{control}"]


def changed_fraction(original: str, rewrite: str) -> float:
    left = re.findall(r"\w+|[^\w\s]", original)
    right = re.findall(r"\w+|[^\w\s]", rewrite)
    matches = sum(block.size for block in difflib.SequenceMatcher(a=left, b=right, autojunk=False).get_matching_blocks())
    return 1 - (2 * matches / max(len(left) + len(right), 1))


def detect_factors(markdown: str) -> set[str]:
    section = re.split(r"^##\s+Blocking questions", markdown, flags=re.I | re.M)[0]
    if re.search(r"^##\s+Highest-leverage changes", section, re.I | re.M):
        section = re.split(r"^##\s+Highest-leverage changes", section, flags=re.I | re.M)[-1]
    return {factor for factor, pattern in FACTOR_PATTERNS.items() if re.search(pattern, section, re.I | re.S)}


def citation_pair(run_dir: Path, lever: str, generator: str, case_id: str, rewrite_text: str, original_text: str) -> dict:
    values = {}
    for variant in ("original", "rewrite"):
        path = run_dir / "citations" / lever / generator / case_id / f"{variant}.json"
        meta = json.loads(path.with_suffix(".meta.json").read_text(encoding="utf-8"))
        result = json.loads(path.read_text(encoding="utf-8"))
        target = meta["target_candidate_id"]
        values[variant] = {
            "cited": target in result["cited_candidate_ids"],
            "accurate": target in result["accurately_used_candidate_ids"] and not result["unsupported_claims"],
            "unsupported": bool(result["unsupported_claims"]),
            "null_case": meta["null_case"],
            "selector": meta["selector"],
        }
    rewrite_success = values["rewrite"]["cited"] and values["rewrite"]["accurate"] and not values["rewrite"]["unsupported"]
    original_success = values["original"]["cited"] and values["original"]["accurate"] and not values["original"]["unsupported"]
    return {
        "citation_selected_original": values["original"]["cited"],
        "citation_selected_rewrite": values["rewrite"]["cited"],
        "accurate_use_original": values["original"]["accurate"],
        "accurate_use_rewrite": values["rewrite"]["accurate"],
        "unsupported_original": values["original"]["unsupported"],
        "unsupported_rewrite": values["rewrite"]["unsupported"],
        "citation_win": 1.0 if rewrite_success and not original_success else (0.0 if original_success and not rewrite_success else 0.5),
        "changed_token_fraction": changed_fraction(original_text, rewrite_text),
        "null_case": values["original"]["null_case"],
        "selector": values["original"]["selector"],
    }


def interval(values: list[float]) -> dict:
    low, high = wilson(sum(values), len(values)) if values else (None, None)
    return {"estimate": sum(values) / len(values) if values else None, "low": low, "high": high, "n": len(values)}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--run-dir", type=Path, default=EVAL / "runs" / "dev")
    parser.add_argument("--analysis", type=Path)
    parser.add_argument("--mode", choices=["dev", "fresh"], default="dev")
    args = parser.parse_args()
    run_dir = args.run_dir.resolve()
    cases = json.loads((EVAL / "fixtures" / "cases.json").read_text(encoding="utf-8"))["cases"]
    by_id = {case["id"]: case for case in cases}
    ablations = json.loads((EVAL / "fixtures" / "ablation-manifest.json").read_text(encoding="utf-8"))

    factor_tp = factor_fp = factor_fn = 0
    quality_rows = []
    all_fidelity = []
    for generator in GENERATORS:
        for index, case in enumerate(cases):
            order = "AB" if index % 2 == 0 else "BA"
            judgment_path = run_dir / "judgments" / "skill" / generator / f"{case['id']}.{order}.json"
            judgment = json.loads(judgment_path.read_text(encoding="utf-8"))
            skill_output = json.loads((run_dir / "generations" / generator / case["id"] / "skill.json").read_text(encoding="utf-8"))
            predicted = detect_factors(skill_output["markdown"])
            expected = set(case["expected_factors"])
            factor_tp += len(predicted & expected)
            factor_fp += len(predicted - expected)
            factor_fn += len(expected - predicted)
            row = {
                "generator": generator, "case_id": case["id"], "document_type": case["document_type"],
                "topic_family": case["topic_family"], "win": treatment_value(judgment, order),
                "quality": treatment_field(judgment, order, "quality"),
                "quality_control": control_field(judgment, order, "quality"),
                "fidelity": treatment_field(judgment, order, "fact_fidelity"),
                "readability": treatment_field(judgment, order, "readability"),
            }
            all_fidelity.append(row["fidelity"])
            quality_rows.append(row)
    precision = factor_tp / max(factor_tp + factor_fp, 1)
    recall = factor_tp / max(factor_tp + factor_fn, 1)
    audit_f1 = 2 * precision * recall / max(precision + recall, 1e-12)

    simulation_rows = []
    lever_quality = {}
    for lever, case_ids in ablations["levers"].items():
        lever_quality[lever] = {}
        for index, case_id in enumerate(case_ids):
            generator = GENERATORS[index % 2]
            case = by_id[case_id]
            rewrite_path = run_dir / "generations" / generator / case_id / "levers" / f"{lever}.json"
            rewrite = json.loads(rewrite_path.read_text(encoding="utf-8"))["markdown"]
            order = "AB" if index % 2 == 0 else "BA"
            judge = json.loads((run_dir / "judgments" / "levers" / lever / generator / f"{case_id}.{order}.json").read_text(encoding="utf-8"))
            pair = citation_pair(run_dir, lever, generator, case_id, rewrite, case["draft"])
            pair.update({
                "lever_id": lever, "case_id": case_id, "generator": generator,
                "document_type": case["document_type"], "topic_family": case["topic_family"],
                "human_quality_delta": treatment_field(judge, order, "quality") - control_field(judge, order, "quality"),
                "fact_fidelity_delta": int(treatment_field(judge, order, "fact_fidelity")) - int(control_field(judge, order, "fact_fidelity")),
                "rewrite_fact_fidelity": treatment_field(judge, order, "fact_fidelity") and compare_fact_ledgers(case["draft"], rewrite)["pass"],
            })
            all_fidelity.append(pair["rewrite_fact_fidelity"])
            simulation_rows.append(pair)

    combined_rows = []
    combined_exclusions = []
    for index, case_id in enumerate(ablations["combined_skill_cases"]):
        generator = GENERATORS[index % 2]
        case = by_id[case_id]
        rewrite_path = run_dir / "generations" / generator / case_id / "skill_rewrite.json"
        validation_path = rewrite_path.with_name("skill_rewrite.validation.json")
        validation = json.loads(validation_path.read_text(encoding="utf-8")) if validation_path.exists() else {}
        if validation_path.exists() and not validation.get("pass"):
            combined_exclusions.append({"case_id": case_id, "generator": generator, "reason": "deterministic_fact_guard"})
            continue
        rewrite = json.loads(rewrite_path.read_text(encoding="utf-8"))["markdown"]
        pair = citation_pair(run_dir, "COMBINED", generator, case_id, rewrite, case["draft"])
        pair.update({"case_id": case_id, "generator": generator, "document_type": case["document_type"], "topic_family": case["topic_family"]})
        pair["rewrite_fact_fidelity"] = bool(validation.get("pass"))
        all_fidelity.append(pair["rewrite_fact_fidelity"])
        combined_rows.append(pair)

    analysis = json.loads(args.analysis.read_text(encoding="utf-8")) if args.analysis and args.analysis.exists() else {}
    adjusted = (((analysis.get("multivariable") or {}).get("google_adjusted") or {}).get("effects") or {})
    template = json.loads((EVAL / "levers.json").read_text(encoding="utf-8"))
    rows_by_lever = {lever: [row for row in simulation_rows if row["lever_id"] == lever] for lever in ablations["levers"]}
    for row in template["levers"]:
        lever = row["id"]
        factor = LEVER_FACTOR[lever]
        if factor in adjusted:
            row["google"] = {"estimate": adjusted[factor].get("estimate_pp_per_sd"), "n": ((analysis.get("multivariable") or {}).get("google_adjusted") or {}).get("n", 0), "interval": adjusted[factor].get("interval_95_pp")}
        if lever not in rows_by_lever:
            continue
        values = rows_by_lever[lever]
        citation_lift = 100 * sum(int(item["citation_selected_rewrite"]) - int(item["citation_selected_original"]) for item in values) / len(values)
        accurate_lift = 100 * sum(int(item["accurate_use_rewrite"]) - int(item["accurate_use_original"]) for item in values) / len(values)
        changed = median(item["changed_token_fraction"] for item in values)
        quality = sum(item["human_quality_delta"] for item in values) / len(values)
        fidelity = sum(item["fact_fidelity_delta"] for item in values) / len(values)
        by_generator = {}
        for generator in GENERATORS:
            selected = [item for item in values if item["generator"] == generator]
            by_generator[generator] = {
                "citation_lift_pp": 100 * sum(int(item["citation_selected_rewrite"]) - int(item["citation_selected_original"]) for item in selected) / len(selected),
                "accurate_use_lift_pp": 100 * sum(int(item["accurate_use_rewrite"]) - int(item["accurate_use_original"]) for item in selected) / len(selected),
            }
        consistency = all(value["citation_lift_pp"] >= 0 and value["accurate_use_lift_pp"] >= 0 for value in by_generator.values())
        row.update({
            "disposition": "promising" if citation_lift > 0 and accurate_lift > 0 and quality >= 0 and fidelity >= 0 and consistency else "insufficient",
            "citation_selection_lift_pp": citation_lift, "accurate_answer_use_lift_pp": accurate_lift,
            "human_quality_delta": quality, "fact_fidelity_delta": fidelity,
            "median_changed_token_fraction": changed,
            "executor_consistency": {"same_nonnegative_direction": consistency, **by_generator},
            "efficiency_index": ((citation_lift + accurate_lift) / 2) * 0.10 / max(changed, 0.10),
            "limitations": "Eight constructed post-retrieval pairs, including one null; this is not live retrieval or citation lift.",
        })
    tested = [row for row in template["levers"] if row["id"] in rows_by_lever]
    tested.sort(key=lambda row: row.get("efficiency_index") if isinstance(row.get("efficiency_index"), (int, float)) else float("-inf"), reverse=True)
    untested = [row for row in template["levers"] if row["id"] not in rows_by_lever]
    ordered = tested + untested
    for rank, row in enumerate(ordered, 1):
        row["rank"] = rank
    template["levers"] = ordered
    eligible_winners = [row for row in tested if row["disposition"] == "promising" and row["citation_selection_lift_pp"] > 0 and row["accurate_answer_use_lift_pp"] > 0 and row["human_quality_delta"] >= 0 and row["fact_fidelity_delta"] >= 0 and row["executor_consistency"]["same_nonnegative_direction"]]
    template["atomic_simulation_leader"] = eligible_winners[0]["id"] if eligible_winners else None
    template["winner"] = None
    template["winner_reason"] = "pending overall pilot gates"
    template["status"] = "simulation_compiled"

    quality_values = [row["win"] for row in quality_rows]
    readability_values = [1.0 if row["readability"] else 0.0 for row in quality_rows]
    citation_values = [row["citation_win"] for row in combined_rows]
    strata = {}
    for field in ("generator", "document_type", "topic_family"):
        strata[field] = {level: interval([row["win"] for row in quality_rows if row[field] == level]) for level in sorted({row[field] for row in quality_rows})}
    clear_regression = any(value["high"] is not None and value["high"] < 0.5 for groups in strata.values() for value in groups.values())
    stage0 = subprocess.run([sys.executable, "-m", "unittest", "discover", "-s", str(EVAL / "tests")], capture_output=True).returncode == 0
    lint = subprocess.run([str(EVAL / "harness" / "lint.sh")], capture_output=True).returncode == 0
    try:
        verify_ledger(EVAL / "runs" / "cost-ledger.jsonl")
        ledger_ok = True
    except Exception:
        ledger_ok = False
    components = {
        "audit_factor_f1": audit_f1,
        "blind_quality_win_rate": sum(quality_values) / len(quality_values),
        "paired_citation_selection_win_rate": sum(citation_values) / len(citation_values),
        "readability_pass_rate": sum(readability_values) / len(readability_values),
    }
    metrics = {
        "mode": args.mode, "components": components, "composite": composite(components),
        "combined_cases": {"planned": len(ablations["combined_skill_cases"]), "completed": len(combined_rows), "exclusions": combined_exclusions},
        "intervals": {"blind_quality": interval(quality_values), "paired_citation_selection": interval(citation_values), "readability": interval(readability_values)},
        "audit_counts": {"tp": factor_tp, "fp": factor_fp, "fn": factor_fn, "precision": precision, "recall": recall},
        "strata": strata,
        "hard_gates": {
            "zero_invented_facts": all(all_fidelity),
            "schema_complete": not combined_exclusions and len(combined_rows) == len(ablations["combined_skill_cases"]),
            "no_forbidden_claim": lint, "no_clear_stratum_regression": not clear_regression,
            "stage0_green": stage0, "ledger_reconciled": ledger_ok,
        },
        "causal_boundary": "Post-retrieval simulation only; no live crawl, index, rank, retrieval, citation, or traffic effect is measured.",
    }
    bars = {"audit_factor_f1": 0.70, "blind_quality_win_rate": 0.60, "paired_citation_selection_win_rate": 0.55, "readability_pass_rate": 0.90}
    metrics["pilot_gates_pass"] = bool(
        metrics["composite"] >= 65.0
        and all(components[key] >= value for key, value in bars.items())
        and all(metrics["hard_gates"].values())
    )
    if metrics["pilot_gates_pass"] and eligible_winners:
        template["winner"] = eligible_winners[0]["id"]
        template["winner_reason"] = "highest eligible post-retrieval efficiency index; not a live causal effect"
    else:
        template["winner_reason"] = "no supported efficiency winner; overall pilot gates did not pass"
    run_dir.mkdir(parents=True, exist_ok=True)
    (run_dir / "citation-simulations.jsonl").write_text("".join(json.dumps(row, sort_keys=True) + "\n" for row in simulation_rows), encoding="utf-8")
    (run_dir / "combined-citation-simulations.jsonl").write_text("".join(json.dumps(row, sort_keys=True) + "\n" for row in combined_rows), encoding="utf-8")
    (run_dir / "levers.json").write_text(json.dumps(template, indent=2) + "\n", encoding="utf-8")
    (run_dir / "metrics.json").write_text(json.dumps(metrics, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"metrics": metrics, "winner": template["winner"], "winner_reason": template["winner_reason"]}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
