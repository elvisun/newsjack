#!/usr/bin/env python3
"""Rebuild results.json from per-brand verdict files already on disk.

Crash insurance: the workflow writes each brand's verdict-*.json the moment that
brand is judged, so a run that dies partway still leaves every completed brand's
verdicts on disk. This script scans a run directory and reconstructs the same
results.json shape that scripts/run.js returns and aggregate.py consumes — from
whatever exists. Run it anytime to checkpoint partial progress.

Usage:
    python3 collect.py runs/<run>            # writes runs/<run>/results.json
    python3 collect.py runs/<run> --stdout   # print, don't write
"""
import json
import os
import re
import sys

# verdict-ord1-Aopus-Bfable.json  ->  ordering, A_model, B_model
VERDICT_RE = re.compile(r"^verdict-(ord\d+)-A([a-z0-9]+)-B([a-z0-9]+)\.json$")
BRAND_DIR_RE = re.compile(r"^brand-(\d+)-(.+)$")


def main(run_dir, write=True):
    if not os.path.isdir(run_dir):
        sys.exit(f"not a directory: {run_dir}")

    records = []
    brand_dirs = sorted(
        d for d in os.listdir(run_dir)
        if BRAND_DIR_RE.match(d) and os.path.isdir(os.path.join(run_dir, d))
    )
    n_brands = 0
    for bd in brand_dirs:
        m = BRAND_DIR_RE.match(bd)
        brand_id = int(m.group(1))
        slug = m.group(2)
        path = os.path.join(run_dir, bd)
        found_here = 0
        for fn in sorted(os.listdir(path)):
            vm = VERDICT_RE.match(fn)
            if not vm:
                continue
            ordering, a_model, b_model = vm.group(1), vm.group(2), vm.group(3)
            fp = os.path.join(path, fn)
            try:
                with open(fp) as f:
                    verdict = json.load(f)
            except (json.JSONDecodeError, OSError) as e:
                print(f"  WARN: skipping unreadable {fn}: {e}", file=sys.stderr)
                continue
            # sanity: must have a winner + scores to be usable downstream
            if "winner" not in verdict or "scores" not in verdict:
                print(f"  WARN: {fn} missing winner/scores, skipping", file=sys.stderr)
                continue
            records.append({
                "brand_id": brand_id,
                "brand": slug.replace("-", " ").title(),
                "ordering": ordering,
                "A_model": a_model,
                "B_model": b_model,
                "verdict_file": f"{bd}/{fn}",
                "verdict": verdict,
            })
            found_here += 1
        if found_here:
            n_brands += 1
        status = {0: "MISSING", 1: "partial(1/2)"}.get(found_here, "ok")
        print(f"  brand-{brand_id:02d}-{slug}: {found_here} verdict(s) [{status}]",
              file=sys.stderr)

    out = {
        "run": os.path.basename(os.path.normpath(run_dir)),
        "study": "Fable 5 vs Opus 4.8 — story-angle quality, GPT-5.5 (meanest-editor) blind judge",
        "judge_model": "gpt-5.5",
        "generator_skill": "skills/angle-generator",
        "judge_skill": "skills/meanest-editor",
        "note": "Reconstructed from on-disk verdict files by collect.py. One record per judgment; ordering encodes which model sat in slot A vs B.",
        "records": records,
    }

    print(f"\nReconstructed {len(records)} judgment(s) across {n_brands} brand(s) "
          f"with at least one verdict.", file=sys.stderr)

    if write:
        out_path = os.path.join(run_dir, "results.json")
        with open(out_path, "w") as f:
            json.dump(out, f, indent=2)
        print(f"Wrote {out_path}", file=sys.stderr)
    else:
        print(json.dumps(out, indent=2))


if __name__ == "__main__":
    argv = [a for a in sys.argv[1:] if a != "--stdout"]
    if len(argv) != 1:
        print(__doc__)
        sys.exit(1)
    main(argv[0], write=("--stdout" not in sys.argv))
