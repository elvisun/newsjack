# Open Source Installation and Distribution Spec

This is the target distribution model for Newsjack as an open-source project.
It supersedes the earlier Vercel-bundled latest-main distribution plan for
public release distribution.

## Hard Constraints

- `curl -fsSL newsjack.sh | bash` is a first-class install path and must install
  the full product by default: CLI, managed bundle, skills, and best-effort
  runtime/MCP setup.
- GitHub Releases are the canonical source for versioned CLI artifacts.
- Agent skill marketplaces are first-class discovery/install surfaces for the
  skill layer.
- Marketplace-installed skills may install the CLI on demand, but they must not
  silently overwrite marketplace-owned skill files.
- Users must not need Go installed for normal installs.
- The pipe installer must not ask for secrets. Secrets belong in `newsjack setup`,
  `newsjack login`, or a first live run that needs them.

## Distribution Surfaces

Newsjack has three public distribution surfaces:

1. **Branded full installer:** `newsjack.sh`
2. **Versioned artifacts:** GitHub Releases
3. **Skill marketplaces:** skills.sh and Claude plugin/skills marketplaces

These surfaces serve different jobs. `newsjack.sh` is the simplest complete
install path. GitHub Releases are the audit and version source. Marketplaces
install the instruction layer and bootstrap the CLI only when a workflow needs
local executable support.

## Branded Full Installer

The branded installer remains:

```bash
curl -fsSL newsjack.sh | bash
```

By default this installs the complete stack:

- `~/.newsjack/bin/newsjack`
- `~/.newsjack/newsjack`
- Newsjack-managed skill folders in detected runtimes
- optional Medialyst MCP configuration where the runtime has a reliable
  noninteractive setup path

Default ownership mode:

```text
skills_mode = managed
```

In managed mode, Newsjack owns the generated runtime skill files and may update
them on future `newsjack update` or auto-update runs.

The branded installer should still be served from `newsjack.sh`, but it should
download versioned CLI artifacts from GitHub Releases instead of treating a site
deployment as the release source.

## GitHub Release Artifacts

Every release tag, for example `v0.2.0`, should produce a GitHub Release with:

```text
install.sh
manifest.json
checksums.txt
newsjack_darwin_amd64.tar.gz
newsjack_darwin_arm64.tar.gz
newsjack_linux_amd64.tar.gz
newsjack_linux_arm64.tar.gz
```

Each platform archive contains:

```text
bin/newsjack
skills/
README.md
LICENSE
VERSION
COMMIT
manifest.json
skills-manifest.json
```

`VERSION` stores the release tag:

```text
v0.2.0
```

`COMMIT` stores the source commit SHA used to build the release.

`manifest.json` should include at least:

```json
{
  "version": "v0.2.0",
  "commit": "abc123...",
  "channel": "stable",
  "built_at": "2026-06-01T00:00:00Z",
  "artifacts": [
    {
      "name": "newsjack_darwin_arm64.tar.gz",
      "os": "darwin",
      "arch": "arm64",
      "sha256": "..."
    }
  ]
}
```

`skills-manifest.json` should include each packaged skill name, file hash, and
source path so `newsjack doctor` and `newsjack skills status` can report whether
managed local skills match the installed bundle.

## Installer Resolution

Latest stable install:

```bash
curl -fsSL newsjack.sh | bash
```

Pinned install:

```bash
curl -fsSL newsjack.sh | NEWSJACK_VERSION=v0.2.0 bash
```

Installer resolution rules:

- If `NEWSJACK_VERSION` is set, download assets from
  `https://github.com/<owner>/newsjack/releases/download/$NEWSJACK_VERSION/`.
- Otherwise resolve the latest stable GitHub Release and download from
  `https://github.com/<owner>/newsjack/releases/latest/download/`.
- Verify `checksums.txt` before extracting any archive.
- Install atomically by staging into `~/.newsjack/newsjack.new`, then moving the
  previous bundle to `~/.newsjack/newsjack.previous`.
- Install the compiled CLI to `~/.newsjack/bin/newsjack`.

The installer must support these environment variables:

```text
NEWSJACK_VERSION          release tag, for example v0.2.0
NEWSJACK_REPO             GitHub owner/repo override
NEWSJACK_HOME             install home, default ~/.newsjack
NEWSJACK_RUNTIMES         auto, all, none, codex, claude, openclaw, hermes
NEWSJACK_INSTALL_SKILLS   1 by default; 0 for marketplace CLI bootstrap
NEWSJACK_INSTALL_MCP      1 by default; 0 to skip MCP configuration
NEWSJACK_FORCE            0 by default; 1 to overwrite user-owned conflicts
NEWSJACK_AUTO_UPDATE      1 by default; 0 to disable CLI auto-update
```

## Installed State

The installer writes:

```text
~/.newsjack/install.json
```

Example full-domain install:

```json
{
  "version": "v0.2.0",
  "commit": "abc123...",
  "channel": "stable",
  "repo": "elvisun/newsjack",
  "install_url": "https://newsjack.sh",
  "skills_mode": "managed",
  "runtimes": ["codex", "claude"],
  "install_mcp": true,
  "installed_at": "2026-06-01T00:00:00Z"
}
```

Example marketplace CLI bootstrap:

```json
{
  "version": "v0.2.0",
  "commit": "abc123...",
  "channel": "stable",
  "repo": "elvisun/newsjack",
  "install_url": "https://newsjack.sh",
  "skills_mode": "external",
  "runtimes": [],
  "install_mcp": false,
  "installed_at": "2026-06-01T00:00:00Z"
}
```

`skills_mode` controls update behavior:

- `managed`: Newsjack may update runtime skill files that carry Newsjack marker
  files.
- `external`: Newsjack updates only the CLI and managed bundle. Marketplace-owned
  skill files are not overwritten unless the user explicitly runs a forced skill
  install.

## Auto-Update

Installed binaries auto-update before normal user-facing commands.

Auto-update checks GitHub Releases:

```text
installed VERSION != latest release manifest.version
```

When stale:

1. Run the branded installer with the previous install state.
2. Preserve `skills_mode`.
3. Preserve selected runtimes for managed installs.
4. Preserve `NEWSJACK_INSTALL_SKILLS=0` for external marketplace installs.
5. Re-run the original command on the updated binary.

Auto-update should skip commands that are installer/update internals or are
machine-facing plumbing, including:

- `help`
- `version`
- `install`
- `update`
- `auth`
- `mcp-bridge`

`NEWSJACK_AUTO_UPDATE=0` disables the behavior for CI and debugging.

## Marketplace Skill Path

Newsjack should publish marketplace-safe skills to:

- skills.sh
- Claude community or official plugin marketplace
- optional Claude.ai upload bundles

Marketplace install examples:

```bash
npx skills add <owner>/newsjack
```

```text
/plugin install newsjack@claude-community
```

Marketplace-installed skills are instruction packages first. They should work in
instruction-only mode when the CLI is unavailable.

CLI-backed workflows should check for the CLI before requiring it:

```bash
command -v newsjack || test -x "$HOME/.newsjack/bin/newsjack"
```

If the CLI is missing, the skill should ask for explicit user permission before
running:

```bash
curl -fsSL newsjack.sh | NEWSJACK_INSTALL_SKILLS=0 NEWSJACK_INSTALL_MCP=0 bash
```

This creates `skills_mode=external` and installs only the CLI plus managed bundle.
It does not overwrite the marketplace-owned skill files.

## Marketplace Package Shape

Canonical skill source remains:

```text
skills/<skill-name>/SKILL.md
skills/<skill-name>/references/
skills/<skill-name>/examples.md
skills/ETHICS.md
skills/WHY-NOT-SPAM.md
```

Generated Claude plugin package:

```text
plugins/newsjack/
  .claude-plugin/plugin.json
  skills/
    angle-generator/
    crisis-holding/
    fact-check/
    journalist-fit-check/
    meanest-editor/
    media-list-manager/
    newsjack-detector/
    newsjack-setup/
    newsworthiness-check/
    reactive-comment/
    relevance-coarse-filter/
    story-origin-check/
    voice-extractor/
```

`skills.sh.json` should group the repository page for scanability, but it does
not change installation semantics.

Claude.ai skill packages should be zipped with the skill directory as the zip
root's child directory, not with `SKILL.md` directly at the zip root.

## Skill Authoring Requirements

Marketplace skills must:

- keep `name` lowercase, numeric, or hyphenated
- keep `description` at or below 200 characters for Claude.ai compatibility
- avoid silent downloads or silent installation
- degrade gracefully when network, shell, or CLI access is unavailable
- clearly mark CLI-backed steps as requiring local execution
- never ask for API keys in the pipe installer path
- continue local artifact fallback behavior when MCP is unavailable

CLI-backed skills should include a short dependency section:

```markdown
## CLI Dependency

Some workflows use the local `newsjack` CLI. If it is missing, ask the user
before installing it:

`curl -fsSL newsjack.sh | NEWSJACK_INSTALL_SKILLS=0 NEWSJACK_INSTALL_MCP=0 bash`
```

## Release Workflows

Required GitHub Actions:

- `ci.yml`
  - Go tests
  - site checks if the site is touched
  - skill validation
  - local installer smoke
- `release.yml`
  - runs on `v*` tags
  - validates the tag
  - builds all platform archives
  - writes `manifest.json`, `skills-manifest.json`, and `checksums.txt`
  - creates or updates the GitHub Release
  - uploads all assets
- `post-release-smoke.yml`
  - installs through `newsjack.sh`
  - installs through direct GitHub Release assets
  - exercises `NEWSJACK_INSTALL_SKILLS=0`
  - verifies `newsjack version`, `newsjack skills list`, `newsjack doctor`, and a
    mock detector run

## Security And Trust

- GitHub tags and releases are the public audit surface.
- Checksums are required before extraction.
- Signed releases or provenance attestations should be added before enterprise
  positioning.
- Marketplace skills must never hide external install behavior.
- MCP setup is best effort and must not block CLI or skill installation.
- Existing non-Newsjack user skill files are never overwritten unless
  `NEWSJACK_FORCE=1` is set.

## Remaining Implementation Work

- Add marketplace-safe CLI dependency text to CLI-backed skills.
- Add `skills.sh.json`.
- Add Claude plugin package generation.
- Add signed release provenance or attestations.
- Run the post-release smoke workflow against the public repository after the
  repo visibility switch.
