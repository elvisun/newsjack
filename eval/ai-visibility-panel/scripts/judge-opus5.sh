#!/usr/bin/env bash
# Judge completed AI visibility panel artifacts with fresh, tools-off Opus 5 sessions.
#
# Usage:
#   judge-opus5.sh RUN_DIR [options]
#
# Options:
#   --cases FILE          Case dataset (default: eval/ai-visibility-panel/cases.json)
#   --gold FILE           Private gold/assertions (default: eval/.../gold.json)
#   --instructions FILE   Blind judge prompt (default: eval/.../harness/judge.md)
#   --schema FILE         Claude structured-output schema (default: harness/judge-schema.json)
#   --ids ID1,ID2         Judge only these case IDs/names (default: all)
#   --split NAME          Judge only this split
#   --concurrency N       Parallel fresh judge sessions (default: 2)
#   --effort LEVEL        Claude effort (default: high)
#   --max-budget-usd USD  Hard per-judge budget (default: 4)
#   --dry-run             Assemble prompts/manifests without invoking Claude

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMMON="$SCRIPT_DIR/_common.py"

MODEL_ID="claude-opus-5"
CASES_FILE="$EVAL_DIR/cases.json"
GOLD_FILE="$EVAL_DIR/gold.json"
JUDGE_INSTRUCTIONS="$EVAL_DIR/harness/judge.md"
JUDGE_SCHEMA="$EVAL_DIR/harness/judge-schema.json"
CASE_IDS="all"
SPLIT=""
CONCURRENCY=2
EFFORT="high"
MAX_BUDGET_USD="4"
DRY_RUN=0

usage() {
  sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
}

die() {
  echo "judge-opus5.sh: $*" >&2
  exit 2
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi
[ "$#" -ge 1 ] || { usage >&2; exit 2; }
RUN_DIR="$1"
shift

while [ "$#" -gt 0 ]; do
  case "$1" in
    --cases|--gold|--instructions|--schema|--ids|--split|--concurrency|--effort|--max-budget-usd)
      [ "$#" -ge 2 ] || die "$1 needs a value"
      option="$1"
      value="$2"
      case "$option" in
        --cases) CASES_FILE="$value" ;;
        --gold) GOLD_FILE="$value" ;;
        --instructions) JUDGE_INSTRUCTIONS="$value" ;;
        --schema) JUDGE_SCHEMA="$value" ;;
        --ids) CASE_IDS="$value" ;;
        --split) SPLIT="$value" ;;
        --concurrency) CONCURRENCY="$value" ;;
        --effort) EFFORT="$value" ;;
        --max-budget-usd) MAX_BUDGET_USD="$value" ;;
      esac
      shift 2
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

[[ "$CONCURRENCY" =~ ^[1-9][0-9]*$ ]] || die "concurrency must be a positive integer"
[[ "$MAX_BUDGET_USD" =~ ^[0-9]+([.][0-9]+)?$ ]] || die "max budget must be a non-negative number"
case "$EFFORT" in
  low|medium|high|xhigh|max) ;;
  *) die "invalid effort: $EFFORT" ;;
esac

[ -d "$RUN_DIR" ] || die "run directory does not exist: $RUN_DIR"
RUN_DIR="$(cd "$RUN_DIR" && pwd)"
for variable in CASES_FILE GOLD_FILE JUDGE_INSTRUCTIONS JUDGE_SCHEMA; do
  value="${!variable}"
  value="$(cd "$(dirname "$value")" && pwd)/$(basename "$value")"
  printf -v "$variable" '%s' "$value"
  [ -f "$value" ] || die "missing $variable: $value"
done

command -v python3 >/dev/null 2>&1 || die "python3 is required"
python3 - "$JUDGE_SCHEMA" <<'PY'
import json, sys
value = json.load(open(sys.argv[1]))
if not isinstance(value, dict) or value.get("type") != "object":
    raise SystemExit("judge schema must be a top-level JSON object schema")
PY

if [ "$DRY_RUN" -eq 0 ]; then
  CLAUDE_BIN="${CLAUDE_BIN:-claude}"
  command -v "$CLAUDE_BIN" >/dev/null 2>&1 || die "Claude Code not found: $CLAUDE_BIN"
else
  CLAUDE_BIN="${CLAUDE_BIN:-claude}"
fi

TASK_FILE="$(mktemp -t ai-panel-judge-tasks.XXXXXX)"
trap 'rm -f "$TASK_FILE"' EXIT

STAGE_ARGS=(
  stage-judgments
  --cases "$CASES_FILE"
  --gold "$GOLD_FILE"
  --run-dir "$RUN_DIR"
  --ids "$CASE_IDS"
  --judge-instructions "$JUDGE_INSTRUCTIONS"
  --tasks "$TASK_FILE"
)
if [ -n "$SPLIT" ]; then
  STAGE_ARGS+=(--split "$SPLIT")
fi
STAGE_OUTPUT="$(python3 "$COMMON" "${STAGE_ARGS[@]}")"
SELECTED_COUNT="$(
  python3 - "$STAGE_OUTPUT" <<'PY'
import json, sys
print(json.loads(sys.argv[1])["selected"])
PY
)"

JUDGE_APPARATUS_HASH="$(
  python3 - "$CASES_FILE" "$GOLD_FILE" "$JUDGE_INSTRUCTIONS" "$JUDGE_SCHEMA" <<'PY'
import hashlib, json, sys
items = {}
for raw in sys.argv[1:]:
    with open(raw, "rb") as handle:
        items[raw] = hashlib.sha256(handle.read()).hexdigest()
payload = json.dumps(items, sort_keys=True, separators=(",", ":")).encode()
print(hashlib.sha256(payload).hexdigest())
PY
)"

python3 - \
  "$RUN_DIR/JUDGE_MANIFEST.json" \
  "$CASES_FILE" \
  "$GOLD_FILE" \
  "$JUDGE_INSTRUCTIONS" \
  "$JUDGE_SCHEMA" \
  "$MODEL_ID" \
  "$JUDGE_APPARATUS_HASH" \
  "$CONCURRENCY" \
  "$EFFORT" \
  "$MAX_BUDGET_USD" \
  "$STAGE_OUTPUT" <<'PY'
import datetime as dt, hashlib, json, os, sys
from pathlib import Path

out, cases, gold, instructions, schema, model, apparatus, concurrency, effort, budget, staged = sys.argv[1:]
def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()
manifest = {
    "schema_version": "1.0.0",
    "kind": "ai_visibility_panel_blind_judge",
    "created_at": dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
    "model": model,
    "cases_sha256": digest(cases),
    "gold_sha256": digest(gold),
    "judge_instructions_sha256": digest(instructions),
    "judge_schema_sha256": digest(schema),
    "judge_apparatus_sha256": apparatus,
    "selected_case_ids": json.loads(staged)["case_ids"],
    "config": {
        "concurrency": int(concurrency),
        "effort": effort,
        "max_budget_usd_per_case": float(budget),
        "session_persistence": False,
        "safe_mode": True,
        "tools": [],
    },
}
path = Path(out)
if path.exists():
    old = json.loads(path.read_text())
    immutable = (
        "kind", "model", "cases_sha256", "gold_sha256",
        "judge_instructions_sha256", "judge_schema_sha256",
        "judge_apparatus_sha256", "selected_case_ids", "config",
    )
    changed = [key for key in immutable if old.get(key) != manifest.get(key)]
    if changed:
        raise SystemExit(f"{path}: immutable judge config differs ({', '.join(changed)}); use a new run directory")
else:
    part = path.with_name(path.name + ".part")
    part.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
    os.replace(part, path)
PY

echo ">>> staged $SELECTED_COUNT blind judgment(s) in $RUN_DIR/judgments" >&2
echo ">>> exact judge model: $MODEL_ID; tools off; concurrency: $CONCURRENCY" >&2

if [ "$DRY_RUN" -eq 1 ]; then
  echo ">>> dry run: prompts and manifest assembled; no Claude sessions launched" >&2
  exit 0
fi

JUDGE_SCHEMA_JSON="$(tr -d '\r\n' < "$JUDGE_SCHEMA")"

judge_case() {
  local judgment_dir="$1"
  local action
  local raw_part="$judgment_dir/judge.raw.json.part"
  local verdict_part="$judgment_dir/verdict.json.part"
  local log_part="$judgment_dir/judge.log.part"

  action="$(
    python3 "$COMMON" judgment-current \
      --judgment-dir "$judgment_dir" \
      --apparatus-hash "$JUDGE_APPARATUS_HASH"
  )"
  if [ "$action" = "skip" ]; then
    echo ">>> skip current judgment $(basename "$judgment_dir")" >&2
    return 0
  fi

  if [ -e "$judgment_dir/verdict.json" ] || [ -e "$judgment_dir/judge.raw.json" ] || [ -e "$judgment_dir/judge_run.json" ]; then
    local stamp archive
    stamp="$(date -u +%Y%m%dT%H%M%SZ)-$$"
    archive="$judgment_dir/_attempts/$stamp"
    mkdir -p "$archive"
    for name in verdict.json judge.raw.json judge.log judge_run.json; do
      [ ! -e "$judgment_dir/$name" ] || mv "$judgment_dir/$name" "$archive/$name"
    done
  fi
  rm -f "$raw_part" "$verdict_part" "$log_part"

  echo ">>> judge $(basename "$judgment_dir") via fresh $MODEL_ID" >&2
  (
    cd "$judgment_dir"
    "$CLAUDE_BIN" -p \
      --model "$MODEL_ID" \
      --effort "$EFFORT" \
      --max-budget-usd "$MAX_BUDGET_USD" \
      --safe-mode \
      --disable-slash-commands \
      --no-chrome \
      --no-session-persistence \
      --permission-mode dontAsk \
      --strict-mcp-config \
      --mcp-config '{"mcpServers":{}}' \
      --tools "" \
      --output-format json \
      --json-schema "$JUDGE_SCHEMA_JSON" \
      < "$judgment_dir/judge_prompt.txt" \
      > "$raw_part" \
      2> "$log_part"
  )

  python3 "$COMMON" extract-judge --raw "$raw_part" --out "$verdict_part"
  mv "$raw_part" "$judgment_dir/judge.raw.json"
  mv "$verdict_part" "$judgment_dir/verdict.json"
  mv "$log_part" "$judgment_dir/judge.log"
  python3 "$COMMON" complete-judge \
    --judgment-dir "$judgment_dir" \
    --apparatus-hash "$JUDGE_APPARATUS_HASH"
  echo ">>> complete judgment $(basename "$judgment_dir")" >&2
}

export -f judge_case
export COMMON JUDGE_APPARATUS_HASH CLAUDE_BIN MODEL_ID EFFORT MAX_BUDGET_USD
export JUDGE_SCHEMA_JSON

xargs -0 -n 1 -P "$CONCURRENCY" bash -c \
  'set -euo pipefail; judge_case "$1"' _ < "$TASK_FILE"

python3 - "$RUN_DIR/JUDGE_MANIFEST.json" "$RUN_DIR" "$SELECTED_COUNT" <<'PY'
import json, os, sys
from pathlib import Path
manifest_path = Path(sys.argv[1])
run_dir = Path(sys.argv[2])
expected = int(sys.argv[3])
complete = list((run_dir / "judgments").glob("*/judge_run.json"))
if len(complete) < expected:
    raise SystemExit(f"only {len(complete)} completed judgment(s), expected {expected}")
manifest = json.loads(manifest_path.read_text())
manifest["completed_case_count"] = len(complete)
part = manifest_path.with_name(manifest_path.name + ".part")
part.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
os.replace(part, manifest_path)
PY

echo ">>> blind judge run complete: $RUN_DIR" >&2
