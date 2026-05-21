#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
QUERY="${NEWSJACK_QUERY:-${1:-AI customer support}}"
PROFILE="${NEWSJACK_PROFILE:-${2:-profile.chatbase.json}}"
if [[ "$#" -gt 0 ]]; then shift; fi
if [[ "$#" -gt 0 ]]; then shift; fi

exec "$SCRIPT_DIR/agent-env.sh" \
  python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run \
  "$QUERY" \
  --profile "$PROFILE" \
  "$@" \
  --sources news_search,x \
  --lookback-days 7 \
  --depth quick \
  --save \
  --emit json
