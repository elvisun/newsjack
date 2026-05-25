#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '%s\n' "harness-runtimes: $*" >&2
}

log "installing Node runtime CLIs"
npm install -g \
  "@openai/codex@${CODEX_VERSION}" \
  "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
  "openclaw@${OPENCLAW_VERSION}" \
  "acpx@${ACPX_VERSION}" \
  "@zed-industries/codex-acp@${CODEX_ACP_VERSION}" \
  "@zed-industries/claude-agent-acp@${CLAUDE_AGENT_ACP_VERSION}"

log "installing Hermes Agent"
pipx install "hermes-agent[acp]==${HERMES_AGENT_VERSION}"

log "installed versions"
codex --version || true
claude --version || true
openclaw --version || true
hermes --version || true
acpx --version || true
codex-acp --version || true
claude-agent-acp --version || true

