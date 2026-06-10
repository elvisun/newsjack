#!/usr/bin/env bash
# Blind GPT-5.5 judge for the Fable-vs-Opus angle study.
#
# Drives `codex exec` (GPT-5.5) as an LLM-as-judge grounded in the meanest-editor
# skill. The judge sees only the company update and two anonymized angle sets (A,
# B) — never which model wrote which. Output is schema-validated JSON.
#
# Usage:
#   judge.sh UPDATE_FILE A_FILE B_FILE OUT_FILE
#
# UPDATE_FILE  company update text both models were given
# A_FILE       angle set shown to the judge as "A"
# B_FILE       angle set shown to the judge as "B"
# OUT_FILE     where the verdict JSON is written (codex --output-last-message)
#
# Caller owns the A/B -> model mapping (for position-bias counterbalancing) and
# records it alongside OUT_FILE. This script is deliberately blind.

set -euo pipefail

UPDATE_FILE="${1:?need UPDATE_FILE}"
A_FILE="${2:?need A_FILE}"
B_FILE="${3:?need B_FILE}"
OUT_FILE="${4:?need OUT_FILE}"

# Resolve repo paths relative to this script so it works from any cwd.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$EVAL_DIR/../.." && pwd)"

JUDGE_MD="$EVAL_DIR/harness/judge.md"
SCHEMA="$EVAL_DIR/harness/judge-schema.json"
SKILL="$REPO_ROOT/skills/meanest-editor/SKILL.md"

for f in "$JUDGE_MD" "$SCHEMA" "$SKILL" "$UPDATE_FILE" "$A_FILE" "$B_FILE"; do
  [ -f "$f" ] || { echo "judge.sh: missing file: $f" >&2; exit 2; }
done

PROMPT_FILE="$(mktemp -t fable-judge-prompt.XXXXXX)"
trap 'rm -f "$PROMPT_FILE"' EXIT

{
  cat "$JUDGE_MD"
  printf '\n\n=== COMPANY UPDATE ===\n'
  cat "$UPDATE_FILE"
  printf '\n\n=== ANGLE SET A ===\n'
  cat "$A_FILE"
  printf '\n\n=== ANGLE SET B ===\n'
  cat "$B_FILE"
  printf '\n\n=== MEANEST EDITOR SKILL ===\n'
  cat "$SKILL"
} > "$PROMPT_FILE"

codex exec \
  --skip-git-repo-check \
  --sandbox read-only \
  --model gpt-5.5 \
  --output-schema "$SCHEMA" \
  --output-last-message "$OUT_FILE" \
  - < "$PROMPT_FILE" > /dev/null 2>"${OUT_FILE%.json}.codex.log"

# codex writes the schema-validated final message to OUT_FILE; surface it.
cat "$OUT_FILE"
