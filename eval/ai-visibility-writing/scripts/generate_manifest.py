#!/usr/bin/env python3
"""Generate reproducible balanced query manifests from frozen protocol strata."""

from __future__ import annotations

import argparse
import hashlib
import json
import random
import secrets
from datetime import date, datetime, timezone
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def _rows(protocol: dict, *, phase: str, seed_material: str, query_date: date) -> list[dict]:
    sampling = protocol["sampling"]
    seed = int(hashlib.sha256(seed_material.encode()).hexdigest(), 16)
    rng = random.Random(seed)
    rows = []
    for family, topics in sampling["topic_families"].items():
        shuffled = list(topics)
        rng.shuffle(shuffled)
        for intent_index, (intent, template) in enumerate(sampling["intents"].items()):
            rotated = shuffled[intent_index:] + shuffled[:intent_index]
            for topic in rotated:
                query = template.format(topic=topic, year=query_date.year, month_year=query_date.strftime("%B %Y"))
                rows.append({
                    "phase": phase,
                    "topic_family": family,
                    "intent": intent,
                    "topic": topic,
                    "query": query,
                })
    rng.shuffle(rows)
    return rows


def _balanced_subset(rows: list[dict], count: int) -> list[dict]:
    buckets: dict[tuple[str, str], list[dict]] = {}
    for row in rows:
        buckets.setdefault((row["topic_family"], row["intent"]), []).append(row)
    selected = []
    cell_counts = {key: 0 for key in buckets}
    family_counts = {key[0]: 0 for key in buckets}
    intent_counts = {key[1]: 0 for key in buckets}
    while len(selected) < count and any(buckets.values()):
        available = [key for key in buckets if buckets[key]]
        key = min(available, key=lambda item: (
            cell_counts[item], family_counts[item[0]], intent_counts[item[1]], item,
        ))
        selected.append(buckets[key].pop())
        cell_counts[key] += 1
        family_counts[key[0]] += 1
        intent_counts[key[1]] += 1
    return selected


def build(phase: str, seed_material: str | None, count: int | None) -> dict:
    protocol = json.loads((ROOT / "protocol.json").read_text(encoding="utf-8"))
    sampling = protocol["sampling"]
    generated = datetime.now(timezone.utc)
    if phase == "fresh":
        seed_material = seed_material or secrets.token_hex(32)
        query_date = generated.date()
    elif phase == "calibration":
        seed_material = seed_material or sampling["main_seed_material"] + "|calibration"
        query_date = date.fromisoformat(sampling["main_query_as_of"])
    elif phase == "repeats":
        seed_material = seed_material or sampling["main_seed_material"] + "|repeats"
        query_date = date.fromisoformat(sampling["main_query_as_of"])
    else:
        seed_material = seed_material or sampling["main_seed_material"]
        query_date = date.fromisoformat(sampling["main_query_as_of"])
    all_main_rows = _rows(protocol, phase="main", seed_material=sampling["main_seed_material"], query_date=query_date)
    main_families = set(sampling.get("main_topic_families", sampling["topic_families"]))
    main_rows = _balanced_subset(
        [row for row in all_main_rows if row["topic_family"] in main_families],
        sampling["main_pairs_target"],
    )
    for index, row in enumerate(main_rows, 1):
        row["paired_unit_id"] = f"main-{index:04d}"
    if phase == "main":
        rows = main_rows
    elif phase == "calibration":
        calibration_rng = random.Random(int(hashlib.sha256(seed_material.encode()).hexdigest(), 16))
        by_family: dict[str, list[dict]] = {}
        for row in all_main_rows:
            by_family.setdefault(row["topic_family"], []).append(row)
        rows = []
        intents = list(sampling["intents"])
        for index, family in enumerate(sorted(by_family)):
            choices = [row for row in by_family[family] if row["intent"] == intents[index % len(intents)]]
            rows.append(calibration_rng.choice(choices))
        rows = [{**row, "phase": "calibration"} for row in rows]
    elif phase == "repeats":
        repeat_seed = seed_material or sampling["main_seed_material"] + "|repeats"
        repeat_rng = random.Random(int(hashlib.sha256(repeat_seed.encode()).hexdigest(), 16))
        shuffled = list(main_rows)
        repeat_rng.shuffle(shuffled)
        rows = _balanced_subset(shuffled, count or 60)
        rows = [{**row, "repeat_of": row["paired_unit_id"], "phase": "repeats"} for row in rows]
        seed_material = repeat_seed
    else:
        rows = _balanced_subset(_rows(protocol, phase="fresh", seed_material=seed_material, query_date=query_date), count or 60)
    if count is not None and phase == "main" and count < len(rows):
        rows = _balanced_subset(rows, count)
    for index, row in enumerate(rows, 1):
        row["paired_unit_id"] = f"{phase}-{index:04d}"
        row["requests"] = [
            {"platform": "google", "request_tag": f"{phase}-{index:04d}-google"},
            {"platform": "chatgpt", "request_tag": f"{phase}-{index:04d}-chatgpt"},
        ]
    return {
        "protocol_id": protocol["protocol_id"],
        "phase": phase,
        "generated_at": generated.isoformat().replace("+00:00", "Z"),
        "query_as_of": query_date.isoformat(),
        "seed_material": seed_material,
        "seed_sha256": hashlib.sha256(seed_material.encode()).hexdigest(),
        "pair_count": len(rows),
        "units": rows,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--phase", choices=["calibration", "main", "repeats", "fresh"], required=True)
    parser.add_argument("--seed-material")
    parser.add_argument("--count", type=int)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.phase == "calibration" and args.count not in (None, 6):
        raise SystemExit("calibration manifest must contain 6 pairs")
    protocol = json.loads((ROOT / "protocol.json").read_text(encoding="utf-8"))
    target = protocol["sampling"].get(f"{args.phase[:-1] if args.phase == 'repeats' else args.phase}_pairs_target")
    if args.phase != "calibration" and args.count not in (None, target):
        raise SystemExit(f"{args.phase} manifest must contain {target} pairs")
    default_count = 6 if args.phase == "calibration" else target
    data = build(args.phase, args.seed_material, args.count or default_count)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
