package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

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
		opts := installOptions{Source: repo, Runtimes: "codex", CLI: commandInvocation{Command: filepath.Join(home, ".newsjack", "bin", "newsjack")}, Repo: defaultRepo, Ref: defaultRef}
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
