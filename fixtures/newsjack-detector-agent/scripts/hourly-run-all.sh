#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
NEWSJACK_BIN="${NEWSJACK_BIN:-$FIXTURE_DIR/../../bin/newsjack}"

STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
RUN_DIR="${NEWSJACK_RUN_DIR:-$FIXTURE_DIR/runs/$STAMP}"
SOURCES="${NEWSJACK_SOURCES:-news_search,x}"
LOOKBACK_DAYS="${NEWSJACK_LOOKBACK_DAYS:-1}"
DEPTH="${NEWSJACK_DEPTH:-quick}"
MAX_AGE_HOURS="${NEWSJACK_MAX_AGE_HOURS:-24}"
EXTRA_ARGS=("$@")

mkdir -p "$RUN_DIR"

profiles=(
  "bluebottle|specialty coffee|profile.bluebottle.json"
  "localfalcon|AI search visibility|profile.localfalcon.json"
  "nofar-method|reformer Pilates|profile.nofar-method.json"
  "simular|computer-use agents|profile.simular.json"
  "slite|AI knowledge base|profile.slite.json"
  "property-saviour|UK property chain collapse|profile.property-saviour.json"
  "clearnym|data broker removal|profile.clearnym.json"
)

status=0
report_rows=()

echo "newsjack hourly run: $STAMP"
echo "output: $RUN_DIR"

for row in "${profiles[@]}"; do
  IFS="|" read -r slug query profile <<<"$row"
  profile_dir="$RUN_DIR/$slug"
  output="$profile_dir/candidates.json"
  error_log="$profile_dir/detector.stderr.log"
  summary="$profile_dir/summary.json"
  markdown="$profile_dir/run.md"

  mkdir -p "$profile_dir"

  echo
  echo "== $slug =="
  echo "query: $query"
  echo "profile: $profile"
  profile_path="$FIXTURE_DIR/$profile"

  if "$SCRIPT_DIR/agent-env.sh" \
    "$NEWSJACK_BIN" detector run \
      "$query" \
      --profile "$profile_path" \
      --sources "$SOURCES" \
      --lookback-days "$LOOKBACK_DAYS" \
      --depth "$DEPTH" \
      "${EXTRA_ARGS[@]}" \
      --save \
      --new-only \
      --max-age-hours "$MAX_AGE_HOURS" \
      --emit json \
      >"$output" 2>"$error_log"; then
    if "$NEWSJACK_BIN" summarize-run "$output" --output "$summary" --markdown "$markdown"; then
      echo "ok: $markdown"
      report_rows+=("| $slug | [$profile]($slug/run.md) | [$slug/candidates.json]($slug/candidates.json) | ok |")
    else
      status=1
      echo "failed: $slug summary; see $summary / $markdown"
      report_rows+=("| $slug | [$profile]($slug/run.md) | [$slug/candidates.json]($slug/candidates.json) | summary failed |")
    fi
  else
    status=1
    echo "failed: $slug; see $error_log"
    report_rows+=("| $slug | $profile | $slug/candidates.json | detector failed |")
  fi
done

INDEX="$RUN_DIR/index.md"
{
  echo "# Newsjack Beta Run"
  echo
  echo "- Generated: $STAMP"
  echo "- Sources: $SOURCES"
  echo "- Lookback: $LOOKBACK_DAYS day(s)"
  echo "- Max source item age: $MAX_AGE_HOURS hour(s)"
  echo "- Depth: $DEPTH"
  echo "- Beta freshness gate: final surfaced stories require LLM-verified first public timestamp within 24 hours and canonical major coverage when recoverable"
  echo
  echo "| Beta profile | Report | Detector JSON | Status |"
  echo "|---|---|---|---|"
  for report_row in "${report_rows[@]}"; do
    echo "$report_row"
  done
  echo
  echo "Open each report's \`run.md\`. Detector-only reports are clearly marked as previews until a final editorial pass writes \`final_report.md\` and rerenders the brief."
} > "$INDEX"

echo
echo "done: $RUN_DIR"
echo "index: $INDEX"
exit "$status"
