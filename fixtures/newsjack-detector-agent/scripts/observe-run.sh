#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
NEWSJACK_BIN="${NEWSJACK_BIN:-$FIXTURE_DIR/../../bin/newsjack}"

SLUG="${NEWSJACK_OBSERVE_SLUG:-${1:-simular}}"
QUERY="${NEWSJACK_OBSERVE_QUERY:-${2:-computer-use agents}}"
PROFILE="${NEWSJACK_OBSERVE_PROFILE:-${3:-profile.simular.json}}"
if [[ "$#" -gt 0 ]]; then shift; fi
if [[ "$#" -gt 0 ]]; then shift; fi
if [[ "$#" -gt 0 ]]; then shift; fi

if [[ "$PROFILE" != /* ]]; then
  if [[ -f "$FIXTURE_DIR/$PROFILE" ]]; then
    PROFILE="$FIXTURE_DIR/$PROFILE"
  elif [[ -f "$PROFILE" ]]; then
    PROFILE="$(cd "$(dirname "$PROFILE")" && pwd)/$(basename "$PROFILE")"
  fi
fi

STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
RUN_DIR="${NEWSJACK_RUN_DIR:-$FIXTURE_DIR/runs/$STAMP-$SLUG-observe}"
SOURCES="${NEWSJACK_SOURCES:-news_search,x}"
LOOKBACK_DAYS="${NEWSJACK_LOOKBACK_DAYS:-1}"
DEPTH="${NEWSJACK_DEPTH:-quick}"
LIMIT="${NEWSJACK_LIMIT:-80}"
MIN_QUEUE_PRIORITY="${NEWSJACK_MIN_QUEUE_PRIORITY:-40}"
MIN_MAJOR_NEWS="${NEWSJACK_MIN_MAJOR_NEWS:-0.55}"
MAX_AGE_HOURS="${NEWSJACK_MAX_AGE_HOURS:-24}"
INCLUDE_ALL_SCORED="${NEWSJACK_INCLUDE_ALL_SCORED:-1}"

mkdir -p "$RUN_DIR"

CANDIDATES="$RUN_DIR/candidates.json"
STDERR_LOG="$RUN_DIR/detector.stderr.log"
SUMMARY="$RUN_DIR/summary.json"
RUN_MARKDOWN="$RUN_DIR/run.md"
COMMAND_LOG="$RUN_DIR/commands.log"

{
  echo "timestamp=$STAMP"
  echo "slug=$SLUG"
  echo "query=$QUERY"
  echo "profile=$PROFILE"
  echo "sources=$SOURCES"
  echo "lookback_days=$LOOKBACK_DAYS"
  echo "depth=$DEPTH"
  echo "limit=$LIMIT"
  echo "min_queue_priority=$MIN_QUEUE_PRIORITY"
  echo "min_major_news=$MIN_MAJOR_NEWS"
  echo "max_age_hours=$MAX_AGE_HOURS"
  echo "include_all_scored=$INCLUDE_ALL_SCORED"
} > "$COMMAND_LOG"

echo "observed run: $RUN_DIR"
echo "running detector..."

detector_args=(
  detector run
  "$QUERY"
  --profile "$PROFILE"
  --sources "$SOURCES"
  --lookback-days "$LOOKBACK_DAYS"
  --depth "$DEPTH"
  --limit "$LIMIT"
  --min-queue-priority "$MIN_QUEUE_PRIORITY"
  --min-major-news "$MIN_MAJOR_NEWS"
  --max-age-hours "$MAX_AGE_HOURS"
)

if [[ "$INCLUDE_ALL_SCORED" != "0" && "$INCLUDE_ALL_SCORED" != "false" ]]; then
  detector_args+=(--include-all-scored)
fi

detector_args+=("$@" --emit json)

{
  printf 'detector_command='
  printf '%q ' "$SCRIPT_DIR/agent-env.sh" "$NEWSJACK_BIN" "${detector_args[@]}"
  echo
} >> "$COMMAND_LOG"

if "$SCRIPT_DIR/agent-env.sh" \
  "$NEWSJACK_BIN" "${detector_args[@]}" \
    > "$CANDIDATES" 2> "$STDERR_LOG"; then
  echo "detector_status=ok" >> "$COMMAND_LOG"
else
  status=$?
  echo "detector_status=failed:$status" >> "$COMMAND_LOG"
  echo "failed: see $STDERR_LOG" >&2
  exit "$status"
fi

"$NEWSJACK_BIN" summarize-run "$CANDIDATES" --output "$SUMMARY" --markdown "$RUN_MARKDOWN"

echo "wrote:"
echo "- $CANDIDATES"
echo "- $SUMMARY"
echo "- $RUN_MARKDOWN"
echo "- $STDERR_LOG"
echo "- $COMMAND_LOG"
