#!/bin/sh
set -eu

NEWSJACK_REPO="${NEWSJACK_REPO:-elvisun/newsjack}"
NEWSJACK_REF="${NEWSJACK_REF:-main}"
NEWSJACK_HOME="${NEWSJACK_HOME:-$HOME/.newsjack}"
NEWSJACK_INSTALL_DIR="${NEWSJACK_INSTALL_DIR:-$NEWSJACK_HOME/newsjack}"
NEWSJACK_RUNTIMES="${NEWSJACK_RUNTIMES:-auto}"
NEWSJACK_INSTALL_MCP="${NEWSJACK_INSTALL_MCP:-1}"
NEWSJACK_FORCE="${NEWSJACK_FORCE:-0}"
NEWSJACK_DIST_BASE="${NEWSJACK_DIST_BASE:-https://newsjack.sh/dist}"
NEWSJACK_CHANNEL="${NEWSJACK_CHANNEL:-main}"

banner() {
  cat >&2 <<'EOF'
newsjack
the operating system for agentic PR

EOF
}

log() {
  printf '%s\n' "[info] $*" >&2
}

success() {
  printf '%s\n' "[success] $*" >&2
}

note() {
  printf '%s\n' "- $*" >&2
}

die() {
  printf '%s\n' "[error] $*" >&2
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

detect_platform() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m | tr '[:upper:]' '[:lower:]')

  case "$os" in
    darwin) os=darwin ;;
    linux) os=linux ;;
    *) die "unsupported operating system: $os" ;;
  esac

  case "$arch" in
    x86_64|amd64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) die "unsupported architecture: $arch" ;;
  esac

  printf '%s_%s\n' "$os" "$arch"
}

sha256_file() {
  file=$1
  if have sha256sum; then
    sha256sum "$file" | awk '{print $1}'
  elif have shasum; then
    shasum -a 256 "$file" | awk '{print $1}'
  else
    die "sha256sum or shasum is required"
  fi
}

ensure_compiled_binary() {
  bin=$1
  [ -f "$bin" ] || die "missing CLI binary: $bin"
  if [ "$(dd if="$bin" bs=2 count=1 2>/dev/null)" = "#!" ]; then
    die "CLI at $bin is a script; expected a compiled binary"
  fi
  "$bin" version >/dev/null 2>&1 || die "CLI at $bin is not executable on this platform"
}

copy_tree() {
  ct_src=$1
  ct_dest=$2
  rm -rf "$ct_dest"
  mkdir -p "$ct_dest"
  (cd "$ct_src" && tar -cf - .) | (cd "$ct_dest" && tar -xf -)
  find "$ct_dest" -type f \( -name ".env" -o -name ".env.*" \) ! -name ".env.example" -exec rm -f {} \;
}

fetch_dist() {
  tmp=$1
  platform=$(detect_platform)
  version="${NEWSJACK_VERSION:-}"

  if [ -z "$version" ]; then
    version=$(download "$NEWSJACK_DIST_BASE/channels/$NEWSJACK_CHANNEL.txt" | tr -d '[:space:]')
  fi
  [ -n "$version" ] || die "could not resolve newsjack version from channel: $NEWSJACK_CHANNEL"

  archive="$tmp/newsjack_$platform.tar.gz"
  artifact="newsjack_$platform.tar.gz"
  url="$NEWSJACK_DIST_BASE/commits/$version/$artifact"
  log "fetching newsjack $version for $platform"
  download "$url" >"$archive"

  expected=$(download "$url.sha256" | awk '{print $1}')
  actual=$(sha256_file "$archive")
  [ "$expected" = "$actual" ] || die "checksum mismatch for $artifact"
  success "verified checksum for $artifact"

  mkdir -p "$tmp/source"
  tar -xzf "$archive" -C "$tmp/source"
  [ -d "$tmp/source/skills" ] || die "artifact does not contain a skills directory"
  [ -f "$tmp/source/.newsjack-prebuilt" ] || die "artifact is missing the prebuilt marker"
  [ -f "$tmp/source/bin/newsjack" ] || die "artifact does not contain bin/newsjack"
  ensure_compiled_binary "$tmp/source/bin/newsjack"
  printf '%s\n' "$tmp/source"
}

fetch_source_archive() {
  tmp=$1
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

fetch_source() {
  tmp=$1

  if [ "${NEWSJACK_SOURCE_DIR:-}" ]; then
    src=$NEWSJACK_SOURCE_DIR
    [ -d "$src/skills" ] || die "NEWSJACK_SOURCE_DIR does not look like the newsjack repo: $src"
    copy_tree "$src" "$tmp/source"
    printf '%s\n' "$tmp/source"
    return
  fi

  if [ "${NEWSJACK_TARBALL_URL:-}" ]; then
    fetch_source_archive "$tmp"
    return
  fi

  fetch_dist "$tmp"
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
  success "installed bundle to $NEWSJACK_INSTALL_DIR"
}

install_cli() {
  mkdir -p "$NEWSJACK_HOME/bin"
  cli_dest="$NEWSJACK_HOME/bin/newsjack"
  cli_tmp="$NEWSJACK_HOME/bin/newsjack.new"
  rm -f "$cli_tmp"

  if [ "${NEWSJACK_CLI_BINARY:-}" ]; then
    [ -f "$NEWSJACK_CLI_BINARY" ] || die "NEWSJACK_CLI_BINARY not found: $NEWSJACK_CLI_BINARY"
    cp "$NEWSJACK_CLI_BINARY" "$cli_tmp"
  elif [ -f "$NEWSJACK_INSTALL_DIR/.newsjack-prebuilt" ] && [ -f "$NEWSJACK_INSTALL_DIR/bin/newsjack" ]; then
    cp "$NEWSJACK_INSTALL_DIR/bin/newsjack" "$cli_tmp"
  else
    die "missing prebuilt CLI binary in $NEWSJACK_INSTALL_DIR; set NEWSJACK_CLI_BINARY for source installs"
  fi

  chmod 755 "$cli_tmp"
  ensure_compiled_binary "$cli_tmp"
  mv "$cli_tmp" "$cli_dest"
  success "installed CLI to $cli_dest"
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
  if command -v newsjack >/dev/null 2>&1; then
    setup_cmd="newsjack setup"
  else
    setup_cmd="$NEWSJACK_HOME/bin/newsjack setup"
  fi

  cat <<EOF

NEWSJACK INSTALLED

  bundle                 $NEWSJACK_INSTALL_DIR
  cli                    $NEWSJACK_HOME/bin/newsjack

NEXT
  $setup_cmd
EOF
}

main() {
  banner
  note "curl path installs the CLI, copies skills, then hands off to newsjack setup."
  tmp=$(mktemp -d "${TMPDIR:-/tmp}/newsjack.XXXXXX")
  trap 'rm -rf "$tmp"' EXIT HUP INT TERM

  src=$(fetch_source "$tmp")
  install_source "$src"
  install_cli
  run_go_install
  print_next_steps
}

main "$@"
