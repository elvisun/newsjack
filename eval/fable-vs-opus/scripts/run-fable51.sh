#!/usr/bin/env bash
# Six-model round-robin extension for the Fable 5.1 launch:
#   Fable 5.1 (f51) · Opus 5 (o5) · Fable 5 (fable) · Opus 4.8 (opus)
#   · Sonnet 5 (s5) · Sonnet 4.6 (s46)
#
# Only Fable 5.1 is generated. Every other angle set is reused byte-for-byte
# (o5/fable/opus from 2026-07-24-opus5; s5/s46 from 2026-06-30-4model), as are
# all 800 existing judgments. Fresh judging covers the 5 new Fable-5.1 pairs and
# the 2 pairs the earlier runs never crossed (Opus 5 vs each Sonnet), so the
# round-robin is complete: C(6,2)=15 pairs x 100 = 1500 judgments.
#
# Usage:
#   run-fable51.sh RUNDIR PHASE [CONCURRENCY]
#     PHASE = stage | gen | judge | all
set -uo pipefail

RUNDIR="$(cd "$(dirname "${1:?need RUNDIR}")" && pwd)/$(basename "$1")"
PHASE="${2:?need PHASE: stage|gen|judge|all}"
CONC="${3:-4}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$EVAL_DIR/../.." && pwd)"
cd "$REPO_ROOT"

SRC_O5="${SRC_O5:-$EVAL_DIR/runs/2026-07-24-opus5}"
SRC_4M="${SRC_4M:-$EVAL_DIR/runs/2026-06-30-4model}"
GEN="$SCRIPT_DIR/generate.sh"
JUDGE="$SCRIPT_DIR/judge.sh"
NEW="f51"
NEW_ID="claude-fable-5-1"
# Frozen ETHICS.md (sha 897f4fd0…) so the generator apparatus matches every prior run.
export ETHICS_FILE="$RUNDIR/apparatus/ETHICS.md"
# Pairs to judge fresh: "x y" -> ord1 A=x B=y, ord2 A=y B=x
PAIRS=( "fable f51" "o5 f51" "opus f51" "s5 f51" "s46 f51" "o5 s5" "o5 s46" )

brand_dirs() { find "$RUNDIR" -maxdepth 1 -type d -name 'brand-*' -print 2>/dev/null | sort; }

stage() {
  echo ">>> stage: copy frozen artifacts into $RUNDIR" >&2
  local n=0
  for src in "$SRC_O5"/brand-*; do
    [ -d "$src" ] || continue
    local b dst s4
    b="$(basename "$src")"; dst="$RUNDIR/$b"; s4="$SRC_4M/$b"
    mkdir -p "$dst"
    for f in update.txt opus.md fable.md o5.md \
             verdict-ord1-Aopus-Bfable.json verdict-ord2-Afable-Bopus.json \
             verdict-ord1-Afable-Bo5.json verdict-ord2-Ao5-Bfable.json \
             verdict-ord1-Aopus-Bo5.json verdict-ord2-Ao5-Bopus.json; do
      [ -f "$src/$f" ] && cp -n "$src/$f" "$dst/$f"
    done
    if [ -d "$s4" ]; then
      # the 4-model run must have used the same frozen fable/opus sets
      for m in fable opus; do
        cmp -s "$src/$m.md" "$s4/$m.md" || echo "WARN $b: $m.md differs between source runs" >&2
      done
      for f in s5.md s46.md \
               verdict-ord1-Afable-Bs46.json verdict-ord1-Afable-Bs5.json \
               verdict-ord1-Aopus-Bs46.json verdict-ord1-Aopus-Bs5.json \
               verdict-ord1-As46-Bs5.json verdict-ord2-As46-Bfable.json \
               verdict-ord2-As46-Bopus.json verdict-ord2-As5-Bfable.json \
               verdict-ord2-As5-Bopus.json verdict-ord2-As5-Bs46.json; do
        [ -f "$s4/$f" ] && cp -n "$s4/$f" "$dst/$f"
      done
    else
      echo "WARN no 4-model dir for $b" >&2
    fi
    n=$((n+1))
  done
  echo ">>> staged $n brand dir(s)" >&2
}

valid_verdict() {
  python3 -c "import json,sys;d=json.load(open('$1'));sys.exit(0 if('winner'in d and'scores'in d)else 1)" 2>/dev/null
}

build_gen_tasks() {
  local tasks="$1"; : > "$tasks"
  while IFS= read -r d; do
    [ -f "$d/update.txt" ] || { echo "WARN no update.txt in $d" >&2; continue; }
    [ -s "$d/$NEW.md" ] && continue
    printf '%s %s %s\n' "$NEW_ID" "$d/update.txt" "$d/$NEW.md" >> "$tasks"
  done < <(brand_dirs)
}

build_judge_tasks() {
  local tasks="$1"; : > "$tasks"
  while IFS= read -r d; do
    local upd="$d/update.txt"
    [ -f "$upd" ] || continue
    for p in "${PAIRS[@]}"; do
      set -- $p
      local x="$1" y="$2" xf="$d/$1.md" yf="$d/$2.md"
      [ -s "$xf" ] && [ -s "$yf" ] || { echo "WARN missing $xf or $yf" >&2; continue; }
      local o1="$d/verdict-ord1-A${x}-B${y}.json" o2="$d/verdict-ord2-A${y}-B${x}.json"
      valid_verdict "$o1" || printf '%s %s %s %s\n' "$upd" "$xf" "$yf" "$o1" >> "$tasks"
      valid_verdict "$o2" || printf '%s %s %s %s\n' "$upd" "$yf" "$xf" "$o2" >> "$tasks"
    done
  done < <(brand_dirs)
}

run_gen() {
  local tasks total
  tasks="$(mktemp -t f51-gen-tasks.XXXXXX)"
  build_gen_tasks "$tasks"
  total=$(wc -l < "$tasks" | tr -d ' ')
  echo ">>> gen: $total Fable 5.1 generation task(s) pending, concurrency $CONC" >&2
  [ "$total" -eq 0 ] && { rm -f "$tasks"; return 0; }
  xargs -P "$CONC" -L 1 bash "$GEN" < "$tasks"
  rm -f "$tasks"
  echo ">>> gen: $(find "$RUNDIR" -path "*/brand-*/$NEW.md" -type f -size +0c | wc -l | tr -d ' ') / $(brand_dirs | wc -l | tr -d ' ') Fable 5.1 angle sets present" >&2
}

run_judge() {
  local tasks total
  tasks="$(mktemp -t f51-judge-tasks.XXXXXX)"
  build_judge_tasks "$tasks"
  total=$(wc -l < "$tasks" | tr -d ' ')
  echo ">>> judge: $total judge task(s) pending, concurrency $CONC" >&2
  [ "$total" -eq 0 ] && { rm -f "$tasks"; return 0; }
  xargs -P "$CONC" -L 1 bash "$JUDGE" < "$tasks" >/dev/null
  rm -f "$tasks"
  echo ">>> judge: $(find "$RUNDIR" -path '*/brand-*/verdict-*.json' -type f | wc -l | tr -d ' ') verdict file(s) present (target 1500)" >&2
}

case "$PHASE" in
  stage) stage ;;
  gen) run_gen ;;
  judge) run_judge ;;
  all) stage; run_gen; run_judge ;;
  *) echo "unknown PHASE: $PHASE" >&2; exit 1 ;;
esac
