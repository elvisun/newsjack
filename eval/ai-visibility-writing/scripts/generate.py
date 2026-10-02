#!/usr/bin/env python3
"""Generate one clean-context baseline, skill, or atomic-lever output."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path

from model_runner import EVAL, REPO, run_structured
from pilotlib import PilotError, compare_fact_ledgers, sha256_text


def revision_scope(markdown: str) -> tuple[str, str]:
    """Return the rewritten passage, excluding an audit wrapper when present."""
    heading = re.search(r"(?im)^#{1,6}\s+Revision(?:\s+only\s+if\s+requested)?\s*$", markdown)
    if not heading:
        return markdown, "full_markdown"
    tail = markdown[heading.end():]
    next_heading = re.search(r"(?im)^#{1,6}\s+(?:Fact-preservation\s+note|Measurement\s+caveat)\s*$", tail)
    revision = tail[:next_heading.start()] if next_heading else tail
    revision = revision.strip()
    if not revision:
        return markdown, "full_markdown"
    return revision, "revision_section"


def protected_item_preserved(item: str, rewrite: str) -> bool:
    def normalize_phrase(text: str) -> str:
        return re.sub(r"\blanguages?\s+other\s+than\s+english\b", "other languages", text.lower())

    needle = re.findall(r"[a-z0-9]+", normalize_phrase(item))
    haystack = re.findall(r"[a-z0-9]+", normalize_phrase(rewrite))

    def variants(token: str) -> set[str]:
        values = {token}
        if len(token) > 5 and token.endswith("ing"):
            stem = token[:-3]
            values.update({stem, stem + "e"})
            if len(stem) > 2 and stem[-1] == stem[-2]:
                values.add(stem[:-1])
        return values

    def matches(left: str, right: str) -> bool:
        return bool(variants(left) & variants(right))

    if not needle:
        return True
    for start, token in enumerate(haystack):
        if not matches(token, needle[0]):
            continue
        position = start + 1
        gaps = 0
        for expected in needle[1:]:
            while position < len(haystack) and not matches(haystack[position], expected):
                gaps += 1
                position += 1
            if position == len(haystack) or gaps > 2:
                break
            position += 1
        else:
            return True
    return False


def fact_guard_result(case: dict, markdown: str) -> dict:
    rewrite, scope = revision_scope(markdown)
    guard = compare_fact_ledgers(case["draft"], rewrite)
    supplied_url = case.get("source_url") if not case.get("constructed", True) else None
    allowed_provenance_urls = [
        value for value in guard["added"]["urls"] if supplied_url and value == supplied_url
    ]
    if allowed_provenance_urls:
        guard["added"]["urls"] = [
            value for value in guard["added"]["urls"] if value not in allowed_provenance_urls
        ]
        guard["pass"] = not any(
            guard[side][category]
            for side in ("missing", "added")
            for category in ("numbers", "dates", "quotes", "urls")
        )
    protected_missing = [item for item in case.get("protected_items", []) if not protected_item_preserved(item, rewrite)]
    return {
        "pass": bool(guard["pass"] and not protected_missing),
        "compared_scope": scope,
        "compared_text_sha256": sha256_text(rewrite),
        "ledger_diff": guard,
        "allowed_provenance_urls": allowed_provenance_urls,
        "protected_missing": protected_missing,
    }


def enforce_fact_guard(condition: str, result: dict, output: Path) -> None:
    sidecar = output.resolve().with_suffix(".validation.json")
    sidecar.write_text(json.dumps({"condition": condition, **result}, indent=2) + "\n", encoding="utf-8")
    if not result["pass"] and condition not in {"bare", "editorial"}:
        raise PilotError("generated rewrite failed deterministic fact guard")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--case-id", required=True)
    parser.add_argument("--executor", choices=["codex", "claude"], required=True)
    parser.add_argument("--condition", choices=["bare", "editorial", "skill", "skill_rewrite", "lever"], required=True)
    parser.add_argument("--lever-id")
    parser.add_argument("--model")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()
    if args.condition == "lever" and not args.lever_id:
        raise SystemExit("--lever-id is required for the lever condition")
    cases = json.loads((EVAL / "fixtures" / "cases.json").read_text(encoding="utf-8"))["cases"]
    case = next((item for item in cases if item["id"] == args.case_id), None)
    if not case:
        raise SystemExit("unknown case id")
    requested_behavior = "rewrite" if args.condition == "skill_rewrite" else case["behavior"]
    if case.get("constructed", True):
        provenance = "CONSTRUCTED FIXTURE — NEVER PRESENT AS REAL"
    elif args.condition == "lever":
        provenance = "ATTRIBUTED PUBLIC-DOMAIN SOURCE ADAPTATION — edit only the supplied draft; do not reproduce source metadata"
    else:
        provenance = f"ATTRIBUTED PUBLIC-DOMAIN SOURCE ADAPTATION\nSource: {case['source_url']}\nLicense basis: {case['license_basis']}"
    common = f"""You are working in a clean context with no tools. Use only the supplied fixture as facts. Do not search, invent, or infer outside facts. The requested behavior is {requested_behavior}. Return JSON matching the supplied schema and set `behavior` to `{requested_behavior}`. Put the complete human-facing Markdown in `markdown`; keep it concise (maximum 700 words).\n\n{provenance}\nDocument type: {case['document_type']}\nTarget audience/query: {case['target_query']}\nDraft:\n{case['draft']}\n"""
    if args.condition == "bare":
        instruction = "\nOptimize this writing for AI visibility."
    elif args.condition == "editorial":
        instruction = """
\nAct as a careful senior editor. Identify the audience and answer the likely query directly. Preserve all facts, numbers, dates, quotations, attributions, scope, and caveats. Ask only questions that block a sound result. Rank no more than three useful edits with reasons and tradeoffs. Rewrite only when requested and safe. Separate page/indexing issues from prose. Never promise rankings, citations, traffic, or coverage. Use headings: Verdict; Highest-leverage changes; Blocking questions; Revision only if requested; Fact-preservation note; Measurement caveat.
"""
    elif args.condition in {"skill", "skill_rewrite"}:
        skill = (REPO / "skills" / "ai-visibility-writing" / "SKILL.md").read_text(encoding="utf-8")
        instruction = "\nApply the following skill faithfully:\n\n" + skill
        if args.condition == "skill_rewrite":
            instruction += "\n\nFor this post-retrieval experiment, put only the safely revised document passage in `markdown`, without the audit, verdict, table, questions, change note, or measurement caveat. If the skill would block the rewrite, set `blocking` true and return the original passage unchanged."
    else:
        levers = json.loads((EVAL / "levers.json").read_text(encoding="utf-8"))["levers"]
        lever = next((item for item in levers if item["id"] == args.lever_id), None)
        if not lever:
            raise SystemExit("unknown lever id")
        instruction = f"""
\nMake exactly one writing intervention and no other optimization: {lever['lever']} — {lever['intervention']} In `markdown`, return only the edited candidate passage, not an audit, headings, explanation, or change note. Do not add a source, link, citation, label, or note unless the named intervention explicitly requires it. Preserve every supplied fact, relationship, caveat, and the document's voice. Do not reuse a supplied number in a new timeframe, denominator, relationship, or scope that the draft does not explicitly state. If the intervention cannot be made without missing facts, set `blocking` true and return the original passage unchanged. Make no claim about evaluation or guaranteed outcomes.
"""
    prompt = common + instruction
    default_model = "gpt-5.6-sol" if args.executor == "codex" else "claude-opus-4-8"
    value = run_structured(
        executor=args.executor, prompt=prompt,
        schema_path=EVAL / "harness" / "generation-schema.json",
        output=args.output, kind="generation", case_id=args.case_id,
        condition=args.condition if not args.lever_id else f"lever:{args.lever_id}",
        model=args.model or default_model, dry_run=args.dry_run,
    )
    if not args.dry_run and ((requested_behavior == "rewrite" and args.condition != "lever") or args.condition == "lever") and not value.get("blocking"):
        result = fact_guard_result(case, value["markdown"])
        result["model_claimed_preserved"] = bool(value.get("protected_items_preserved"))
        result["pass"] = bool(result["pass"] and result["model_claimed_preserved"])
        enforce_fact_guard(args.condition, result, args.output)
    if args.dry_run:
        print(json.dumps(value, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
