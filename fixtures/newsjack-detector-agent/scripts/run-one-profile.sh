#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
NEWSJACK_BIN="$("$SCRIPT_DIR/resolve-newsjack-bin.sh")"

SLUG="${NEWSJACK_PROFILE_SLUG:-${1:-simular}}"
QUERY="${NEWSJACK_PROFILE_QUERY:-${2:-computer-use agents}}"
PROFILE="${NEWSJACK_PROFILE_FILE:-${3:-profile.simular.json}}"
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
RUN_DIR="${NEWSJACK_RUN_DIR:-$FIXTURE_DIR/runs/$STAMP-$SLUG-profile}"
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
SCORED_CANDIDATES="$RUN_DIR/scored_candidates.json"
STDERR_LOG="$RUN_DIR/detector.stderr.log"
SUMMARY="$RUN_DIR/summary.json"
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
  echo "newsjack_bin=$NEWSJACK_BIN"
} > "$COMMAND_LOG"

echo "one-profile fixture run: $RUN_DIR"
echo "newsjack binary: $NEWSJACK_BIN"
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
  detector_args+=(--include-all-scored --scored-output "$SCORED_CANDIDATES")
fi

detector_args+=("$@")

{
  printf 'detector_command='
  printf '%q ' "$SCRIPT_DIR/with-fixture-env.sh" "$NEWSJACK_BIN" "${detector_args[@]}"
  echo
} >> "$COMMAND_LOG"

if "$SCRIPT_DIR/with-fixture-env.sh" \
  "$NEWSJACK_BIN" "${detector_args[@]}" \
    > "$CANDIDATES" 2> "$STDERR_LOG"; then
  echo "detector_status=ok" >> "$COMMAND_LOG"
else
  status=$?
  echo "detector_status=failed:$status" >> "$COMMAND_LOG"
  echo "failed: see $STDERR_LOG" >&2
  exit "$status"
fi

"$NEWSJACK_BIN" run-summary "$CANDIDATES" --output "$SUMMARY"

echo "wrote:"
echo "- $CANDIDATES"
if [[ -f "$SCORED_CANDIDATES" ]]; then
  echo "- $SCORED_CANDIDATES"
fi
echo "- $SUMMARY"
echo "- $STDERR_LOG"
echo "- $COMMAND_LOG"
echo
echo "note: this fixture writes detector JSON and a machine-readable summary. A pitch-ready run.md requires the full skill pipeline."
