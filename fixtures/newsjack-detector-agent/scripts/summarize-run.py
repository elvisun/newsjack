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
        "cheap_filter": payload.get("cheap_filter") or {},
        "cheap_filter_file": _summarize_decisions(artifact_paths["filter_decisions"]),
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
    lines: list[str] = []

    lines.append("# Newsjack Run")
    lines.append("")
    lines.append("## Result")
    final_report = summary.get("final_report_file") or {}
    final_text = final_report.get("content")
    if final_text:
        lines.append(final_text.rstrip())
        lines.append("")
    else:
        lines.append("_No final LLM report yet. Run the cheap filter, apply it, then write `final_report.md` and rerender this file._")
        lines.append("")

    lines.append("## Pipeline")
    lines.extend(_render_table([
        (stage.get("stage"), f"{stage.get('status')} - {stage.get('artifact')}")
        for stage in summary.get("pipeline") or []
    ]))
    lines.append("")

    lines.append("## Run Context")
    lines.append(f"- **Input:** `{summary.get('input_path')}`")
    lines.append(f"- **Profile:** {_md_inline(monitor.get('profile_name') or '(unknown)')}")
    queries = monitor.get("queries") or []
    lines.append(f"- **Queries:** {_md_inline(', '.join(queries) if queries else '(none)')}")
    lines.append(f"- **Sources used:** `{', '.join(monitor.get('sources_used') or [])}`")

    lines.append("")
    lines.append("## Detector Counts")
    lines.extend(_render_table([
        ("scored", counts.get("total_scored_signals")),
        ("selected_unique", counts.get("selected_unique_signals")),
        ("debug_rows", counts.get("debug_all_scored_rows")),
        ("debug_selected_rows", counts.get("debug_selected_rows")),
        ("debug_unselected_rows", counts.get("debug_unselected_rows")),
        ("debug_duplicate_rows", counts.get("debug_duplicate_scored_rows")),
        ("source_errors", counts.get("source_errors")),
    ]))

    cheap_summary = summary.get("cheap_filter_file") or {}
    if cheap_summary.get("exists"):
        lines.append("")
        lines.append("## Cheap Filter")
        lines.extend(_render_table([
            ("decisions", cheap_summary.get("decision_count")),
            *[
                (f"decision.{key}", value)
                for key, value in (cheap_summary.get("decisions_by_outcome") or {}).items()
            ],
            *[
                (f"reason.{key}", value)
                for key, value in (cheap_summary.get("decisions_by_reason") or {}).items()
            ],
        ]))

    targeted_summary = summary.get("targeted_candidates_file") or {}
    if targeted_summary.get("exists"):
        lines.append("")
        lines.append("## Targeted Candidates")
        lines.extend(_render_table([
            ("selected_signals", targeted_summary.get("selected_signals")),
            ("input_signals", targeted_summary.get("input_signals")),
            ("rejected_signals", targeted_summary.get("rejected_signals")),
        ]))

    if selection:
        lines.append("")
        lines.append("## Selection")
        lines.extend(_render_table(selection.items()))

    lines.append("")
    lines.append("## Lanes")
    lane_rows: list[tuple[str, Any]] = []
    for group in ("scored", "emitted", "dropped_debug"):
        for lane, count in (lanes.get(group) or {}).items():
            lane_rows.append((f"{group}.{lane}", count))
    lines.extend(_render_table(lane_rows))

    lines.append("")
    lines.append("## Sources")
    source_rows: list[tuple[str, Any]] = []
    for source, count in (sources.get("evidence_by_source") or {}).items():
        source_rows.append((f"evidence.{source}", count))
    for source, error in (sources.get("source_errors") or {}).items():
        source_rows.append((f"error.{source}", error))
    lines.extend(_render_table(source_rows))

    hygiene = summary.get("hygiene_rejections") or {}
    if hygiene:
        lines.append("")
        lines.append("## Hygiene Rejections")
        lines.extend(_render_table(hygiene.items()))

    lines.append("")
    lines.append("## Selected Signals Before LLM")
    lines.extend(_render_signal_list(summary.get("top_signals") or []))

    dropped = summary.get("top_dropped_signals") or []
    if dropped:
        lines.append("")
        lines.append("## Top Dropped Signals Before LLM")
        lines.extend(_render_signal_list(dropped))

    return "\n".join(lines).rstrip() + "\n"


def _render_signal_list(signals: list[dict[str, Any]]) -> list[str]:
    if not signals:
        return ["- (none)"]
    lines: list[str] = []
    for index, signal in enumerate(signals, start=1):
        title = _md_inline(signal.get("title") or "(untitled)")
        lines.append(
            f"{index}. **{title}**  \n"
            f"   `id={signal.get('id')}` `lane={signal.get('lane')}` "
            f"`queue={_fmt(signal.get('queue_priority'))}` "
            f"`profile={_fmt(signal.get('profile_match'))}` "
            f"`major={_fmt(signal.get('major_news'))}`"
        )
        if signal.get("query"):
            lines.append(f"   - Query: `{_md_inline(signal['query'])}`")
        if signal.get("cheap_filter"):
            lines.append(f"   - Cheap filter: {_md_inline(_format_mapping(signal['cheap_filter']))}")
        for evidence in (signal.get("evidence") or [])[:3]:
            source = _md_inline(evidence.get("source") or "source")
            ev_title = _md_inline(evidence.get("title") or "(no title)")
            url = evidence.get("url") or ""
            if url:
                lines.append(f"   - {source}: {ev_title} - <{url}>")
            else:
                lines.append(f"   - {source}: {ev_title}")
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
        "cheap_filter": signal.get("cheap_filter"),
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
        _stage("cheap_filter", paths["filter_decisions"]),
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
    cheap_filter = payload.get("cheap_filter") or {}
    return {
        "exists": True,
        "path": str(path),
        "selected_signals": len(payload.get("signals") or []),
        "input_signals": cheap_filter.get("input_signal_count"),
        "rejected_signals": cheap_filter.get("rejected_count"),
    }


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
