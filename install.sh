#!/bin/sh
set -eu

NEWSJACK_REPO="${NEWSJACK_REPO:-elvisun/newsjack}"
NEWSJACK_REF="${NEWSJACK_REF:-main}"
NEWSJACK_HOME="${NEWSJACK_HOME:-$HOME/.newsjack}"
NEWSJACK_INSTALL_DIR="${NEWSJACK_INSTALL_DIR:-$NEWSJACK_HOME/newsjack}"
NEWSJACK_RUNTIMES="${NEWSJACK_RUNTIMES:-auto}"
NEWSJACK_INSTALL_MCP="${NEWSJACK_INSTALL_MCP:-1}"
NEWSJACK_FORCE="${NEWSJACK_FORCE:-0}"

log() {
  printf '%s\n' "newsjack: $*" >&2
}

die() {
  printf '%s\n' "newsjack: error: $*" >&2
  exit 1
}

have() {
  command -v "$1" >/dev/null 2>&1
}

download() {
  url=$1
  if have curl; then
    curl -fsSL "$url"
  elif have wget; then
    wget -qO- "$url"
  else
    die "curl or wget is required"
  fi
}

copy_tree() {
  ct_src=$1
  ct_dest=$2
  rm -rf "$ct_dest"
  mkdir -p "$ct_dest"
  (cd "$ct_src" && tar -cf - .) | (cd "$ct_dest" && tar -xf -)
}

fetch_source() {
  tmp=$1

  if [ "${NEWSJACK_SOURCE_DIR:-}" ]; then
    src=$NEWSJACK_SOURCE_DIR
    [ -d "$src/skills" ] || die "NEWSJACK_SOURCE_DIR does not look like the newsjack repo: $src"
    copy_tree "$src" "$tmp/source"
    printf '%s\n' "$tmp/source"
    return
  fi

  archive="$tmp/newsjack.tar.gz"
  url="${NEWSJACK_TARBALL_URL:-https://codeload.github.com/$NEWSJACK_REPO/tar.gz/$NEWSJACK_REF}"
  log "fetching $NEWSJACK_REPO@$NEWSJACK_REF"
  download "$url" >"$archive"
  tar -xzf "$archive" -C "$tmp"
  src=$(find "$tmp" -mindepth 1 -maxdepth 1 -type d ! -name source | head -n 1)
  [ -n "$src" ] || die "could not unpack newsjack archive"
  [ -d "$src/skills" ] || die "archive does not contain a skills directory"
  printf '%s\n' "$src"
}

install_source() {
  src=$1
  mkdir -p "$NEWSJACK_HOME"

  new_dir="$NEWSJACK_INSTALL_DIR.new"
  old_dir="$NEWSJACK_INSTALL_DIR.previous"

  rm -rf "$new_dir"
  copy_tree "$src" "$new_dir"

  rm -rf "$old_dir"
  if [ -d "$NEWSJACK_INSTALL_DIR" ]; then
    mv "$NEWSJACK_INSTALL_DIR" "$old_dir"
  fi
  mv "$new_dir" "$NEWSJACK_INSTALL_DIR"
  log "installed source to $NEWSJACK_INSTALL_DIR"
}

install_cli() {
  mkdir -p "$NEWSJACK_HOME/bin"

  if [ "${NEWSJACK_CLI_BINARY:-}" ]; then
    [ -f "$NEWSJACK_CLI_BINARY" ] || die "NEWSJACK_CLI_BINARY not found: $NEWSJACK_CLI_BINARY"
    cp "$NEWSJACK_CLI_BINARY" "$NEWSJACK_HOME/bin/newsjack"
  else
    [ -d "$NEWSJACK_INSTALL_DIR/apps/cli" ] || die "missing Go CLI source at $NEWSJACK_INSTALL_DIR/apps/cli"
    have go || die "Go is required to build newsjack until release binaries are published"
    (cd "$NEWSJACK_INSTALL_DIR/apps/cli" && go build -buildvcs=false -o "$NEWSJACK_HOME/bin/newsjack" ./cmd/newsjack)
  fi

  chmod 755 "$NEWSJACK_HOME/bin/newsjack"
  log "installed newsjack CLI to $NEWSJACK_HOME/bin/newsjack"
}

run_go_install() {
  set -- "$NEWSJACK_HOME/bin/newsjack" install \
    --source "$NEWSJACK_INSTALL_DIR" \
    --runtimes "$NEWSJACK_RUNTIMES"

  if [ "$NEWSJACK_INSTALL_MCP" = "0" ]; then
    set -- "$@" --mcp=false
  fi

  if [ "$NEWSJACK_FORCE" = "1" ]; then
    set -- "$@" --force
  fi

  "$@"
}

print_next_steps() {
  cat <<EOF

newsjack installed.

Installed bundle: $NEWSJACK_INSTALL_DIR
CLI:              $NEWSJACK_HOME/bin/newsjack

Next:
  newsjack skills list
  newsjack doctor
  newsjack login   # optional Medialyst API key for MCP-backed media lists

If 'newsjack' is not on PATH, add this to your shell profile:
  export PATH="\$HOME/.newsjack/bin:\$PATH"
EOF
}

main() {
  tmp=$(mktemp -d "${TMPDIR:-/tmp}/newsjack.XXXXXX")
  trap 'rm -rf "$tmp"' EXIT HUP INT TERM

  src=$(fetch_source "$tmp")
  install_source "$src"
  install_cli
  run_go_install
  print_next_steps
}

main "$@"
