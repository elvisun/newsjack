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
		if candidates := stringValue(testPayload["candidates"]); candidates == "" || !fileExists(candidates) {
			t.Fatalf("candidates artifact missing: %#v", testPayload)
		}
		if summary := stringValue(testPayload["summary"]); summary == "" || !fileExists(summary) {
			t.Fatalf("summary artifact missing: %#v", testPayload)
		}
		reportTarget := stringValue(testPayload["report_target"])
		if reportTarget == "" || fileExists(reportTarget) {
			t.Fatalf("report_target=%q, want planned non-written run.md path", reportTarget)
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
		if schedule["suggested_minute"] != float64(suggestedScheduleMinute("fixture-coffee")) {
			t.Fatalf("suggested_minute=%v, want %d", schedule["suggested_minute"], suggestedScheduleMinute("fixture-coffee"))
		}
		schedulePath := stringValue(schedule["schedule_path"])
		scheduleBody, readErr := os.ReadFile(schedulePath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(scheduleBody), "crontab") || strings.Contains(string(scheduleBody), "launchd") || strings.Contains(string(scheduleBody), "systemd") {
			t.Fatalf("schedule should not include system scheduler instructions:\n%s", scheduleBody)
		}
		if !strings.Contains(string(scheduleBody), "never minute 0") {
			t.Fatalf("schedule should include cron jitter guidance:\n%s", scheduleBody)
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
		if stringValue(status["latest_report_path"]) != "" {
			t.Fatalf("latest_report_path should be empty until a skill writes run.md: %#v", status)
		}
	})
}

func TestMonitorBriefScaffoldAndCommand(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	profilePath := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(profilePath, []byte(`{"company":"Clearnym","topics":["consumer privacy"],"search_terms":["data broker removal"]}`), 0o644); err != nil {
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
		briefPath := stringValue(initPayload["brief_path"])
		if briefPath == "" || !fileExists(briefPath) {
			t.Fatalf("init should scaffold a brief: %#v", initPayload)
		}
		body, err := os.ReadFile(briefPath)
		if err != nil {
			t.Fatal(err)
		}
		// Scaffold must carry the company name and the inert section headings, and
		// must NOT activate any rule (bodies are comments).
		for _, want := range []string{"# Clearnym — Brief", "## We never pitch", "## How to surface", "source of truth"} {
			if !strings.Contains(string(body), want) {
				t.Fatalf("brief scaffold missing %q:\n%s", want, body)
			}
		}

		// The brief is user-authored policy: re-init with --force must NOT clobber it.
		custom := "# Clearnym — Brief\n\n## We never pitch\n- Federal legislation.\n"
		if err := os.WriteFile(briefPath, []byte(custom), 0o644); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		errBuf.Reset()
		if code := runCLI([]string{"monitor", "init", "--profile", profilePath, "--force"}, &out, &errBuf); code != 0 {
			t.Fatalf("monitor init --force code=%d stderr=%s", code, errBuf.String())
		}
		after, _ := os.ReadFile(briefPath)
		if string(after) != custom {
			t.Fatalf("--force clobbered the user brief:\n%s", after)
		}

		// `monitor brief <slug>` prints the contents.
		out.Reset()
		errBuf.Reset()
		if code := runCLI([]string{"monitor", "brief", "clearnym"}, &out, &errBuf); code != 0 {
			t.Fatalf("monitor brief code=%d stderr=%s", code, errBuf.String())
		}
		if !strings.Contains(out.String(), "Federal legislation.") {
			t.Fatalf("monitor brief should print contents:\n%s", out.String())
		}

		// `monitor brief <slug> --json` reports path + presence.
		out.Reset()
		errBuf.Reset()
		if code := runCLI([]string{"monitor", "brief", "clearnym", "--json"}, &out, &errBuf); code != 0 {
			t.Fatalf("monitor brief --json code=%d stderr=%s", code, errBuf.String())
		}
		var meta map[string]any
		if json.Unmarshal(out.Bytes(), &meta) != nil {
			t.Fatalf("invalid brief JSON: %s", out.String())
		}
		if meta["brief_present"] != true || stringValue(meta["brief_path"]) != briefPath {
			t.Fatalf("brief --json mismatch: %#v", meta)
		}

		// status surfaces the brief as a fact for skills.
		out.Reset()
		errBuf.Reset()
		if code := runCLI([]string{"monitor", "status", "clearnym"}, &out, &errBuf); code != 0 {
			t.Fatalf("monitor status code=%d stderr=%s", code, errBuf.String())
		}
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil {
			t.Fatalf("invalid status JSON: %s", out.String())
		}
		if status["brief_present"] != true || stringValue(status["brief_path"]) != briefPath {
			t.Fatalf("status should surface brief facts: %#v", status)
		}
	})
}

func TestSuggestedScheduleMinute(t *testing.T) {
	slugs := []string{
		"fixture-coffee",
		"medialyst",
		"acme-corp",
		"newsjack",
		"localfalcon",
		"property-saviour",
		"Blue Bottle Coffee",
		"",
	}
	for _, slug := range slugs {
		first := suggestedScheduleMinute(slug)
		second := suggestedScheduleMinute(slug)
		if first != second {
			t.Fatalf("suggestedScheduleMinute(%q)=%d then %d, want deterministic", slug, first, second)
		}
		if first < 1 || first > 59 {
			t.Fatalf("suggestedScheduleMinute(%q)=%d, want [1, 59]", slug, first)
		}
		if first == 0 {
			t.Fatalf("suggestedScheduleMinute(%q)=0, want nonzero", slug)
		}
	}
}

func TestSetupDefaultsToClaudeCode(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"MEDIALYST_API_KEY":       "",
		"X_BEARER_TOKEN":          "",
		"TWITTER_BEARER_TOKEN":    "",
		"PATH":                    t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"setup"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s", code, errBuf.String())
		}
		text := out.String()
		if !strings.Contains(text, "Supported agent runtimes:") {
			t.Fatalf("setup should show runtime choices:\n%s", text)
		}
		if !strings.Contains(text, "this installs Newsjack skills and owns scheduled runs") {
			t.Fatalf("setup should explain the combined runtime choice:\n%s", text)
		}
		if strings.Contains(text, "Which harness should own scheduled runs?") {
			t.Fatalf("setup should not ask a separate scheduler-runtime question:\n%s", text)
		}
		if !strings.Contains(text, "installed skills for Claude Code") {
			t.Fatalf("setup should install Claude Code skills by default:\n%s", text)
		}
		if !strings.Contains(text, "Claude Code is selected for skill installation but is not installed.") {
			t.Fatalf("setup should warn that Claude Code is missing:\n%s", text)
		}
		if !strings.Contains(text, "SCHEDULER RUNTIME MISSING") {
			t.Fatalf("setup should stop before launch when scheduler runtime is missing:\n%s", text)
		}
		if strings.Contains(text, "Command: claude ") {
			t.Fatalf("setup should not print a runnable Claude command when Claude Code is missing:\n%s", text)
		}
		if strings.Contains(text, "Run auto-setup in Claude Code now? [Y/n]:") && strings.Contains(text, "Installing Claude Code") {
			t.Fatalf("noninteractive setup should not launch or install Claude Code:\n%s", text)
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

func TestSetupRecommendsHighestPriorityDetectedHarness(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	fakeBin := t.TempDir()
	for _, name := range []string{"codex", "claude", "openclaw", "hermes"} {
		path := filepath.Join(fakeBin, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"PATH":                    fakeBin,
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"setup", "--json"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup --json code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid setup JSON: %s", out.String())
		}
		if payload["recommended_runtime"] != "hermes" {
			t.Fatalf("recommended_runtime=%v, want hermes", payload["recommended_runtime"])
		}
		if payload["recommended_scheduler"] != "hermes" {
			t.Fatalf("recommended_scheduler=%v, want hermes", payload["recommended_scheduler"])
		}
		if stringValue(payload["agent_command"]) != "hermes chat" {
			t.Fatalf("agent_command=%q, want interactive hermes chat command", stringValue(payload["agent_command"]))
		}
	})
}

func TestSetupInstallsClaudeCodeAfterConfirmation(t *testing.T) {
	repo := repoRootForTest(t)
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
		"NEWSJACK_ROOT":                   repo,
		"NEWSJACK_NO_AUTO_UPDATE":         "1",
		"NEWSJACK_IGNORE_DOTENV":          "1",
		"NEWSJACK_CLAUDE_INSTALL_COMMAND": shellQuote(installer),
		"MEDIALYST_API_KEY":               "",
		"X_BEARER_TOKEN":                  "",
		"TWITTER_BEARER_TOKEN":            "",
		"PATH":                            fakeBin + ":/bin:/usr/bin",
	}, func() {
		var out, errBuf bytes.Buffer
		input := strings.Join([]string{
			"claude",
			"yes",
			"",
			"",
		}, "\n")
		code := runCLIWithIO([]string{"setup", "--skip-credentials", "--no-launch"}, strings.NewReader(input), &out, &errBuf)
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
		if !strings.Contains(text, "Ready to run low-effort auto-setup in Claude Code") {
			t.Fatalf("setup should proceed to Claude Code launch after install:\n%s", text)
		}
	})
}

func TestSetupDoesNotTreatSkillDirectoryAsInstalledRuntime(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"PATH":                    t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		input := strings.Join([]string{
			"claude",
			"n",
			"claude",
		}, "\n")
		code := runCLIWithIO([]string{"setup", "--skip-credentials", "--no-launch"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		text := out.String()
		if !fileExists(filepath.Join(home, ".claude", "skills", "newsjack-monitor-setup", "SKILL.md")) {
			t.Fatalf("setup should still copy skills for a selected missing runtime")
		}
		if !strings.Contains(text, "claude     Claude Code    not detected") {
			t.Fatalf("skills directory should not make Claude Code look detected:\n%s", text)
		}
		if !strings.Contains(text, "to install Claude Code, run:") {
			t.Fatalf("setup should print install help for the missing runtime:\n%s", text)
		}
		if !strings.Contains(text, "SCHEDULER RUNTIME MISSING") {
			t.Fatalf("setup should not continue to launch a missing scheduler runtime:\n%s", text)
		}
	})
}

func TestSetupPreservesSeparateRuntimeFlags(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	fakeBin := t.TempDir()
	hermesPath := filepath.Join(fakeBin, "hermes")
	if err := os.WriteFile(hermesPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"PATH":                    fakeBin + ":/bin:/usr/bin",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLIWithIO([]string{"setup", "--runtime", "codex", "--schedule-runtime", "hermes", "--skip-credentials", "--no-launch"}, strings.NewReader(""), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		text := out.String()
		if !strings.Contains(text, "skills                 Codex") || !strings.Contains(text, "schedule               Hermes") {
			t.Fatalf("setup should summarize separate runtime flags:\n%s", text)
		}
		if !strings.Contains(text, "installed skills for Codex") {
			t.Fatalf("setup should install skills into the requested runtime:\n%s", text)
		}
		if !strings.Contains(text, "Ready to run low-effort auto-setup in Hermes") {
			t.Fatalf("setup should schedule in the requested runtime:\n%s", text)
		}
		if strings.Contains(text, "Which agent should run Newsjack?") || strings.Contains(text, "Which harness should own scheduled runs?") {
			t.Fatalf("setup should not prompt when separate runtime flags are explicit:\n%s", text)
		}
	})
}

func TestSetupStoresOptionalCredentials(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"MEDIALYST_API_KEY":       "",
		"X_BEARER_TOKEN":          "",
		"TWITTER_BEARER_TOKEN":    "",
		"PATH":                    t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		input := strings.Join([]string{
			"manual",
			"value",
			"mlst_test_key_12345",
		}, "\n")
		code := runCLIWithIO([]string{"setup"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		text := out.String()
		if !strings.Contains(text, "Configure X API (Optional)") {
			t.Fatalf("setup should title the X API section plainly:\n%s", text)
		}
		if !strings.Contains(text, "Configure Medialyst API (Optional)") {
			t.Fatalf("setup should title the Medialyst API section plainly:\n%s", text)
		}
		if !strings.Contains(text, "live news search, media database, find journalists") {
			t.Fatalf("setup should explain Medialyst usage plainly:\n%s", text)
		}
		if !strings.Contains(text, "https://medialyst.ai/agents") {
			t.Fatalf("setup should link to Medialyst agent API key page:\n%s", text)
		}
		if strings.Contains(text, "MANUAL SKILL INSTALL") || strings.Contains(text, "MANUAL SCHEDULER") {
			t.Fatalf("manual setup should be one merged section:\n%s", text)
		}
		if !strings.Contains(text, "MANUAL SETUP") || !strings.Contains(text, "copy the skills, then give your agent this prompt") {
			t.Fatalf("manual setup should include one merged handoff section:\n%s", text)
		}
		if !strings.Contains(text, "Prompt:\nUse the newsjack-monitor-setup skill to set up newsjack monitoring for my company.") {
			t.Fatalf("manual setup should print the short agent prompt:\n%s", text)
		}
		envPath := filepath.Join(home, ".newsjack", ".env")
		envBody, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(envBody), `X_BEARER_TOKEN="value"`) {
			t.Fatalf("X token was not saved to newsjack env:\n%s", envBody)
		}
		mode, err := os.Stat(envPath)
		if err != nil {
			t.Fatal(err)
		}
		if mode.Mode().Perm() != 0o600 {
			t.Fatalf("env permissions=%o, want 600", mode.Mode().Perm())
		}
		key, source := loadAPIKey()
		if key != "mlst_test_key_12345" || !strings.HasPrefix(source, "credentials:") {
			t.Fatalf("medialyst key/source=%q/%q", key, source)
		}
	})
}

func TestSetupLaunchesSelectedHarnessWithSetupSkillPrompt(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	fakeBin := t.TempDir()
	capture := filepath.Join(t.TempDir(), "hermes-args.txt")
	hermesPath := filepath.Join(fakeBin, "hermes")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + shellQuote(capture) + "\n"
	if err := os.WriteFile(hermesPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"MEDIALYST_API_KEY":       "",
		"X_BEARER_TOKEN":          "",
		"TWITTER_BEARER_TOKEN":    "",
		"PATH":                    fakeBin + ":/bin:/usr/bin",
	}, func() {
		var out, errBuf bytes.Buffer
		input := strings.Join([]string{
			"hermes",
			"",
			"",
			"yes",
			"",
		}, "\n")
		code := runCLIWithIO([]string{"setup"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		text := out.String()
		if !strings.Contains(text, "Ready to run low-effort auto-setup in Hermes") || !strings.Contains(text, "Command: hermes chat") {
			t.Fatalf("setup should prepare Hermes launch:\n%s", text)
		}
		args, err := os.ReadFile(capture)
		if err != nil {
			t.Fatal(err)
		}
		argText := string(args)
		if strings.TrimSpace(argText) != "chat" {
			t.Fatalf("Hermes should launch the interactive chat REPL, got args:\n%s", argText)
		}
		if strings.Contains(argText, "--query") {
			t.Fatalf("Hermes should not launch in one-shot --query mode:\n%s", argText)
		}
		if !strings.Contains(text, "send this prompt first:") || !strings.Contains(text, "Use the newsjack-monitor-setup skill to set up newsjack monitoring for my company.") {
			t.Fatalf("Hermes setup should print the seed prompt to send:\n%s", text)
		}
	})
}

func TestSetupOpenClawFallsBackToManualPrompt(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	fakeBin := t.TempDir()
	capture := filepath.Join(t.TempDir(), "openclaw-args.txt")
	openclawPath := filepath.Join(fakeBin, "openclaw")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + shellQuote(capture) + "\n"
	if err := os.WriteFile(openclawPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"MEDIALYST_API_KEY":       "",
		"X_BEARER_TOKEN":          "",
		"TWITTER_BEARER_TOKEN":    "",
		"PATH":                    fakeBin + ":/bin:/usr/bin",
	}, func() {
		var out, errBuf bytes.Buffer
		input := strings.Join([]string{
			"openclaw",
			"",
			"",
			"yes",
			"",
		}, "\n")
		code := runCLIWithIO([]string{"setup"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		text := out.String()
		// OpenClaw's agent command is one-shot only, so setup must not auto-launch
		// it (it would stop the moment the agent needs input); it hands off the
		// prompt for the user to run manually instead.
		if !strings.Contains(text, "OpenClaw runs one-shot and can't host an interactive setup session") {
			t.Fatalf("setup should explain OpenClaw has no interactive session:\n%s", text)
		}
		if !strings.Contains(text, "Prompt:\nUse the newsjack-monitor-setup skill to set up newsjack monitoring for my company.") {
			t.Fatalf("setup should print the prompt for OpenClaw:\n%s", text)
		}
		if _, err := os.Stat(capture); err == nil {
			t.Fatalf("OpenClaw should not be executed for interactive setup")
		}
	})
}

func TestSetupPrintsContinuationPromptWhenAgentLaunchFails(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	fakeBin := t.TempDir()
	codexPath := filepath.Join(fakeBin, "codex")
	script := "#!/bin/sh\necho codex failed >&2\nexit 7\n"
	if err := os.WriteFile(codexPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"MEDIALYST_API_KEY":       "",
		"X_BEARER_TOKEN":          "",
		"TWITTER_BEARER_TOKEN":    "",
		"PATH":                    fakeBin + ":/bin:/usr/bin",
	}, func() {
		var out, errBuf bytes.Buffer
		input := strings.Join([]string{
			"codex",
			"",
			"",
			"yes",
			"",
		}, "\n")
		code := runCLIWithIO([]string{"setup"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("setup should handle agent launch failure without a generic installer fallback, code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		text := out.String()
		if !strings.Contains(text, "auto-setup did not complete in Codex") {
			t.Fatalf("setup should explain that auto-setup failed:\n%s", text)
		}
		if !strings.Contains(text, "CONTINUE SETUP") || !strings.Contains(text, "Prompt:\nUse the newsjack-monitor-setup skill to set up newsjack monitoring for my company.") {
			t.Fatalf("setup should print the continuation prompt:\n%s", text)
		}
		if !strings.Contains(errBuf.String(), "codex failed") {
			t.Fatalf("setup should preserve agent stderr:\n%s", errBuf.String())
		}
	})
}
