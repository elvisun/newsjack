#!/usr/bin/env python3
"""Deterministically grade AI visibility panel eval artifacts.

Usage:
    python3 grade.py RUN_DIR [--cases FILE] [--gold FILE] [--out FILE]
                            [--ids ID1,ID2] [--split NAME] [--require-judge]

The grader first enforces the shared artifact contract, graph references,
contamination boundaries, enum consistency, and panel invariants. It then
applies any machine-checkable per-case assertions in gold.json. Judgment-only
assertions remain explicitly pending unless a judge verdict exists.
"""

from __future__ import annotations

import argparse
import collections
import json
import math
import re
import sys
from pathlib import Path
from typing import Any, Iterable

sys.path.insert(0, str(Path(__file__).resolve().parent))
import _common as common  # noqa: E402


SHA_RE = re.compile(r"^sha256:[0-9a-f]{64}$")
RFC3339_RE = re.compile(
    r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$"
)

ENUMS = {
    "information_act": {
        "explain", "diagnose", "plan", "generate", "compare", "recommend",
        "verify", "navigate", "buy", "implement", "troubleshoot",
    },
    "journey_state": {
        "problem_identification", "exploration", "requirements_building",
        "supplier_selection", "adoption", "post_purchase",
    },
    "funnel": {"TOFU", "MOFU", "BOFU", None},
    "proximity_band": {
        "B0_direct_brand_product", "B1_comparison_purchase", "B2_category",
        "B3_problem_need", "B4_job_goal", "B5_broad_discovery_story",
    },
    "aided_status": {"target_aided", "competitor_aided", "category_aided", "unaided"},
    "lane": {"closed_model", "retrieval", "consumer_surface", "campaign_experiment"},
    "partition": {"core", "rotating", "sentinel", "control", "aided"},
    "transformation": {
        "verbatim", "lightly_normalized", "search_query_expanded",
        "human_written", "llm_expanded", "translated", "locale_transcreated",
    },
    "evidence_grade": {"A", "B", "C", "D"},
}


class Checks:
    def __init__(self) -> None:
        self.items: list[dict[str, Any]] = []

    def add(
        self,
        ident: str,
        passed: bool,
        evidence: Any,
        *,
        text: str | None = None,
        severity: str = "error",
    ) -> None:
        self.items.append({
            "id": ident,
            "text": text or ident.replace("_", " "),
            "passed": bool(passed),
            "severity": severity,
            "evidence": evidence,
        })

    @property
    def passed(self) -> bool:
        return all(
            item["passed"]
            for item in self.items
            if item["severity"] in ("error", "critical")
        )


def as_list(value: Any) -> list[Any]:
    if value is None:
        return []
    return value if isinstance(value, list) else [value]


def unique_nonempty(values: Iterable[Any]) -> set[str]:
    return {str(value) for value in values if value is not None and str(value).strip()}


def normalize(value: str) -> str:
    return re.sub(r"\s+", " ", re.sub(r"[^\w]+", " ", value.casefold())).strip()


def term_in_text(term: str, text: str) -> bool:
    needle = normalize(term)
    haystack = f" {normalize(text)} "
    return bool(needle and f" {needle} " in haystack)


def recursive_values(value: Any, keys: set[str]) -> list[Any]:
    found: list[Any] = []
    if isinstance(value, dict):
        for key, child in value.items():
            if key in keys:
                found.extend(as_list(child))
            found.extend(recursive_values(child, keys))
    elif isinstance(value, list):
        for child in value:
            found.extend(recursive_values(child, keys))
    return found


def recursive_source_refs(value: Any) -> set[str]:
    refs: set[str] = set()
    if isinstance(value, dict):
        for key, child in value.items():
            if (
                key in {"source_id", "source_ids", "reason_source_ids", "supporting_source_ids",
                        "counterevidence_source_ids"}
                or key.endswith("_source_ids")
            ):
                refs |= unique_nonempty(as_list(child))
            refs |= recursive_source_refs(child)
    elif isinstance(value, list):
        for child in value:
            refs |= recursive_source_refs(child)
    return refs


def flatten_universe(universe: dict[str, Any]) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    cells = universe.get("canonical_cells")
    if not isinstance(cells, list):
        return [], []
    candidates: list[dict[str, Any]] = []
    valid_cells: list[dict[str, Any]] = []
    for raw_cell in cells:
        if not isinstance(raw_cell, dict):
            continue
        valid_cells.append(raw_cell)
        for raw_candidate in as_list(raw_cell.get("candidates")):
            if not isinstance(raw_candidate, dict):
                continue
            merged = dict(raw_cell)
            merged.pop("candidates", None)
            merged.update(raw_candidate)
            merged["_cell"] = raw_cell
            candidates.append(merged)
    return valid_cells, candidates


def decision_map(qa: dict[str, Any]) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for decision in as_list(qa.get("decisions")):
        if isinstance(decision, dict) and decision.get("candidate_id") is not None:
            result[str(decision["candidate_id"])] = decision
    return result


def accepted_ids(qa: dict[str, Any], decisions: dict[str, dict[str, Any]]) -> set[str]:
    explicit = unique_nonempty(as_list(qa.get("accepted_candidate_ids")))
    if explicit:
        return explicit
    return {
        ident for ident, decision in decisions.items()
        if decision.get("status") == "pass"
    }


def validate_envelopes(artifacts: dict[str, Any], checks: Checks) -> None:
    machine_names = common.JSON_ARTIFACTS + common.YAML_ARTIFACTS
    hashes: dict[str, str] = {}
    panel = artifacts.get("panel.yaml")
    provisional = (
        isinstance(panel, dict)
        and panel.get("status") == "provisional_directional"
    )
    for name in machine_names:
        value = artifacts.get(name)
        if not isinstance(value, dict):
            continue
        missing = [
            key for key in (
                "schema_version", "artifact_id", "created_at", "created_by",
                "source_manifest_hash", "warnings",
            )
            if key not in value
        ]
        checks.add(f"envelope:{name}", not missing, missing or "complete")
        if missing:
            continue
        checks.add(
            f"schema_version:{name}",
            value.get("schema_version") == "1.0.0",
            value.get("schema_version"),
        )
        checks.add(
            f"created_at:{name}",
            isinstance(value.get("created_at"), str)
            and bool(RFC3339_RE.match(value["created_at"])),
            value.get("created_at"),
        )
        source_hash = value.get("source_manifest_hash")
        hash_valid = (
            isinstance(source_hash, str)
            and bool(SHA_RE.match(source_hash))
        )
        provisional_null = provisional and source_hash is None
        checks.add(
            f"source_hash_shape:{name}",
            hash_valid or provisional_null,
            (
                source_hash
                if not provisional_null
                else "null; permitted provisionally and required before freeze"
            ),
            severity="warning" if provisional_null else "error",
        )
        if provisional_null:
            warnings = value.get("warnings")
            checks.add(
                f"source_hash_warning:{name}",
                isinstance(warnings, list)
                and any("hash" in str(item).casefold() for item in warnings),
                warnings,
            )
        checks.add(
            f"warnings_array:{name}",
            isinstance(value.get("warnings"), list),
            type(value.get("warnings")).__name__,
        )
        if isinstance(value.get("source_manifest_hash"), str):
            hashes[name] = value["source_manifest_hash"]
    distinct = sorted(set(hashes.values()))
    checks.add(
        "source_manifest_hash_consistent",
        len(distinct) <= 1,
        {"distinct_hashes": distinct, "artifacts": hashes},
    )


def validate_source_manifest(
    artifacts: dict[str, Any], checks: Checks
) -> tuple[set[str], dict[str, dict[str, Any]]]:
    manifest = artifacts.get("source_manifest.json")
    sources = manifest.get("sources") if isinstance(manifest, dict) else None
    if not isinstance(sources, list):
        checks.add("sources_array", False, "source_manifest.sources is not an array")
        return set(), {}
    declared_hash = manifest.get("source_manifest_hash")
    if isinstance(declared_hash, str) and SHA_RE.match(declared_hash):
        computed_hash = "sha256:" + common.canonical_hash(sources)
        checks.add(
            "source_manifest_hash_content",
            declared_hash == computed_hash,
            {"declared": declared_hash, "computed": computed_hash},
        )
    source_map: dict[str, dict[str, Any]] = {}
    bad: list[str] = []
    for index, source in enumerate(sources):
        if not isinstance(source, dict) or not source.get("source_id"):
            bad.append(f"index {index}: missing source_id")
            continue
        ident = str(source["source_id"])
        if ident in source_map:
            bad.append(f"duplicate {ident}")
        source_map[ident] = source
        permission = source.get("permission")
        if permission not in {"public", "user_authorized", "generated"}:
            bad.append(f"{ident}: invalid permission")
        if source.get("evidence_grade") not in ENUMS["evidence_grade"]:
            bad.append(f"{ident}: invalid evidence_grade")
        if source.get("source_class") not in {
            "company_asserted", "buyer_behavior", "independent", "search_proxy", "llm_hypothesis",
        }:
            bad.append(f"{ident}: invalid source_class")
        if permission == "generated":
            if (
                source.get("source_class") != "llm_hypothesis"
                or source.get("evidence_grade") != "D"
                or source.get("url") is not None
            ):
                bad.append(
                    f"{ident}: generated sources require llm_hypothesis, grade D, url null"
                )
        elif not source.get("url"):
            bad.append(f"{ident}: evidence source missing URL/locator")
        if not source.get("span") or not source.get("span_locator"):
            bad.append(f"{ident}: missing short evidence span/locator")
    checks.add("source_records_valid", not bad, bad or f"{len(source_map)} sources")

    counts = collections.Counter(
        str(source.get("source_class")) for source in source_map.values()
    )
    checks.add(
        "source_mix_target",
        counts["company_asserted"] >= 1,
        dict(counts),
        severity="warning",
    )
    checks.add(
        "source_mix_independent",
        counts["independent"] >= 2,
        dict(counts),
        severity="warning",
    )
    checks.add(
        "source_mix_buyer_language",
        counts["buyer_behavior"] + counts["search_proxy"] >= 3,
        dict(counts),
        severity="warning",
    )
    return set(source_map), source_map


def validate_graph(artifacts: dict[str, Any], source_ids: set[str], checks: Checks) -> dict[str, Any]:
    icp = artifacts.get("icp_hypotheses.json", {})
    jobs_art = artifacts.get("buyer_jobs.json", {})
    architecture = artifacts.get("prompt_architecture.json", {})
    universe = artifacts.get("prompt_universe.json", {})
    qa = artifacts.get("prompt_qa.json", {})
    if not all(isinstance(item, dict) for item in (icp, jobs_art, architecture, universe, qa)):
        return {}

    icps = [item for item in as_list(icp.get("icps")) if isinstance(item, dict)]
    jobs = [item for item in as_list(jobs_art.get("jobs")) if isinstance(item, dict)]
    specs = [item for item in as_list(architecture.get("cells")) if isinstance(item, dict)]
    cells, candidates = flatten_universe(universe)
    decisions = decision_map(qa)
    accepted = accepted_ids(qa, decisions)

    id_groups = {
        "icp_id": [item.get("icp_id") for item in icps],
        "job_id": [item.get("job_id") for item in jobs],
        "cell_spec_id": [item.get("cell_spec_id") for item in specs],
        "canonical_cell_id": [item.get("canonical_cell_id") for item in cells],
        "candidate_id": [item.get("candidate_id") for item in candidates],
    }
    id_sets: dict[str, set[str]] = {}
    for key, values in id_groups.items():
        present = [str(value) for value in values if value is not None and str(value)]
        id_sets[key] = set(present)
        checks.add(
            f"unique:{key}",
            len(present) == len(set(present)) and len(present) == len(values),
            {"count": len(present), "unique": len(set(present)), "records": len(values)},
        )

    all_source_refs: set[str] = set()
    for name, artifact in artifacts.items():
        if name != "source_manifest.json" and isinstance(artifact, (dict, list)):
            all_source_refs |= recursive_source_refs(artifact)
    unresolved_sources = sorted(all_source_refs - source_ids)
    checks.add(
        "source_references_resolve",
        not unresolved_sources,
        unresolved_sources or f"{len(all_source_refs)} references",
    )

    job_icp_refs = unique_nonempty(
        value for job in jobs for value in as_list(job.get("icp_ids"))
    )
    checks.add(
        "job_icp_references_resolve",
        job_icp_refs <= id_sets["icp_id"],
        sorted(job_icp_refs - id_sets["icp_id"]) or "resolved",
    )
    spec_job_refs = unique_nonempty(spec.get("job_id") for spec in specs)
    checks.add(
        "architecture_job_references_resolve",
        spec_job_refs <= id_sets["job_id"],
        sorted(spec_job_refs - id_sets["job_id"]) or "resolved",
    )
    cell_spec_refs = unique_nonempty(cell.get("cell_spec_id") for cell in cells)
    cell_job_refs = unique_nonempty(cell.get("job_id") for cell in cells)
    checks.add(
        "universe_spec_references_resolve",
        cell_spec_refs <= id_sets["cell_spec_id"],
        sorted(cell_spec_refs - id_sets["cell_spec_id"]) or "resolved",
    )
    checks.add(
        "universe_job_references_resolve",
        cell_job_refs <= id_sets["job_id"],
        sorted(cell_job_refs - id_sets["job_id"]) or "resolved",
    )

    candidate_ids = id_sets["candidate_id"]
    checks.add(
        "qa_one_decision_per_candidate",
        set(decisions) == candidate_ids,
        {
            "missing": sorted(candidate_ids - set(decisions)),
            "unknown": sorted(set(decisions) - candidate_ids),
        },
    )
    checks.add(
        "qa_accepted_ids_resolve",
        accepted <= candidate_ids,
        sorted(accepted - candidate_ids) or "resolved",
    )
    invalid_accepts = sorted(
        ident for ident in accepted
        if decisions.get(ident, {}).get("status") != "pass"
    )
    checks.add("qa_accepted_are_pass", not invalid_accepts, invalid_accepts or "all pass")
    counts = qa.get("counts")
    if counts is not None:
        status_counts = collections.Counter(
            str(decision.get("status"))
            for decision in decisions.values()
            if decision.get("status") in {"pass", "revise", "quarantine", "reject"}
        )
        expected_counts = {
            "total_candidates": len(decisions),
            "pass": status_counts["pass"],
            "revise": status_counts["revise"],
            "quarantine": status_counts["quarantine"],
            "reject": status_counts["reject"],
            "accepted": len(accepted),
        }
        if isinstance(counts, dict):
            mismatches = {
                key: {"expected": expected, "actual": counts.get(key)}
                for key, expected in expected_counts.items()
                if counts.get(key) != expected
            }
        else:
            mismatches = {"shape": type(counts).__name__}
        checks.add(
            "qa_counts_reconcile",
            isinstance(counts, dict) and not mismatches,
            mismatches or expected_counts,
        )
    checks.add(
        "qa_baseline_blinded",
        qa.get("baseline_fields_blinded") is True,
        qa.get("baseline_fields_blinded"),
    )

    required_cell_fields = {
        "canonical_cell_id", "cell_spec_id", "job_id", "icp_ids",
        "information_act", "journey_state", "funnel", "proximity_band",
        "aided_status", "campaign_exposed", "persona_id", "locale", "language",
        "material_constraints", "expected_answer_kind", "turn_form",
        "lane_eligibility", "partition", "evidence_grade", "reason_source_ids",
        "candidates",
    }
    missing_cell_fields = {
        str(cell.get("canonical_cell_id", f"index-{index}")): sorted(required_cell_fields - set(cell))
        for index, cell in enumerate(cells)
        if required_cell_fields - set(cell)
    }
    checks.add("canonical_cell_fields_complete", not missing_cell_fields, missing_cell_fields or "complete")

    required_candidate_fields = {
        "candidate_id", "variant_role", "text", "transformation", "source_ids",
        "evidence_grade", "locale_review_status", "generation_provenance",
    }
    missing_candidate_fields = {
        str(candidate.get("candidate_id", f"index-{index}")):
            sorted(required_candidate_fields - set(candidate))
        for index, candidate in enumerate(candidates)
        if required_candidate_fields - set(candidate)
    }
    checks.add(
        "candidate_fields_complete",
        not missing_candidate_fields,
        missing_candidate_fields or "complete",
    )

    enum_errors: list[str] = []
    for record in candidates:
        ident = str(record.get("candidate_id"))
        for field in ("information_act", "journey_state", "funnel", "proximity_band",
                      "aided_status", "partition", "transformation", "evidence_grade"):
            if record.get(field) not in ENUMS[field]:
                enum_errors.append(f"{ident}: {field}={record.get(field)!r}")
        lanes = as_list(record.get("lane_eligibility"))
        invalid_lanes = [lane for lane in lanes if lane not in ENUMS["lane"]]
        if invalid_lanes:
            enum_errors.append(f"{ident}: invalid lanes {invalid_lanes}")
    checks.add("prompt_enums_valid", not enum_errors, enum_errors or "valid")

    consistency_errors: list[str] = []
    for record in candidates:
        ident = str(record.get("candidate_id"))
        band = record.get("proximity_band")
        aided = record.get("aided_status")
        partition = record.get("partition")
        campaign = record.get("campaign_exposed")
        if band == "B0_direct_brand_product" and aided != "target_aided":
            consistency_errors.append(f"{ident}: B0 must be target_aided")
        if aided == "target_aided" and band not in {
            "B0_direct_brand_product",
            "B1_comparison_purchase",
        }:
            consistency_errors.append(f"{ident}: target_aided outside B0/B1")
        if aided == "competitor_aided" and band != "B1_comparison_purchase":
            consistency_errors.append(f"{ident}: competitor_aided outside B1")
        if aided == "category_aided" and band != "B2_category":
            consistency_errors.append(f"{ident}: category_aided outside B2")
        if aided == "target_aided" and partition != "aided":
            consistency_errors.append(f"{ident}: target_aided must use aided partition")
        if campaign is True and partition == "core":
            consistency_errors.append(f"{ident}: campaign-exposed prompt in core")
    checks.add(
        "band_aided_partition_consistent",
        not consistency_errors,
        consistency_errors or "consistent",
    )

    accepted_records = [
        record for record in candidates if str(record.get("candidate_id")) in accepted
    ]
    normalized = collections.defaultdict(list)
    for record in accepted_records:
        text = record.get("text")
        if isinstance(text, str):
            normalized[normalize(text)].append(str(record.get("candidate_id")))
    duplicates = {text: ids for text, ids in normalized.items() if text and len(ids) > 1}
    checks.add("accepted_prompts_exact_unique", not duplicates, duplicates or "unique")

    grade_d_core = []
    for record in accepted_records:
        if record.get("partition") != "core" or record.get("evidence_grade") != "D":
            continue
        promotion = record.get("promotion")
        promoted = record.get("promoted") is True or (
            isinstance(promotion, dict) and promotion.get("approved") is True
        )
        if not promoted:
            grade_d_core.append(str(record.get("candidate_id")))
    checks.add("no_unpromoted_grade_d_core", not grade_d_core, grade_d_core or "none")

    return {
        "icps": icps,
        "jobs": jobs,
        "specs": specs,
        "cells": cells,
        "candidates": candidates,
        "accepted_ids": accepted,
        "accepted": accepted_records,
        "decisions": decisions,
        "id_sets": id_sets,
    }


def contamination_terms(register: dict[str, Any]) -> dict[str, list[str]]:
    target = register.get("target_terms")
    target = target if isinstance(target, dict) else {}
    result = {
        "target": [],
        "campaign": [str(item) for item in as_list(target.get("campaign_terms")) if str(item)],
        "competitor": [str(item) for item in as_list(register.get("competitor_terms")) if str(item)],
    }
    for key, values in target.items():
        if key != "campaign_terms":
            result["target"].extend(str(item) for item in as_list(values) if str(item))
    return result


def validate_contamination(
    artifacts: dict[str, Any],
    graph: dict[str, Any],
    checks: Checks,
) -> None:
    register = artifacts.get("contamination_register.yaml")
    if not isinstance(register, dict):
        return
    terms = contamination_terms(register)
    leaks: list[str] = []
    for record in graph.get("accepted", []):
        ident = str(record.get("candidate_id"))
        text = str(record.get("text", ""))
        aided = record.get("aided_status")
        partition = record.get("partition")
        campaign_exposed = record.get("campaign_exposed") is True
        if aided != "target_aided":
            for term in terms["target"]:
                if term_in_text(term, text):
                    leaks.append(f"{ident}: target term {term!r} outside target_aided")
        competitor_stimulus_allowed = (
            aided == "target_aided"
            and record.get("proximity_band") == "B1_comparison_purchase"
            and partition == "aided"
        )
        if aided != "competitor_aided" and not competitor_stimulus_allowed:
            for term in terms["competitor"]:
                if term_in_text(term, text):
                    leaks.append(f"{ident}: competitor term {term!r} outside competitor_aided")
        if not campaign_exposed:
            for term in terms["campaign"]:
                if term_in_text(term, text):
                    leaks.append(f"{ident}: campaign term {term!r} without campaign_exposed")
        if aided == "unaided" and partition == "core":
            for term in terms["target"] + terms["campaign"] + terms["competitor"]:
                if term_in_text(term, text):
                    leaks.append(f"{ident}: unaided-core leak {term!r}")
    checks.add("accepted_prompt_contamination_zero", not leaks, leaks or "zero leaks")

    brief = artifacts.get("blind_design_brief.json")
    serialized = json.dumps(brief, ensure_ascii=False) if isinstance(brief, dict) else ""
    brief_leaks = sorted({
        term for term in terms["target"] + terms["campaign"]
        if term_in_text(term, serialized)
    })
    checks.add("blind_brief_target_free", not brief_leaks, brief_leaks or "target-free")


def validate_panel(
    artifacts: dict[str, Any],
    graph: dict[str, Any],
    checks: Checks,
) -> None:
    panel = artifacts.get("panel.yaml")
    if not isinstance(panel, dict):
        return
    weight = panel.get("weight", panel.get("weights"))
    weight_ok = isinstance(weight, dict) and "exposure" in weight and "priority" in weight
    checks.add("weights_exposure_priority_separate", weight_ok, weight if weight_ok else "missing")

    selected = unique_nonempty(recursive_values(
        panel,
        {"selected_candidate_ids", "candidate_ids", "selected_candidates"},
    ))
    if not selected:
        # Some panel schemas store candidate IDs one at a time.
        selected = unique_nonempty(recursive_values(panel, {"candidate_id"}))
    checks.add(
        "panel_selected_candidates_present",
        bool(selected),
        sorted(selected) if selected else "none",
    )
    accepted = set(graph.get("accepted_ids", set()))
    checks.add(
        "panel_selected_candidates_qa_approved",
        bool(selected) and selected <= accepted,
        sorted(selected - accepted) if selected else "none selected",
    )

    report = artifacts.get("panel_report.md")
    tracking = artifacts.get("tracking_plan.md")
    joined = "\n".join(text for text in (report, tracking) if isinstance(text, str))
    checks.add(
        "conditional_panel_language",
        "conditional on this panel" in joined.casefold(),
        "phrase present" if "conditional on this panel" in joined.casefold() else "phrase missing",
    )
    missing_report_ids = sorted(
        ident for ident in accepted
        if isinstance(report, str) and ident not in report
    )
    checks.add(
        "report_lists_accepted_prompts",
        not missing_report_ids,
        missing_report_ids or f"{len(accepted)} accepted IDs present",
    )
    missing_report_prompts = sorted(
        str(record.get("candidate_id"))
        for record in graph.get("accepted", [])
        if isinstance(report, str)
        and isinstance(record.get("text"), str)
        and record["text"] not in report
        and record["text"].replace("|", r"\|") not in report
    )
    checks.add(
        "report_lists_exact_prompt_text",
        not missing_report_prompts,
        missing_report_prompts or f"{len(accepted)} exact prompt texts present",
    )

    status = panel.get("status")
    approvals = panel.get("approvals")
    approvals_pending = False
    if isinstance(approvals, list):
        approvals_pending = any(
            isinstance(item, dict) and item.get("status") == "pending"
            for item in approvals
        )
    elif isinstance(approvals, dict):
        approvals_pending = any(
            value == "pending"
            or (isinstance(value, dict) and value.get("status") == "pending")
            for value in approvals.values()
        )
    checks.add(
        "pending_gates_not_frozen",
        not approvals_pending or status in {"provisional_directional", "provisional"},
        {"status": status, "approvals_pending": approvals_pending},
    )


def _path_parts(path: str) -> list[str]:
    return [part for part in path.replace("/", ".").split(".") if part and part != "$"]


def resolve_path(root: Any, path: str | None) -> list[Any]:
    if not path:
        return [root]
    current = [root]
    for part in _path_parts(path):
        match = re.fullmatch(r"([^\[]+)?(?:\[(\*|\d+)\])?", part)
        if not match:
            return []
        key, index = match.groups()
        next_values: list[Any] = []
        for value in current:
            targets = value if isinstance(value, list) and key is None else [value]
            for target in targets:
                child = target.get(key) if key and isinstance(target, dict) else target
                if key and not isinstance(target, dict):
                    continue
                if index == "*":
                    if isinstance(child, list):
                        next_values.extend(child)
                elif index is not None:
                    if isinstance(child, list) and int(index) < len(child):
                        next_values.append(child[int(index)])
                elif isinstance(child, list):
                    next_values.extend(child)
                elif child is not None:
                    next_values.append(child)
        current = next_values
    return current


def _haystack(values: list[Any]) -> str:
    return json.dumps(values, ensure_ascii=False, sort_keys=True).casefold()


def evaluate_check(check: dict[str, Any], root: Any) -> tuple[bool, Any]:
    op = check.get("op", "exists")
    if op in {"all", "any"}:
        results = [evaluate_check(child, root) for child in as_list(check.get("checks"))]
        passed = all(item[0] for item in results) if op == "all" else any(item[0] for item in results)
        return passed, [item[1] for item in results]
    if op == "not":
        passed, evidence = evaluate_check(check.get("check", {}), root)
        return not passed, evidence

    values = resolve_path(root, check.get("path"))
    expected = check.get("value")
    if op == "exists":
        return bool(values), values
    if op == "not_exists":
        return not values, values
    if op == "==":
        return any(value == expected for value in values), values
    if op == "!=":
        return bool(values) and all(value != expected for value in values), values
    if op == "in":
        allowed = as_list(expected)
        return any(value in allowed for value in values), values
    if op == "contains":
        return str(expected).casefold() in _haystack(values), values
    if op == "contains_any":
        needles = [str(item).casefold() for item in as_list(expected)]
        haystack = _haystack(values)
        return any(needle in haystack for needle in needles), values
    if op == "contains_all":
        needles = [str(item).casefold() for item in as_list(expected)]
        haystack = _haystack(values)
        return all(needle in haystack for needle in needles), values
    if op == "regex":
        pattern = re.compile(str(expected), re.IGNORECASE)
        return any(pattern.search(str(value)) for value in values), values
    if op.startswith("count_"):
        count = len(values)
        relation = op.removeprefix("count_")
        return compare_number(count, relation, expected), {"count": count, "values": values}
    if op in {"<", "<=", ">", ">=", "between"}:
        numeric = [value for value in values if isinstance(value, (int, float)) and not isinstance(value, bool)]
        return any(compare_number(value, op, expected) for value in numeric), numeric
    raise ValueError(f"unsupported assertion op: {op}")


def compare_number(value: float, op: str, expected: Any) -> bool:
    if op in ("=", "=="):
        return value == expected
    if op == "<":
        return value < expected
    if op == "<=":
        return value <= expected
    if op == ">":
        return value > expected
    if op == ">=":
        return value >= expected
    if op == "between":
        low, high = expected
        return low <= value <= high
    return False


def apply_gold(
    case: dict[str, Any],
    gold: dict[str, Any],
    artifacts: dict[str, Any],
    graph: dict[str, Any],
    checks: Checks,
) -> list[dict[str, Any]]:
    context = {
        "case": case,
        "gold": gold,
        "artifacts": {
            name: value for name, value in artifacts.items()
        },
        **{
            Path(name).stem: value
            for name, value in artifacts.items()
        },
    }
    pending: list[dict[str, Any]] = []
    for index, assertion in enumerate(as_list(gold.get("assertions"))):
        if not isinstance(assertion, dict):
            continue
        ident = str(assertion.get("id", f"gold-{index + 1}"))
        machine_check = assertion.get("check")
        if not isinstance(machine_check, dict):
            pending.append(assertion)
            continue
        artifact_name = machine_check.get("artifact")
        root = artifacts.get(str(artifact_name), context) if artifact_name else context
        check_spec = dict(machine_check)
        check_spec.pop("artifact", None)
        try:
            passed, evidence = evaluate_check(check_spec, root)
        except Exception as exc:  # noqa: BLE001
            passed, evidence = False, f"assertion error: {exc}"
        checks.add(
            f"gold:{ident}",
            passed,
            evidence,
            text=str(assertion.get("text", ident)),
            severity=str(assertion.get("severity", "error")),
        )

    required = gold.get("required_coverage")
    if not isinstance(required, dict):
        expected = gold.get("expected")
        required = expected.get("required_coverage") if isinstance(expected, dict) else {}
    required = required if isinstance(required, dict) else {}
    candidates = graph.get("accepted", [])
    dimensions = {
        "bands": unique_nonempty(item.get("proximity_band") for item in candidates),
        "proximity_bands": unique_nonempty(item.get("proximity_band") for item in candidates),
        "aided_statuses": unique_nonempty(item.get("aided_status") for item in candidates),
        "information_acts": unique_nonempty(item.get("information_act") for item in candidates),
        "journey_states": unique_nonempty(item.get("journey_state") for item in candidates),
        "partitions": unique_nonempty(item.get("partition") for item in candidates),
        "locales": unique_nonempty(item.get("locale") for item in candidates),
        "lanes": unique_nonempty(
            lane for item in candidates for lane in as_list(item.get("lane_eligibility"))
        ),
        "evidence_grades": unique_nonempty(item.get("evidence_grade") for item in candidates),
        "transformations": unique_nonempty(item.get("transformation") for item in candidates),
    }
    for key, expected_values in required.items():
        if key not in dimensions or not isinstance(expected_values, list):
            continue
        wanted = unique_nonempty(expected_values)
        missing = sorted(wanted - dimensions[key])
        checks.add(
            f"gold:coverage:{key}",
            not missing,
            {"missing": missing, "actual": sorted(dimensions[key])},
        )

    forbidden = as_list(gold.get("forbidden_prompt_terms"))
    if forbidden:
        leaks = [
            f"{item.get('candidate_id')}:{term}"
            for item in candidates
            for term in forbidden
            if term_in_text(str(term), str(item.get("text", "")))
        ]
        checks.add("gold:forbidden_prompt_terms", not leaks, leaks or "none")
    forbidden_unaided = as_list(gold.get("forbidden_unaided_terms"))
    if forbidden_unaided:
        leaks = [
            f"{item.get('candidate_id')}:{term}"
            for item in candidates
            if item.get("aided_status") == "unaided"
            for term in forbidden_unaided
            if term_in_text(str(term), str(item.get("text", "")))
        ]
        checks.add("gold:forbidden_unaided_terms", not leaks, leaks or "none")

    count = len(candidates)
    if isinstance(gold.get("min_prompt_count"), int):
        checks.add("gold:min_prompt_count", count >= gold["min_prompt_count"], count)
    if isinstance(gold.get("max_prompt_count"), int):
        checks.add("gold:max_prompt_count", count <= gold["max_prompt_count"], count)
    return pending


def grade_case(
    case: dict[str, Any],
    gold: dict[str, Any],
    run_dir: Path,
    require_judge: bool,
) -> dict[str, Any]:
    slug = common.case_slug(case)
    case_dir = run_dir / "cases" / slug
    checks = Checks()
    artifacts, parse_errors = common.read_artifacts(case_dir)
    checks.add(
        "required_artifacts_parse",
        not parse_errors,
        parse_errors or f"{len(common.REQUIRED_ARTIFACTS)} artifacts",
        severity="critical",
    )
    graph: dict[str, Any] = {}
    pending: list[dict[str, Any]] = []
    if not parse_errors:
        validate_envelopes(artifacts, checks)
        source_ids, _ = validate_source_manifest(artifacts, checks)
        graph = validate_graph(artifacts, source_ids, checks)
        validate_contamination(artifacts, graph, checks)
        validate_panel(artifacts, graph, checks)
        pending = apply_gold(case, gold, artifacts, graph, checks)

    verdict_path = run_dir / "judgments" / slug / "verdict.json"
    verdict = None
    if verdict_path.is_file():
        try:
            verdict = common.load_json(verdict_path)
        except Exception as exc:  # noqa: BLE001
            checks.add("judge_verdict_parse", False, str(exc))
    elif require_judge:
        checks.add("judge_verdict_present", False, str(verdict_path))

    return {
        "case_id": common.case_id(case),
        "case_name": common.case_name(case),
        "case_slug": slug,
        "split": case.get("split"),
        "tags": case.get("tags", []),
        "passed": checks.passed,
        "checks": checks.items,
        "pending_judgment_assertions": pending,
        "judge_verdict": verdict,
        "artifact_sha256": common.artifact_hashes(case_dir) if case_dir.is_dir() else {},
        "counts": {
            "sources": len(as_list(artifacts.get("source_manifest.json", {}).get("sources")))
            if isinstance(artifacts.get("source_manifest.json"), dict) else 0,
            "canonical_cells": len(graph.get("cells", [])),
            "candidates": len(graph.get("candidates", [])),
            "accepted_candidates": len(graph.get("accepted", [])),
        },
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("run_dir")
    parser.add_argument("--cases")
    parser.add_argument("--gold")
    parser.add_argument("--out")
    parser.add_argument("--ids", default="all")
    parser.add_argument("--split")
    parser.add_argument("--require-judge", action="store_true")
    args = parser.parse_args()

    script_dir = Path(__file__).resolve().parent
    eval_dir = script_dir.parent
    run_dir = Path(args.run_dir).resolve()
    cases_file = Path(args.cases).resolve() if args.cases else eval_dir / "cases.json"
    gold_file = Path(args.gold).resolve() if args.gold else eval_dir / "gold.json"
    out_path = Path(args.out).resolve() if args.out else run_dir / "grading.json"

    _, records = common.load_records(cases_file)
    selected = common.select_cases(records, ids=args.ids, split=args.split)
    _, gold_records = common.load_records(gold_file)

    results = []
    for case in selected:
        gold = common.find_record(gold_records, common.case_id(case))
        results.append(grade_case(case, gold, run_dir, args.require_judge))

    passed = sum(result["passed"] for result in results)
    failed_checks = collections.Counter(
        check["id"]
        for result in results
        for check in result["checks"]
        if not check["passed"] and check["severity"] in ("error", "critical")
    )
    output = {
        "schema_version": "1.0.0",
        "graded_at": common.utc_now(),
        "run": run_dir.name,
        "cases_sha256": common.sha256_file(cases_file),
        "gold_sha256": common.sha256_file(gold_file),
        "summary": {
            "cases": len(results),
            "passed": passed,
            "failed": len(results) - passed,
            "pass_rate": passed / len(results) if results else math.nan,
            "top_failed_checks": failed_checks.most_common(20),
        },
        "cases": results,
    }
    common.atomic_write_json(out_path, output)

    print("=== AI visibility panel deterministic grade ===")
    print(f"cases: {len(results)}  passed: {passed}  failed: {len(results) - passed}")
    for result in results:
        print(
            f"[{'PASS' if result['passed'] else 'FAIL'}] "
            f"{result['case_id']} {result['case_name']} "
            f"(cells={result['counts']['canonical_cells']}, "
            f"accepted={result['counts']['accepted_candidates']})"
        )
        for check in result["checks"]:
            if not check["passed"] and check["severity"] in ("error", "critical"):
                print(f"  - {check['id']}: {check['evidence']}")
    print(f"wrote {out_path}")
    return 0 if passed == len(results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
