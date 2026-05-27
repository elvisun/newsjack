#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '%s\n' "harness-runtimes: $*" >&2
}

raw="${NEWSJACK_HARNESS_RUNTIMES:-all}"
raw="${raw//claude-code/claude}"
raw="${raw//claude_code/claude}"
raw="${raw//cladue/claude}"
raw="${raw//,/ }"

declare -a packages=()
declare -A selected=()

select_runtime() {
  case "$1" in
    all|auto)
      selected[codex]=1
      selected[claude]=1
      selected[openclaw]=1
      selected[hermes]=1
      ;;
    none)
      ;;
    codex|claude|openclaw|hermes)
      selected["$1"]=1
      ;;
    "")
      ;;
    *)
      log "unsupported harness runtime: $1"
      exit 1
      ;;
  esac
}

for rt in $raw; do
  select_runtime "$rt"
done

if [ -z "${raw// }" ]; then
  select_runtime all
fi

if [ "${selected[codex]:-}" = "1" ]; then
  packages+=("@openai/codex@${CODEX_VERSION}")
  packages+=("@zed-industries/codex-acp@${CODEX_ACP_VERSION}")
fi
if [ "${selected[claude]:-}" = "1" ]; then
  packages+=("@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}")
  packages+=("@zed-industries/claude-agent-acp@${CLAUDE_AGENT_ACP_VERSION}")
fi
if [ "${selected[openclaw]:-}" = "1" ]; then
  packages+=("openclaw@${OPENCLAW_VERSION}")
fi
if [ "${selected[hermes]:-}" = "1" ]; then
  packages+=("hermes-agent@${HERMES_AGENT_VERSION}")
  packages+=("acpx@${ACPX_VERSION}")
fi

if [ "${#packages[@]}" -gt 0 ]; then
  log "installing Node runtime CLIs: ${!selected[*]}"
  npm install -g "${packages[@]}"
else
  log "installing no agent runtime CLIs"
fi

log "installed versions"
for bin in codex claude openclaw hermes hermes-agent acpx codex-acp claude-agent-acp; do
  if command -v "$bin" >/dev/null 2>&1; then
    "$bin" --version || true
  fi
done
