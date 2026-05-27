#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:local}"
RUNTIMES="${NEWSJACK_RUNTIMES:-all}"
SOURCE_MODE="local"
INSTALLER_URL="${NEWSJACK_INSTALLER_URL:-https://newsjack.sh}"

usage() {
  cat <<'USAGE'
Usage: harness/scripts/run-ci-installer.sh [options]

Options:
  --runtime <runtime>      Runtime target: all, codex, claude, openclaw, hermes
  --image <image>          Harness container image (default: newsjack-agent-harness:local)
  --local-source           Build and install from the checked-out source tree (default)
  --production-path        Install through the hosted newsjack.sh path
  --installer-url <url>    Hosted installer URL for --production-path
USAGE
}

log() {
  printf '%s\n' "ci-installer: $*" >&2
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --runtime)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      RUNTIMES="$2"
      shift 2
      ;;
    --image)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      IMAGE="$2"
      shift 2
      ;;
    --local-source)
      SOURCE_MODE="local"
      shift
      ;;
    --production-path)
      SOURCE_MODE="production"
      shift
      ;;
    --installer-url)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      INSTALLER_URL="$2"
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

if ! command -v docker >/dev/null 2>&1; then
  log "docker is required to run the installer harness"
  exit 1
fi

log "running ${SOURCE_MODE} installer smoke in ${IMAGE} for runtimes: ${RUNTIMES}"

docker run --rm --interactive \
  --mount "type=bind,src=${REPO_DIR},dst=/repo,readonly" \
  --workdir /repo \
  --env "HOME=/tmp/newsjack-home" \
  --env "XDG_CONFIG_HOME=/tmp/newsjack-home/.config" \
  --env "XDG_CACHE_HOME=/tmp/newsjack-home/.cache" \
  --env "XDG_DATA_HOME=/tmp/newsjack-home/.local/share" \
  --env "NEWSJACK_RUNTIMES=${RUNTIMES}" \
  --env "NEWSJACK_NO_AUTO_UPDATE=1" \
  --env "NEWSJACK_HARNESS_SOURCE_MODE=${SOURCE_MODE}" \
  --env "NEWSJACK_INSTALLER_URL=${INSTALLER_URL}" \
  "${IMAGE}" \
  bash -s <<'EOF'
set -euo pipefail

log() {
  printf '%s\n' "ci-installer(container): $*" >&2
}

selected_skill_dirs() {
  local raw="${NEWSJACK_RUNTIMES:-all}"
  raw="${raw//claude-code/claude}"
  raw="${raw//claude_code/claude}"
  raw="${raw//cladue/claude}"
  raw="${raw//,/ }"

  local dirs=()
  local rt
  for rt in $raw; do
    case "$rt" in
      all|auto)
        printf '%s\n' \
          "$HOME/.agents/skills" \
          "$HOME/.claude/skills" \
          "$HOME/.openclaw/skills" \
          "$HOME/.hermes/skills"
        return
        ;;
      none)
        return
        ;;
      codex)
        dirs+=("$HOME/.agents/skills")
        ;;
      claude)
        dirs+=("$HOME/.claude/skills")
        ;;
      openclaw)
        dirs+=("$HOME/.openclaw/skills")
        ;;
      hermes)
        dirs+=("$HOME/.hermes/skills")
        ;;
      "")
        ;;
      *)
        log "unsupported runtime selection: $rt"
        exit 1
        ;;
    esac
  done

  printf '%s\n' "${dirs[@]}"
}

export PATH="$HOME/.newsjack/bin:$PATH"
export NEWSJACK_NO_AUTO_UPDATE=1

case "${NEWSJACK_HARNESS_SOURCE_MODE:-local}" in
  local)
    log "building and testing local CLI"
    (cd /repo/apps/cli && go test ./... && go build -buildvcs=false -o /tmp/newsjack ./cmd/newsjack)

    log "running local-source installer"
    NEWSJACK_SOURCE_DIR=/repo \
    NEWSJACK_CLI_BINARY=/tmp/newsjack \
    NEWSJACK_INSTALL_MCP=1 \
    NEWSJACK_FORCE=1 \
    /repo/install.sh
    ;;
  production)
    log "running hosted installer: ${NEWSJACK_INSTALLER_URL}"
    curl -fsSL "${NEWSJACK_INSTALLER_URL}" | \
      NEWSJACK_RUNTIMES="${NEWSJACK_RUNTIMES}" \
      NEWSJACK_INSTALL_MCP=1 \
      NEWSJACK_FORCE=1 \
      sh
    ;;
  *)
    log "unsupported source mode: ${NEWSJACK_HARNESS_SOURCE_MODE}"
    exit 1
    ;;
esac

cd /tmp

log "checking installed CLI"
newsjack version
newsjack doctor | tee /tmp/newsjack-doctor.json
jq -e '.root_ok == true and .dependencies.npx == true' /tmp/newsjack-doctor.json >/dev/null

log "checking installed runtime skills"
mapfile -t skill_dirs < <(selected_skill_dirs)
for dir in "${skill_dirs[@]}"; do
  test -f "$dir/newsjack-detector/SKILL.md"
  test -f "$dir/newsjack-setup/SKILL.md"
  test ! -d "$dir/newsjack-detector/scripts"
done

log "running mock detector smoke"
newsjack detector run "AI search visibility" --mock --limit 1 --emit json \
  | tee /tmp/newsjack-detector.json \
  | jq -e '.monitor.mock == true and (.monitor.queries | index("AI search visibility")) and (.signals | length) >= 1' >/dev/null

log "installer smoke complete"
EOF
