#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:local}"
ENV_FILE="${NEWSJACK_HARNESS_ENV_FILE:-}"

usage() {
  cat <<'USAGE'
Usage: harness/scripts/docker-shell.sh [options]

Options:
  --image <image>       Harness container image (default: newsjack-agent-harness:local)
  --with-local-env      Load harness/.env.local into the container
  --env-file <path>     Load an explicit Docker env file into the container
  --help, -h            Show this help

This script passes env files at docker-run time only. It never prints env values.
USAGE
}

log() {
  printf '%s\n' "harness-shell: $*" >&2
}

abs_path() {
  local path="$1"
  local dir
  dir="$(cd "$(dirname "$path")" && pwd)"
  printf '%s/%s\n' "$dir" "$(basename "$path")"
}

validate_env_file() {
  local file="$1"
  [ -f "$file" ] || { log "env file not found: $file"; exit 1; }

  local abs
  abs="$(abs_path "$file")"
  case "$abs" in
    "$REPO_DIR"/*)
      local rel="${abs#"$REPO_DIR"/}"
      if ! git -C "$REPO_DIR" check-ignore -q "$rel"; then
        log "refusing to pass repo env file that is not git-ignored: $rel"
        exit 1
      fi
      ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --image)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      IMAGE="$2"
      shift 2
      ;;
    --with-local-env)
      ENV_FILE="$REPO_DIR/harness/.env.local"
      shift
      ;;
    --env-file)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      ENV_FILE="$2"
      shift 2
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

docker_args=(
  --rm
  -it
  --mount "type=bind,src=${REPO_DIR},dst=/repo"
  --workdir /repo
)

if [ -n "$ENV_FILE" ]; then
  validate_env_file "$ENV_FILE"
  docker_args+=(--env-file "$ENV_FILE")
  log "loading env file into container: ${ENV_FILE#$REPO_DIR/}"
fi

docker_args+=(
  --env "HOME=/tmp/newsjack-home"
  --env "XDG_CONFIG_HOME=/tmp/newsjack-home/.config"
  --env "XDG_CACHE_HOME=/tmp/newsjack-home/.cache"
  --env "XDG_DATA_HOME=/tmp/newsjack-home/.local/share"
  --env "PATH=/tmp/newsjack-home/.newsjack/bin:/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin"
)

docker run "${docker_args[@]}" "$IMAGE" bash
