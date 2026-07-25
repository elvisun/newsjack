#!/usr/bin/env python3
"""Blind, counterbalanced pairwise quality judge."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

from model_runner import EVAL, run_structured


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--case-id", required=True)
    parser.add_argument("--executor", choices=["codex", "claude"], required=True)
    parser.add_argument("--generator", choices=["codex", "claude"], required=True)
    parser.add_argument("--a", type=Path, required=True)
    parser.add_argument("--b", type=Path, required=True)
    parser.add_argument("--ordering", choices=["AB", "BA"], required=True)
    parser.add_argument("--model")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()
    if args.executor == args.generator:
        raise SystemExit("a provider may not judge its own generated pair")
    cases = json.loads((EVAL / "fixtures" / "cases.json").read_text(encoding="utf-8"))["cases"]
    case = next((item for item in cases if item["id"] == args.case_id), None)
    if not case:
        raise SystemExit("unknown case id")
    left = json.loads(args.a.read_text(encoding="utf-8"))["markdown"]
    right = json.loads(args.b.read_text(encoding="utf-8"))["markdown"]
    prompt = f"""You are a blind senior-editor judge in a clean context with no tools. Compare two anonymized outputs for the same constructed source. You do not know the treatments. Judge usefulness to the stated audience, specificity, evidentiary rigor, preservation of meaning and voice, economy, and honesty about limitations. Longer is not better. A changed or invented number, date, entity, quotation, attribution, source relationship, or caveat is a severe failure. Return only schema-valid JSON.\n\nSOURCE FACTS\nDocument type: {case['document_type']}\nRequested behavior: {case['behavior']}\nTarget query: {case['target_query']}\nDraft:\n{case['draft']}\n\nOUTPUT A\n{left}\n\nOUTPUT B\n{right}\n"""
    default_model = "gpt-5.6-sol" if args.executor == "codex" else "claude-opus-4-8"
    value = run_structured(
        executor=args.executor, prompt=prompt, schema_path=EVAL / "harness" / "judge-schema.json",
        output=args.output, kind="judge", case_id=args.case_id,
        condition=f"blind:{args.ordering}", model=args.model or default_model, dry_run=args.dry_run,
    )
    if args.dry_run:
        print(json.dumps(value, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
