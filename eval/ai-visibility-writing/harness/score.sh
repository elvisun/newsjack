#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

MODE="dev"
INPUT=""
PLANT=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --fresh) MODE="fresh"; shift ;;
    --input) INPUT="${2:?--input needs a path}"; shift 2 ;;
    --plant) PLANT="${2:?--plant needs a value}"; shift 2 ;;
    *) echo "usage: score.sh [--fresh] [--input metrics.json] [--plant kind]" >&2; exit 2 ;;
  esac
done

lint_args=()
if [ -n "$PLANT" ]; then
  lint_args=(--plant "$PLANT")
fi
lint_output="$("$SCRIPT_DIR/lint.sh" "${lint_args[@]}" 2>/dev/null)" || {
  echo "VOID: constraint violation"
  exit 3
}

if ! python3 -m unittest discover -s "$EVAL_DIR/tests" -p 'test_*.py' >/tmp/ai-visibility-writing-tests.log 2>&1; then
  echo "VOID: constraint violation"
  exit 3
fi

if [ -z "$INPUT" ]; then
  INPUT="$EVAL_DIR/runs/$MODE/metrics.json"
fi
if [ ! -f "$INPUT" ]; then
  echo "INCOMPLETE: no scored $MODE run" >&2
  exit 4
fi

if [ "$MODE" = "fresh" ]; then
  exec python3 "$SCRIPT_DIR/score.py" --fresh --input "$INPUT"
fi
exec python3 "$SCRIPT_DIR/score.py" --input "$INPUT"
