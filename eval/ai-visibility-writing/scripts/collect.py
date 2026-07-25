#!/usr/bin/env python3
"""Idempotent, budget-gated DataForSEO collector with mock support."""

from __future__ import annotations

import argparse
import base64
import hashlib
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

from pilotlib import (
    BudgetError, PilotError, RetryableResponse, append_ledger, canonical_json,
    check_batch_budget, parse_chatgpt_response, parse_google_response, sha256_text,
    terminal_events, TOTAL_LIMIT, utc_now,
)


ROOT = Path(__file__).resolve().parents[1]
API = "https://api.dataforseo.com/v3"


class GooglePostError(PilotError):
    def __init__(self, message: str, raw: dict):
        super().__init__(message)
        self.raw = raw


def env_file_values(path: Path | None) -> dict[str, str]:
    if path is None:
        return {}
    values = {}
    for line_number, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            raise PilotError(f"invalid env file line {line_number}")
        key, value = line.split("=", 1)
        key = key.strip().removeprefix("export ").strip()
        if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            raise PilotError(f"invalid env file key on line {line_number}")
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
            value = value[1:-1]
        values[key] = value
    return values


def credentials(env_file: Path | None = None) -> tuple[str, str]:
    source = env_file_values(env_file)
    source.update(os.environ)
    login = source.get("DATAFORSEO_LOGIN") or source.get("DATAFORSEO_USERNAME") or source.get("DATAFORSEO_API_LOGIN")
    password = source.get("DATAFORSEO_PASSWORD") or source.get("DATAFORSEO_API_PASSWORD")
    if login and password:
        return login, password
    api_key = source.get("DATAFORSEO_API_KEY")
    if api_key:
        token = api_key.strip()
        if token.lower().startswith("basic "):
            token = token.split(None, 1)[1].strip()
        try:
            decoded = base64.b64decode(token, validate=True).decode("utf-8")
        except (ValueError, UnicodeDecodeError) as exc:
            raise PilotError("DataForSEO API key is not a valid Basic credential") from exc
        if ":" not in decoded:
            raise PilotError("DataForSEO API key is not a valid Basic credential")
        return tuple(decoded.split(":", 1))
    raise PilotError("DataForSEO credentials are absent")


def http_json(method: str, url: str, payload: object | None, auth: tuple[str, str]) -> dict:
    data = canonical_json(payload).encode("utf-8") if payload is not None else None
    token = base64.b64encode(f"{auth[0]}:{auth[1]}".encode()).decode()
    request = urllib.request.Request(url, data=data, method=method)
    request.add_header("Authorization", f"Basic {token}")
    request.add_header("Content-Type", "application/json")
    value = None
    for attempt in range(4):
        try:
            with urllib.request.urlopen(request, timeout=150) as response:
                value = json.load(response)
            break
        except urllib.error.HTTPError as exc:
            if exc.code not in {408, 429, 500, 502, 503, 504} or attempt == 3:
                raise PilotError(f"DataForSEO HTTP {exc.code}") from exc
        except urllib.error.URLError as exc:
            if attempt == 3:
                raise PilotError("DataForSEO network failure") from exc
        time.sleep(2 ** attempt)
    if value is None:
        raise PilotError("DataForSEO request produced no response")
    if not isinstance(value, dict):
        raise PilotError("DataForSEO response was not an object")
    return value


def provider_daily_money_limit(auth: tuple[str, str]) -> float:
    raw = http_json("GET", f"{API}/appendix/user_data", None, auth)
    tasks = raw.get("tasks")
    task = tasks[0] if isinstance(tasks, list) and tasks and isinstance(tasks[0], dict) else {}
    results = task.get("result")
    result = results[0] if isinstance(results, list) and results and isinstance(results[0], dict) else {}
    if raw.get("status_code") != 20000 or task.get("status_code") != 20000:
        raise PilotError("DataForSEO provider-cap verification failed")
    if isinstance(raw.get("cost"), (int, float)) and raw["cost"] > 0:
        raise PilotError("DataForSEO provider-cap verification was not free")
    limit = (((result.get("money") or {}).get("limits") or {}).get("day") or {}).get("total")
    if isinstance(limit, bool) or not isinstance(limit, (int, float)) or not 0 < float(limit):
        raise BudgetError("provider daily money limit is absent or disabled")
    return float(limit)


def verify_provider_cap(auth: tuple[str, str]) -> float:
    limit = provider_daily_money_limit(auth)
    if limit > TOTAL_LIMIT:
        raise BudgetError("provider daily money limit is absent, disabled, or exceeds $10")
    return limit


def provider_safety(auth: tuple[str, str], *, capped: bool, free_credit_only: bool) -> tuple[str, float]:
    if capped == free_credit_only:
        raise PilotError("choose exactly one provider safety mode")
    if capped:
        return "verified_daily_cap_at_most_10", verify_provider_cap(auth)
    return "user_confirmed_free_credit_only_no_payment_method", provider_daily_money_limit(auth)


def choose_chatgpt_model(auth: tuple[str, str], requested: str | None) -> str:
    if requested:
        return requested
    raw = http_json("GET", f"{API}/ai_optimization/chat_gpt/llm_responses/models", None, auth)
    candidates = []
    for node in [raw] + list(_walk(raw)):
        if isinstance(node, dict):
            name = node.get("model_name") or node.get("name")
            supports = node.get("web_search_supported", node.get("web_search", node.get("supports_web_search")))
            reasoning = node.get("reasoning")
            if isinstance(name, str) and supports is True and reasoning is False:
                candidates.append(name)
    if not candidates:
        raise PilotError("no supported non-reasoning ChatGPT web-search model found")
    def model_key(name: str) -> tuple:
        numbers = tuple(int(value) for value in re.findall(r"\d+", name))
        dated = bool(re.search(r"20\d{2}[-_]\d{2}[-_]\d{2}", name))
        compact = not any(value in name.lower() for value in ("mini", "nano"))
        return numbers, dated, compact, name
    return max(set(candidates), key=model_key)


def _walk(value):
    if isinstance(value, dict):
        for child in value.values():
            yield child
            yield from _walk(child)
    elif isinstance(value, list):
        for child in value:
            yield child
            yield from _walk(child)


def ensure_google_task(request: dict, auth: tuple[str, str], pending_path: Path) -> str:
    payload = [{
        "keyword": request["query"], "location_code": 2840,
        "language_code": "en", "device": "desktop", "os": "windows",
        "depth": 10, "load_async_ai_overview": True, "tag": request["request_tag"],
    }]
    if pending_path.exists():
        pending = json.loads(pending_path.read_text(encoding="utf-8"))
        if pending.get("request_tag") != request["request_tag"] or not pending.get("task_id"):
            raise PilotError("Google pending-task state mismatch")
        task_id = pending["task_id"]
    else:
        posted = http_json("POST", f"{API}/serp/google/organic/task_post", payload, auth)
        tasks = posted.get("tasks") or []
        if len(tasks) != 1 or not tasks[0].get("id"):
            raise GooglePostError("Google task POST returned no task id", posted)
        task_id = tasks[0]["id"]
        atomic_raw(pending_path, {
            "request_tag": request["request_tag"], "task_id": task_id,
            "posted_at": utc_now(), "post_response_sha256": sha256_text(canonical_json(posted)),
        })
    return task_id


def google_live(request: dict, auth: tuple[str, str], pending_path: Path) -> dict:
    task_id = ensure_google_task(request, auth, pending_path)
    for attempt in range(8):
        result = http_json("GET", f"{API}/serp/google/organic/task_get/advanced/{task_id}", None, auth)
        try:
            parse_google_response(result, request)
            return result
        except RetryableResponse:
            if attempt == 7:
                raise
            time.sleep(min(2 ** attempt, 30))
        except PilotError as exc:
            if "has no result" not in str(exc):
                return result
            if attempt == 7:
                raise
            time.sleep(min(2 ** attempt, 30))
    raise RetryableResponse("Google polling exhausted")


def chatgpt_live(request: dict, auth: tuple[str, str], model: str) -> dict:
    payload = [{
        "user_prompt": request["query"], "model_name": model,
        "web_search": True, "force_web_search": True,
        "web_search_country_iso_code": "US", "max_output_tokens": 700,
        "temperature": 0.2, "tag": request["request_tag"],
    }]
    return http_json("POST", f"{API}/ai_optimization/chat_gpt/llm_responses/live", payload, auth)


def atomic_raw(path: Path, raw: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    descriptor = os.open(path, flags, 0o600)
    with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
        json.dump(raw, handle, ensure_ascii=False)
        handle.write("\n")
        handle.flush()
        os.fsync(handle.fileno())


def append_record(path: Path, record: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(canonical_json(record) + "\n")
        handle.flush()
        os.fsync(handle.fileno())


def record_terminal_error(
    *, ledger: Path, raw_path: Path, raw: dict | None, request: dict,
    phase: str, platform: str, provider_mode: str, provider_limit: float | None,
    exc: Exception,
) -> None:
    if isinstance(raw, dict) and not raw_path.exists():
        atomic_raw(raw_path, raw)
    cost = response_cost(raw)
    append_ledger(ledger, {
        "request_tag": request["request_tag"], "phase": phase, "platform": platform,
        "terminal": True, "status": "error" if cost is not None else "unknown_cost", "cost_usd": cost,
        "provider_safety_mode": provider_mode,
        "provider_daily_money_limit_usd": provider_limit,
        "retry_of": request.get("retry_of"),
        "error_class": type(exc).__name__,
        "request_sha256": sha256_text(canonical_json({"query": request["query"], "platform": platform})),
        "response_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest() if raw_path.exists() else None,
    })


def response_cost(raw: dict | None) -> float | None:
    if not isinstance(raw, dict):
        return None
    tasks = raw.get("tasks")
    candidates = [raw.get("cost")]
    if isinstance(tasks, list) and tasks and isinstance(tasks[0], dict):
        candidates.insert(0, tasks[0].get("cost"))
    for value in candidates:
        if isinstance(value, (int, float)) and value >= 0:
            return round(float(value), 10)
    return None


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--ledger", type=Path, default=ROOT / "runs" / "cost-ledger.jsonl")
    parser.add_argument("--run-dir", type=Path, default=ROOT / "runs" / "current")
    parser.add_argument("--mock-dir", type=Path)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--env-file", type=Path)
    safety = parser.add_mutually_exclusive_group()
    safety.add_argument(
        "--confirm-provider-cap", action="store_true",
        help="confirm auto-recharge is disabled; the provider daily money limit is verified by API",
    )
    safety.add_argument(
        "--confirm-free-credit-only-account", action="store_true",
        help="confirm the account has no payment method or auto-recharge and only free credits are at risk",
    )
    parser.add_argument("--chatgpt-model")
    parser.add_argument("--limit", type=int)
    parser.add_argument("--retry-errors-from-reserve", action="store_true")
    parser.add_argument(
        "--queue-google", action="store_true",
        help="journal outstanding Google task IDs without polling; resume with the same command minus this flag",
    )
    args = parser.parse_args()
    if args.live == bool(args.mock_dir):
        raise SystemExit("choose exactly one of --live or --mock-dir")
    manifest = json.loads(args.manifest.read_text(encoding="utf-8"))
    phase = manifest["phase"]
    requests = []
    for unit in manifest["units"]:
        for descriptor in unit["requests"]:
            requests.append({**unit, **descriptor})
    if args.limit is not None:
        requests = requests[: args.limit]
    terminal = terminal_events(args.ledger)
    if args.retry_errors_from_reserve:
        retryable = []
        for request in requests:
            event = terminal.get(request["request_tag"])
            if event and event.get("status") != "success" and isinstance(event.get("cost_usd"), (int, float)):
                retryable.append({**request, "request_tag": f"{request['request_tag']}-reserve-retry-1", "phase": "reserve", "retry_of": request["request_tag"]})
        requests = [request for request in retryable if request["request_tag"] not in terminal]
        phase = "reserve"
    else:
        requests = [request for request in requests if request["request_tag"] not in terminal]
    if args.queue_google:
        if not args.live:
            raise SystemExit("--queue-google requires --live")
        requests = [
            request for request in requests
            if request["platform"] == "google"
            and not (args.run_dir / "pending" / f"{request['request_tag']}.json").exists()
            and not (args.run_dir / "raw" / f"{request['request_tag']}.json").exists()
        ]
    counts = {platform: sum(r["platform"] == platform for r in requests) for platform in ("google", "chatgpt")}
    projections = check_batch_budget(args.ledger, phase, counts)
    print(canonical_json({"phase": phase, "requests": len(requests), "projections": {k: v.__dict__ for k, v in projections.items()}}), file=sys.stderr)
    if args.live and not (args.confirm_provider_cap or args.confirm_free_credit_only_account):
        raise SystemExit("live collection requires one explicit provider safety mode")
    try:
        auth = credentials(args.env_file) if args.live else ("", "")
    except PilotError as exc:
        raise SystemExit(f"live collection stopped: {exc}") from None
    if args.live:
        try:
            provider_mode, provider_limit = provider_safety(
                auth, capped=args.confirm_provider_cap,
                free_credit_only=args.confirm_free_credit_only_account,
            )
        except PilotError as exc:
            raise SystemExit(f"live collection stopped: {exc}") from None
        print(canonical_json({"provider_daily_money_limit_usd": provider_limit, "provider_safety_mode": provider_mode}), file=sys.stderr)
    else:
        provider_mode, provider_limit = "mock", None
    model = choose_chatgpt_model(auth, args.chatgpt_model) if args.live and counts["chatgpt"] else args.chatgpt_model
    if args.queue_google:
        queued = 0
        for request in requests:
            pending_path = args.run_dir / "pending" / f"{request['request_tag']}.json"
            try:
                ensure_google_task(request, auth, pending_path)
                queued += 1
            except Exception as exc:
                raw = exc.raw if isinstance(exc, GooglePostError) else None
                record_terminal_error(
                    ledger=args.ledger,
                    raw_path=args.run_dir / "raw" / f"{request['request_tag']}.json",
                    raw=raw, request=request, phase=phase, platform="google",
                    provider_mode=provider_mode, provider_limit=provider_limit, exc=exc,
                )
                raise
        print(canonical_json({"queued_google_tasks": queued}), file=sys.stderr)
        return 0
    for request in requests:
        tag = request["request_tag"]
        platform = request["platform"]
        raw = None
        raw_path = args.run_dir / "raw" / f"{tag}.json"
        pending_path = args.run_dir / "pending" / f"{tag}.json"
        try:
            if raw_path.exists():
                raw = json.loads(raw_path.read_text(encoding="utf-8"))
            elif args.mock_dir:
                fixture = args.mock_dir / f"{platform}.json"
                raw = json.loads(fixture.read_text(encoding="utf-8"))
            elif platform == "google":
                raw = google_live(request, auth, pending_path)
            else:
                raw = chatgpt_live(request, auth, model)
            parser_fn = parse_google_response if platform == "google" else parse_chatgpt_response
            record = parser_fn(raw, request)
            if not raw_path.exists():
                atomic_raw(raw_path, raw)
            observations_path = args.run_dir / "observations.jsonl"
            existing_tags = {
                json.loads(line).get("request_tag")
                for line in observations_path.read_text(encoding="utf-8").splitlines()
                if line.strip()
            } if observations_path.exists() else set()
            if tag not in existing_tags:
                append_record(observations_path, record)
            append_ledger(args.ledger, {
                "request_tag": tag, "phase": phase, "platform": platform,
                "terminal": True, "status": "success", "cost_usd": record["cost_usd"],
                "money_spent_usd": record.get("money_spent_usd"),
                "provider_safety_mode": provider_mode,
                "provider_daily_money_limit_usd": provider_limit,
                "retry_of": request.get("retry_of"),
                "request_sha256": sha256_text(canonical_json({"query": request["query"], "platform": platform})),
                "response_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest(),
            })
        except RetryableResponse as exc:
            if platform == "google" and pending_path.exists() and raw is None:
                raise SystemExit(f"live collection paused with saved task id: {exc}") from None
            record_terminal_error(
                ledger=args.ledger, raw_path=raw_path, raw=raw, request=request,
                phase=phase, platform=platform, provider_mode=provider_mode,
                provider_limit=provider_limit, exc=exc,
            )
            raise
        except Exception as exc:
            if raw is None and isinstance(exc, GooglePostError):
                raw = exc.raw
            record_terminal_error(
                ledger=args.ledger, raw_path=raw_path, raw=raw, request=request,
                phase=phase, platform=platform, provider_mode=provider_mode,
                provider_limit=provider_limit, exc=exc,
            )
            raise
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
