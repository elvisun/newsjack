#!/usr/bin/env python3
"""Constraint linter. Public output is deliberately non-diagnostic on failure."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from pathlib import Path


EVAL = Path(__file__).resolve().parents[1]
REPO = EVAL.parents[1]
SKILL = REPO / "skills" / "ai-visibility-writing" / "SKILL.md"
sys.path.insert(0, str(EVAL / "scripts"))
from pilotlib import PilotError, read_jsonl, terminal_events, unresolved_unknown_tags, verify_ledger  # noqa: E402
from collect import response_cost  # noqa: E402


def details_path() -> Path:
    return Path(os.environ.get("AI_VISIBILITY_LINT_LOG", "/tmp/ai-visibility-writing-lint.log"))


def add(condition: bool, message: str, errors: list[str]) -> None:
    if condition:
        errors.append(message)


def git_show(commit: str, relative: str) -> bytes | None:
    result = subprocess.run(["git", "show", f"{commit}:{relative}"], cwd=REPO, capture_output=True)
    return result.stdout if result.returncode == 0 else None


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--plant", choices=["capacity", "claim", "credential", "ledger", "checksum", "corpus"])
    args = parser.parse_args()
    errors: list[str] = []
    try:
        skill_text = SKILL.read_text(encoding="utf-8")
        lines = skill_text.splitlines()
        add(len(lines) > 500, "skill line capacity", errors)
        skill_files = [path for path in SKILL.parent.rglob("*") if path.is_file()]
        add(len([path for path in skill_files if path.name != "SKILL.md"]) > 2, "skill reference-file capacity", errors)
        factors = json.loads((EVAL / "factors.json").read_text(encoding="utf-8"))["factors"]
        add(len(factors) > 12, "factor capacity", errors)
        add(skill_text.count("**Thin release:**") + skill_text.count("**Technical blog:**") + skill_text.count("**Freshness-sensitive article:**") > 6, "example capacity", errors)
        positive_claims = [
            r"\b(?:will|does|can) guarantee\b.{0,30}\b(?:rank|citation|traffic|mention)",
            r"\bguaranteed\s+(?:rank|citation|traffic|visibility|coverage)",
            r"\b(?:ensure|secure)\s+(?:a\s+)?(?:ranking|citation|mention)",
        ]
        add(any(re.search(pattern, skill_text, re.I | re.S) for pattern in positive_claims), "forbidden guarantee", errors)
        active_blackhat = [r"recommend\b.{0,40}\bkeyword stuffing", r"add\s+hidden\s+text", r"fabricate\w*\s+(?:FAQ|citation|source)"]
        add(any(re.search(pattern, skill_text, re.I | re.S) for pattern in active_blackhat), "black-hat advice", errors)
        eval_literals = ["audit_factor_F1", "blind_quality_win_rate", "paired_citation_selection_win_rate", "readability_pass_rate", "efficiency_index", "pilot signal bar", "constraint violation"]
        add(any(value.lower() in skill_text.lower() for value in eval_literals), "evaluation-language leakage", errors)
        longest_list = current_list = 0
        for line in lines:
            if re.match(r"^\s*(?:[-*]|\d+\.)\s+", line):
                current_list += 1
                longest_list = max(longest_list, current_list)
            elif line.strip() and not line.startswith("  "):
                current_list = 0
        add(longest_list > 20, "keyword/list capacity", errors)
        all_paths = [path for path in (list((REPO / "skills" / "ai-visibility-writing").rglob("*")) + list(EVAL.rglob("*"))) if path.is_file() and ".git" not in path.parts]
        credential_patterns = [re.compile(rb"Authorization:\s*Basic\s+[A-Za-z0-9+/=]{12,}"), re.compile(rb"DATAFORSEO_(?:PASSWORD|LOGIN)\s*=\s*[^\s$<{][^\s]*")]
        for path in all_paths:
            payload = path.read_bytes()
            if any(pattern.search(payload) for pattern in credential_patterns):
                errors.append("credential-like literal")
                break
        ledger = EVAL / "runs" / "cost-ledger.jsonl"
        if ledger.exists():
            verify_ledger(ledger)
            ledger_rows = read_jsonl(ledger)
            add(bool(unresolved_unknown_tags(ledger)), "unknown request cost", errors)
            terminal_by_tag = terminal_events(ledger)
            observations = {}
            for observation_file in (EVAL / "runs").glob("*/observations.jsonl"):
                for row in read_jsonl(observation_file):
                    observations[row.get("request_tag")] = row
            for tag, event in terminal_by_tag.items():
                raw_paths = list((EVAL / "runs").glob(f"*/raw/{tag}.json"))
                if event.get("status") == "success":
                    add(len(raw_paths) != 1, "successful request raw reconciliation", errors)
                    add(tag not in observations, "successful request observation reconciliation", errors)
                if len(raw_paths) == 1:
                    raw = json.loads(raw_paths[0].read_text(encoding="utf-8"))
                    add(event.get("response_sha256") != hashlib.sha256(raw_paths[0].read_bytes()).hexdigest(), "raw response hash mismatch", errors)
                    raw_cost = response_cost(raw)
                    add(raw_cost is None or abs(float(event.get("cost_usd")) - raw_cost) > 1e-9, "raw response cost mismatch", errors)
                if tag in observations:
                    add(abs(float(event.get("cost_usd")) - float(observations[tag].get("cost_usd"))) > 1e-9, "observation cost mismatch", errors)
            manifests = []
            for manifest_path in (EVAL / "manifests").glob("*.json") if (EVAL / "manifests").exists() else []:
                manifests.append(json.loads(manifest_path.read_text(encoding="utf-8")))
            for manifest in manifests:
                expected = {descriptor["request_tag"] for unit in manifest.get("units", []) for descriptor in unit.get("requests", [])}
                phase_events = {tag for tag, row in terminal_by_tag.items() if row.get("phase") == manifest.get("phase") and not row.get("retry_of")}
                if phase_events:
                    add(phase_events != expected, "started manifest is incomplete or has unexpected tags", errors)
        corpus_literals = []
        for manifest in (EVAL / "manifests").glob("*.json") if (EVAL / "manifests").exists() else []:
            data = json.loads(manifest.read_text(encoding="utf-8"))
            corpus_literals.extend(row.get("query", "") for row in data.get("units", []))
        for observation_file in (EVAL / "runs").glob("*/observations.jsonl") if (EVAL / "runs").exists() else []:
            for row in read_jsonl(observation_file):
                corpus_literals.extend(c.get("canonical_url", "") for c in row.get("citations", []))
        add(any(len(value) >= 20 and value in skill_text for value in corpus_literals if value), "corpus literal leakage", errors)
        frozen_commit_file = EVAL / "FROZEN_COMMIT"
        if frozen_commit_file.exists():
            commit = frozen_commit_file.read_text(encoding="utf-8").strip()
            frozen = json.loads((EVAL / "frozen-checksums.json").read_text(encoding="utf-8"))
            for relative, expected in frozen["files"].items():
                current = (REPO / relative).read_bytes()
                add(hashlib.sha256(current).hexdigest() != expected, "frozen checksum mismatch", errors)
                baseline = git_show(commit, relative)
                add(baseline is None or baseline != current, "frozen commit mismatch", errors)
    except Exception as exc:
        errors.append(f"linter exception: {type(exc).__name__}")
    if args.plant:
        errors.append(f"deliberate {args.plant} detection plant")
    log = details_path()
    log.parent.mkdir(parents=True, exist_ok=True)
    log.write_text("\n".join(errors) + ("\n" if errors else "OK\n"), encoding="utf-8")
    if errors:
        print("VOID: constraint violation")
        return 3
    print("OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
