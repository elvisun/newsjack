package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonitorInitTestStatusAndSchedule(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	profilePath := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(profilePath, []byte(`{
  "company": "Fixture Coffee",
  "description": "Specialty coffee company.",
  "topics": ["coffee supply chain"],
  "search_terms": ["coffee supply chain"],
  "feed_urls": ["https://example.com/feed.xml"],
  "x_news": {"enabled": true},
  "x_trends": {"mode": "none", "woeids": [], "locations": []}
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"monitor", "init", "--profile", profilePath}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("monitor init code=%d stderr=%s", code, errBuf.String())
		}
		var initPayload map[string]any
		if json.Unmarshal(out.Bytes(), &initPayload) != nil {
			t.Fatalf("invalid init JSON: %s", out.String())
		}
		if initPayload["slug"] != "fixture-coffee" {
			t.Fatalf("slug=%v, want fixture-coffee", initPayload["slug"])
		}
		if !fileExists(filepath.Join(home, ".newsjack", "monitors", "fixture-coffee", "profile.json")) {
			t.Fatal("profile was not saved")
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"monitor", "test", "fixture-coffee", "--mock", "--limit", "2"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("monitor test code=%d stderr=%s", code, errBuf.String())
		}
		var testPayload map[string]any
		if json.Unmarshal(out.Bytes(), &testPayload) != nil {
			t.Fatalf("invalid test JSON: %s", out.String())
		}
		runMD := stringValue(testPayload["run_markdown"])
		if runMD == "" || !fileExists(runMD) {
			t.Fatalf("run_markdown=%q does not exist", runMD)
		}
		body, readErr := os.ReadFile(runMD)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if !strings.Contains(string(body), "Fixture Coffee Newsjack Brief") {
			t.Fatalf("run.md missing profile heading:\n%s", body)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"monitor", "schedule", "fixture-coffee", "--runtime", "openclaw", "--every", "1h"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("monitor schedule code=%d stderr=%s", code, errBuf.String())
		}
		var schedule map[string]any
		if json.Unmarshal(out.Bytes(), &schedule) != nil {
			t.Fatalf("invalid schedule JSON: %s", out.String())
		}
		if schedule["system_cron"] != false {
			t.Fatalf("system_cron=%v, want false", schedule["system_cron"])
		}
		schedulePath := stringValue(schedule["schedule_path"])
		scheduleBody, readErr := os.ReadFile(schedulePath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(scheduleBody), "crontab") || strings.Contains(string(scheduleBody), "launchd") || strings.Contains(string(scheduleBody), "systemd") {
			t.Fatalf("schedule should not include system scheduler instructions:\n%s", scheduleBody)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"monitor", "status", "fixture-coffee"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("monitor status code=%d stderr=%s", code, errBuf.String())
		}
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil {
			t.Fatalf("invalid status JSON: %s", out.String())
		}
		if status["run_count"] != float64(1) {
			t.Fatalf("run_count=%v, want 1", status["run_count"])
		}
		if stringValue(status["latest_run_markdown"]) == "" {
			t.Fatalf("latest_run_markdown missing: %#v", status)
		}
	})
}

func TestSetupDefaultsToClaudeCode(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"PATH":                    t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"setup"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s", code, errBuf.String())
		}
		text := out.String()
		if !strings.Contains(text, "Recommended runtime: Claude Code") {
			t.Fatalf("setup should recommend Claude Code:\n%s", text)
		}
		if !strings.Contains(text, "Claude Code command: claude ") {
			t.Fatalf("setup should print a Claude Code command:\n%s", text)
		}
		if strings.Contains(text, "Recommended runtime: OpenClaw") {
			t.Fatalf("setup should not prefer OpenClaw by default:\n%s", text)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"setup", "--json"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup --json code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid setup JSON: %s", out.String())
		}
		if payload["recommended_runtime"] != "claude" {
			t.Fatalf("recommended_runtime=%v, want claude", payload["recommended_runtime"])
		}
		if payload["claude_installed"] != false {
			t.Fatalf("claude_installed=%v, want false", payload["claude_installed"])
		}
		if !strings.HasPrefix(stringValue(payload["agent_command"]), "claude ") {
			t.Fatalf("agent_command=%q, want claude command", stringValue(payload["agent_command"]))
		}
	})
}

func TestSetupInstallsClaudeCodeAfterConfirmation(t *testing.T) {
	home := t.TempDir()
	fakeBin := t.TempDir()
	installer := filepath.Join(t.TempDir(), "install-claude.sh")
	claudePath := filepath.Join(fakeBin, "claude")
	script := "#!/bin/sh\n" +
		"cat > " + shellQuote(claudePath) + " <<'SCRIPT'\n" +
		"#!/bin/sh\n" +
		"echo claude-test\n" +
		"SCRIPT\n" +
		"chmod +x " + shellQuote(claudePath) + "\n"
	if err := os.WriteFile(installer, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                            home,
		"NEWSJACK_HOME":                   "",
		"NEWSJACK_NO_AUTO_UPDATE":         "1",
		"NEWSJACK_CLAUDE_INSTALL_COMMAND": shellQuote(installer),
		"PATH":                            fakeBin + ":/bin:/usr/bin",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLIWithIO([]string{"setup"}, strings.NewReader("yes\n"), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		if !fileExists(claudePath) {
			t.Fatalf("Claude installer did not create fake claude at %s", claudePath)
		}
		text := out.String()
		if !strings.Contains(text, "Installing Claude Code") {
			t.Fatalf("setup did not run the approved Claude installer:\n%s", text)
		}
		if !strings.Contains(text, "Recommended runtime: Claude Code") {
			t.Fatalf("setup should still recommend Claude Code after install:\n%s", text)
		}
	})
}
