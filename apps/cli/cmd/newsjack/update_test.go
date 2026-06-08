package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatestReleaseVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/release/manifest.json" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"version":"v0.1.0"}`))
	}))
	defer server.Close()

	withTempEnv(t, map[string]string{
		"NEWSJACK_RELEASE_BASE": server.URL + "/release",
	}, func() {
		got, err := latestReleaseVersion()
		if err != nil {
			t.Fatal(err)
		}
		if got != "v0.1.0" {
			t.Fatalf("latestReleaseVersion() = %q, want v0.1.0", got)
		}
	})
}

func TestReadInstalledVersion(t *testing.T) {
	home := t.TempDir()
	versionPath := filepath.Join(home, ".newsjack", "newsjack", "VERSION")
	if err := os.MkdirAll(filepath.Dir(versionPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(versionPath, []byte("abc123\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withTempEnv(t, map[string]string{"NEWSJACK_HOME": filepath.Join(home, ".newsjack")}, func() {
		if got := readInstalledVersion(); got != "abc123" {
			t.Fatalf("readInstalledVersion() = %q, want abc123", got)
		}
	})
}

func TestAutoUpdateDisabled(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_AUTO_UPDATE": "0"}, func() {
		if !autoUpdateDisabled() {
			t.Fatal("NEWSJACK_AUTO_UPDATE=0 should disable auto-update")
		}
	})
	withTempEnv(t, map[string]string{"NEWSJACK_NO_AUTO_UPDATE": "1"}, func() {
		if !autoUpdateDisabled() {
			t.Fatal("NEWSJACK_NO_AUTO_UPDATE=1 should disable auto-update")
		}
	})
}

func TestNPMDistributionSkipsAutoUpdate(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_DISTRIBUTION": "npm"}, func() {
		if shouldAutoUpdate([]string{"detector"}) {
			t.Fatal("npm distribution should not use GitHub Release auto-update")
		}
	})
}

func TestNPMUpdatePrintsInstructions(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_DISTRIBUTION": "npm"}, func() {
		var out, err bytes.Buffer
		if code := cmdUpdate(nil, &out, &err); code != 0 {
			t.Fatalf("cmdUpdate() code = %d stderr=%s", code, err.String())
		}
		if !strings.Contains(out.String(), "npm i -g newsjack@latest") {
			t.Fatalf("npm update instructions missing from %q", out.String())
		}
	})
}

func TestInstallerEnvPreservesExternalSkillMode(t *testing.T) {
	state := installState{
		Repo:        "elvisun/newsjack",
		SkillsMode:  skillsModeExternal,
		RuntimesRaw: "codex,claude",
		InstallMCP:  false,
	}
	env := installerEnv([]string{"NEWSJACK_INSTALL_SKILLS=1", "OTHER=value"}, state)
	joined := "\n" + strings.Join(env, "\n") + "\n"
	for _, want := range []string{
		"\nNEWSJACK_AUTO_UPDATE_RUNNING=1\n",
		"\nNEWSJACK_INSTALL_SKILLS=0\n",
		"\nNEWSJACK_INSTALL_MCP=0\n",
		"\nNEWSJACK_RUNTIMES=codex,claude\n",
		"\nNEWSJACK_REPO=elvisun/newsjack\n",
		"\nOTHER=value\n",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("env missing %q in %#v", want, env)
		}
	}
}
