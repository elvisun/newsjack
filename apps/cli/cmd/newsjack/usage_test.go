package main

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
)

// hiddenCommands are dispatched on purpose but intentionally not advertised in
// `newsjack help` (internal plumbing, not a user-facing command).
var hiddenCommands = map[string]bool{
	"mcp-bridge": true,
}

// topLevelCaseRe matches the top-level dispatch cases in runCLIWithIO: lines
// indented exactly one tab (nested sub-command switches are indented deeper).
var topLevelCaseRe = regexp.MustCompile(`(?m)^\tcase (".*"):`)
var quotedRe = regexp.MustCompile(`"([^"]+)"`)

// TestHelpListsEveryDispatchedCommand prevents drift between the command
// dispatch switch and the help screen: the exact failure that made a container
// harness believe filter-apply/cluster/origin-apply did not exist and fall back
// to simulating the deterministic gates. Every command the CLI dispatches must
// be discoverable in `newsjack help` (minus flag aliases and hidden plumbing).
func TestHelpListsEveryDispatchedCommand(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	var help bytes.Buffer
	printUsage(&help)
	helpText := help.String()

	seen := map[string]bool{}
	var missing []string
	for _, line := range topLevelCaseRe.FindAllStringSubmatch(string(src), -1) {
		for _, tok := range quotedRe.FindAllStringSubmatch(line[1], -1) {
			cmd := tok[1]
			if strings.HasPrefix(cmd, "-") || hiddenCommands[cmd] || seen[cmd] {
				continue // flag aliases (--help, -h, --version) and hidden plumbing
			}
			seen[cmd] = true
			if !strings.Contains(helpText, cmd) {
				missing = append(missing, cmd)
			}
		}
	}
	if len(seen) == 0 {
		t.Fatal("parsed zero dispatch cases from main.go — regex likely needs updating")
	}
	if len(missing) > 0 {
		t.Fatalf("commands dispatched but missing from `newsjack help`: %v\n"+
			"add them to printUsage (usage.go) or to hiddenCommands if intentionally internal", missing)
	}
}

// TestHelpPipelineCommandsPresent pins the specific pipeline gates that a
// multi-harness run depends on, so they can never silently drop out of help.
func TestHelpPipelineCommandsPresent(t *testing.T) {
	var help bytes.Buffer
	printUsage(&help)
	for _, cmd := range []string{"filter-apply", "cluster", "origin-apply", "run-summary"} {
		if !strings.Contains(help.String(), cmd) {
			t.Errorf("pipeline command %q missing from `newsjack help`", cmd)
		}
	}
}

func TestHelpShowsAPIRecoveryCommands(t *testing.T) {
	var help bytes.Buffer
	printUsage(&help)
	for _, want := range []string{
		"API SETUP",
		"auth set-medialyst",
		"auth set-x",
		"https://medialyst.ai/agents",
		"live news search, media database, find journalists",
		"newsjack mcp-bridge",
		"~/.newsjack/credentials.json",
	} {
		if !strings.Contains(help.String(), want) {
			t.Fatalf("newsjack help missing %q:\n%s", want, help.String())
		}
	}

	help.Reset()
	if !printCommandHelp(&help, "auth") {
		t.Fatal("auth help topic was not handled")
	}
	for _, want := range []string{
		"newsjack auth set --medialyst-key <mlst_...> --x-bearer-token <token>",
		"newsjack auth set-medialyst --key <mlst_...>",
		"newsjack auth set-x --bearer-token <token>",
		"https://medialyst.ai/agents",
		"credentials.json",
		"newsjack mcp-bridge",
	} {
		if !strings.Contains(help.String(), want) {
			t.Fatalf("newsjack help auth missing %q:\n%s", want, help.String())
		}
	}
}

func TestDetectorHelpIsDiscoverable(t *testing.T) {
	var out, err bytes.Buffer
	if code := runCLI([]string{"help", "detector"}, &out, &err); code != 0 {
		t.Fatalf("newsjack help detector exited %d: %s", code, err.String())
	}
	for _, want := range []string{
		"newsjack detector run --help",
		"news_search",
		"x_news",
		"search_terms",
		"retrieval uses these instead of raw topics + competitors",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("newsjack help detector missing %q:\n%s", want, out.String())
		}
	}

	out.Reset()
	err.Reset()
	if code := runCLI([]string{"detector", "--help"}, &out, &err); code != 0 {
		t.Fatalf("newsjack detector --help exited %d: %s", code, err.String())
	}
	if !strings.Contains(out.String(), "newsjack detector run --help") {
		t.Fatalf("newsjack detector --help did not print detector help:\n%s", out.String())
	}
}

func TestDetectorRunHelpExitsZero(t *testing.T) {
	var out, err bytes.Buffer
	if code := runCLI([]string{"detector", "run", "--help"}, &out, &err); code != 0 {
		t.Fatalf("newsjack detector run --help exited %d: %s", code, err.String())
	}
	for _, want := range []string{"Usage of detector run", "-profile", "-topic", "-sources"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("detector run help missing %q:\n%s", want, out.String())
		}
	}
}
