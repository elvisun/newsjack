package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

const (
	version                 = "0.2.0-go"
	defaultRepo             = "elvisun/newsjack"
	defaultRef              = "main"
	defaultMinQueuePriority = 40.0
	defaultMinMajorNews     = 0.55
	medialystMCPURL         = "https://medialyst.ai/api/mcp"
	envMedialystKey         = "MEDIALYST_API_KEY"
)

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	cmd := "help"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	case "version", "--version":
		fmt.Fprintln(stdout, version)
		return 0
	case "path":
		root, err := newsjackRoot()
		if err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintln(stdout, root)
		return 0
	case "skills":
		if len(args) > 1 {
			switch args[1] {
			case "list":
				return cmdSkillsList(args[2:], stdout, stderr)
			case "install":
				return cmdInstall(args[2:], stdout, stderr)
			case "--help", "-h":
				fmt.Fprintln(stdout, "Usage: newsjack skills list|install")
				return 0
			default:
				return failf(stderr, "unknown skills command: %s", args[1])
			}
		}
		return cmdSkillsList(nil, stdout, stderr)
	case "install":
		return cmdInstall(args[1:], stdout, stderr)
	case "update":
		return cmdUpdate(args[1:], stdout, stderr)
	case "doctor":
		return cmdDoctor(args[1:], stdout, stderr)
	case "runtimes":
		if len(args) > 1 && args[1] == "detect" {
			return cmdRuntimesDetect(args[2:], stdout, stderr)
		}
		return fail(stderr, errors.New("usage: newsjack runtimes detect"))
	case "login":
		return cmdLogin(args[1:], stdout, stderr)
	case "auth":
		return cmdAuth(args[1:], stdout, stderr)
	case "mcp":
		if len(args) > 1 {
			switch args[1] {
			case "setup":
				return cmdMCPSetup(args[2:], stdout, stderr)
			case "status":
				return cmdMCPStatus(args[2:], stdout, stderr)
			}
		}
		return fail(stderr, errors.New("usage: newsjack mcp setup|status"))
	case "mcp-bridge":
		return cmdMCPBridge(args[1:], stdout, stderr)
	case "detector":
		return cmdDetector(args[1:], stdout, stderr)
	case "filter-apply":
		return cmdFilterApply(args[1:], stdout, stderr)
	case "summarize-run":
		return cmdSummarizeRun(args[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return failf(stderr, "unknown command: %s", cmd)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `newsjack

Usage:
  newsjack help
  newsjack version
  newsjack path
  newsjack doctor
  newsjack install [--source DIR]
  newsjack skills [list]
  newsjack skills install [--source DIR]
  newsjack runtimes detect
  newsjack login [--key KEY]
  newsjack auth status|headers|logout
  newsjack detector run|diagnose|recent ...
  newsjack filter-apply ...
  newsjack summarize-run ...
  newsjack mcp setup|status
  newsjack mcp-bridge
  newsjack update
`)
}

func fail(w io.Writer, err error) int {
	fmt.Fprintf(w, "newsjack: error: %v\n", err)
	return 1
}

func failf(w io.Writer, format string, args ...any) int {
	return fail(w, fmt.Errorf(format, args...))
}

func warn(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "newsjack: warning: "+format+"\n", args...)
}

func logf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "newsjack: "+format+"\n", args...)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}

func newsjackHome() string {
	if v := os.Getenv("NEWSJACK_HOME"); v != "" {
		return expandPath(v)
	}
	return filepath.Join(homeDir(), ".newsjack")
}

func newsjackRoot() (string, error) {
	if v := os.Getenv("NEWSJACK_ROOT"); v != "" {
		root := expandPath(v)
		if dirExists(filepath.Join(root, "skills")) {
			return root, nil
		}
		return "", fmt.Errorf("NEWSJACK_ROOT does not look like the newsjack repo: %s", root)
	}
	root := filepath.Join(newsjackHome(), "newsjack")
	if dirExists(filepath.Join(root, "skills")) {
		return root, nil
	}
	if cwd, err := os.Getwd(); err == nil {
		if repo, ok := findRepoRoot(cwd); ok {
			return repo, nil
		}
	}
	return "", fmt.Errorf("newsjack is not installed at %s; run: curl newsjack.sh | sh", root)
}

func findRepoRoot(start string) (string, bool) {
	dir := start
	for {
		if fileExists(filepath.Join(dir, "skills", "newsjack-detector", "SKILL.md")) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func expandPath(path string) string {
	if path == "~" {
		return homeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir(), path[2:])
	}
	if path != "" && !filepath.IsAbs(path) {
		if base := os.Getenv("NEWSJACK_WORKDIR"); base != "" {
			return filepath.Join(base, path)
		}
	}
	return path
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func cmdSkillsList(_ []string, stdout, stderr io.Writer) int {
	root, err := newsjackRoot()
	if err != nil {
		return fail(stderr, err)
	}
	names, err := skillNames(root)
	if err != nil {
		return fail(stderr, err)
	}
	for _, name := range names {
		fmt.Fprintln(stdout, name)
	}
	return 0
}

func skillNames(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if fileExists(filepath.Join(root, "skills", entry.Name(), "SKILL.md")) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func cmdUpdate(_ []string, _ io.Writer, stderr io.Writer) int {
	shell := "sh"
	var cmd *exec.Cmd
	if curl, err := exec.LookPath("curl"); err == nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q -fsSL https://newsjack.sh/install.sh | sh", curl))
	} else if wget, err := exec.LookPath("wget"); err == nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q -qO- https://newsjack.sh/install.sh | sh", wget))
	} else {
		return fail(stderr, errors.New("curl or wget is required"))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fail(stderr, err)
	}
	return 0
}

type runtimeTarget struct {
	Key      string
	Label    string
	DirEnv   string
	Default  string
	Binary   string
	HomeHint string
}

var runtimeTargets = []runtimeTarget{
	{"codex", "Codex", "NEWSJACK_CODEX_SKILLS_DIR", filepath.Join(homeDir(), ".agents", "skills"), "codex", filepath.Join(homeDir(), ".agents")},
	{"claude", "Claude Code", "NEWSJACK_CLAUDE_SKILLS_DIR", filepath.Join(homeDir(), ".claude", "skills"), "claude", filepath.Join(homeDir(), ".claude")},
	{"openclaw", "OpenClaw", "NEWSJACK_OPENCLAW_SKILLS_DIR", filepath.Join(homeDir(), ".openclaw", "skills"), "openclaw", filepath.Join(homeDir(), ".openclaw")},
	{"hermes", "Hermes", "NEWSJACK_HERMES_SKILLS_DIR", filepath.Join(homeDir(), ".hermes", "skills"), "hermes", filepath.Join(homeDir(), ".hermes")},
}

func targetDir(rt runtimeTarget) string {
	if v := os.Getenv(rt.DirEnv); v != "" {
		return expandPath(v)
	}
	return rt.Default
}

func normalizeRuntimeList(raw string) []string {
	if raw == "" {
		raw = "auto"
	}
	raw = strings.ToLower(raw)
	raw = strings.ReplaceAll(raw, "claude-code", "claude")
	raw = strings.ReplaceAll(raw, "claude_code", "claude")
	raw = strings.ReplaceAll(raw, "cladue", "claude")
	raw = strings.NewReplacer("\n", ",", "\t", ",", " ", ",").Replace(raw)
	parts := strings.Split(raw, ",")
	var out []string
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	if len(out) == 0 {
		return []string{"auto"}
	}
	return out
}

func runtimeDetected(rt runtimeTarget) bool {
	if _, err := exec.LookPath(rt.Binary); err == nil {
		return true
	}
	return dirExists(rt.HomeHint) || dirExists(filepath.Dir(rt.Default))
}

func selectedRuntimes(raw string) []runtimeTarget {
	list := normalizeRuntimeList(raw)
	for _, item := range list {
		if item == "none" {
			return nil
		}
	}
	has := func(key string) bool {
		for _, item := range list {
			if item == key {
				return true
			}
		}
		return false
	}
	var out []runtimeTarget
	if has("all") {
		return append(out, runtimeTargets...)
	}
	if has("auto") {
		for _, rt := range runtimeTargets {
			if runtimeDetected(rt) {
				out = append(out, rt)
			}
		}
		return out
	}
	for _, rt := range runtimeTargets {
		if has(rt.Key) {
			out = append(out, rt)
		}
	}
	return out
}

func cmdRuntimesDetect(_ []string, stdout, _ io.Writer) int {
	payload := map[string]any{}
	for _, rt := range runtimeTargets {
		payload[rt.Key] = map[string]any{
			"detected":   runtimeDetected(rt),
			"skills_dir": targetDir(rt),
			"binary":     rt.Binary,
		}
	}
	writeJSON(stdout, payload)
	return 0
}

type installOptions struct {
	Source     string
	Runtimes   string
	InstallMCP bool
	Force      bool
	StrictMCP  bool
	CLIPath    string
	Repo       string
	Ref        string
}

func cmdInstall(args []string, stdout, stderr io.Writer) int {
	rootDefault, _ := newsjackRoot()
	if rootDefault == "" {
		rootDefault = os.Getenv("NEWSJACK_INSTALL_DIR")
	}
	if rootDefault == "" {
		rootDefault = filepath.Join(newsjackHome(), "newsjack")
	}
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	source := fs.String("source", rootDefault, "Newsjack source bundle")
	runtimes := fs.String("runtimes", getenv("NEWSJACK_RUNTIMES", "auto"), "Runtime selection: auto, all, none, codex,claude,openclaw,hermes")
	force := fs.Bool("force", os.Getenv("NEWSJACK_FORCE") == "1", "Overwrite Newsjack-managed targets and existing user-owned skills")
	installMCP := fs.Bool("mcp", os.Getenv("NEWSJACK_INSTALL_MCP") != "0", "Configure supported MCP clients")
	strictMCP := fs.Bool("strict-mcp", false, "Fail instead of warning on MCP setup errors")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	opts := installOptions{
		Source:     expandPath(*source),
		Runtimes:   *runtimes,
		InstallMCP: *installMCP,
		Force:      *force,
		StrictMCP:  *strictMCP,
		CLIPath:    filepath.Join(newsjackHome(), "bin", "newsjack"),
		Repo:       getenv("NEWSJACK_REPO", defaultRepo),
		Ref:        getenv("NEWSJACK_REF", defaultRef),
	}
	if err := installRuntimeSkills(opts, stdout, stderr); err != nil {
		return fail(stderr, err)
	}
	if opts.InstallMCP {
		if err := configureMCP(opts, stdout, stderr); err != nil {
			if opts.StrictMCP {
				return fail(stderr, err)
			}
			warn(stderr, "%v", err)
		}
	}
	return 0
}

func installRuntimeSkills(opts installOptions, stdout, stderr io.Writer) error {
	if !dirExists(filepath.Join(opts.Source, "skills")) {
		return fmt.Errorf("missing skills directory in %s", opts.Source)
	}
	targets := selectedRuntimes(opts.Runtimes)
	if len(targets) == 0 && !strings.Contains(","+strings.Join(normalizeRuntimeList(opts.Runtimes), ",")+",", ",none,") {
		warn(stderr, "no supported runtime detected; installing portable skills into %s", filepath.Join(homeDir(), ".agents", "skills"))
		targets = []runtimeTarget{{Key: "portable", Label: "portable Agent Skills", Default: filepath.Join(homeDir(), ".agents", "skills")}}
	}
	for _, rt := range targets {
		if err := installSkillsTo(opts, rt.Label, targetDir(rt)); err != nil {
			return err
		}
		logf(stdout, "installed skills for %s into %s", rt.Label, targetDir(rt))
	}
	return nil
}

func installSkillsTo(opts installOptions, _ string, targetRoot string) error {
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(filepath.Join(opts.Source, "skills"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		src := filepath.Join(opts.Source, "skills", entry.Name())
		if !fileExists(filepath.Join(src, "SKILL.md")) {
			continue
		}
		if err := installOneSkill(opts, src, filepath.Join(targetRoot, entry.Name())); err != nil {
			return err
		}
	}
	return installDoctrineFiles(opts, targetRoot)
}

func installOneSkill(opts installOptions, src, dest string) error {
	marker := filepath.Join(dest, ".newsjack-installed")
	if _, err := os.Stat(dest); err == nil && !fileExists(marker) && !opts.Force {
		return nil
	}
	tmp := dest + ".tmp"
	_ = os.RemoveAll(tmp)
	if err := copySkillTree(src, tmp, opts); err != nil {
		return err
	}
	markerText := fmt.Sprintf("repo=%s\nref=%s\nsource=%s\n", opts.Repo, opts.Ref, opts.Source)
	if err := os.WriteFile(filepath.Join(tmp, ".newsjack-installed"), []byte(markerText), 0o644); err != nil {
		return err
	}
	_ = os.RemoveAll(dest)
	return os.Rename(tmp, dest)
}

func copySkillTree(src, dest string, opts installOptions) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dest, 0o755)
		}
		if d.IsDir() && d.Name() == "scripts" {
			return filepath.SkipDir
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			data = []byte(rewriteSkillText(string(data), opts.CLIPath))
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		mode := info.Mode() & 0o777
		if mode == 0 {
			mode = 0o644
		}
		return os.WriteFile(target, data, mode)
	})
}

func rewriteSkillText(text, cliPath string) string {
	cmd := cliPath
	replacements := []struct{ old, new string }{
		{"python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run", cmd + " detector run"},
		{"python3 skills/newsjack-detector/scripts/newsjack_detector.py run", cmd + " detector run"},
		{"../../skills/newsjack-detector/scripts/newsjack_detector.py run", cmd + " detector run"},
		{"skills/newsjack-detector/scripts/newsjack_detector.py run", cmd + " detector run"},
		{"python3 ../../skills/newsjack-detector/scripts/newsjack_filter_apply.py", cmd + " filter-apply"},
		{"python3 skills/newsjack-detector/scripts/newsjack_filter_apply.py", cmd + " filter-apply"},
		{"../../skills/newsjack-detector/scripts/newsjack_filter_apply.py", cmd + " filter-apply"},
		{"skills/newsjack-detector/scripts/newsjack_filter_apply.py", cmd + " filter-apply"},
		{"python3 \"$SCRIPT_DIR/summarize-run.py\"", cmd + " summarize-run"},
		{"python3 scripts/summarize-run.py", cmd + " summarize-run"},
		{"summarize-run.py", "newsjack summarize-run"},
	}
	for _, repl := range replacements {
		text = strings.ReplaceAll(text, repl.old, repl.new)
	}
	return text
}

func installDoctrineFiles(opts installOptions, targetRoot string) error {
	for _, name := range []string{"ETHICS.md", "WHY-NOT-SPAM.md"} {
		src := filepath.Join(opts.Source, "skills", name)
		if !fileExists(src) {
			continue
		}
		dest := filepath.Join(targetRoot, name)
		marker := filepath.Join(targetRoot, "."+name+".newsjack-installed")
		if fileExists(dest) && !fileExists(marker) && !opts.Force {
			continue
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(marker, []byte(""), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func configureMCP(opts installOptions, stdout, stderr io.Writer) error {
	targets := selectedRuntimes(opts.Runtimes)
	var errs []string
	for _, rt := range targets {
		var err error
		switch rt.Key {
		case "codex":
			err = configureCodexMCP(opts.CLIPath, stdout, stderr)
		case "claude":
			err = configureClaudeMCP(opts.CLIPath, stdout, stderr)
		case "openclaw":
			err = configureOpenClawMCP(opts.CLIPath, stdout, stderr)
		case "hermes":
			err = configureHermesMCP(opts.CLIPath, stdout)
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

func configureCodexMCP(cli string, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("codex"); err != nil {
		return nil
	}
	cmd := exec.Command("codex", "mcp", "add", "medialyst", "--", cli, "mcp-bridge")
	cmd.Stdin = nil
	if out, err := cmd.CombinedOutput(); err != nil {
		warn(stderr, "Codex MCP server 'medialyst' may already exist; leaving existing config in place: %s", strings.TrimSpace(string(out)))
		return nil
	}
	logf(stdout, "configured Codex MCP server: medialyst")
	return nil
}

func configureClaudeMCP(cli string, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return nil
	}
	payload := map[string]any{"type": "stdio", "command": cli, "args": []string{"mcp-bridge"}}
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

func configureOpenClawMCP(cli string, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("openclaw"); err != nil {
		return nil
	}
	payload := map[string]any{"command": cli, "args": []string{"mcp-bridge"}}
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

func configureHermesMCP(cli string, stdout io.Writer) error {
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
	block := fmt.Sprintf("  medialyst:\n    command: %q\n    args:\n      - \"mcp-bridge\"\n", cli)
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
	opts := installOptions{Runtimes: *runtimes, InstallMCP: true, CLIPath: filepath.Join(newsjackHome(), "bin", "newsjack")}
	if err := configureMCP(opts, stdout, stderr); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func cmdMCPStatus(_ []string, stdout, _ io.Writer) int {
	payload := map[string]any{
		"bridge_command": []string{filepath.Join(newsjackHome(), "bin", "newsjack"), "mcp-bridge"},
		"npx_available":  commandAvailable("npx"),
	}
	writeJSON(stdout, payload)
	return 0
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func cmdDoctor(_ []string, stdout, _ io.Writer) int {
	root, rootErr := newsjackRoot()
	key, source := loadAPIKey()
	payload := map[string]any{
		"version":       version,
		"newsjack_home": newsjackHome(),
		"newsjack_root": root,
		"root_ok":       rootErr == nil,
		"auth": map[string]any{
			"medialyst_configured": key != "",
			"source":               nullableString(source),
		},
		"dependencies": map[string]any{
			"npx":  commandAvailable("npx"),
			"xurl": commandAvailable("xurl"),
		},
	}
	payload["runtimes"] = runtimeStatus()
	writeJSON(stdout, payload)
	return 0
}

func runtimeStatus() map[string]any {
	out := map[string]any{}
	for _, rt := range runtimeTargets {
		out[rt.Key] = map[string]any{
			"detected":   runtimeDetected(rt),
			"skills_dir": targetDir(rt),
		}
	}
	return out
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func cmdLogin(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(stderr)
	key := fs.String("key", "", "Medialyst API key")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	apiKey := strings.TrimSpace(*key)
	if apiKey == "" {
		fmt.Fprint(stderr, "Medialyst API key: ")
		var line string
		if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
			return fail(stderr, err)
		}
		apiKey = strings.TrimSpace(line)
	}
	if err := validateAPIKey(apiKey); err != nil {
		return failf(stderr, "invalid key: %v", err)
	}
	path, err := writeCredentials(apiKey)
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Saved Medialyst credentials to %s\n", path)
	fmt.Fprintln(stdout, "MCP-compatible runtimes can now use newsjack mcp-bridge without MEDIALYST_API_KEY exports.")
	return 0
}

func cmdAuth(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, errors.New("usage: newsjack auth status|headers|logout"))
	}
	switch args[0] {
	case "status":
		key, source := loadAPIKey()
		payload := map[string]any{"configured": key != "", "source": nullableString(source)}
		writeJSON(stdout, payload)
		if key == "" {
			return 1
		}
		return 0
	case "headers":
		key, source := loadAPIKey()
		if key == "" {
			fmt.Fprintln(stderr, "Medialyst API key not found. Run: ~/.newsjack/bin/newsjack login")
			return 1
		}
		writeJSONCompact(stdout, map[string]string{"Authorization": "Bearer " + key})
		if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
			fmt.Fprintf(stderr, "Loaded Medialyst API key from %s\n", source)
		}
		return 0
	case "logout":
		path := credentialsPath()
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintln(stdout, "No saved Medialyst credentials found.")
				return 0
			}
			return fail(stderr, err)
		}
		fmt.Fprintf(stdout, "Removed saved Medialyst credentials at %s\n", path)
		return 0
	default:
		return failf(stderr, "unknown auth command: %s", args[0])
	}
}

func credentialsPath() string {
	return filepath.Join(newsjackHome(), "credentials.json")
}

func loadAPIKey() (string, string) {
	if v := strings.TrimSpace(os.Getenv(envMedialystKey)); v != "" {
		return v, "environment:" + envMedialystKey
	}
	path := credentialsPath()
	if data, err := os.ReadFile(path); err == nil {
		var payload map[string]any
		if json.Unmarshal(data, &payload) == nil {
			if med, ok := payload["medialyst"].(map[string]any); ok {
				if v, ok := med["api_key"].(string); ok && strings.TrimSpace(v) != "" {
					return strings.TrimSpace(v), "credentials:" + path
				}
			}
			if v, ok := payload[envMedialystKey].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), "credentials:" + path
			}
		}
	}
	for _, path := range candidateEnvPaths() {
		if v := readDotenvKey(path, envMedialystKey); v != "" {
			return v, "dotenv:" + path
		}
	}
	return "", ""
}

func candidateEnvPaths() []string {
	var out []string
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for {
			out = append(out, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	out = append(out, filepath.Join(newsjackHome(), ".env"))
	return out
}

func readDotenvKey(path, key string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		k, v, _ := strings.Cut(line, "=")
		if strings.TrimSpace(k) != key {
			continue
		}
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if v != "" {
			return v
		}
	}
	return ""
}

func validateAPIKey(key string) error {
	if !strings.HasPrefix(key, "mlst_") {
		return errors.New("Medialyst API keys should start with 'mlst_'")
	}
	if len(key) < 12 {
		return errors.New("Medialyst API key is too short")
	}
	return nil
}

func writeCredentials(apiKey string) (string, error) {
	path := credentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	payload := map[string]any{
		"medialyst": map[string]any{
			"api_key":    apiKey,
			"created_at": time.Now().UTC().Format(time.RFC3339Nano),
			"source":     "newsjack-local-login",
		},
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func cmdMCPBridge(_ []string, _ io.Writer, stderr io.Writer) int {
	key, source := loadAPIKey()
	if key == "" {
		fmt.Fprintln(stderr, "Medialyst API key not found. Run: ~/.newsjack/bin/newsjack login")
		return 1
	}
	npx, err := exec.LookPath("npx")
	if err != nil {
		fmt.Fprintln(stderr, "npx is required for the Medialyst MCP stdio bridge. Install Node.js, then retry.")
		return 1
	}
	if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
		fmt.Fprintf(stderr, "Loaded Medialyst API key from %s\n", source)
	}
	env := os.Environ()
	env = setEnv(env, envMedialystKey, key)
	args := []string{npx, "-y", "mcp-remote", medialystMCPURL, "--header", "Authorization: Bearer ${" + envMedialystKey + "}"}
	if err := syscall.Exec(npx, args, env); err != nil {
		return fail(stderr, err)
	}
	return 127
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

type detectorOptions struct {
	Command                         string
	Query                           []string
	Topics                          []string
	ProfilePath                     string
	Sources                         string
	MajorFeeds                      bool
	FeedURLs                        []string
	FeedFiles                       []string
	FeedOnly                        bool
	NoProfileFeeds                  bool
	NoXNews                         bool
	NoXTrends                       bool
	NoHygieneFilter                 bool
	Depth                           string
	LookbackDays                    int
	MaxAgeHours                     float64
	XNewsMinProfileMatch            float64
	XPostsMinProfileMatch           float64
	ProfileRelevanceMinProfileMatch float64
	MajorNewsMinProfileMatch        float64
	XTrendsMinProfileMatch          float64
	MinQueuePriority                float64
	MinMajorNews                    float64
	LaneCaps                        string
	NewOnly                         bool
	Limit                           int
	Mock                            bool
	Save                            bool
	Store                           string
	MonitorName                     string
	IncludeAllScored                bool
	Emit                            string
}

func cmdDetector(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{"run"}
	}
	commands := map[string]bool{"run": true, "diagnose": true, "recent": true}
	if !commands[args[0]] && args[0] != "-h" && args[0] != "--help" {
		args = append([]string{"run"}, args...)
	}
	switch args[0] {
	case "run":
		opts, code := parseDetectorRun(args[1:], stderr)
		if code != 0 {
			return code
		}
		if err := detectorRun(opts, stdout); err != nil {
			return fail(stderr, err)
		}
		return 0
	case "diagnose":
		fs := flag.NewFlagSet("detector diagnose", flag.ContinueOnError)
		fs.SetOutput(stderr)
		sources := fs.String("sources", "", "Comma-separated sources")
		store := fs.String("store", "", "SQLite store path")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		return detectorDiagnose(*sources, *store, stdout, stderr)
	case "recent":
		fs := flag.NewFlagSet("detector recent", flag.ContinueOnError)
		fs.SetOutput(stderr)
		limit := fs.Int("limit", 10, "Number of recent runs")
		store := fs.String("store", "", "SQLite store path")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		rows, err := recentRuns(*limit, *store)
		if err != nil {
			return fail(stderr, err)
		}
		writeJSON(stdout, rows)
		return 0
	default:
		return fail(stderr, errors.New("usage: newsjack detector run|diagnose|recent"))
	}
}

func parseDetectorRun(args []string, stderr io.Writer) (detectorOptions, int) {
	var topics, feedURLs, feedFiles stringList
	fs := flag.NewFlagSet("detector run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := detectorOptions{
		Command:                         "run",
		Depth:                           "quick",
		LookbackDays:                    1,
		MaxAgeHours:                     24.0,
		XNewsMinProfileMatch:            0.05,
		XPostsMinProfileMatch:           0.08,
		ProfileRelevanceMinProfileMatch: 0.05,
		MajorNewsMinProfileMatch:        0.05,
		XTrendsMinProfileMatch:          0.05,
		MinQueuePriority:                defaultMinQueuePriority,
		MinMajorNews:                    defaultMinMajorNews,
		Limit:                           20,
		Emit:                            "json",
	}
	fs.Var(&topics, "topic", "Topic to monitor. Repeatable.")
	fs.StringVar(&opts.ProfilePath, "profile", "", "Path to monitor profile JSON")
	fs.StringVar(&opts.Sources, "sources", "", "Comma-separated sources")
	fs.BoolVar(&opts.MajorFeeds, "major-feeds", false, "Include default curated major-news RSS feeds")
	fs.Var(&feedURLs, "feed-url", "RSS/Atom feed URL or local XML path. Repeatable.")
	fs.Var(&feedFiles, "feed-file", "File containing feed URLs. Repeatable.")
	fs.BoolVar(&opts.FeedOnly, "feed-only", false, "Skip query/profile searches")
	fs.BoolVar(&opts.NoProfileFeeds, "no-profile-feeds", false, "Do not include feed_urls from profile")
	fs.BoolVar(&opts.NoXNews, "no-x-news", false, "Do not auto-include x_news")
	fs.BoolVar(&opts.NoXTrends, "no-x-trends", false, "Do not include x_trends")
	fs.BoolVar(&opts.NoHygieneFilter, "no-hygiene-filter", false, "Disable deterministic hygiene filter")
	fs.StringVar(&opts.Depth, "depth", "quick", "quick, default, or deep")
	fs.IntVar(&opts.LookbackDays, "lookback-days", 1, "Lookback window in days")
	fs.Float64Var(&opts.MaxAgeHours, "max-age-hours", 24.0, "Drop items older than this many hours; 0 disables")
	fs.Float64Var(&opts.XNewsMinProfileMatch, "x-news-min-profile-match", 0.05, "Demote X News clusters below this profile-match score")
	fs.Float64Var(&opts.XPostsMinProfileMatch, "x-posts-min-profile-match", 0.08, "Demote raw X posts below this profile-match score")
	fs.Float64Var(&opts.ProfileRelevanceMinProfileMatch, "profile-relevance-min-profile-match", 0.05, "Demote profile query results below this score")
	fs.Float64Var(&opts.MajorNewsMinProfileMatch, "major-news-min-profile-match", 0.05, "Demote major-feed-only stories below this score")
	fs.Float64Var(&opts.XTrendsMinProfileMatch, "x-trends-min-profile-match", 0.05, "Demote X trends below this score")
	fs.Float64Var(&opts.MinQueuePriority, "min-queue-priority", defaultMinQueuePriority, "Mechanical queue priority floor")
	fs.Float64Var(&opts.MinMajorNews, "min-major-news", defaultMinMajorNews, "Major-news fallback floor")
	fs.StringVar(&opts.LaneCaps, "lane-caps", "", "Comma-separated per-lane output caps")
	fs.BoolVar(&opts.NewOnly, "new-only", false, "Suppress signals already seen in the monitor store")
	fs.IntVar(&opts.Limit, "limit", 20, "Maximum emitted signals")
	fs.BoolVar(&opts.Mock, "mock", false, "Use deterministic mock sources")
	fs.BoolVar(&opts.Save, "save", false, "Save run and seen URLs")
	fs.StringVar(&opts.Store, "store", "", "Override SQLite store path")
	fs.StringVar(&opts.MonitorName, "monitor-name", "", "Monitor name")
	fs.BoolVar(&opts.IncludeAllScored, "include-all-scored", false, "Include all scored signals in debug output")
	fs.StringVar(&opts.Emit, "emit", "json", "json or brief")
	valueFlags := stringSet([]string{
		"topic", "profile", "sources", "feed-url", "feed-file", "depth", "lookback-days", "max-age-hours",
		"x-news-min-profile-match", "x-posts-min-profile-match", "profile-relevance-min-profile-match",
		"major-news-min-profile-match", "x-trends-min-profile-match", "min-queue-priority", "min-major-news",
		"lane-caps", "limit", "store", "monitor-name", "emit",
	})
	if err := fs.Parse(reorderIntermixedFlags(args, valueFlags)); err != nil {
		return opts, 2
	}
	opts.Topics = topics
	opts.FeedURLs = feedURLs
	opts.FeedFiles = feedFiles
	opts.Query = fs.Args()
	if opts.Depth != "quick" && opts.Depth != "default" && opts.Depth != "deep" {
		fmt.Fprintln(stderr, "invalid --depth")
		return opts, 2
	}
	if opts.Emit != "json" && opts.Emit != "brief" {
		fmt.Fprintln(stderr, "invalid --emit")
		return opts, 2
	}
	return opts, 0
}

type monitorProfile struct {
	Company      string
	Description  string
	Topics       []string
	Competitors  []string
	SearchTerms  []string
	FeedURLs     []string
	XNews        map[string]any
	XTrends      map[string]any
	Spokespeople []string
	ProofAssets  []string
	Standing     []string
	Exclusions   []string
	Raw          map[string]any
}

func defaultProfile() monitorProfile {
	return monitorProfile{
		Topics:       []string{},
		Competitors:  []string{},
		SearchTerms:  []string{},
		FeedURLs:     []string{},
		Spokespeople: []string{},
		ProofAssets:  []string{},
		Standing:     []string{},
		Exclusions:   []string{},
		XNews:        map[string]any{"enabled": true},
		XTrends:      map[string]any{"mode": "none", "woeids": []any{}, "locations": []any{}},
	}
}

func profileFromFile(path string) (monitorProfile, error) {
	data, err := os.ReadFile(expandPath(path))
	if err != nil {
		return monitorProfile{}, err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return monitorProfile{}, err
	}
	return profileFromMap(payload), nil
}

func profileFromMap(payload map[string]any) monitorProfile {
	p := defaultProfile()
	p.Raw = payload
	p.Company = firstString(payload["company"], payload["client"])
	p.Description = stringValue(payload["description"])
	p.Topics = stringListValue(firstValue(payload, "topics", "beats"))
	p.Competitors = stringListValue(payload["competitors"])
	p.SearchTerms = stringListValue(firstValue(payload, "search_terms", "queries"))
	p.FeedURLs = stringListValue(firstValue(payload, "feed_urls", "feeds", "rss_feeds"))
	p.XNews = dictValue(payload["x_news"], map[string]any{"enabled": true})
	p.XTrends = dictValue(payload["x_trends"], map[string]any{"mode": "none", "woeids": []any{}, "locations": []any{}})
	p.Spokespeople = stringListValue(firstValue(payload, "spokespeople", "experts"))
	p.ProofAssets = stringListValue(firstValue(payload, "proof_assets", "proof"))
	p.Standing = stringListValue(firstValue(payload, "standing", "expertise"))
	p.Exclusions = stringListValue(firstValue(payload, "exclusions", "do_not_newsjack"))
	return p
}

func (p monitorProfile) queryTerms() []string {
	if len(p.SearchTerms) > 0 {
		return dedupeStrings(p.SearchTerms)
	}
	var out []string
	out = append(out, p.Topics...)
	for _, term := range p.Competitors {
		out = append(out, competitorQuery(term))
	}
	return dedupeStrings(out)
}

func (p monitorProfile) matchText() string {
	var parts []string
	parts = append(parts, p.Company, p.Description)
	parts = append(parts, p.Topics...)
	parts = append(parts, p.Competitors...)
	parts = append(parts, p.FeedURLs...)
	parts = append(parts, stringListValue(p.XTrends["locations"])...)
	parts = append(parts, p.Spokespeople...)
	parts = append(parts, p.ProofAssets...)
	parts = append(parts, p.Standing...)
	return strings.TrimSpace(strings.Join(parts, " "))
}

func (p monitorProfile) publicDict() map[string]any {
	return map[string]any{
		"company":      nullableString(p.Company),
		"topics":       p.Topics,
		"competitors":  p.Competitors,
		"search_terms": p.SearchTerms,
		"feed_urls":    p.FeedURLs,
		"x_news":       p.XNews,
		"x_trends":     p.XTrends,
		"spokespeople": p.Spokespeople,
		"proof_assets": p.ProofAssets,
		"standing":     p.Standing,
		"exclusions":   p.Exclusions,
	}
}

func firstValue(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func firstString(values ...any) string {
	for _, v := range values {
		if s := stringValue(v); s != "" {
			return s
		}
	}
	return ""
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	return s
}

func dictValue(v any, def map[string]any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		out := map[string]any{}
		for k, v := range m {
			out[k] = v
		}
		return out
	}
	out := map[string]any{}
	for k, v := range def {
		out[k] = v
	}
	return out
}

func stringListValue(v any) []string {
	if v == nil {
		return emptyStrings()
	}
	switch x := v.(type) {
	case string:
		parts := regexp.MustCompile(`[,;\n]`).Split(x, -1)
		var out []string
		for _, part := range parts {
			if s := strings.TrimSpace(part); s != "" {
				out = append(out, s)
			}
		}
		return nonNilStrings(out)
	case []any:
		var out []string
		for _, item := range x {
			if m, ok := item.(map[string]any); ok {
				keys := make([]string, 0, len(m))
				for k := range m {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					if s := stringValue(m[k]); s != "" {
						out = append(out, s)
					}
				}
				continue
			}
			if s := stringValue(item); s != "" {
				out = append(out, s)
			}
		}
		return nonNilStrings(out)
	case []string:
		var out []string
		for _, item := range x {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}
		return nonNilStrings(out)
	default:
		if s := stringValue(v); s != "" {
			return []string{s}
		}
		return emptyStrings()
	}
}

func competitorQuery(value string) string {
	term := strings.TrimSpace(value)
	if term == "" || strings.HasPrefix(term, `"`) || !strings.Contains(term, " ") {
		return term
	}
	return `"` + strings.ReplaceAll(term, `"`, `\"`) + `"`
}

type evidenceItem struct {
	Source      string
	Title       string
	URL         string
	Excerpt     string
	Author      string
	Container   string
	PublishedAt string
	Engagement  map[string]any
	Metadata    map[string]any
}

func evidenceFromMap(m map[string]any) evidenceItem {
	return evidenceItem{
		Source:      firstString(m["source"], "unknown"),
		Title:       strings.TrimSpace(firstString(m["title"], m["excerpt"])),
		URL:         strings.TrimSpace(stringValue(m["url"])),
		Excerpt:     strings.TrimSpace(stringValue(m["excerpt"])),
		Author:      stringValue(m["author"]),
		Container:   stringValue(m["container"]),
		PublishedAt: stringValue(m["published_at"]),
		Engagement:  dictValue(m["engagement"], map[string]any{}),
		Metadata:    dictValue(m["metadata"], map[string]any{}),
	}
}

func (e evidenceItem) text() string {
	return strings.Join(nonEmpty(e.Title, e.Excerpt, e.Author, e.Container), " ")
}

func (e evidenceItem) publicDict() map[string]any {
	payload := map[string]any{
		"source":       e.Source,
		"title":        e.Title,
		"url":          e.URL,
		"author":       nullableString(e.Author),
		"container":    nullableString(e.Container),
		"published_at": nullableString(e.PublishedAt),
		"excerpt":      truncate(e.Excerpt, 500),
		"engagement":   e.Engagement,
	}
	if len(e.Metadata) > 0 {
		payload["metadata"] = e.Metadata
	}
	return payload
}

type signalCluster struct {
	Evidence []evidenceItem
}

func (c signalCluster) title() string {
	for _, item := range c.Evidence {
		if item.Source == "news_search" && item.Title != "" {
			return item.Title
		}
	}
	if len(c.Evidence) > 0 {
		return c.Evidence[0].Title
	}
	return ""
}

func (c signalCluster) text() string {
	var parts []string
	for _, item := range c.Evidence {
		parts = append(parts, item.text())
	}
	return strings.Join(parts, " ")
}

func (c signalCluster) sources() []string {
	set := map[string]bool{}
	for _, item := range c.Evidence {
		set[item.Source] = true
	}
	var out []string
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (c signalCluster) urls() []string {
	var out []string
	for _, item := range c.Evidence {
		if item.URL != "" {
			out = append(out, item.URL)
		}
	}
	return out
}

var stopWords = stringSet([]string{
	"about", "after", "again", "against", "also", "and", "are", "because", "am", "an", "as", "at",
	"been", "being", "but", "can", "could", "did", "does", "doing", "for", "be", "by", "do", "go", "he", "if", "in", "is", "it", "me", "my",
	"no", "of", "on", "or", "so", "to", "up", "us", "we", "from", "had", "has", "have", "her", "here", "him", "his", "how",
	"into", "its", "just", "more", "new", "not", "now", "our", "out", "over", "per", "said", "say", "says", "she", "should", "than", "that",
	"the", "their", "them", "then", "there", "these", "they", "this", "those", "through", "too", "under", "was", "what", "when", "where",
	"which", "while", "who", "why", "will", "with", "would", "you", "your",
})

var tokenRe = regexp.MustCompile(`[a-z0-9][a-z0-9+._-]{1,}`)

func tokens(text string) map[string]bool {
	out := map[string]bool{}
	for _, token := range tokenRe.FindAllString(strings.ToLower(text), -1) {
		if !stopWords[token] {
			out[token] = true
		}
	}
	return out
}

func jaccard(left, right string) float64 {
	lt, rt := tokens(left), tokens(right)
	if len(lt) == 0 || len(rt) == 0 {
		return 0
	}
	inter := 0
	for t := range lt {
		if rt[t] {
			inter++
		}
	}
	union := len(lt)
	for t := range rt {
		if !lt[t] {
			union++
		}
	}
	return float64(inter) / float64(union)
}

func profileMatches(profile monitorProfile, text string) []string {
	lower := strings.ToLower(text)
	var terms []string
	terms = append(terms, profile.Topics...)
	terms = append(terms, profile.Competitors...)
	terms = append(terms, profile.Standing...)
	terms = append(terms, profile.ProofAssets...)
	var matches []string
	seen := map[string]bool{}
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term != "" && strings.Contains(lower, strings.ToLower(term)) && !seen[term] {
			matches = append(matches, term)
			seen[term] = true
			if len(matches) >= 12 {
				break
			}
		}
	}
	return nonNilStrings(matches)
}

func profileMatchScore(profile monitorProfile, text string) float64 {
	context := profile.matchText()
	if context == "" {
		return 0.4
	}
	overlap := jaccard(context, text)
	phraseBonus := math.Min(0.4, 0.08*float64(len(profileMatches(profile, text))))
	return math.Min(1.0, overlap+phraseBonus)
}

func parseTime(value string) (time.Time, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, false
	}
	if strings.HasSuffix(raw, "Z") {
		raw = strings.TrimSuffix(raw, "Z") + "+00:00"
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), true
		}
	}
	if len(raw) >= 10 {
		if parsed, err := time.Parse("2006-01-02", raw[:10]); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func minAgeHours(cluster signalCluster, now time.Time) *float64 {
	var ages []float64
	for _, item := range cluster.Evidence {
		if parsed, ok := parseTime(item.PublishedAt); ok {
			age := now.Sub(parsed).Hours()
			if age < 0 {
				age = 0
			}
			ages = append(ages, age)
		}
	}
	if len(ages) == 0 {
		return nil
	}
	sort.Float64s(ages)
	return &ages[0]
}

func freshnessScore(age *float64, lookbackDays int) float64 {
	if age == nil {
		return 0.35
	}
	if *age <= 4 {
		return 1.0
	}
	if *age <= 24 {
		return 0.86
	}
	window := math.Max(24, float64(lookbackDays*24))
	return math.Max(0.05, 1.0-(*age/window))
}

func decayBucket(age *float64) string {
	if age == nil {
		return "unknown"
	}
	switch {
	case *age <= 1:
		return "30min"
	case *age <= 4:
		return "4hr"
	case *age <= 24:
		return "24hr"
	case *age <= 168:
		return "week"
	default:
		return "month"
	}
}

func filterItemsByAge(items []evidenceItem, now time.Time, maxAgeHours float64) []evidenceItem {
	if maxAgeHours <= 0 {
		return items
	}
	var out []evidenceItem
	for _, item := range items {
		parsed, ok := parseTime(item.PublishedAt)
		if !ok || math.Max(0, now.Sub(parsed).Hours()) <= maxAgeHours {
			out = append(out, item)
		}
	}
	return out
}

var socialSources = stringSet([]string{"x", "x_news", "x_trends", "reddit", "hackernews"})
var docsHostPrefixes = []string{"docs.", "doc.", "help.", "support.", "developer.", "developers."}
var docsPathParts = stringSet([]string{"docs", "documentation", "help", "support", "kb", "knowledge-base", "api", "api-reference", "reference", "manual", "guide", "guides", "tutorial", "tutorials", "connector", "connectors", "integration", "integrations"})
var productPathParts = stringSet([]string{"product", "products", "shop", "store", "cart", "checkout", "pricing", "plans", "marketplace", "app-store", "apps"})
var seoPathPatterns = mustRegexes([]string{`\bsell[-_]house[-_]fast\b`, `\bsell[-_]my[-_]house[-_]fast\b`, `\bcash[-_]house[-_]buyers?\b`, `\bwe[-_]buy[-_]houses?\b`, `\bquick[-_]house[-_]sale\b`, `\bbest[-_][a-z0-9-]+`})
var seoTitlePatterns = mustRegexes([]string{`\bbest\s+\d+\b`, `\bbest\s+[a-z0-9\s-]+\s+(tools|services|companies|platforms)\b`, `\bhow to sell (your )?house fast\b`, `\bcash house buyers?\b`})

func hygieneRejectionReason(item evidenceItem) string {
	if socialSources[item.Source] {
		return ""
	}
	parsed, _ := url.Parse(item.URL)
	host := strings.ToLower(parsed.Hostname())
	var parts []string
	for _, part := range strings.Split(strings.ToLower(parsed.Path), "/") {
		if strings.TrimSpace(part) != "" {
			parts = append(parts, strings.TrimSpace(part))
		}
	}
	pathText := strings.ToLower(parsed.Path)
	titleText := strings.ToLower(item.Title)
	combined := strings.Join([]string{titleText, strings.ToLower(item.Container), strings.ToLower(item.Excerpt)}, " ")
	for _, prefix := range docsHostPrefixes {
		if strings.HasPrefix(host, prefix) || strings.Contains(host, "readthedocs.io") {
			return "owned_docs_or_help"
		}
	}
	for _, part := range parts {
		if docsPathParts[part] {
			return "owned_docs_or_help"
		}
	}
	if regexp.MustCompile(`\b(documentation|api reference|developer docs|help center|support article)\b`).MatchString(combined) {
		return "owned_docs_or_help"
	}
	for _, part := range parts {
		if productPathParts[part] {
			return "product_or_ecommerce_page"
		}
	}
	if regexp.MustCompile(`\b(add to cart|buy now|pricing plans|product page|shop now)\b`).MatchString(combined) {
		return "product_or_ecommerce_page"
	}
	for _, re := range seoPathPatterns {
		if re.MatchString(pathText) {
			return "seo_landing_page"
		}
	}
	for _, re := range seoTitlePatterns {
		if re.MatchString(titleText) {
			return "seo_landing_page"
		}
	}
	return ""
}

func filterItemsByHygiene(items []evidenceItem, enabled bool) ([]evidenceItem, map[string]int) {
	if !enabled {
		return items, map[string]int{}
	}
	var out []evidenceItem
	counts := map[string]int{}
	for _, item := range items {
		if reason := hygieneRejectionReason(item); reason != "" {
			counts[reason]++
			continue
		}
		out = append(out, item)
	}
	return out, counts
}

var sourceQuality = map[string]float64{
	"major_feed":  0.88,
	"news_search": 0.95,
	"x":           0.70,
	"x_news":      0.86,
	"x_trends":    0.68,
	"reddit":      0.62,
	"hackernews":  0.72,
}

func sourceAgreementScore(sources []string) float64 {
	if len(sources) >= 3 {
		return 1.0
	}
	if len(sources) == 2 {
		return 0.78
	}
	if len(sources) == 1 && sources[0] == "major_feed" {
		return 0.62
	}
	if len(sources) == 1 && sources[0] == "news_search" {
		return 0.55
	}
	return 0.32
}

func sourceQualityScore(cluster signalCluster) float64 {
	if len(cluster.Evidence) == 0 {
		return 0
	}
	sum := 0.0
	for _, item := range cluster.Evidence {
		if q, ok := sourceQuality[item.Source]; ok {
			sum += q
		} else {
			sum += 0.5
		}
	}
	return sum / float64(len(cluster.Evidence))
}

var engagementFields = []string{"score", "num_comments", "comments", "likes", "reposts", "replies", "quotes", "bookmarks", "views", "points"}

func engagementScore(cluster signalCluster) float64 {
	total := 0.0
	for _, item := range cluster.Evidence {
		for _, field := range engagementFields {
			value := floatValue(item.Engagement[field])
			if value > 0 {
				total += math.Log1p(value)
			}
		}
	}
	return math.Min(1.0, total/24.0)
}

func noveltyScore(urls []string, seen map[string]map[string]any) float64 {
	if len(urls) == 0 {
		return 0.50
	}
	unseen := 0
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			unseen++
		}
	}
	if unseen == 0 {
		return 0.10
	}
	return float64(unseen) / float64(len(urls))
}

var majorNewsTerms = []string{"acquire", "acquired", "acquisition", "antitrust", "ban", "billion", "breach", "contract", "deal", "funding", "hack", "investigation", "ipo", "lawsuit", "launch", "launched", "layoffs", "merger", "outage", "probe", "regulation", "regulator", "ruling", "sec", "settlement", "shutdown", "sues", "valuation"}
var majorEntityTerms = []string{"amazon", "anthropic", "apple", "doj", "ftc", "google", "meta", "microsoft", "nvidia", "openai", "pentagon", "salesforce", "sec", "spacex", "tesla", "white house", "xai"}

func majorNewsScore(cluster signalCluster, age *float64) float64 {
	var feedItems []evidenceItem
	for _, item := range cluster.Evidence {
		if item.Source == "major_feed" {
			feedItems = append(feedItems, item)
		}
	}
	if len(feedItems) == 0 {
		return 0
	}
	best := 999
	for _, item := range feedItems {
		pos := intValue(item.Metadata["feed_position"], 999)
		if pos < best {
			best = pos
		}
	}
	positionScore := 0.42
	switch {
	case best <= 3:
		positionScore = 1.0
	case best <= 10:
		positionScore = 0.82
	case best <= 25:
		positionScore = 0.64
	}
	text := strings.ToLower(cluster.text())
	toks := tokens(text)
	stakeHits := 0
	for _, term := range majorNewsTerms {
		if termMatches(term, text, toks) {
			stakeHits++
		}
	}
	entityHits := 0
	for _, term := range majorEntityTerms {
		if termMatches(term, text, toks) {
			entityHits++
		}
	}
	stakeScore := math.Min(1.0, 0.22*float64(stakeHits))
	entityScore := math.Min(1.0, 0.18*float64(entityHits))
	freshness := freshnessScore(age, 7)
	return math.Min(1.0, 0.44*positionScore+0.24*freshness+0.20*stakeScore+0.12*entityScore)
}

func termMatches(term, text string, toks map[string]bool) bool {
	if strings.Contains(term, " ") {
		return strings.Contains(text, term)
	}
	return toks[term]
}

func signalLane(cluster signalCluster, majorNews, profileMatch float64, opts detectorOptions) string {
	sources := stringSet(cluster.sources())
	if sources["x_news"] {
		if profileMatch < opts.XNewsMinProfileMatch {
			return "x_news_unmatched"
		}
		return "x_news"
	}
	if sources["x_trends"] {
		if profileMatch < opts.XTrendsMinProfileMatch {
			return "x_trends_unmatched"
		}
		return "x_trends"
	}
	if len(sources) == 1 && sources["x"] {
		for _, item := range cluster.Evidence {
			if item.Metadata["x_signal_type"] == "query_trend" {
				if profileMatch < opts.XTrendsMinProfileMatch {
					return "x_trends_unmatched"
				}
				return "x_trends"
			}
		}
		if profileMatch < opts.XPostsMinProfileMatch {
			return "x_posts_weak"
		}
		return "x_posts"
	}
	if majorNews > 0 {
		if profileMatch < opts.MajorNewsMinProfileMatch {
			return "major_news_unmatched"
		}
		return "major_news"
	}
	if profileMatch < opts.ProfileRelevanceMinProfileMatch {
		return "profile_relevance_weak"
	}
	return "profile_relevance"
}

func scoreSignal(cluster signalCluster, profile monitorProfile, seen map[string]map[string]any, now time.Time, opts detectorOptions) map[string]any {
	text := cluster.text()
	sources := cluster.sources()
	urls := cluster.urls()
	age := minAgeHours(cluster, now)
	freshness := freshnessScore(age, opts.LookbackDays)
	novelty := noveltyScore(urls, seen)
	sourceAgreement := sourceAgreementScore(sources)
	sourceQuality := sourceQualityScore(cluster)
	engagement := engagementScore(cluster)
	profileMatch := profileMatchScore(profile, text)
	majorNews := majorNewsScore(cluster, age)
	lane := signalLane(cluster, majorNews, profileMatch, opts)
	queue := 0.0
	switch lane {
	case "major_news":
		queue = round1(100 * (0.30*majorNews + 0.22*freshness + 0.14*novelty + 0.12*profileMatch + 0.12*sourceQuality + 0.10*sourceAgreement))
	case "major_news_unmatched":
		queue = round1(math.Min(39.9, 100*(0.18*majorNews+0.16*freshness+0.12*novelty+0.10*sourceQuality+0.08*sourceAgreement)))
	case "x_news", "x_trends":
		queue = round1(100 * (0.24*freshness + 0.20*profileMatch + 0.18*novelty + 0.16*sourceQuality + 0.12*sourceAgreement + 0.10*engagement))
	case "x_trends_unmatched":
		queue = round1(math.Min(39.9, 100*(0.16*freshness+0.14*novelty+0.12*sourceQuality+0.10*engagement)))
	case "x_news_unmatched", "profile_relevance_weak", "x_posts_weak":
		queue = round1(math.Min(44.0, 100*(0.18*freshness+0.16*novelty+0.12*sourceQuality+0.10*sourceAgreement+0.08*engagement)))
	case "x_posts":
		queue = round1(math.Min(64.0, 100*(0.22*freshness+0.20*engagement+0.18*profileMatch+0.16*novelty+0.14*sourceQuality+0.10*sourceAgreement)))
	default:
		queue = round1(100 * (0.24*freshness + 0.20*sourceAgreement + 0.18*novelty + 0.16*profileMatch + 0.12*sourceQuality + 0.10*engagement))
	}
	var evidence []map[string]any
	for i, item := range cluster.Evidence {
		if i >= 8 {
			break
		}
		evidence = append(evidence, item.publicDict())
	}
	features := map[string]any{
		"decay_bucket":    decayBucket(age),
		"source_count":    len(sources),
		"evidence_count":  len(cluster.Evidence),
		"seen_before":     len(urls) > 0 && allSeen(urls, seen),
		"seen_urls":       seenSubset(urls, seen),
		"profile_matches": profileMatches(profile, text),
		"safety_flags":    safetyFlags(text, profile.Exclusions),
	}
	if age != nil {
		features["age_hours"] = roundN(*age, 2)
	}
	return map[string]any{
		"id":       signalID(cluster.title(), urls, text),
		"title":    cluster.title(),
		"sources":  sources,
		"evidence": evidence,
		"features": features,
		"routing": map[string]any{
			"lane":           lane,
			"queue_priority": queue,
			"demoted":        strings.HasSuffix(lane, "_unmatched") || strings.HasSuffix(lane, "_weak"),
		},
		"mechanical_scores": map[string]any{
			"freshness":        roundN(freshness, 3),
			"source_agreement": roundN(sourceAgreement, 3),
			"novelty":          roundN(novelty, 3),
			"profile_match":    roundN(profileMatch, 3),
			"source_quality":   roundN(sourceQuality, 3),
			"momentum":         roundN(engagement, 3),
			"major_news":       roundN(majorNews, 3),
		},
	}
}

func signalID(title string, urls []string, text string) string {
	basisParts := append([]string{title}, firstN(urls, 5)...)
	basis := strings.Join(basisParts, "|")
	if strings.Trim(basis, "|") == "" {
		basis = truncate(text, 400)
	}
	sum := sha1.Sum([]byte(basis))
	return hex.EncodeToString(sum[:])[:16]
}

func allSeen(urls []string, seen map[string]map[string]any) bool {
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			return false
		}
	}
	return true
}

func seenSubset(urls []string, seen map[string]map[string]any) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, u := range urls {
		if v, ok := seen[u]; ok {
			out[u] = v
		}
	}
	return out
}

var hardSafetyTerms = []string{"humanitarian crisis", "sexual violence", "missing child", "missing person", "mass shooting", "terror attack", "child abuse", "hate crime", "war crime", "earthquake", "genocide", "hostage", "assault", "bombing", "murder", "abuse", "rape"}

func safetyFlags(text string, exclusions []string) []map[string]string {
	lower := strings.ToLower(text)
	var flags []map[string]string
	for _, term := range hardSafetyTerms {
		if strings.Contains(lower, term) {
			flags = append(flags, map[string]string{"type": "hard_safety_term", "term": term, "note": "Review against tragedy and human-suffering newsjacking rules."})
		}
	}
	for _, term := range exclusions {
		if term != "" && strings.Contains(lower, strings.ToLower(term)) {
			flags = append(flags, map[string]string{"type": "profile_exclusion", "term": term, "note": "Matched a monitor-profile exclusion."})
		}
	}
	if flags == nil {
		return []map[string]string{}
	}
	return flags
}

func clusterItems(items []evidenceItem) []signalCluster {
	sort.SliceStable(items, func(i, j int) bool {
		ki, kj := 1, 1
		if items[i].Source == "news_search" {
			ki = 0
		}
		if items[j].Source == "news_search" {
			kj = 0
		}
		if ki != kj {
			return ki < kj
		}
		return items[i].PublishedAt < items[j].PublishedAt
	})
	var clusters []signalCluster
	for _, item := range items {
		if item.Title == "" && item.Excerpt == "" {
			continue
		}
		placed := false
		for i := range clusters {
			if item.URL != "" && contains(clusters[i].urls(), item.URL) {
				clusters[i].Evidence = append(clusters[i].Evidence, item)
				placed = true
				break
			}
			if jaccard(item.text(), clusters[i].text()) >= 0.32 {
				clusters[i].Evidence = append(clusters[i].Evidence, item)
				placed = true
				break
			}
		}
		if !placed {
			clusters = append(clusters, signalCluster{Evidence: []evidenceItem{item}})
		}
	}
	return clusters
}

func detectorRun(opts detectorOptions, stdout io.Writer) error {
	profile := defaultProfile()
	var err error
	if opts.ProfilePath != "" {
		profile, err = profileFromFile(opts.ProfilePath)
		if err != nil {
			return err
		}
	}
	queries := []string{}
	if !opts.FeedOnly {
		queries = buildQueries(opts, profile)
	}
	feedURLs, err := buildFeedURLs(opts, profile)
	if err != nil {
		return err
	}
	config := configFromEnv()
	requestedSources := []string{}
	if !opts.FeedOnly {
		requestedSources, err = requestedSourcesFor(opts, profile)
		if err != nil {
			return err
		}
	}
	trendRequested := contains(requestedSources, "x_trends") && !opts.FeedOnly
	if len(queries) == 0 && len(feedURLs) == 0 && !trendRequested {
		return errors.New("Provide a query, --topic, --major-feeds, --feed-url, or --profile with topics/competitors.")
	}
	querySources := querySources(requestedSources)
	sources := []string{}
	if len(queries) > 0 {
		if opts.Mock {
			sources = querySources
		} else {
			sources = availableSources(config, querySources)
		}
		if len(sources) == 0 {
			return errors.New("No requested sources are available. Configure MEDIALYST_API_KEY and xurl auth, or rerun with --mock.")
		}
	}
	now := time.Now().UTC()
	var allSignals []map[string]any
	var seenURLsToMark []string
	sourceErrors := map[string]any{}
	evidenceSourceCounts := map[string]int{}
	hygieneRejections := map[string]int{}
	noteItems := func(items []evidenceItem) {
		for _, item := range items {
			evidenceSourceCounts[item.Source]++
		}
	}
	processItems := func(query string, items []evidenceItem, errors map[string]string) error {
		items = filterItemsByAge(items, now, opts.MaxAgeHours)
		var rejected map[string]int
		items, rejected = filterItemsByHygiene(items, !opts.NoHygieneFilter)
		mergeCounts(hygieneRejections, rejected)
		noteItems(items)
		if len(errors) > 0 {
			sourceErrors[query] = errors
		}
		clusters := clusterItems(items)
		var urls []string
		for _, cluster := range clusters {
			urls = append(urls, cluster.urls()...)
		}
		seenURLsToMark = append(seenURLsToMark, urls...)
		seen, err := seenStatus(urls, opts.Store)
		if err != nil {
			return err
		}
		for _, cluster := range clusters {
			signal := scoreSignal(cluster, profile, seen, now, opts)
			signal["query"] = query
			if !(opts.NewOnly && signalIsSeen(signal)) {
				allSignals = append(allSignals, signal)
			}
		}
		return nil
	}
	for _, query := range queries {
		items, errors := collectQuery(query, sources, config, opts, now)
		if err := processItems(query, items, errors); err != nil {
			return err
		}
	}
	if len(feedURLs) > 0 {
		items, errors := collectFeeds(feedURLs, opts.Depth, opts.Mock, now)
		if err := processItems("major_news_feed", items, errors); err != nil {
			return err
		}
	}
	if trendRequested {
		items, errors := collectXTrends(profile, config, opts, now)
		if err := processItems("x_trends", items, errors); err != nil {
			return err
		}
	}
	laneCaps := parseLaneCaps(opts.LaneCaps)
	signals := selectSignals(allSignals, opts.Limit, laneCaps, opts.MinQueuePriority, opts.MinMajorNews)
	var runID any
	runID = nil
	if opts.Save {
		id, err := recordRun(opts.MonitorName, profile.publicDict(), queries, signals, seenURLsToMark, opts.Store)
		if err != nil {
			return err
		}
		runID = id
	}
	payload := map[string]any{
		"monitor": map[string]any{
			"name":              nullableString(opts.MonitorName),
			"generated_at":      now.Format(time.RFC3339Nano),
			"profile":           profile.publicDict(),
			"queries":           nonNilStrings(queries),
			"feed_urls":         nonNilStrings(feedURLs),
			"sources_requested": requestedSources,
			"sources_used":      append(append([]string{}, sources...), append(boolSlice(trendRequested, "x_trends"), boolSlice(len(feedURLs) > 0, "major_feed")...)...),
			"lookback_days":     opts.LookbackDays,
			"max_age_hours":     opts.MaxAgeHours,
			"new_only":          opts.NewOnly,
			"depth":             opts.Depth,
			"mock":              opts.Mock,
		},
		"signals": signals,
		"diagnostics": map[string]any{
			"evidence_by_source":    evidenceSourceCounts,
			"hygiene_rejections":    hygieneRejections,
			"signals_by_lane":       countByLanes(allSignals),
			"emitted_by_lane":       countByLanes(signals),
			"lane_caps":             laneCaps,
			"selection":             map[string]any{"mode": map[bool]string{true: "lane_caps", false: "mechanical_floor"}[laneCaps != nil], "limit": opts.Limit, "min_queue_priority": opts.MinQueuePriority, "min_major_news": opts.MinMajorNews},
			"total_scored_signals":  len(allSignals),
			"total_emitted_signals": len(signals),
		},
		"source_errors": sourceErrors,
		"store": map[string]any{
			"saved":  opts.Save,
			"run_id": runID,
			"path":   storePathForOutput(opts.Store, opts.Save),
		},
	}
	if opts.IncludeAllScored {
		selected := map[string]bool{}
		for _, signal := range signals {
			if id := stringValue(signal["id"]); id != "" {
				selected[id] = true
			}
		}
		var dropped []string
		for _, signal := range allSignals {
			id := stringValue(signal["id"])
			if id != "" && !selected[id] {
				dropped = append(dropped, id)
			}
		}
		payload["debug"] = map[string]any{"all_scored_signals": allSignals, "dropped_signal_ids": nonNilStrings(dropped), "include_all_scored": true}
	}
	if opts.Emit == "brief" {
		fmt.Fprint(stdout, detectorBrief(payload))
	} else {
		writeJSON(stdout, payload)
	}
	return nil
}

func buildQueries(opts detectorOptions, profile monitorProfile) []string {
	var queries []string
	queries = append(queries, opts.Topics...)
	if len(opts.Query) > 0 {
		queries = append(queries, strings.TrimSpace(strings.Join(opts.Query, " ")))
	}
	queries = append(queries, profile.queryTerms()...)
	return dedupeStrings(queries)
}

var defaultMajorFeeds = []string{"https://www.techmeme.com/feed.xml"}

func buildFeedURLs(opts detectorOptions, profile monitorProfile) ([]string, error) {
	var feeds []string
	if !opts.NoProfileFeeds {
		feeds = append(feeds, profile.FeedURLs...)
	}
	if opts.MajorFeeds {
		envFeeds := envMajorFeeds()
		if len(envFeeds) > 0 {
			feeds = append(feeds, envFeeds...)
		} else if len(profile.FeedURLs) == 0 {
			feeds = append(feeds, defaultMajorFeeds...)
		}
	}
	feeds = append(feeds, opts.FeedURLs...)
	for _, file := range opts.FeedFiles {
		read, err := readFeedURLs(file)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, read...)
	}
	return dedupeStrings(feeds), nil
}

func envMajorFeeds() []string {
	raw := os.Getenv("NEWSJACK_MAJOR_FEEDS")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return splitFields(raw)
}

func readFeedURLs(path string) ([]string, error) {
	data, err := os.ReadFile(expandPath(path))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out, nil
}

func requestedSourcesFor(opts detectorOptions, profile monitorProfile) ([]string, error) {
	sources, err := parseSources(opts.Sources)
	if err != nil {
		return nil, err
	}
	if truthy(profile.XNews["enabled"], true) && !opts.NoXNews && !contains(sources, "x_news") {
		sources = append(sources, "x_news")
	}
	trendsMode := strings.ToLower(stringValue(profile.XTrends["mode"]))
	if trendsMode != "" && trendsMode != "none" && trendsMode != "off" && trendsMode != "false" && !opts.NoXTrends && !contains(sources, "x_trends") {
		sources = append(sources, "x_trends")
	}
	if opts.NoXNews {
		sources = removeString(sources, "x_news")
	}
	if opts.NoXTrends {
		sources = removeString(sources, "x_trends")
	}
	return sources, nil
}

func querySources(sources []string) []string {
	var out []string
	for _, s := range sources {
		if s != "x_trends" {
			out = append(out, s)
		}
	}
	return out
}

var defaultSources = []string{"news_search", "x_news", "x"}
var allSources = stringSet([]string{"news_search", "x_news", "x", "x_trends", "reddit", "hackernews"})

func parseSources(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return append([]string{}, defaultSources...), nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		key := strings.ToLower(strings.TrimSpace(part))
		switch key {
		case "":
			continue
		case "hn":
			key = "hackernews"
		case "news":
			key = "news_search"
		case "twitter", "x_posts":
			key = "x"
		}
		if !allSources[key] {
			return nil, fmt.Errorf("unsupported source for v0: %s", part)
		}
		if !contains(out, key) {
			out = append(out, key)
		}
	}
	return out, nil
}

func configFromEnv() map[string]string {
	fileEnv := envFileValues()
	get := func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fileEnv[key]
	}
	return map[string]string{
		"MEDIALYST_API_KEY":        get("MEDIALYST_API_KEY"),
		"MEDIALYST_API_BASE":       get("MEDIALYST_API_BASE"),
		"MEDIALYST_NEWS_PATH":      get("MEDIALYST_NEWS_PATH"),
		"TWITTER_BEARER_TOKEN":     get("TWITTER_BEARER_TOKEN"),
		"X_BEARER_TOKEN":           get("X_BEARER_TOKEN"),
		"X_API_BEARER_TOKEN":       get("X_API_BEARER_TOKEN"),
		"TWITTER_API_BEARER_TOKEN": get("TWITTER_API_BEARER_TOKEN"),
	}
}

func envFileValues() map[string]string {
	paths := []string{}
	if root, err := newsjackRoot(); err == nil {
		paths = append(paths, filepath.Join(root, ".env"))
	}
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".env"))
	}
	out := map[string]string{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, raw := range strings.Split(string(data), "\n") {
			line := strings.TrimSpace(raw)
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			k = strings.TrimSpace(k)
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if k != "" {
				out[k] = v
			}
		}
	}
	return out
}

func availableSources(config map[string]string, requested []string) []string {
	var out []string
	xurl := xurlAvailable()
	bearer := bearerToken(config) != ""
	for _, source := range requested {
		switch {
		case source == "news_search" && config["MEDIALYST_API_KEY"] != "":
			out = append(out, source)
		case source == "x" && xurl:
			out = append(out, source)
		case source == "x_news" && (xurl || bearer):
			out = append(out, source)
		case source == "x_trends" && (xurl || bearer):
			out = append(out, source)
		case source == "reddit" || source == "hackernews":
			out = append(out, source)
		}
	}
	return out
}

func bearerToken(config map[string]string) string {
	for _, key := range []string{"TWITTER_BEARER_TOKEN", "X_BEARER_TOKEN", "X_API_BEARER_TOKEN", "TWITTER_API_BEARER_TOKEN"} {
		if config[key] != "" {
			return config[key]
		}
	}
	return ""
}

func mockItems(query string, now time.Time) []evidenceItem {
	today := now.Format("2006-01-02")
	h1 := sha1.Sum([]byte(query))
	h2 := sha1.Sum([]byte(query + "x"))
	return []evidenceItem{
		{Source: "news_search", Title: "Regulators open inquiry tied to " + query, URL: "https://example.com/news/" + hex.EncodeToString(h1[:])[:8], Container: "Example News", PublishedAt: today, Excerpt: "Officials are examining claims and compliance practices around " + query + ".", Engagement: map[string]any{}, Metadata: map[string]any{}},
		{Source: "x", Title: "Experts are reacting to " + query, URL: "https://x.com/example/status/" + hex.EncodeToString(h2[:])[:8], Author: "example", Container: "x.com", PublishedAt: today, Excerpt: "Thread: the " + query + " inquiry is moving faster than vendors expected.", Engagement: map[string]any{"likes": 120, "reposts": 22, "replies": 9}, Metadata: map[string]any{}},
	}
}

func mockFeedItems(now time.Time) []evidenceItem {
	published := now.Format(time.RFC3339Nano)
	return []evidenceItem{
		{Source: "major_feed", Title: "Salesforce launches free AI customer service agents for startups", URL: "https://example.com/major/salesforce-ai-agents", Container: "Example Major Feed", PublishedAt: published, Excerpt: "A major CRM vendor is targeting startup and SMB customer-support workflows with free AI agents.", Engagement: map[string]any{}, Metadata: map[string]any{"feed_title": "Example Major Feed", "feed_url": "mock://major-feed", "feed_position": 1}},
		{Source: "major_feed", Title: "Pentagon launches task force for safe deployment of AI tools", URL: "https://example.com/major/pentagon-ai-task-force", Container: "Example Major Feed", PublishedAt: published, Excerpt: "The Pentagon is studying how to deploy leading AI tools across sensitive government workflows.", Engagement: map[string]any{}, Metadata: map[string]any{"feed_title": "Example Major Feed", "feed_url": "mock://major-feed", "feed_position": 2}},
	}
}

func collectQuery(query string, sources []string, config map[string]string, opts detectorOptions, now time.Time) ([]evidenceItem, map[string]string) {
	if opts.Mock {
		return mockItems(query, now), map[string]string{}
	}
	from, to := dateRange(opts.LookbackDays)
	errors := map[string]string{}
	var items []evidenceItem
	for _, source := range sources {
		rawItems, err := collectSource(source, query, from, to, opts.Depth, config)
		if err != "" {
			errors[source] = err
		}
		for _, raw := range rawItems {
			item := evidenceFromMap(raw)
			if item.Source != "news_search" && item.Source != "x_news" && jaccard(query, item.text()) < 0.08 {
				continue
			}
			if item.Title != "" || item.Excerpt != "" {
				items = append(items, item)
			}
		}
	}
	return items, errors
}

func collectFeeds(feedURLs []string, depth string, mock bool, now time.Time) ([]evidenceItem, map[string]string) {
	if mock {
		return mockFeedItems(now), map[string]string{}
	}
	limit := map[string]int{"quick": 15, "default": 30, "deep": 60}[depth]
	var items []evidenceItem
	errors := map[string]string{}
	for _, feed := range feedURLs {
		raw, errText := collectFeed(feed, limit)
		if errText != "" {
			errors[feed] = errText
		}
		for _, r := range raw {
			item := evidenceFromMap(r)
			if item.Title != "" || item.Excerpt != "" {
				items = append(items, item)
			}
		}
	}
	return items, errors
}

func collectXTrends(profile monitorProfile, config map[string]string, opts detectorOptions, now time.Time) ([]evidenceItem, map[string]string) {
	mode := strings.ToLower(stringValue(profile.XTrends["mode"]))
	if mode == "" || mode == "none" || mode == "off" || mode == "false" {
		return nil, nil
	}
	if opts.Mock {
		return []evidenceItem{{
			Source:      "x_trends",
			Title:       "Meta Cuts 8,000 Jobs to Focus on AI Future",
			URL:         "https://x.com/search?q=Meta%20Cuts%208000%20Jobs&f=live",
			Author:      "x-trends",
			Container:   "x.com/trends",
			PublishedAt: now.Format(time.RFC3339Nano),
			Excerpt:     "Personalized X trend, 8.7K posts, trending for 5 hours.",
			Engagement:  map[string]any{"score": 500},
			Metadata:    map[string]any{"x_signal_type": "trend", "x_trend_mode": mode, "x_trend_post_count": "8.7K posts"},
		}}, nil
	}
	raw, errText := collectXTrendsRaw(profile.XTrends, opts.Depth, bearerToken(config))
	var items []evidenceItem
	for _, r := range raw {
		item := evidenceFromMap(mapXTrend(r))
		if item.Title != "" || item.Excerpt != "" {
			items = append(items, item)
		}
	}
	if errText != "" {
		return items, map[string]string{"x_trends": errText}
	}
	return items, nil
}

func collectSource(source, query, fromDate, toDate, depth string, config map[string]string) (items []map[string]any, errText string) {
	defer func() {
		if value := recover(); value != nil {
			items = nil
			errText = fmt.Sprintf("panic: %v", value)
		}
	}()
	switch source {
	case "news_search":
		items, err := searchNews(query, fromDate, toDate, limitForDepth(depth), config)
		return items, err
	case "x":
		response := searchX(query, depth)
		if err := stringValue(response["error"]); err != "" {
			return nil, err
		}
		counts := recentCountSummary(query, bearerToken(config))
		parsed := parseXResponse(response, query, counts)
		var out []map[string]any
		for _, item := range parsed {
			if keepXItem(item) {
				out = append(out, mapX(item))
			}
		}
		return out, ""
	case "x_news":
		response := searchXNews(query, depth, lookbackHours(fromDate, toDate), bearerToken(config))
		if err := stringValue(response["error"]); err != "" {
			return nil, err
		}
		var out []map[string]any
		for _, item := range parseXNewsResponse(response, query) {
			out = append(out, mapXNews(item))
		}
		return out, ""
	case "reddit":
		var out []map[string]any
		for _, item := range searchRedditPublic(query, fromDate, toDate, depth) {
			out = append(out, mapReddit(item))
		}
		return out, ""
	case "hackernews":
		response, errText := searchHackerNews(query, fromDate, toDate, depth)
		if errText != "" {
			return nil, errText
		}
		var out []map[string]any
		for _, item := range parseHackerNewsResponse(response, query) {
			out = append(out, mapHackerNews(item))
		}
		return out, ""
	default:
		return nil, "Unsupported source: " + source
	}
}

func limitForDepth(depth string) int {
	return map[string]int{"quick": 10, "default": 25, "deep": 50}[depth]
}

func dateRange(days int) (string, string) {
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -maxInt(days-1, 0))
	return from.Format("2006-01-02"), now.Format("2006-01-02")
}

func detectorDiagnose(sourcesRaw, store string, stdout, stderr io.Writer) int {
	config := configFromEnv()
	requested, err := parseSources(sourcesRaw)
	if err != nil {
		return fail(stderr, err)
	}
	available := availableSources(config, requested)
	payload := map[string]any{
		"sources_requested":         requested,
		"sources_available":         available,
		"news_search_configured":    config["MEDIALYST_API_KEY"] != "",
		"xurl_available":            contains(availableSources(config, []string{"x"}), "x"),
		"x_news_available":          contains(availableSources(config, []string{"x_news"}), "x_news"),
		"x_trends_available":        contains(availableSources(config, []string{"x_trends"}), "x_trends"),
		"twitter_bearer_configured": bearerToken(config) != "",
		"store_path":                dbPathFromEnv(store),
	}
	writeJSON(stdout, payload)
	return 0
}

func parseLaneCaps(raw string) map[string]int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	out := map[string]int{}
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		if n < 0 {
			n = 0
		}
		out[strings.TrimSpace(key)] = n
	}
	return out
}

func selectSignals(all []map[string]any, limit int, laneCaps map[string]int, minQueuePriority, minMajorNews float64) []map[string]any {
	sortedSignals := dedupeSignalsByURL(sortSignalsByQueue(all))
	if laneCaps == nil {
		var selected []map[string]any
		for _, signal := range sortedSignals {
			if passesSelectionFloor(signal, minQueuePriority, minMajorNews) {
				selected = append(selected, signal)
			}
		}
		if limit > 0 && len(selected) > limit {
			return selected[:limit]
		}
		return selected
	}
	var lanes []string
	for lane := range laneCaps {
		lanes = append(lanes, lane)
	}
	sort.Strings(lanes)
	var selected []map[string]any
	selectedIDs := map[string]bool{}
	for _, lane := range lanes {
		count := 0
		for _, signal := range sortedSignals {
			if signalLaneValue(signal) != lane || count >= laneCaps[lane] {
				continue
			}
			id := stringValue(signal["id"])
			if selectedIDs[id] {
				continue
			}
			selected = append(selected, signal)
			selectedIDs[id] = true
			count++
			if limit > 0 && len(selected) >= limit {
				return sortSignalsByQueue(selected)
			}
		}
	}
	for _, signal := range sortedSignals {
		id := stringValue(signal["id"])
		if selectedIDs[id] {
			continue
		}
		if _, ok := laneCaps[signalLaneValue(signal)]; ok {
			continue
		}
		selected = append(selected, signal)
		selectedIDs[id] = true
		if limit > 0 && len(selected) >= limit {
			break
		}
	}
	return sortSignalsByQueue(selected)
}

func passesSelectionFloor(signal map[string]any, minQueuePriority, minMajorNews float64) bool {
	mech, _ := signal["mechanical_scores"].(map[string]any)
	return queuePriority(signal) >= minQueuePriority || floatValue(mech["major_news"]) >= minMajorNews
}

func queuePriority(signal map[string]any) float64 {
	routing, _ := signal["routing"].(map[string]any)
	return floatValue(routing["queue_priority"])
}

func signalLaneValue(signal map[string]any) string {
	routing, _ := signal["routing"].(map[string]any)
	lane := stringValue(routing["lane"])
	if lane == "" {
		return "unknown"
	}
	return lane
}

func sortSignalsByQueue(signals []map[string]any) []map[string]any {
	out := append([]map[string]any{}, signals...)
	sort.SliceStable(out, func(i, j int) bool { return queuePriority(out[i]) > queuePriority(out[j]) })
	return out
}

func dedupeSignalsByURL(signals []map[string]any) []map[string]any {
	seenIDs := map[string]bool{}
	seenURLs := map[string]bool{}
	var out []map[string]any
	for _, signal := range signals {
		id := stringValue(signal["id"])
		urls := evidenceURLs(signal)
		overlap := false
		for _, u := range urls {
			if seenURLs[u] {
				overlap = true
				break
			}
		}
		if seenIDs[id] || (len(urls) > 0 && overlap) {
			continue
		}
		seenIDs[id] = true
		for _, u := range urls {
			seenURLs[u] = true
		}
		out = append(out, signal)
	}
	return out
}

func evidenceURLs(signal map[string]any) []string {
	var out []string
	for _, item := range anySlice(signal["evidence"]) {
		if m, ok := item.(map[string]any); ok {
			if u := stringValue(m["url"]); u != "" {
				out = append(out, u)
			}
		}
	}
	return out
}

func countByLanes(signals []map[string]any) map[string]int {
	out := map[string]int{}
	for _, signal := range signals {
		out[signalLaneValue(signal)]++
	}
	return sortedCountMap(out)
}

func signalIsSeen(signal map[string]any) bool {
	features, _ := signal["features"].(map[string]any)
	v, _ := features["seen_before"].(bool)
	return v
}

func detectorBrief(payload map[string]any) string {
	var lines []string
	lines = append(lines, "newsjack monitor", "")
	diagnostics, _ := payload["diagnostics"].(map[string]any)
	if len(diagnostics) > 0 {
		lines = append(lines, fmt.Sprintf("scored=%v emitted=%v", diagnostics["total_scored_signals"], diagnostics["total_emitted_signals"]))
		if m, ok := diagnostics["evidence_by_source"].(map[string]int); ok && len(m) > 0 {
			lines = append(lines, "evidence_by_source="+joinCounts(m))
		} else if m, ok := diagnostics["evidence_by_source"].(map[string]any); ok && len(m) > 0 {
			lines = append(lines, "evidence_by_source="+joinAnyCounts(m))
		}
		if m, ok := diagnostics["hygiene_rejections"].(map[string]int); ok && len(m) > 0 {
			lines = append(lines, "hygiene_rejections="+joinCounts(m))
		}
		if m, ok := diagnostics["emitted_by_lane"].(map[string]int); ok && len(m) > 0 {
			lines = append(lines, "emitted_by_lane="+joinCounts(m))
		}
		lines = append(lines, "")
	}
	for i, signal := range signalSlice(payload["signals"]) {
		features, _ := signal["features"].(map[string]any)
		routing, _ := signal["routing"].(map[string]any)
		mech, _ := signal["mechanical_scores"].(map[string]any)
		lines = append(lines, fmt.Sprintf("%d. %s (%s, %s, %s)", i+1, signal["title"], routing["lane"], features["decay_bucket"], strings.Join(toStringSlice(signal["sources"]), ", ")))
		lines = append(lines, fmt.Sprintf("   mechanical: queue_priority=%v, profile_match=%v, major_news=%v, momentum=%v", routing["queue_priority"], mech["profile_match"], mech["major_news"], mech["momentum"]))
		for _, link := range briefEvidenceLinks(signal, 3) {
			lines = append(lines, "   evidence: "+link)
		}
		if flags := anySlice(features["safety_flags"]); len(flags) > 0 {
			lines = append(lines, fmt.Sprintf("   safety_flags=%d", len(flags)))
		}
	}
	if len(signalSlice(payload["signals"])) == 0 {
		lines = append(lines, "No signals returned.")
	}
	return strings.Join(lines, "\n") + "\n"
}

func briefEvidenceLinks(signal map[string]any, limit int) []string {
	var out []string
	seen := map[string]bool{}
	for _, item := range anySlice(signal["evidence"]) {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		u := stringValue(m["url"])
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		source := stringValue(m["source"])
		title := firstString(m["title"], m["container"])
		if title != "" {
			out = append(out, fmt.Sprintf("%s: %s - %s", source, truncate(title, 110), u))
		} else {
			out = append(out, fmt.Sprintf("%s: %s", source, u))
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func searchNews(query, fromDate, toDate string, limit int, config map[string]string) ([]map[string]any, string) {
	apiKey := config["MEDIALYST_API_KEY"]
	if apiKey == "" {
		return nil, ""
	}
	base := config["MEDIALYST_API_BASE"]
	if base == "" {
		base = "https://medialyst.ai/api"
	}
	path := config["MEDIALYST_NEWS_PATH"]
	if path == "" {
		path = "/v1/news/search"
	}
	body := map[string]any{"q": query, "num": limit, "gl": "us", "hl": "en", "tbs": tbsForRange(fromDate, toDate)}
	payload, err := httpJSON("POST", strings.TrimRight(base, "/")+path, map[string]string{"Authorization": "Bearer " + apiKey}, body, 30*time.Second)
	if err != nil {
		return nil, err.Error()
	}
	return parseNewsResponse(payload), ""
}

func parseNewsResponse(payload map[string]any) []map[string]any {
	rawItems := firstArray(payload, "items", "results", "news", "organic")
	var items []map[string]any
	for i, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := firstAny(item, "title", "headline", "name")
		link := firstAny(item, "url", "link")
		snippet := firstAny(item, "snippet", "summary", "description", "content")
		source := firstAny(item, "source", "publication", "publisher", "site")
		if sm, ok := source.(map[string]any); ok {
			source = firstAny(sm, "name", "domain")
		}
		published := firstAny(item, "published_at", "publishedAt", "published", "date", "created_at")
		if stringValue(title) == "" && stringValue(snippet) == "" {
			continue
		}
		id := firstString(firstAny(item, "id", "uuid"), fmt.Sprintf("ML%d", i+1))
		items = append(items, map[string]any{
			"id":           id,
			"source":       "news_search",
			"title":        strings.TrimSpace(firstString(title, snippet)),
			"url":          strings.TrimSpace(stringValue(link)),
			"author":       nullableString(stringValue(firstAny(item, "author", "byline"))),
			"container":    nullableString(strings.TrimSpace(stringValue(source))),
			"published_at": nullableString(normalizeLooseDate(stringValue(published))),
			"excerpt":      strings.TrimSpace(stringValue(snippet)),
			"engagement":   map[string]any{},
			"metadata":     map[string]any{"raw_source": source},
		})
	}
	return items
}

func collectFeed(urlOrPath string, limit int) ([]map[string]any, string) {
	var text string
	if regexp.MustCompile(`(?i)^https?://`).MatchString(urlOrPath) {
		resp, err := httpGetRaw(urlOrPath, map[string]string{"Accept": "application/rss+xml, application/atom+xml, text/xml, */*"}, 20*time.Second)
		if err != nil {
			return nil, fmt.Sprintf("%T: %v", err, err)
		}
		text = resp
	} else {
		data, err := os.ReadFile(expandPath(urlOrPath))
		if err != nil {
			return nil, fmt.Sprintf("%T: %v", err, err)
		}
		text = string(data)
	}
	items, err := parseFeed(text, urlOrPath, limit)
	if err != nil {
		return nil, fmt.Sprintf("%T: %v", err, err)
	}
	return items, ""
}

func parseFeed(xmlText, feedURL string, limit int) ([]map[string]any, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlText))
	var stack []string
	var current map[string]string
	var inItem bool
	feedTitle := feedURL
	var items []map[string]any
	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			stack = append(stack, name)
			if name == "item" || name == "entry" {
				inItem = true
				current = map[string]string{}
			}
			if inItem && name == "link" {
				for _, attr := range t.Attr {
					if strings.ToLower(attr.Name.Local) == "href" {
						current["link"] = attr.Value
					}
				}
			}
		case xml.EndElement:
			name := strings.ToLower(t.Name.Local)
			if (name == "item" || name == "entry") && inItem {
				position := len(items) + 1
				if item := feedItemDict(current, feedTitle, feedURL, position); item != nil {
					items = append(items, item)
					if len(items) >= limit {
						return items, nil
					}
				}
				inItem = false
				current = nil
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text == "" || len(stack) == 0 {
				continue
			}
			name := stack[len(stack)-1]
			if inItem {
				if current[name] == "" {
					current[name] = text
				} else {
					current[name] += " " + text
				}
			} else if name == "title" && feedTitle == feedURL {
				feedTitle = cleanText(text)
			}
		}
	}
	return items, nil
}

func feedItemDict(m map[string]string, feedTitle, feedURL string, position int) map[string]any {
	title := cleanText(m["title"])
	link := firstString(m["link"], m["guid"], m["id"])
	excerpt := cleanText(firstString(m["description"], m["summary"], m["content"]))
	published := normalizeLooseDate(firstString(m["pubDate"], m["published"], m["updated"]))
	container := cleanText(firstString(m["source"], feedTitle))
	guid := firstString(m["guid"], m["id"])
	if title == "" && excerpt == "" {
		return nil
	}
	if title == "" {
		title = truncate(excerpt, 120)
	}
	id := firstString(guid, link, fmt.Sprintf("%s#%d", feedURL, position))
	return map[string]any{
		"id":           id,
		"source":       "major_feed",
		"title":        title,
		"url":          strings.TrimSpace(link),
		"author":       nil,
		"container":    firstString(container, feedTitle),
		"published_at": nullableString(published),
		"excerpt":      excerpt,
		"engagement":   map[string]any{},
		"metadata":     map[string]any{"feed_title": feedTitle, "feed_url": feedURL, "feed_position": position},
	}
}

func cleanText(value string) string {
	text := html.UnescapeString(value)
	text = regexp.MustCompile(`(?is)<(script|style).*?</\1>`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`(?s)<[^>]+>`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func xurlAvailable() bool {
	ctx, cancel := contextWithTimeout(10 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xurl", "whoami")
	out, err := cmd.Output()
	return err == nil && strings.Contains(string(out), `"username"`)
}

func searchX(query, depth string) map[string]any {
	maxResults := maxInt(10, minInt(100, map[string]int{"quick": 10, "default": 30, "deep": 60}[depth]))
	normalized := normalizeXQuery(query)
	params := url.Values{}
	params.Set("query", normalized)
	params.Set("max_results", strconv.Itoa(maxResults))
	params.Set("sort_order", "relevancy")
	params.Set("tweet.fields", "created_at,public_metrics,author_id,conversation_id,referenced_tweets,lang,possibly_sensitive")
	params.Set("expansions", "author_id")
	params.Set("user.fields", "username,name,verified,is_identity_verified,public_metrics")
	response := xurlGet("/2/tweets/search/recent?" + params.Encode())
	if response["error"] == nil {
		response["_newsjack_query"] = normalized
		return response
	}
	fallback := xurlSearchShortcut(query, maxResults)
	if fallback["error"] != nil {
		return response
	}
	fallback["_newsjack_query"] = query
	return fallback
}

func recentCountSummary(query, bearer string) map[string]any {
	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(-24 * time.Hour)
	params := url.Values{}
	params.Set("query", normalizeXQuery(query))
	params.Set("granularity", "hour")
	params.Set("start_time", isoZ(start))
	params.Set("end_time", isoZ(now))
	path := "/2/tweets/counts/recent?" + params.Encode()
	var response map[string]any
	if bearer != "" {
		response = apiGet(path, bearer)
	} else {
		response = xurlGetAuth(path, "app")
	}
	if response["error"] != nil {
		return nil
	}
	return summarizeCounts(response)
}

func searchXNews(query, depth string, maxAgeHours int, bearer string) map[string]any {
	maxResults := map[string]int{"quick": 5, "default": 10, "deep": 20}[depth]
	params := url.Values{}
	params.Set("query", strings.TrimSpace(query))
	params.Set("max_results", strconv.Itoa(maxResults))
	params.Set("max_age_hours", strconv.Itoa(maxAgeHours))
	params.Set("news.fields", "id,name,summary,hook,contexts,cluster_posts_results,updated_at,keywords,category")
	path := "/2/news/search?" + params.Encode()
	if bearer != "" {
		return apiGet(path, bearer)
	}
	return xurlGet(path)
}

func parseXNewsResponse(response map[string]any, topic string) []map[string]any {
	var items []map[string]any
	for i, raw := range anySlice(response["data"]) {
		story, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(stringValue(story["name"]))
		hook := strings.TrimSpace(stringValue(story["hook"]))
		summary := strings.TrimSpace(stringValue(story["summary"]))
		if name == "" && hook == "" && summary == "" {
			continue
		}
		postIDs := storyPostIDs(story)
		text := strings.Join(nonEmpty(name, hook, summary), " ")
		engagement := map[string]any{"score": minInt(500, maxInt(1, len(postIDs))*10)}
		if len(postIDs) > 0 {
			engagement["comments"] = len(postIDs)
		}
		items = append(items, map[string]any{
			"id":            fmt.Sprintf("XNEWS%d", i+1),
			"title":         firstString(name, truncate(hook, 120)),
			"text":          truncate(text, 1200),
			"url":           "https://x.com/search?q=" + url.QueryEscape(firstString(name, topic)) + "&f=live",
			"author_handle": "x-news",
			"date":          story["updated_at"],
			"engagement":    engagement,
			"metadata": map[string]any{
				"x_signal_type":             "story_cluster",
				"x_news_id":                 firstAny(story, "id", "rest_id"),
				"x_news_category":           story["category"],
				"x_news_keywords":           valueOrEmptyArray(story["keywords"]),
				"x_news_contexts":           valueOrEmptyMap(story["contexts"]),
				"x_news_cluster_post_ids":   postIDs,
				"x_news_cluster_post_count": len(postIDs),
				"x_news_disclaimer":         story["disclaimer"],
			},
			"why_relevant": "X News story cluster",
			"relevance":    tokenOverlapRelevance(topic, text),
		})
	}
	return items
}

func collectXTrendsRaw(config map[string]any, depth, bearer string) ([]map[string]any, string) {
	mode := strings.ToLower(stringValue(config["mode"]))
	if mode == "personalized" {
		resp := xurlGet("/2/users/personalized_trends?personalized_trend.fields=trend_name%2Cpost_count%2Ccategory%2Ctrending_since")
		return parseTrendsResponse(resp, mode, "", ""), stringValue(resp["error"])
	}
	if mode == "location" {
		if bearer == "" {
			return nil, "x_trends location mode requires TWITTER_BEARER_TOKEN or X_BEARER_TOKEN"
		}
		maxResults := map[string]int{"quick": 10, "default": 20, "deep": 50}[depth]
		locations := stringListValue(config["locations"])
		var items []map[string]any
		var errors []string
		for i, raw := range anySlice(config["woeids"]) {
			woeid := stringValue(raw)
			resp := apiGet("/2/trends/by/woeid/"+url.PathEscape(woeid)+"?max_trends="+strconv.Itoa(maxResults), bearer)
			if err := stringValue(resp["error"]); err != "" {
				errors = append(errors, woeid+": "+err)
				continue
			}
			location := woeid
			if i < len(locations) {
				location = locations[i]
			}
			items = append(items, parseTrendsResponse(resp, mode, woeid, location)...)
		}
		return items, strings.Join(errors, "; ")
	}
	return nil, "Unsupported x_trends mode: " + mode
}

func parseTrendsResponse(response map[string]any, mode, woeid, location string) []map[string]any {
	if response["error"] != nil {
		return nil
	}
	var items []map[string]any
	for i, raw := range anySlice(response["data"]) {
		trend, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(stringValue(trend["trend_name"]))
		if name == "" {
			continue
		}
		countText := firstString(trend["post_count"], trend["tweet_count"])
		count := parseCountText(countText)
		context := strings.Join(nonEmpty(stringValue(trend["category"]), countText, stringValue(trend["trending_since"]), location), ", ")
		engagement := map[string]any{}
		if count > 0 {
			engagement["score"] = minInt(500, count)
		}
		items = append(items, map[string]any{
			"id":            fmt.Sprintf("XTREND%d", i+1),
			"title":         name,
			"text":          strings.TrimSpace(name + ". " + context),
			"url":           "https://x.com/search?q=" + url.QueryEscape(name) + "&f=live",
			"author_handle": "x-trends",
			"date":          time.Now().UTC().Truncate(time.Second).Format(time.RFC3339),
			"engagement":    engagement,
			"metadata": map[string]any{
				"x_signal_type":      "trend",
				"x_trend_mode":       mode,
				"x_trend_woeid":      nullableString(woeid),
				"x_trend_location":   nullableString(location),
				"x_trend_category":   trend["category"],
				"x_trend_post_count": countText,
				"x_trend_since":      trend["trending_since"],
			},
			"why_relevant": "X trend",
			"relevance":    0.5,
		})
	}
	return items
}

func parseXResponse(response map[string]any, topic string, counts map[string]any) []map[string]any {
	var items []map[string]any
	if counts != nil && truthy(counts["is_trending"], false) {
		items = append(items, trendItem(topic, counts))
	}
	data := anySlice(response["data"])
	if len(data) == 0 {
		return items
	}
	authors := map[string]map[string]any{}
	if includes, ok := response["includes"].(map[string]any); ok {
		for _, raw := range anySlice(includes["users"]) {
			if user, ok := raw.(map[string]any); ok {
				authors[stringValue(user["id"])] = user
			}
		}
	}
	for i, raw := range data {
		tweet, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		author := authors[stringValue(tweet["author_id"])]
		username := stringValue(author["username"])
		metrics := valueOrEmptyMap(tweet["public_metrics"])
		authorMetrics := valueOrEmptyMap(author["public_metrics"])
		engagement := map[string]any{}
		if len(metrics) > 0 {
			engagement = map[string]any{
				"likes":     intValue(metrics["like_count"], 0),
				"reposts":   intValue(metrics["retweet_count"], 0),
				"replies":   intValue(metrics["reply_count"], 0),
				"quotes":    intValue(metrics["quote_count"], 0),
				"bookmarks": intValue(metrics["bookmark_count"], 0),
				"views":     intValue(metrics["impression_count"], 0),
			}
		}
		verified := truthy(author["verified"], false) || truthy(author["is_identity_verified"], false)
		proof := socialProof(metrics, authorMetrics, verified)
		date := stringValue(tweet["created_at"])
		text := strings.TrimSpace(stringValue(tweet["text"]))
		u := ""
		if username != "" && stringValue(tweet["id"]) != "" {
			u = "https://x.com/" + username + "/status/" + stringValue(tweet["id"])
		}
		items = append(items, map[string]any{
			"id":            fmt.Sprintf("XURL%d", i+1),
			"text":          truncate(text, 500),
			"url":           u,
			"author_handle": username,
			"date":          nullableString(date),
			"engagement":    engagement,
			"metadata": map[string]any{
				"x_signal_type":      "post",
				"x_author_followers": intValue(authorMetrics["followers_count"], 0),
				"x_author_listed":    intValue(authorMetrics["listed_count"], 0),
				"x_author_verified":  verified,
				"x_low_reach":        len(proof) == 0,
				"x_social_proof":     proof,
				"x_query_counts":     counts,
			},
			"why_relevant": "",
			"relevance":    tokenOverlapRelevance(topic, text),
		})
	}
	return items
}

func keepXItem(item map[string]any) bool {
	metadata := valueOrEmptyMap(item["metadata"])
	if metadata["x_signal_type"] == "query_trend" {
		return true
	}
	return !truthy(metadata["x_low_reach"], false)
}

func mapX(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "x", "title": truncate(stringValue(item["text"]), 120), "url": item["url"], "author": item["author_handle"], "container": "x.com", "published_at": item["date"], "excerpt": item["text"], "engagement": valueOrEmptyMap(item["engagement"]), "metadata": valueOrEmptyMap(item["metadata"])}
}

func mapXNews(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "x_news", "title": firstString(item["title"], truncate(stringValue(item["text"]), 120)), "url": item["url"], "author": item["author_handle"], "container": "x.com/news", "published_at": item["date"], "excerpt": item["text"], "engagement": valueOrEmptyMap(item["engagement"]), "metadata": valueOrEmptyMap(item["metadata"])}
}

func mapXTrend(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "x_trends", "title": firstString(item["title"], truncate(stringValue(item["text"]), 120)), "url": item["url"], "author": item["author_handle"], "container": "x.com/trends", "published_at": item["date"], "excerpt": item["text"], "engagement": valueOrEmptyMap(item["engagement"]), "metadata": valueOrEmptyMap(item["metadata"])}
}

func mapReddit(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "reddit", "title": item["title"], "url": item["url"], "author": item["author"], "container": item["subreddit"], "published_at": item["date"], "excerpt": firstString(item["selftext"], item["body"], item["snippet"]), "engagement": valueOrEmptyMap(item["engagement"]), "metadata": map[string]any{"top_comments": valueOrEmptyArray(item["top_comments"])}}
}

func mapHackerNews(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "hackernews", "title": item["title"], "url": item["url"], "author": item["author"], "container": "news.ycombinator.com", "published_at": item["date"], "excerpt": firstString(item["text"], item["snippet"]), "engagement": valueOrEmptyMap(item["engagement"]), "metadata": map[string]any{}}
}

func xurlGet(path string) map[string]any { return xurlGetAuth(path, "") }

func xurlGetAuth(path, auth string) map[string]any {
	args := []string{}
	if auth != "" {
		args = append(args, "--auth", auth)
	}
	args = append(args, path)
	ctx, cancel := contextWithTimeout(30 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xurl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), contextDeadlineExceeded()) {
			return map[string]any{"error": "xurl request timed out (30s)"}
		}
		if errors.Is(err, exec.ErrNotFound) {
			return map[string]any{"error": "xurl not found in PATH"}
		}
		return map[string]any{"error": "xurl request failed: " + cleanError(string(out))}
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		return map[string]any{"error": "Invalid JSON from xurl: " + err.Error()}
	}
	return payload
}

func apiGet(path, bearer string) map[string]any {
	req, _ := http.NewRequest("GET", "https://api.x.com"+path, nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("%T: %v", err, err)}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if json.Unmarshal(data, &payload) != nil {
		return map[string]any{"error": "Invalid JSON from X API"}
	}
	if resp.StatusCode >= 400 {
		detail := firstString(payload["detail"], payload["title"], truncate(string(data), 300))
		if reason := stringValue(payload["reason"]); reason != "" {
			detail += " (" + reason + ")"
		}
		return map[string]any{"error": fmt.Sprintf("X API HTTP %d: %s", resp.StatusCode, detail)}
	}
	return payload
}

func xurlSearchShortcut(query string, maxResults int) map[string]any {
	ctx, cancel := contextWithTimeout(30 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xurl", "search", query, "-n", strconv.Itoa(maxResults))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{"error": "xurl search failed: " + cleanError(string(out))}
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		return map[string]any{"error": "Invalid JSON from xurl: " + err.Error()}
	}
	return payload
}

func normalizeXQuery(query string) string {
	out := strings.TrimSpace(query)
	var additions []string
	if !strings.Contains(out, "lang:") {
		additions = append(additions, "lang:en")
	}
	for _, op := range []string{"-is:retweet", "-is:reply", "-is:nullcast"} {
		if !strings.Contains(out, op) && !strings.Contains(out, op[1:]) {
			additions = append(additions, op)
		}
	}
	if len(additions) > 0 {
		out += " " + strings.Join(additions, " ")
	}
	return truncate(out, 512)
}

func socialProof(metrics, authorMetrics map[string]any, verified bool) []string {
	likes := intValue(metrics["like_count"], 0)
	reposts := intValue(metrics["retweet_count"], 0)
	replies := intValue(metrics["reply_count"], 0)
	quotes := intValue(metrics["quote_count"], 0)
	views := intValue(metrics["impression_count"], 0)
	followers := intValue(authorMetrics["followers_count"], 0)
	listed := intValue(authorMetrics["listed_count"], 0)
	total := likes + reposts + replies + quotes
	hasView := metrics["impression_count"] != nil
	minEngagement := envInt("NEWSJACK_X_MIN_ENGAGEMENT", 3)
	minFollowers := envInt("NEWSJACK_X_MIN_AUTHOR_FOLLOWERS", 2000)
	minViews := envInt("NEWSJACK_X_MIN_VIEWS", 1000)
	var proof []string
	if total >= minEngagement {
		proof = append(proof, "post_engagement")
	}
	if reposts > 0 || quotes > 0 {
		proof = append(proof, "reshared")
	}
	if views >= minViews {
		proof = append(proof, "views")
	}
	if followers >= minFollowers && !hasView {
		proof = append(proof, "author_followers")
	}
	if listed >= 25 && !hasView {
		proof = append(proof, "author_listed")
	}
	if verified && followers >= minFollowers && !hasView {
		proof = append(proof, "verified_author")
	}
	return proof
}

func summarizeCounts(response map[string]any) map[string]any {
	var counts []int
	for _, raw := range anySlice(response["data"]) {
		if m, ok := raw.(map[string]any); ok {
			counts = append(counts, intValue(m["tweet_count"], 0))
		}
	}
	total24 := sumInts(counts)
	recent6 := sumInts(lastInts(counts, 6))
	previous := counts
	if len(counts) > 6 {
		previous = counts[:len(counts)-6]
	} else {
		previous = nil
	}
	recentPerHour := 0.0
	if len(counts) > 0 {
		recentPerHour = float64(recent6) / 6.0
	}
	previousPerHour := 0.0
	if len(previous) > 0 {
		previousPerHour = float64(sumInts(previous)) / float64(len(previous))
	}
	velocity := (recentPerHour + 1.0) / (previousPerHour + 1.0)
	min24 := envInt("NEWSJACK_X_TREND_MIN_24H", 25)
	min6 := envInt("NEWSJACK_X_TREND_MIN_6H", 8)
	minVelocity := envFloat("NEWSJACK_X_TREND_MIN_VELOCITY", 2.0)
	return map[string]any{"total_24h": total24, "recent_6h": recent6, "previous_hourly_avg": roundN(previousPerHour, 2), "velocity": roundN(velocity, 2), "is_trending": total24 >= min24 && (recent6 >= min6 || velocity >= minVelocity), "bucket_count": len(counts)}
}

func trendItem(topic string, counts map[string]any) map[string]any {
	query := strings.TrimSpace(topic)
	text := fmt.Sprintf("X conversation volume for %q is elevated: %v posts in the last 6h, %v in the last 24h (velocity %vx).", query, counts["recent_6h"], counts["total_24h"], counts["velocity"])
	return map[string]any{
		"id":            "X-TREND",
		"text":          text,
		"url":           "https://x.com/search?q=" + url.QueryEscape(query) + "&f=live",
		"author_handle": "x-search",
		"date":          time.Now().UTC().Truncate(time.Second).Format(time.RFC3339),
		"engagement":    map[string]any{"score": minInt(500, intValue(counts["total_24h"], 0)), "comments": intValue(counts["recent_6h"], 0)},
		"metadata":      map[string]any{"x_signal_type": "query_trend", "x_social_proof": []string{"query_volume"}, "x_query_counts": counts},
		"why_relevant":  "X query-volume trend",
		"relevance":     0.7,
	}
}

func storyPostIDs(story map[string]any) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range anySlice(story["cluster_posts_results"]) {
		if m, ok := raw.(map[string]any); ok {
			id := firstString(m["post_id"], m["id"])
			if id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

func parseCountText(value string) int {
	text := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), ",", "")
	m := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([km]?)`).FindStringSubmatch(text)
	if len(m) == 0 {
		return 0
	}
	n, _ := strconv.ParseFloat(m[1], 64)
	if m[2] == "k" {
		n *= 1000
	} else if m[2] == "m" {
		n *= 1000000
	}
	return int(n)
}

func searchRedditPublic(query, fromDate, toDate, depth string) []map[string]any {
	limit := map[string]int{"quick": 10, "default": 25, "deep": 50}[depth]
	u := "https://www.reddit.com/search.json?q=" + url.QueryEscape(query) + "&sort=relevance&t=month&limit=" + strconv.Itoa(limit) + "&raw_json=1"
	payload, err := httpJSON("GET", u, map[string]string{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"}, nil, 15*time.Second)
	if err != nil {
		return nil
	}
	var posts []map[string]any
	data := valueOrEmptyMap(payload["data"])
	for _, childRaw := range anySlice(data["children"]) {
		child, ok := childRaw.(map[string]any)
		if !ok || child["kind"] != "t3" {
			continue
		}
		post := valueOrEmptyMap(child["data"])
		permalink := strings.TrimSpace(stringValue(post["permalink"]))
		if permalink == "" || !strings.Contains(permalink, "/comments/") {
			continue
		}
		score := intValue(post["score"], 0)
		comments := intValue(post["num_comments"], 0)
		date := ""
		if created := floatValue(post["created_utc"]); created > 0 {
			date = time.Unix(int64(created), 0).UTC().Format("2006-01-02")
		}
		if date != "" && (date < fromDate || date > toDate) {
			continue
		}
		posts = append(posts, map[string]any{
			"id":           fmt.Sprintf("R%d", len(posts)+1),
			"title":        strings.TrimSpace(stringValue(post["title"])),
			"url":          "https://www.reddit.com" + permalink,
			"score":        score,
			"num_comments": comments,
			"subreddit":    strings.TrimSpace(stringValue(post["subreddit"])),
			"author":       firstString(post["author"], "[deleted]"),
			"selftext":     truncate(stringValue(post["selftext"]), 500),
			"date":         nullableString(date),
			"engagement":   map[string]any{"score": score, "num_comments": comments, "upvote_ratio": post["upvote_ratio"]},
			"relevance":    roundN((math.Min(1.0, float64(score)/500.0)*0.6)+(math.Min(1.0, float64(comments)/200.0)*0.4), 3),
		})
	}
	sort.SliceStable(posts, func(i, j int) bool {
		return intValue(valueOrEmptyMap(posts[i]["engagement"])["score"], 0) > intValue(valueOrEmptyMap(posts[j]["engagement"])["score"], 0)
	})
	return posts
}

func searchHackerNews(query, fromDate, toDate, depth string) (map[string]any, string) {
	count := map[string]int{"quick": 15, "default": 30, "deep": 60}[depth]
	fromTS := dateToUnix(fromDate)
	toTS := dateToUnix(toDate) + 86400
	core := flattenQueryForAlgolia(query)
	params := url.Values{}
	params.Set("query", core)
	params.Set("tags", "story")
	params.Set("numericFilters", fmt.Sprintf("created_at_i>%d,created_at_i<%d,points>2", fromTS, toTS))
	params.Set("hitsPerPage", strconv.Itoa(count))
	toks := strings.Fields(core)
	if len(toks) > 1 {
		params.Set("optionalWords", strings.Join(toks[1:], " "))
	}
	payload, err := httpJSON("GET", "https://hn.algolia.com/api/v1/search?"+params.Encode(), nil, nil, 30*time.Second)
	if err != nil {
		return map[string]any{"hits": []any{}}, err.Error()
	}
	return payload, ""
}

func parseHackerNewsResponse(response map[string]any, query string) []map[string]any {
	hits := anySlice(response["hits"])
	var items []map[string]any
	for i, raw := range hits {
		hit, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := stringValue(hit["title"])
		if query != "" && !titleMatchesQuery(title, query) {
			continue
		}
		objectID := stringValue(hit["objectID"])
		points := intValue(hit["points"], 0)
		comments := intValue(hit["num_comments"], 0)
		date := ""
		if ts := intValue(hit["created_at_i"], 0); ts > 0 {
			date = time.Unix(int64(ts), 0).UTC().Format("2006-01-02")
		}
		rank := math.Max(0.3, 1.0-(float64(i)*0.02))
		engagementBoost := math.Min(0.2, math.Log1p(float64(points))/40)
		relevance := math.Min(1.0, 0.6*rank+0.4*tokenOverlapRelevance(query, title)+engagementBoost)
		items = append(items, map[string]any{"id": objectID, "title": title, "url": stringValue(hit["url"]), "hn_url": "https://news.ycombinator.com/item?id=" + objectID, "author": stringValue(hit["author"]), "date": nullableString(date), "engagement": map[string]any{"points": points, "comments": comments}, "relevance": roundN(relevance, 2), "why_relevant": "HN story about " + truncate(title, 60)})
	}
	return items
}

func titleMatchesQuery(title, query string) bool {
	check := strings.ToLower(regexp.MustCompile(`(?i)^(Tell HN|Show HN|Ask HN|Launch HN)\s*:\s*`).ReplaceAllString(title, ""))
	for _, word := range strings.Fields(flattenQueryForAlgolia(strings.ToLower(query))) {
		if regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`).MatchString(check) {
			return true
		}
	}
	return strings.TrimSpace(query) == ""
}

func flattenQueryForAlgolia(text string) string {
	return strings.Join(strings.Fields(strings.NewReplacer(",", " ", "-", " ").Replace(text)), " ")
}

func tokenOverlapRelevance(query, text string) float64 {
	qt, tt := tokens(query), tokens(text)
	if len(qt) == 0 || len(tt) == 0 {
		return 0.5
	}
	hits := 0
	for t := range qt {
		if tt[t] {
			hits++
		}
	}
	return math.Min(1.0, float64(hits)/float64(len(qt)))
}

func dbPathFromEnv(override string) string {
	if override != "" {
		return expandPath(override)
	}
	if v := os.Getenv("NEWSJACK_STORE"); v != "" {
		return expandPath(v)
	}
	return filepath.Join(homeDir(), ".local", "share", "newsjack", "monitor.db")
}

func initDB(override string) error {
	path := dbPathFromEnv(override)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA busy_timeout=5000;
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
CREATE TABLE IF NOT EXISTS seen_urls (url TEXT PRIMARY KEY, first_seen TEXT NOT NULL, last_seen TEXT NOT NULL, sighting_count INTEGER NOT NULL DEFAULT 1);
CREATE TABLE IF NOT EXISTS monitor_runs (id INTEGER PRIMARY KEY, monitor_name TEXT, profile_json TEXT, query_json TEXT NOT NULL, generated_at TEXT NOT NULL, signal_count INTEGER NOT NULL DEFAULT 0, created_at TEXT DEFAULT (datetime('now')));
CREATE TABLE IF NOT EXISTS signal_snapshots (id INTEGER PRIMARY KEY, run_id INTEGER REFERENCES monitor_runs(id) ON DELETE CASCADE, signal_id TEXT NOT NULL, title TEXT NOT NULL, rank_score REAL NOT NULL, payload_json TEXT NOT NULL, created_at TEXT DEFAULT (datetime('now')));
CREATE INDEX IF NOT EXISTS idx_signal_snapshots_run ON signal_snapshots(run_id);
CREATE INDEX IF NOT EXISTS idx_signal_snapshots_rank ON signal_snapshots(rank_score DESC);`)
	return err
}

func openDB(override string) (*sql.DB, error) {
	if err := initDB(override); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPathFromEnv(override))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func seenStatus(urls []string, override string) (map[string]map[string]any, error) {
	if len(urls) == 0 {
		return map[string]map[string]any{}, nil
	}
	db, err := openDB(override)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	unique := dedupeStrings(urls)
	placeholders := strings.TrimRight(strings.Repeat("?,", len(unique)), ",")
	args := make([]any, len(unique))
	for i, u := range unique {
		args[i] = u
	}
	rows, err := db.Query("SELECT url, first_seen, last_seen, sighting_count FROM seen_urls WHERE url IN ("+placeholders+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]map[string]any{}
	for rows.Next() {
		var u, first, last string
		var count int
		if err := rows.Scan(&u, &first, &last, &count); err != nil {
			return nil, err
		}
		out[u] = map[string]any{"first_seen": first, "last_seen": last, "sighting_count": count}
	}
	return out, rows.Err()
}

func recordRun(monitorName string, profile map[string]any, queries []string, signals []map[string]any, seenURLs []string, override string) (int64, error) {
	db, err := openDB(override)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	profileJSON, _ := json.Marshal(profile)
	queryJSON, _ := json.Marshal(queries)
	result, err := tx.Exec("INSERT INTO monitor_runs (monitor_name, profile_json, query_json, generated_at, signal_count) VALUES (?, ?, ?, ?, ?)", nullSQLString(monitorName), string(profileJSON), string(queryJSON), now, len(signals))
	if err != nil {
		return 0, err
	}
	runID, _ := result.LastInsertId()
	for _, signal := range signals {
		payload, _ := json.Marshal(signal)
		if _, err := tx.Exec("INSERT INTO signal_snapshots (run_id, signal_id, title, rank_score, payload_json) VALUES (?, ?, ?, ?, ?)", runID, signal["id"], signal["title"], queuePriority(signal), string(payload)); err != nil {
			return 0, err
		}
	}
	for _, u := range dedupeStrings(seenURLs) {
		if _, err := tx.Exec(`INSERT INTO seen_urls (url, first_seen, last_seen, sighting_count) VALUES (?, ?, ?, 1)
ON CONFLICT(url) DO UPDATE SET last_seen = excluded.last_seen, sighting_count = seen_urls.sighting_count + 1`, u, now, now); err != nil {
			return 0, err
		}
	}
	return runID, tx.Commit()
}

func recentRuns(limit int, override string) ([]map[string]any, error) {
	db, err := openDB(override)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT id, monitor_name, query_json, generated_at, signal_count FROM monitor_runs ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, count int
		var name sql.NullString
		var queryJSON, generated string
		if err := rows.Scan(&id, &name, &queryJSON, &generated, &count); err != nil {
			return nil, err
		}
		var queries []string
		_ = json.Unmarshal([]byte(queryJSON), &queries)
		out = append(out, map[string]any{"id": id, "monitor_name": nullableSQLString(name), "queries": queries, "generated_at": generated, "signal_count": count})
	}
	return out, rows.Err()
}

func storePathForOutput(override string, saved bool) any {
	if !saved {
		return nil
	}
	return dbPathFromEnv(override)
}

func nullSQLString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullableSQLString(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

var allowedDecisions = stringSet([]string{"keep", "monitor_only", "reject"})
var allowedReasons = stringSet([]string{"relevant_news", "plausible_client_bridge", "major_news_no_bridge", "keyword_collision", "not_news", "owned_docs_or_product_page", "seo_landing_page", "low_reach_x_post", "stale", "safety_risk", "duplicate", "off_beat", "no_profile_bridge"})

func cmdFilterApply(args []string, stdout, stderr io.Writer) int {
	var includes stringList
	fs := flag.NewFlagSet("filter-apply", flag.ContinueOnError)
	fs.SetOutput(stderr)
	candidatesPath := fs.String("candidates", "", "Detector JSON output")
	decisionsPath := fs.String("decisions", "", "Cheap-filter decision JSON")
	outputPath := fs.String("output", "", "Output path")
	fs.Var(&includes, "include", "Decision to include. Repeatable.")
	allowMissing := fs.Bool("allow-missing", false, "Do not fail when a candidate has no decision")
	allowUnknown := fs.Bool("allow-unknown", false, "Do not fail when decisions reference unknown IDs")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *candidatesPath == "" || *decisionsPath == "" {
		return fail(stderr, errors.New("--candidates and --decisions are required"))
	}
	candidates, err := readJSONMap(*candidatesPath)
	if err != nil {
		return fail(stderr, err)
	}
	decisionsPayload, err := readJSONAny(*decisionsPath)
	if err != nil {
		return fail(stderr, err)
	}
	includeSet := map[string]bool{}
	if len(includes) == 0 {
		includeSet["keep"] = true
	} else {
		for _, inc := range includes {
			includeSet[inc] = true
		}
	}
	output, err := applyDecisions(candidates, decisionsPayload, includeSet, *allowMissing, *allowUnknown)
	if err != nil {
		return fail(stderr, err)
	}
	data := marshalJSON(output)
	if *outputPath != "" {
		if err := os.WriteFile(expandPath(*outputPath), data, 0o644); err != nil {
			return fail(stderr, err)
		}
	} else {
		stdout.Write(data)
	}
	return 0
}

func applyDecisions(candidates map[string]any, decisionsPayload any, include map[string]bool, allowMissing, allowUnknown bool) (map[string]any, error) {
	signals := signalSlice(candidates["signals"])
	signalByID := map[string]map[string]any{}
	for _, signal := range signals {
		if id := stringValue(signal["id"]); id != "" {
			signalByID[id] = signal
		}
	}
	decisions, err := normalizeDecisions(decisionsPayload)
	if err != nil {
		return nil, err
	}
	decisionByID := map[string]map[string]any{}
	var errs []string
	for _, decision := range decisions {
		id := strings.TrimSpace(stringValue(decision["signal_id"]))
		if id == "" {
			errs = append(errs, "decision missing signal_id")
			continue
		}
		if _, ok := decisionByID[id]; ok {
			errs = append(errs, "duplicate decision for signal_id="+id)
			continue
		}
		if signalByID[id] == nil && !allowUnknown {
			errs = append(errs, "decision references unknown signal_id="+id)
			continue
		}
		normalized := normalizeDecision(decision)
		if !allowedDecisions[stringValue(normalized["decision"])] {
			errs = append(errs, fmt.Sprintf("%s: unsupported decision=%s", id, normalized["decision"]))
		}
		if !allowedReasons[stringValue(normalized["reason"])] {
			errs = append(errs, fmt.Sprintf("%s: unsupported reason=%s", id, normalized["reason"]))
		}
		decisionByID[id] = normalized
	}
	var missing []string
	for id := range signalByID {
		if decisionByID[id] == nil {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 && !allowMissing {
		sample := strings.Join(firstN(missing, 8), ", ")
		suffix := ""
		if len(missing) > 8 {
			suffix = "..."
		}
		errs = append(errs, fmt.Sprintf("missing decisions for %d signal(s): %s%s", len(missing), sample, suffix))
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}
	var selected []map[string]any
	var rejected []map[string]any
	var missingSignals []map[string]any
	decisionCounts := map[string]int{}
	reasonCounts := map[string]int{}
	for _, d := range decisionByID {
		decisionCounts[stringValue(d["decision"])]++
		reasonCounts[stringValue(d["reason"])]++
	}
	for _, signal := range signals {
		id := stringValue(signal["id"])
		decision := decisionByID[id]
		if decision == nil {
			missingSignals = append(missingSignals, summarySignal(signal))
			continue
		}
		withDecision := cloneMap(signal)
		withDecision["cheap_filter"] = decision
		if include[stringValue(decision["decision"])] {
			selected = append(selected, withDecision)
		} else {
			s := summarySignal(signal)
			s["decision"] = decision["decision"]
			s["reason"] = decision["reason"]
			s["rationale"] = decision["rationale"]
			rejected = append(rejected, s)
		}
	}
	included := mapKeys(include)
	sort.Strings(included)
	return map[string]any{
		"version":              1,
		"generated_at":         time.Now().UTC().Format(time.RFC3339Nano),
		"monitor":              valueOrEmptyMap(candidates["monitor"]),
		"signals":              selected,
		"cheap_filter":         map[string]any{"input_signal_count": len(signals), "decision_count": len(decisionByID), "selected_count": len(selected), "rejected_count": len(rejected), "missing_count": len(missingSignals), "included_decisions": included, "decision_counts": sortedCountMap(decisionCounts), "reason_counts": sortedCountMap(reasonCounts), "rejected_signals": rejected, "missing_signals": missingSignals},
		"detector_diagnostics": valueOrEmptyMap(candidates["diagnostics"]),
		"source_errors":        valueOrEmptyMap(candidates["source_errors"]),
	}, nil
}

func normalizeDecisions(payload any) ([]map[string]any, error) {
	if arr, ok := payload.([]any); ok {
		var out []map[string]any
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out, nil
	}
	if m, ok := payload.(map[string]any); ok {
		if arr, ok := m["decisions"].([]any); ok {
			var out []map[string]any
			for _, item := range arr {
				if d, ok := item.(map[string]any); ok {
					out = append(out, d)
				}
			}
			return out, nil
		}
	}
	return nil, errors.New("decisions JSON must be a list or an object with a decisions list")
}

func normalizeDecision(decision map[string]any) map[string]any {
	return map[string]any{
		"signal_id":     strings.TrimSpace(stringValue(decision["signal_id"])),
		"decision":      strings.TrimSpace(stringValue(decision["decision"])),
		"reason":        strings.TrimSpace(stringValue(decision["reason"])),
		"rationale":     strings.TrimSpace(stringValue(decision["rationale"])),
		"confidence":    firstString(strings.TrimSpace(stringValue(decision["confidence"])), "medium"),
		"evidence_urls": toStringSlice(decision["evidence_urls"]),
	}
}

func summarySignal(signal map[string]any) map[string]any {
	var urls []string
	for _, item := range anySlice(signal["evidence"]) {
		if m, ok := item.(map[string]any); ok {
			if u := stringValue(m["url"]); u != "" {
				urls = append(urls, u)
			}
		}
	}
	return map[string]any{"signal_id": signal["id"], "signal_title": signal["title"], "sources": valueOrEmptyArray(signal["sources"]), "routing": valueOrEmptyMap(signal["routing"]), "evidence_urls": firstN(urls, 5)}
}

func cmdSummarizeRun(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("summarize-run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputPath := fs.String("output", "", "Path to write machine-readable summary JSON")
	markdownPath := fs.String("markdown", "", "Path to write Markdown report")
	briefPath := fs.String("brief", "", "Deprecated alias for --markdown")
	top := fs.Int("top", 25, "Number of selected and dropped signals to include")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"output", "markdown", "brief", "top"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outputPath == "" {
		return fail(stderr, errors.New("usage: newsjack summarize-run INPUT --output summary.json --markdown run.md"))
	}
	mdPath := firstString(*markdownPath, *briefPath)
	if mdPath == "" {
		return fail(stderr, errors.New("--markdown is required"))
	}
	payload, err := readJSONMap(fs.Arg(0))
	if err != nil {
		return fail(stderr, err)
	}
	summary := summarizeRun(payload, expandPath(fs.Arg(0)), maxInt(0, *top))
	if err := os.MkdirAll(filepath.Dir(expandPath(*outputPath)), 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.MkdirAll(filepath.Dir(expandPath(mdPath)), 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(expandPath(*outputPath), marshalJSON(summary), 0o644); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(expandPath(mdPath), []byte(renderSummaryMarkdown(summary)), 0o644); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func summarizeRun(payload map[string]any, inputPath string, top int) map[string]any {
	signals := signalSlice(payload["signals"])
	diagnostics := valueOrEmptyMap(firstNonNil(payload["diagnostics"], payload["detector_diagnostics"]))
	debug := valueOrEmptyMap(payload["debug"])
	allScored := signalSlice(debug["all_scored_signals"])
	selectedIDs := map[string]bool{}
	for _, signal := range signals {
		if id := stringValue(signal["id"]); id != "" {
			selectedIDs[id] = true
		}
	}
	var allIDs []string
	var dropped, selectedDebug []map[string]any
	for _, signal := range allScored {
		id := stringValue(signal["id"])
		if id != "" {
			allIDs = append(allIDs, id)
		}
		if selectedIDs[id] {
			selectedDebug = append(selectedDebug, signal)
		} else {
			dropped = append(dropped, signal)
		}
	}
	runDir := filepath.Dir(inputPath)
	paths := artifactPaths(runDir)
	monitor := valueOrEmptyMap(payload["monitor"])
	sourceErrors := valueOrEmptyMap(payload["source_errors"])
	sort.SliceStable(dropped, func(i, j int) bool { return queuePriority(dropped[i]) > queuePriority(dropped[j]) })
	return map[string]any{
		"generated_at": time.Now().UTC().Format(time.RFC3339Nano),
		"input_path":   inputPath,
		"run_dir":      runDir,
		"artifacts":    artifactStatus(paths),
		"pipeline":     pipelineStatus(paths),
		"monitor": map[string]any{
			"name":              monitor["name"],
			"generated_at":      monitor["generated_at"],
			"profile_name":      profileName(monitor),
			"queries":           valueOrEmptyArray(monitor["queries"]),
			"feed_urls":         valueOrEmptyArray(monitor["feed_urls"]),
			"sources_requested": valueOrEmptyArray(monitor["sources_requested"]),
			"sources_used":      valueOrEmptyArray(monitor["sources_used"]),
			"lookback_days":     monitor["lookback_days"],
			"max_age_hours":     monitor["max_age_hours"],
			"depth":             monitor["depth"],
			"mock":              monitor["mock"],
		},
		"counts": map[string]any{
			"selected_unique_signals":        len(signals),
			"total_scored_signals":           firstNonNil(diagnostics["total_scored_signals"], len(allScored)),
			"total_emitted_signals":          firstNonNil(diagnostics["total_emitted_signals"], len(signals)),
			"debug_all_scored_rows":          len(allScored),
			"debug_unique_scored_signal_ids": len(stringSet(allIDs)),
			"debug_selected_rows":            len(selectedDebug),
			"debug_unselected_rows":          len(dropped),
			"debug_duplicate_scored_rows":    len(allIDs) - len(stringSet(allIDs)),
			"source_errors":                  len(sourceErrors),
		},
		"selection":                valueOrEmptyMap(diagnostics["selection"]),
		"lanes":                    map[string]any{"scored": firstNonNil(diagnostics["signals_by_lane"], countByLanes(allScored)), "emitted": firstNonNil(diagnostics["emitted_by_lane"], countByLanes(signals)), "dropped_debug": countByLanes(dropped)},
		"sources":                  map[string]any{"evidence_by_source": firstNonNil(diagnostics["evidence_by_source"], countEvidenceSources(signals)), "source_errors": sourceErrors},
		"hygiene_rejections":       valueOrEmptyMap(diagnostics["hygiene_rejections"]),
		"cheap_filter":             valueOrEmptyMap(payload["cheap_filter"]),
		"cheap_filter_file":        summarizeDecisions(paths["filter_decisions"]),
		"targeted_candidates_file": summarizeTargeted(paths["targeted_candidates"]),
		"final_report_file":        summarizeFinalReport(paths["final_report"]),
		"top_signals":              summarizeSignals(firstNSignals(signals, top)),
		"top_dropped_signals":      summarizeSignals(firstNSignals(dropped, top)),
	}
}

func renderSummaryMarkdown(summary map[string]any) string {
	monitor := valueOrEmptyMap(summary["monitor"])
	counts := valueOrEmptyMap(summary["counts"])
	profile := firstString(monitor["profile_name"], "Newsjack")
	var lines []string
	lines = append(lines, "# "+mdInline(profile)+" Newsjack Brief", "")
	lines = append(lines, renderTable([][2]any{{"Status", statusText(summary)}, {"Generated", formatDatetime(firstString(monitor["generated_at"], summary["generated_at"]))}, {"Detector candidates", counts["selected_unique_signals"]}})...)
	lines = append(lines, "", "## Candidate Preview", "")
	for i, signal := range mapSlice(summary["top_signals"]) {
		lines = append(lines, fmt.Sprintf("%d. **%s**", i+1, mdInline(firstString(signal["title"], "(untitled)"))))
		lines = append(lines, fmt.Sprintf("   - Why surfaced: %s; queue %s, profile %s, major %s.", label(signal["lane"]), fmtValue(signal["queue_priority"]), fmtValue(signal["profile_match"]), fmtValue(signal["major_news"])))
		for _, raw := range anySlice(signal["evidence"]) {
			if ev, ok := raw.(map[string]any); ok {
				lines = append(lines, "   - "+renderEvidenceLink(ev))
			}
		}
	}
	if len(mapSlice(summary["top_signals"])) == 0 {
		lines = append(lines, "- (none)")
	}
	lines = append(lines, "", "## What Was Scanned", "")
	lines = append(lines, fmt.Sprintf("- **Profile:** %s", mdInline(firstString(monitor["profile_name"], "(unknown)"))))
	lines = append(lines, fmt.Sprintf("- **Queries:** %s", mdInline(formatList(valueOrEmptyArray(monitor["queries"]), 8))))
	lines = append(lines, "", "## Appendix: Provenance", "", "### Pipeline", "")
	lines = append(lines, renderPipeline(valueOrEmptyArray(summary["pipeline"]))...)
	lines = append(lines, "", "### Detector Counts", "")
	lines = append(lines, renderTable([][2]any{{"scored", counts["total_scored_signals"]}, {"selected", counts["selected_unique_signals"]}, {"source_errors", counts["source_errors"]}})...)
	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
}

func statusText(summary map[string]any) string {
	finalReport := valueOrEmptyMap(summary["final_report_file"])
	if truthy(finalReport["exists"], false) {
		return "Editorial review complete"
	}
	return "Detector preview only"
}

func artifactPaths(runDir string) map[string]string {
	return map[string]string{"candidates": filepath.Join(runDir, "candidates.json"), "detector_summary": filepath.Join(runDir, "summary.json"), "commands": filepath.Join(runDir, "commands.log"), "detector_stderr": filepath.Join(runDir, "detector.stderr.log"), "filter_decisions": filepath.Join(runDir, "filter_decisions.json"), "targeted_candidates": filepath.Join(runDir, "targeted_candidates.json"), "final_report": filepath.Join(runDir, "final_report.md"), "run_markdown": filepath.Join(runDir, "run.md")}
}

func artifactStatus(paths map[string]string) map[string]any {
	out := map[string]any{}
	for name, path := range paths {
		info, err := os.Stat(path)
		exists := err == nil
		size := int64(0)
		if exists {
			size = info.Size()
		}
		out[name] = map[string]any{"path": path, "exists": exists, "bytes": size}
	}
	return out
}

func pipelineStatus(paths map[string]string) []map[string]string {
	return []map[string]string{stage("detector", paths["candidates"]), stage("cheap_filter", paths["filter_decisions"]), stage("filter_apply", paths["targeted_candidates"]), stage("final_report", paths["final_report"])}
}

func stage(name, path string) map[string]string {
	status := "pending"
	if fileExists(path) {
		status = "done"
	}
	return map[string]string{"stage": name, "status": status, "artifact": filepath.Base(path)}
}

func summarizeDecisions(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	payload, err := readJSONMap(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	decisions := mapSlice(payload["decisions"])
	outcome := map[string]int{}
	reasons := map[string]int{}
	for _, d := range decisions {
		outcome[firstString(d["decision"], "unknown")]++
		reasons[firstString(d["reason"], "unknown")]++
	}
	return map[string]any{"exists": true, "path": path, "decision_count": len(decisions), "decisions_by_outcome": outcome, "decisions_by_reason": reasons}
}

func summarizeTargeted(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	payload, err := readJSONMap(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	cheap := valueOrEmptyMap(payload["cheap_filter"])
	return map[string]any{"exists": true, "path": path, "selected_signals": len(signalSlice(payload["signals"])), "input_signals": cheap["input_signal_count"], "rejected_signals": cheap["rejected_count"]}
}

func summarizeFinalReport(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	return map[string]any{"exists": true, "path": path, "bytes": len(data), "content": string(data)}
}

func profileName(monitor map[string]any) any {
	profile := valueOrEmptyMap(monitor["profile"])
	if v := firstString(profile["name"], profile["company"], profile["client"]); v != "" {
		return v
	}
	return nil
}

func countEvidenceSources(signals []map[string]any) map[string]int {
	out := map[string]int{}
	for _, signal := range signals {
		for _, raw := range anySlice(signal["evidence"]) {
			if ev, ok := raw.(map[string]any); ok {
				out[firstString(ev["source"], "unknown")]++
			}
		}
	}
	return out
}

func summarizeSignals(signals []map[string]any) []map[string]any {
	var out []map[string]any
	for _, signal := range signals {
		routing := valueOrEmptyMap(signal["routing"])
		mech := valueOrEmptyMap(signal["mechanical_scores"])
		out = append(out, map[string]any{"id": signal["id"], "title": firstString(signal["title"], firstEvidenceValue(signal, "title")), "query": signal["query"], "lane": routing["lane"], "queue_priority": routing["queue_priority"], "decay_bucket": firstNonNil(signal["decay_bucket"], mech["decay_bucket"]), "profile_match": mech["profile_match"], "major_news": mech["major_news"], "momentum": mech["momentum"], "source_agreement": mech["source_agreement"], "cheap_filter": signal["cheap_filter"], "evidence": summarizeEvidence(signal)})
	}
	return out
}

func summarizeEvidence(signal map[string]any) []map[string]any {
	var out []map[string]any
	for _, raw := range anySlice(signal["evidence"]) {
		if ev, ok := raw.(map[string]any); ok {
			out = append(out, map[string]any{"source": ev["source"], "title": ev["title"], "url": ev["url"], "published_at": ev["published_at"], "author": ev["author"], "engagement": valueOrEmptyMap(ev["engagement"])})
		}
	}
	return out
}

func firstEvidenceValue(signal map[string]any, key string) any {
	for _, raw := range anySlice(signal["evidence"]) {
		if ev, ok := raw.(map[string]any); ok && ev[key] != nil && stringValue(ev[key]) != "" {
			return ev[key]
		}
	}
	return nil
}

func firstNSignals(signals []map[string]any, n int) []map[string]any {
	if n < 0 {
		n = 0
	}
	if len(signals) > n {
		return signals[:n]
	}
	return signals
}

func renderEvidenceLink(ev map[string]any) string {
	source := label(ev["source"])
	title := mdInline(firstString(ev["title"], "(no title)"))
	u := strings.TrimSpace(stringValue(ev["url"]))
	suffix := ""
	if published := stringValue(ev["published_at"]); published != "" {
		suffix = " (" + mdInline(published) + ")"
	}
	if u != "" {
		return fmt.Sprintf("%s: [%s](%s)%s", source, escapeLinkText(title), u, suffix)
	}
	return source + ": " + title + suffix
}

func renderPipeline(stages []any) []string {
	var rows [][2]any
	for _, raw := range stages {
		if st, ok := raw.(map[string]any); ok {
			rows = append(rows, [2]any{st["stage"], fmt.Sprintf("%v - %v", st["status"], st["artifact"])})
		}
	}
	return renderTable(rows)
}

func renderTable(rows [][2]any) []string {
	if len(rows) == 0 {
		return []string{"- (none)"}
	}
	lines := []string{"| key | value |", "|---|---|"}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("| %s | %s |", mdCell(row[0]), mdCell(row[1])))
	}
	return lines
}

func mdCell(v any) string { return strings.ReplaceAll(mdInline(v), "|", `\|`) }
func mdInline(v any) string {
	return strings.Join(strings.Fields(fmt.Sprint(v)), " ")
}
func label(v any) string {
	return strings.ReplaceAll(strings.ReplaceAll(mdInline(v), "_", " "), "-", " ")
}
func escapeLinkText(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(v, "[", `\[`), "]", `\]`)
}

func formatDatetime(v string) string {
	if v == "" || v == "<nil>" {
		return "(unknown)"
	}
	raw := strings.TrimSuffix(v, "Z") + strings.TrimPrefix("Z", "Z")
	if strings.HasSuffix(v, "Z") {
		raw = strings.TrimSuffix(v, "Z") + "+00:00"
	}
	if parsed, ok := parseTime(raw); ok {
		return parsed.UTC().Format("2006-01-02 15:04 UTC")
	}
	return v
}

func formatList(values []any, limit int) string {
	var clean []string
	for _, v := range values {
		if s := mdInline(v); s != "" {
			clean = append(clean, s)
		}
	}
	if len(clean) == 0 {
		return "(none)"
	}
	if len(clean) <= limit {
		return strings.Join(clean, ", ")
	}
	return strings.Join(clean[:limit], ", ") + fmt.Sprintf(", plus %d more", len(clean)-limit)
}

func fmtValue(v any) string {
	if v == nil {
		return "-"
	}
	if f, ok := numberValue(v); ok {
		return strconv.FormatFloat(f, 'g', 3, 64)
	}
	return fmt.Sprint(v)
}

func httpJSON(method, rawURL string, headers map[string]string, body any, timeout time.Duration) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func httpGetRaw(rawURL string, headers map[string]string, timeout time.Duration) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	return string(data), nil
}

func writeJSON(w io.Writer, payload any) {
	w.Write(marshalJSON(payload))
}

func writeJSONCompact(w io.Writer, payload any) {
	data, _ := json.Marshal(payload)
	w.Write(append(data, '\n'))
}

func marshalJSON(payload any) []byte {
	data, _ := json.MarshalIndent(payload, "", "  ")
	return append(data, '\n')
}

func readJSONMap(path string) (map[string]any, error) {
	payload, err := readJSONAny(path)
	if err != nil {
		return nil, err
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain a JSON object", path)
	}
	return m, nil
}

func readJSONAny(path string) (any, error) {
	data, err := os.ReadFile(expandPath(path))
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func nonEmpty(values ...string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		out[v] = true
	}
	return out
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, strings.TrimSpace(v))
	}
	return nonNilStrings(out)
}

func emptyStrings() []string {
	return []string{}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func contains(values []string, needle string) bool {
	for _, v := range values {
		if v == needle {
			return true
		}
	}
	return false
}

func removeString(values []string, needle string) []string {
	var out []string
	for _, v := range values {
		if v != needle {
			out = append(out, v)
		}
	}
	return out
}

func mergeCounts(left, right map[string]int) {
	for k, v := range right {
		left[k] += v
	}
}

func floatValue(v any) float64 {
	f, _ := numberValue(v)
	return f
}

func numberValue(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func intValue(v any, fallback int) int {
	if f, ok := numberValue(v); ok {
		return int(f)
	}
	return fallback
}

func truthy(v any, fallback bool) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		if x == "" {
			return fallback
		}
		return x == "1" || strings.EqualFold(x, "true") || strings.EqualFold(x, "yes")
	default:
		if v == nil {
			return fallback
		}
		return fallback
	}
}

func anySlice(v any) []any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []any:
		return x
	case []string:
		out := make([]any, len(x))
		for i, v := range x {
			out[i] = v
		}
		return out
	default:
		return nil
	}
}

func mapSlice(v any) []map[string]any {
	var out []map[string]any
	for _, item := range anySlice(v) {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func signalSlice(v any) []map[string]any { return mapSlice(v) }

func toStringSlice(v any) []string {
	if s, ok := v.(string); ok {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return []string{s}
	}
	var out []string
	for _, item := range anySlice(v) {
		if s := stringValue(item); strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func valueOrEmptyMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok && m != nil {
		return m
	}
	return map[string]any{}
}

func valueOrEmptyArray(v any) []any {
	if arr := anySlice(v); arr != nil {
		return arr
	}
	return []any{}
}

func firstArray(m map[string]any, keys ...string) []any {
	for _, key := range keys {
		if arr := anySlice(m[key]); len(arr) > 0 {
			return arr
		}
	}
	return nil
}

func firstAny(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok && v != nil && stringValue(v) != "" {
			return v
		}
	}
	return nil
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func firstN[T any](values []T, n int) []T {
	if len(values) > n {
		return values[:n]
	}
	return values
}

func mapKeys(m map[string]bool) []string {
	var out []string
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	return out
}

func cloneMap(m map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func sortedCountMap(m map[string]int) map[string]int {
	return m
}

func joinCounts(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func joinAnyCounts(m map[string]any) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%v", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func splitFields(raw string) []string {
	var out []string
	for _, part := range regexp.MustCompile(`[\n,]`).Split(raw, -1) {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return nonNilStrings(out)
}

func reorderIntermixedFlags(args []string, valueFlags map[string]bool) []string {
	var flagsOut []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positional = append(positional, arg)
			continue
		}
		name := strings.TrimLeft(arg, "-")
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		flagsOut = append(flagsOut, arg)
		if strings.Contains(arg, "=") {
			continue
		}
		if valueFlags[name] && i+1 < len(args) {
			i++
			flagsOut = append(flagsOut, args[i])
		}
	}
	return append(flagsOut, positional...)
}

func boolSlice(ok bool, value string) []string {
	if ok {
		return []string{value}
	}
	return nil
}

func roundN(v float64, n int) float64 {
	p := math.Pow10(n)
	return math.Round(v*p) / p
}

func round1(v float64) float64 { return roundN(v, 1) }

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sumInts(values []int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}

func lastInts(values []int, n int) []int {
	if len(values) <= n {
		return values
	}
	return values[len(values)-n:]
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(key), 64); err == nil {
		return v
	}
	return fallback
}

func dateToUnix(date string) int64 {
	t, _ := time.Parse("2006-01-02", date[:10])
	return t.Unix()
}

func lookbackHours(fromDate, toDate string) int {
	start, err1 := time.Parse("2006-01-02", fromDate[:10])
	end, err2 := time.Parse("2006-01-02", toDate[:10])
	if err1 != nil || err2 != nil {
		return 168
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days*24 < 24 {
		return 24
	}
	return days * 24
}

func tbsForRange(fromDate, toDate string) string {
	start, err1 := time.Parse("2006-01-02", fromDate[:10])
	end, err2 := time.Parse("2006-01-02", toDate[:10])
	if err1 != nil || err2 != nil {
		return "qdr:w"
	}
	days := int(end.Sub(start).Hours() / 24)
	if days <= 1 {
		return "qdr:d"
	}
	if days <= 7 {
		return "qdr:w"
	}
	return "qdr:m"
}

func normalizeLooseDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) >= 10 && value[4] == '-' && value[7] == '-' {
		return value
	}
	rel := regexp.MustCompile(`(?i)^\s*(\d+)\s+(minute|minutes|hour|hours|day|days|week|weeks|month|months)\s+ago\s*$`).FindStringSubmatch(value)
	if len(rel) > 0 {
		n, _ := strconv.Atoi(rel[1])
		unit := strings.ToLower(rel[2])
		d := time.Duration(0)
		switch {
		case strings.HasPrefix(unit, "minute"):
			d = time.Duration(n) * time.Minute
		case strings.HasPrefix(unit, "hour"):
			d = time.Duration(n) * time.Hour
		case strings.HasPrefix(unit, "day"):
			d = time.Duration(n) * 24 * time.Hour
		case strings.HasPrefix(unit, "week"):
			d = time.Duration(n) * 7 * 24 * time.Hour
		default:
			d = time.Duration(n) * 30 * 24 * time.Hour
		}
		return time.Now().UTC().Add(-d).Format(time.RFC3339Nano)
	}
	if parsed, err := http.ParseTime(value); err == nil {
		return parsed.UTC().Format(time.RFC3339Nano)
	}
	return value
}

func isoZ(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func cleanError(text string) string {
	return regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`).ReplaceAllString(strings.TrimSpace(text), "")
}

func mustRegexes(patterns []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		out[i] = regexp.MustCompile(pattern)
	}
	return out
}

type cancelFunc func()

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func contextDeadlineExceeded() error { return context.DeadlineExceeded }
