package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const defaultClaudeInstallCommand = "curl -fsSL https://claude.ai/install.sh | bash"
const defaultClaudeInstallCommandWindows = "irm https://claude.ai/install.ps1 | iex"
const xAPIKeyURL = "https://docs.x.com/fundamentals/authentication/oauth-2-0/bearer-tokens"
const medialystAPIKeyURL = "https://medialyst.ai/agents"

func cmdSetup(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "Emit setup status as JSON")
	runtimeRaw := fs.String("runtime", "auto", "Skill runtime target: auto, all, other, codex,claude,openclaw,hermes")
	scheduleRuntimeRaw := fs.String("schedule-runtime", "auto", "Agent harness for scheduled runs")
	installClaude := fs.Bool("install-claude", false, "Install Claude Code if it is missing")
	yes := fs.Bool("yes", false, "Answer yes to setup prompts")
	skipClaudeInstall := fs.Bool("skip-claude-install", false, "Do not offer to install Claude Code")
	noLaunch := fs.Bool("no-launch", false, "Do not launch an agent harness after setup")
	skipCredentials := fs.Bool("skip-credentials", false, "Do not prompt for optional API keys")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	// In --json mode bootstrap progress goes to stderr so stdout stays
	// valid JSON, mirroring the auto-update contract.
	bootstrapOut := stdout
	if *jsonOut {
		bootstrapOut = stderr
	}
	if err := ensureInstalledRoot(*runtimeRaw, bootstrapOut, stderr); err != nil {
		return fail(stderr, err)
	}
	status := setupPayload(*runtimeRaw)
	if *jsonOut {
		writeJSON(stdout, status)
		return 0
	}

	wizard := setupWizard{
		reader:              bufio.NewReader(stdin),
		stdin:               stdin,
		stdout:              stdout,
		stderr:              stderr,
		assumeYes:           *yes,
		installClaude:       *installClaude,
		skipClaudeInstall:   *skipClaudeInstall,
		noLaunch:            *noLaunch,
		skipCredentials:     *skipCredentials,
		defaultSkillRuntime: *runtimeRaw,
		defaultScheduler:    *scheduleRuntimeRaw,
	}
	if err := wizard.run(); err != nil {
		return fail(stderr, err)
	}
	return 0
}

type setupWizard struct {
	reader              *bufio.Reader
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	assumeYes           bool
	installClaude       bool
	skipClaudeInstall   bool
	noLaunch            bool
	skipCredentials     bool
	defaultSkillRuntime string
	defaultScheduler    string
	interactive         bool
	runtimeInstallSeen  map[string]bool
}

func (w *setupWizard) run() error {
	uiProduct(w.stdout, "guided setup", "choose skills, scheduler, optional auth, then launch your agent.")
	fmt.Fprintln(w.stdout)
	uiKV(w.stdout, "home", newsjackHome())
	uiKV(w.stdout, "mock test", "newsjack monitor test <slug> --mock")
	fmt.Fprintln(w.stdout)

	skillRuntime, schedulerRuntime := w.chooseAgentRuntime()
	manualSkillRuntime := runtimeSelectionIncludes(skillRuntime, "manual")
	installSkillRuntime := nonManualRuntimeSelection(skillRuntime)
	if err := w.offerRuntimeInstalls(installSkillRuntime, "skill installation"); err != nil {
		return err
	}
	if installSkillRuntime != "" {
		if err := installSetupSkills(installSkillRuntime, w.stdout, w.stderr); err != nil {
			return err
		}
	} else if !manualSkillRuntime {
		if err := installSetupSkills(skillRuntime, w.stdout, w.stderr); err != nil {
			return err
		}
	}

	schedulerReady := true
	if !isManualRuntime(schedulerRuntime) {
		var err error
		schedulerReady, err = w.ensureRuntimeInstalled(schedulerRuntime, "scheduled runs")
		if err != nil {
			return err
		}
	}
	if !isManualRuntime(schedulerRuntime) && !runtimeSelectionIncludes(skillRuntime, schedulerRuntime) {
		if err := installSetupSkills(schedulerRuntime, w.stdout, w.stderr); err != nil {
			return err
		}
	}

	if !schedulerReady {
		fmt.Fprintln(w.stdout)
		uiSection(w.stdout, "scheduler runtime missing")
		uiNote(w.stdout, "%s must be installed and on PATH before Newsjack can launch it for scheduled runs.", runtimeLabel(schedulerRuntime))
		w.printRuntimeInstallCommand(schedulerRuntime)
		uiNote(w.stdout, "then rerun: newsjack setup --schedule-runtime %s", schedulerRuntime)
		return nil
	}

	if !w.skipCredentials {
		if err := w.promptForXAPIKey(); err != nil {
			return err
		}
		if err := w.promptForMedialystAPIKey(); err != nil {
			return err
		}
	}
	agentPrompt := setupAgentPrompt(schedulerRuntime)
	plan := setupLaunchPlanFor(schedulerRuntime, agentPrompt)
	command := setupAgentCommand(schedulerRuntime, agentPrompt)
	fmt.Fprintln(w.stdout)
	if isManualRuntime(schedulerRuntime) {
		uiSection(w.stdout, "manual setup")
		uiNote(w.stdout, "copy the skills, then give your agent this prompt.")
		fmt.Fprintln(w.stdout)
		fmt.Fprintln(w.stdout, manualSkillInstallInstruction())
		fmt.Fprintln(w.stdout)
		fmt.Fprintln(w.stdout, "Prompt:")
		fmt.Fprintln(w.stdout, agentPrompt)
		return nil
	}
	if len(plan.args) == 0 {
		// The runtime has the skills installed but no interactive terminal
		// session, so it can't answer the agent's follow-up questions in place.
		// Hand the prompt off instead of launching a one-shot run that would stop
		// the moment setup needs input.
		uiSection(w.stdout, "manual setup")
		uiNote(w.stdout, "%s runs one-shot and can't host an interactive setup session.", runtimeLabel(schedulerRuntime))
		uiNote(w.stdout, "open %s and send this prompt:", runtimeLabel(schedulerRuntime))
		fmt.Fprintln(w.stdout)
		fmt.Fprintln(w.stdout, "Prompt:")
		fmt.Fprintln(w.stdout, agentPrompt)
		return nil
	}

	uiSection(w.stdout, "launch")
	uiSuccess(w.stdout, "Ready to run low-effort auto-setup in %s.", runtimeLabel(schedulerRuntime))
	fmt.Fprintf(w.stdout, "Command: %s\n\n", command)
	if w.noLaunch || !w.confirm(fmt.Sprintf("Run auto-setup in %s now?", runtimeLabel(schedulerRuntime)), true) {
		uiNote(w.stdout, "run this command when ready:")
		fmt.Fprintln(w.stdout, command)
		if !plan.seeded {
			fmt.Fprintln(w.stdout)
			uiNote(w.stdout, "then send this prompt in %s:", runtimeLabel(schedulerRuntime))
			fmt.Fprintln(w.stdout, agentPrompt)
		}
		fmt.Fprintln(w.stdout)
		return nil
	}
	return launchSetupAgent(schedulerRuntime, plan, agentPrompt, w.stdin, w.stdout, w.stderr)
}

func (w *setupWizard) chooseAgentRuntime() (string, string) {
	if skillRuntime, schedulerRuntime, ok := w.explicitRuntimePair(); ok {
		uiSection(w.stdout, "agent runtime")
		uiKV(w.stdout, "skills", runtimeSelectionLabel(skillRuntime))
		uiKV(w.stdout, "schedule", runtimeLabel(schedulerRuntime))
		return skillRuntime, schedulerRuntime
	}

	defaultRuntime := w.defaultAgentRuntime()
	uiSection(w.stdout, "agent runtime")
	uiNote(w.stdout, "this installs Newsjack skills and owns scheduled runs.")
	uiNote(w.stdout, "scheduled runs stay inside an agent harness, not system cron.")
	fmt.Fprintln(w.stdout)
	fmt.Fprintln(w.stdout, "Supported agent runtimes:")
	printRuntimeList(w.stdout)
	fmt.Fprintln(w.stdout)

	choices := []setupChoice{
		{Value: "hermes", Label: "Hermes", Hint: runtimeChoiceHint("hermes")},
		{Value: "openclaw", Label: "OpenClaw", Hint: runtimeChoiceHint("openclaw")},
		{Value: "claude", Label: "Claude Code", Hint: runtimeChoiceHint("claude")},
		{Value: "codex", Label: "Codex", Hint: runtimeChoiceHint("codex")},
		{Value: "manual", Label: "Other/manual", Hint: "copy instructions"},
	}
	if value, ok := w.selectSingle("Which agent should run Newsjack?", choices, defaultRuntime); ok {
		return setupAndScheduleRuntime(value, defaultRuntime)
	}

	answer := w.prompt("Which agent should run Newsjack? [hermes,openclaw,claude,codex,other]", defaultRuntime)
	return setupAndScheduleRuntime(answer, defaultRuntime)
}

func (w *setupWizard) explicitRuntimePair() (string, string, bool) {
	if !explicitRuntimeValue(w.defaultSkillRuntime) {
		return "", "", false
	}
	skillRuntime := normalizeSetupRuntimeSelection(w.defaultSkillRuntime, recommendedSetupRuntime())
	schedulerRuntime := normalizeScheduleRuntime(w.defaultScheduler)
	if schedulerRuntime == "auto" {
		schedulerRuntime = scheduleRuntimeForSetupSelection(skillRuntime, recommendedScheduleRuntime())
	}
	if advancedSetupRuntime(skillRuntime) || explicitRuntimeValue(w.defaultScheduler) && schedulerRuntime != firstRuntime(skillRuntime) {
		return skillRuntime, schedulerRuntime, true
	}
	return "", "", false
}

func (w *setupWizard) defaultAgentRuntime() string {
	schedulerRuntime := normalizeScheduleRuntime(w.defaultScheduler)
	if schedulerRuntime != "auto" {
		return schedulerRuntime
	}

	skillRuntime := normalizeSetupRuntimeSelection(w.defaultSkillRuntime, recommendedSetupRuntime())
	switch skillRuntime {
	case "all", "detected", "none":
		return recommendedScheduleRuntime()
	}
	return firstRuntime(skillRuntime)
}

func explicitRuntimeValue(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	return raw != "" && raw != "auto"
}

func advancedSetupRuntime(selection string) bool {
	return selection == "all" || selection == "none" || strings.Contains(selection, ",")
}

func runtimeSelectionLabel(selection string) string {
	switch selection {
	case "all":
		return "all supported runtimes"
	case "none":
		return "none"
	case "manual":
		return runtimeLabel(selection)
	}
	var labels []string
	for _, key := range normalizeRuntimeList(selection) {
		labels = append(labels, runtimeLabel(key))
	}
	return strings.Join(labels, ", ")
}

func scheduleRuntimeForSetupSelection(selection, fallback string) string {
	if selection == "all" || selection == "none" {
		return fallback
	}
	return firstRuntime(selection)
}

func singleAgentRuntime(raw string) (string, bool) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, "claude-code", "claude")
	raw = strings.ReplaceAll(raw, "claude_code", "claude")
	raw = strings.ReplaceAll(raw, "cladue", "claude")
	if raw == "other" || raw == "manual" {
		return "manual", true
	}
	if isSupportedRuntime(raw) {
		return raw, true
	}
	return "", false
}

func (w *setupWizard) promptForXAPIKey() error {
	if bearerToken(configFromEnv()) != "" {
		fmt.Fprintln(w.stdout)
		uiSuccess(w.stdout, "X API bearer token already configured.")
		return nil
	}
	fmt.Fprintln(w.stdout)
	uiSectionExact(w.stdout, "Configure X API (Optional)")
	uiKV(w.stdout, "used for", "X News, X trends, and X post search")
	uiKV(w.stdout, "stored in", "~/.newsjack/.env with user-only permissions")
	uiKV(w.stdout, "get one", xAPIKeyURL)
	value := strings.TrimSpace(w.prompt("X bearer token (press Enter to skip)", ""))
	if isSkipValue(value) {
		uiNote(w.stdout, "skipped X API key.")
		return nil
	}
	if err := writeNewsjackEnv(map[string]string{"X_BEARER_TOKEN": value}); err != nil {
		return err
	}
	uiSuccess(w.stdout, "saved X API bearer token to ~/.newsjack/.env.")
	return nil
}

func (w *setupWizard) promptForMedialystAPIKey() error {
	if medialystConfigured() {
		fmt.Fprintln(w.stdout)
		uiSuccess(w.stdout, "Medialyst API key already configured.")
		return nil
	}
	fmt.Fprintln(w.stdout)
	uiSectionExact(w.stdout, "Configure Medialyst API (Optional)")
	uiKV(w.stdout, "used for", "live news search, media database, find journalists")
	uiKV(w.stdout, "stored in", "~/.newsjack/credentials.json")
	uiKV(w.stdout, "get one", medialystAPIKeyURL)
	value := strings.TrimSpace(w.prompt("Medialyst API key (press Enter to skip)", ""))
	if isSkipValue(value) {
		uiNote(w.stdout, "skipped Medialyst API key.")
		return nil
	}
	if err := validateAPIKey(value); err != nil {
		return fmt.Errorf("invalid Medialyst API key: %w", err)
	}
	path, err := writeCredentials(value)
	if err != nil {
		return err
	}
	uiSuccess(w.stdout, "saved Medialyst credentials to %s.", path)
	return nil
}

func (w *setupWizard) prompt(message, defaultValue string) string {
	if w.assumeYes && defaultValue != "" {
		fmt.Fprintf(w.stdout, "%s %s [%s]: %s\n", uiQuestion(w.stdout), message, defaultValue, defaultValue)
		return defaultValue
	}
	if defaultValue != "" {
		fmt.Fprintf(w.stdout, "%s %s [%s]: ", uiQuestion(w.stdout), message, defaultValue)
	} else {
		fmt.Fprintf(w.stdout, "%s %s: ", uiQuestion(w.stdout), message)
	}
	line, err := w.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return defaultValue
	}
	if errors.Is(err, io.EOF) {
		fmt.Fprintln(w.stdout)
	}
	if err == nil {
		w.interactive = true
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue
	}
	return value
}

func (w *setupWizard) confirm(message string, defaultYes bool) bool {
	if w.assumeYes {
		fmt.Fprintf(w.stdout, "%s yes\n", message)
		return true
	}
	defaultLabel := "y/N"
	if defaultYes {
		defaultLabel = "Y/n"
	}
	answer := strings.ToLower(strings.TrimSpace(w.prompt(message+" ["+defaultLabel+"]", "")))
	if answer == "" {
		return defaultYes && w.interactive
	}
	return answer == "y" || answer == "yes"
}

func setupPayload(runtimeRaw string) map[string]any {
	recommended := selectSetupRuntime(runtimeRaw)
	scheduleRuntime := recommendedScheduleRuntime()
	agentPrompt := setupAgentPrompt(scheduleRuntime)
	return map[string]any{
		"newsjack_home":             newsjackHome(),
		"monitors_dir":              monitorsDir(),
		"runtimes":                  runtimeStatus(),
		"recommended_runtime":       recommended,
		"recommended_runtime_label": runtimeLabel(recommended),
		"recommended_scheduler":     scheduleRuntime,
		"scheduler_label":           runtimeLabel(scheduleRuntime),
		"claude_installed":          claudeCodeInstalled(),
		"claude_install_command":    claudeInstallCommand(),
		"auth": map[string]any{
			"medialyst_configured": medialystConfigured(),
			"x_api_configured":     bearerToken(configFromEnv()) != "",
		},
		"agent_command": setupAgentCommand(scheduleRuntime, agentPrompt),
		"agent_prompt":  agentPrompt,
	}
}

// setupLaunchPlan describes how to open an interactive setup session with an
// agent so it can ask follow-up questions and the user can answer in place.
type setupLaunchPlan struct {
	// args is the command to exec for an interactive session. It is empty when
	// the runtime has no interactive terminal mode, in which case setup falls
	// back to printing the prompt for the user to send manually.
	args []string
	// seeded reports whether args already include the setup prompt so the agent
	// starts on it. When false, the runtime can't seed a prompt while staying
	// interactive, so the prompt is printed for the user to send first.
	seeded bool
}

func setupLaunchPlanFor(runtime, prompt string) setupLaunchPlan {
	switch runtime {
	case "claude":
		// claude <prompt> seeds the first message and stays interactive.
		return setupLaunchPlan{args: []string{"claude", prompt}, seeded: true}
	case "codex":
		// codex <prompt> opens the interactive TUI seeded with the prompt.
		return setupLaunchPlan{args: []string{"codex", prompt}, seeded: true}
	case "hermes":
		// hermes chat is the interactive REPL. `hermes chat --query` is one-shot
		// and exits after the first turn, and there is no flag to seed a prompt
		// while staying interactive, so launch the REPL and print the prompt.
		return setupLaunchPlan{args: []string{"hermes", "chat"}, seeded: false}
	default:
		// openclaw's agent command is one-shot only (no interactive REPL), so a
		// follow-up question would end the session. Fall back to a manual prompt.
		return setupLaunchPlan{}
	}
}

func setupAgentCommand(runtime, prompt string) string {
	plan := setupLaunchPlanFor(runtime, prompt)
	if len(plan.args) == 0 {
		return ""
	}
	parts := make([]string, len(plan.args))
	for i, arg := range plan.args {
		parts[i] = shellQuote(arg)
	}
	return strings.Join(parts, " ")
}

func launchSetupAgent(runtime string, plan setupLaunchPlan, prompt string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(plan.args) == 0 {
		return nil
	}
	// Resolve through the runtime prober so Windows Claude Code installs
	// (which are not on PATH) still launch.
	resolved := ""
	if rt, ok := runtimeByKey(runtime); ok {
		if path, found := resolveRuntimeBinary(rt); found && rt.Binary == plan.args[0] {
			resolved = path
		}
	}
	if resolved == "" {
		path, err := exec.LookPath(plan.args[0])
		if err != nil {
			fmt.Fprintf(stdout, "%s is not on PATH. Run this command when it is available:\n%s\n", plan.args[0], setupAgentCommand(runtime, prompt))
			printSetupContinuationPrompt(stdout, runtime, prompt)
			return nil
		}
		resolved = path
	}
	plan.args = append([]string{resolved}, plan.args[1:]...)
	if !plan.seeded {
		// The runtime opens an interactive session but can't seed a prompt, so
		// show it for the user to send as their first message once the agent is up.
		uiNote(stdout, "%s opens an interactive session — send this prompt first:", runtimeLabel(runtime))
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, prompt)
		fmt.Fprintln(stdout)
	}
	cmd := exec.Command(plan.args[0], plan.args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		uiWarn(stdout, "auto-setup did not complete in %s.", runtimeLabel(runtime))
		uiNote(stdout, "agent command failed: %v", err)
		printSetupContinuationPrompt(stdout, runtime, prompt)
		return nil
	}
	return nil
}

func setupAgentPrompt(scheduleRuntime string) string {
	// Name the skill explicitly so every runtime (including hermes, which can't
	// seed the prompt and only prints it for the user to paste) reliably invokes
	// newsjack-monitor-setup instead of guessing at the right skill.
	return "Use the newsjack-monitor-setup skill to set up newsjack monitoring for my company."
}

func printSetupContinuationPrompt(stdout io.Writer, runtime, prompt string) {
	fmt.Fprintln(stdout)
	uiSection(stdout, "continue setup")
	uiNote(stdout, "open %s and send this prompt.", runtimeLabel(runtime))
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Prompt:")
	fmt.Fprintln(stdout, prompt)
	fmt.Fprintln(stdout)
}

func printRuntimeList(stdout io.Writer) {
	for _, rt := range runtimeTargets {
		state := "not detected"
		if runtimeDetected(rt) {
			state = "detected"
		}
		fmt.Fprintf(stdout, "  %-10s %-14s %s\n", rt.Key, rt.Label, state)
	}
	fmt.Fprintf(stdout, "  %-10s %-14s %s\n", "other", "Manual", "copy instructions")
}

func normalizeSetupRuntimeSelection(raw, fallback string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || raw == "auto" {
		return fallback
	}
	raw = strings.ReplaceAll(raw, "claude-code", "claude")
	raw = strings.ReplaceAll(raw, "claude_code", "claude")
	raw = strings.ReplaceAll(raw, "cladue", "claude")
	switch raw {
	case "other", "manual":
		return "manual"
	case "detected":
		var detected []string
		for _, rt := range runtimeTargets {
			if runtimeDetected(rt) {
				detected = append(detected, rt.Key)
			}
		}
		if len(detected) == 0 {
			return fallback
		}
		return strings.Join(detected, ",")
	case "all", "none":
		return raw
	}
	var out []string
	for _, part := range normalizeRuntimeList(raw) {
		if isSupportedRuntime(part) || isManualRuntime(part) {
			if part == "other" {
				part = "manual"
			}
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return strings.Join(out, ",")
}

func nonManualRuntimeSelection(selection string) string {
	if selection == "all" || selection == "none" {
		return selection
	}
	var out []string
	for _, item := range normalizeRuntimeList(selection) {
		if isSupportedRuntime(item) {
			out = append(out, item)
		}
	}
	return strings.Join(out, ",")
}

func firstRuntime(selection string) string {
	if selection == "all" || selection == "auto" || selection == "" {
		return recommendedSetupRuntime()
	}
	for _, item := range normalizeRuntimeList(selection) {
		if isSupportedRuntime(item) || isManualRuntime(item) {
			return item
		}
	}
	return recommendedSetupRuntime()
}

func setupAndScheduleRuntime(raw, fallback string) (string, string) {
	skillRuntime := normalizeSetupRuntimeSelection(raw, fallback)
	if schedulerRuntime, ok := singleAgentRuntime(raw); ok {
		return skillRuntime, schedulerRuntime
	}
	return skillRuntime, scheduleRuntimeForSetupSelection(skillRuntime, fallback)
}

func normalizeScheduleRuntime(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, "claude-code", "claude")
	raw = strings.ReplaceAll(raw, "claude_code", "claude")
	raw = strings.ReplaceAll(raw, "cladue", "claude")
	switch raw {
	case "", "auto":
		return "auto"
	case "other", "manual":
		return "manual"
	case "hermes", "openclaw", "claude", "codex":
		return raw
	default:
		return "manual"
	}
}

func recommendedScheduleRuntime() string {
	for _, key := range []string{"hermes", "openclaw", "claude", "codex"} {
		for _, target := range runtimeTargets {
			if target.Key == key && commandAvailable(target.Binary) {
				return key
			}
		}
	}
	return "claude"
}

func recommendedSetupRuntime() string {
	return recommendedScheduleRuntime()
}

func runtimeChoiceHint(key string) string {
	for _, target := range runtimeTargets {
		if target.Key == key {
			if commandAvailable(target.Binary) {
				if key == recommendedScheduleRuntime() {
					return "recommended"
				}
				return "detected"
			}
			return "not detected"
		}
	}
	return ""
}

func isSupportedRuntime(key string) bool {
	for _, rt := range runtimeTargets {
		if rt.Key == key {
			return true
		}
	}
	return false
}

func isManualRuntime(key string) bool {
	return key == "manual" || key == "other"
}

func runtimeSelectionIncludes(selection, key string) bool {
	if selection == "all" {
		return isSupportedRuntime(key)
	}
	for _, item := range normalizeRuntimeList(selection) {
		if item == key {
			return true
		}
	}
	return false
}

func installSetupSkills(runtimeRaw string, stdout, stderr io.Writer) error {
	root, err := newsjackRoot()
	if err != nil {
		return err
	}
	opts := installOptions{
		Source:   root,
		Runtimes: runtimeRaw,
		Force:    true,
		CLI:      newsjackCLIInvocation(),
		Repo:     getenv("NEWSJACK_REPO", defaultRepo),
		Ref:      getenv("NEWSJACK_REF", defaultRef),
	}
	return installRuntimeSkills(opts, stdout, stderr)
}

func manualSkillInstallInstruction() string {
	root, err := newsjackRoot()
	if err != nil || root == "" {
		root = filepath.Join(newsjackHome(), "newsjack")
	}
	return fmt.Sprintf("Copy every directory under `%s` into your agent runtime's skills directory, preserving each `SKILL.md`.", filepath.Join(root, "skills"))
}

func isSkipValue(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "" || value == "skip" || value == "s"
}

func writeNewsjackEnv(updates map[string]string) error {
	path := newsjackEnvPath()
	values := map[string]string{}
	if data, err := os.ReadFile(path); err == nil {
		for _, raw := range strings.Split(string(data), "\n") {
			line := strings.TrimSpace(raw)
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			k = strings.TrimSpace(k)
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if k != "" {
				values[k] = v
			}
		}
	}
	for k, v := range updates {
		if strings.TrimSpace(v) != "" {
			values[k] = strings.TrimSpace(v)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# Newsjack local credentials. Do not commit this file.\n")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(dotenvQuote(values[k]))
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func dotenvQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return `"` + value + `"`
}

func selectSetupRuntime(raw string) string {
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "auto" {
		return recommendedSetupRuntime()
	}
	list := normalizeRuntimeList(raw)
	for _, item := range list {
		switch item {
		case "claude", "codex", "openclaw", "hermes", "manual":
			return item
		case "none":
			return "manual"
		}
	}
	return "claude"
}

func runtimeLabel(key string) string {
	for _, target := range runtimeTargets {
		if target.Key == key {
			return target.Label
		}
	}
	if key == "manual" {
		return "Manual agent harness"
	}
	return key
}

func claudeCodeInstalled() bool {
	return commandAvailable("claude")
}

func claudeInstallCommand() string {
	return runtimeInstallCommand("claude")
}

func (w *setupWizard) offerRuntimeInstalls(selection, purpose string) error {
	if selection == "" || selection == "none" {
		return nil
	}
	for _, rt := range selectedRuntimes(selection) {
		if _, err := w.ensureRuntimeInstalled(rt.Key, purpose); err != nil {
			return err
		}
	}
	return nil
}

func (w *setupWizard) ensureRuntimeInstalled(key, purpose string) (bool, error) {
	if isManualRuntime(key) {
		return true, nil
	}
	rt, ok := runtimeTargetForKey(key)
	if !ok {
		return false, nil
	}
	if runtimeCLIInstalled(rt) {
		return true, nil
	}
	if w.runtimeInstallSeen == nil {
		w.runtimeInstallSeen = map[string]bool{}
	}
	if w.runtimeInstallSeen[key] {
		return false, nil
	}
	w.runtimeInstallSeen[key] = true

	label := runtimeLabel(key)
	uiWarn(w.stdout, "%s is selected for %s but is not installed.", label, purpose)
	w.printRuntimeInstallCommand(key)
	if key == "claude" && w.skipClaudeInstall {
		uiNote(w.stdout, "skipped Claude Code install prompt.")
		return false, nil
	}
	autoInstall := w.assumeYes || (key == "claude" && w.installClaude)
	if !autoInstall && !w.confirm(fmt.Sprintf("Install %s now?", label), false) {
		uiNote(w.stdout, "skipped %s install.", label)
		return false, nil
	}

	command := runtimeInstallCommand(key)
	if command == "" {
		return false, nil
	}
	fmt.Fprintln(w.stdout)
	uiInfo(w.stdout, "Installing %s...", label)
	cmd := shellCommand(command)
	cmd.Stdin = w.stdin
	cmd.Stdout = w.stdout
	cmd.Stderr = w.stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("%s install failed: %w", label, err)
	}
	if !runtimeCLIInstalled(rt) {
		uiNote(w.stdout, "%s installer finished, but `%s` is not on PATH yet. Open a new terminal or update PATH before launching it.", label, rt.Binary)
		return false, nil
	}
	return true, nil
}

func (w *setupWizard) printRuntimeInstallCommand(key string) {
	command := runtimeInstallCommand(key)
	if command == "" {
		uiNote(w.stdout, "install %s, then rerun newsjack setup.", runtimeLabel(key))
		return
	}
	uiNote(w.stdout, "to install %s, run:", runtimeLabel(key))
	fmt.Fprintf(w.stdout, "  %s\n", command)
}

func runtimeTargetForKey(key string) (runtimeTarget, bool) {
	for _, target := range runtimeTargets {
		if target.Key == key {
			return target, true
		}
	}
	return runtimeTarget{}, false
}

// shellCommand wraps a runtime install command in the platform shell:
// bash on Unix, PowerShell on Windows.
func shellCommand(command string) *exec.Cmd {
	if goos() == "windows" {
		return exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", command)
	}
	return exec.Command("bash", "-c", command)
}

func runtimeInstallCommand(key string) string {
	if cmd := strings.TrimSpace(os.Getenv(runtimeInstallCommandEnv(key))); cmd != "" {
		return cmd
	}
	switch key {
	case "codex":
		return "npm install -g @openai/codex"
	case "claude":
		if goos() == "windows" {
			return defaultClaudeInstallCommandWindows
		}
		return defaultClaudeInstallCommand
	case "openclaw":
		return "npm install -g openclaw"
	case "hermes":
		return "npm install -g hermes-agent"
	default:
		return ""
	}
}

func runtimeInstallCommandEnv(key string) string {
	switch key {
	case "codex":
		return "NEWSJACK_CODEX_INSTALL_COMMAND"
	case "claude":
		return "NEWSJACK_CLAUDE_INSTALL_COMMAND"
	case "openclaw":
		return "NEWSJACK_OPENCLAW_INSTALL_COMMAND"
	case "hermes":
		return "NEWSJACK_HERMES_INSTALL_COMMAND"
	default:
		return ""
	}
}

func medialystConfigured() bool {
	key, _ := loadAPIKey()
	return key != ""
}

func cmdMonitor(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, errors.New("usage: newsjack monitor init|test|run|schedule|status|open|brief"))
	}
	switch args[0] {
	case "init":
		return cmdMonitorInit(args[1:], stdout, stderr)
	case "test":
		return cmdMonitorTest(args[1:], stdout, stderr)
	case "run":
		return cmdMonitorRun(args[1:], stdout, stderr)
	case "schedule":
		return cmdMonitorSchedule(args[1:], stdout, stderr)
	case "status":
		return cmdMonitorStatus(args[1:], stdout, stderr)
	case "open":
		return cmdMonitorOpen(args[1:], stdout, stderr)
	case "brief":
		return cmdMonitorBrief(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown monitor command: %s", args[0])
	}
}

func cmdMonitorInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profilePath := fs.String("profile", "", "Profile JSON to save")
	force := fs.Bool("force", false, "Overwrite an existing monitor profile")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"profile"}))); err != nil {
		return 2
	}
	if *profilePath == "" {
		return fail(stderr, errors.New("usage: newsjack monitor init [slug] --profile profile.json"))
	}
	payload, err := readJSONMap(*profilePath)
	if err != nil {
		return fail(stderr, err)
	}
	profile := profileFromMap(payload)
	slug := ""
	if fs.NArg() > 0 {
		slug = fs.Arg(0)
	}
	if slug == "" {
		slug = slugify(firstString(profile.Company, "monitor"))
	}
	slug = slugify(slug)
	if slug == "" {
		return fail(stderr, errors.New("monitor slug is empty"))
	}
	dir := monitorDir(slug)
	dest := monitorProfilePath(slug)
	if fileExists(dest) && !*force {
		return failf(stderr, "monitor already exists: %s", slug)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(dest, marshalJSON(payload), 0o644); err != nil {
		return fail(stderr, err)
	}
	// Scaffold a starter brief (client pitch/output policy) next to the profile.
	// Never clobber a user-edited brief, even with --force: it is hand-authored policy.
	briefPath := monitorBriefPath(slug)
	if !fileExists(briefPath) {
		if err := os.WriteFile(briefPath, []byte(briefTemplate(profile)), 0o644); err != nil {
			return fail(stderr, err)
		}
	}
	writeJSON(stdout, map[string]any{"slug": slug, "profile_path": dest, "monitor_dir": dir, "brief_path": briefPath})
	return 0
}

func cmdMonitorTest(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor test", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mock := fs.Bool("mock", false, "Run deterministic mock sources")
	live := fs.Bool("live", false, "Run available live sources")
	limit := fs.Int("limit", 20, "Maximum emitted signals")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"limit"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor test <slug> --mock|--live"))
	}
	if *mock && *live {
		return fail(stderr, errors.New("choose only one of --mock or --live"))
	}
	runMock := true
	if *live {
		runMock = false
	}
	result, err := runMonitor(fs.Arg(0), runMock, true, *limit)
	if err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, result)
	return 0
}

func cmdMonitorRun(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mock := fs.Bool("mock", false, "Run deterministic mock sources")
	limit := fs.Int("limit", 20, "Maximum emitted signals")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"limit"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor run <slug>"))
	}
	result, err := runMonitor(fs.Arg(0), *mock, false, *limit)
	if err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, result)
	return 0
}

func cmdMonitorSchedule(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor schedule", flag.ContinueOnError)
	fs.SetOutput(stderr)
	runtimeRaw := fs.String("runtime", "auto", "Agent runtime: auto, openclaw, hermes, claude, codex")
	every := fs.String("every", "1h", "Schedule interval")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"runtime", "every"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor schedule <slug> --runtime <agent-runtime> --every 1h"))
	}
	slug := fs.Arg(0)
	if !fileExists(monitorProfilePath(slug)) {
		return failf(stderr, "monitor profile not found: %s", slug)
	}
	runtime := selectScheduleRuntime(*runtimeRaw)
	dir := monitorDir(slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail(stderr, err)
	}
	suggestedMinute := suggestedScheduleMinute(slug)
	payload := map[string]any{
		"slug":             slug,
		"runtime":          runtime,
		"every":            *every,
		"suggested_minute": suggestedMinute,
		"system_cron":      false,
		"schedule_path":    monitorSchedulePath(slug),
		"instructions":     scheduleInstructions(slug, runtime, *every),
		"installed_at":     time.Now().UTC().Format(time.RFC3339Nano),
		"run_command":      fmt.Sprintf("newsjack monitor run %s", shellQuote(slug)),
		"artifact_scope":   "agent_harness",
	}
	if err := os.WriteFile(monitorScheduleJSONPath(slug), marshalJSON(payload), 0o644); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(monitorSchedulePath(slug), []byte(renderScheduleMarkdown(payload)), 0o644); err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, payload)
	return 0
}

func cmdMonitorStatus(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor status <slug>"))
	}
	slug := args[0]
	payload := monitorStatus(slug)
	writeJSON(stdout, payload)
	if !truthy(payload["exists"], false) {
		return 1
	}
	return 0
}

func cmdMonitorOpen(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor open <slug>"))
	}
	slug := args[0]
	status := monitorStatus(slug)
	if !truthy(status["exists"], false) {
		return failf(stderr, "monitor not found: %s", slug)
	}
	if latest := stringValue(status["latest_report_path"]); latest != "" {
		fmt.Fprintln(stdout, latest)
		return 0
	}
	fmt.Fprintln(stdout, monitorDir(slug))
	return 0
}

func cmdMonitorBrief(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor brief", flag.ContinueOnError)
	fs.SetOutput(stderr)
	edit := fs.Bool("edit", false, "Open the brief in $EDITOR")
	jsonOut := fs.Bool("json", false, "Emit brief metadata as JSON instead of contents")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor brief <slug> [--edit|--json]"))
	}
	slug := fs.Arg(0)
	if !fileExists(monitorProfilePath(slug)) {
		return failf(stderr, "monitor not found: %s", slug)
	}
	path := monitorBriefPath(slug)
	if *edit {
		editor := firstString(os.Getenv("NEWSJACK_EDITOR"), os.Getenv("VISUAL"), os.Getenv("EDITOR"), "vi")
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fail(stderr, err)
		}
		return 0
	}
	if *jsonOut {
		writeJSON(stdout, map[string]any{"slug": slug, "brief_path": path, "brief_present": fileExists(path)})
		return 0
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return failf(stderr, "no brief found for %s (expected %s)", slug, path)
	}
	_, _ = stdout.Write(data)
	return 0
}

func runMonitor(slug string, mock, test bool, limit int) (map[string]any, error) {
	profilePath := monitorProfilePath(slug)
	if !fileExists(profilePath) {
		return nil, fmt.Errorf("monitor profile not found: %s", slug)
	}
	profile, err := profileFromFile(profilePath)
	if err != nil {
		return nil, err
	}
	runDir := monitorRunDir(slug)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, err
	}
	paths := artifactPaths(runDir)
	opts := detectorOptions{
		ProfilePath:                     profilePath,
		Depth:                           "quick",
		LookbackDays:                    1,
		MaxAgeHours:                     24,
		MinQueuePriority:                defaultMinQueuePriority,
		MinMajorNews:                    defaultMinMajorNews,
		Limit:                           limit,
		Mock:                            mock,
		Save:                            !mock,
		NewOnly:                         !mock,
		MonitorName:                     slug,
		IncludeAllScored:                true,
		MajorFeeds:                      !mock,
		NoHygieneFilter:                 false,
		NoProfileFeeds:                  false,
		XNewsMinProfileMatch:            0.05,
		XPostsMinProfileMatch:           0.08,
		ProfileRelevanceMinProfileMatch: 0.05,
		MajorNewsMinProfileMatch:        0.05,
		XTrendsMinProfileMatch:          0.05,
	}
	if !mock && shouldRunFeedOnly(profile, opts) {
		opts.FeedOnly = true
	}
	command := monitorDetectorCommand(slug, opts)
	if err := os.WriteFile(paths["commands"], []byte(command+"\n"), 0o644); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = detectorRun(opts, &out)
	if err != nil {
		_ = os.WriteFile(paths["detector_stderr"], []byte(err.Error()+"\n"), 0o644)
		return nil, err
	}
	if err := os.WriteFile(paths["detector_stderr"], nil, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths["candidates"], out.Bytes(), 0o644); err != nil {
		return nil, err
	}
	payload, err := readJSONMap(paths["candidates"])
	if err != nil {
		return nil, err
	}
	summaryGeneratedAt := time.Now().UTC().Format(time.RFC3339Nano)
	summary := summarizeRunAt(payload, paths["candidates"], 25, summaryGeneratedAt)
	if err := os.WriteFile(paths["detector_summary"], marshalJSON(summary), 0o644); err != nil {
		return nil, err
	}
	summary = summarizeRunAt(payload, paths["candidates"], 25, summaryGeneratedAt)
	if err := os.WriteFile(paths["detector_summary"], marshalJSON(summary), 0o644); err != nil {
		return nil, err
	}
	return map[string]any{
		"slug":          slug,
		"mock":          mock,
		"test":          test,
		"feed_only":     opts.FeedOnly,
		"run_dir":       runDir,
		"candidates":    paths["candidates"],
		"summary":       paths["detector_summary"],
		"report_target": paths["run_report"],
		"brief_path":    monitorBriefPath(slug),
		"brief_present": fileExists(monitorBriefPath(slug)),
	}, nil
}

func shouldRunFeedOnly(profile monitorProfile, opts detectorOptions) bool {
	requested, err := requestedSourcesFor(opts, profile)
	if err != nil {
		return false
	}
	if len(availableSources(configFromEnv(), querySources(requested))) > 0 {
		return false
	}
	return len(profile.FeedURLs) > 0 || opts.MajorFeeds
}

func monitorDetectorCommand(slug string, opts detectorOptions) string {
	parts := []string{"newsjack", "detector", "run", "--profile", monitorProfilePath(slug), "--monitor-name", slug, "--depth", opts.Depth, "--limit", fmt.Sprint(opts.Limit)}
	if opts.Mock {
		parts = append(parts, "--mock")
	}
	if opts.FeedOnly {
		parts = append(parts, "--feed-only")
	}
	if opts.MajorFeeds {
		parts = append(parts, "--major-feeds")
	}
	if opts.Save {
		parts = append(parts, "--save")
	}
	if opts.NewOnly {
		parts = append(parts, "--new-only")
	}
	return strings.Join(parts, " ")
}

func monitorStatus(slug string) map[string]any {
	dir := monitorDir(slug)
	runs := monitorRuns(slug)
	latestRun := ""
	latestReport := ""
	if len(runs) > 0 {
		latestRun = runs[len(runs)-1]
		if fileExists(filepath.Join(latestRun, "run.md")) {
			latestReport = filepath.Join(latestRun, "run.md")
		}
	}
	return map[string]any{
		"slug":               slug,
		"exists":             fileExists(monitorProfilePath(slug)),
		"monitor_dir":        dir,
		"profile_path":       monitorProfilePath(slug),
		"brief_path":         monitorBriefPath(slug),
		"brief_present":      fileExists(monitorBriefPath(slug)),
		"schedule_path":      nullableStringIfExists(monitorSchedulePath(slug)),
		"schedule_json_path": nullableStringIfExists(monitorScheduleJSONPath(slug)),
		"run_count":          len(runs),
		"latest_run_dir":     nullableString(latestRun),
		"latest_report_path": nullableString(latestReport),
	}
}

func monitorRuns(slug string) []string {
	root := filepath.Join(monitorDir(slug), "runs")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var runs []string
	for _, entry := range entries {
		if entry.IsDir() {
			runs = append(runs, filepath.Join(root, entry.Name()))
		}
	}
	sort.Strings(runs)
	return runs
}

func monitorsDir() string { return filepath.Join(newsjackHome(), "monitors") }

func monitorDir(slug string) string { return filepath.Join(monitorsDir(), slugify(slug)) }

func monitorProfilePath(slug string) string { return filepath.Join(monitorDir(slug), "profile.json") }

func monitorBriefPath(slug string) string { return filepath.Join(monitorDir(slug), "brief.md") }

// briefTemplate is the starter pitch/output policy scaffolded at monitor init.
// It is intentionally inert: the section bodies are HTML comments, so an
// untouched brief carries no active rules and the pipeline keeps default
// behavior. The CLI never parses this file — it only creates and surfaces it;
// the detector and triage skills read it and update it from user feedback.
func briefTemplate(profile monitorProfile) string {
	company := firstString(profile.Company, "This client")
	return fmt.Sprintf(`# %s — Brief

This file is the source of truth for what %s will and won't pitch, and how the
newsjack scan should be presented. It is prose on purpose — edit it like a note.
The detector reads it when judging standing and rendering your report, and
should update it when you give feedback about what you do and don't want.

Delete a comment and write under the heading to activate that section. An
untouched brief changes nothing.

## Audience
<!-- Who is reached, ultimately? e.g. "regular consumers", "CISOs at mid-market SaaS". This sets the altitude of an acceptable pitch. -->

## We pitch
<!-- The concrete kinds of stories that are fair game. Be specific. -->

## We never pitch
<!-- Hard rules: a story matching one can never be pitch-ready. A genuinely big story stays surfaced as awareness-only (never hidden), just not pitched. e.g. "policy/legislation process", "a competitor's own press releases". -->

## How to surface
<!-- Presentation preferences. e.g. "lead with pitch-ready only", "collapse the big-stories/awareness section to a one-line count". -->

## Examples
<!-- Good/bad examples, each dated. This doubles as the feedback log: when you reject something, it lands here. -->
`, company, company)
}

func monitorSchedulePath(slug string) string { return filepath.Join(monitorDir(slug), "schedule.md") }

func monitorScheduleJSONPath(slug string) string {
	return filepath.Join(monitorDir(slug), "schedule.json")
}

func monitorRunDir(slug string) string {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	dir := filepath.Join(monitorDir(slug), "runs", stamp)
	if !dirExists(dir) {
		return dir
	}
	return filepath.Join(monitorDir(slug), "runs", time.Now().UTC().Format("20060102T150405.000000000Z"))
}

func selectScheduleRuntime(raw string) string {
	runtime := selectSetupRuntime(raw)
	switch runtime {
	case "openclaw", "hermes", "claude", "codex":
		return runtime
	default:
		return "manual"
	}
}

func suggestedScheduleMinute(slug string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(slugify(slug)))
	return int(h.Sum32()%59) + 1
}

func scheduleInstructions(slug, runtime, every string) string {
	minute := suggestedScheduleMinute(slug)
	return fmt.Sprintf("Every %s, run `newsjack monitor run %s` inside %s, then use the installed newsjack-detector skill to complete LLM analysis and render run.md from the JSON artifacts. When installing the recurring schedule, pick a stable random minute in [1, 59], never minute 0; suggested_minute is %d from `minute = (fnv32a(slug) %% 59) + 1`. Apply the same jitter to daily and weekly schedules, and avoid common top-of-hour, midnight, and Monday-09:00 defaults. This spreads load across the Newsjack/Medialyst backend so we don't get a thundering-herd spike at the top of every hour.", every, shellQuote(slug), runtime, minute)
}

func renderScheduleMarkdown(payload map[string]any) string {
	return fmt.Sprintf("# Newsjack Agent Schedule\n\n"+
		"- Monitor: %s\n"+
		"- Runtime: %s\n"+
		"- Every: %s\n"+
		"- Suggested minute: %v\n"+
		"- System cron installed: false\n\n"+
		"%s\n\n"+
		"## Cron Jitter\n\n"+
		"Pick a stable random minute in [1, 59], never minute 0. Use\n"+
		"`minute = (fnv32a(slug) %% 59) + 1`; this schedule suggests minute %v for\n"+
		"`%s`. For hourly schedules, use that minute instead of the top of the hour.\n"+
		"For daily or weekly schedules, use the same minute rule and avoid common\n"+
		"midnight or Monday-09:00 defaults. This spreads load across the\n"+
		"Newsjack/Medialyst backend so we don't get a thundering-herd spike at the top\n"+
		"of every hour.\n\n"+
		"This is an agent-harness schedule contract. Do not install this monitor in\n"+
		"system cron for v1; the run must trigger the agent harness so the LLM analysis\n"+
		"step can operate on the detector artifacts.\n",
		payload["slug"], payload["runtime"], payload["every"], payload["suggested_minute"], payload["instructions"], payload["suggested_minute"], payload["slug"])
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "monitor"
	}
	return value
}

func nullableStringIfExists(path string) any {
	if fileExists(path) {
		return path
	}
	return nil
}

func shellQuote(value string) string {
	if regexp.MustCompile(`^[A-Za-z0-9._/-]+$`).MatchString(value) {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
