#!/usr/bin/env bash
# Run the URL-to-panel workflow in one fresh Claude Code Opus 5 session per case.
#
# The runner deliberately addresses Opus 5 by exact model ID, disables session
# persistence/customizations/MCP/Chrome/slash skills, and grants only the tools
# the research workflow needs. It reads the repository skills by exact path so
# a stale installed ~/.claude skill cannot contaminate the eval.
#
# Usage:
#   run-opus5.sh RUN_DIR [options]
#
# Options:
#   --cases FILE          Case dataset (default: eval/ai-visibility-panel/cases.json)
#   --ids ID1,ID2         Run only these case IDs/names (default: all)
#   --split NAME          Run only this split
#   --concurrency N       Parallel Claude sessions (default: 2)
#   --effort LEVEL        Claude effort: low|medium|high|xhigh|max (default: high)
#   --max-budget-usd USD  Hard per-session API budget (default: 8)
#   --dry-run             Stage inputs and manifest without invoking Claude
#
# A completed case is resumed only when its input, exact skill hashes, required
# artifacts, and recorded artifact hashes all still match.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$EVAL_DIR/../.." && pwd)"
COMMON="$SCRIPT_DIR/_common.py"
EXECUTOR_INSTRUCTIONS="$EVAL_DIR/harness/executor.md"

MODEL_ID="claude-opus-5"
CASES_FILE="$EVAL_DIR/cases.json"
CASE_IDS="all"
SPLIT=""
CONCURRENCY=2
EFFORT="high"
MAX_BUDGET_USD="8"
DRY_RUN=0

usage() {
  sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
}

die() {
  echo "run-opus5.sh: $*" >&2
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
    --cases)
      [ "$#" -ge 2 ] || die "--cases needs a value"
      CASES_FILE="$2"
      shift 2
      ;;
    --ids)
      [ "$#" -ge 2 ] || die "--ids needs a value"
      CASE_IDS="$2"
      shift 2
      ;;
    --split)
      [ "$#" -ge 2 ] || die "--split needs a value"
      SPLIT="$2"
      shift 2
      ;;
    --concurrency)
      [ "$#" -ge 2 ] || die "--concurrency needs a value"
      CONCURRENCY="$2"
      shift 2
      ;;
    --effort)
      [ "$#" -ge 2 ] || die "--effort needs a value"
      EFFORT="$2"
      shift 2
      ;;
    --max-budget-usd)
      [ "$#" -ge 2 ] || die "--max-budget-usd needs a value"
      MAX_BUDGET_USD="$2"
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

CASES_FILE="$(cd "$(dirname "$CASES_FILE")" && pwd)/$(basename "$CASES_FILE")"
mkdir -p "$RUN_DIR"
RUN_DIR="$(cd "$RUN_DIR" && pwd)"

[ -f "$CASES_FILE" ] || die "missing cases file: $CASES_FILE"
[ -f "$COMMON" ] || die "missing helper: $COMMON"
[ -f "$EXECUTOR_INSTRUCTIONS" ] || die "missing executor harness: $EXECUTOR_INSTRUCTIONS"
command -v python3 >/dev/null 2>&1 || die "python3 is required"

APPARATUS=(
  "$REPO_ROOT/skills/build-ai-visibility-panel/SKILL.md"
  "$REPO_ROOT/skills/build-ai-visibility-panel/references/artifact-contracts.md"
  "$REPO_ROOT/skills/icp-evidence-analysis/SKILL.md"
  "$REPO_ROOT/skills/buyer-job-intent-analysis/SKILL.md"
  "$REPO_ROOT/skills/prompt-proximity-architecture/SKILL.md"
  "$REPO_ROOT/skills/realistic-prompt-generation/SKILL.md"
  "$REPO_ROOT/skills/prompt-set-qa/SKILL.md"
  "$REPO_ROOT/skills/ai-visibility-panel-design/SKILL.md"
  "$REPO_ROOT/skills/ETHICS.md"
)
for path in "${APPARATUS[@]}"; do
  [ -f "$path" ] || die "missing apparatus file: $path"
done

if [ "$DRY_RUN" -eq 0 ]; then
  CLAUDE_BIN="${CLAUDE_BIN:-claude}"
  command -v "$CLAUDE_BIN" >/dev/null 2>&1 || die "Claude Code not found: $CLAUDE_BIN"
else
  CLAUDE_BIN="${CLAUDE_BIN:-claude}"
fi

TASK_FILE="$(mktemp -t ai-panel-executor-tasks.XXXXXX)"
trap 'rm -f "$TASK_FILE"' EXIT

APPARATUS_ARGS=()
for path in "${APPARATUS[@]}"; do
  APPARATUS_ARGS+=(--apparatus "$path")
done

CONFIG_JSON="$(
  python3 - "$CONCURRENCY" "$EFFORT" "$MAX_BUDGET_USD" "$EXECUTOR_INSTRUCTIONS" <<'PY'
import hashlib, json, sys
print(json.dumps({
    "concurrency": int(sys.argv[1]),
    "effort": sys.argv[2],
    "max_budget_usd_per_case": float(sys.argv[3]),
    "executor_instructions_sha256": hashlib.sha256(open(sys.argv[4], "rb").read()).hexdigest(),
    "session_persistence": False,
    "safe_mode": True,
    "slash_commands": False,
    "mcp_servers": [],
    "chrome": False,
    "tools": ["Read", "Glob", "Grep", "WebSearch", "WebFetch", "Write", "Edit"],
}))
PY
)"

STAGE_ARGS=(
  stage
  --cases "$CASES_FILE"
  --run-dir "$RUN_DIR"
  --ids "$CASE_IDS"
  --tasks "$TASK_FILE"
  "${APPARATUS_ARGS[@]}"
  --config "$CONFIG_JSON"
)
if [ -n "$SPLIT" ]; then
  STAGE_ARGS+=(--split "$SPLIT")
fi
STAGE_OUTPUT="$(python3 "$COMMON" "${STAGE_ARGS[@]}")"
APPARATUS_HASH="$(
  python3 - "$STAGE_OUTPUT" <<'PY'
import json, sys
print(json.loads(sys.argv[1])["apparatus_combined_sha256"])
PY
)"
SELECTED_COUNT="$(
  python3 - "$STAGE_OUTPUT" <<'PY'
import json, sys
print(json.loads(sys.argv[1])["selected"])
PY
)"

echo ">>> staged $SELECTED_COUNT case(s) in $RUN_DIR" >&2
echo ">>> exact model: $MODEL_ID; concurrency: $CONCURRENCY; effort: $EFFORT" >&2

if [ "$DRY_RUN" -eq 1 ]; then
  echo ">>> dry run: manifest and inputs staged; no Claude sessions launched" >&2
  exit 0
fi

run_case() {
  local case_dir="$1"
  local action
  local claude_exit=0
  local terminal_reason
  local stream_part="$case_dir/executor.stream.jsonl.part"
  local log_part="$case_dir/executor.log.part"
  local case_apparatus_args=(
    --apparatus "$REPO_ROOT/skills/build-ai-visibility-panel/SKILL.md"
    --apparatus "$REPO_ROOT/skills/build-ai-visibility-panel/references/artifact-contracts.md"
    --apparatus "$REPO_ROOT/skills/icp-evidence-analysis/SKILL.md"
    --apparatus "$REPO_ROOT/skills/buyer-job-intent-analysis/SKILL.md"
    --apparatus "$REPO_ROOT/skills/prompt-proximity-architecture/SKILL.md"
    --apparatus "$REPO_ROOT/skills/realistic-prompt-generation/SKILL.md"
    --apparatus "$REPO_ROOT/skills/prompt-set-qa/SKILL.md"
    --apparatus "$REPO_ROOT/skills/ai-visibility-panel-design/SKILL.md"
    --apparatus "$REPO_ROOT/skills/ETHICS.md"
  )

  action="$(
    python3 "$COMMON" executor-prompt \
      --case-dir "$case_dir" \
      --apparatus-hash "$APPARATUS_HASH" \
      --instructions "$EXECUTOR_INSTRUCTIONS" \
      "${case_apparatus_args[@]}"
  )"
  if [ "$action" = "skip" ]; then
    echo ">>> skip current case $(basename "$case_dir")" >&2
    return 0
  fi

  rm -f "$stream_part" "$log_part"
  echo ">>> run $(basename "$case_dir") via $MODEL_ID" >&2

  (
    cd "$case_dir"
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
      --tools "Read,Glob,Grep,WebSearch,WebFetch,Write,Edit" \
      --allowed-tools "Read,Glob,Grep,WebSearch,WebFetch,Write,Edit" \
      --disallowed-tools "Bash,MultiEdit,NotebookEdit,Task,Skill,TodoWrite" \
      --add-dir "$REPO_ROOT/skills" \
      --output-format stream-json \
      --verbose \
      < "$case_dir/executor_prompt.txt" \
      > "$stream_part" \
      2> "$log_part"
  ) || claude_exit=$?

  python3 "$COMMON" validate-candidate --case-dir "$case_dir" >/dev/null
  if [ "$claude_exit" -ne 0 ]; then
    terminal_reason="$(
      python3 - "$stream_part" <<'PY'
import json, sys
terminal = {}
with open(sys.argv[1], encoding="utf-8") as handle:
    for line in handle:
        if line.strip():
            value = json.loads(line)
            if isinstance(value, dict):
                terminal = value
print(terminal.get("terminal_reason", ""))
PY
    )"
    if [ "$terminal_reason" != "budget_exhausted" ]; then
      echo ">>> Claude exited $claude_exit ($terminal_reason); refusing incomplete case" >&2
      return "$claude_exit"
    fi
    echo ">>> accept validated artifacts after Claude budget exhaustion" >&2
  fi
  mv "$stream_part" "$case_dir/executor.stream.jsonl"
  mv "$log_part" "$case_dir/executor.log"
  python3 "$COMMON" complete-executor \
    --case-dir "$case_dir" \
    --apparatus-hash "$APPARATUS_HASH" \
    --claude-exit-code "$claude_exit"
  echo ">>> complete $(basename "$case_dir")" >&2
}

export -f run_case
export COMMON APPARATUS_HASH CLAUDE_BIN MODEL_ID EFFORT MAX_BUDGET_USD REPO_ROOT
export EXECUTOR_INSTRUCTIONS

# xargs receives one NUL-terminated absolute case directory per task. Each child
# re-enables strict mode so any failed Claude call or validation fails the run.
xargs -0 -n 1 -P "$CONCURRENCY" bash -c \
  'set -euo pipefail; run_case "$1"' _ < "$TASK_FILE"

python3 - "$RUN_DIR" "$SELECTED_COUNT" <<'PY'
import json, sys
from pathlib import Path
run_dir = Path(sys.argv[1])
expected = int(sys.argv[2])
complete = list((run_dir / "cases").glob("*/executor_run.json"))
if len(complete) < expected:
    raise SystemExit(f"only {len(complete)} completed case(s), expected {expected}")
manifest_path = run_dir / "MANIFEST.json"
manifest = json.loads(manifest_path.read_text())
manifest["completed_case_count"] = len(complete)
manifest_path.with_suffix(".json.part").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
manifest_path.with_suffix(".json.part").replace(manifest_path)
PY

echo ">>> executor run complete: $RUN_DIR" >&2
