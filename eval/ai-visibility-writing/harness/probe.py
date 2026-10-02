#!/usr/bin/env python3
"""Generate deterministic robustness probes and summarize available gaps."""

from __future__ import annotations

import argparse
import json
import re
from collections import Counter
from pathlib import Path


EVAL = Path(__file__).resolve().parents[1]


def variants(case: dict) -> list[dict]:
    draft = case["draft"]
    entity = re.sub(r"\b(?:Acme|Northline|Meridian|Harbor|Cedar|Atlas)\b", "Juniper", draft, count=1)
    number = re.sub(r"\b(\d[\d,.]*%?)\b", "17.3%", draft, count=1)
    date = re.sub(r"\b(?:20\d{2}|(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{1,2},\s+20\d{2})\b", "June 3, 2026", draft, count=1)
    query = case.get("target_query", "")
    paraphrase = query.replace("How should", "What is the right way to").replace("What are", "Which are")
    return [
        {"probe": "query_paraphrase", "target_query": paraphrase, "draft": draft},
        {"probe": "entity_swap", "target_query": query, "draft": entity},
        {"probe": "number_swap", "target_query": query, "draft": number},
        {"probe": "date_swap", "target_query": query, "draft": date},
    ]


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--cases", type=Path, default=EVAL / "fixtures" / "cases.json")
    parser.add_argument("--metrics", type=Path)
    parser.add_argument("--fresh-metrics", type=Path)
    args = parser.parse_args()
    cases = json.loads(args.cases.read_text(encoding="utf-8"))["cases"]
    generated = [{"case_id": case["id"], **variant} for case in cases for variant in variants(case)]
    families = Counter(case["topic_family"] for case in cases)
    documents = Counter(case["document_type"] for case in cases)
    result = {
        "cases": len(cases), "probe_variants": len(generated),
        "probe_types": dict(Counter(row["probe"] for row in generated)),
        "leave_topic_family_out": {key: len(cases)-value for key, value in sorted(families.items())},
        "leave_document_type_out": {key: len(cases)-value for key, value in sorted(documents.items())},
        "dev_fresh_gap": None,
    }
    if args.metrics and args.fresh_metrics and args.metrics.exists() and args.fresh_metrics.exists():
        dev = json.loads(args.metrics.read_text(encoding="utf-8"))["composite"]
        fresh = json.loads(args.fresh_metrics.read_text(encoding="utf-8"))["composite"]
        result["dev_fresh_gap"] = dev - fresh
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
