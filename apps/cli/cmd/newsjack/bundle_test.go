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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func buildBundleTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		content := files[name]
		header := &tar.Header{Name: "./" + name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testBundleFiles(version string) map[string]string {
	return map[string]string{
		"skills/newsjack-detector/SKILL.md":      "# detector\n",
		"skills/newsjack-monitor-setup/SKILL.md": "# monitor setup\n",
		".newsjack-prebuilt":                     "1\n",
		"VERSION":                                version + "\n",
		"COMMIT":                                 "deadbeef\n",
		"skills-manifest.json":                   "{}\n",
		"bin/" + installedBinaryName():           "fake binary payload\n",
		".env":                                   "SECRET=1\n",
	}
}

func serveRelease(t *testing.T, payload []byte, version string) *httptest.Server {
	t.Helper()
	artifact := releaseArtifactName()
	sum := sha256.Sum256(payload)
	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), artifact)
	})
	mux.HandleFunc("/"+artifact, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	})
	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"version":%q}`, version)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func bootstrapEnv(home, releaseBase string) map[string]string {
	return map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           "",
		"NEWSJACK_RELEASE_BASE":   releaseBase,
		"NEWSJACK_RUNTIMES":       "claude",
		"NEWSJACK_INSTALL_MCP":    "0",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"PATH":                    os.TempDir(),
	}
}

func TestBootstrapInstallFromBareBinary(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v9.9.9-test"))
	server := serveRelease(t, payload, "v9.9.9-test")
	home := t.TempDir()

	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		var out, errBuf bytes.Buffer
		if err := bootstrapInstall("", &out, &errBuf); err != nil {
			t.Fatalf("bootstrapInstall: %v\nstderr=%s", err, errBuf.String())
		}

		install := filepath.Join(home, ".newsjack", "newsjack")
		if got := readTrimmedFile(filepath.Join(install, "VERSION")); got != "v9.9.9-test" {
			t.Fatalf("installed VERSION=%q, want v9.9.9-test", got)
		}
		if !fileExists(filepath.Join(home, ".newsjack", "bin", installedBinaryName())) {
			t.Fatal("self binary was not installed into the managed bin dir")
		}
		if fileExists(filepath.Join(install, ".env")) {
			t.Fatal("bundle .env file should be stripped during unpack")
		}
		if !fileExists(filepath.Join(home, ".claude", "skills", "newsjack-detector", "SKILL.md")) {
			t.Fatal("skills were not installed for the selected runtime")
		}

		data, err := os.ReadFile(filepath.Join(home, ".newsjack", "install.json"))
		if err != nil {
			t.Fatalf("install.json missing: %v", err)
		}
		var state installState
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatalf("install.json invalid: %s", data)
		}
		if state.Version != "v9.9.9-test" || state.Commit != "deadbeef" {
			t.Fatalf("install state version/commit = %q/%q", state.Version, state.Commit)
		}
		if state.SkillsMode != skillsModeManaged || state.RuntimesRaw != "claude" {
			t.Fatalf("install state mode/runtimes = %q/%q", state.SkillsMode, state.RuntimesRaw)
		}

		root, err := newsjackRoot()
		if err != nil || root != install {
			t.Fatalf("newsjackRoot() = %q, %v; want %q", root, err, install)
		}
	})
}

func TestApplyReleaseBundleChecksumMismatchLeavesInstallUntouched(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v9.9.9-test"))
	artifact := releaseArtifactName()
	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", strings.Repeat("0", 64), artifact)
	})
	mux.HandleFunc("/"+artifact, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	home := t.TempDir()
	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		install := managedInstallDir()
		sentinel := filepath.Join(install, "sentinel.txt")
		if err := os.MkdirAll(install, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sentinel, []byte("keep me"), 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := applyReleaseBundle(server.URL, io.Discard); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
			t.Fatalf("expected checksum mismatch error, got %v", err)
		}
		if !fileExists(sentinel) {
			t.Fatal("existing install was modified after a failed download")
		}
		if dirExists(install + ".new") {
			t.Fatal("failed apply left a staging directory behind")
		}
	})
}

func TestApplyReleaseBundleTruncatedArtifactFails(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v9.9.9-test"))
	artifact := releaseArtifactName()
	sum := sha256.Sum256(payload)
	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), artifact)
	})
	mux.HandleFunc("/"+artifact, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload[:len(payload)/2])
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	home := t.TempDir()
	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		if _, err := applyReleaseBundle(server.URL, io.Discard); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
			t.Fatalf("truncated artifact should fail checksum verification, got %v", err)
		}
	})
}

func TestApplyReleaseBundleMissingPlatformArtifact(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, strings.Repeat("0", 64)+"  newsjack_plan9_mips.tar.gz\n")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	home := t.TempDir()
	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		_, err := applyReleaseBundle(server.URL, io.Discard)
		if err == nil || !strings.Contains(err.Error(), releaseArtifactName()) {
			t.Fatalf("missing platform artifact should name the artifact, got %v", err)
		}
	})
}

func TestApplyReleaseBundleRerunKeepsPreviousInstall(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v9.9.9-test"))
	server := serveRelease(t, payload, "v9.9.9-test")
	home := t.TempDir()

	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		if _, err := applyReleaseBundle(server.URL, io.Discard); err != nil {
			t.Fatal(err)
		}
		if _, err := applyReleaseBundle(server.URL, io.Discard); err != nil {
			t.Fatal(err)
		}
		if !dirExists(managedInstallDir() + ".previous") {
			t.Fatal("re-apply should keep the previous install for rollback")
		}
		if got := readTrimmedFile(filepath.Join(managedInstallDir(), "VERSION")); got != "v9.9.9-test" {
			t.Fatalf("VERSION=%q after re-apply", got)
		}
	})
}

func TestSetupJSONBootstrapsWhenRootMissing(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v9.9.9-test"))
	server := serveRelease(t, payload, "v9.9.9-test")
	home := t.TempDir()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	})
	// Leave the repo so newsjackRoot cannot fall back to the source checkout.
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"setup", "--json"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup --json code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("setup --json should still emit valid JSON after bootstrap:\n%s", out.String())
		}
		root, err := newsjackRoot()
		if err != nil {
			t.Fatalf("root missing after setup bootstrap: %v", err)
		}
		if root != filepath.Join(home, ".newsjack", "newsjack") {
			t.Fatalf("root=%q after bootstrap", root)
		}
	})
}

func TestEnsureInstalledRootFinishesInterruptedBootstrap(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v9.9.9-test"))
	server := serveRelease(t, payload, "v9.9.9-test")
	home := t.TempDir()

	withTempEnv(t, bootstrapEnv(home, server.URL), func() {
		// Simulate a bootstrap that died after the bundle swap: the managed
		// root exists, but the binary and install state were never written.
		if _, err := applyReleaseBundle(server.URL, io.Discard); err != nil {
			t.Fatal(err)
		}
		if fileExists(installedBinaryPath()) || fileExists(installStatePath()) {
			t.Fatal("precondition: interrupted install should have no binary or state")
		}

		var out, errBuf bytes.Buffer
		if err := ensureInstalledRoot("", &out, &errBuf); err != nil {
			t.Fatalf("ensureInstalledRoot: %v\nstderr=%s", err, errBuf.String())
		}
		if !fileExists(installedBinaryPath()) {
			t.Fatal("repair should install the managed binary")
		}
		var state installState
		data, err := os.ReadFile(installStatePath())
		if err != nil {
			t.Fatalf("repair should write install state: %v", err)
		}
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatal(err)
		}
		if state.Version != "v9.9.9-test" {
			t.Fatalf("repaired state version = %q", state.Version)
		}
		if !fileExists(filepath.Join(home, ".claude", "skills", "newsjack-detector", "SKILL.md")) {
			t.Fatal("repair should install runtime skills")
		}

		// A second call must be a no-op, not another repair.
		errBuf.Reset()
		if err := ensureInstalledRoot("", &out, &errBuf); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(errBuf.String(), "interrupted") {
			t.Fatalf("completed install should not re-trigger repair:\n%s", errBuf.String())
		}
	})
}

func TestUntarGzRejectsPathTraversal(t *testing.T) {
	payload := buildBundleTarGz(t, map[string]string{"../evil.txt": "nope"})
	dest := filepath.Join(t.TempDir(), "out")
	if err := untarGz(payload, dest); err == nil || !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
	if fileExists(filepath.Join(filepath.Dir(dest), "evil.txt")) {
		t.Fatal("traversal entry escaped the destination directory")
	}
}
