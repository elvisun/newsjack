#!/usr/bin/env python3
"""Apply coarse-filter LLM decisions to detector candidate JSON.

This script intentionally does not call an LLM. It validates a harness-written
decision file, attaches decisions to the original signals, and emits a smaller
candidate set for the expensive newsworthiness pass.
"""

from __future__ import annotations

import argparse
import json
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


ALLOWED_DECISIONS = {"keep", "monitor_only", "reject"}
ALLOWED_REASONS = {
    "relevant_news",
    "plausible_client_bridge",
    "major_news_no_bridge",
    "keyword_collision",
    "not_news",
    "owned_docs_or_product_page",
    "seo_landing_page",
    "low_reach_x_post",
    "stale",
    "freshness_unverified",
    "safety_risk",
    "duplicate",
    "off_beat",
    "no_profile_bridge",
}


def main() -> int:
    parser = argparse.ArgumentParser(description="Apply coarse LLM filter decisions to newsjack detector candidates.")
    parser.add_argument("--candidates", required=True, help="Detector JSON output from newsjack_detector.py run --emit json.")
    parser.add_argument("--decisions", required=True, help="Coarse-filter decision JSON written by a harness/LLM.")
    parser.add_argument("--output", help="Output path. Defaults to stdout.")
    parser.add_argument(
        "--include",
        action="append",
        choices=["keep", "monitor_only"],
        default=[],
        help="Decision to include in targeted output. Repeatable. Defaults to keep only.",
    )
    parser.add_argument("--allow-missing", action="store_true", help="Do not fail when a candidate signal has no decision.")
    parser.add_argument("--allow-unknown", action="store_true", help="Do not fail when decisions reference unknown signal IDs.")
    parser.add_argument(
        "--require-fresh-first-publication",
        action="store_true",
        help="Fail if an included decision lacks first_publication.status=fresh or fresh_new_development.",
    )
    args = parser.parse_args()

    candidates = _read_json(Path(args.candidates))
    decisions_payload = _read_json(Path(args.decisions))
    output = apply_decisions(
        candidates,
        decisions_payload,
        include=set(args.include or ["keep"]),
        allow_missing=args.allow_missing,
        allow_unknown=args.allow_unknown,
        require_fresh_first_publication=args.require_fresh_first_publication,
    )
    rendered = json.dumps(output, indent=2, sort_keys=True)
    if args.output:
        Path(args.output).expanduser().write_text(f"{rendered}\n", encoding="utf-8")
    else:
        print(rendered)
    return 0


def apply_decisions(
    candidates: dict[str, Any],
    decisions_payload: dict[str, Any] | list[dict[str, Any]],
    *,
    include: set[str],
    allow_missing: bool = False,
    allow_unknown: bool = False,
    require_fresh_first_publication: bool = False,
) -> dict[str, Any]:
    signals = list(candidates.get("signals") or [])
    signal_by_id = {str(signal.get("id")): signal for signal in signals if signal.get("id")}
    decisions = _normalize_decisions(decisions_payload)
    decision_by_id: dict[str, dict[str, Any]] = {}
    errors: list[str] = []

    for decision in decisions:
        signal_id = str(decision.get("signal_id") or "").strip()
        if not signal_id:
            errors.append("decision missing signal_id")
            continue
        if signal_id in decision_by_id:
            errors.append(f"duplicate decision for signal_id={signal_id}")
            continue
        if signal_id not in signal_by_id and not allow_unknown:
            errors.append(f"decision references unknown signal_id={signal_id}")
            continue
        normalized = _normalize_decision(decision)
        if normalized["decision"] not in ALLOWED_DECISIONS:
            errors.append(f"{signal_id}: unsupported decision={normalized['decision']}")
        if normalized["reason"] not in ALLOWED_REASONS:
            errors.append(f"{signal_id}: unsupported reason={normalized['reason']}")
        if (
            require_fresh_first_publication
            and normalized["decision"] in include
            and not _fresh_first_publication_status(normalized.get("first_publication"))
        ):
            errors.append(
                f"{signal_id}: included decision requires first_publication.status=fresh or fresh_new_development"
            )
        decision_by_id[signal_id] = normalized

    missing_ids = [signal_id for signal_id in signal_by_id if signal_id not in decision_by_id]
    if missing_ids and not allow_missing:
        sample = ", ".join(missing_ids[:8])
        suffix = "..." if len(missing_ids) > 8 else ""
        errors.append(f"missing decisions for {len(missing_ids)} signal(s): {sample}{suffix}")
    if errors:
        raise SystemExit("\n".join(errors))

    selected = []
    rejected = []
    missing = []
    for signal in signals:
        signal_id = str(signal.get("id") or "")
        decision = decision_by_id.get(signal_id)
        if not decision:
            missing.append(_summary_signal(signal))
            continue
        signal_with_decision = dict(signal)
        signal_with_decision["coarse_filter"] = decision
        if decision["decision"] in include:
            selected.append(signal_with_decision)
        else:
            rejected_signal = {
                **_summary_signal(signal),
                "decision": decision["decision"],
                "reason": decision["reason"],
                "rationale": decision["rationale"],
            }
            if "first_publication" in decision:
                rejected_signal["first_publication"] = decision["first_publication"]
            rejected.append(rejected_signal)

    decision_counts = Counter(decision["decision"] for decision in decision_by_id.values())
    reason_counts = Counter(decision["reason"] for decision in decision_by_id.values())
    return {
        "version": 1,
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "monitor": candidates.get("monitor") or {},
        "signals": selected,
        "coarse_filter": {
            "input_signal_count": len(signals),
            "decision_count": len(decision_by_id),
            "selected_count": len(selected),
            "rejected_count": len(rejected),
            "missing_count": len(missing),
            "included_decisions": sorted(include),
            "decision_counts": dict(sorted(decision_counts.items())),
            "reason_counts": dict(sorted(reason_counts.items())),
            "rejected_signals": rejected,
            "missing_signals": missing,
        },
        "detector_diagnostics": candidates.get("diagnostics") or {},
        "source_errors": candidates.get("source_errors") or {},
    }


def _normalize_decisions(payload: dict[str, Any] | list[dict[str, Any]]) -> list[dict[str, Any]]:
    if isinstance(payload, list):
        return payload
    if isinstance(payload, dict):
        decisions = payload.get("decisions")
        if isinstance(decisions, list):
            return decisions
    raise SystemExit("decisions JSON must be a list or an object with a decisions list")


def _normalize_decision(decision: dict[str, Any]) -> dict[str, Any]:
    evidence_urls = decision.get("evidence_urls") or []
    if isinstance(evidence_urls, str):
        evidence_urls = [evidence_urls]
    normalized = {
        "signal_id": str(decision.get("signal_id") or "").strip(),
        "decision": str(decision.get("decision") or "").strip(),
        "reason": str(decision.get("reason") or "").strip(),
        "rationale": str(decision.get("rationale") or "").strip(),
        "confidence": str(decision.get("confidence") or "").strip() or "medium",
        "evidence_urls": [str(url) for url in evidence_urls if str(url).strip()],
    }
    if isinstance(decision.get("first_publication"), dict):
        normalized["first_publication"] = decision["first_publication"]
    return normalized


def _fresh_first_publication_status(value: Any) -> bool:
    if not isinstance(value, dict):
        return False
    return str(value.get("status") or "").strip() in {"fresh", "fresh_new_development"}


def _summary_signal(signal: dict[str, Any]) -> dict[str, Any]:
    return {
        "signal_id": signal.get("id"),
        "signal_title": signal.get("title"),
        "sources": signal.get("sources") or [],
        "routing": signal.get("routing") or {},
        "evidence_urls": [
            evidence.get("url")
            for evidence in signal.get("evidence") or []
            if evidence.get("url")
        ][:5],
    }


def _read_json(path: Path) -> Any:
    with path.expanduser().open("r", encoding="utf-8") as handle:
        return json.load(handle)


if __name__ == "__main__":
    sys.exit(main())
