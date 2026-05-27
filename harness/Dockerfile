FROM node:22-bookworm

ARG CODEX_VERSION=0.133.0
ARG CLAUDE_CODE_VERSION=2.1.150
ARG OPENCLAW_VERSION=2026.5.22
ARG HERMES_AGENT_VERSION=0.14.0
ARG ACPX_VERSION=0.10.0
ARG CODEX_ACP_VERSION=0.15.0
ARG CLAUDE_AGENT_ACP_VERSION=0.23.1
ARG GO_VERSION=1.26.3
ARG NEWSJACK_HARNESS_RUNTIMES=all

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    git \
    jq \
    python3-pip \
    ripgrep \
    tmux \
  && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
  arch="$(dpkg --print-architecture)"; \
  case "$arch" in \
    amd64) go_arch=amd64 ;; \
    arm64) go_arch=arm64 ;; \
    *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
  esac; \
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${go_arch}.tar.gz" -o /tmp/go.tgz; \
  rm -rf /usr/local/go; \
  tar -C /usr/local -xzf /tmp/go.tgz; \
  rm /tmp/go.tgz; \
  go version

COPY harness/scripts/install-runtimes.sh /usr/local/bin/install-newsjack-harness-runtimes

RUN CODEX_VERSION="$CODEX_VERSION" \
    CLAUDE_CODE_VERSION="$CLAUDE_CODE_VERSION" \
    OPENCLAW_VERSION="$OPENCLAW_VERSION" \
    HERMES_AGENT_VERSION="$HERMES_AGENT_VERSION" \
    ACPX_VERSION="$ACPX_VERSION" \
    CODEX_ACP_VERSION="$CODEX_ACP_VERSION" \
    CLAUDE_AGENT_ACP_VERSION="$CLAUDE_AGENT_ACP_VERSION" \
    NEWSJACK_HARNESS_RUNTIMES="$NEWSJACK_HARNESS_RUNTIMES" \
    PIP_BREAK_SYSTEM_PACKAGES=1 \
    install-newsjack-harness-runtimes

WORKDIR /workspace
