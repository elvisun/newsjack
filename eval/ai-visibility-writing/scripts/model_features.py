#!/usr/bin/env python3
"""Dependency-free, query-fixed-effect linear probability models for page features."""

from __future__ import annotations

import math
from collections import Counter, defaultdict
from typing import Any


FACTORS = [f"F{i}" for i in range(1, 11)]
PREREGISTERED = ["F1", "F2", "F3", "F5", "F8", "F10"]


def inverse(matrix: list[list[float]]) -> list[list[float]]:
    size = len(matrix)
    augmented = [row[:] + [1.0 if i == j else 0.0 for j in range(size)] for i, row in enumerate(matrix)]
    for column in range(size):
        pivot = max(range(column, size), key=lambda row: abs(augmented[row][column]))
        if abs(augmented[pivot][column]) < 1e-12:
            raise ValueError("singular design")
        augmented[column], augmented[pivot] = augmented[pivot], augmented[column]
        scale = augmented[column][column]
        augmented[column] = [value / scale for value in augmented[column]]
        for row in range(size):
            if row == column:
                continue
            scale = augmented[row][column]
            if scale:
                augmented[row] = [left - scale * right for left, right in zip(augmented[row], augmented[column])]
    return [row[size:] for row in augmented]


def matvec(matrix: list[list[float]], vector: list[float]) -> list[float]:
    return [sum(left * right for left, right in zip(row, vector)) for row in matrix]


def bh_adjust(pvalues: dict[str, float]) -> dict[str, float]:
    ordered = sorted(pvalues.items(), key=lambda item: item[1])
    adjusted: dict[str, float] = {}
    running = 1.0
    total = len(ordered)
    for rank in range(total, 0, -1):
        key, value = ordered[rank - 1]
        running = min(running, value * total / rank)
        adjusted[key] = min(1.0, running)
    return adjusted


def _eligible(rows: list[dict[str, Any]], platform: str) -> list[dict[str, Any]]:
    candidates = [
        row for row in rows
        if row.get("platform") == platform
        and isinstance(row.get("features"), dict)
        and not row.get("exclusion_reason")
        and (
            platform != "google"
            or (row.get("risk_set") == "google_organic" and isinstance(row.get("organic_position"), (int, float)))
        )
    ]
    groups: dict[str, set[bool]] = defaultdict(set)
    for row in candidates:
        groups[str(row.get("request_tag"))].add(bool(row.get("cited")))
    mixed = {key for key, values in groups.items() if len(values) == 2}
    return [row for row in candidates if str(row.get("request_tag")) in mixed]


def _design(rows: list[dict[str, Any]], include_rank_domain: bool) -> tuple[list[list[float]], list[float], list[str], list[str]]:
    domain_counts = Counter(str(row.get("publisher_domain") or "unknown") for row in rows)
    top_domains = {name for name, _ in domain_counts.most_common(12)}
    documents = sorted({str(row["features"].get("document_type") or "unknown") for row in rows})
    domains = sorted(top_domains) if include_rank_domain else []
    names = FACTORS + ["log_age", "log_length"]
    if include_rank_domain:
        names += ["organic_position", "organic_position_missing"]
    names += [f"doc:{name}" for name in documents[1:]]
    names += [f"domain:{name}" for name in domains[1:]]
    raw: list[list[float]] = []
    groups = []
    y = []
    positions = [float(row.get("organic_position")) for row in rows if isinstance(row.get("organic_position"), (int, float))]
    position_fill = sorted(positions)[len(positions) // 2] if positions else 10.0
    for row in rows:
        features = row["features"]
        factor_values = features.get("factor_features") or {}
        age = features.get("publication_age_days")
        length = features.get("content_length")
        vector = [float(factor_values.get(name, 0.0)) for name in FACTORS]
        vector += [math.log1p(max(0.0, float(age))) if isinstance(age, (int, float)) else 0.0, math.log1p(max(0.0, float(length))) if isinstance(length, (int, float)) else 0.0]
        if include_rank_domain:
            position = row.get("organic_position")
            vector += [float(position) if isinstance(position, (int, float)) else position_fill, 0.0 if isinstance(position, (int, float)) else 1.0]
        document = str(features.get("document_type") or "unknown")
        vector += [1.0 if document == name else 0.0 for name in documents[1:]]
        domain = str(row.get("publisher_domain") or "unknown")
        vector += [1.0 if domain == name else 0.0 for name in domains[1:]]
        raw.append(vector)
        y.append(1.0 if row.get("cited") else 0.0)
        groups.append(str(row.get("request_tag")))
    means = [sum(row[index] for row in raw) / len(raw) for index in range(len(names))]
    scales = []
    for index, mean in enumerate(means):
        variance = sum((row[index] - mean) ** 2 for row in raw) / max(len(raw) - 1, 1)
        scales.append(math.sqrt(variance))
    keep = [index for index, scale in enumerate(scales) if scale > 1e-9]
    names = [names[index] for index in keep]
    standardized = [[(row[index] - means[index]) / scales[index] for index in keep] for row in raw]
    group_indices: dict[str, list[int]] = defaultdict(list)
    for index, group in enumerate(groups):
        group_indices[group].append(index)
    transformed_x = [row[:] for row in standardized]
    transformed_y = y[:]
    for indices in group_indices.values():
        y_mean = sum(y[index] for index in indices) / len(indices)
        x_means = [sum(standardized[index][column] for index in indices) / len(indices) for column in range(len(names))]
        for index in indices:
            transformed_y[index] -= y_mean
            transformed_x[index] = [value - x_means[column] for column, value in enumerate(standardized[index])]
    varying = [column for column in range(len(names)) if sum(row[column] ** 2 for row in transformed_x) > 1e-9]
    return [[row[column] for column in varying] for row in transformed_x], transformed_y, [names[column] for column in varying], groups


def fit(rows: list[dict[str, Any]], *, platform: str = "google", include_rank_domain: bool = True) -> dict[str, Any]:
    eligible = _eligible(rows, platform)
    groups = {row.get("request_tag") for row in eligible}
    if len(eligible) < 30 or len(groups) < 8:
        return {"status": "insufficient", "platform": platform, "n": len(eligible), "queries": len(groups), "effects": {}}
    x, y, names, clusters = _design(eligible, include_rank_domain)
    if not names:
        return {"status": "insufficient", "platform": platform, "n": len(eligible), "queries": len(groups), "effects": {}}
    p = len(names)
    xtx = [[sum(row[i] * row[j] for row in x) + (1e-7 if i == j else 0.0) for j in range(p)] for i in range(p)]
    xty = [sum(row[i] * target for row, target in zip(x, y)) for i in range(p)]
    bread = inverse(xtx)
    beta = matvec(bread, xty)
    residuals = [target - sum(value * coefficient for value, coefficient in zip(row, beta)) for row, target in zip(x, y)]
    scores: dict[str, list[float]] = defaultdict(lambda: [0.0] * p)
    for row, residual, cluster in zip(x, residuals, clusters):
        for index, value in enumerate(row):
            scores[cluster][index] += value * residual
    correction = (len(scores) / max(len(scores) - 1, 1)) * ((len(x) - 1) / max(len(x) - p, 1))
    standard_errors = []
    for index in range(p):
        projection = [sum(bread[index][column] * score[column] for column in range(p)) for score in scores.values()]
        standard_errors.append(math.sqrt(max(0.0, correction * sum(value * value for value in projection))))
    effects = {}
    pvalues = {}
    for name, coefficient, standard_error in zip(names, beta, standard_errors):
        if name not in FACTORS:
            continue
        pvalue = math.erfc(abs(coefficient / standard_error) / math.sqrt(2)) if standard_error > 0 else 1.0
        effects[name] = {
            "estimate_pp_per_sd": 100 * coefficient,
            "standard_error_pp": 100 * standard_error,
            "interval_95_pp": [100 * (coefficient - 1.96 * standard_error), 100 * (coefficient + 1.96 * standard_error)],
            "p_value": pvalue,
        }
        if name in PREREGISTERED:
            pvalues[name] = pvalue
    adjusted = bh_adjust(pvalues)
    for name, value in adjusted.items():
        effects[name]["fdr_q_value"] = value
    return {
        "status": "estimated", "platform": platform, "n": len(eligible), "queries": len(groups),
        "method": "query-fixed-effect linear probability model; cluster-robust query standard errors; numeric predictors standardized; top-12 domain and document-type controls",
        "include_rank_domain": include_rank_domain, "effects": effects,
    }


def analyze(rows: list[dict[str, Any]]) -> dict[str, Any]:
    google = fit(rows, platform="google", include_rank_domain=True)
    ablation = fit(rows, platform="google", include_rank_domain=False)
    chatgpt_positive = sum(row.get("platform") == "chatgpt" and row.get("cited") for row in rows)
    interactions: dict[str, dict[str, Any]] = {}
    for field in ("topic_family", "intent"):
        interactions[field] = {}
        for level in sorted({str(row.get(field)) for row in rows if row.get(field)}):
            model = fit([row for row in rows if str(row.get(field)) == level], platform="google", include_rank_domain=True)
            interactions[field][level] = {"status": model["status"], "n": model["n"], "effects": model.get("effects", {})}
    documents = sorted({str((row.get("features") or {}).get("document_type")) for row in rows if isinstance(row.get("features"), dict)})
    interactions["document_type"] = {}
    for level in documents:
        model = fit([row for row in rows if str((row.get("features") or {}).get("document_type")) == level], platform="google", include_rank_domain=True)
        interactions["document_type"][level] = {"status": model["status"], "n": model["n"], "effects": model.get("effects", {})}
    sensitivity_rows = [row for row in rows if not any(row.get(flag) for flag in ("syndicated", "duplicated", "ambiguous", "inaccessible"))]
    sensitivity = fit(sensitivity_rows, platform="google", include_rank_domain=True)
    return {
        "google_adjusted": google,
        "chatgpt_adjusted": {"status": "unestimable", "positive_pages": chatgpt_positive, "reason": "ChatGPT does not expose a retrieved-but-rejected risk set; Google organic controls are not relabeled as ChatGPT retrievals."},
        "rank_domain_ablation": ablation,
        "interactions": interactions,
        "sensitivity_exclusions": sensitivity,
    }
