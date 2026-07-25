#!/usr/bin/env python3
"""Compute the frozen v1 pilot composite from schema-complete metrics."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))
from pilotlib import PilotError, composite  # noqa: E402


REQUIRED_GATES = [
    "zero_invented_facts", "schema_complete", "no_forbidden_claim",
    "no_clear_stratum_regression", "stage0_green", "ledger_reconciled",
]


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--fresh", action="store_true")
    args = parser.parse_args()
    try:
        data = json.loads(args.input.read_text(encoding="utf-8"))
        if args.fresh and data.get("mode") != "fresh":
            raise PilotError("fresh score requires fresh metrics")
        gates = data.get("hard_gates") or {}
        if set(gates) != set(REQUIRED_GATES) or not all(gates.values()):
            print("VOID: constraint violation")
            return 3
        value = composite(data.get("components") or {})
        for key, component in data["components"].items():
            print(f"{key}={float(component):.4f}", file=sys.stderr)
        print(f"composite={value:.2f}", file=sys.stderr)
        print(f"{value:.2f}")
        return 0
    except (OSError, json.JSONDecodeError, PilotError, TypeError, ValueError):
        print("VOID: constraint violation")
        return 3


if __name__ == "__main__":
    raise SystemExit(main())
