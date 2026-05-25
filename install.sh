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

warn() {
  printf '%s\n' "newsjack: warning: $*" >&2
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

json_escape() {
  # JSON string escaping for paths used in MCP command payloads.
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

yaml_dq_escape() {
  # YAML double-quoted scalar escaping for simple generated config blocks.
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
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
  src="$NEWSJACK_INSTALL_DIR/bin/newsjack"
  [ -f "$src" ] || return 0

  mkdir -p "$NEWSJACK_HOME/bin"
  cp "$src" "$NEWSJACK_HOME/bin/newsjack"
  chmod 755 "$NEWSJACK_HOME/bin/newsjack"
  log "installed newsjack shim to $NEWSJACK_HOME/bin/newsjack"
}

runtime_list() {
  printf '%s' "$NEWSJACK_RUNTIMES" |
    tr '[:upper:]' '[:lower:]' |
    tr '[:space:]' ',' |
    sed 's/claude-code/claude/g; s/claude_code/claude/g; s/cladue/claude/g; s/,,*/,/g; s/^,//; s/,$//'
}

detect_runtime() {
  runtime=$1
  case "$runtime" in
    codex)
      have codex || [ -d "$HOME/.codex" ] || [ -d "$HOME/.agents" ]
      ;;
    claude)
      have claude || [ -d "$HOME/.claude" ]
      ;;
    openclaw)
      have openclaw || [ -d "$HOME/.openclaw" ]
      ;;
    hermes)
      have hermes || [ -d "$HOME/.hermes" ]
      ;;
    *)
      return 1
      ;;
  esac
}

want_runtime() {
  runtime=$1
  selected=$(runtime_list)
  case ",$selected," in
    *,none,*)
      return 1
      ;;
    *,all,*)
      return 0
      ;;
    *,auto,*)
      detect_runtime "$runtime"
      return $?
      ;;
    *,"$runtime",*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

copy_one_skill() {
  src=$1
  target_root=$2
  name=$(basename "$src")
  dest="$target_root/$name"
  marker="$dest/.newsjack-installed"

  if [ -e "$dest" ] && [ ! -f "$marker" ] && [ "$NEWSJACK_FORCE" != "1" ]; then
    warn "skipping existing non-newsjack skill: $dest (set NEWSJACK_FORCE=1 to overwrite)"
    return 0
  fi

  mkdir -p "$target_root"
  tmp_dest="$dest.tmp"
  rm -rf "$tmp_dest"
  copy_tree "$src" "$tmp_dest"
  {
    printf 'repo=%s\n' "$NEWSJACK_REPO"
    printf 'ref=%s\n' "$NEWSJACK_REF"
    printf 'source=%s\n' "$NEWSJACK_INSTALL_DIR"
  } >"$tmp_dest/.newsjack-installed"
  rm -rf "$dest"
  mv "$tmp_dest" "$dest"
}

install_doctrine_files() {
  target_root=$1
  mkdir -p "$target_root"
  for file in ETHICS.md WHY-NOT-SPAM.md; do
    if [ -f "$NEWSJACK_INSTALL_DIR/skills/$file" ]; then
      file_marker="$target_root/.$file.newsjack-installed"
      if [ -e "$target_root/$file" ] && [ ! -f "$file_marker" ] && [ "$NEWSJACK_FORCE" != "1" ]; then
        warn "skipping existing non-newsjack doctrine file: $target_root/$file (set NEWSJACK_FORCE=1 to overwrite)"
        continue
      fi
      cp "$NEWSJACK_INSTALL_DIR/skills/$file" "$target_root/$file"
      : >"$file_marker"
    fi
  done
}

install_skills_to() {
  label=$1
  target_root=$2

  [ -d "$NEWSJACK_INSTALL_DIR/skills" ] || die "missing skills directory in $NEWSJACK_INSTALL_DIR"
  mkdir -p "$target_root"

  for skill_dir in "$NEWSJACK_INSTALL_DIR"/skills/*; do
    [ -d "$skill_dir" ] || continue
    [ -f "$skill_dir/SKILL.md" ] || continue
    copy_one_skill "$skill_dir" "$target_root"
  done
  install_doctrine_files "$target_root"
  log "installed skills for $label into $target_root"
}

configure_codex_mcp() {
  bridge=$1
  have codex || return 0
  [ "$NEWSJACK_INSTALL_MCP" = "1" ] || return 0

  if codex mcp add medialyst -- python3 "$bridge" >/dev/null 2>&1 </dev/null; then
    log "configured Codex MCP server: medialyst"
  else
    warn "Codex MCP server 'medialyst' may already exist; leaving existing config in place"
  fi
}

configure_claude_mcp() {
  bridge=$1
  have claude || return 0
  [ "$NEWSJACK_INSTALL_MCP" = "1" ] || return 0

  escaped_bridge=$(json_escape "$bridge")
  json="{\"type\":\"stdio\",\"command\":\"python3\",\"args\":[\"$escaped_bridge\"]}"
  if claude mcp add-json --scope user medialyst "$json" >/dev/null 2>&1 </dev/null; then
    log "configured Claude Code MCP server: medialyst"
  else
    warn "Claude Code MCP server 'medialyst' may already exist; leaving existing config in place"
  fi
}

configure_openclaw_mcp() {
  bridge=$1
  have openclaw || return 0
  [ "$NEWSJACK_INSTALL_MCP" = "1" ] || return 0

  escaped_bridge=$(json_escape "$bridge")
  json="{\"command\":\"python3\",\"args\":[\"$escaped_bridge\"]}"
  if openclaw mcp set medialyst "$json" >/dev/null 2>&1 </dev/null; then
    log "configured OpenClaw MCP server: medialyst"
  else
    warn "could not configure OpenClaw MCP automatically"
  fi
}

hermes_has_medialyst() {
  config=$1
  [ -f "$config" ] || return 1
  awk '
    /^[^[:space:]#][^:]*:/ {
      top=$0
      sub(/:.*/, "", top)
    }
    top == "mcp_servers" && /^[[:space:]]+medialyst:[[:space:]]*$/ {
      found=1
    }
    END { exit found ? 0 : 1 }
  ' "$config"
}

configure_hermes_mcp() {
  bridge=$1
  [ "$NEWSJACK_INSTALL_MCP" = "1" ] || return 0

  config="${HERMES_CONFIG:-$HOME/.hermes/config.yaml}"
  if hermes_has_medialyst "$config"; then
    log "Hermes MCP server already configured: medialyst"
    return 0
  fi

  mkdir -p "$(dirname "$config")"
  [ -f "$config" ] || : >"$config"

  escaped_bridge=$(yaml_dq_escape "$bridge")

  if grep -q '^mcp_servers:' "$config" && ! grep -q '^mcp_servers:[[:space:]]*\(#.*\)\?$' "$config"; then
    warn "Hermes config has an inline mcp_servers key; add medialyst manually in $config"
    return 0
  fi

  tmp_config="$config.newsjack-tmp"
  awk -v bridge="$escaped_bridge" '
    function print_block() {
      print "  medialyst:"
      print "    command: \"python3\""
      print "    args:"
      print "      - \"" bridge "\""
    }
    !inserted && /^mcp_servers:[[:space:]]*(#.*)?$/ {
      print
      print_block()
      inserted=1
      next
    }
    { print }
    END {
      if (!inserted) {
        print ""
        print "mcp_servers:"
        print_block()
      }
    }
  ' "$config" >"$tmp_config"
  mv "$tmp_config" "$config"
  log "configured Hermes MCP server: medialyst"
}

configure_mcp() {
  bridge="$NEWSJACK_INSTALL_DIR/skills/media-list-manager/scripts/medialyst_mcp_bridge.py"
  [ -f "$bridge" ] || return 0

  want_runtime codex && configure_codex_mcp "$bridge"
  want_runtime claude && configure_claude_mcp "$bridge"
  want_runtime openclaw && configure_openclaw_mcp "$bridge"
  want_runtime hermes && configure_hermes_mcp "$bridge"
  return 0
}

install_runtime_skills() {
  installed_any=0

  if want_runtime codex; then
    install_skills_to "Codex" "${NEWSJACK_CODEX_SKILLS_DIR:-$HOME/.agents/skills}"
    installed_any=1
  fi

  if want_runtime claude; then
    install_skills_to "Claude Code" "${NEWSJACK_CLAUDE_SKILLS_DIR:-$HOME/.claude/skills}"
    installed_any=1
  fi

  if want_runtime openclaw; then
    install_skills_to "OpenClaw" "${NEWSJACK_OPENCLAW_SKILLS_DIR:-$HOME/.openclaw/skills}"
    installed_any=1
  fi

  if want_runtime hermes; then
    install_skills_to "Hermes" "${NEWSJACK_HERMES_SKILLS_DIR:-$HOME/.hermes/skills}"
    installed_any=1
  fi

  if [ "$installed_any" = "0" ]; then
    warn "no supported runtime detected; installing portable skills into $HOME/.agents/skills"
    install_skills_to "portable Agent Skills" "$HOME/.agents/skills"
  fi
}

print_next_steps() {
  cat <<EOF

newsjack installed.

Installed repo: $NEWSJACK_INSTALL_DIR
CLI shim:       $NEWSJACK_HOME/bin/newsjack

Next:
  newsjack skills
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
  install_runtime_skills
  configure_mcp
  print_next_steps
}

main "$@"
