#!/usr/bin/env python3
"""Deterministic utilities for the AI visibility writing pilot."""

from __future__ import annotations

import hashlib
import json
import math
import os
import re
import statistics
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Iterable
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit


TRACKING_PARAMS = {
    "fbclid", "gclid", "mc_cid", "mc_eid", "ref", "ref_src",
    "utm_campaign", "utm_content", "utm_medium", "utm_source", "utm_term",
}

PHASE_LIMITS = {
    "calibration": 1.0,
    "main": 4.82,
    "repeats": 2.0,
    "fresh": 2.0,
    "reserve": 0.18,
}
TOTAL_LIMIT = 10.0
PRECALIBRATION_CEILINGS = {"google": 0.0012, "chatgpt": 0.04}


class PilotError(RuntimeError):
    pass


class RetryableResponse(PilotError):
    pass


class BudgetError(PilotError):
    pass


def canonical_json(value: Any) -> str:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False)


def sha256_text(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def canonicalize_url(url: str) -> str:
    """Canonicalize for cross-query dedupe without pretending URLs are identical pages."""
    if not isinstance(url, str) or not url.strip():
        return ""
    raw = url.strip()
    if "://" not in raw:
        raw = "https://" + raw
    parts = urlsplit(raw)
    scheme = (parts.scheme or "https").lower()
    host = (parts.hostname or "").lower().rstrip(".")
    if not host:
        return ""
    try:
        port = parts.port
    except ValueError:
        return ""
    if port and not ((scheme == "http" and port == 80) or (scheme == "https" and port == 443)):
        host = f"{host}:{port}"
    path = re.sub(r"/{2,}", "/", parts.path or "/")
    if path != "/":
        path = path.rstrip("/")
    query = [
        (key, value)
        for key, value in parse_qsl(parts.query, keep_blank_values=True)
        if key.lower() not in TRACKING_PARAMS and not key.lower().startswith("utm_")
    ]
    query.sort()
    return urlunsplit((scheme, host, path, urlencode(query, doseq=True), ""))


def walk(value: Any) -> Iterable[Any]:
    yield value
    if isinstance(value, dict):
        for child in value.values():
            yield from walk(child)
    elif isinstance(value, list):
        for child in value:
            yield from walk(child)


def _task(raw: dict[str, Any]) -> dict[str, Any]:
    if raw.get("status_code") != 20000:
        raise PilotError(f"top-level API error {raw.get('status_code')}")
    tasks = raw.get("tasks")
    if not isinstance(tasks, list) or len(tasks) != 1 or not isinstance(tasks[0], dict):
        raise PilotError("expected exactly one task")
    task = tasks[0]
    code = int(task.get("status_code") or 0)
    if (20000 <= code < 30000 and code != 20000) or code in {40601, 40602}:
        raise RetryableResponse(f"task pending/retryable {code}")
    if code != 20000:
        raise PilotError(f"task API error {code}")
    return task


def _task_cost(raw: dict[str, Any], task: dict[str, Any]) -> float:
    candidates = [task.get("cost"), raw.get("cost")]
    for candidate in candidates:
        if isinstance(candidate, (int, float)) and candidate >= 0:
            return round(float(candidate), 10)
    raise PilotError("successful response has no numeric cost")


def _url_record(item: dict[str, Any], default_label: str) -> dict[str, Any] | None:
    url = item.get("url") or item.get("source_url") or item.get("link")
    canonical = canonicalize_url(url) if isinstance(url, str) else ""
    if not canonical:
        return None
    return {
        "url": url,
        "canonical_url": canonical,
        "domain": urlsplit(canonical).hostname,
        "title": item.get("title") or item.get("source_name") or "",
        "label": default_label,
    }


def parse_google_response(raw: dict[str, Any], request: dict[str, Any]) -> dict[str, Any]:
    task = _task(raw)
    result = task.get("result") or []
    if not isinstance(result, list) or not result:
        raise PilotError("Google task has no result")
    roots = result
    overviews = [node for node in walk(roots) if isinstance(node, dict) and node.get("type") in {"ai_overview", "ai_overview_item"}]
    overview = overviews[0] if overviews else None
    organic: list[dict[str, Any]] = []
    for node in walk(roots):
        if isinstance(node, dict) and node.get("type") == "organic":
            record = _url_record(node, "organic_not_ai_cited")
            if record:
                record["organic_position"] = node.get("rank_absolute") or node.get("rank_group")
                organic.append(record)
    references: list[dict[str, Any]] = []
    if overview:
        for node in walk(overview):
            if not isinstance(node, dict):
                continue
            node_type = str(node.get("type") or "")
            if node_type in {"ai_overview_reference", "reference", "source"} or any(k in node for k in ("source_url", "url")):
                record = _url_record(node, "inline_cited_not_named")
                if record:
                    references.append(record)
    reference_urls = {r["canonical_url"] for r in references}
    for record in organic:
        if record["canonical_url"] in reference_urls:
            record["label"] = "inline_cited_not_named"
    references = dedupe_records(references)
    organic = dedupe_records(organic, keep_position=True)
    if not overview:
        outcome = "no_ai_answer"
    elif not references:
        outcome = "ai_answer_no_citations"
    else:
        outcome = "has_citations"
    return {
        "request_tag": request["request_tag"],
        "paired_unit_id": request["paired_unit_id"],
        "phase": request["phase"],
        "platform": "google",
        "query": request["query"],
        "topic_family": request["topic_family"],
        "intent": request["intent"],
        "observed_at": utc_now(),
        "cost_usd": _task_cost(raw, task),
        "outcome": outcome,
        "answer_text": collect_text(overview) if overview else "",
        "citations": references,
        "organic_results": organic,
        "raw_task_id": task.get("id"),
    }


def parse_chatgpt_response(raw: dict[str, Any], request: dict[str, Any]) -> dict[str, Any]:
    task = _task(raw)
    result = task.get("result") or []
    if not isinstance(result, list) or not result or not isinstance(result[0], dict):
        raise PilotError("ChatGPT task has no result")
    response = result[0]
    answer_parts: list[str] = []
    references: list[dict[str, Any]] = []
    for node in walk(response.get("items") or []):
        if not isinstance(node, dict):
            continue
        if node.get("type") in {"text", "output_text"} and isinstance(node.get("text"), str):
            answer_parts.append(node["text"])
        annotations = node.get("annotations")
        if isinstance(annotations, list):
            for annotation in annotations:
                if isinstance(annotation, dict):
                    record = _url_record(annotation, "inline_cited_not_named")
                    if record:
                        references.append(record)
    answer = "\n".join(dict.fromkeys(answer_parts)).strip()
    references = dedupe_records(references)
    for record in references:
        title = str(record.get("title") or "").strip().lower()
        domain = str(record.get("domain") or "").split(".")[0].lower()
        if (title and title in answer.lower()) or (domain and len(domain) > 3 and domain in answer.lower()):
            record["label"] = "answer_mentioned_and_cited"
    if not answer:
        outcome = "no_ai_answer"
    elif not references:
        outcome = "ai_answer_no_citations"
    else:
        outcome = "has_citations"
    money_spent = response.get("money_spent")
    if money_spent is not None and not isinstance(money_spent, (int, float)):
        raise PilotError("ChatGPT money_spent is not numeric")
    return {
        "request_tag": request["request_tag"],
        "paired_unit_id": request["paired_unit_id"],
        "phase": request["phase"],
        "platform": "chatgpt",
        "query": request["query"],
        "topic_family": request["topic_family"],
        "intent": request["intent"],
        "observed_at": utc_now(),
        "cost_usd": _task_cost(raw, task),
        "money_spent_usd": round(float(money_spent), 10) if money_spent is not None else None,
        "outcome": outcome,
        "answer_text": answer,
        "citations": references,
        "organic_results": [],
        "resolved_model": response.get("model_name"),
        "input_tokens": response.get("input_tokens"),
        "output_tokens": response.get("output_tokens"),
        "web_search_used": response.get("web_search"),
        "raw_task_id": task.get("id"),
    }


def collect_text(value: Any) -> str:
    parts: list[str] = []
    for node in walk(value):
        if isinstance(node, dict):
            for key in ("text", "markdown"):
                text = node.get(key)
                if isinstance(text, str) and text.strip():
                    parts.append(text.strip())
    return "\n".join(dict.fromkeys(parts))


def dedupe_records(records: list[dict[str, Any]], keep_position: bool = False) -> list[dict[str, Any]]:
    seen: dict[str, dict[str, Any]] = {}
    for record in records:
        key = record["canonical_url"]
        if key not in seen:
            seen[key] = record
        elif keep_position:
            old = seen[key].get("organic_position")
            new = record.get("organic_position")
            if isinstance(new, int) and (not isinstance(old, int) or new < old):
                seen[key] = record
    return list(seen.values())


def build_page_index(observations: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Deduplicate pages across queries while preserving every citation event."""
    pages: dict[str, dict[str, Any]] = {}
    for observation in observations:
        for surface in ("citations", "organic_results"):
            for record in observation.get(surface, []):
                canonical = record.get("canonical_url") or canonicalize_url(record.get("url", ""))
                if not canonical:
                    continue
                page = pages.setdefault(canonical, {
                    "canonical_url": canonical,
                    "domain": urlsplit(canonical).hostname,
                    "titles": [],
                    "events": [],
                })
                title = record.get("title")
                if title and title not in page["titles"]:
                    page["titles"].append(title)
                page["events"].append({
                    "request_tag": observation.get("request_tag"),
                    "paired_unit_id": observation.get("paired_unit_id"),
                    "platform": observation.get("platform"),
                    "label": record.get("label"),
                    "organic_position": record.get("organic_position"),
                })
    return sorted(pages.values(), key=lambda row: row["canonical_url"])


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def read_jsonl(path: Path) -> list[dict[str, Any]]:
    if not path.exists():
        return []
    rows: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as handle:
        for line_number, line in enumerate(handle, 1):
            if not line.strip():
                continue
            try:
                value = json.loads(line)
            except json.JSONDecodeError as exc:
                raise PilotError(f"invalid JSONL at line {line_number}") from exc
            if not isinstance(value, dict):
                raise PilotError(f"non-object JSONL at line {line_number}")
            rows.append(value)
    return rows


def verify_ledger(path: Path) -> tuple[float, dict[str, float]]:
    previous = "GENESIS"
    total = 0.0
    phases = {phase: 0.0 for phase in PHASE_LIMITS}
    seen_tags: set[str] = set()
    unknown_terminals: dict[str, dict[str, Any]] = {}
    reconciled_tags: set[str] = set()
    for index, event in enumerate(read_jsonl(path), 1):
        supplied_hash = event.get("event_hash")
        body = dict(event)
        body.pop("event_hash", None)
        expected = sha256_text(canonical_json(body))
        if supplied_hash != expected or body.get("prev_hash") != previous:
            raise PilotError(f"cost ledger chain failure at record {index}")
        previous = supplied_hash
        if body.get("terminal"):
            tag = body.get("request_tag")
            if not isinstance(tag, str) or tag in seen_tags:
                raise PilotError(f"duplicate or missing terminal request tag at record {index}")
            seen_tags.add(tag)
            cost = body.get("cost_usd")
            if body.get("status") == "success" and (not isinstance(cost, (int, float)) or cost < 0):
                raise PilotError(f"successful terminal event lacks cost at record {index}")
            if not isinstance(cost, (int, float)):
                unknown_terminals[tag] = body
            if isinstance(cost, (int, float)):
                if cost < 0:
                    raise PilotError(f"negative terminal cost at record {index}")
                phase = body.get("phase")
                if phase not in PHASE_LIMITS:
                    raise PilotError(f"unknown phase at record {index}")
                total += float(cost)
                phases[phase] += float(cost)
        target = body.get("reconciles_request_tag")
        if target is not None:
            if body.get("terminal") or not isinstance(target, str) or target not in unknown_terminals or target in reconciled_tags:
                raise PilotError(f"invalid cost reconciliation at record {index}")
            original = unknown_terminals[target]
            cost = body.get("cost_usd")
            phase = body.get("phase")
            if body.get("status") not in {"error", "success"} or not isinstance(cost, (int, float)) or cost < 0:
                raise PilotError(f"invalid cost reconciliation value at record {index}")
            if phase != original.get("phase") or body.get("platform") != original.get("platform") or phase not in PHASE_LIMITS:
                raise PilotError(f"cost reconciliation provenance mismatch at record {index}")
            total += float(cost)
            phases[phase] += float(cost)
            reconciled_tags.add(target)
    for phase, spent in phases.items():
        if spent > PHASE_LIMITS[phase] + 1e-9:
            raise PilotError(f"phase budget exceeded: {phase}")
    if total > TOTAL_LIMIT + 1e-9:
        raise PilotError("total budget exceeded")
    return round(total, 10), {k: round(v, 10) for k, v in phases.items()}


def append_ledger(path: Path, event: dict[str, Any]) -> dict[str, Any]:
    rows = read_jsonl(path)
    total, phases = verify_ledger(path)
    target = event.get("reconciles_request_tag")
    if target is not None:
        terminals = {
            row.get("request_tag"): row for row in rows
            if row.get("terminal") and isinstance(row.get("request_tag"), str)
        }
        reconciled = {row.get("reconciles_request_tag") for row in rows if row.get("reconciles_request_tag")}
        original = terminals.get(target)
        cost = event.get("cost_usd")
        phase = event.get("phase")
        if event.get("terminal") or not original or target in reconciled or isinstance(original.get("cost_usd"), (int, float)):
            raise PilotError("invalid or duplicate cost reconciliation")
        if event.get("status") not in {"error", "success"} or not isinstance(cost, (int, float)) or cost < 0:
            raise PilotError("invalid cost reconciliation value")
        if phase != original.get("phase") or event.get("platform") != original.get("platform") or phase not in PHASE_LIMITS:
            raise PilotError("cost reconciliation provenance mismatch")
        if phases[phase] + float(cost) > PHASE_LIMITS[phase] + 1e-9 or total + float(cost) > TOTAL_LIMIT + 1e-9:
            raise PilotError("cost reconciliation exceeds budget")
    if event.get("terminal"):
        tag = event.get("request_tag")
        if not isinstance(tag, str) or any(row.get("terminal") and row.get("request_tag") == tag for row in rows):
            raise PilotError("duplicate or missing terminal request tag")
        cost = event.get("cost_usd")
        phase = event.get("phase")
        if isinstance(cost, (int, float)):
            if cost < 0 or phase not in PHASE_LIMITS:
                raise PilotError("invalid terminal cost or phase")
            if phases[phase] + float(cost) > PHASE_LIMITS[phase] + 1e-9:
                raise PilotError(f"phase budget exceeded: {phase}")
            if total + float(cost) > TOTAL_LIMIT + 1e-9:
                raise PilotError("total budget exceeded")
    previous = rows[-1].get("event_hash") if rows else "GENESIS"
    body = dict(event)
    body["prev_hash"] = previous
    body.setdefault("recorded_at", utc_now())
    body["event_hash"] = sha256_text(canonical_json(body))
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(canonical_json(body) + "\n")
        handle.flush()
        os.fsync(handle.fileno())
    verify_ledger(path)
    return body


def successful_tags(path: Path) -> set[str]:
    return {
        tag for tag, row in terminal_events(path).items()
        if row.get("status") == "success"
    }


def terminal_events(path: Path) -> dict[str, dict[str, Any]]:
    events = {
        str(row["request_tag"]): row
        for row in read_jsonl(path)
        if row.get("terminal") and isinstance(row.get("request_tag"), str)
    }
    for row in read_jsonl(path):
        target = row.get("reconciles_request_tag")
        if isinstance(target, str) and target in events:
            events[target] = {
                **events[target], "status": row.get("status"),
                "cost_usd": row.get("cost_usd"),
                "reconciliation_event_hash": row.get("event_hash"),
                "cost_source": row.get("cost_source"),
            }
    return events


def unresolved_unknown_tags(path: Path) -> set[str]:
    return {
        tag for tag, row in terminal_events(path).items()
        if not isinstance(row.get("cost_usd"), (int, float))
    }


def observed_p95(path: Path, platform: str) -> float | None:
    costs = sorted(
        float(row["cost_usd"])
        for row in terminal_events(path).values()
        if row.get("status") == "success"
        and row.get("phase") == "calibration" and row.get("platform") == platform
        and isinstance(row.get("cost_usd"), (int, float))
    )
    if not costs:
        return None
    index = max(0, math.ceil(0.95 * len(costs)) - 1)
    return costs[index]


@dataclass(frozen=True)
class Projection:
    total_actual: float
    phase_actual: float
    projected: float
    unit_cost: float
    source: str


def check_budget(path: Path, phase: str, platform: str, count: int) -> Projection:
    if phase not in PHASE_LIMITS or platform not in PRECALIBRATION_CEILINGS or count < 0:
        raise BudgetError("invalid budget request")
    if unresolved_unknown_tags(path):
        raise BudgetError("collection stopped: a terminal request has unknown cost")
    total, phases = verify_ledger(path)
    p95 = observed_p95(path, platform)
    source = "calibration_p95" if p95 is not None else "preregistered_precalibration_ceiling"
    unit = p95 if p95 is not None else PRECALIBRATION_CEILINGS[platform]
    projected = round(unit * count, 10)
    if phases[phase] + projected > PHASE_LIMITS[phase] + 1e-9:
        raise BudgetError(f"projected {phase} spend exceeds allocation")
    if total + projected > TOTAL_LIMIT + 1e-9:
        raise BudgetError("projected total spend exceeds $10")
    if phase == "fresh" and phases["fresh"] == 0 and TOTAL_LIMIT - total < PHASE_LIMITS["fresh"] - 1e-9:
        raise BudgetError("fresh run cancelled: less than its full $2 allocation remains")
    return Projection(total, phases[phase], projected, unit, source)


def check_batch_budget(path: Path, phase: str, counts: dict[str, int]) -> dict[str, Projection]:
    projections = {
        platform: check_budget(path, phase, platform, count)
        for platform, count in counts.items() if count
    }
    if not projections:
        return {}
    first = next(iter(projections.values()))
    combined = sum(item.projected for item in projections.values())
    if first.phase_actual + combined > PHASE_LIMITS[phase] + 1e-9:
        raise BudgetError(f"combined projected {phase} spend exceeds allocation")
    if first.total_actual + combined > TOTAL_LIMIT + 1e-9:
        raise BudgetError("combined projected total spend exceeds $10")
    return projections


NUMBER_RE = re.compile(r"(?<!\w)(?:[$£€]\s*)?(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?\s*(?:%|percent|million|billion|thousand|days?|weeks?|months?|years?)?", re.I)
DATE_RE = re.compile(r"\b(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+\d{1,2}(?:,\s*\d{4})?|\b\d{4}-\d{2}-\d{2}\b|\b(?:19|20)\d{2}\b", re.I)
QUOTE_RE = re.compile(r"(?:\"[^\"\n]+\"|“[^”\n]+”|(?<!\w)'[^'\n]{8,}'(?!\w))")
URL_RE = re.compile(r"https?://[^\s)>\]]+")
ENTITY_RE = re.compile(r"\b(?:[A-Z][a-z]+(?:[-'][A-Z]?[a-z]+)?)(?:\s+(?:[A-Z][a-z]+|[A-Z]{2,}|of|the)){0,3}\b")


def fact_ledger(text: str) -> dict[str, list[str]]:
    prose = re.sub(r"(?m)^\s*\d+[.)]\s+", "", text)
    prose = re.sub(r"\[\d+\]", "", prose)
    urls = list(dict.fromkeys(match.group(0).strip() for match in URL_RE.finditer(prose)))
    prose_without_urls = URL_RE.sub("", prose)
    def unique(pattern: re.Pattern[str]) -> list[str]:
        return list(dict.fromkeys(match.group(0).strip() for match in pattern.finditer(prose_without_urls)))
    entities = [value for value in unique(ENTITY_RE) if value.lower() not in {"the", "a", "an", "on", "in", "this"}]
    quotes = list(dict.fromkeys(
        re.sub(r"\s+", " ", match.group(0)[1:-1].strip()).rstrip(".,")
        for match in QUOTE_RE.finditer(prose_without_urls)
        if len(re.findall(r"\w+", match.group(0)[1:-1])) >= 3
    ))
    return {
        "numbers": unique(NUMBER_RE),
        "dates": unique(DATE_RE),
        "quotes": quotes,
        "urls": urls,
        "entities": entities,
    }


def compare_fact_ledgers(original: str, rewrite: str) -> dict[str, Any]:
    before = fact_ledger(original)
    after = fact_ledger(rewrite)
    missing: dict[str, list[str]] = {}
    added: dict[str, list[str]] = {}
    for key in before:
        missing[key] = [value for value in before[key] if value not in after[key]]
        added[key] = [value for value in after[key] if value not in before[key]]
    hard_categories = ("numbers", "dates", "quotes", "urls")
    pass_hard = not any(missing[key] or added[key] for key in hard_categories)
    return {"pass": pass_hard, "missing": missing, "added": added, "before": before, "after": after}


def wilson(successes: float, n: int, z: float = 1.96) -> tuple[float, float]:
    if n <= 0:
        return (math.nan, math.nan)
    p = successes / n
    denominator = 1 + z * z / n
    center = (p + z * z / (2 * n)) / denominator
    spread = z * math.sqrt((p * (1 - p) + z * z / (4 * n)) / n) / denominator
    return max(0.0, center - spread), min(1.0, center + spread)


def percentile(values: list[float], q: float) -> float | None:
    if not values:
        return None
    ordered = sorted(values)
    index = max(0, min(len(ordered) - 1, math.ceil(q * len(ordered)) - 1))
    return ordered[index]


def composite(metrics: dict[str, Any]) -> float:
    keys = ("audit_factor_f1", "blind_quality_win_rate", "paired_citation_selection_win_rate", "readability_pass_rate")
    if any(not isinstance(metrics.get(key), (int, float)) or not 0 <= float(metrics[key]) <= 1 for key in keys):
        raise PilotError("missing or invalid score component")
    return round(100 * (
        0.35 * metrics["audit_factor_f1"]
        + 0.30 * metrics["blind_quality_win_rate"]
        + 0.25 * metrics["paired_citation_selection_win_rate"]
        + 0.10 * metrics["readability_pass_rate"]
    ), 2)
