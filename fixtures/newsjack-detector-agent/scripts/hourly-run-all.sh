#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
RUN_DIR="${NEWSJACK_RUN_DIR:-$FIXTURE_DIR/runs/$STAMP}"
SOURCES="${NEWSJACK_SOURCES:-news_search,x}"
LOOKBACK_DAYS="${NEWSJACK_LOOKBACK_DAYS:-1}"
DEPTH="${NEWSJACK_DEPTH:-quick}"
MAX_AGE_HOURS="${NEWSJACK_MAX_AGE_HOURS:-24}"
EXTRA_ARGS=("$@")

mkdir -p "$RUN_DIR"

profiles=(
  "localfalcon|AI search visibility|profile.localfalcon.json"
  "simular|computer-use agents|profile.simular.json"
  "slite|AI knowledge base|profile.slite.json"
  "property-saviour|UK property chain collapse|profile.property-saviour.json"
  "clearnym|data broker removal|profile.clearnym.json"
)

status=0

echo "newsjack hourly run: $STAMP"
echo "output: $RUN_DIR"

for row in "${profiles[@]}"; do
  IFS="|" read -r slug query profile <<<"$row"
  output="$RUN_DIR/$slug.json"
  error_log="$RUN_DIR/$slug.stderr.log"

  echo
  echo "== $slug =="
  echo "query: $query"
  echo "profile: $profile"

  if "$SCRIPT_DIR/agent-env.sh" \
    python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run \
      "$query" \
      --profile "$profile" \
      --sources "$SOURCES" \
      --lookback-days "$LOOKBACK_DAYS" \
      --depth "$DEPTH" \
      "${EXTRA_ARGS[@]}" \
      --save \
      --new-only \
      --max-age-hours "$MAX_AGE_HOURS" \
      --emit json \
      >"$output" 2>"$error_log"; then
    echo "ok: $output"
  else
    status=1
    echo "failed: $slug; see $error_log"
  fi
done

echo
echo "done: $RUN_DIR"
exit "$status"
