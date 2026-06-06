#!/bin/sh
set -eu

NEWSJACK_REPO="${NEWSJACK_REPO:-elvisun/newsjack}"
NEWSJACK_REF="${NEWSJACK_REF:-main}"
NEWSJACK_HOME="${NEWSJACK_HOME:-$HOME/.newsjack}"
NEWSJACK_INSTALL_DIR="${NEWSJACK_INSTALL_DIR:-$NEWSJACK_HOME/newsjack}"
NEWSJACK_RUNTIMES="${NEWSJACK_RUNTIMES:-auto}"
NEWSJACK_INSTALL_SKILLS="${NEWSJACK_INSTALL_SKILLS:-1}"
NEWSJACK_INSTALL_MCP="${NEWSJACK_INSTALL_MCP:-1}"
NEWSJACK_FORCE="${NEWSJACK_FORCE:-0}"
NEWSJACK_RUN_SETUP="${NEWSJACK_RUN_SETUP:-auto}"

esc=$(printf '\033')
c_reset=
c_bold=
c_dim=
c_red=
c_green=
c_blue=
o_reset=
o_bold=
o_dim=
o_green=

case "${NEWSJACK_COLOR:-auto}" in
  always|1|true|yes|on)
    color_stderr=1
    color_stdout=1
    ;;
  never|0|false|no|off)
    color_stderr=0
    color_stdout=0
    ;;
  *)
    if [ "${NO_COLOR:-}" ]; then
      color_stderr=0
      color_stdout=0
    else
      color_stderr=0
      color_stdout=0
      [ "${TERM:-}" != "dumb" ] && [ -t 2 ] && color_stderr=1
      [ "${TERM:-}" != "dumb" ] && [ -t 1 ] && color_stdout=1
    fi
    ;;
esac

if [ "$color_stderr" = "1" ]; then
  c_reset="${esc}[0m"
  c_bold="${esc}[1m"
  c_dim="${esc}[2m"
  c_red="${esc}[31m"
  c_green="${esc}[32m"
  c_blue="${esc}[34m"
fi

if [ "$color_stdout" = "1" ]; then
  o_reset="${esc}[0m"
  o_bold="${esc}[1m"
  o_dim="${esc}[2m"
  o_green="${esc}[32m"
fi

banner() {
  bar=$(printf '%*s' 42 '' | sed 's/ /─/g')
  printf '%s┌%s┐%s\n' "$c_blue" "$bar" "$c_reset" >&2
  printf '%s│%s %s%-40s%s %s│%s\n' "$c_blue" "$c_reset" "$c_bold$c_blue" "newsjack" "$c_reset" "$c_blue" "$c_reset" >&2
  printf '%s├%s┤%s\n' "$c_blue" "$bar" "$c_reset" >&2
  printf '%s│%s %s%-40s%s %s│%s\n' "$c_blue" "$c_reset" "$c_dim" "the operating system for agentic PR" "$c_reset" "$c_blue" "$c_reset" >&2
  printf '%s└%s┘%s\n\n' "$c_blue" "$bar" "$c_reset" >&2
}

log() {
  printf '%s[info]%s %s\n' "$c_blue" "$c_reset" "$*" >&2
}

success() {
  printf '%s[success]%s %s\n' "$c_green" "$c_reset" "$*" >&2
}

warn() {
  printf '%s[warn]%s %s\n' "$c_red" "$c_reset" "$*" >&2
}

note() {
  printf '%s-%s %s%s%s\n' "$c_dim" "$c_reset" "$c_dim" "$*" "$c_reset" >&2
}

die() {
  printf '%s[error]%s %s\n' "$c_red" "$c_reset" "$*" >&2
  exit 1
}

have() {
  command -v "$1" >/dev/null 2>&1
}

download() {
  download_url=$1
  if have curl; then
    curl -fsSL "$download_url"
  elif have wget; then
    wget -qO- "$download_url"
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

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

json_string() {
  printf '"%s"' "$(json_escape "$1")"
}

json_bool() {
  case "$1" in
    1|true|yes|on) printf 'true' ;;
    *) printf 'false' ;;
  esac
}

runtimes_json() {
  raw=$1
  raw=$(printf '%s' "$raw" | tr '[:space:]' ',')
  old_ifs=$IFS
  IFS=,
  first=1
  printf '['
  for item in $raw; do
    [ -n "$item" ] || continue
    if [ "$first" = "1" ]; then
      first=0
    else
      printf ', '
    fi
    json_string "$item"
  done
  IFS=$old_ifs
  printf ']'
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

release_base() {
  if [ "${NEWSJACK_RELEASE_BASE:-}" ]; then
    printf '%s\n' "${NEWSJACK_RELEASE_BASE%/}"
    return
  fi
  if [ "${NEWSJACK_VERSION:-}" ]; then
    printf 'https://github.com/%s/releases/download/%s\n' "$NEWSJACK_REPO" "$NEWSJACK_VERSION"
    return
  fi
  printf 'https://github.com/%s/releases/latest/download\n' "$NEWSJACK_REPO"
}

fetch_release() {
  tmp=$1
  platform=$(detect_platform)
  base=$(release_base)

  archive="$tmp/newsjack_$platform.tar.gz"
  artifact="newsjack_$platform.tar.gz"
  url="$base/$artifact"
  checksums="$tmp/checksums.txt"
  log "fetching newsjack release for $platform"
  download "$base/manifest.json" >"$tmp/release-manifest.json"
  download "$base/checksums.txt" >"$checksums"
  download "$url" >"$archive"

  expected=$(awk -v artifact="$artifact" '$2 == artifact {print $1}' "$checksums" | head -n 1)
  [ -n "$expected" ] || die "checksum missing for $artifact"
  actual=$(sha256_file "$archive")
  [ "$expected" = "$actual" ] || die "checksum mismatch for $artifact"
  success "verified checksum for $artifact"

  mkdir -p "$tmp/source"
  tar -xzf "$archive" -C "$tmp/source"
  [ -d "$tmp/source/skills" ] || die "artifact does not contain a skills directory"
  [ -f "$tmp/source/.newsjack-prebuilt" ] || die "artifact is missing the prebuilt marker"
  [ -f "$tmp/source/bin/newsjack" ] || die "artifact does not contain bin/newsjack"
  [ -f "$tmp/source/VERSION" ] || die "artifact does not contain VERSION"
  [ -f "$tmp/source/COMMIT" ] || die "artifact does not contain COMMIT"
  [ -f "$tmp/source/skills-manifest.json" ] || die "artifact does not contain skills-manifest.json"
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

  fetch_release "$tmp"
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
  if [ "$NEWSJACK_INSTALL_SKILLS" = "0" ]; then
    note "skipping runtime skill install because NEWSJACK_INSTALL_SKILLS=0"
    return
  fi

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

write_install_state() {
  mkdir -p "$NEWSJACK_HOME"
  version=
  commit=
  [ -f "$NEWSJACK_INSTALL_DIR/VERSION" ] && version=$(tr -d '[:space:]' <"$NEWSJACK_INSTALL_DIR/VERSION")
  [ -f "$NEWSJACK_INSTALL_DIR/COMMIT" ] && commit=$(tr -d '[:space:]' <"$NEWSJACK_INSTALL_DIR/COMMIT")
  [ -n "$version" ] || version="${NEWSJACK_VERSION:-source}"
  [ -n "$commit" ] || commit="${NEWSJACK_COMMIT:-}"
  if [ "$NEWSJACK_INSTALL_SKILLS" = "0" ]; then
    skills_mode=external
  else
    skills_mode=managed
  fi
  installed_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  install_url="${NEWSJACK_INSTALL_URL:-https://newsjack.sh}"
  {
    printf '{\n'
    printf '  "version": '; json_string "$version"; printf ',\n'
    printf '  "commit": '; json_string "$commit"; printf ',\n'
    printf '  "channel": "stable",\n'
    printf '  "repo": '; json_string "$NEWSJACK_REPO"; printf ',\n'
    printf '  "install_url": '; json_string "$install_url"; printf ',\n'
    printf '  "skills_mode": '; json_string "$skills_mode"; printf ',\n'
    printf '  "runtimes": '; runtimes_json "$NEWSJACK_RUNTIMES"; printf ',\n'
    printf '  "runtimes_raw": '; json_string "$NEWSJACK_RUNTIMES"; printf ',\n'
    printf '  "install_mcp": '; json_bool "$NEWSJACK_INSTALL_MCP"; printf ',\n'
    printf '  "installed_at": '; json_string "$installed_at"; printf '\n'
    printf '}\n'
  } >"$NEWSJACK_HOME/install.json"
  success "wrote install state to $NEWSJACK_HOME/install.json"
}

setup_command() {
  if command -v newsjack >/dev/null 2>&1; then
    printf '%s\n' "newsjack setup"
  else
    printf '%s\n' "$NEWSJACK_HOME/bin/newsjack setup"
  fi
}

setup_prompt() {
  printf '%s\n' "Use newsjack to set up monitoring for my company."
}

print_setup_continue_prompt() {
  printf '\n%sCONTINUE SETUP%s\n\n' "$o_bold" "$o_reset"
  printf 'Open your agent and send this prompt:\n\n'
  printf 'Prompt:\n'
  setup_prompt
  printf '\n'
}

print_install_summary() {
  printf '\n%sNEWSJACK INSTALLED%s\n\n' "$o_bold" "$o_reset"
  printf '  %s%-22s%s %s\n' "$o_dim" "bundle" "$o_reset" "$NEWSJACK_INSTALL_DIR"
  printf '  %s%-22s%s %s\n' "$o_dim" "cli" "$o_reset" "$NEWSJACK_HOME/bin/newsjack"
}

print_next_steps() {
  setup_cmd=$(setup_command)
  printf '\n%sNEXT%s\n' "$o_bold" "$o_reset"
  printf '  %s%s%s\n' "$o_green" "$setup_cmd" "$o_reset"
  printf '\n\n'
}

should_run_setup() {
  if [ "${NEWSJACK_AUTO_UPDATE_RUNNING:-}" = "1" ]; then
    return 1
  fi
  case "$NEWSJACK_RUN_SETUP" in
    0|false|no|off)
      return 1
      ;;
    1|true|yes|on)
      return 0
      ;;
  esac
  [ "$NEWSJACK_INSTALL_SKILLS" != "0" ] || return 1
  ( : </dev/tty >/dev/tty ) 2>/dev/null
}

run_setup() {
  should_run_setup || return 1
  setup_cmd=$(setup_command)
  printf '\n%sSETUP%s\n\n' "$o_bold" "$o_reset" > /dev/tty
  printf '  %s%s%s\n\n' "$o_green" "$setup_cmd" "$o_reset" > /dev/tty
  if "$NEWSJACK_HOME/bin/newsjack" setup </dev/tty >/dev/tty 2>/dev/tty; then
    return 0
  fi
  warn "setup did not complete"
  print_setup_continue_prompt > /dev/tty
  return 0
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
  write_install_state
  print_install_summary
  if ! run_setup; then
    print_next_steps
  fi
}

main "$@"
