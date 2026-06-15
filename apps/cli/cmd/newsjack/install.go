package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type runtimeTarget struct {
	Key      string
	Label    string
	DirEnv   string
	Default  string
	Binary   string
	HomeHint string
}

var runtimeTargets = []runtimeTarget{
	{"codex", "Codex", "NEWSJACK_CODEX_SKILLS_DIR", "", "codex", ""},
	{"claude", "Claude Code", "NEWSJACK_CLAUDE_SKILLS_DIR", "", "claude", ""},
	{"openclaw", "OpenClaw", "NEWSJACK_OPENCLAW_SKILLS_DIR", "", "openclaw", ""},
	{"hermes", "Hermes", "NEWSJACK_HERMES_SKILLS_DIR", "", "hermes", ""},
}

func targetDir(rt runtimeTarget) string {
	if v := os.Getenv(rt.DirEnv); v != "" {
		return expandPath(v)
	}
	return defaultRuntimeSkillsDir(rt.Key)
}

func defaultRuntimeSkillsDir(key string) string {
	switch key {
	case "codex":
		return filepath.Join(homeDir(), ".agents", "skills")
	case "claude":
		return filepath.Join(homeDir(), ".claude", "skills")
	case "openclaw":
		return filepath.Join(homeDir(), ".openclaw", "skills")
	case "hermes":
		return filepath.Join(homeDir(), ".hermes", "skills")
	default:
		return filepath.Join(homeDir(), ".agents", "skills")
	}
}

func runtimeHomeHint(key string) string {
	return filepath.Dir(defaultRuntimeSkillsDir(key))
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
	return runtimeCLIInstalled(rt)
}

func runtimeByKey(key string) (runtimeTarget, bool) {
	for _, rt := range runtimeTargets {
		if rt.Key == key {
			return rt, true
		}
	}
	return runtimeTarget{}, false
}

func runtimeCLIInstalled(rt runtimeTarget) bool {
	_, ok := resolveRuntimeBinary(rt)
	return ok
}

// resolveRuntimeBinary finds the runtime CLI. PATH wins; on Windows,
// Claude Code's installers do not put `claude` on PATH, so its known
// install locations are probed as a fallback.
func resolveRuntimeBinary(rt runtimeTarget) (string, bool) {
	if rt.Binary == "" {
		return "", false
	}
	if path, err := exec.LookPath(rt.Binary); err == nil {
		return path, true
	}
	if goos() == "windows" && rt.Key == "claude" {
		if path := findWindowsClaudeBinary(os.Getenv("APPDATA"), homeDir()); path != "" {
			return path, true
		}
	}
	return "", false
}

// findWindowsClaudeBinary probes Claude Code's Windows install locations:
// the native installer's %USERPROFILE%\.local\bin and the per-version
// directories under %APPDATA%\Claude\claude-code.
func findWindowsClaudeBinary(appData, home string) string {
	if home != "" {
		candidate := filepath.Join(home, ".local", "bin", "claude.exe")
		if fileExists(candidate) {
			return candidate
		}
	}
	if appData == "" {
		return ""
	}
	root := filepath.Join(appData, "Claude", "claude-code")
	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	// Version-named subdirectories; prefer the lexically newest.
	for i := len(entries) - 1; i >= 0; i-- {
		if !entries[i].IsDir() {
			continue
		}
		candidate := filepath.Join(root, entries[i].Name(), "claude.exe")
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
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
	Source   string
	Runtimes string
	Force    bool
	CLI      commandInvocation
	Repo     string
	Ref      string
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
	if err := fs.Parse(args); err != nil {
		return 2
	}
	opts := installOptions{
		Source:   expandPath(*source),
		Runtimes: *runtimes,
		Force:    *force,
		CLI:      newsjackCLIInvocation(),
		Repo:     getenv("NEWSJACK_REPO", defaultRepo),
		Ref:      getenv("NEWSJACK_REF", defaultRef),
	}
	if shouldAdoptBundle(opts.Source) {
		logf(stdout, "adopting bundle %s into %s", opts.Source, managedInstallDir())
		if err := adoptBundle(opts.Source, opts.Runtimes, stderr); err != nil {
			return fail(stderr, err)
		}
		opts.Source = managedInstallDir()
		opts.CLI = newsjackCLIInvocation()
	}
	if err := installRuntimeSkills(opts, stdout, stderr); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func installRuntimeSkills(opts installOptions, stdout, stderr io.Writer) error {
	if !dirExists(filepath.Join(opts.Source, "skills")) {
		return fmt.Errorf("missing skills directory in %s", opts.Source)
	}
	targets := selectedRuntimes(opts.Runtimes)
	if len(targets) == 0 && !strings.Contains(","+strings.Join(normalizeRuntimeList(opts.Runtimes), ",")+",", ",none,") {
		warn(stderr, "no supported runtime detected; installing portable skills into %s — note Claude Code reads %s, so run `newsjack install --runtimes claude` to target it explicitly", filepath.Join(homeDir(), ".agents", "skills"), filepath.Join(homeDir(), ".claude", "skills"))
		targets = []runtimeTarget{{Key: "portable", Label: "portable Agent Skills"}}
	}
	for _, rt := range targets {
		if err := installSkillsTo(opts, rt.Label, targetDir(rt)); err != nil {
			return err
		}
		successf(stdout, "installed skills for %s into %s", rt.Label, targetDir(rt))
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

func copySkillTree(src, dest string, _ installOptions) error {
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
