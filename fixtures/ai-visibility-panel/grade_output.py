#!/usr/bin/env python3
"""Validate an AI visibility panel run and optionally grade semantic QA.

The implementation intentionally uses only the Python standard library.  It
accepts JSON plus the conservative YAML subset used by Newsjack panel files.
It does not attempt to replace semantic or human review.
"""

from __future__ import annotations

import argparse
import ast
import hashlib
import json
import re
import sys
import unicodedata
from collections import Counter, defaultdict
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Iterable


REQUIRED_FILES = (
    "panel_report.md",
    "tracking_plan.md",
    "measurement_charter.json",
    "source_manifest.json",
    "icp_hypotheses.json",
    "buyer_jobs.json",
    "contamination_register.yaml",
    "blind_design_brief.json",
    "prompt_architecture.json",
    "prompt_universe.json",
    "prompt_qa.json",
    "panel.yaml",
    "run_manifest_template.json",
    "panel_change_ledger.json",
)

MACHINE_FILES = tuple(name for name in REQUIRED_FILES if not name.endswith(".md"))

HASH_RE = re.compile(r"^sha256:[0-9a-f]{64}$")
RFC3339_RE = re.compile(
    r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$"
)
SLUG_RE = re.compile(r"^[a-z0-9][a-z0-9._:-]*$")
SEMVER_RE = re.compile(r"^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$")

ESTIMANDS = {
    "unaided_brand_presence",
    "aided_brand_knowledge",
    "competitive_mention_share",
    "citation_presence",
    "answer_framing",
    "campaign_response",
}
SOURCE_CLASSES = {
    "company_asserted",
    "buyer_behavior",
    "independent",
    "search_proxy",
    "llm_hypothesis",
}
SOURCE_TYPES = {
    "product",
    "pricing",
    "support",
    "technical",
    "review",
    "forum",
    "interview",
    "query",
    "procurement",
    "news",
    "other",
}
EVIDENCE_GRADES = {"A", "B", "C", "D"}
CONFIDENCE = {"high", "medium", "low"}
INFORMATION_ACTS = {
    "explain",
    "diagnose",
    "plan",
    "generate",
    "compare",
    "recommend",
    "verify",
    "navigate",
    "buy",
    "implement",
    "troubleshoot",
}
JOURNEYS = {
    "problem_identification",
    "exploration",
    "requirements_building",
    "supplier_selection",
    "adoption",
    "post_purchase",
}
FUNNELS = {"TOFU", "MOFU", "BOFU", None}
BANDS = {
    "B0_direct_brand_product",
    "B1_comparison_purchase",
    "B2_category",
    "B3_problem_need",
    "B4_job_goal",
    "B5_broad_discovery_story",
}
AIDED_STATUSES = {
    "target_aided",
    "competitor_aided",
    "category_aided",
    "unaided",
}
LANES = {
    "closed_model",
    "retrieval",
    "consumer_surface",
    "campaign_experiment",
}
PARTITIONS = {"core", "rotating", "sentinel", "control", "aided"}
TRANSFORMATIONS = {
    "verbatim",
    "lightly_normalized",
    "search_query_expanded",
    "human_written",
    "llm_expanded",
    "translated",
    "locale_transcreated",
}
QA_STATUSES = {"pass", "revise", "quarantine", "reject"}
RULE_STATUSES = {"pass", "fail", "review"}
DUPLICATE_ACTIONS = {
    "retain_variant",
    "merge_exact",
    "merge_semantic",
    "split",
}
ROUTES = {
    "realistic-prompt-generation",
    "prompt-proximity-architecture",
    "buyer-job-intent-analysis",
    "human_gate_3",
    None,
}


class YamlSubsetError(ValueError):
    """Raised when a file exceeds the deliberately small YAML subset."""


def _strip_yaml_comment(value: str) -> str:
    quote: str | None = None
    escaped = False
    for index, char in enumerate(value):
        if escaped:
            escaped = False
            continue
        if char == "\\" and quote == '"':
            escaped = True
            continue
        if char in {"'", '"'}:
            if quote is None:
                quote = char
            elif quote == char:
                quote = None
            continue
        if char == "#" and quote is None and (
            index == 0 or value[index - 1].isspace()
        ):
            return value[:index].rstrip()
    return value.rstrip()


def _split_top_level(value: str, delimiter: str = ",") -> list[str]:
    parts: list[str] = []
    start = 0
    depth = 0
    quote: str | None = None
    escaped = False
    for index, char in enumerate(value):
        if escaped:
            escaped = False
            continue
        if char == "\\" and quote == '"':
            escaped = True
            continue
        if char in {"'", '"'}:
            if quote is None:
                quote = char
            elif quote == char:
                quote = None
            continue
        if quote is not None:
            continue
        if char in "[{(":
            depth += 1
        elif char in "]})":
            depth -= 1
        elif char == delimiter and depth == 0:
            parts.append(value[start:index].strip())
            start = index + 1
    parts.append(value[start:].strip())
    return parts


def _split_yaml_key(value: str) -> tuple[str, str] | None:
    depth = 0
    quote: str | None = None
    escaped = False
    for index, char in enumerate(value):
        if escaped:
            escaped = False
            continue
        if char == "\\" and quote == '"':
            escaped = True
            continue
        if char in {"'", '"'}:
            if quote is None:
                quote = char
            elif quote == char:
                quote = None
            continue
        if quote is not None:
            continue
        if char in "[{(":
            depth += 1
        elif char in "]})":
            depth -= 1
        elif char == ":" and depth == 0:
            return value[:index].strip(), value[index + 1 :].strip()
    return None


def _parse_yaml_scalar(value: str) -> Any:
    value = _strip_yaml_comment(value.strip())
    if value == "":
        return None
    if value in {"null", "Null", "NULL", "~"}:
        return None
    if value.lower() == "true":
        return True
    if value.lower() == "false":
        return False
    if value.startswith('"') and value.endswith('"'):
        return json.loads(value)
    if value.startswith("'") and value.endswith("'"):
        return ast.literal_eval(value)
    if value.startswith("[") and value.endswith("]"):
        inner = value[1:-1].strip()
        return [] if not inner else [
            _parse_yaml_scalar(part) for part in _split_top_level(inner)
        ]
    if value.startswith("{") and value.endswith("}"):
        inner = value[1:-1].strip()
        result: dict[str, Any] = {}
        if not inner:
            return result
        for part in _split_top_level(inner):
            pair = _split_yaml_key(part)
            if pair is None:
                raise YamlSubsetError(f"invalid inline mapping item: {part!r}")
            key, item = pair
            result[str(_parse_yaml_scalar(key))] = _parse_yaml_scalar(item)
        return result
    if re.fullmatch(r"-?\d+", value):
        return int(value)
    if re.fullmatch(r"-?(?:\d+\.\d*|\d*\.\d+)", value):
        return float(value)
    return value


def parse_yaml_subset(text: str) -> Any:
    """Parse mappings/lists/scalars from the YAML subset in panel artifacts."""

    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass

    tokens: list[tuple[int, str, int]] = []
    for line_no, raw_line in enumerate(text.splitlines(), start=1):
        if not raw_line.strip() or raw_line.lstrip().startswith(("#", "---", "...")):
            continue
        if "\t" in raw_line[: len(raw_line) - len(raw_line.lstrip())]:
            raise YamlSubsetError(f"line {line_no}: tabs are not supported")
        indent = len(raw_line) - len(raw_line.lstrip(" "))
        tokens.append((indent, raw_line.strip(), line_no))

    if not tokens:
        return {}

    def parse_block(index: int, indent: int) -> tuple[Any, int]:
        if index >= len(tokens):
            return {}, index
        is_list = tokens[index][1] == "-" or tokens[index][1].startswith("- ")
        container: Any = [] if is_list else {}

        while index < len(tokens):
            current_indent, content, line_no = tokens[index]
            if current_indent < indent:
                break
            if current_indent > indent:
                raise YamlSubsetError(
                    f"line {line_no}: unexpected indentation {current_indent}"
                )

            if is_list:
                if not (content == "-" or content.startswith("- ")):
                    break
                rest = content[1:].strip()
                index += 1
                if not rest:
                    if index >= len(tokens) or tokens[index][0] <= indent:
                        container.append(None)
                    else:
                        child, index = parse_block(index, tokens[index][0])
                        container.append(child)
                    continue

                pair = _split_yaml_key(rest)
                if pair is None:
                    container.append(_parse_yaml_scalar(rest))
                    if index < len(tokens) and tokens[index][0] > indent:
                        raise YamlSubsetError(
                            f"line {tokens[index][2]}: scalar list item has children"
                        )
                    continue

                key, raw_value = pair
                item: dict[str, Any] = {}
                if raw_value:
                    item[key] = _parse_yaml_scalar(raw_value)
                elif index < len(tokens) and tokens[index][0] > indent:
                    child, index = parse_block(index, tokens[index][0])
                    item[key] = child
                else:
                    item[key] = None

                if index < len(tokens) and tokens[index][0] > indent:
                    child_indent = tokens[index][0]
                    child, index = parse_block(index, child_indent)
                    if not isinstance(child, dict):
                        raise YamlSubsetError(
                            f"line {tokens[index - 1][2]}: list mapping continuation "
                            "must be a mapping"
                        )
                    item.update(child)
                container.append(item)
                continue

            if content == "-" or content.startswith("- "):
                break
            pair = _split_yaml_key(content)
            if pair is None:
                raise YamlSubsetError(f"line {line_no}: expected key: value")
            key, raw_value = pair
            if not key:
                raise YamlSubsetError(f"line {line_no}: empty mapping key")
            index += 1
            if raw_value:
                container[key] = _parse_yaml_scalar(raw_value)
            elif index < len(tokens) and tokens[index][0] > indent:
                child, index = parse_block(index, tokens[index][0])
                container[key] = child
            else:
                container[key] = None
        return container, index

    parsed, final_index = parse_block(0, tokens[0][0])
    if final_index != len(tokens):
        _, _, line_no = tokens[final_index]
        raise YamlSubsetError(f"line {line_no}: could not parse remaining content")
    return parsed


def normalize_text(value: str) -> str:
    value = unicodedata.normalize("NFKC", value).casefold()
    value = "".join(
        char if char.isalnum() else " "
        for char in value
    )
    return " ".join(value.split())


def iter_strings(value: Any, *, skip_keys: set[str] | None = None) -> Iterable[str]:
    skip_keys = skip_keys or set()
    if isinstance(value, str):
        yield value
    elif isinstance(value, list):
        for item in value:
            yield from iter_strings(item, skip_keys=skip_keys)
    elif isinstance(value, dict):
        for key, item in value.items():
            if key in skip_keys or key.endswith(("_id", "_ids", "_hash")):
                continue
            yield from iter_strings(item, skip_keys=skip_keys)


def list_value(value: Any) -> list[Any]:
    if value is None:
        return []
    return value if isinstance(value, list) else [value]


def nested_values(value: Any, key_names: set[str]) -> list[Any]:
    found: list[Any] = []
    if isinstance(value, dict):
        for key, item in value.items():
            if key in key_names:
                found.extend(list_value(item))
            found.extend(nested_values(item, key_names))
    elif isinstance(value, list):
        for item in value:
            found.extend(nested_values(item, key_names))
    return found


@dataclass(frozen=True)
class Finding:
    level: str
    code: str
    path: str
    message: str


class Validator:
    def __init__(self, root: Path) -> None:
        self.root = root
        self.data: dict[str, Any] = {}
        self.text: dict[str, str] = {}
        self.findings: list[Finding] = []
        self.source_ids: set[str] = set()
        self.icp_ids: set[str] = set()
        self.job_ids: set[str] = set()
        self.specs: dict[str, dict[str, Any]] = {}
        self.cells: dict[str, dict[str, Any]] = {}
        self.candidates: dict[str, dict[str, Any]] = {}
        self.candidate_cells: dict[str, dict[str, Any]] = {}
        self.decisions: dict[str, dict[str, Any]] = {}
        self.accepted_ids: set[str] = set()

    def provisional(self) -> bool:
        panel = self.data.get("panel.yaml")
        return (
            isinstance(panel, dict)
            and panel.get("status") == "provisional_directional"
        )

    def add(self, level: str, code: str, path: str, message: str) -> None:
        self.findings.append(Finding(level, code, path, message))

    def error(self, code: str, path: str, message: str) -> None:
        self.add("error", code, path, message)

    def warning(self, code: str, path: str, message: str) -> None:
        self.add("warning", code, path, message)

    def load(self) -> None:
        if not self.root.is_dir():
            self.error("output-dir", str(self.root), "output directory does not exist")
            return
        for name in REQUIRED_FILES:
            path = self.root / name
            if not path.is_file():
                self.error("required-file", name, "required artifact is missing")
                continue
            try:
                raw = path.read_text(encoding="utf-8")
            except (OSError, UnicodeError) as exc:
                self.error("artifact-read", name, str(exc))
                continue
            self.text[name] = raw
            if name.endswith(".md"):
                if not raw.strip():
                    self.error("empty-human-output", name, "human output is empty")
                continue
            try:
                if name.endswith(".json"):
                    parsed = json.loads(raw)
                else:
                    parsed = parse_yaml_subset(raw)
            except (json.JSONDecodeError, YamlSubsetError, ValueError) as exc:
                self.error("artifact-parse", name, str(exc))
                continue
            if not isinstance(parsed, dict):
                self.error("artifact-shape", name, "top level must be a mapping")
                continue
            self.data[name] = parsed

    def check_envelopes(self) -> None:
        seen_artifact_ids: dict[str, str] = {}
        manifest = self.data.get("source_manifest.json")
        expected_manifest_hash = (
            manifest.get("source_manifest_hash")
            if isinstance(manifest, dict)
            else None
        )
        for name in MACHINE_FILES:
            artifact = self.data.get(name)
            if not isinstance(artifact, dict):
                continue
            prefix = name
            required = (
                "schema_version",
                "artifact_id",
                "created_at",
                "created_by",
                "source_manifest_hash",
                "warnings",
            )
            for key in required:
                if key not in artifact:
                    self.error("envelope-field", prefix, f"missing {key}")
            if artifact.get("schema_version") != "1.0.0":
                self.error(
                    "schema-version",
                    prefix,
                    "schema_version must be 1.0.0 for these fixtures",
                )
            artifact_id = artifact.get("artifact_id")
            if not isinstance(artifact_id, str) or not SLUG_RE.fullmatch(artifact_id):
                self.error("artifact-id", prefix, "artifact_id must be a stable slug")
            elif artifact_id in seen_artifact_ids:
                self.error(
                    "artifact-id-duplicate",
                    prefix,
                    f"artifact_id also appears in {seen_artifact_ids[artifact_id]}",
                )
            else:
                seen_artifact_ids[artifact_id] = name
            if not isinstance(artifact.get("created_at"), str) or not RFC3339_RE.fullmatch(
                artifact.get("created_at", "")
            ):
                self.error("created-at", prefix, "created_at must be RFC3339")
            if not isinstance(artifact.get("created_by"), str) or not artifact.get(
                "created_by", ""
            ).strip():
                self.error("created-by", prefix, "created_by must be declared")
            source_hash = artifact.get("source_manifest_hash")
            if source_hash is None and self.provisional():
                self.warning(
                    "source-manifest-hash-provisional",
                    prefix,
                    "source_manifest_hash is null; compute it before freeze",
                )
                warnings = artifact.get("warnings")
                if not (
                    isinstance(warnings, list)
                    and any("hash" in str(item).casefold() for item in warnings)
                ):
                    self.error(
                        "source-manifest-hash-warning",
                        prefix,
                        "a provisional null hash requires a freeze-blocking hash warning",
                    )
            elif not isinstance(source_hash, str) or not HASH_RE.fullmatch(source_hash):
                self.error(
                    "source-manifest-hash",
                    prefix,
                    "source_manifest_hash must be SHA-256-shaped, or null only "
                    "for a warned provisional_directional panel",
                )
            elif (
                isinstance(expected_manifest_hash, str)
                and artifact.get("source_manifest_hash") != expected_manifest_hash
            ):
                self.error(
                    "source-manifest-hash-mismatch",
                    prefix,
                    "source_manifest_hash differs from source_manifest.json",
                )
            if not isinstance(artifact.get("warnings"), list):
                self.error("warnings-shape", prefix, "warnings must be a list")

    def check_sources(self) -> None:
        manifest = self.data.get("source_manifest.json")
        if not isinstance(manifest, dict):
            return
        sources = manifest.get("sources")
        if not isinstance(sources, list) or not sources:
            self.error("sources", "source_manifest.json", "sources must be a non-empty list")
            return
        declared_hash = manifest.get("source_manifest_hash")
        if isinstance(declared_hash, str) and HASH_RE.fullmatch(declared_hash):
            canonical = json.dumps(
                sources,
                ensure_ascii=False,
                sort_keys=True,
                separators=(",", ":"),
            ).encode("utf-8")
            computed_hash = "sha256:" + hashlib.sha256(canonical).hexdigest()
            if declared_hash != computed_hash:
                self.error(
                    "source-manifest-hash-content",
                    "source_manifest.json",
                    f"declared {declared_hash}, computed {computed_hash} from sources",
                )
        for index, source in enumerate(sources):
            path = f"source_manifest.json:sources[{index}]"
            if not isinstance(source, dict):
                self.error("source-shape", path, "source must be a mapping")
                continue
            source_id = source.get("source_id")
            if not isinstance(source_id, str) or not SLUG_RE.fullmatch(source_id):
                self.error("source-id", path, "source_id must be a stable slug")
            elif source_id in self.source_ids:
                self.error("source-id-duplicate", path, f"duplicate {source_id}")
            else:
                self.source_ids.add(source_id)
            self.enum(source.get("source_class"), SOURCE_CLASSES, "source-class", path)
            self.enum(source.get("source_type"), SOURCE_TYPES, "source-type", path)
            self.enum(source.get("evidence_grade"), EVIDENCE_GRADES, "evidence-grade", path)
            self.enum(source.get("confidence"), CONFIDENCE, "confidence", path)
            self.enum(
                source.get("permission"),
                {"public", "user_authorized", "generated"},
                "permission",
                path,
            )
            for field in ("title", "span", "span_locator", "fact_type"):
                if not isinstance(source.get(field), str) or not source.get(field, "").strip():
                    self.error("source-field", path, f"{field} must be a non-empty string")
            if source.get("permission") == "generated":
                if (
                    source.get("source_class") != "llm_hypothesis"
                    or source.get("evidence_grade") != "D"
                    or source.get("url") is not None
                ):
                    self.error(
                        "generated-source-shape",
                        path,
                        "generated sources require source_class=llm_hypothesis, "
                        "evidence_grade=D, and url=null",
                    )
            else:
                for field in ("url", "publisher"):
                    if not isinstance(source.get(field), str) or not source.get(
                        field, ""
                    ).strip():
                        self.error(
                            "source-field",
                            path,
                            f"{field} must be a non-empty string for "
                            "public/user-authorized evidence",
                        )
            if len(source.get("span", "").split()) > 80:
                self.warning(
                    "source-span-length",
                    path,
                    "span exceeds 80 words; keep only a short locatable excerpt/paraphrase",
                )

    def enum(self, value: Any, allowed: set[Any], code: str, path: str) -> None:
        if all(value != item for item in allowed):
            rendered = ", ".join(sorted(str(item) for item in allowed))
            self.error(code, path, f"{value!r} is not one of: {rendered}")

    def check_charter(self) -> None:
        charter = self.data.get("measurement_charter.json")
        if not isinstance(charter, dict):
            return
        for field in (
            "business_decision",
            "target_population",
            "products",
            "markets",
            "locales",
            "time_horizon",
            "surfaces",
            "lanes",
            "reporting_strata",
            "run_budget",
            "review_budget",
            "precision_status",
            "approver",
            "approval_status",
        ):
            if field not in charter:
                self.error("charter-field", "measurement_charter.json", f"missing {field}")
        for lane in list_value(charter.get("lanes")):
            self.enum(lane, LANES, "lane", "measurement_charter.json:lanes")
        estimands = charter.get("estimands")
        if not isinstance(estimands, list) or not estimands:
            self.error("estimands", "measurement_charter.json", "estimands must be non-empty")
            return
        for index, estimand in enumerate(estimands):
            path = f"measurement_charter.json:estimands[{index}]"
            if isinstance(estimand, str):
                self.enum(estimand, ESTIMANDS, "estimand", path)
                self.error(
                    "estimand-denominator",
                    path,
                    "estimand must declare numerator, denominator, eligible "
                    "partitions/lanes, and limits",
                )
                continue
            if not isinstance(estimand, dict):
                self.error("estimand-shape", path, "estimand must be a mapping")
                continue
            self.enum(estimand.get("name"), ESTIMANDS, "estimand", path)
            for field in (
                "numerator",
                "denominator",
                "eligible_partitions",
                "eligible_lanes",
                "does_not_prove",
            ):
                if not estimand.get(field):
                    self.error("estimand-field", path, f"missing or empty {field}")

    def check_reference_ids(
        self,
        artifact_name: str,
        source_keys: set[str],
    ) -> None:
        artifact = self.data.get(artifact_name)
        if not isinstance(artifact, dict):
            return

        def walk(value: Any, path: str) -> None:
            if isinstance(value, dict):
                for key, item in value.items():
                    child_path = f"{path}.{key}"
                    if key in source_keys:
                        for source_id in list_value(item):
                            if source_id not in self.source_ids:
                                self.error(
                                    "source-reference",
                                    child_path,
                                    f"unknown source ID {source_id!r}",
                                )
                    else:
                        walk(item, child_path)
            elif isinstance(value, list):
                for index, item in enumerate(value):
                    walk(item, f"{path}[{index}]")

        walk(artifact, artifact_name)

    def check_icps(self) -> None:
        artifact = self.data.get("icp_hypotheses.json")
        if not isinstance(artifact, dict):
            return
        icps = artifact.get("icps")
        if not isinstance(icps, list):
            self.error("icps", "icp_hypotheses.json", "icps must be a list")
            return
        for index, icp in enumerate(icps):
            path = f"icp_hypotheses.json:icps[{index}]"
            if not isinstance(icp, dict):
                self.error("icp-shape", path, "ICP must be a mapping")
                continue
            icp_id = icp.get("icp_id")
            if not isinstance(icp_id, str) or not SLUG_RE.fullmatch(icp_id):
                self.error("icp-id", path, "icp_id must be a stable slug")
            elif icp_id in self.icp_ids:
                self.error("icp-id-duplicate", path, f"duplicate {icp_id}")
            else:
                self.icp_ids.add(icp_id)
            self.enum(icp.get("confidence"), CONFIDENCE, "confidence", path)
            self.enum(
                icp.get("status"),
                {"supported", "hypothesis_only", "excluded"},
                "icp-status",
                path,
            )
            if not list_value(icp.get("supporting_source_ids")) and icp.get(
                "status"
            ) == "supported":
                self.error(
                    "icp-evidence",
                    path,
                    "supported ICP needs at least one supporting source",
                )

    def check_jobs(self) -> None:
        artifact = self.data.get("buyer_jobs.json")
        if not isinstance(artifact, dict):
            return
        jobs = artifact.get("jobs")
        if not isinstance(jobs, list):
            self.error("jobs", "buyer_jobs.json", "jobs must be a list")
            return
        for index, job in enumerate(jobs):
            path = f"buyer_jobs.json:jobs[{index}]"
            if not isinstance(job, dict):
                self.error("job-shape", path, "job must be a mapping")
                continue
            job_id = job.get("job_id")
            if not isinstance(job_id, str) or not SLUG_RE.fullmatch(job_id):
                self.error("job-id", path, "job_id must be a stable slug")
            elif job_id in self.job_ids:
                self.error("job-id-duplicate", path, f"duplicate {job_id}")
            else:
                self.job_ids.add(job_id)
            for icp_id in list_value(job.get("icp_ids")):
                if icp_id not in self.icp_ids:
                    self.error("icp-reference", path, f"unknown ICP ID {icp_id!r}")
            for information_act in list_value(job.get("information_acts")):
                self.enum(information_act, INFORMATION_ACTS, "information-act", path)
            for journey in list_value(job.get("journey_states")):
                self.enum(journey, JOURNEYS, "journey-state", path)
            self.enum(job.get("evidence_grade"), EVIDENCE_GRADES, "evidence-grade", path)
            self.enum(job.get("confidence"), CONFIDENCE, "confidence", path)
            self.enum(
                job.get("status"),
                {"supported", "hypothesis_only"},
                "job-status",
                path,
            )
            if job.get("status") == "supported" and not list_value(
                job.get("supporting_source_ids")
            ):
                self.error("job-evidence", path, "supported job needs source evidence")

    def check_architecture(self) -> None:
        artifact = self.data.get("prompt_architecture.json")
        if not isinstance(artifact, dict):
            return
        cells = artifact.get("cells")
        if not isinstance(cells, list):
            self.error("architecture-cells", "prompt_architecture.json", "cells must be a list")
            return
        partition_counts: Counter[str] = Counter()
        for index, cell in enumerate(cells):
            path = f"prompt_architecture.json:cells[{index}]"
            if not isinstance(cell, dict):
                self.error("cell-spec-shape", path, "cell spec must be a mapping")
                continue
            spec_id = cell.get("cell_spec_id")
            if not isinstance(spec_id, str) or not SLUG_RE.fullmatch(spec_id):
                self.error("cell-spec-id", path, "cell_spec_id must be a stable slug")
            elif spec_id in self.specs:
                self.error("cell-spec-id-duplicate", path, f"duplicate {spec_id}")
            else:
                self.specs[spec_id] = cell
            if cell.get("job_id") not in self.job_ids:
                self.error("job-reference", path, f"unknown job ID {cell.get('job_id')!r}")
            self.enum(
                cell.get("information_act"),
                INFORMATION_ACTS,
                "information-act",
                path,
            )
            self.enum(
                cell.get("journey_state"),
                JOURNEYS,
                "journey-state",
                path,
            )
            self.enum(cell.get("funnel"), FUNNELS, "funnel", path)
            self.enum(cell.get("proximity_band"), BANDS, "proximity-band", path)
            self.enum(cell.get("aided_status"), AIDED_STATUSES, "aided-status", path)
            self.enum(cell.get("partition"), PARTITIONS, "partition", path)
            self.enum(cell.get("evidence_grade"), EVIDENCE_GRADES, "evidence-grade", path)
            for lane in list_value(cell.get("lane_eligibility")):
                self.enum(lane, LANES, "lane", path)
            partition_counts[str(cell.get("partition"))] += 1
            if (
                cell.get("proximity_band") == "B0_direct_brand_product"
                and cell.get("aided_status") != "target_aided"
            ):
                self.error(
                    "b0-aided-status",
                    path,
                    "B0 direct-brand cell must be target_aided",
                )
            if cell.get("campaign_exposed") is True and cell.get("partition") == "core":
                self.error(
                    "campaign-core",
                    path,
                    "campaign-exposed cell cannot enter evergreen core",
                )
            if cell.get("evidence_grade") == "D" and cell.get("partition") == "core":
                self.error("grade-d-core", path, "grade-D cell cannot enter core")

        allocation = artifact.get("allocation")
        if not isinstance(allocation, dict):
            self.error("allocation", "prompt_architecture.json", "allocation must be a mapping")
            return
        self.enum(
            allocation.get("budget_status"),
            {"within_budget", "conflict", "waived"},
            "budget-status",
            "prompt_architecture.json:allocation",
        )
        allocation_fields = {
            "core": "core_cells",
            "rotating": "rotating_cells",
            "sentinel": "sentinel_cells",
            "control": "control_cells",
            "aided": "aided_cells",
        }
        for partition, field in allocation_fields.items():
            target = allocation.get(field)
            if not isinstance(target, int) or target < 0:
                self.error(
                    "allocation-count",
                    "prompt_architecture.json:allocation",
                    f"{field} must be a non-negative integer",
                )
            elif partition_counts[partition] > target:
                self.error(
                    "allocation-overflow",
                    "prompt_architecture.json:allocation",
                    f"{partition_counts[partition]} {partition} cells exceed {field}={target}",
                )

    def check_universe(self) -> None:
        artifact = self.data.get("prompt_universe.json")
        if not isinstance(artifact, dict):
            return
        cells = artifact.get("canonical_cells")
        if not isinstance(cells, list):
            self.error("canonical-cells", "prompt_universe.json", "canonical_cells must be a list")
            return
        for index, cell in enumerate(cells):
            path = f"prompt_universe.json:canonical_cells[{index}]"
            if not isinstance(cell, dict):
                self.error("canonical-cell-shape", path, "canonical cell must be a mapping")
                continue
            cell_id = cell.get("canonical_cell_id")
            if not isinstance(cell_id, str) or not SLUG_RE.fullmatch(cell_id):
                self.error("canonical-cell-id", path, "canonical_cell_id must be a stable slug")
            elif cell_id in self.cells:
                self.error("canonical-cell-id-duplicate", path, f"duplicate {cell_id}")
            else:
                self.cells[cell_id] = cell
            spec_id = cell.get("cell_spec_id")
            spec = self.specs.get(spec_id)
            if spec is None:
                self.error("cell-spec-reference", path, f"unknown cell spec {spec_id!r}")
                spec = {}
            if cell.get("job_id") not in self.job_ids:
                self.error("job-reference", path, f"unknown job ID {cell.get('job_id')!r}")
            elif spec and cell.get("job_id") != spec.get("job_id"):
                self.error("job-cell-mismatch", path, "canonical cell and spec job IDs differ")
            required_cell_fields = {
                "canonical_cell_id",
                "cell_spec_id",
                "job_id",
                "icp_ids",
                "information_act",
                "journey_state",
                "funnel",
                "proximity_band",
                "aided_status",
                "campaign_exposed",
                "persona_id",
                "locale",
                "language",
                "material_constraints",
                "expected_answer_kind",
                "turn_form",
                "lane_eligibility",
                "partition",
                "evidence_grade",
                "reason_source_ids",
                "candidates",
            }
            missing = sorted(required_cell_fields - set(cell))
            if missing:
                self.error(
                    "canonical-cell-fields",
                    path,
                    f"missing required flat fields: {missing}",
                )
            if spec:
                copied_fields = required_cell_fields - {
                    "canonical_cell_id",
                    "candidates",
                }
                mismatched = sorted(
                    field
                    for field in copied_fields
                    if field in cell
                    and field in spec
                    and cell.get(field) != spec.get(field)
                )
                if mismatched:
                    self.error(
                        "canonical-cell-spec-mismatch",
                        path,
                        f"fields differ from referenced architecture cell: {mismatched}",
                    )
            candidates = cell.get("candidates")
            if not isinstance(candidates, list) or not candidates:
                self.error("candidates", path, "canonical cell needs candidates")
                continue
            for candidate_index, candidate in enumerate(candidates):
                candidate_path = f"{path}:candidates[{candidate_index}]"
                if not isinstance(candidate, dict):
                    self.error("candidate-shape", candidate_path, "candidate must be a mapping")
                    continue
                candidate_id = candidate.get("candidate_id")
                if not isinstance(candidate_id, str) or not SLUG_RE.fullmatch(
                    candidate_id
                ):
                    self.error(
                        "candidate-id",
                        candidate_path,
                        "candidate_id must be a stable slug",
                    )
                elif candidate_id in self.candidates:
                    self.error(
                        "candidate-id-duplicate",
                        candidate_path,
                        f"duplicate {candidate_id}",
                    )
                else:
                    self.candidates[candidate_id] = candidate
                    self.candidate_cells[candidate_id] = {
                        **spec,
                        **{
                            key: value
                            for key, value in cell.items()
                            if key != "candidates"
                        },
                    }
                if not isinstance(candidate.get("text"), str) or not candidate.get(
                    "text", ""
                ).strip():
                    self.error("prompt-text", candidate_path, "text must be non-empty")
                self.enum(
                    candidate.get("transformation"),
                    TRANSFORMATIONS,
                    "transformation",
                    candidate_path,
                )
                self.enum(
                    candidate.get("evidence_grade"),
                    EVIDENCE_GRADES,
                    "evidence-grade",
                    candidate_path,
                )
                for source_id in list_value(candidate.get("source_ids")):
                    if source_id not in self.source_ids:
                        self.error(
                            "source-reference",
                            candidate_path,
                            f"unknown source ID {source_id!r}",
                        )
                provenance = candidate.get("generation_provenance")
                if not isinstance(provenance, dict) or not provenance.get("model"):
                    self.error(
                        "generation-provenance",
                        candidate_path,
                        "generation_provenance.model is required",
                    )
                else:
                    prompt_hash = provenance.get("prompt_hash")
                    if prompt_hash is None and self.provisional():
                        self.warning(
                            "generation-prompt-hash-provisional",
                            candidate_path,
                            "prompt_hash is null; compute it before freeze",
                        )
                    elif not isinstance(prompt_hash, str) or not HASH_RE.fullmatch(
                        prompt_hash
                    ):
                        self.error(
                            "generation-prompt-hash",
                            candidate_path,
                            "generation_provenance.prompt_hash must be "
                            "SHA-256-shaped, or null only for a warned "
                            "provisional_directional panel",
                        )

    def check_qa(self) -> None:
        artifact = self.data.get("prompt_qa.json")
        if not isinstance(artifact, dict):
            return
        if artifact.get("baseline_fields_blinded") is not True:
            self.error(
                "baseline-blinding",
                "prompt_qa.json",
                "baseline_fields_blinded must be true",
            )
        decisions = artifact.get("decisions")
        if not isinstance(decisions, list):
            self.error("qa-decisions", "prompt_qa.json", "decisions must be a list")
            return
        for index, decision in enumerate(decisions):
            path = f"prompt_qa.json:decisions[{index}]"
            if not isinstance(decision, dict):
                self.error("qa-decision-shape", path, "decision must be a mapping")
                continue
            candidate_id = decision.get("candidate_id")
            if candidate_id not in self.candidates:
                self.error("candidate-reference", path, f"unknown candidate {candidate_id!r}")
            if candidate_id in self.decisions:
                    self.error(
                        "qa-decision-duplicate",
                        path,
                        f"duplicate decision for {candidate_id}",
                    )
            elif isinstance(candidate_id, str):
                self.decisions[candidate_id] = decision
            self.enum(decision.get("status"), QA_STATUSES, "qa-status", path)
            self.enum(decision.get("review_confidence"), CONFIDENCE, "confidence", path)
            self.enum(decision.get("route_to"), ROUTES, "route-to", path)
            rule_results = decision.get("rule_results")
            if not isinstance(rule_results, list) or not rule_results:
                self.error("qa-rule-results", path, "rule_results must be non-empty")
            else:
                seen_rules: set[str] = set()
                for rule_index, result in enumerate(rule_results):
                    rule_path = f"{path}:rule_results[{rule_index}]"
                    if not isinstance(result, dict) or not result.get("rule_id"):
                        self.error("qa-rule-result-shape", rule_path, "rule result needs rule_id")
                        continue
                    rule_id = result["rule_id"]
                    if rule_id in seen_rules:
                        self.error("qa-rule-duplicate", rule_path, f"duplicate rule {rule_id}")
                    seen_rules.add(rule_id)
                    self.enum(result.get("status"), RULE_STATUSES, "qa-rule-status", rule_path)
                    if not isinstance(result.get("evidence"), list):
                        self.error("qa-rule-evidence", rule_path, "evidence must be a list")
            duplicate = decision.get("duplicate_decision")
            if not isinstance(duplicate, dict):
                self.error("duplicate-decision", path, "duplicate_decision is required")
            else:
                self.enum(
                    duplicate.get("action"),
                    DUPLICATE_ACTIONS,
                    "duplicate-action",
                    path,
                )
            if not isinstance(decision.get("reason"), str) or not decision.get(
                "reason", ""
            ).strip():
                self.error("qa-reason", path, "reason must be specific and non-empty")

        missing = set(self.candidates) - set(self.decisions)
        for candidate_id in sorted(missing):
            self.error(
                "qa-decision-missing",
                "prompt_qa.json",
                f"no decision for {candidate_id}",
            )
        accepted = artifact.get("accepted_candidate_ids")
        if not isinstance(accepted, list):
            self.error(
                "accepted-candidates",
                "prompt_qa.json",
                "accepted_candidate_ids must be a list",
            )
            return
        if len(set(accepted)) != len(accepted):
            self.error(
                "accepted-candidate-duplicate",
                "prompt_qa.json",
                "accepted_candidate_ids contains duplicates",
            )
        self.accepted_ids = set(accepted)
        for candidate_id in sorted(self.accepted_ids):
            decision = self.decisions.get(candidate_id)
            if decision is None:
                self.error(
                    "accepted-candidate-reference",
                    "prompt_qa.json",
                    f"accepted candidate {candidate_id!r} has no decision",
                )
            elif decision.get("status") != "pass":
                self.error(
                    "accepted-candidate-status",
                    "prompt_qa.json",
                    f"accepted candidate {candidate_id!r} is {decision.get('status')!r}",
                )
        passed = {
            candidate_id
            for candidate_id, decision in self.decisions.items()
            if decision.get("status") == "pass"
        }
        for candidate_id in sorted(passed - self.accepted_ids):
            self.error(
                "passed-not-accepted",
                "prompt_qa.json",
                f"passing candidate {candidate_id!r} is absent from accepted_candidate_ids",
            )
        counts = artifact.get("counts")
        if counts is not None:
            expected_counts = Counter(
                str(decision.get("status"))
                for decision in self.decisions.values()
                if decision.get("status") in QA_STATUSES
            )
            expected = {
                "total_candidates": len(self.decisions),
                "pass": expected_counts["pass"],
                "revise": expected_counts["revise"],
                "quarantine": expected_counts["quarantine"],
                "reject": expected_counts["reject"],
                "accepted": len(self.accepted_ids),
            }
            if not isinstance(counts, dict):
                self.error("qa-counts", "prompt_qa.json", "counts must be a mapping")
            else:
                mismatches = {
                    key: {"expected": value, "actual": counts.get(key)}
                    for key, value in expected.items()
                    if counts.get(key) != value
                }
                if mismatches:
                    self.error(
                        "qa-counts-mismatch",
                        "prompt_qa.json",
                        f"derived counts do not reconcile: {mismatches}",
                    )

    def contamination_terms(self) -> tuple[dict[str, list[str]], list[str]]:
        register = self.data.get("contamination_register.yaml")
        if not isinstance(register, dict):
            return {}, []
        target = register.get("target_terms")
        if not isinstance(target, dict):
            self.error(
                "contamination-target-terms",
                "contamination_register.yaml",
                "target_terms must be a mapping",
            )
            target = {}
        expected_classes = (
            "brands",
            "products",
            "domains",
            "people",
            "slogans",
            "proprietary_categories",
            "campaign_terms",
            "flattering_claims",
        )
        result: dict[str, list[str]] = {}
        for term_class in expected_classes:
            values = target.get(term_class)
            if not isinstance(values, list):
                self.error(
                    "contamination-term-class",
                    "contamination_register.yaml",
                    f"target_terms.{term_class} must be a list",
                )
                result[term_class] = []
            else:
                result[term_class] = [
                    str(value) for value in values if str(value).strip()
                ]
        competitors = register.get("competitor_terms")
        if not isinstance(competitors, list):
            self.error(
                "competitor-terms",
                "contamination_register.yaml",
                "competitor_terms must be a list",
            )
            competitors = []
        if not isinstance(register.get("allowed_exceptions"), list):
            self.error(
                "allowed-exceptions",
                "contamination_register.yaml",
                "allowed_exceptions must be a list",
            )
        return result, [str(value) for value in competitors if str(value).strip()]

    @staticmethod
    def lexical_hits(text: str, terms: Iterable[str]) -> list[str]:
        normalized = f" {normalize_text(text)} "
        hits = []
        for term in terms:
            normalized_term = normalize_text(term)
            if normalized_term and f" {normalized_term} " in normalized:
                hits.append(term)
        return hits

    def check_blinding_and_acceptance(self) -> None:
        target_terms, competitor_terms = self.contamination_terms()
        all_target_terms = [
            term for values in target_terms.values() for term in values
        ]
        brief = self.data.get("blind_design_brief.json")
        if isinstance(brief, dict):
            skip_keys = {
                "schema_version",
                "artifact_id",
                "created_at",
                "created_by",
                "warnings",
                "approval_status",
            }
            blind_text = "\n".join(iter_strings(brief, skip_keys=skip_keys))
            hits = self.lexical_hits(blind_text, all_target_terms)
            if hits:
                self.error(
                    "blind-brief-leak",
                    "blind_design_brief.json",
                    f"target/campaign terms found in blind content: {sorted(set(hits))}",
                )

        normalized_accepted: defaultdict[str, list[str]] = defaultdict(list)
        for candidate_id in sorted(self.accepted_ids):
            candidate = self.candidates.get(candidate_id)
            cell = self.candidate_cells.get(candidate_id, {})
            if not isinstance(candidate, dict):
                continue
            text = str(candidate.get("text", ""))
            normalized_accepted[normalize_text(text)].append(candidate_id)
            is_core_unaided = (
                cell.get("partition") == "core"
                and cell.get("aided_status") == "unaided"
                and cell.get("campaign_exposed") is not True
            )
            if is_core_unaided:
                hits = self.lexical_hits(text, all_target_terms)
                if hits:
                    self.error(
                        "accepted-unaided-leak",
                        f"prompt_universe.json:{candidate_id}",
                        "accepted unaided core includes target/campaign terms: "
                        f"{sorted(set(hits))}",
                    )
                competitor_hits = self.lexical_hits(text, competitor_terms)
                if competitor_hits:
                    self.error(
                        "accepted-competitor-mislabel",
                        f"prompt_universe.json:{candidate_id}",
                        "competitor term requires competitor_aided: "
                        f"{sorted(set(competitor_hits))}",
                    )
            if cell.get("campaign_exposed") is True and cell.get("partition") == "core":
                self.error(
                    "accepted-campaign-core",
                    f"prompt_universe.json:{candidate_id}",
                    "campaign-exposed accepted candidate cannot enter evergreen core",
                )
            if (
                cell.get("partition") == "core"
                and (
                    cell.get("evidence_grade") == "D"
                    or candidate.get("evidence_grade") == "D"
                )
            ):
                self.error(
                    "accepted-grade-d-core",
                    f"prompt_universe.json:{candidate_id}",
                    "accepted grade-D prompt cannot enter core",
                )
            transformation = candidate.get("transformation")
            review_status = candidate.get("locale_review_status")
            if cell.get("partition") == "core" and transformation == "translated":
                self.error(
                    "accepted-machine-translation-core",
                    f"prompt_universe.json:{candidate_id}",
                    "translated candidate must be transcreated and approved before core",
                )
            if (
                cell.get("partition") == "core"
                and transformation == "locale_transcreated"
                and review_status != "approved"
            ):
                self.error(
                    "accepted-locale-review",
                    f"prompt_universe.json:{candidate_id}",
                    "core locale transcreation requires approved review",
                )
            if candidate.get("answer_derived") is True and cell.get("partition") == "core":
                self.error(
                    "accepted-answer-derived-core",
                    f"prompt_universe.json:{candidate_id}",
                    "answer-derived candidate cannot enter core",
                )

        for normalized, candidate_ids in sorted(normalized_accepted.items()):
            if normalized and len(candidate_ids) > 1:
                self.error(
                    "accepted-exact-duplicate",
                    "prompt_qa.json",
                    f"normalized accepted duplicate remains: {candidate_ids}",
                )

    def check_panel(self) -> None:
        panel = self.data.get("panel.yaml")
        if not isinstance(panel, dict):
            return
        if not isinstance(panel.get("panel_id"), str) or not SLUG_RE.fullmatch(
            panel.get("panel_id", "")
        ):
            self.error("panel-id", "panel.yaml", "panel_id must be a stable slug")
        if not isinstance(panel.get("version"), str) or not SEMVER_RE.fullmatch(
            panel.get("version", "")
        ):
            self.error("panel-version", "panel.yaml", "version must be semantic")
        if panel.get("status") not in {
            "provisional_directional",
            "frozen",
            "retired",
        }:
            self.error(
                "panel-status",
                "panel.yaml",
                "status must be provisional_directional, frozen, or retired",
            )

        partition_ids = nested_values(
            panel.get("partitions", {}),
            {"canonical_cell_ids"},
        )
        for cell_id in partition_ids:
            if cell_id not in self.cells:
                self.error("panel-cell-reference", "panel.yaml", f"unknown cell {cell_id!r}")

        selected_ids = set(
            str(value)
            for value in nested_values(
                panel,
                {
                    "selected_candidate_ids",
                    "candidate_ids",
                    "selected_variants",
                },
            )
            if isinstance(value, str)
        )
        for candidate_id in sorted(selected_ids):
            if candidate_id not in self.accepted_ids:
                self.error(
                    "panel-candidate-reference",
                    "panel.yaml",
                    f"selected candidate {candidate_id!r} is not QA-accepted",
                )

        weights = panel.get("weight", panel.get("weights"))
        if not isinstance(weights, dict):
            self.error(
                "panel-weights",
                "panel.yaml",
                "separate exposure and priority weights are required",
            )
        else:
            for kind in ("exposure", "priority"):
                if kind not in weights:
                    self.error(
                        "panel-weight-kind",
                        "panel.yaml",
                        f"missing separate {kind} weight",
                    )

        limitations = list_value(panel.get("limitations"))
        limitation_text = " ".join(str(item) for item in limitations).casefold()
        if "conditional" not in limitation_text or "panel" not in limitation_text:
            self.error(
                "panel-conditional-limit",
                "panel.yaml",
                "limitations must say estimates are conditional on this panel",
            )
        if not any(
            phrase in limitation_text
            for phrase in (
                "non-probability",
                "not a probability",
                "do not represent a probability",
                "coverage",
            )
        ):
            self.error(
                "panel-coverage-limit",
                "panel.yaml",
                "limitations must disclose non-probability or coverage limits",
            )

        statistics = panel.get("statistics")
        if not isinstance(statistics, dict):
            self.error("panel-statistics", "panel.yaml", "statistics are required")
        elif statistics.get("cluster_unit") != "canonical_cell_id":
            self.error(
                "panel-cluster-unit",
                "panel.yaml",
                "variants/repetitions must cluster under canonical_cell_id",
            )

    def check_human_outputs(self) -> None:
        report = self.text.get("panel_report.md", "")
        tracking = self.text.get("tracking_plan.md", "")
        if report and "conditional on this panel" not in report.casefold():
            self.error(
                "report-conditional-label",
                "panel_report.md",
                "report must include 'conditional on this panel'",
            )
        if tracking and "conditional on this panel" not in tracking.casefold():
            self.error(
                "tracking-conditional-label",
                "tracking_plan.md",
                "tracking plan must include 'conditional on this panel'",
            )
        for candidate_id, candidate in sorted(self.candidates.items()):
            text = str(candidate.get("text", "")).strip()
            markdown_table_text = text.replace("|", r"\|")
            if text and text not in report and markdown_table_text not in report:
                self.error(
                    "report-prompt-missing",
                    "panel_report.md",
                    f"exact prompt text missing for {candidate_id}",
                )

    def check_panel_support_files(self) -> None:
        run_manifest = self.data.get("run_manifest_template.json")
        if isinstance(run_manifest, dict):
            required_observation_fields = {
                "exact_prompt",
                "configuration_hash",
                "provider",
                "model",
                "surface",
                "lane",
                "locale",
                "search_policy",
                "retrieval_used",
                "session_state",
                "response_payload_hash",
                "citation_payload_hash",
                "timestamp",
                "retry_status",
                "validity_status",
                "parser_version",
            }
            all_keys: set[str] = set()

            def collect_keys(value: Any) -> None:
                if isinstance(value, dict):
                    all_keys.update(str(key) for key in value)
                    for item in value.values():
                        collect_keys(item)
                elif isinstance(value, list):
                    for item in value:
                        collect_keys(item)

            collect_keys(run_manifest)
            missing = required_observation_fields - all_keys
            if missing:
                self.error(
                    "run-manifest-fields",
                    "run_manifest_template.json",
                    f"missing observation fields: {sorted(missing)}",
                )
        ledger = self.data.get("panel_change_ledger.json")
        if isinstance(ledger, dict):
            if not isinstance(ledger.get("changes"), list):
                self.error(
                    "change-ledger",
                    "panel_change_ledger.json",
                    "changes must be an append-only list",
                )

    def run(self) -> list[Finding]:
        self.load()
        self.check_envelopes()
        self.check_sources()
        self.check_charter()
        source_keys = {
            "source_id",
            "source_ids",
            "supporting_source_ids",
            "counterevidence_source_ids",
            "reason_source_ids",
        }
        for artifact in (
            "icp_hypotheses.json",
            "buyer_jobs.json",
            "blind_design_brief.json",
            "prompt_architecture.json",
            "prompt_universe.json",
            "prompt_qa.json",
            "panel.yaml",
        ):
            self.check_reference_ids(artifact, source_keys)
        self.check_icps()
        self.check_jobs()
        self.check_architecture()
        self.check_universe()
        self.check_qa()
        self.check_blinding_and_acceptance()
        self.check_panel()
        self.check_human_outputs()
        self.check_panel_support_files()
        return self.findings


def grade_gold(
    qa_path: Path,
    gold_path: Path,
) -> tuple[list[Finding], dict[str, dict[str, int]]]:
    findings: list[Finding] = []
    by_category: dict[str, dict[str, int]] = defaultdict(
        lambda: {"correct": 0, "total": 0}
    )
    try:
        qa = json.loads(qa_path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as exc:
        return [
            Finding("error", "gold-qa-read", str(qa_path), str(exc))
        ], {}
    try:
        gold = json.loads(gold_path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as exc:
        return [
            Finding("error", "gold-fixture-read", str(gold_path), str(exc))
        ], {}

    decisions = {
        item.get("candidate_id"): item
        for item in list_value(qa.get("decisions"))
        if isinstance(item, dict) and isinstance(item.get("candidate_id"), str)
    }
    cases = gold.get("cases")
    if not isinstance(cases, list) or not cases:
        return [
            Finding("error", "gold-cases", str(gold_path), "cases must be non-empty")
        ], {}

    for case in cases:
        if not isinstance(case, dict):
            findings.append(
                Finding("error", "gold-case-shape", str(gold_path), "case is not a mapping")
            )
            continue
        candidate_id = case.get("candidate_id")
        category = str(case.get("category", "uncategorized"))
        expected = case.get("expected", {})
        by_category[category]["total"] += 1
        decision = decisions.get(candidate_id)
        case_correct = True
        if decision is None:
            findings.append(
                Finding(
                    "error",
                    "gold-decision-missing",
                    "prompt_qa.json",
                    f"missing gold decision for {candidate_id}",
                )
            )
            continue
        if decision.get("status") != expected.get("status"):
            case_correct = False
            findings.append(
                Finding(
                    "error",
                    "gold-status",
                    f"prompt_qa.json:{candidate_id}",
                    f"expected {expected.get('status')!r}, got {decision.get('status')!r}",
                )
            )
        actual_rules = {
            result.get("rule_id"): result.get("status")
            for result in list_value(decision.get("rule_results"))
            if isinstance(result, dict)
        }
        for rule_id, expected_status in expected.get("rule_results", {}).items():
            if actual_rules.get(rule_id) != expected_status:
                case_correct = False
                findings.append(
                    Finding(
                        "error",
                        "gold-rule",
                        f"prompt_qa.json:{candidate_id}",
                        f"rule {rule_id!r}: expected {expected_status!r}, "
                        f"got {actual_rules.get(rule_id)!r}",
                    )
                )
        duplicate = decision.get("duplicate_decision")
        actual_action = duplicate.get("action") if isinstance(duplicate, dict) else None
        if actual_action != expected.get("duplicate_action"):
            case_correct = False
            findings.append(
                Finding(
                    "error",
                    "gold-duplicate-action",
                    f"prompt_qa.json:{candidate_id}",
                    f"expected {expected.get('duplicate_action')!r}, got {actual_action!r}",
                )
            )
        if decision.get("route_to") != expected.get("route_to"):
            case_correct = False
            findings.append(
                Finding(
                    "error",
                    "gold-route",
                    f"prompt_qa.json:{candidate_id}",
                    f"expected {expected.get('route_to')!r}, "
                    f"got {decision.get('route_to')!r}",
                )
            )
        if case_correct:
            by_category[category]["correct"] += 1

    unexpected = sorted(set(decisions) - {
        case.get("candidate_id") for case in cases if isinstance(case, dict)
    })
    if unexpected:
        findings.append(
            Finding(
                "warning",
                "gold-extra-decisions",
                "prompt_qa.json",
                f"extra decisions ignored: {unexpected}",
            )
        )
    return findings, dict(sorted(by_category.items()))


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate an AI visibility panel output directory."
    )
    parser.add_argument("output_dir", type=Path)
    parser.add_argument(
        "--gold",
        type=Path,
        help="grade output_dir/prompt_qa.json against the adversarial gold fixture",
    )
    parser.add_argument(
        "--qa-only",
        action="store_true",
        help="skip end-to-end validation and grade only prompt_qa.json",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        dest="json_output",
        help="emit a machine-readable summary",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    findings: list[Finding] = []
    categories: dict[str, dict[str, int]] = {}

    if args.qa_only and args.gold is None:
        findings.append(
            Finding("error", "qa-only-without-gold", "--qa-only", "--gold is required")
        )
    if not args.qa_only:
        findings.extend(Validator(args.output_dir).run())
    if args.gold is not None:
        gold_findings, categories = grade_gold(
            args.output_dir / "prompt_qa.json",
            args.gold,
        )
        findings.extend(gold_findings)

    findings = sorted(
        findings,
        key=lambda item: (item.level != "error", item.code, item.path, item.message),
    )
    errors = sum(item.level == "error" for item in findings)
    warnings = sum(item.level == "warning" for item in findings)
    passed = errors == 0

    if args.json_output:
        payload = {
            "status": "pass" if passed else "fail",
            "errors": errors,
            "warnings": warnings,
            "findings": [asdict(item) for item in findings],
            "gold_categories": categories,
        }
        print(json.dumps(payload, indent=2, sort_keys=True))
    else:
        for item in findings:
            print(
                f"{item.level.upper()} [{item.code}] {item.path}: {item.message}"
            )
        if categories:
            print("Gold categories:")
            for category, result in categories.items():
                print(f"  {category}: {result['correct']}/{result['total']}")
        print(
            f"{'PASS' if passed else 'FAIL'}: {errors} error(s), "
            f"{warnings} warning(s)"
        )
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
