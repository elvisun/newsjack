package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUpdateInstalledBinaryFromBundleSwap(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":          home,
		"NEWSJACK_HOME": "",
	}, func() {
		bundleBin := filepath.Join(managedInstallDir(), "bin", installedBinaryName())
		if err := os.MkdirAll(filepath.Dir(bundleBin), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(bundleBin, []byte("new binary"), 0o755); err != nil {
			t.Fatal(err)
		}
		dest := installedBinaryPath()
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, []byte("old binary"), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := updateInstalledBinaryFromBundle(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(dest)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "new binary" {
			t.Fatalf("installed binary content = %q", got)
		}
		parked := dest + ".old"
		if runtime.GOOS == "windows" {
			// The swap never overwrites in place; the previous binary is
			// parked so a running exe survives its own update.
			old, err := os.ReadFile(parked)
			if err != nil {
				t.Fatalf("parked binary missing on windows: %v", err)
			}
			if string(old) != "old binary" {
				t.Fatalf("parked binary content = %q", old)
			}
		} else if fileExists(parked) {
			t.Fatal("parked binary should be removed immediately on unix")
		}

		cleanupStaleBinary()
		if fileExists(parked) {
			t.Fatal("cleanupStaleBinary should remove the parked binary")
		}
	})
}

func TestRunNativeUpdateAppliesBundleAndPreservesState(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v2.0.0-test"))
	server := serveRelease(t, payload, "v2.0.0-test")
	home := t.TempDir()

	env := bootstrapEnv(home, server.URL)
	env["NEWSJACK_NATIVE_UPDATE"] = "1"
	delete(env, "NEWSJACK_RUNTIMES")
	withTempEnv(t, env, func() {
		// Simulate an existing v1 install with recorded state.
		if err := os.MkdirAll(filepath.Join(managedInstallDir(), "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(managedInstallDir(), "VERSION"), []byte("v1.0.0\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := writeInstallStateFile(normalizeInstallState(installState{
			Version:     "v1.0.0",
			Commit:      "oldcommit",
			RuntimesRaw: "claude",
			Runtimes:    []string{"claude"},
			SkillsMode:  skillsModeManaged,
			InstallMCP:  false,
		})); err != nil {
			t.Fatal(err)
		}
		dest := installedBinaryPath()
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, []byte("v1 binary"), 0o755); err != nil {
			t.Fatal(err)
		}

		var out, errBuf bytes.Buffer
		if err := runNativeUpdate(&out, &errBuf); err != nil {
			t.Fatalf("runNativeUpdate: %v\nstderr=%s", err, errBuf.String())
		}

		if got := readTrimmedFile(filepath.Join(managedInstallDir(), "VERSION")); got != "v2.0.0-test" {
			t.Fatalf("bundle VERSION=%q after update", got)
		}
		binary, err := os.ReadFile(dest)
		if err != nil {
			t.Fatal(err)
		}
		if string(binary) != "fake binary payload\n" {
			t.Fatalf("installed binary not refreshed from bundle: %q", binary)
		}

		data, err := os.ReadFile(installStatePath())
		if err != nil {
			t.Fatal(err)
		}
		var state installState
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatal(err)
		}
		if state.Version != "v2.0.0-test" || state.Commit != "deadbeef" {
			t.Fatalf("state version/commit = %q/%q", state.Version, state.Commit)
		}
		if state.RuntimesRaw != "claude" || state.InstallMCP != false {
			t.Fatalf("update must preserve recorded runtimes/mcp: %+v", state)
		}
		if !fileExists(filepath.Join(home, ".claude", "skills", "newsjack-detector", "SKILL.md")) {
			t.Fatal("update should refresh runtime skills for the recorded runtimes")
		}
	})
}

func TestRunNativeUpdateExternalSkillsModeLeavesRuntimesAlone(t *testing.T) {
	payload := buildBundleTarGz(t, testBundleFiles("v2.0.0-test"))
	server := serveRelease(t, payload, "v2.0.0-test")
	home := t.TempDir()

	env := bootstrapEnv(home, server.URL)
	env["NEWSJACK_NATIVE_UPDATE"] = "1"
	withTempEnv(t, env, func() {
		if err := writeInstallStateFile(normalizeInstallState(installState{
			Version:    "v1.0.0",
			SkillsMode: skillsModeExternal,
		})); err != nil {
			t.Fatal(err)
		}

		var out, errBuf bytes.Buffer
		if err := runNativeUpdate(&out, &errBuf); err != nil {
			t.Fatalf("runNativeUpdate: %v\nstderr=%s", err, errBuf.String())
		}
		if fileExists(filepath.Join(home, ".claude", "skills", "newsjack-detector", "SKILL.md")) {
			t.Fatal("external skills mode must not write runtime skill dirs")
		}
		var state installState
		data, _ := os.ReadFile(installStatePath())
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatal(err)
		}
		if state.SkillsMode != skillsModeExternal || state.Version != "v2.0.0-test" {
			t.Fatalf("state after external update: %+v", state)
		}
	})
}

func TestUseNativeUpdateRespectsOverride(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_NATIVE_UPDATE": "1"}, func() {
		if !useNativeUpdate() {
			t.Fatal("NEWSJACK_NATIVE_UPDATE=1 should force the native path")
		}
	})
	withTempEnv(t, map[string]string{"NEWSJACK_NATIVE_UPDATE": "0"}, func() {
		if useNativeUpdate() {
			t.Fatal("NEWSJACK_NATIVE_UPDATE=0 should force the hosted path")
		}
	})
	withTempEnv(t, map[string]string{"NEWSJACK_NATIVE_UPDATE": ""}, func() {
		if got, want := useNativeUpdate(), runtime.GOOS == "windows"; got != want {
			t.Fatalf("default native update = %v on %s", got, runtime.GOOS)
		}
	})
}
