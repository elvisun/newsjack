#!/usr/bin/env python3
"""Resumable clean-context simulation driver with a frozen 436-call plan."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

from model_runner import CAP, EVAL, INVOCATIONS
from pilotlib import read_jsonl


GENERATORS = ("codex", "claude")
OTHER = {"codex": "claude", "claude": "codex"}
BASELINE_CALIBRATION_CASES = (0, 3, 6, 9, 12, 15)
POSITION_REPLICATIONS = (("codex", 0), ("claude", 6), ("codex", 12), ("claude", 18))
PLANNED_INVOCATIONS = {
    "base_generations": 144,
    "combined_generations": 12,
    "lever_generations": 48,
    "baseline_judgments": 12,
    "skill_judgments": 48,
    "lever_judgments": 48,
    "position_replications": 4,
    "lever_citation_trials": 96,
    "combined_citation_trials": 24,
}


def invoke(command: list[str], dry_run: bool) -> None:
    if dry_run:
        return
    subprocess.run(command, check=True, cwd=EVAL.parents[1])


def generation_command(case_id: str, generator: str, condition: str, output: Path, lever: str | None = None) -> list[str]:
    command = [sys.executable, str(EVAL / "scripts" / "generate.py"), "--case-id", case_id, "--executor", generator, "--condition", condition, "--output", str(output)]
    if lever:
        command += ["--lever-id", lever]
    return command


def judge_command(case_id: str, generator: str, left: Path, right: Path, ordering: str, output: Path) -> list[str]:
    return [sys.executable, str(EVAL / "scripts" / "judge.py"), "--case-id", case_id, "--executor", OTHER[generator], "--generator", generator, "--a", str(left), "--b", str(right), "--ordering", ordering, "--output", str(output)]


def original_file(run_dir: Path, case: dict) -> Path:
    path = run_dir / "originals" / f"{case['id']}.json"
    if not path.exists():
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps({"markdown": case["draft"]}, indent=2) + "\n", encoding="utf-8")
    return path


def stronger_baselines(run_dir: Path, cases: list[dict], dry_run: bool) -> dict[str, str]:
    if dry_run:
        return {generator: "editorial" for generator in GENERATORS}
    choices = {}
    for generator in GENERATORS:
        wins = {"bare": 0.0, "editorial": 0.0}
        for index in BASELINE_CALIBRATION_CASES:
            order = "AB" if index % 2 == 0 else "BA"
            path = run_dir / "judgments" / "baseline" / generator / f"{cases[index]['id']}.{order}.json"
            result = json.loads(path.read_text(encoding="utf-8"))
            if result["winner"] == "tie":
                wins["bare"] += 0.5
                wins["editorial"] += 0.5
            else:
                winner = "bare" if (order == "AB" and result["winner"] == "A") or (order == "BA" and result["winner"] == "B") else "editorial"
                wins[winner] += 1
        choices[generator] = "editorial" if wins["editorial"] >= wins["bare"] else "bare"
    return choices


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--run-dir", type=Path, default=EVAL / "runs" / "dev")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()
    planned = sum(PLANNED_INVOCATIONS.values())
    existing = len(read_jsonl(INVOCATIONS))
    if existing >= CAP and not args.dry_run:
        raise SystemExit(f"model invocation cap already reached: existing={existing}, cap={CAP}")
    cases = json.loads((EVAL / "fixtures" / "cases.json").read_text(encoding="utf-8"))["cases"]
    ablations = json.loads((EVAL / "fixtures" / "ablation-manifest.json").read_text(encoding="utf-8"))
    run_dir = args.run_dir.resolve()

    for generator in GENERATORS:
        for case in cases:
            for condition in ("bare", "editorial", "skill"):
                output = run_dir / "generations" / generator / case["id"] / f"{condition}.json"
                invoke(generation_command(case["id"], generator, condition, output), args.dry_run)

    for generator in GENERATORS:
        for index in BASELINE_CALIBRATION_CASES:
            case = cases[index]
            order = "AB" if index % 2 == 0 else "BA"
            bare = run_dir / "generations" / generator / case["id"] / "bare.json"
            editorial = run_dir / "generations" / generator / case["id"] / "editorial.json"
            left, right = (bare, editorial) if order == "AB" else (editorial, bare)
            output = run_dir / "judgments" / "baseline" / generator / f"{case['id']}.{order}.json"
            invoke(judge_command(case["id"], generator, left, right, order, output), args.dry_run)

    baselines = stronger_baselines(run_dir, cases, args.dry_run)
    for generator in GENERATORS:
        for index, case in enumerate(cases):
            order = "AB" if index % 2 == 0 else "BA"
            baseline = run_dir / "generations" / generator / case["id"] / f"{baselines[generator]}.json"
            skill = run_dir / "generations" / generator / case["id"] / "skill.json"
            left, right = (baseline, skill) if order == "AB" else (skill, baseline)
            output = run_dir / "judgments" / "skill" / generator / f"{case['id']}.{order}.json"
            invoke(judge_command(case["id"], generator, left, right, order, output), args.dry_run)

    by_id = {case["id"]: case for case in cases}
    for lever, case_ids in ablations["levers"].items():
        for index, case_id in enumerate(case_ids):
            case = by_id[case_id]
            generator = GENERATORS[index % 2]
            rewrite = run_dir / "generations" / generator / case_id / "levers" / f"{lever}.json"
            invoke(generation_command(case_id, generator, "lever", rewrite, lever), args.dry_run)
            original = original_file(run_dir, case) if not args.dry_run else run_dir / "originals" / f"{case_id}.json"
            order = "AB" if index % 2 == 0 else "BA"
            left, right = (original, rewrite) if order == "AB" else (rewrite, original)
            judgment = run_dir / "judgments" / "levers" / lever / generator / f"{case_id}.{order}.json"
            invoke(judge_command(case_id, generator, left, right, order, judgment), args.dry_run)
            for variant in ("original", "rewrite"):
                output = run_dir / "citations" / lever / generator / case_id / f"{variant}.json"
                meta = output.with_suffix(".meta.json")
                command = [sys.executable, str(EVAL / "scripts" / "citation_simulate.py"), "--case-id", case_id, "--lever-id", lever, "--variant", variant, "--executor", OTHER[generator], "--generator", generator, "--output", str(output), "--metadata", str(meta)]
                if variant == "rewrite":
                    command += ["--rewrite", str(rewrite)]
                invoke(command, args.dry_run)

    for index, case_id in enumerate(ablations["combined_skill_cases"]):
        generator = GENERATORS[index % 2]
        rewrite = run_dir / "generations" / generator / case_id / "skill_rewrite.json"
        validation = rewrite.with_name("skill_rewrite.validation.json")
        if rewrite.exists() and validation.exists() and not json.loads(validation.read_text(encoding="utf-8")).get("pass"):
            continue
        invoke(generation_command(case_id, generator, "skill_rewrite", rewrite), args.dry_run)
        for variant in ("original", "rewrite"):
            output = run_dir / "citations" / "COMBINED" / generator / case_id / f"{variant}.json"
            meta = output.with_suffix(".meta.json")
            command = [sys.executable, str(EVAL / "scripts" / "citation_simulate.py"), "--case-id", case_id, "--lever-id", "COMBINED", "--variant", variant, "--executor", OTHER[generator], "--generator", generator, "--output", str(output), "--metadata", str(meta)]
            if variant == "rewrite":
                command += ["--rewrite", str(rewrite)]
            invoke(command, args.dry_run)

    for generator, index in POSITION_REPLICATIONS:
        case = cases[index]
        original_order = "AB" if index % 2 == 0 else "BA"
        reverse = "BA" if original_order == "AB" else "AB"
        baseline = run_dir / "generations" / generator / case["id"] / f"{baselines[generator]}.json"
        skill = run_dir / "generations" / generator / case["id"] / "skill.json"
        left, right = (baseline, skill) if reverse == "AB" else (skill, baseline)
        output = run_dir / "judgments" / "position" / generator / f"{case['id']}.{reverse}.json"
        invoke(judge_command(case["id"], generator, left, right, reverse, output), args.dry_run)

    print(json.dumps({"planned_invocations": planned, "existing_invocations": existing, "capacity": CAP, "fresh_plan_buffer": CAP - planned, "components": PLANNED_INVOCATIONS, "baseline_policy": "per-generator winner across six blind calibration pairs; ties select editorial", "run_dir": str(run_dir), "dry_run": args.dry_run}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
