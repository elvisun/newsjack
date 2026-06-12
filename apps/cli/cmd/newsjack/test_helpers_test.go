package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRootForTest(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if root, ok := findRepoRoot(cwd); ok {
		return root
	}
	t.Fatal("repo root not found")
	return ""
}

func withTempEnv(t *testing.T, env map[string]string, fn func()) {
	t.Helper()
	if home, ok := env["HOME"]; ok && runtime.GOOS == "windows" {
		if _, explicit := env["USERPROFILE"]; !explicit {
			// os.UserHomeDir resolves USERPROFILE on Windows; keep both in sync
			// so HOME-faking tests isolate the same way on every platform.
			mirrored := map[string]string{"USERPROFILE": home}
			for key, value := range env {
				mirrored[key] = value
			}
			env = mirrored
		}
	}
	old := map[string]string{}
	present := map[string]bool{}
	for key, value := range env {
		old[key], present[key] = os.LookupEnv(key)
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
	defer func() {
		for key := range env {
			if present[key] {
				os.Setenv(key, old[key])
			} else {
				os.Unsetenv(key)
			}
		}
	}()
	fn()
}

// requirePOSIXShell skips tests whose fake agent CLIs are POSIX shell
// scripts. The logic they cover is OS-independent and stays covered on the
// Linux CI leg; Windows runtime-exec coverage comes from the Windows harness
// battery (docs/2026-06-11-windows-support-test-plan.md).
func requirePOSIXShell(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake CLIs are POSIX shell scripts; covered on the Linux CI leg")
	}
}

// writeFakeTool creates an executable no-op command for LookPath-based
// runtime detection tests on every platform.
func writeFakeTool(t *testing.T, dir, name string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		if err := os.WriteFile(filepath.Join(dir, name+".cmd"), []byte("@exit /b 0\r\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// testPATH builds a PATH value from the given directories, keeping the
// system shell directories available on Unix so product code can spawn sh.
func testPATH(dirs ...string) string {
	entries := append([]string{}, dirs...)
	if runtime.GOOS != "windows" {
		entries = append(entries, "/bin", "/usr/bin")
	}
	return strings.Join(entries, string(os.PathListSeparator))
}

// assertOwnerOnlyFile checks 0600 permissions where the OS supports POSIX
// permission bits; Windows reduces modes to the read-only attribute.
func assertOwnerOnlyFile(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("%s permissions=%o, want 600", path, info.Mode().Perm())
	}
}
