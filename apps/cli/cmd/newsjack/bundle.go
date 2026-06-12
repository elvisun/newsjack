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
	"path"
	"path/filepath"
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
	base := releaseBaseForUpdate()
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
// re-downloading the release.
func finalizeBootstrap(version, runtimeSelection string, stdout, stderr io.Writer) error {
	if err := installSelfBinary(); err != nil {
		return err
	}
	if err := writeInstallStateFile(bootstrapInstallState(version, runtimeSelection)); err != nil {
		return err
	}
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

// installSelfBinary copies the currently running executable into the
// managed bin directory so future runs and MCP configs use a stable path.
func installSelfBinary() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dest := installedBinaryPath()
	if samePath(exe, dest) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		return err
	}
	staging := dest + ".new"
	if err := os.WriteFile(staging, data, 0o755); err != nil {
		return err
	}
	// Windows cannot rename over an existing file; the destination is not
	// the running executable here, so removing it first is safe.
	if fileExists(dest) {
		if err := os.Remove(dest); err != nil {
			_ = os.Remove(staging)
			return err
		}
	}
	return os.Rename(staging, dest)
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
