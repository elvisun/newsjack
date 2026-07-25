#!/usr/bin/env python3
"""Reproduce descriptive corpus, stability, and paired simulation summaries."""

from __future__ import annotations

import argparse
import json
import math
from collections import Counter, defaultdict
from pathlib import Path
from urllib.parse import urlsplit

from pilotlib import read_jsonl, wilson
from model_features import analyze as analyze_features


ROOT = Path(__file__).resolve().parents[1]


def interval(successes: float, n: int) -> dict:
    low, high = wilson(successes, n)
    return {"estimate": successes / n if n else None, "low": low if n else None, "high": high if n else None, "n": n}


def jaccard(left: set[str], right: set[str]) -> float | None:
    union = left | right
    return len(left & right) / len(union) if union else None


def bh_adjust(pvalues: dict[str, float]) -> dict[str, float]:
    ordered = sorted(pvalues.items(), key=lambda item: item[1])
    adjusted = {}
    running = 1.0
    for reverse_index, (key, value) in enumerate(reversed(ordered), 1):
        rank = len(ordered) - reverse_index + 1
        running = min(running, value * len(ordered) / rank)
        adjusted[key] = min(1.0, running)
    return adjusted


def summarize(observations: list[dict], simulations: list[dict], features: list[dict] | None = None) -> dict:
    platforms = {}
    by_platform = defaultdict(list)
    for row in observations:
        by_platform[row.get("platform")].append(row)
    for platform in ("google", "chatgpt"):
        rows = by_platform.get(platform, [])
        outcomes = Counter(row.get("outcome") for row in rows)
        labels = Counter(
            citation.get("label")
            for row in rows for citation in row.get("citations", [])
        )
        platforms[platform] = {
            "n": len(rows), "outcomes": dict(outcomes), "citation_labels": dict(labels),
            "has_citations": interval(sum(row.get("outcome") == "has_citations" for row in rows), len(rows)),
            "no_ai_answer": interval(outcomes.get("no_ai_answer", 0), len(rows)),
            "ai_answer_no_citations": interval(outcomes.get("ai_answer_no_citations", 0), len(rows)),
        }
    pair_rows = defaultdict(dict)
    for row in observations:
        pair_rows[row.get("paired_unit_id")][row.get("platform")] = row
    url_overlap = []
    domain_overlap = []
    for pair in pair_rows.values():
        if set(pair) != {"google", "chatgpt"}:
            continue
        google_urls = {c["canonical_url"] for c in pair["google"].get("citations", [])}
        chat_urls = {c["canonical_url"] for c in pair["chatgpt"].get("citations", [])}
        google_domains = {urlsplit(url).hostname for url in google_urls}
        chat_domains = {urlsplit(url).hostname for url in chat_urls}
        u = jaccard(google_urls, chat_urls)
        d = jaccard(google_domains, chat_domains)
        if u is not None:
            url_overlap.append(u)
        if d is not None:
            domain_overlap.append(d)
    stability_groups = defaultdict(list)
    for row in observations:
        base = row.get("query") or row.get("repeat_of") or row.get("paired_unit_id")
        stability_groups[(base, row.get("platform"))].append(row)
    stability = []
    for rows in stability_groups.values():
        if len(rows) < 2:
            continue
        first = {c["canonical_url"] for c in rows[0].get("citations", [])}
        for row in rows[1:]:
            other = {c["canonical_url"] for c in row.get("citations", [])}
            value = jaccard(first, other)
            if value is not None:
                stability.append(value)
    ablations = {}
    for lever in sorted({row.get("lever_id") for row in simulations if row.get("lever_id")}):
        rows = [row for row in simulations if row.get("lever_id") == lever]
        citation_values = [float(row["citation_selected_rewrite"]) - float(row["citation_selected_original"]) for row in rows if all(key in row for key in ("citation_selected_rewrite", "citation_selected_original"))]
        use_values = [float(row["accurate_use_rewrite"]) - float(row["accurate_use_original"]) for row in rows if all(key in row for key in ("accurate_use_rewrite", "accurate_use_original"))]
        quality_values = [float(row.get("human_quality_delta", 0)) for row in rows]
        fidelity_values = [float(row.get("fact_fidelity_delta", 0)) for row in rows]
        changed = [float(row["changed_token_fraction"]) for row in rows if isinstance(row.get("changed_token_fraction"), (int, float))]
        citation_pp = 100 * sum(citation_values) / len(citation_values) if citation_values else None
        use_pp = 100 * sum(use_values) / len(use_values) if use_values else None
        median_changed = sorted(changed)[len(changed)//2] if changed else None
        efficiency = None
        if citation_pp is not None and use_pp is not None and median_changed is not None:
            efficiency = ((citation_pp + use_pp) / 2) * 0.10 / max(median_changed, 0.10)
        ablations[lever] = {
            "n": len(rows), "citation_selection_lift_pp": citation_pp,
            "accurate_answer_use_lift_pp": use_pp,
            "human_quality_delta": sum(quality_values)/len(quality_values) if quality_values else None,
            "fact_fidelity_delta": sum(fidelity_values)/len(fidelity_values) if fidelity_values else None,
            "median_changed_token_fraction": median_changed, "efficiency_index": efficiency,
        }
    result = {
        "observations": len(observations), "paired_units": len(pair_rows),
        "platforms": platforms,
        "platform_overlap": {
            "mean_url_jaccard": sum(url_overlap)/len(url_overlap) if url_overlap else None,
            "mean_domain_jaccard": sum(domain_overlap)/len(domain_overlap) if domain_overlap else None,
            "n": len(url_overlap),
        },
        "repeat_stability": {"mean_url_jaccard": sum(stability)/len(stability) if stability else None, "n": len(stability)},
        "ablations": ablations,
        "method_note": "Descriptive and paired estimates are not live causal effects. ChatGPT source selection is not estimable without an exposed retrieved-but-rejected risk set.",
    }
    result["multivariable"] = analyze_features(features) if features is not None else {"status": "not_run", "reason": "no extracted feature table supplied"}
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--observations", type=Path, default=ROOT / "runs" / "current" / "observations.jsonl")
    parser.add_argument("--simulations", type=Path, default=ROOT / "runs" / "current" / "citation-simulations.jsonl")
    parser.add_argument("--features", type=Path)
    parser.add_argument("--output", type=Path, default=ROOT / "runs" / "current" / "analysis.json")
    args = parser.parse_args()
    result = summarize(read_jsonl(args.observations), read_jsonl(args.simulations), read_jsonl(args.features) if args.features else None)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
