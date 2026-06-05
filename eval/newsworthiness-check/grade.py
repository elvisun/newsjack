#!/usr/bin/env python3
"""Deterministic grader for the newsworthiness-check eval.

Reads evals.json + a directory of skill outputs (one <id>-<eval_name>.json per
case) and evaluates each assertion's machine-checkable `check`. Writes
grading.json in the Anthropic {text, passed, evidence} shape and prints a
summary. Reusable across model runs and skill versions: same assertions, any
outputs dir.

Usage:
  python3 grade.py <outputs_dir> [--evals evals.json] [--out grading.json]
"""
import argparse, json, os, sys


def get(obj, path):
    """Dot-path lookup. Returns (found, value)."""
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict) and part in cur:
            cur = cur[part]
        else:
            return False, None
    return True, cur


def as_text(v):
    """Flatten a string or list-of-strings/dicts to one lowercased blob."""
    if v is None:
        return ""
    if isinstance(v, str):
        return v.lower()
    if isinstance(v, list):
        return " \n ".join(as_text(x) for x in v).lower()
    if isinstance(v, dict):
        return " \n ".join(as_text(x) for x in v.values()).lower()
    return str(v).lower()


def caps_iter(doc):
    caps = doc.get("caps_applied")
    return caps if isinstance(caps, list) else []


def eval_check(doc, chk):
    """Return (passed: bool, evidence: str)."""
    op = chk.get("op")

    if op in ("all", "any"):
        subs = [eval_check(doc, c) for c in chk["checks"]]
        passed = all(p for p, _ in subs) if op == "all" else any(p for p, _ in subs)
        ev = " | ".join(e for _, e in subs)
        return passed, ev

    if op == "cap_applied":
        sub = chk["value"].lower()
        for c in caps_iter(doc):
            name = str(c.get("cap", "")).lower()
            if sub in name and c.get("applied") is True:
                return True, f"cap '{c.get('cap')}' applied=true"
        return False, f"no applied cap matching '{chk['value']}' (caps: {[c.get('cap') for c in caps_iter(doc)]})"

    if op == "cap_not_applied":
        sub = chk["value"].lower()
        for c in caps_iter(doc):
            name = str(c.get("cap", "")).lower()
            if sub in name and c.get("applied") is True:
                return False, f"unexpected applied cap '{c.get('cap')}'"
        return True, f"no applied cap matching '{chk['value']}'"

    found, val = get(doc, chk["path"])
    path = chk["path"]

    if op == "is_null":
        return (found and val is None), f"{path}={val!r}"
    if op == "not_null":
        return (found and val is not None), f"{path}={val!r}"

    if op == "contains_any":
        blob = as_text(val)
        hits = [w for w in chk["value"] if w.lower() in blob]
        return (len(hits) > 0), (f"{path} matched {hits}" if hits else f"{path} matched none of {chk['value']}")
    if op == "contains_all":
        blob = as_text(val)
        miss = [w for w in chk["value"] if w.lower() not in blob]
        return (len(miss) == 0), (f"{path} missing {miss}" if miss else f"{path} contains all")

    if not found or val is None:
        return False, f"{path} missing/null (val={val!r})"

    if op == "==":
        return val == chk["value"], f"{path}={val!r} (want =={chk['value']!r})"
    if op == "in":
        return val in chk["value"], f"{path}={val!r} (want in {chk['value']})"
    if op == "between":
        lo, hi = chk["value"]
        return (lo <= val <= hi), f"{path}={val!r} (want {lo}..{hi})"

    try:
        num = float(val)
    except (TypeError, ValueError):
        return False, f"{path}={val!r} not numeric for op {op}"
    cmp = {"<=": num <= chk["value"], ">=": num >= chk["value"],
           "<": num < chk["value"], ">": num > chk["value"]}.get(op)
    if cmp is None:
        return False, f"unknown op {op}"
    return cmp, f"{path}={val!r} (want {op}{chk['value']})"


def load_output(outputs_dir, case):
    fn = f"{case['id']}-{case['eval_name']}.json"
    p = os.path.join(outputs_dir, fn)
    if not os.path.exists(p):
        return None, f"missing output file {fn}"
    raw = open(p).read()
    try:
        return json.loads(raw), None
    except json.JSONDecodeError as e:
        return None, f"invalid JSON in {fn}: {e}"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("outputs_dir")
    here = os.path.dirname(os.path.abspath(__file__))
    ap.add_argument("--evals", default=os.path.join(here, "evals.json"))
    ap.add_argument("--out", default=None)
    args = ap.parse_args()

    evals = json.load(open(args.evals))
    out_path = args.out or os.path.join(args.outputs_dir, os.pardir, "grading.json")

    cases_out, n_assert, n_pass, n_case_pass = [], 0, 0, 0
    json_invalid = []
    for case in evals["evals"]:
        doc, err = load_output(args.outputs_dir, case)
        exps = []
        if err:
            json_invalid.append(case["id"])
            for a in case["assertions"]:
                exps.append({"text": a["text"], "passed": False, "evidence": err})
        else:
            for a in case["assertions"]:
                p, ev = eval_check(doc, a["check"])
                exps.append({"text": a["text"], "passed": bool(p), "evidence": ev})
        cpass = all(e["passed"] for e in exps)
        n_case_pass += int(cpass)
        n_assert += len(exps)
        n_pass += sum(e["passed"] for e in exps)
        cases_out.append({
            "id": case["id"], "eval_name": case["eval_name"],
            "industry": case.get("industry"), "mode": case.get("mode"),
            "overall_score": (doc or {}).get("overall_score") if doc else None,
            "recommended_action": (doc or {}).get("recommended_action") if doc else None,
            "json_valid": err is None, "passed": cpass, "expectations": exps,
        })

    result = {
        "skill_name": evals["skill_name"],
        "outputs_dir": os.path.abspath(args.outputs_dir),
        "totals": {
            "cases": len(cases_out), "cases_passed": n_case_pass,
            "assertions": n_assert, "assertions_passed": n_pass,
            "case_pass_rate": round(n_case_pass / len(cases_out), 4) if cases_out else 0,
            "assertion_pass_rate": round(n_pass / n_assert, 4) if n_assert else 0,
            "json_invalid_cases": json_invalid,
        },
        "cases": cases_out,
    }
    out_path = os.path.abspath(out_path)
    json.dump(result, open(out_path, "w"), indent=2)

    print(f"\n=== {evals['skill_name']} :: {args.outputs_dir} ===")
    print(f"cases {n_case_pass}/{len(cases_out)} | assertions {n_pass}/{n_assert} "
          f"| json-invalid: {json_invalid or 'none'}")
    for c in cases_out:
        flag = "PASS" if c["passed"] else ("BAD-JSON" if not c["json_valid"] else "FAIL")
        print(f"  [{flag}] #{c['id']:>2} {c['industry'] or '':<18} "
              f"score={c['overall_score']} action={c['recommended_action']} "
              f"({c['eval_name']})")
        if not c["passed"]:
            for e in c["expectations"]:
                if not e["passed"]:
                    print(f"        - FAILED: {e['text']}  [{e['evidence']}]")
    print(f"\nwrote {out_path}")
    return 0 if n_case_pass == len(cases_out) else 1


if __name__ == "__main__":
    sys.exit(main())
