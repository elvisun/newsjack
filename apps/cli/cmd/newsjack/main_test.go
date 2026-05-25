package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthStatusMissingAndLoginHeaders(t *testing.T) {
	withTempEnv(t, map[string]string{
		"HOME":              t.TempDir(),
		"NEWSJACK_HOME":     "",
		"MEDIALYST_API_KEY": "",
	}, func() {
		var out, err bytes.Buffer
		code := runCLI([]string{"auth", "status"}, &out, &err)
		if code != 1 {
			t.Fatalf("status code=%d, want 1; stderr=%s", code, err.String())
		}
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil || status["configured"] != false {
			t.Fatalf("unexpected status JSON: %s", out.String())
		}

		out.Reset()
		err.Reset()
		testKey := "mlst_" + strings.Repeat("x", 12)
		code = runCLI([]string{"login", "--key", testKey}, &out, &err)
		if code != 0 {
			t.Fatalf("login code=%d stderr=%s", code, err.String())
		}
		info, statErr := os.Stat(credentialsPath())
		if statErr != nil {
			t.Fatalf("credentials not written: %v", statErr)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("credentials mode=%o, want 600", got)
		}

		out.Reset()
		err.Reset()
		code = runCLI([]string{"auth", "headers"}, &out, &err)
		if code != 0 {
			t.Fatalf("headers code=%d stderr=%s", code, err.String())
		}
		if out.String() != "{\"Authorization\":\"Bearer "+testKey+"\"}\n" {
			t.Fatalf("unexpected headers: %q", out.String())
		}
	})
}

func TestDetectorMockRunAcceptsFlagsAfterQuery(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":          t.TempDir(),
		"NEWSJACK_ROOT": repo,
	}, func() {
		var out, err bytes.Buffer
		code := runCLI([]string{"detector", "run", "AI customer support", "--mock", "--include-all-scored", "--emit", "json"}, &out, &err)
		if code != 0 {
			t.Fatalf("detector code=%d stderr=%s", code, err.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid JSON: %s", out.String())
		}
		monitor := valueOrEmptyMap(payload["monitor"])
		if monitor["mock"] != true {
			t.Fatalf("mock=false payload=%s", out.String())
		}
		signals := signalSlice(payload["signals"])
		if len(signals) != 2 {
			t.Fatalf("signals=%d, want 2", len(signals))
		}
		if signals[0]["id"] != "e32ebc6ac34ee9d2" || signals[1]["id"] != "578832fabe7e6a64" {
			t.Fatalf("unexpected signal ids: %#v %#v", signals[0]["id"], signals[1]["id"])
		}
		debug := valueOrEmptyMap(payload["debug"])
		if dropped := anySlice(debug["dropped_signal_ids"]); len(dropped) != 0 {
			t.Fatalf("dropped=%v, want empty", dropped)
		}
	})
}

func TestFilterApplyFixture(t *testing.T) {
	repo := repoRootForTest(t)
	var out, err bytes.Buffer
	code := runCLI([]string{
		"filter-apply",
		"--candidates", filepath.Join(repo, "fixtures/golden/go-port/detector/mock-query.json"),
		"--decisions", filepath.Join(repo, "fixtures/golden/go-port/filter/decisions.json"),
		"--include", "keep",
	}, &out, &err)
	if code != 0 {
		t.Fatalf("filter code=%d stderr=%s", code, err.String())
	}
	var payload map[string]any
	if json.Unmarshal(out.Bytes(), &payload) != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 2 {
		t.Fatalf("selected signals=%d, want 2", len(signals))
	}
	cheap := valueOrEmptyMap(payload["cheap_filter"])
	if cheap["selected_count"].(float64) != 2 || cheap["rejected_count"].(float64) != 0 {
		t.Fatalf("unexpected cheap summary: %#v", cheap)
	}
}

func TestInstallGeneratesInstructionOnlySkills(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	codexSkills := filepath.Join(home, ".agents", "skills")
	withTempEnv(t, map[string]string{
		"HOME":                         home,
		"NEWSJACK_CODEX_SKILLS_DIR":    codexSkills,
		"NEWSJACK_CLAUDE_SKILLS_DIR":   filepath.Join(home, ".claude", "skills"),
		"NEWSJACK_OPENCLAW_SKILLS_DIR": filepath.Join(home, ".openclaw", "skills"),
		"NEWSJACK_HERMES_SKILLS_DIR":   filepath.Join(home, ".hermes", "skills"),
	}, func() {
		opts := installOptions{Source: repo, Runtimes: "codex", CLIPath: filepath.Join(home, ".newsjack", "bin", "newsjack"), Repo: defaultRepo, Ref: defaultRef}
		var out, err bytes.Buffer
		if installErr := installRuntimeSkills(opts, &out, &err); installErr != nil {
			t.Fatalf("install failed: %v stderr=%s", installErr, err.String())
		}
		skillDir := filepath.Join(codexSkills, "newsjack-detector")
		if !fileExists(filepath.Join(skillDir, "SKILL.md")) {
			t.Fatalf("SKILL.md not installed")
		}
		if dirExists(filepath.Join(skillDir, "scripts")) {
			t.Fatalf("scripts directory copied into runtime skill dir")
		}
		if !fileExists(filepath.Join(skillDir, ".newsjack-installed")) {
			t.Fatalf("marker missing")
		}
	})
}

func TestSummarizeRunAcceptsFlagsAfterInput(t *testing.T) {
	repo := repoRootForTest(t)
	dir := t.TempDir()
	summary := filepath.Join(dir, "summary.json")
	runMD := filepath.Join(dir, "run.md")
	var out, err bytes.Buffer
	code := runCLI([]string{
		"summarize-run",
		filepath.Join(repo, "fixtures/golden/go-port/detector/mock-query.json"),
		"--output", summary,
		"--markdown", runMD,
	}, &out, &err)
	if code != 0 {
		t.Fatalf("summarize code=%d stderr=%s", code, err.String())
	}
	if !fileExists(summary) || !fileExists(runMD) {
		t.Fatalf("summary artifacts missing")
	}
}

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
