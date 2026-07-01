#!/usr/bin/env bash
# Round-robin driver for the 4-model angle study:
#   Opus 4.8 (opus) · Fable 5 (fable) · Sonnet 5 (s5) · Sonnet 4.6 (s46)
#
# opus.md and fable.md are REUSED from a prior 2-model run (Fable is disabled and
# can't be regenerated; reusing keeps the already-judged opus<->fable pair valid).
# Only the two Sonnet sets are generated fresh, then every not-yet-judged pair is
# judged blind in both orderings by codex GPT-5.5. Fully resumable: re-run any
# phase and completed work is skipped.
#
# Usage:
#   run-4model.sh RUNDIR PHASE [CONCURRENCY]
#     PHASE = stage | gen | judge | all
#     RUNDIR  e.g. eval/fable-vs-opus/runs/2026-06-30-4model
#
# Requires SRCRUN (the run to copy reused artifacts from) — defaults to the
# 2026-06-09-full run in the same runs/ dir.
set -uo pipefail

RUNDIR="${1:?need RUNDIR}"
PHASE="${2:?need PHASE: stage|gen|judge|all}"
CONC="${3:-4}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$EVAL_DIR/../.." && pwd)"
cd "$REPO_ROOT"

SRCRUN="${SRCRUN:-$EVAL_DIR/runs/2026-06-09-full}"
GEN="$SCRIPT_DIR/generate.sh"
JUDGE="$SCRIPT_DIR/judge.sh"

# model slug -> exact id (only the two we generate)
declare -A MODEL_ID=( [s5]="claude-sonnet-5" [s46]="claude-sonnet-4-6" )
GEN_SLUGS=( s5 s46 )
# new pairs to judge (opus<->fable already exists and is copied in stage)
PAIRS=( "opus s5" "opus s46" "fable s5" "fable s46" "s46 s5" )

brand_dirs() { ls -d "$RUNDIR"/brand-* 2>/dev/null | sort; }

stage() {
  echo ">>> stage: copy reused artifacts from $SRCRUN into $RUNDIR" >&2
  local n=0
  for src in "$SRCRUN"/brand-*; do
    [ -d "$src" ] || continue
    local b; b="$(basename "$src")"
    local dst="$RUNDIR/$b"
    mkdir -p "$dst"
    for f in update.txt opus.md fable.md \
             verdict-ord1-Aopus-Bfable.json verdict-ord2-Afable-Bopus.json; do
      [ -f "$src/$f" ] && cp -n "$src/$f" "$dst/$f"
    done
    n=$((n+1))
  done
  echo ">>> staged $n brand dir(s)" >&2
}

build_gen_tasks() {
  local tasks="$1"; : > "$tasks"
  for d in $(brand_dirs); do
    local upd="$d/update.txt"
    [ -f "$upd" ] || { echo "WARN no update.txt in $d" >&2; continue; }
    for s in "${GEN_SLUGS[@]}"; do
      local out="$d/$s.md"
      [ -s "$out" ] && continue                    # resume
      printf '%s %s %s\n' "${MODEL_ID[$s]}" "$upd" "$out" >> "$tasks"
    done
  done
}

build_judge_tasks() {
  local tasks="$1"; : > "$tasks"
  for d in $(brand_dirs); do
    local upd="$d/update.txt"
    [ -f "$upd" ] || continue
    for p in "${PAIRS[@]}"; do
      set -- $p; local x="$1" y="$2"
      local xf="$d/$x.md" yf="$d/$y.md"
      [ -s "$xf" ] && [ -s "$yf" ] || { echo "WARN missing $xf or $yf" >&2; continue; }
      local o1="$d/verdict-ord1-A${x}-B${y}.json"
      local o2="$d/verdict-ord2-A${y}-B${x}.json"
      # resume: skip already-valid verdicts (judge.sh also self-skips)
      python3 -c "import json,sys;d=json.load(open('$o1'));sys.exit(0 if('winner'in d and'scores'in d)else 1)" 2>/dev/null \
        || printf '%s %s %s %s\n' "$upd" "$xf" "$yf" "$o1" >> "$tasks"
      python3 -c "import json,sys;d=json.load(open('$o2'));sys.exit(0 if('winner'in d and'scores'in d)else 1)" 2>/dev/null \
        || printf '%s %s %s %s\n' "$upd" "$yf" "$xf" "$o2" >> "$tasks"
    done
  done
}

run_gen() {
  local tasks; tasks="$(mktemp -t gen-tasks.XXXXXX)"
  build_gen_tasks "$tasks"
  local total; total=$(wc -l < "$tasks" | tr -d ' ')
  echo ">>> gen: $total generation task(s) pending, concurrency $CONC" >&2
  [ "$total" -eq 0 ] && { echo ">>> gen: nothing to do"; rm -f "$tasks"; return 0; }
  xargs -P "$CONC" -L 1 bash "$GEN" < "$tasks"
  rm -f "$tasks"
  local done; done=$(ls "$RUNDIR"/brand-*/s5.md "$RUNDIR"/brand-*/s46.md 2>/dev/null | wc -l | tr -d ' ')
  echo ">>> gen: $done / $(( $(brand_dirs | wc -l | tr -d ' ') * 2 )) Sonnet angle sets present" >&2
}

run_judge() {
  local tasks; tasks="$(mktemp -t judge-tasks.XXXXXX)"
  build_judge_tasks "$tasks"
  local total; total=$(wc -l < "$tasks" | tr -d ' ')
  echo ">>> judge: $total judge task(s) pending, concurrency $CONC" >&2
  [ "$total" -eq 0 ] && { echo ">>> judge: nothing to do"; rm -f "$tasks"; return 0; }
  xargs -P "$CONC" -L 1 bash "$JUDGE" < "$tasks" >/dev/null
  rm -f "$tasks"
  local nv; nv=$(ls "$RUNDIR"/brand-*/verdict-*.json 2>/dev/null | wc -l | tr -d ' ')
  echo ">>> judge: $nv verdict file(s) present (target 600)" >&2
}

case "$PHASE" in
  stage) stage ;;
  gen)   run_gen ;;
  judge) run_judge ;;
  all)   stage; run_gen; run_judge ;;
  *) echo "unknown PHASE: $PHASE" >&2; exit 1 ;;
esac
