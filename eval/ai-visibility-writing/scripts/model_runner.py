#!/usr/bin/env python3
"""Shared clean-context Codex/Claude execution and invocation accounting."""

from __future__ import annotations

import json
import os
import subprocess
import tempfile
from pathlib import Path
from typing import Any

from pilotlib import PilotError, canonical_json, read_jsonl, sha256_text, utc_now


EVAL = Path(__file__).resolve().parents[1]
REPO = EVAL.parents[1]
INVOCATIONS = EVAL / "runs" / "model-invocations.jsonl"
CAP = 450


def _validate(value: Any, schema: dict[str, Any], path: str = "$") -> None:
    expected = schema.get("type")
    if expected == "object" and not isinstance(value, dict):
        raise PilotError(f"{path} is not an object")
    if expected == "array" and not isinstance(value, list):
        raise PilotError(f"{path} is not an array")
    if expected == "string" and not isinstance(value, str):
        raise PilotError(f"{path} is not a string")
    if expected == "boolean" and not isinstance(value, bool):
        raise PilotError(f"{path} is not a boolean")
    if expected == "integer" and (not isinstance(value, int) or isinstance(value, bool)):
        raise PilotError(f"{path} is not an integer")
    if "enum" in schema and value not in schema["enum"]:
        raise PilotError(f"{path} is outside enum")
    if isinstance(value, str) and len(value) < int(schema.get("minLength", 0)):
        raise PilotError(f"{path} is too short")
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        if "minimum" in schema and value < schema["minimum"]:
            raise PilotError(f"{path} is below minimum")
        if "maximum" in schema and value > schema["maximum"]:
            raise PilotError(f"{path} is above maximum")
    if isinstance(value, list):
        if schema.get("uniqueItems") and len({canonical_json(item) for item in value}) != len(value):
            raise PilotError(f"{path} contains duplicate items")
        if isinstance(schema.get("items"), dict):
            for index, item in enumerate(value):
                _validate(item, schema["items"], f"{path}[{index}]")
        return
    if not isinstance(value, dict):
        return
    required = schema.get("required", [])
    if any(key not in value for key in required):
        raise PilotError(f"{path} is missing required fields")
    if schema.get("additionalProperties") is False and set(value) - set(schema.get("properties", {})):
        raise PilotError(f"{path} has extra fields")
    for key, definition in schema.get("properties", {}).items():
        if key not in value:
            continue
        _validate(value[key], definition, f"{path}.{key}")


def _codex_schema(value: Any) -> Any:
    """Drop schema keywords the Codex response-format API rejects.

    The full local schema remains authoritative and is applied by `_validate`
    after generation, so this compatibility projection does not weaken the
    harness contract.
    """
    if isinstance(value, dict):
        return {key: _codex_schema(child) for key, child in value.items() if key != "uniqueItems"}
    if isinstance(value, list):
        return [_codex_schema(child) for child in value]
    return value


def _extract_claude(stdout: str) -> tuple[dict[str, Any], dict[str, Any]]:
    envelope = json.loads(stdout)
    candidate = envelope.get("structured_output")
    if candidate is None:
        candidate = envelope.get("result")
        if isinstance(candidate, str):
            candidate = json.loads(candidate)
    if not isinstance(candidate, dict):
        candidate = envelope if isinstance(envelope, dict) else None
    if not isinstance(candidate, dict):
        raise PilotError("Claude returned no structured output")
    return candidate, envelope


def _find_numbers(value: Any, names: set[str]) -> int:
    total = 0
    if isinstance(value, dict):
        for key, child in value.items():
            if key in names and isinstance(child, int):
                total += child
            else:
                total += _find_numbers(child, names)
    elif isinstance(value, list):
        total += sum(_find_numbers(child, names) for child in value)
    return total


def _find_model(value: Any, requested: str) -> str | None:
    if isinstance(value, dict) and isinstance(value.get("modelUsage"), dict):
        models = list(value["modelUsage"])
        if requested in models:
            return requested
        if len(models) == 1:
            return models[0]
    if isinstance(value, dict):
        for key in ("model_name", "model"):
            if isinstance(value.get(key), str):
                return value[key]
        for child in value.values():
            found = _find_model(child, requested)
            if found:
                return found
    elif isinstance(value, list):
        for child in value:
            found = _find_model(child, requested)
            if found:
                return found
    return None


def _usage(metadata: dict[str, Any]) -> dict[str, Any]:
    usage = metadata.get("usage")
    if isinstance(usage, dict):
        return {
            "input_tokens": usage.get("input_tokens", 0),
            "cache_creation_input_tokens": usage.get("cache_creation_input_tokens", 0),
            "cache_read_input_tokens": usage.get("cache_read_input_tokens", 0),
            "output_tokens": usage.get("output_tokens", 0),
            "model_usage": metadata.get("modelUsage"),
        }
    return {
        "input_tokens": _find_numbers(metadata, {"input_tokens", "inputTokens"}),
        "cache_creation_input_tokens": _find_numbers(metadata, {"cache_creation_input_tokens", "cacheCreationInputTokens"}),
        "cache_read_input_tokens": _find_numbers(metadata, {"cache_read_input_tokens", "cacheReadInputTokens"}),
        "output_tokens": _find_numbers(metadata, {"output_tokens", "outputTokens"}),
        "model_usage": metadata.get("modelUsage"),
    }


def append_invocation(event: dict[str, Any]) -> None:
    INVOCATIONS.parent.mkdir(parents=True, exist_ok=True)
    with INVOCATIONS.open("a", encoding="utf-8") as handle:
        handle.write(canonical_json(event) + "\n")
        handle.flush()
        os.fsync(handle.fileno())


def run_structured(*, executor: str, prompt: str, schema_path: Path, output: Path,
                   kind: str, case_id: str, condition: str, model: str,
                   dry_run: bool = False) -> dict[str, Any]:
    if executor not in {"codex", "claude"}:
        raise PilotError("executor must be codex or claude")
    output = output.resolve()
    schema_path = schema_path.resolve()
    if output.exists() and output.stat().st_size:
        value = json.loads(output.read_text(encoding="utf-8"))
        _validate(value, json.loads(schema_path.read_text(encoding="utf-8")))
        return value
    if len(read_jsonl(INVOCATIONS)) >= CAP:
        raise PilotError("model invocation cap reached")
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    if dry_run:
        return {"dry_run": True, "prompt_sha256": sha256_text(prompt), "executor": executor, "model": model}
    started = utc_now()
    output.parent.mkdir(parents=True, exist_ok=True)
    part = output.with_suffix(output.suffix + ".part")
    log = output.with_suffix(output.suffix + f".{executor}.log")
    metadata: dict[str, Any] = {}
    status = "error"
    command: list[str] = []
    try:
        with tempfile.TemporaryDirectory(prefix="ai-visibility-run-") as directory:
            if executor == "codex":
                codex_schema_path = Path(directory) / "codex-output-schema.json"
                codex_schema_path.write_text(canonical_json(_codex_schema(schema)), encoding="utf-8")
                command = [
                    "codex", "exec", "--skip-git-repo-check", "--sandbox", "read-only",
                    "--ephemeral", "--ignore-user-config", "--model", model, "--json",
                    "--output-schema", str(codex_schema_path), "--output-last-message", str(part), "-",
                ]
                result = subprocess.run(command, input=prompt, text=True, capture_output=True, cwd=directory, timeout=300)
                log.write_text(result.stdout + "\n--- STDERR ---\n" + result.stderr, encoding="utf-8")
                if result.returncode != 0:
                    raise PilotError(f"Codex exited {result.returncode}")
                value = json.loads(part.read_text(encoding="utf-8"))
                events = [json.loads(line) for line in result.stdout.splitlines() if line.strip().startswith("{")]
                metadata = {"events": events}
            elif executor == "claude":
                command = [
                    "claude", "-p", "--safe-mode", "--tools", "", "--no-session-persistence",
                    "--model", model, "--output-format", "json", "--json-schema", canonical_json(schema),
                ]
                result = subprocess.run(command, input=prompt, text=True, capture_output=True, cwd=directory, timeout=300)
                log.write_text(result.stdout + "\n--- STDERR ---\n" + result.stderr, encoding="utf-8")
                if result.returncode != 0:
                    raise PilotError(f"Claude exited {result.returncode}")
                value, metadata = _extract_claude(result.stdout)
                part.write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")
        _validate(value, schema)
        os.replace(part, output)
        status = "success"
        return value
    finally:
        if part.exists():
            part.unlink()
        usage = _usage(metadata)
        append_invocation({
            "started_at": started, "ended_at": utc_now(), "executor": executor,
            "kind": kind, "case_id": case_id, "condition": condition,
            "requested_model": model, "resolved_model": _find_model(metadata, model),
            "command": command, "prompt_sha256": sha256_text(prompt), "status": status,
            **usage,
            "output": str(output.relative_to(REPO)) if output.is_relative_to(REPO) else str(output),
        })
