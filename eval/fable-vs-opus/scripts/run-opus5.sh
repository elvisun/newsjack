#!/usr/bin/env bash
# Three-model round-robin extension:
#   Opus 5 (o5) · Fable 5 (fable) · Opus 4.8 (opus)
#
# The frozen Opus 4.8 and Fable 5 angle sets and their already-counterbalanced
# judgments are reused from the original 50-brand study. Opus 5 is generated
# through headless Claude Code by exact model id, then its two new pairings are
# judged blind in both orderings by the unchanged GPT-5.5 judge.
#
# Usage:
#   run-opus5.sh RUNDIR PHASE [CONCURRENCY]
#     PHASE = stage | gen | judge | all
#     RUNDIR e.g. eval/fable-vs-opus/runs/2026-07-24-opus5
#
# Requires SRCRUN only when overriding the default frozen source run.
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
OPUS5_ID="claude-opus-5"
PAIRS=( "opus o5" "fable o5" )

brand_dirs() { find "$RUNDIR" -maxdepth 1 -type d -name 'brand-*' -print 2>/dev/null | sort; }

stage() {
  echo ">>> stage: copy frozen artifacts from $SRCRUN into $RUNDIR" >&2
  local n=0
  for src in "$SRCRUN"/brand-*; do
    [ -d "$src" ] || continue
    local b dst
    b="$(basename "$src")"
    dst="$RUNDIR/$b"
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
  local tasks="$1"
  : > "$tasks"
  while IFS= read -r d; do
    local upd="$d/update.txt"
    [ -f "$upd" ] || { echo "WARN no update.txt in $d" >&2; continue; }
    [ -s "$d/o5.md" ] && continue
    printf '%s %s %s\n' "$OPUS5_ID" "$upd" "$d/o5.md" >> "$tasks"
  done < <(brand_dirs)
}

build_judge_tasks() {
  local tasks="$1"
  : > "$tasks"
  while IFS= read -r d; do
    local upd="$d/update.txt"
    [ -f "$upd" ] || continue
    for p in "${PAIRS[@]}"; do
      set -- $p
      local x="$1" y="$2"
      local xf="$d/$x.md" yf="$d/$y.md"
      [ -s "$xf" ] && [ -s "$yf" ] || {
        echo "WARN missing $xf or $yf" >&2
        continue
      }
      local o1="$d/verdict-ord1-A${x}-B${y}.json"
      local o2="$d/verdict-ord2-A${y}-B${x}.json"
      python3 -c "import json,sys;d=json.load(open('$o1'));sys.exit(0 if('winner'in d and'scores'in d)else 1)" 2>/dev/null \
        || printf '%s %s %s %s\n' "$upd" "$xf" "$yf" "$o1" >> "$tasks"
      python3 -c "import json,sys;d=json.load(open('$o2'));sys.exit(0 if('winner'in d and'scores'in d)else 1)" 2>/dev/null \
        || printf '%s %s %s %s\n' "$upd" "$yf" "$xf" "$o2" >> "$tasks"
    done
  done < <(brand_dirs)
}

run_gen() {
  local tasks total
  tasks="$(mktemp -t opus5-gen-tasks.XXXXXX)"
  build_gen_tasks "$tasks"
  total=$(wc -l < "$tasks" | tr -d ' ')
  echo ">>> gen: $total Opus 5 generation task(s) pending, concurrency $CONC" >&2
  if [ "$total" -eq 0 ]; then
    echo ">>> gen: nothing to do" >&2
    rm -f "$tasks"
    return 0
  fi
  xargs -P "$CONC" -L 1 bash "$GEN" < "$tasks"
  rm -f "$tasks"
  local done target
  done=$(find "$RUNDIR" -path '*/brand-*/o5.md' -type f -size +0c | wc -l | tr -d ' ')
  target=$(brand_dirs | wc -l | tr -d ' ')
  echo ">>> gen: $done / $target Opus 5 angle sets present" >&2
}

run_judge() {
  local tasks total
  tasks="$(mktemp -t opus5-judge-tasks.XXXXXX)"
  build_judge_tasks "$tasks"
  total=$(wc -l < "$tasks" | tr -d ' ')
  echo ">>> judge: $total judge task(s) pending, concurrency $CONC" >&2
  if [ "$total" -eq 0 ]; then
    echo ">>> judge: nothing to do" >&2
    rm -f "$tasks"
    return 0
  fi
  xargs -P "$CONC" -L 1 bash "$JUDGE" < "$tasks" >/dev/null
  rm -f "$tasks"
  local nv
  nv=$(find "$RUNDIR" -path '*/brand-*/verdict-*.json' -type f | wc -l | tr -d ' ')
  echo ">>> judge: $nv verdict file(s) present (target 300)" >&2
}

case "$PHASE" in
  stage) stage ;;
  gen) run_gen ;;
  judge) run_judge ;;
  all) stage; run_gen; run_judge ;;
  *) echo "unknown PHASE: $PHASE" >&2; exit 1 ;;
esac
