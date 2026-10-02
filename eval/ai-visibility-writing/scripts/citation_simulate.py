#!/usr/bin/env python3
"""Run one blinded original-or-rewrite post-retrieval source-selection trial."""

from __future__ import annotations

import argparse
import hashlib
import json
import random
import re
from pathlib import Path

from model_runner import EVAL, run_structured
from pilotlib import canonical_json, sha256_text


def _sentences(text: str) -> list[str]:
    values = [part.strip() for part in re.split(r"(?<=[.!?])\s+", text.strip()) if part.strip()]
    return values or [text.strip()]


def pack(case: dict, target_text: str, lever_id: str, null_case: bool) -> tuple[str, list[dict], str]:
    material = f"{lever_id}|{case['id']}|candidate-pack-v1"
    rng = random.Random(int(hashlib.sha256(material.encode()).hexdigest(), 16))
    ids = ["SOURCE-1", "SOURCE-2", "SOURCE-3", "SOURCE-4"]
    rng.shuffle(ids)
    target_id = ids[0]
    source_sentences = _sentences(case["draft"])
    candidates = [
        {"id": target_id, "passage": target_text},
        {"id": ids[1], "passage": source_sentences[0]},
        {"id": ids[2], "passage": source_sentences[-1]},
        {"id": ids[3], "passage": "This constructed candidate provides no additional supported fact for the question."},
    ]
    rng.shuffle(candidates)
    if null_case:
        cases = json.loads((EVAL / "fixtures" / "cases.json").read_text(encoding="utf-8"))["cases"]
        replacement = next(item for item in cases if item["topic_family"] != case["topic_family"])
        query = replacement["target_query"]
    else:
        query = case["target_query"]
    return query, candidates, target_id


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--case-id", required=True)
    parser.add_argument("--lever-id", choices=[f"L{i}" for i in range(1, 7)] + ["COMBINED"], required=True)
    parser.add_argument("--variant", choices=["original", "rewrite"], required=True)
    parser.add_argument("--rewrite", type=Path)
    parser.add_argument("--executor", choices=["codex", "claude"], required=True)
    parser.add_argument("--generator", choices=["codex", "claude"], required=True)
    parser.add_argument("--model")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--metadata", type=Path, required=True)
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()
    if args.executor == args.generator:
        raise SystemExit("a provider may not select citations for its own generated rewrite")
    cases = json.loads((EVAL / "fixtures" / "cases.json").read_text(encoding="utf-8"))["cases"]
    case = next((item for item in cases if item["id"] == args.case_id), None)
    manifest = json.loads((EVAL / "fixtures" / "ablation-manifest.json").read_text(encoding="utf-8"))
    eligible = manifest["combined_skill_cases"] if args.lever_id == "COMBINED" else manifest["levers"].get(args.lever_id, [])
    if not case or args.case_id not in eligible:
        raise SystemExit("case is not preregistered for this lever")
    index = eligible.index(args.case_id)
    null_case = args.lever_id != "COMBINED" and index == manifest["null_case_index"]
    if args.variant == "rewrite":
        if not args.rewrite:
            raise SystemExit("--rewrite is required for rewrite variant")
        target_text = json.loads(args.rewrite.read_text(encoding="utf-8"))["markdown"]
    else:
        target_text = case["draft"]
    query, candidates, target_id = pack(case, target_text, args.lever_id, null_case)
    rendered = "\n\n".join(f"[{item['id']}]\n{item['passage']}" for item in candidates)
    prompt = f"""You are running a post-retrieval source-selection task in a clean context with no tools. Answer the query using only the supplied constructed candidate passages. Cite a candidate ID immediately after each claim it supports. Cite only candidates that materially support the answer; citing none is correct when the pack is irrelevant or insufficient. Do not use outside knowledge. In `cited_candidate_ids`, list only IDs actually cited in the answer. In `accurately_used_candidate_ids`, list cited IDs whose information you used without changing its meaning, scope, or caveats. List every unsupported claim in `unsupported_claims`. Return only schema-valid JSON.\n\nQUERY\n{query}\n\nCONSTRUCTED CANDIDATE PACK\n{rendered}\n"""
    default_model = "gpt-5.6-sol" if args.executor == "codex" else "claude-opus-4-8"
    value = run_structured(
        executor=args.executor, prompt=prompt,
        schema_path=EVAL / "harness" / "citation-schema.json",
        output=args.output, kind="citation_selection", case_id=args.case_id,
        condition=f"{args.lever_id}:{args.variant}", model=args.model or default_model,
        dry_run=args.dry_run,
    )
    metadata = {
        "case_id": args.case_id,
        "lever_id": args.lever_id,
        "variant": args.variant,
        "generator": args.generator,
        "selector": args.executor,
        "null_case": null_case,
        "target_candidate_id": target_id,
        "candidate_order": [item["id"] for item in candidates],
        "candidate_sha256": {item["id"]: sha256_text(item["passage"]) for item in candidates},
        "query_sha256": sha256_text(query),
        "prompt_sha256": value.get("prompt_sha256") if args.dry_run else sha256_text(prompt),
    }
    if args.dry_run:
        print(json.dumps({"runner": value, "metadata": metadata}, indent=2))
    else:
        args.metadata.parent.mkdir(parents=True, exist_ok=True)
        args.metadata.write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
