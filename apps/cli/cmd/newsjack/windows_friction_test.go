package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Regression tests for the Windows install friction report (2026-06-12).

func TestIsReleaseTag(t *testing.T) {
	cases := map[string]bool{
		"v0.1.10":         true,
		"v0.1.10-rc.1":    true,
		"v1.2.3-beta.2":   true,
		"v0.1.0-dev":      false,
		"v0.1.0-dev+abc1": false,
		"dev":             false,
		"":                false,
		"0.1.10":          false,
	}
	for tag, want := range cases {
		if got := isReleaseTag(tag); got != want {
			t.Errorf("isReleaseTag(%q) = %v, want %v", tag, got, want)
		}
	}
}

func TestReleaseBaseForBootstrapPinsOwnVersion(t *testing.T) {
	saved := version
	t.Cleanup(func() { version = saved })

	withTempEnv(t, map[string]string{
		"HOME":                  t.TempDir(),
		"NEWSJACK_HOME":         "",
		"NEWSJACK_RELEASE_BASE": "",
		"NEWSJACK_VERSION":      "",
		"NEWSJACK_REPO":         "",
	}, func() {
		// A release-tagged binary must fetch its own bundle: `latest` may
		// not carry this platform yet (friction report #5).
		version = "v0.1.10-rc.1"
		if got := releaseBaseForBootstrap(); !strings.HasSuffix(got, "/releases/download/v0.1.10-rc.1") {
			t.Fatalf("release binary should pin its own tag, got %s", got)
		}
		// Dev builds have no published release; fall back to latest.
		version = "v0.1.0-dev"
		if got := releaseBaseForBootstrap(); !strings.HasSuffix(got, "/releases/latest/download") {
			t.Fatalf("dev build should use latest, got %s", got)
		}
	})

	withTempEnv(t, map[string]string{
		"HOME":             t.TempDir(),
		"NEWSJACK_HOME":    "",
		"NEWSJACK_VERSION": "v9.9.9",
		"NEWSJACK_REPO":    "",
	}, func() {
		if got := releaseBaseForBootstrap(); !strings.HasSuffix(got, "/releases/download/v9.9.9") {
			t.Fatalf("NEWSJACK_VERSION should override the pinned tag, got %s", got)
		}
	})
}

func TestFindWindowsClaudeBinary(t *testing.T) {
	appData := t.TempDir()
	home := t.TempDir()

	if got := findWindowsClaudeBinary(appData, home); got != "" {
		t.Fatalf("nothing installed should find nothing, got %q", got)
	}

	versioned := filepath.Join(appData, "Claude", "claude-code", "2.1.170")
	if err := os.MkdirAll(versioned, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versioned, "claude.exe"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := findWindowsClaudeBinary(appData, home); got != filepath.Join(versioned, "claude.exe") {
		t.Fatalf("AppData install not found, got %q", got)
	}

	// The native installer location wins when both exist.
	local := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "claude.exe"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := findWindowsClaudeBinary(appData, home); got != filepath.Join(local, "claude.exe") {
		t.Fatalf(".local/bin install should win, got %q", got)
	}
}

func TestInstallSourceAdoptsPrebuiltBundle(t *testing.T) {
	home := t.TempDir()
	bundleDir := t.TempDir()
	for name, content := range testBundleFiles("v9.9.9-test") {
		path := filepath.Join(bundleDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           "",
		"NEWSJACK_RUNTIMES":       "claude",
		"NEWSJACK_INSTALL_MCP":    "0",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"PATH":                    os.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"install", "--source", bundleDir, "--runtimes", "claude", "--mcp=false"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("install --source code=%d stderr=%s", code, errBuf.String())
		}

		// Friction report #9 and #13: the managed root, managed binary,
		// and install state must all exist after a --source install so
		// doctor is clean and the registered MCP bridge path is real.
		if got := readTrimmedFile(filepath.Join(managedInstallDir(), "VERSION")); got != "v9.9.9-test" {
			t.Fatalf("managed root not adopted, VERSION=%q", got)
		}
		if !fileExists(installedBinaryPath()) {
			t.Fatal("managed binary missing after adoption — MCP bridge command would point at nothing")
		}
		var state installState
		data, err := os.ReadFile(installStatePath())
		if err != nil {
			t.Fatalf("install state missing after adoption: %v", err)
		}
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatal(err)
		}
		if state.Version != "v9.9.9-test" || state.RuntimesRaw != "claude" {
			t.Fatalf("adopted state = %+v", state)
		}
		if !fileExists(filepath.Join(home, ".claude", "skills", "newsjack-detector", "SKILL.md")) {
			t.Fatal("skills not installed from adopted bundle")
		}
		root, err := newsjackRoot()
		if err != nil || root != managedInstallDir() {
			t.Fatalf("doctor's install root should be clean after adoption: %q, %v", root, err)
		}

		// Re-running against the managed root itself must not re-adopt.
		if shouldAdoptBundle(managedInstallDir()) {
			t.Fatal("the managed root must never be adopted into itself")
		}
	})
}

func TestConfigureClaudeMCPRepairsStaleEntry(t *testing.T) {
	requirePOSIXShell(t)
	home := t.TempDir()
	fakeBin := t.TempDir()
	capture := filepath.Join(t.TempDir(), "claude-args.txt")
	marker := filepath.Join(t.TempDir(), "existing-entry")
	if err := os.WriteFile(marker, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Fake claude: add-json fails with "already exists" while the marker
	// is present; remove deletes the marker; every call is recorded.
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> " + shellQuote(capture) + "\n" +
		"case \"$2\" in\n" +
		"add-json) if [ -f " + shellQuote(marker) + " ]; then echo 'MCP server medialyst already exists in user config'; exit 1; fi ;;\n" +
		"remove) rm -f " + shellQuote(marker) + " ;;\n" +
		"esac\n" +
		"exit 0\n"
	if err := os.WriteFile(filepath.Join(fakeBin, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":          home,
		"NEWSJACK_HOME": "",
		"PATH":          testPATH(fakeBin),
	}, func() {
		var out, errBuf bytes.Buffer
		if err := configureClaudeMCP(newsjackCLIInvocation(), &out, &errBuf); err != nil {
			t.Fatal(err)
		}
		calls, err := os.ReadFile(capture)
		if err != nil {
			t.Fatal(err)
		}
		text := string(calls)
		if !strings.Contains(text, "mcp remove --scope user medialyst") {
			t.Fatalf("stale entry should be removed, calls:\n%s", text)
		}
		if strings.Count(text, "add-json") != 2 {
			t.Fatalf("expected add-json retry after remove, calls:\n%s", text)
		}
		if !strings.Contains(out.String(), "refreshed Claude Code MCP server") {
			t.Fatalf("repair should be reported:\n%s", out.String())
		}
		if fileExists(marker) {
			t.Fatal("stale entry marker should be gone")
		}
	})
}

func TestDoctorExitsZeroWhenOnlyOptionalKeysMissing(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":                     t.TempDir(),
		"NEWSJACK_HOME":            "",
		"NEWSJACK_ROOT":            repo,
		"NEWSJACK_NO_AUTO_UPDATE":  "1",
		"NEWSJACK_IGNORE_DOTENV":   "1",
		"MEDIALYST_API_KEY":        "",
		"X_BEARER_TOKEN":           "",
		"TWITTER_BEARER_TOKEN":     "",
		"X_API_BEARER_TOKEN":       "",
		"TWITTER_API_BEARER_TOKEN": "",
		"PATH":                     t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{"doctor"}, &out, &errBuf); code != 0 {
			t.Fatalf("doctor with only optional keys missing must exit 0, got %d", code)
		}
		if code := runCLI([]string{"doctor", "--json"}, &out, &errBuf); code != 0 {
			t.Fatalf("doctor --json with only optional keys missing must exit 0, got %d", code)
		}
	})
}

func TestHelpCoversAllListedCommands(t *testing.T) {
	for _, topic := range []string{"install", "skills", "doctor", "setup", "auth"} {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"help", topic}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("help %s exited %d: %s", topic, code, errBuf.String())
		}
		if strings.Contains(errBuf.String(), "unknown help topic") {
			t.Fatalf("help %s should exist", topic)
		}
	}
}

func TestParseRegQueryValue(t *testing.T) {
	out := "\r\nHKEY_CURRENT_USER\\Environment\r\n    Path    REG_EXPAND_SZ    C:\\Program Files\\Foo;C:\\Users\\jane smith\\bin\r\n\r\n"
	got := parseRegQueryValue(out)
	if got != `C:\Program Files\Foo;C:\Users\jane smith\bin` {
		t.Fatalf("parseRegQueryValue = %q", got)
	}
	if parseRegQueryValue("garbage") != "" {
		t.Fatal("garbage input should parse to empty")
	}
}

func TestPathListContains(t *testing.T) {
	list := `C:\Program Files\Foo;C:\Users\jane\.newsjack\bin\;D:\tools`
	if !pathListContains(list, `c:\users\jane\.newsjack\bin`) {
		t.Fatal("case- and trailing-slash-insensitive match expected")
	}
	if pathListContains(list, `C:\Users\other\.newsjack\bin`) {
		t.Fatal("must not match a different dir")
	}
}
