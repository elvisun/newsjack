package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func installedBinaryName() string {
	if runtime.GOOS == "windows" {
		return "newsjack.exe"
	}
	return "newsjack"
}

func installedBinaryPath() string {
	return filepath.Join(newsjackHome(), "bin", installedBinaryName())
}

func managedInstallDir() string {
	return filepath.Join(newsjackHome(), "newsjack")
}

func releaseArtifactName() string {
	return fmt.Sprintf("newsjack_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
}

// bootstrapInstall turns a bare downloaded binary into a full install:
// fetch the release bundle for this platform, verify its checksum, unpack
// it, install this binary as the managed CLI, write install state, then
// run the regular skills and MCP install. No shell, tar, curl, Node, or
// git is required on the machine. runtimeSelection comes from setup's
// --runtime flag; NEWSJACK_RUNTIMES still wins when set.
func bootstrapInstall(runtimeSelection string, stdout, stderr io.Writer) error {
	base := releaseBaseForBootstrap()
	logf(stderr, "newsjack is not installed yet; fetching the release bundle from %s", base)
	version, err := applyReleaseBundle(base, stderr)
	if err != nil {
		return err
	}
	return finalizeBootstrap(version, runtimeSelection, stdout, stderr)
}

// finalizeBootstrap is the post-bundle half of first install: managed
// binary, install state, skills, MCP. Split out so an interrupted
// bootstrap (bundle applied, rest missing) can be resumed without
// re-downloading the release. The managed binary comes from the verified
// bundle, not the running exe: a stale downloaded binary must not become
// the installed CLI while install.json records the bundle's version.
func finalizeBootstrap(version, runtimeSelection string, stdout, stderr io.Writer) error {
	if err := updateInstalledBinaryFromBundle(); err != nil {
		return err
	}
	if err := writeInstallStateFile(bootstrapInstallState(version, runtimeSelection)); err != nil {
		return err
	}
	maybeAddWindowsUserPath(stderr)
	if os.Getenv("NEWSJACK_INSTALL_SKILLS") == "0" {
		successf(stdout, "installed newsjack %s (runtime skill install skipped: NEWSJACK_INSTALL_SKILLS=0)", version)
		return nil
	}
	opts := installOptions{
		Source:     managedInstallDir(),
		Runtimes:   bootstrapRuntimes(runtimeSelection),
		InstallMCP: os.Getenv("NEWSJACK_INSTALL_MCP") != "0",
		Force:      os.Getenv("NEWSJACK_FORCE") == "1",
		CLI:        newsjackCLIInvocation(),
		Repo:       getenv("NEWSJACK_REPO", defaultRepo),
		Ref:        getenv("NEWSJACK_REF", defaultRef),
	}
	if err := installRuntimeSkills(opts, stdout, stderr); err != nil {
		return err
	}
	if opts.InstallMCP {
		if err := configureMCP(opts, stdout, stderr); err != nil {
			warn(stderr, "%v", err)
		}
	}
	successf(stdout, "installed newsjack %s to %s", version, managedInstallDir())
	return nil
}

func bootstrapRuntimes(runtimeSelection string) string {
	if v := strings.TrimSpace(os.Getenv("NEWSJACK_RUNTIMES")); v != "" {
		return v
	}
	if v := strings.TrimSpace(runtimeSelection); v != "" && v != "other" && v != "manual" {
		return v
	}
	return "auto"
}

// ensureInstalledRoot bootstraps a bare binary on first run. Explicit
// NEWSJACK_ROOT overrides, source checkouts, and npm installs never
// trigger a release download. A managed root left behind by an
// interrupted bootstrap (bundle present, binary or state missing) is
// finished here without re-downloading.
func ensureInstalledRoot(runtimeSelection string, stdout, stderr io.Writer) error {
	if npmDistribution() {
		return nil
	}
	if os.Getenv("NEWSJACK_ROOT") != "" {
		_, err := newsjackRoot()
		return err
	}
	if root, err := newsjackRoot(); err == nil {
		if root != managedInstallDir() {
			return nil
		}
		if fileExists(installedBinaryPath()) && fileExists(installStatePath()) {
			return nil
		}
		logf(stderr, "finishing an interrupted newsjack install at %s", root)
		return finalizeBootstrap(readTrimmedFile(filepath.Join(root, "VERSION")), runtimeSelection, stdout, stderr)
	}
	return bootstrapInstall(runtimeSelection, stdout, stderr)
}

// releaseBaseForBootstrap picks where a bare binary fetches its bundle.
// Unlike update (which wants latest), first install must use the bundle
// matching the running binary: a prerelease exe pointed at `latest` would
// fail whenever latest does not carry this platform yet, and a stale exe
// must not install a mismatched newer bundle. Overrides:
// NEWSJACK_RELEASE_BASE (full URL) and NEWSJACK_VERSION (tag).
func releaseBaseForBootstrap() string {
	if base := strings.TrimSpace(os.Getenv("NEWSJACK_RELEASE_BASE")); base != "" {
		return strings.TrimRight(base, "/")
	}
	state := readInstallStateOrDefault()
	repo := strings.TrimSpace(getenv("NEWSJACK_REPO", state.Repo))
	if repo == "" {
		repo = defaultRepo
	}
	tag := strings.TrimSpace(getenv("NEWSJACK_VERSION", version))
	if isReleaseTag(tag) {
		return "https://github.com/" + repo + "/releases/download/" + tag
	}
	return "https://github.com/" + repo + "/releases/latest/download"
}

var releaseTagPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z][0-9A-Za-z.-]*)?$`)

// isReleaseTag reports whether a version string names a published release
// tag. Dev builds (v0.1.0-dev, v0.1.0-dev+commit) fall back to latest.
func isReleaseTag(tag string) bool {
	return releaseTagPattern.MatchString(tag) && !strings.Contains(tag, "dev")
}

// applyReleaseBundle downloads, verifies, unpacks, and atomically swaps in
// the platform release bundle, returning the installed version. The
// previous install is kept at <dir>.previous for rollback.
func applyReleaseBundle(base string, logw io.Writer) (string, error) {
	artifact := releaseArtifactName()
	checksums, err := fetchReleaseBytes(base + "/checksums.txt")
	if err != nil {
		return "", fmt.Errorf("fetching checksums.txt: %w", err)
	}
	expected := checksumFor(string(checksums), artifact)
	if expected == "" {
		return "", fmt.Errorf("release has no checksum for %s; this platform may not be supported by %s", artifact, base)
	}
	payload, err := fetchReleaseBytes(base + "/" + artifact)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", artifact, err)
	}
	sum := sha256.Sum256(payload)
	if hex.EncodeToString(sum[:]) != expected {
		return "", fmt.Errorf("checksum mismatch for %s", artifact)
	}
	logf(logw, "verified checksum for %s", artifact)

	if err := os.MkdirAll(newsjackHome(), 0o755); err != nil {
		return "", err
	}
	staging := managedInstallDir() + ".new"
	if err := os.RemoveAll(staging); err != nil {
		return "", err
	}
	if err := untarGz(payload, staging); err != nil {
		_ = os.RemoveAll(staging)
		return "", fmt.Errorf("unpacking %s: %w", artifact, err)
	}
	if err := validateBundleLayout(staging); err != nil {
		_ = os.RemoveAll(staging)
		return "", err
	}
	version := readTrimmedFile(filepath.Join(staging, "VERSION"))
	if err := swapInstallDir(staging); err != nil {
		return "", err
	}
	return version, nil
}

func checksumFor(checksums, artifact string) string {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == artifact {
			return fields[0]
		}
	}
	return ""
}

func validateBundleLayout(dir string) error {
	checks := []struct {
		path string
		dir  bool
	}{
		{"skills", true},
		{".newsjack-prebuilt", false},
		{filepath.Join("bin", installedBinaryName()), false},
		{"VERSION", false},
		{"COMMIT", false},
		{"skills-manifest.json", false},
	}
	for _, check := range checks {
		target := filepath.Join(dir, check.path)
		if check.dir && !dirExists(target) {
			return fmt.Errorf("release bundle is missing directory %s", check.path)
		}
		if !check.dir && !fileExists(target) {
			return fmt.Errorf("release bundle is missing %s", check.path)
		}
	}
	return nil
}

func swapInstallDir(staging string) error {
	installDir := managedInstallDir()
	previous := installDir + ".previous"
	if err := os.RemoveAll(previous); err != nil {
		return err
	}
	if dirExists(installDir) {
		if err := os.Rename(installDir, previous); err != nil {
			return err
		}
	}
	if err := os.Rename(staging, installDir); err != nil {
		if dirExists(previous) && !dirExists(installDir) {
			_ = os.Rename(previous, installDir)
		}
		return err
	}
	return nil
}

func untarGz(data []byte, dest string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(filepath.ToSlash(header.Name), "./")
		if name == "" {
			continue
		}
		clean := path.Clean(name)
		if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
			return fmt.Errorf("unsafe path in archive: %s", header.Name)
		}
		// Strip env files exactly like the shell installer's copy_tree.
		base := path.Base(clean)
		if (base == ".env" || strings.HasPrefix(base, ".env.")) && base != ".env.example" {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(clean))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := fs.FileMode(header.Mode) & 0o777
			if mode == 0 {
				mode = 0o644
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, reader); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		default:
			// Release bundles contain only files and directories.
			continue
		}
	}
}

// adoptBundle copies a prebuilt source bundle into the managed install
// root and finishes the managed layout (binary + install state), so a
// manually downloaded bundle given to `install --source` behaves exactly
// like a bootstrap install: doctor is clean, update works, and the MCP
// bridge command points at a binary that exists.
func adoptBundle(source, runtimeSelection string, stderr io.Writer) error {
	if err := os.MkdirAll(newsjackHome(), 0o755); err != nil {
		return err
	}
	staging := managedInstallDir() + ".new"
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	if err := copyDirTree(source, staging); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	if err := validateBundleLayout(staging); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	if err := swapInstallDir(staging); err != nil {
		return err
	}
	if err := updateInstalledBinaryFromBundle(); err != nil {
		return err
	}
	bundleVersion := readTrimmedFile(filepath.Join(managedInstallDir(), "VERSION"))
	if err := writeInstallStateFile(bootstrapInstallState(bundleVersion, runtimeSelection)); err != nil {
		return err
	}
	maybeAddWindowsUserPath(stderr)
	return nil
}

// shouldAdoptBundle reports whether an `install --source` target is a
// prebuilt bundle that should be promoted into the managed root. The
// installer scripts pass the managed root itself (no-op), and source
// checkouts have no prebuilt marker.
func shouldAdoptBundle(source string) bool {
	if npmDistribution() {
		return false
	}
	if !fileExists(filepath.Join(source, ".newsjack-prebuilt")) {
		return false
	}
	if samePath(source, managedInstallDir()) {
		return false
	}
	return !dirExists(managedInstallDir()) || !fileExists(installedBinaryPath()) || !fileExists(installStatePath())
}

// copyDirTree mirrors a bundle directory, preserving file modes and
// stripping env files exactly like untarGz.
func copyDirTree(src, dest string) error {
	return filepath.WalkDir(src, func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, walkPath)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dest, 0o755)
		}
		base := filepath.Base(rel)
		if !d.IsDir() && (base == ".env" || strings.HasPrefix(base, ".env.")) && base != ".env.example" {
			return nil
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(walkPath)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		mode := info.Mode() & 0o777
		if mode == 0 {
			mode = 0o644
		}
		return os.WriteFile(target, data, mode)
	})
}

// maybeAddWindowsUserPath appends the managed bin dir to the per-user
// Path registry value so new shells can run `newsjack` directly. The
// shell installers handle PATH on Unix; the bare-exe flow must do it
// itself. Best-effort: a failure only prints the manual step. Opt out
// with NEWSJACK_NO_PATH_UPDATE=1.
func maybeAddWindowsUserPath(stderr io.Writer) {
	if goos() != "windows" || os.Getenv("NEWSJACK_NO_PATH_UPDATE") == "1" {
		return
	}
	binDir := filepath.Dir(installedBinaryPath())
	current := readWindowsUserPath()
	if pathListContains(current, binDir) {
		return
	}
	updated := binDir
	if strings.TrimSpace(current) != "" {
		updated = strings.TrimRight(current, ";") + ";" + binDir
	}
	// reg.exe instead of setx: setx truncates Path values at 1024 chars.
	cmd := exec.Command("reg", "add", `HKCU\Environment`, "/v", "Path", "/t", "REG_EXPAND_SZ", "/d", updated, "/f")
	if out, err := cmd.CombinedOutput(); err != nil {
		warn(stderr, "could not add %s to your user PATH (%s); add it manually for new shells", binDir, strings.TrimSpace(string(out)))
		return
	}
	logf(stderr, "added %s to your user PATH; open a new terminal to run `newsjack` directly", binDir)
}

func readWindowsUserPath() string {
	out, err := exec.Command("reg", "query", `HKCU\Environment`, "/v", "Path").Output()
	if err != nil {
		return ""
	}
	return parseRegQueryValue(string(out))
}

// parseRegQueryValue extracts the data column from `reg query /v Path`:
//
//	HKEY_CURRENT_USER\Environment
//	    Path    REG_EXPAND_SZ    C:\foo;C:\bar
//
// Internal spacing in the value is preserved.
func parseRegQueryValue(out string) string {
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(trimmed), "path") {
			continue
		}
		typeIdx := strings.Index(trimmed, "REG_")
		if typeIdx < 0 {
			continue
		}
		rest := trimmed[typeIdx:]
		valueIdx := strings.IndexAny(rest, " \t")
		if valueIdx < 0 {
			continue
		}
		return strings.TrimSpace(rest[valueIdx:])
	}
	return ""
}

func pathListContains(list, dir string) bool {
	normalize := func(entry string) string {
		return strings.ToLower(strings.TrimRight(strings.TrimSpace(entry), `\`))
	}
	want := normalize(dir)
	for _, entry := range strings.Split(list, ";") {
		if entry != "" && normalize(entry) == want {
			return true
		}
	}
	return false
}

// updateInstalledBinaryFromBundle replaces the managed CLI binary with the
// one shipped in the freshly applied bundle. The destination may be the
// currently running executable: Windows allows renaming a running exe but
// not deleting or overwriting it, so the old binary is parked at .old and
// removed by cleanupStaleBinary on a later run.
func updateInstalledBinaryFromBundle() error {
	src := filepath.Join(managedInstallDir(), "bin", installedBinaryName())
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	dest := installedBinaryPath()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	staging := dest + ".new"
	if err := os.WriteFile(staging, data, 0o755); err != nil {
		return err
	}
	parked := dest + ".old"
	_ = os.Remove(parked)
	if fileExists(dest) {
		if err := os.Rename(dest, parked); err != nil {
			_ = os.Remove(staging)
			return err
		}
	}
	if err := os.Rename(staging, dest); err != nil {
		if fileExists(parked) && !fileExists(dest) {
			_ = os.Rename(parked, dest)
		}
		return err
	}
	if goos() != "windows" {
		_ = os.Remove(parked)
	}
	return nil
}

// cleanupStaleBinary removes the parked previous binary left behind by a
// Windows in-place update once it is no longer running. Removal fails
// silently while the old process still holds the file; a later run wins.
func cleanupStaleBinary() {
	_ = os.Remove(installedBinaryPath() + ".old")
}

func samePath(a, b string) bool {
	if resolvedA, err := filepath.EvalSymlinks(a); err == nil {
		a = resolvedA
	}
	if resolvedB, err := filepath.EvalSymlinks(b); err == nil {
		b = resolvedB
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func bootstrapInstallState(version, runtimeSelection string) installState {
	skillsMode := skillsModeManaged
	if os.Getenv("NEWSJACK_INSTALL_SKILLS") == "0" {
		skillsMode = skillsModeExternal
	}
	runtimesRaw := bootstrapRuntimes(runtimeSelection)
	return normalizeInstallState(installState{
		Version:     version,
		Commit:      readTrimmedFile(filepath.Join(managedInstallDir(), "COMMIT")),
		Channel:     "stable",
		Repo:        getenv("NEWSJACK_REPO", defaultRepo),
		InstallURL:  getenv("NEWSJACK_INSTALL_URL", defaultInstallURL),
		SkillsMode:  skillsMode,
		Runtimes:    normalizeRuntimeList(runtimesRaw),
		RuntimesRaw: runtimesRaw,
		InstallMCP:  os.Getenv("NEWSJACK_INSTALL_MCP") != "0",
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func writeInstallStateFile(state installState) error {
	if err := os.MkdirAll(newsjackHome(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(installStatePath(), append(data, '\n'), 0o644)
}

func readTrimmedFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func fetchReleaseBytes(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, rawURL)
	}
	return io.ReadAll(resp.Body)
}
