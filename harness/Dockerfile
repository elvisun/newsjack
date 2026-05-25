FROM node:22-bookworm

ARG CODEX_VERSION=0.133.0
ARG CLAUDE_CODE_VERSION=2.1.150
ARG OPENCLAW_VERSION=2026.5.22
ARG ACPX_VERSION=0.10.0
ARG CODEX_ACP_VERSION=0.15.0
ARG CLAUDE_AGENT_ACP_VERSION=0.23.1
ARG HERMES_AGENT_VERSION=0.14.0

ENV DEBIAN_FRONTEND=noninteractive
ENV PIPX_HOME=/opt/pipx
ENV PIPX_BIN_DIR=/usr/local/bin
ENV PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    git \
    jq \
    pipx \
    python3 \
    python3-venv \
    ripgrep \
    tmux \
  && rm -rf /var/lib/apt/lists/*

COPY harness/scripts/install-runtimes.sh /usr/local/bin/install-newsjack-harness-runtimes

RUN CODEX_VERSION="$CODEX_VERSION" \
    CLAUDE_CODE_VERSION="$CLAUDE_CODE_VERSION" \
    OPENCLAW_VERSION="$OPENCLAW_VERSION" \
    ACPX_VERSION="$ACPX_VERSION" \
    CODEX_ACP_VERSION="$CODEX_ACP_VERSION" \
    CLAUDE_AGENT_ACP_VERSION="$CLAUDE_AGENT_ACP_VERSION" \
    HERMES_AGENT_VERSION="$HERMES_AGENT_VERSION" \
    install-newsjack-harness-runtimes

WORKDIR /workspace

