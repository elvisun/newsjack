#!/usr/bin/env bash
# Open a disposable container with the branch's newsjack CLI installed, the
# TypeSafe key saved, and fixture runs mounted, ready to test the Jev
# coarse-filter path. Builds the harness image (Claude runtime) if missing.
#
# usage: harness/scripts/open-jev-shell.sh [--image IMAGE] [--env-file FILE] [--rebuild]
# env file default: harness/.env.local (git-ignored). Expected keys:
#   TYPESAFE_API_KEY (required for live Jev calls), optional MEDIALYST_API_KEY,
#   ANTHROPIC_API_KEY (for running the detector skill in Claude Code).
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:jev}"
ENV_FILE="$REPO_DIR/harness/.env.local"
REBUILD=0

while [ "$#" -gt 0 ]; do
  case "$1" in
    --image) IMAGE="$2"; shift 2 ;;
    --env-file) ENV_FILE="$2"; shift 2 ;;
    --rebuild) REBUILD=1; shift ;;
    -h|--help) sed -n '2,10p' "$0"; exit 0 ;;
    *) echo "unknown option: $1" >&2; exit 2 ;;
  esac
done

if [ "$REBUILD" = "1" ] || ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "harness-jev: building $IMAGE (Claude runtime only)" >&2
  "$REPO_DIR/harness/scripts/build-image.sh" --harness claude --image "$IMAGE"
fi

docker_args=(
  --rm -it
  --mount "type=bind,src=${REPO_DIR},dst=/repo"
  --workdir /repo
  --env "HOME=/tmp/newsjack-home"
  --env "XDG_CONFIG_HOME=/tmp/newsjack-home/.config"
  --env "XDG_CACHE_HOME=/tmp/newsjack-home/.cache"
  --env "XDG_DATA_HOME=/tmp/newsjack-home/.local/share"
  --env "PATH=/tmp/newsjack-home/.newsjack/bin:/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin"
  --env "NEWSJACK_AUTO_UPDATE=0"
)

if [ -f "$ENV_FILE" ]; then
  rel="${ENV_FILE#"$REPO_DIR"/}"
  if [ "$rel" != "$ENV_FILE" ] && ! git -C "$REPO_DIR" check-ignore -q "$rel"; then
    echo "harness-jev: refusing to pass repo env file that is not git-ignored: $rel" >&2
    exit 1
  fi
  docker_args+=(--env-file "$ENV_FILE")
  echo "harness-jev: loading env file: $rel" >&2
else
  echo "harness-jev: no env file at $ENV_FILE; live Jev calls will need a key inside the container" >&2
fi

if [ ! -d "$REPO_DIR/fixtures/newsjack-detector-agent/runs/20260603T175505Z_simular" ]; then
  echo "harness-jev: fixture run 20260603T175505Z_simular not found under fixtures/newsjack-detector-agent/runs; the cheat sheet's chunk commands will not work" >&2
fi

exec docker run "${docker_args[@]}" "$IMAGE" \
  bash -c 'source /repo/harness/scripts/jev-container-setup.sh && exec bash'
