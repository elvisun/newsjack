#!/usr/bin/env bash
# Headless angle-set generator for the model round-robin study.
#
# Runs one Claude model (addressed by EXACT id, e.g. claude-sonnet-5,
# claude-sonnet-4-6) as a clean-context angle-generator and captures its markdown
# angle set to OUT_FILE. Mirror of judge.sh: the apparatus (generator role + the
# angle-generator skill + ETHICS + WHY-NOT-SPAM) is assembled into one
# self-contained prompt and piped to `claude -p` with ALL tools disabled, so the
# run is a pure, fact-bounded completion — no web, no outside knowledge, no
# permission prompts. The only variable across models is --model.
#
# Why exact ids and not the Workflow model alias: the alias resolves only coarse
# tiers ('sonnet' -> current Sonnet), so it can't address a specific prior Sonnet
# (4.6). Reused opus.md/fable.md keep their original apparatus; this script
# produces the two Sonnet sets on matched, deterministic footing.
#
# Usage:
#   generate.sh MODEL_ID UPDATE_FILE OUT_FILE
#
# Idempotent: if OUT_FILE already exists and is non-empty, it is left untouched
# (resumability). A failed/empty generation leaves no OUT_FILE so a rerun retries.

set -euo pipefail

MODEL="${1:?need MODEL_ID}"
UPDATE_FILE="${2:?need UPDATE_FILE}"
OUT_FILE="${3:?need OUT_FILE}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$EVAL_DIR/../.." && pwd)"

GEN_MD="$EVAL_DIR/harness/generator.md"
SKILL="$REPO_ROOT/skills/angle-generator/SKILL.md"
ETHICS="$REPO_ROOT/skills/ETHICS.md"
NOSPAM="$REPO_ROOT/skills/WHY-NOT-SPAM.md"

for f in "$GEN_MD" "$SKILL" "$ETHICS" "$NOSPAM" "$UPDATE_FILE"; do
  [ -f "$f" ] || { echo "generate.sh: missing file: $f" >&2; exit 2; }
done

# Resume: keep a good prior result.
if [ -s "$OUT_FILE" ]; then
  echo "generate.sh: skip (exists) $OUT_FILE" >&2
  exit 0
fi

PROMPT_FILE="$(mktemp -t angle-gen-prompt.XXXXXX)"
trap 'rm -f "$PROMPT_FILE"' EXIT

{
  cat "$GEN_MD"
  printf '\n\nNOTE: In this run you have NO file or web tools. The skill, ETHICS,\n'
  printf 'and WHY-NOT-SPAM documents referenced above are pasted in full below —\n'
  printf 'read and apply them from here. Use ONLY the company update as facts.\n'
  printf '\n\n=== angle-generator/SKILL.md ===\n'
  cat "$SKILL"
  printf '\n\n=== ETHICS.md ===\n'
  cat "$ETHICS"
  printf '\n\n=== WHY-NOT-SPAM.md ===\n'
  cat "$NOSPAM"
  printf '\n\n=== COMPANY UPDATE (the only facts you may use) ===\n'
  cat "$UPDATE_FILE"
  printf '\n\n=== YOUR TASK ===\n'
  printf 'Apply the angle-generator skill in the default `pitch` mode to the company\n'
  printf 'update above. Output ONLY the readable markdown angle list exactly as the\n'
  printf "skill's Output Format specifies (angles first, then Refused angles,\n"
  printf 'Uncomfortable questions, Next step). No preamble, no meta-commentary, no\n'
  printf 'mention of evaluation. Begin your response with the first angle.\n'
} > "$PROMPT_FILE"

# Pure completion: disable every tool so this is deterministic and fact-bounded.
claude -p --model "$MODEL" \
  --disallowed-tools Bash Read Write Edit MultiEdit NotebookEdit WebSearch WebFetch Glob Grep Task TodoWrite \
  < "$PROMPT_FILE" > "${OUT_FILE}.part" 2> "${OUT_FILE%.md}.gen.log"

# Only promote a non-empty result.
if [ -s "${OUT_FILE}.part" ]; then
  mv "${OUT_FILE}.part" "$OUT_FILE"
  echo "generate.sh: wrote $OUT_FILE ($(wc -l < "$OUT_FILE") lines) via $MODEL" >&2
else
  rm -f "${OUT_FILE}.part"
  echo "generate.sh: EMPTY output for $OUT_FILE via $MODEL (see ${OUT_FILE%.md}.gen.log)" >&2
  exit 3
fi
