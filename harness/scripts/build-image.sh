#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:local}"
NO_CACHE=0
declare -a HARNESSES=()

usage() {
  cat <<'USAGE'
Usage: harness/scripts/build-image.sh [options]

Options:
  --harness <runtime>      Preinstall runtime CLI: all, none, codex, claude, openclaw, hermes.
                           Repeat this flag to test combinations.
  --image <image>          Image tag (default: newsjack-agent-harness:local)
  --no-cache               Build without Docker layer cache
  --help, -h               Show this help

Examples:
  harness/scripts/build-image.sh --harness all
  harness/scripts/build-image.sh --harness claude --harness openclaw
  harness/scripts/build-image.sh --harness none --image newsjack-agent-harness:none
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --harness)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      HARNESSES+=("$2")
      shift 2
      ;;
    --image)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      IMAGE="$2"
      shift 2
      ;;
    --no-cache)
      NO_CACHE=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
done

if [ "${#HARNESSES[@]}" -eq 0 ]; then
  HARNESSES=(all)
fi

normalized=""
for rt in "${HARNESSES[@]}"; do
  rt="${rt//claude-code/claude}"
  rt="${rt//claude_code/claude}"
  rt="${rt//cladue/claude}"
  case "$rt" in
    all|none|codex|claude|openclaw|hermes)
      ;;
    *)
      printf 'unsupported harness runtime: %s\n' "$rt" >&2
      exit 2
      ;;
  esac
  if [ -z "$normalized" ]; then
    normalized="$rt"
  else
    normalized="$normalized,$rt"
  fi
done

declare -a build_args=()
if [ "$NO_CACHE" = "1" ]; then
  build_args+=(--no-cache)
fi

docker build "${build_args[@]}" \
  -f "$REPO_DIR/harness/Dockerfile" \
  --build-arg "NEWSJACK_HARNESS_RUNTIMES=$normalized" \
  -t "$IMAGE" \
  "$REPO_DIR"
