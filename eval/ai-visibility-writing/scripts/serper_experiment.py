#!/usr/bin/env python3
"""Collect and analyze the frozen Serper instrumentation experiment."""

from __future__ import annotations

import argparse
import concurrent.futures
import hashlib
import json
import os
import re
import statistics
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from collections import Counter, defaultdict
from datetime import datetime, timezone
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
PROTOCOL_PATH = ROOT / "serper-experiment-protocol.json"
RUN_DIR = ROOT / "runs" / "serper-25"
API = "https://google.serper.dev/search"
TRACKING_KEYS = {
    "fbclid", "gclid", "msclkid", "srsltid", "ved", "ei", "oq", "sa",
    "source", "sourceid", "ref", "ref_", "mc_cid", "mc_eid",
}
EXPLICIT_AI_KEYS = {
    "aioverview", "ai_overview", "generativeanswer", "generative_answer",
    "aimode", "ai_mode", "aicitations", "ai_citations",
}


def canonical_json(value: object) -> str:
    return json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def percentile(values: list[float], fraction: float) -> float | None:
    if not values:
        return None
    ordered = sorted(values)
    index = (len(ordered) - 1) * fraction
    lower = int(index)
    upper = min(lower + 1, len(ordered) - 1)
    weight = index - lower
    return ordered[lower] * (1 - weight) + ordered[upper] * weight


def normalize_url(value: str) -> str:
    if not isinstance(value, str) or not value.strip():
        return ""
    parsed = urllib.parse.urlsplit(value.strip())
    host = (parsed.hostname or "").lower()
    if host.startswith("www."):
        host = host[4:]
    if not host:
        return ""
    port = f":{parsed.port}" if parsed.port and parsed.port not in {80, 443} else ""
    path = re.sub(r"/{2,}", "/", parsed.path or "/")
    if path != "/":
        path = path.rstrip("/")
    query = []
    for key, item in urllib.parse.parse_qsl(parsed.query, keep_blank_values=True):
        lowered = key.lower()
        if lowered.startswith("utm_") or lowered in TRACKING_KEYS:
            continue
        query.append((key, item))
    encoded = urllib.parse.urlencode(sorted(query))
    return urllib.parse.urlunsplit(("https", host + port, path, encoded, ""))


def domain_of(value: str) -> str:
    normalized = normalize_url(value)
    return urllib.parse.urlsplit(normalized).hostname or ""


def jaccard(left: set[str], right: set[str]) -> float:
    union = left | right
    return len(left & right) / len(union) if union else 1.0


def env_values(path: Path) -> dict[str, str]:
    result: dict[str, str] = {}
    for number, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            raise ValueError(f"invalid env line {number}")
        key, value = line.split("=", 1)
        key = key.strip().removeprefix("export ").strip()
        if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            raise ValueError(f"invalid env key on line {number}")
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
            value = value[1:-1]
        result[key] = value
    return result


def query_units(protocol: dict) -> list[dict]:
    seen: dict[str, dict] = {}
    for phase in protocol["inputs"]["manifest_phases"]:
        manifest = json.loads((ROOT / "manifests" / f"{phase}.json").read_text(encoding="utf-8"))
        for unit in manifest["units"]:
            query = unit["query"]
            if query not in seen:
                seen[query] = {
                    "query": query,
                    "topic_family": unit.get("topic_family"),
                    "intent": unit.get("intent"),
                    "first_phase": phase,
                }
    units = list(seen.values())
    for unit in units:
        digest = hashlib.sha256(unit["query"].encode()).hexdigest()
        unit["query_id"] = f"q-{digest[:12]}"
    expected = protocol["inputs"]["expected_unique_queries"]
    if len(units) != expected:
        raise ValueError(f"query count changed: expected {expected}, got {len(units)}")
    return units


def ordered_units(protocol: dict, wave: int) -> list[dict]:
    protocol_id = protocol["protocol_id"]
    return sorted(
        query_units(protocol),
        key=lambda item: hashlib.sha256(
            f"{protocol_id}|{wave}|{item['query']}".encode()
        ).hexdigest(),
    )


def request_one(unit: dict, wave: int, api_key: str) -> tuple[dict, dict]:
    payload = canonical_json({
        "q": unit["query"], "gl": "us", "hl": "en", "num": 10,
    }).encode()
    request = urllib.request.Request(API, data=payload, method="POST")
    request.add_header("X-API-KEY", api_key)
    request.add_header("Content-Type", "application/json")
    started = time.monotonic()
    last_error: Exception | None = None
    for attempt in range(3):
        try:
            with urllib.request.urlopen(request, timeout=45) as response:
                raw = json.load(response)
                headers = {key.lower(): value for key, value in response.headers.items()}
            if not isinstance(raw, dict) or not isinstance(raw.get("organic"), list):
                raise ValueError("Serper response lacks an organic array")
            elapsed_ms = round((time.monotonic() - started) * 1000)
            record = {
                "query_id": unit["query_id"],
                "wave": wave,
                "observed_at": utc_now(),
                "attempts": attempt + 1,
                "elapsed_ms": elapsed_ms,
                "organic_count": len(raw["organic"]),
                "top_level_keys": sorted(raw),
                "rate_limit_remaining": _integer(headers.get("x-ratelimit-remaining")),
                "response_credits": _number(raw.get("credits")),
            }
            return record, raw
        except (urllib.error.HTTPError, urllib.error.URLError, TimeoutError, ValueError) as exc:
            last_error = exc
            if isinstance(exc, urllib.error.HTTPError) and exc.code not in {408, 429, 500, 502, 503, 504}:
                break
            if attempt < 2:
                time.sleep(2 ** attempt)
    raise RuntimeError(f"Serper request failed for {unit['query_id']}: {type(last_error).__name__}")


def _integer(value: object) -> int | None:
    try:
        return int(value) if value is not None else None
    except (TypeError, ValueError):
        return None


def _number(value: object) -> float | None:
    return float(value) if isinstance(value, (int, float)) and not isinstance(value, bool) else None


def collect(wave: int, env_file: Path, workers: int) -> dict:
    protocol = json.loads(PROTOCOL_PATH.read_text(encoding="utf-8"))
    units = ordered_units(protocol, wave)
    values = env_values(env_file)
    api_key = os.environ.get("SERPER_API_KEY") or values.get("SERPER_API_KEY")
    if not api_key:
        raise ValueError("SERPER_API_KEY is absent")
    raw_dir = RUN_DIR / f"wave-{wave}" / "raw"
    raw_dir.mkdir(parents=True, exist_ok=True)
    ledger_path = RUN_DIR / "ledger.jsonl"
    existing = {}
    if ledger_path.exists():
        for line in ledger_path.read_text(encoding="utf-8").splitlines():
            if line.strip():
                row = json.loads(line)
                existing[(row["wave"], row["query_id"])] = row
    completed_total = len(existing)
    pending = [unit for unit in units if (wave, unit["query_id"]) not in existing]
    cap = protocol["serper_request"]["hard_successful_request_cap"]
    if completed_total + len(pending) > cap:
        raise ValueError("planned requests exceed the frozen successful-request cap")
    results: list[tuple[dict, dict]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        futures = {pool.submit(request_one, unit, wave, api_key): unit for unit in pending}
        for future in concurrent.futures.as_completed(futures):
            results.append(future.result())
    for record, raw in sorted(results, key=lambda item: item[0]["query_id"]):
        raw_path = raw_dir / f"{record['query_id']}.json"
        flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
        descriptor = os.open(raw_path, flags, 0o600)
        with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
            handle.write(canonical_json(raw) + "\n")
        record["response_sha256"] = hashlib.sha256(raw_path.read_bytes()).hexdigest()
        record["replacement_value_usd"] = 0.001
        with ledger_path.open("a", encoding="utf-8") as handle:
            handle.write(canonical_json(record) + "\n")
    summary = {
        "wave": wave,
        "new_successes": len(results),
        "total_successes": completed_total + len(results),
        "planned_wave_size": len(units),
        "min_rate_limit_remaining": min(
            (record["rate_limit_remaining"] for record, _ in results if record["rate_limit_remaining"] is not None),
            default=None,
        ),
    }
    print(canonical_json(summary))
    return summary


def _items(raw: dict) -> list[dict]:
    values = raw.get("organic") or []
    return [item for item in values[:10] if isinstance(item, dict)]


def _url_set(items: list[dict]) -> set[str]:
    return {url for item in items if (url := normalize_url(item.get("link") or item.get("url") or ""))}


def _domain_set(items: list[dict]) -> set[str]:
    return {domain_of(item.get("link") or item.get("url") or "") for item in items} - {""}


def load_observations() -> list[dict]:
    rows = []
    for phase in ["calibration", "main", "repeats", "fresh"]:
        path = ROOT / "runs" / phase / "observations.jsonl"
        if not path.exists():
            raise FileNotFoundError(f"required prior observation file is absent: {path}")
        rows.extend(json.loads(line) for line in path.read_text(encoding="utf-8").splitlines() if line.strip())
    return rows


def analyze() -> dict:
    protocol = json.loads(PROTOCOL_PATH.read_text(encoding="utf-8"))
    units = query_units(protocol)
    raw_by_wave: dict[int, dict[str, dict]] = {1: {}, 2: {}}
    for wave in (1, 2):
        for unit in units:
            path = RUN_DIR / f"wave-{wave}" / "raw" / f"{unit['query_id']}.json"
            raw_by_wave[wave][unit["query"]] = json.loads(path.read_text(encoding="utf-8"))
    ledger = [
        json.loads(line) for line in (RUN_DIR / "ledger.jsonl").read_text(encoding="utf-8").splitlines()
        if line.strip()
    ]
    observations = load_observations()
    by_query: dict[str, list[dict]] = defaultdict(list)
    for row in observations:
        by_query[row["query"]].append(row)

    wave_url_jaccards: list[float] = []
    wave_domain_jaccards: list[float] = []
    wave_intervals_seconds: list[float] = []
    wave_url_by_intent: dict[str, list[float]] = defaultdict(list)
    cross_url_jaccards: list[float] = []
    cross_domain_jaccards: list[float] = []
    citation_totals = Counter()
    citation_hits = Counter()
    matched_pairs = Counter()
    explicit_ai_responses = 0
    top_keys = Counter()
    organic_rows = 0

    for unit in units:
        query = unit["query"]
        wave_items = {wave: _items(raw_by_wave[wave][query]) for wave in (1, 2)}
        wave_urls = {wave: _url_set(wave_items[wave]) for wave in (1, 2)}
        wave_domains = {wave: _domain_set(wave_items[wave]) for wave in (1, 2)}
        wave_url_jaccards.append(jaccard(wave_urls[1], wave_urls[2]))
        wave_domain_jaccards.append(jaccard(wave_domains[1], wave_domains[2]))
        wave_url_by_intent[unit["intent"]].append(wave_url_jaccards[-1])
        for wave in (1, 2):
            raw = raw_by_wave[wave][query]
            organic_rows += len(wave_items[wave])
            top_keys.update(raw.keys())
            lowered = {str(key).lower() for key in raw}
            explicit_ai_responses += bool(lowered & EXPLICIT_AI_KEYS)
        for observation in by_query[query]:
            platform = observation["platform"]
            if platform == "google":
                organic = observation.get("organic_results") or []
                organic_urls = _url_set(organic)
                organic_domains = _domain_set(organic)
                for wave in (1, 2):
                    cross_url_jaccards.append(jaccard(wave_urls[wave], organic_urls))
                    cross_domain_jaccards.append(jaccard(wave_domains[wave], organic_domains))
                    matched_pairs["dataforseo_organic"] += 1
            citations = observation.get("citations") or []
            citation_urls = {normalize_url(item.get("canonical_url") or item.get("url") or "") for item in citations}
            citation_urls.discard("")
            citation_domains = {domain_of(url) for url in citation_urls} - {""}
            for wave in (1, 2):
                matched_pairs[platform] += 1
                citation_totals[f"{platform}_url"] += len(citation_urls)
                citation_hits[f"{platform}_url"] += len(citation_urls & wave_urls[wave])
                citation_totals[f"{platform}_domain"] += len(citation_domains)
                citation_hits[f"{platform}_domain"] += len(citation_domains & wave_domains[wave])

    ledger_by_unit = {(row["wave"], row["query_id"]): row for row in ledger}
    for unit in units:
        first = ledger_by_unit[(1, unit["query_id"])]["observed_at"]
        second = ledger_by_unit[(2, unit["query_id"])]["observed_at"]
        first_time = datetime.fromisoformat(first.replace("Z", "+00:00"))
        second_time = datetime.fromisoformat(second.replace("Z", "+00:00"))
        wave_intervals_seconds.append((second_time - first_time).total_seconds())

    def rate(key: str) -> float | None:
        total = citation_totals[key]
        return citation_hits[key] / total if total else None

    success_rate = len(ledger) / protocol["serper_request"]["planned_successful_requests"]
    organic_counts = [row["organic_count"] for row in ledger]
    median_wave_url = statistics.median(wave_url_jaccards)
    median_cross_url = statistics.median(cross_url_jaccards)
    operational_pass = success_rate >= 0.99 and len(ledger) == 276 and statistics.median(organic_counts) >= 8
    stable_pass = operational_pass and median_wave_url >= 0.70 and median_cross_url >= 0.50
    direct_ai_pass = explicit_ai_responses == len(ledger)
    raw_paths = sorted((RUN_DIR / "wave-1" / "raw").glob("*.json")) + sorted(
        (RUN_DIR / "wave-2" / "raw").glob("*.json")
    )
    raw_tree = "\n".join(
        f"{path.relative_to(ROOT)}:{sha256_file(path)}" for path in raw_paths
    )
    result = {
        "protocol_id": protocol["protocol_id"],
        "analyzed_at": max(row["observed_at"] for row in ledger),
        "sample": {
            "unique_queries": len(units),
            "serper_waves": 2,
            "serper_responses": len(ledger),
            "serper_organic_rows": organic_rows,
            "existing_ai_observations": len(observations),
            "matched_response_wave_pairs": dict(matched_pairs),
        },
        "budget": {
            "approved_cap_usd": 25.0,
            "incremental_cash_spend_usd": 0.0,
            "free_serper_credits_used": len(ledger),
            "response_credits_charged": sum(row.get("response_credits") or 0 for row in ledger),
            "starter_price_replacement_value_usd": round(len(ledger) * 0.001, 3),
            "incremental_dataforseo_spend_usd": 0.0,
        },
        "operational": {
            "success_rate": success_rate,
            "malformed_responses": 0,
            "organic_count_median": statistics.median(organic_counts),
            "latency_ms_median": statistics.median(row["elapsed_ms"] for row in ledger),
            "latency_ms_p95": percentile([row["elapsed_ms"] for row in ledger], 0.95),
            "minimum_reported_rate_limit_remaining": min(
                (row["rate_limit_remaining"] for row in ledger if row.get("rate_limit_remaining") is not None),
                default=None,
            ),
            "top_level_key_counts": dict(sorted(top_keys.items())),
        },
        "stability": {
            "within_serper_url_jaccard_median": median_wave_url,
            "within_serper_url_jaccard_mean": statistics.mean(wave_url_jaccards),
            "within_serper_url_jaccard_p25": percentile(wave_url_jaccards, 0.25),
            "within_serper_url_jaccard_p75": percentile(wave_url_jaccards, 0.75),
            "within_serper_url_jaccard_zero_share": sum(value == 0 for value in wave_url_jaccards) / len(wave_url_jaccards),
            "within_serper_url_jaccard_at_least_0_50_share": sum(value >= 0.50 for value in wave_url_jaccards) / len(wave_url_jaccards),
            "within_serper_url_jaccard_at_least_0_70_share": sum(value >= 0.70 for value in wave_url_jaccards) / len(wave_url_jaccards),
            "within_serper_domain_jaccard_median": statistics.median(wave_domain_jaccards),
            "within_serper_domain_jaccard_mean": statistics.mean(wave_domain_jaccards),
            "wave_interval_seconds_median": statistics.median(wave_intervals_seconds),
            "wave_interval_seconds_p95": percentile(wave_intervals_seconds, 0.95),
            "within_serper_url_jaccard_by_intent": {
                intent: {
                    "queries": len(values),
                    "median": statistics.median(values),
                    "mean": statistics.mean(values),
                }
                for intent, values in sorted(wave_url_by_intent.items())
            },
            "serper_vs_dataforseo_url_jaccard_median": median_cross_url,
            "serper_vs_dataforseo_url_jaccard_mean": statistics.mean(cross_url_jaccards),
            "serper_vs_dataforseo_domain_jaccard_median": statistics.median(cross_domain_jaccards),
            "serper_vs_dataforseo_domain_jaccard_mean": statistics.mean(cross_domain_jaccards),
        },
        "ai_join": {
            "explicit_ai_field_responses": explicit_ai_responses,
            "explicit_ai_field_rate": explicit_ai_responses / len(ledger),
            "google_aio_citation_url_recall_in_serper_top10": rate("google_url"),
            "google_aio_citation_domain_recall_in_serper_top10": rate("google_domain"),
            "chatgpt_citation_url_recall_in_serper_top10": rate("chatgpt_url"),
            "chatgpt_citation_domain_recall_in_serper_top10": rate("chatgpt_domain"),
            "citation_event_denominators": dict(citation_totals),
        },
        "decisions": {
            "operational_pass": operational_pass,
            "stable_organic_control_pass": stable_pass,
            "direct_ai_instrument_pass": direct_ai_pass,
            "lever_ranking_update_allowed": False,
        },
        "provenance": {
            "protocol_sha256": sha256_file(PROTOCOL_PATH),
            "ledger_sha256": sha256_file(RUN_DIR / "ledger.jsonl"),
            "raw_response_count": len(raw_paths),
            "raw_tree_sha256": hashlib.sha256(raw_tree.encode()).hexdigest(),
            "manifest_sha256": {
                phase: sha256_file(ROOT / "manifests" / f"{phase}.json")
                for phase in protocol["inputs"]["manifest_phases"]
            },
            "observation_sha256": {
                phase: sha256_file(ROOT / "runs" / phase / "observations.jsonl")
                for phase in protocol["inputs"]["manifest_phases"]
            },
        },
    }
    output = ROOT / "serper-experiment-analysis.json"
    output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(result, indent=2, sort_keys=True))
    return result


def self_test() -> None:
    assert normalize_url("http://WWW.Example.com/a/?utm_source=x&b=2#z") == "https://example.com/a?b=2"
    assert normalize_url("https://example.com/a?srsltid=1") == "https://example.com/a"
    assert domain_of("http://www.Example.com/a") == "example.com"
    assert jaccard({"a", "b"}, {"b", "c"}) == 1 / 3
    assert jaccard(set(), set()) == 1.0
    protocol = json.loads(PROTOCOL_PATH.read_text(encoding="utf-8"))
    assert len(query_units(protocol)) == 138
    print("serper experiment self-test: ok")


def main() -> int:
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)
    collect_parser = subparsers.add_parser("collect")
    collect_parser.add_argument("--wave", type=int, choices=(1, 2), required=True)
    collect_parser.add_argument("--env-file", type=Path, required=True)
    collect_parser.add_argument("--workers", type=int, default=10)
    subparsers.add_parser("analyze")
    subparsers.add_parser("self-test")
    args = parser.parse_args()
    if args.command == "collect":
        collect(args.wave, args.env_file, args.workers)
    elif args.command == "analyze":
        analyze()
    else:
        self_test()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
