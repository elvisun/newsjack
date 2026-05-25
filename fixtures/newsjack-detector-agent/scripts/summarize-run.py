#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def main() -> int:
    parser = argparse.ArgumentParser(description="Summarize an observed newsjack detector run.")
    parser.add_argument("input", help="Detector candidates JSON or targeted candidates JSON.")
    parser.add_argument("--output", required=True, help="Path to write machine-readable summary JSON.")
    parser.add_argument("--markdown", help="Path to write the observable Markdown run report.")
    parser.add_argument("--brief", help="Deprecated alias for --markdown.")
    parser.add_argument("--top", type=int, default=25, help="Number of selected and dropped signals to include.")
    args = parser.parse_args()
    markdown_target = args.markdown or args.brief
    if not markdown_target:
        parser.error("--markdown is required")

    input_path = Path(args.input)
    payload = json.loads(input_path.read_text(encoding="utf-8"))
    summary = summarize(payload, input_path=input_path, top=max(0, args.top))

    output_path = Path(args.output)
    markdown_path = Path(markdown_target)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    markdown_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    markdown_path.write_text(render_markdown(summary), encoding="utf-8")
    return 0


def summarize(payload: dict[str, Any], *, input_path: Path, top: int) -> dict[str, Any]:
    signals = list(payload.get("signals") or [])
    diagnostics = dict(payload.get("diagnostics") or payload.get("detector_diagnostics") or {})
    debug = dict(payload.get("debug") or {})
    all_scored = list(debug.get("all_scored_signals") or [])
    selected_ids = {str(signal.get("id")) for signal in signals if signal.get("id")}
    all_scored_ids = [str(signal.get("id")) for signal in all_scored if signal.get("id")]
    dropped_signals = [
        signal for signal in all_scored
        if str(signal.get("id")) not in selected_ids
    ]
    selected_debug_rows = [
        signal for signal in all_scored
        if str(signal.get("id")) in selected_ids
    ]

    monitor = dict(payload.get("monitor") or {})
    source_errors = dict(payload.get("source_errors") or {})
    run_dir = input_path.parent
    artifact_paths = _artifact_paths(run_dir)

    summary = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "input_path": str(input_path),
        "run_dir": str(run_dir),
        "artifacts": _artifact_status(artifact_paths),
        "pipeline": _pipeline_status(artifact_paths),
        "monitor": {
            "name": monitor.get("name"),
            "generated_at": monitor.get("generated_at"),
            "profile_name": _profile_name(monitor),
            "queries": monitor.get("queries") or [],
            "feed_urls": monitor.get("feed_urls") or [],
            "sources_requested": monitor.get("sources_requested") or [],
            "sources_used": monitor.get("sources_used") or [],
            "lookback_days": monitor.get("lookback_days"),
            "max_age_hours": monitor.get("max_age_hours"),
            "depth": monitor.get("depth"),
            "mock": monitor.get("mock"),
        },
        "counts": {
            "selected_unique_signals": len(signals),
            "total_scored_signals": diagnostics.get("total_scored_signals", len(all_scored) or None),
            "total_emitted_signals": diagnostics.get("total_emitted_signals", len(signals)),
            "debug_all_scored_rows": len(all_scored),
            "debug_unique_scored_signal_ids": len(set(all_scored_ids)),
            "debug_selected_rows": len(selected_debug_rows),
            "debug_unselected_rows": len(dropped_signals),
            "debug_duplicate_scored_rows": len(all_scored_ids) - len(set(all_scored_ids)),
            "source_errors": len(source_errors),
        },
        "selection": diagnostics.get("selection") or {},
        "lanes": {
            "scored": diagnostics.get("signals_by_lane") or _count_lanes(all_scored),
            "emitted": diagnostics.get("emitted_by_lane") or _count_lanes(signals),
            "dropped_debug": _count_lanes(dropped_signals),
        },
        "sources": {
            "evidence_by_source": diagnostics.get("evidence_by_source") or _count_evidence_sources(signals),
            "source_errors": source_errors,
        },
        "hygiene_rejections": diagnostics.get("hygiene_rejections") or {},
        "coarse_filter": _coarse_filter(payload),
        "coarse_filter_file": _summarize_decisions(artifact_paths["filter_decisions"]),
        "targeted_candidates_file": _summarize_targeted_candidates(artifact_paths["targeted_candidates"]),
        "final_report_file": _summarize_final_report(artifact_paths["final_report"]),
        "top_signals": [_summarize_signal(signal) for signal in signals[:top]],
        "top_dropped_signals": [
            _summarize_signal(signal)
            for signal in sorted(dropped_signals, key=_queue_priority, reverse=True)[:top]
        ],
    }
    return summary


def render_markdown(summary: dict[str, Any]) -> str:
    monitor = summary.get("monitor") or {}
    counts = summary.get("counts") or {}
    lanes = summary.get("lanes") or {}
    sources = summary.get("sources") or {}
    selection = summary.get("selection") or {}
    coarse_summary = summary.get("coarse_filter_file") or {}
    targeted_summary = summary.get("targeted_candidates_file") or {}
    final_report = summary.get("final_report_file") or {}
    final_text = final_report.get("content")
    profile_name = monitor.get("profile_name") or "Newsjack"
    lines: list[str] = []

    lines.append(f"# {profile_name} Newsjack Brief")
    lines.append("")
    lines.extend(_render_brief_header(summary))
    lines.append("")

    if final_text:
        lines.append("## Editorial Judgment")
        lines.extend(_render_final_report(final_text))
    else:
        lines.append("## Review Status")
        lines.append("")
        lines.append(
            "**Not ready to forward as recommendations yet.** The detector run is complete, "
            "but the editorial judgment pass has not produced `final_report.md`."
        )
        lines.append("")
        lines.append("Next action: run the coarse filter, apply it, write `final_report.md`, then rerender this file.")
        lines.append("")
        lines.append("## Candidate Preview")
        lines.append("")
        lines.append("These are the highest-priority signals for review. They are not final pitch recommendations.")
        lines.append("")
        lines.extend(_render_candidate_cards(
            summary.get("top_signals") or [],
            limit=8,
            include_ids=False,
            total_count=counts.get("selected_unique_signals"),
        ))
    lines.append("")

    lines.append("## What Was Scanned")
    lines.append("")
    lines.extend(_render_scan_summary(monitor, sources))
    lines.append("")

    if source_issues := _source_issue_rows(sources):
        lines.append("### Source Issues")
        lines.append("")
        for label, detail in source_issues:
            lines.append(f"- **{_md_inline(label)}:** {_md_inline(detail)}")
        lines.append("")

    lines.append("## Run Notes")
    lines.append("")
    lines.extend(_render_run_notes(summary))
    lines.append("")

    lines.append("## Appendix: Provenance")
    lines.append("")
    lines.append("### Pipeline")
    lines.append("")
    lines.extend(_render_table([
        (stage.get("stage"), f"{stage.get('status')} - {stage.get('artifact')}")
        for stage in summary.get("pipeline") or []
    ]))
    lines.append("")

    lines.append("### Run Context")
    lines.append("")
    lines.append(f"- **Input:** `{summary.get('input_path')}`")
    lines.append(f"- **Profile:** {_md_inline(monitor.get('profile_name') or '(unknown)')}")
    queries = monitor.get("queries") or []
    lines.append(f"- **Queries:** {_md_inline(', '.join(queries) if queries else '(none)')}")
    lines.append(f"- **Sources used:** `{', '.join(monitor.get('sources_used') or [])}`")

    lines.append("")
    lines.append("### Detector Counts")
    lines.append("")
    lines.extend(_render_table([
        ("scored", counts.get("total_scored_signals")),
        ("selected", counts.get("selected_unique_signals")),
        ("debug rows", counts.get("debug_all_scored_rows")),
        ("debug selected rows", counts.get("debug_selected_rows")),
        ("debug unselected rows", counts.get("debug_unselected_rows")),
        ("debug duplicate rows", counts.get("debug_duplicate_scored_rows")),
        ("source_errors", counts.get("source_errors")),
    ]))

    if coarse_summary.get("exists"):
        lines.append("")
        lines.append("### Coarse Filter")
        lines.append("")
        lines.extend(_render_table([
            ("decisions", coarse_summary.get("decision_count")),
            *[
                (f"decision.{key}", value)
                for key, value in (coarse_summary.get("decisions_by_outcome") or {}).items()
            ],
            *[
                (f"reason.{key}", value)
                for key, value in (coarse_summary.get("decisions_by_reason") or {}).items()
            ],
        ]))

    if targeted_summary.get("exists"):
        lines.append("")
        lines.append("### Targeted Candidates")
        lines.append("")
        lines.extend(_render_table([
            ("selected signals", targeted_summary.get("selected_signals")),
            ("input signals", targeted_summary.get("input_signals")),
            ("rejected signals", targeted_summary.get("rejected_signals")),
        ]))

    if selection:
        lines.append("")
        lines.append("### Selection")
        lines.append("")
        lines.extend(_render_table(selection.items()))

    lines.append("")
    lines.append("### Lanes")
    lines.append("")
    lane_rows: list[tuple[str, Any]] = []
    for group in ("scored", "emitted", "dropped_debug"):
        for lane, count in (lanes.get(group) or {}).items():
            lane_rows.append((f"{group}.{lane}", count))
    lines.extend(_render_table(lane_rows))

    lines.append("")
    lines.append("### Evidence Sources")
    lines.append("")
    source_rows: list[tuple[str, Any]] = []
    for source, count in (sources.get("evidence_by_source") or {}).items():
        source_rows.append((f"evidence.{source}", count))
    for source, error in (sources.get("source_errors") or {}).items():
        source_rows.append((f"error.{source}", error))
    lines.extend(_render_table(source_rows))

    hygiene = summary.get("hygiene_rejections") or {}
    if hygiene:
        lines.append("")
        lines.append("### Hygiene Rejections")
        lines.append("")
        lines.extend(_render_table(hygiene.items()))

    lines.append("")
    lines.append("### Candidate Queue")
    lines.append("")
    lines.extend(_render_candidate_cards(
        summary.get("top_signals") or [],
        limit=25,
        include_ids=True,
        total_count=counts.get("selected_unique_signals"),
    ))

    dropped = summary.get("top_dropped_signals") or []
    if dropped:
        lines.append("")
        lines.append("### Top Dropped Candidates")
        lines.append("")
        lines.extend(_render_candidate_cards(
            dropped,
            limit=25,
            include_ids=True,
            total_count=counts.get("debug_unselected_rows"),
        ))

    return "\n".join(lines).rstrip() + "\n"


def _render_brief_header(summary: dict[str, Any]) -> list[str]:
    monitor = summary.get("monitor") or {}
    counts = summary.get("counts") or {}
    final_report = summary.get("final_report_file") or {}
    coarse_summary = summary.get("coarse_filter_file") or {}
    targeted_summary = summary.get("targeted_candidates_file") or {}
    final_payload = _parse_report_json(final_report.get("content") or "")
    parsed_final = isinstance(final_payload, dict)
    opportunities = list(final_payload.get("opportunities") or []) if parsed_final else []
    pitch_now = sum(1 for item in opportunities if item.get("verdict") == "pitch_now")
    develop = sum(1 for item in opportunities if item.get("verdict") == "develop_angle")
    monitor_count = sum(1 for item in opportunities if item.get("verdict") == "monitor")
    generated = monitor.get("generated_at") or summary.get("generated_at")
    selected = counts.get("selected_unique_signals")
    targeted = targeted_summary.get("selected_signals") if targeted_summary.get("exists") else None

    status = "Editorial review complete" if final_report.get("exists") else "Detector preview only"
    coarse_status = "complete" if coarse_summary.get("exists") else "pending"
    targeted_status = f"{targeted} targeted" if targeted is not None else "pending"
    action = _action_summary(
        pitch_now=pitch_now,
        develop=develop,
        monitor=monitor_count,
        has_final=final_report.get("exists"),
        parsed_final=parsed_final,
    )

    rows = [
        ("Status", status),
        ("Generated", _format_datetime(generated)),
        ("Action queue", action),
        ("Detector candidates", selected),
        ("Coarse filter", coarse_status),
        ("Targeted set", targeted_status),
    ]
    lines = _render_table(rows)
    return lines


def _action_summary(*, pitch_now: int, develop: int, monitor: int, has_final: bool, parsed_final: bool) -> str:
    if not has_final:
        return "Needs editorial pass before sharing"
    if not parsed_final:
        return "See editorial judgment"
    parts = []
    if pitch_now:
        parts.append(f"{pitch_now} pitch now")
    if develop:
        parts.append(f"{develop} develop")
    if monitor:
        parts.append(f"{monitor} monitor")
    return ", ".join(parts) if parts else "No recommended action"


def _render_final_report(content: str) -> list[str]:
    payload = _parse_report_json(content)
    if not isinstance(payload, dict):
        return ["", content.rstrip(), ""]

    lines: list[str] = [""]
    opportunities = list(payload.get("opportunities") or [])
    blocks = list(payload.get("brand_safety_blocks") or [])
    rejects = list(payload.get("rejected_signals") or [])
    monitor_notes = list(payload.get("monitor_notes") or [])

    action_items = [
        item for item in opportunities
        if item.get("verdict") in {"pitch_now", "develop_angle"}
    ]
    monitor_items = [
        item for item in opportunities
        if item.get("verdict") == "monitor"
    ]

    if action_items:
        lines.append("### Recommended Actions")
        lines.append("")
        for item in action_items:
            lines.extend(_render_opportunity(item))
            lines.append("")
    else:
        lines.append("### Recommended Actions")
        lines.append("")
        lines.append("- No pitch-ready or angle-development opportunities in this run.")
        lines.append("")

    if monitor_items:
        lines.append("### Watch List")
        lines.append("")
        for item in monitor_items:
            lines.extend(_render_opportunity(item))
            lines.append("")

    if blocks:
        lines.append("### Do Not Use")
        lines.append("")
        for item in blocks:
            title = item.get("signal_title") or item.get("title") or "(untitled)"
            reason = _label(item.get("reason") or "blocked")
            lines.append(f"- **{_md_inline(title)}** - {reason}")
        lines.append("")

    if rejects:
        lines.append("### Rejected")
        lines.append("")
        reason_counts = Counter(str(item.get("reason") or "unknown") for item in rejects)
        lines.append(
            "- "
            + ", ".join(f"{_label(reason)}: {count}" for reason, count in sorted(reason_counts.items()))
        )
        notable = rejects[:8]
        if notable:
            lines.append("")
            for item in notable:
                title = item.get("signal_title") or item.get("title") or "(untitled)"
                reason = _label(item.get("reason") or "rejected")
                lines.append(f"- **{_md_inline(title)}** - {reason}")
        lines.append("")

    if monitor_notes:
        lines.append("### Notes")
        lines.append("")
        for note in monitor_notes:
            lines.append(f"- {_md_inline(note)}")
        lines.append("")

    return lines


def _render_opportunity(item: dict[str, Any]) -> list[str]:
    title = item.get("signal_title") or item.get("title") or "(untitled)"
    verdict = _label(item.get("verdict") or "review")
    decay = item.get("decay") if isinstance(item.get("decay"), dict) else {}
    stage = decay.get("stage") if isinstance(decay, dict) else None
    next_skill = item.get("next_skill") or item.get("next")
    standing = item.get("client_standing") if isinstance(item.get("client_standing"), dict) else {}
    journalist = item.get("journalist_shape") if isinstance(item.get("journalist_shape"), dict) else {}
    proof = item.get("required_proof") or []
    evidence = item.get("evidence_used") or item.get("evidence") or []

    lines = [f"#### {_md_inline(title)}"]
    meta = [verdict]
    if stage:
        meta.append(f"decay: {stage}")
    if next_skill:
        meta.append(f"next: {next_skill}")
    lines.append("")
    lines.append(f"**Call:** {'; '.join(_md_inline(part) for part in meta)}")
    if item.get("why_newsjacking_worthy"):
        lines.append(f"**Why it matters:** {_md_inline(item['why_newsjacking_worthy'])}")
    if standing:
        standing_bits = []
        if standing.get("assessment"):
            standing_bits.append(_label(standing.get("assessment")))
        if standing.get("rationale"):
            standing_bits.append(_md_inline(standing.get("rationale")))
        if standing_bits:
            lines.append(f"**Client standing:** {' - '.join(standing_bits)}")
    if proof:
        lines.append("**Proof needed:**")
        for proof_item in proof:
            lines.append(f"- {_md_inline(proof_item)}")
    if journalist:
        beat = journalist.get("beat_description")
        why_now = journalist.get("why_they_care_now")
        avoid = journalist.get("do_not_target")
        if beat:
            lines.append(f"**Reporter shape:** {_md_inline(beat)}")
        if why_now:
            lines.append(f"**Why now:** {_md_inline(why_now)}")
        if avoid:
            lines.append(f"**Do not target:** {_md_inline(avoid)}")
    if evidence:
        lines.append("**Evidence:**")
        for evidence_item in evidence[:5]:
            lines.append(f"- {_render_evidence_link(evidence_item)}")
    return lines


def _render_scan_summary(monitor: dict[str, Any], sources: dict[str, Any]) -> list[str]:
    queries = monitor.get("queries") or []
    feed_urls = monitor.get("feed_urls") or []
    used = set(monitor.get("sources_used") or [])
    evidence_counts = sources.get("evidence_by_source") or {}
    source_bits = [
        _coverage_label("News search", "news_search", used, evidence_counts),
        _coverage_label("X News", "x_news", used, evidence_counts),
        _coverage_label("X trends", "x_trends", used, evidence_counts),
        _coverage_label("X posts", "x", used, evidence_counts),
        f"RSS feeds: {'used' if feed_urls else 'not configured'} ({len(feed_urls)} feed{'s' if len(feed_urls) != 1 else ''})",
    ]
    return [
        f"- **Profile:** {_md_inline(monitor.get('profile_name') or '(unknown)')}",
        f"- **Queries:** {_md_inline(_format_list(queries, limit=8) if queries else '(feed-only run)')}",
        f"- **Lookback:** {_fmt(monitor.get('lookback_days'))} day(s); max item age {_fmt(monitor.get('max_age_hours'))} hour(s); depth {_md_inline(monitor.get('depth') or '(unknown)')}",
        f"- **Coverage:** {'; '.join(source_bits)}",
    ]


def _coverage_label(label: str, key: str, used: set[str], evidence_counts: dict[str, Any]) -> str:
    if key in used:
        count = evidence_counts.get(key, 0)
        return f"{label}: used ({count} item{'s' if count != 1 else ''})"
    return f"{label}: not used"


def _source_issue_rows(sources: dict[str, Any]) -> list[tuple[str, str]]:
    rows: list[tuple[str, str]] = []
    for key, value in (sources.get("source_errors") or {}).items():
        if isinstance(value, dict):
            for nested_key, nested_value in value.items():
                rows.append((f"{key}.{nested_key}", str(nested_value)))
        else:
            rows.append((str(key), str(value)))
    return rows


def _render_run_notes(summary: dict[str, Any]) -> list[str]:
    counts = summary.get("counts") or {}
    hygiene = summary.get("hygiene_rejections") or {}
    coarse_summary = summary.get("coarse_filter_file") or {}
    targeted_summary = summary.get("targeted_candidates_file") or {}
    final_report = summary.get("final_report_file") or {}
    notes = []
    if not final_report.get("exists"):
        notes.append("This report is a detector preview. Do not treat candidates as approved outreach hooks.")
    if counts.get("selected_unique_signals") == 0:
        notes.append("No new candidate signals were emitted for this run.")
    if coarse_summary.get("exists"):
        decisions = coarse_summary.get("decisions_by_outcome") or {}
        notes.append("Coarse filter decisions: " + ", ".join(f"{_label(key)} {value}" for key, value in sorted(decisions.items())) + ".")
    if targeted_summary.get("exists"):
        notes.append(
            f"Targeted set: {targeted_summary.get('selected_signals')} selected from "
            f"{targeted_summary.get('input_signals')} detector candidates."
        )
    if hygiene:
        notes.append("Hygiene filter removed " + ", ".join(f"{_label(key)} {value}" for key, value in sorted(hygiene.items())) + ".")
    if not notes:
        notes.append("No operational notes.")
    return [f"- {note}" for note in notes]


def _render_candidate_cards(
    signals: list[dict[str, Any]],
    *,
    limit: int,
    include_ids: bool,
    total_count: Any = None,
) -> list[str]:
    if not signals:
        return ["- (none)"]
    lines: list[str] = []
    for index, signal in enumerate(signals[:limit], start=1):
        title = _md_inline(signal.get("title") or "(untitled)")
        lane = _label(signal.get("lane") or "unknown")
        score_bits = [
            f"queue {_fmt(signal.get('queue_priority'))}",
            f"profile {_fmt(signal.get('profile_match'))}",
            f"major {_fmt(signal.get('major_news'))}",
        ]
        lines.append(f"{index}. **{title}**")
        lines.append(f"   - Why surfaced: {lane}; {', '.join(score_bits)}.")
        if include_ids and signal.get("id"):
            lines.append(f"   - Signal ID: `{signal.get('id')}`")
        if signal.get("query"):
            lines.append(f"   - Query: {_md_inline(signal['query'])}")
        if signal.get("coarse_filter"):
            lines.append(f"   - Coarse filter: {_md_inline(_format_mapping(signal['coarse_filter']))}")
        for evidence in (signal.get("evidence") or [])[:3]:
            lines.append(f"   - {_render_evidence_link(evidence)}")
    total = _int_or_none(total_count)
    displayed = min(len(signals), limit)
    remaining = (total - displayed) if total is not None else (len(signals) - displayed)
    if remaining > 0:
        lines.append(f"- Plus {remaining} more candidate(s) in the machine-readable summary.")
    return lines


def _summarize_signal(signal: dict[str, Any]) -> dict[str, Any]:
    routing = dict(signal.get("routing") or {})
    mechanical = dict(signal.get("mechanical_scores") or {})
    evidence = list(signal.get("evidence") or [])
    return {
        "id": signal.get("id"),
        "title": signal.get("title") or _first_evidence_value(evidence, "title"),
        "query": signal.get("query"),
        "lane": routing.get("lane"),
        "queue_priority": routing.get("queue_priority"),
        "decay_bucket": signal.get("decay_bucket") or mechanical.get("decay_bucket"),
        "profile_match": mechanical.get("profile_match"),
        "major_news": mechanical.get("major_news"),
        "momentum": mechanical.get("momentum"),
        "source_agreement": mechanical.get("source_agreement"),
        "coarse_filter": signal.get("coarse_filter") or signal.get("cheap_filter"),
        "evidence": [_summarize_evidence(item) for item in evidence],
    }


def _summarize_evidence(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "source": item.get("source"),
        "title": item.get("title"),
        "url": item.get("url"),
        "published_at": item.get("published_at"),
        "author": item.get("author"),
        "engagement": item.get("engagement") or {},
    }


def _artifact_paths(run_dir: Path) -> dict[str, Path]:
    return {
        "candidates": run_dir / "candidates.json",
        "detector_summary": run_dir / "summary.json",
        "commands": run_dir / "commands.log",
        "detector_stderr": run_dir / "detector.stderr.log",
        "filter_decisions": run_dir / "filter_decisions.json",
        "targeted_candidates": run_dir / "targeted_candidates.json",
        "final_report": run_dir / "final_report.md",
        "run_markdown": run_dir / "run.md",
    }


def _artifact_status(paths: dict[str, Path]) -> dict[str, dict[str, Any]]:
    status: dict[str, dict[str, Any]] = {}
    for name, path in paths.items():
        exists = path.exists()
        status[name] = {
            "path": str(path),
            "exists": exists,
            "bytes": path.stat().st_size if exists else 0,
        }
    return status


def _pipeline_status(paths: dict[str, Path]) -> list[dict[str, str]]:
    return [
        _stage("detector", paths["candidates"]),
        _stage("coarse_filter", paths["filter_decisions"]),
        _stage("filter_apply", paths["targeted_candidates"]),
        _stage("final_report", paths["final_report"]),
    ]


def _stage(name: str, path: Path) -> dict[str, str]:
    return {
        "stage": name,
        "status": "done" if path.exists() else "pending",
        "artifact": path.name,
    }


def _summarize_decisions(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"exists": False, "path": str(path)}
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        return {"exists": True, "path": str(path), "error": str(exc)}
    decisions = list(payload.get("decisions") or [])
    return {
        "exists": True,
        "path": str(path),
        "decision_count": len(decisions),
        "decisions_by_outcome": dict(Counter(str(item.get("decision") or "unknown") for item in decisions)),
        "decisions_by_reason": dict(Counter(str(item.get("reason") or "unknown") for item in decisions)),
    }


def _summarize_targeted_candidates(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"exists": False, "path": str(path)}
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        return {"exists": True, "path": str(path), "error": str(exc)}
    coarse_filter = _coarse_filter(payload)
    return {
        "exists": True,
        "path": str(path),
        "selected_signals": len(payload.get("signals") or []),
        "input_signals": coarse_filter.get("input_signal_count"),
        "rejected_signals": coarse_filter.get("rejected_count"),
    }


def _coarse_filter(payload: dict[str, Any]) -> dict[str, Any]:
    return dict(payload.get("coarse_filter") or payload.get("cheap_filter") or {})


def _summarize_final_report(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"exists": False, "path": str(path)}
    try:
        content = path.read_text(encoding="utf-8")
    except OSError as exc:
        return {"exists": True, "path": str(path), "error": str(exc)}
    return {
        "exists": True,
        "path": str(path),
        "bytes": len(content.encode("utf-8")),
        "content": content,
    }


def _profile_name(monitor: dict[str, Any]) -> str | None:
    profile = monitor.get("profile") or {}
    if isinstance(profile, dict):
        value = profile.get("name") or profile.get("company") or profile.get("client")
        return str(value) if value else None
    return None


def _first_evidence_value(evidence: list[dict[str, Any]], key: str) -> Any:
    for item in evidence:
        if item.get(key):
            return item.get(key)
    return None


def _count_lanes(signals: list[dict[str, Any]]) -> dict[str, int]:
    return dict(Counter(str((signal.get("routing") or {}).get("lane") or "unknown") for signal in signals))


def _count_evidence_sources(signals: list[dict[str, Any]]) -> dict[str, int]:
    counts: Counter[str] = Counter()
    for signal in signals:
        for item in signal.get("evidence") or []:
            counts[str(item.get("source") or "unknown")] += 1
    return dict(counts)


def _queue_priority(signal: dict[str, Any]) -> float:
    try:
        return float((signal.get("routing") or {}).get("queue_priority") or 0.0)
    except (TypeError, ValueError):
        return 0.0


def _format_mapping(mapping: dict[str, Any]) -> str:
    return ", ".join(f"{key}={value}" for key, value in mapping.items())


def _parse_report_json(content: str) -> Any:
    text = content.strip()
    if not text:
        return None
    candidates = [text]
    if text.startswith("```"):
        lines = text.splitlines()
        if lines and lines[0].startswith("```"):
            body = lines[1:]
            if body and body[-1].startswith("```"):
                body = body[:-1]
            candidates.append("\n".join(body).strip())
    first = text.find("{")
    last = text.rfind("}")
    if first >= 0 and last > first:
        candidates.append(text[first:last + 1])
    for candidate in candidates:
        try:
            return json.loads(candidate)
        except json.JSONDecodeError:
            continue
    return None


def _render_evidence_link(item: dict[str, Any]) -> str:
    source = _md_inline(_label(item.get("source") or "source"))
    title = _md_inline(item.get("title") or "(no title)")
    url = str(item.get("url") or "").strip()
    published = item.get("published_at")
    suffix = f" ({_md_inline(published)})" if published else ""
    if url:
        return f"{source}: [{_escape_link_text(title)}]({url}){suffix}"
    return f"{source}: {title}{suffix}"


def _format_datetime(value: Any) -> str:
    if not value:
        return "(unknown)"
    text = str(value)
    raw = text[:-1] + "+00:00" if text.endswith("Z") else text
    try:
        parsed = datetime.fromisoformat(raw)
    except ValueError:
        return text
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")


def _format_list(values: list[Any], *, limit: int) -> str:
    clean = [_md_inline(value) for value in values if _md_inline(value)]
    if len(clean) <= limit:
        return ", ".join(clean)
    return ", ".join(clean[:limit]) + f", plus {len(clean) - limit} more"


def _int_or_none(value: Any) -> int | None:
    if value is None:
        return None
    try:
        return int(value)
    except (TypeError, ValueError):
        return None


def _label(value: Any) -> str:
    text = _md_inline(value)
    if not text:
        return ""
    return text.replace("_", " ").replace("-", " ")


def _escape_link_text(value: str) -> str:
    return value.replace("[", "\\[").replace("]", "\\]")


def _render_table(rows: Any) -> list[str]:
    normalized = list(rows)
    if not normalized:
        return ["- (none)"]
    lines = ["| key | value |", "|---|---|"]
    for key, value in normalized:
        lines.append(f"| {_md_cell(key)} | {_md_cell(value)} |")
    return lines


def _md_cell(value: Any) -> str:
    return _md_inline(value).replace("|", "\\|")


def _md_inline(value: Any) -> str:
    return " ".join(str(value).split())


def _fmt(value: Any) -> str:
    if value is None:
        return "-"
    try:
        return f"{float(value):.3g}"
    except (TypeError, ValueError):
        return str(value)


if __name__ == "__main__":
    raise SystemExit(main())
