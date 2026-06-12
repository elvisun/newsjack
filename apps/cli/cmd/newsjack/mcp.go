package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func configureMCP(opts installOptions, stdout, stderr io.Writer) error {
	targets := selectedRuntimes(opts.Runtimes)
	var errs []string
	for _, rt := range targets {
		var err error
		switch rt.Key {
		case "codex":
			err = configureCodexMCP(opts.CLI, stdout, stderr)
		case "claude":
			err = configureClaudeMCP(opts.CLI, stdout, stderr)
		case "openclaw":
			err = configureOpenClawMCP(opts.CLI, stdout, stderr)
		case "hermes":
			err = configureHermesMCP(opts.CLI, stdout)
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func configureCodexMCP(cli commandInvocation, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("codex"); err != nil {
		return nil
	}
	args := append([]string{"mcp", "add", "medialyst", "--", cli.Command}, cli.With("mcp-bridge")...)
	cmd := exec.Command("codex", args...)
	cmd.Stdin = nil
	if out, err := cmd.CombinedOutput(); err != nil {
		warn(stderr, "Codex MCP server 'medialyst' may already exist; leaving existing config in place: %s", strings.TrimSpace(string(out)))
		return nil
	}
	logf(stdout, "configured Codex MCP server: medialyst")
	return nil
}

func configureClaudeMCP(cli commandInvocation, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return nil
	}
	payload := map[string]any{"type": "stdio", "command": cli.Command, "args": cli.With("mcp-bridge")}
	data, _ := json.Marshal(payload)
	cmd := exec.Command("claude", "mcp", "add-json", "--scope", "user", "medialyst", string(data))
	cmd.Stdin = nil
	if out, err := cmd.CombinedOutput(); err != nil {
		warn(stderr, "Claude Code MCP server 'medialyst' may already exist; leaving existing config in place: %s", strings.TrimSpace(string(out)))
		return nil
	}
	logf(stdout, "configured Claude Code MCP server: medialyst")
	return nil
}

func configureOpenClawMCP(cli commandInvocation, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("openclaw"); err != nil {
		return nil
	}
	payload := map[string]any{"command": cli.Command, "args": cli.With("mcp-bridge")}
	data, _ := json.Marshal(payload)
	cmd := exec.Command("openclaw", "mcp", "set", "medialyst", string(data))
	cmd.Stdin = nil
	if out, err := cmd.CombinedOutput(); err != nil {
		warn(stderr, "could not configure OpenClaw MCP automatically: %s", strings.TrimSpace(string(out)))
		return nil
	}
	logf(stdout, "configured OpenClaw MCP server: medialyst")
	return nil
}

func configureHermesMCP(cli commandInvocation, stdout io.Writer) error {
	config := getenv("HERMES_CONFIG", filepath.Join(homeDir(), ".hermes", "config.yaml"))
	data, _ := os.ReadFile(config)
	if hermesHasMedialyst(string(data)) {
		logf(stdout, "Hermes MCP server already configured: medialyst")
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(config), 0o755); err != nil {
		return err
	}
	text := string(data)
	block := fmt.Sprintf("  medialyst:\n    command: %q\n    args:\n", cli.Command)
	for _, arg := range cli.With("mcp-bridge") {
		block += fmt.Sprintf("      - %q\n", arg)
	}
	lines := strings.Split(text, "\n")
	inserted := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "mcp_servers:" {
			lines = append(lines[:i+1], append(strings.Split(strings.TrimRight(block, "\n"), "\n"), lines[i+1:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		if strings.TrimSpace(text) != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += "\nmcp_servers:\n" + block
	} else {
		text = strings.Join(lines, "\n")
	}
	if err := os.WriteFile(config, []byte(text), 0o644); err != nil {
		return err
	}
	logf(stdout, "configured Hermes MCP server: medialyst")
	return nil
}

func hermesHasMedialyst(text string) bool {
	inMCP := false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "mcp_servers:") {
			inMCP = true
			continue
		}
		if inMCP && strings.HasPrefix(line, "  medialyst:") {
			return true
		}
		if inMCP && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			inMCP = false
		}
	}
	return false
}

func cmdMCPSetup(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mcp setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	runtimes := fs.String("runtimes", getenv("NEWSJACK_RUNTIMES", "auto"), "Runtime selection")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	opts := installOptions{Runtimes: *runtimes, InstallMCP: true, CLI: newsjackCLIInvocation()}
	if err := configureMCP(opts, stdout, stderr); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func cmdMCPStatus(_ []string, stdout, _ io.Writer) int {
	cli := newsjackCLIInvocation()
	payload := map[string]any{
		"bridge_command": cli.CommandLine("mcp-bridge"),
		"transport":      "native",
		"endpoint":       medialystMCPEndpoint(),
	}
	writeJSON(stdout, payload)
	return 0
}
