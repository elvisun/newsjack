package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func cmdSetup(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "Emit setup status as JSON")
	runtimeRaw := fs.String("runtime", "auto", "Preferred agent runtime")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	status := setupPayload(*runtimeRaw)
	if *jsonOut {
		writeJSON(stdout, status)
		return 0
	}
	fmt.Fprintln(stdout, "newsjack setup")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Home: %s\n", status["newsjack_home"])
	fmt.Fprintf(stdout, "Recommended runtime: %s\n", status["recommended_runtime"])
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Next, run this inside your agent harness:")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, stringValue(status["agent_prompt"]))
	return 0
}

func setupPayload(runtimeRaw string) map[string]any {
	recommended := selectSetupRuntime(runtimeRaw)
	return map[string]any{
		"newsjack_home":       newsjackHome(),
		"monitors_dir":        monitorsDir(),
		"runtimes":            runtimeStatus(),
		"recommended_runtime": recommended,
		"auth": map[string]any{
			"medialyst_configured": medialystConfigured(),
			"x_api_configured":     bearerToken(configFromEnv()) != "",
		},
		"agent_prompt": "Use Newsjack to set up an hourly monitor for my company, install the schedule in this agent harness, and run a mock test. Ask only for facts you cannot infer safely.",
	}
}

func selectSetupRuntime(raw string) string {
	list := normalizeRuntimeList(raw)
	for _, item := range list {
		if item != "auto" {
			return item
		}
	}
	for _, rt := range []string{"openclaw", "hermes", "claude", "codex"} {
		for _, target := range runtimeTargets {
			if target.Key == rt && runtimeDetected(target) {
				return rt
			}
		}
	}
	return "manual"
}

func medialystConfigured() bool {
	key, _ := loadAPIKey()
	return key != ""
}

func cmdMonitor(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, errors.New("usage: newsjack monitor init|test|run|schedule|status|open"))
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
	writeJSON(stdout, map[string]any{"slug": slug, "profile_path": dest, "monitor_dir": dir})
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
	payload := map[string]any{
		"slug":           slug,
		"runtime":        runtime,
		"every":          *every,
		"system_cron":    false,
		"schedule_path":  monitorSchedulePath(slug),
		"instructions":   scheduleInstructions(slug, runtime, *every),
		"installed_at":   time.Now().UTC().Format(time.RFC3339Nano),
		"run_command":    fmt.Sprintf("newsjack monitor run %s", shellQuote(slug)),
		"artifact_scope": "agent_harness",
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
	if latest := stringValue(status["latest_run_markdown"]); latest != "" {
		fmt.Fprintln(stdout, latest)
		return 0
	}
	fmt.Fprintln(stdout, monitorDir(slug))
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
		Emit:                            "json",
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
	summary := summarizeRun(payload, paths["candidates"], 25)
	if err := os.WriteFile(paths["detector_summary"], marshalJSON(summary), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths["run_markdown"], []byte(renderSummaryMarkdown(summary)), 0o644); err != nil {
		return nil, err
	}
	return map[string]any{
		"slug":         slug,
		"mock":         mock,
		"test":         test,
		"feed_only":    opts.FeedOnly,
		"run_dir":      runDir,
		"candidates":   paths["candidates"],
		"summary":      paths["detector_summary"],
		"run_markdown": paths["run_markdown"],
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
	parts := []string{"newsjack", "detector", "run", "--profile", monitorProfilePath(slug), "--monitor-name", slug, "--depth", opts.Depth, "--limit", fmt.Sprint(opts.Limit), "--emit", "json"}
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
	latestMD := ""
	if len(runs) > 0 {
		latestRun = runs[len(runs)-1]
		if fileExists(filepath.Join(latestRun, "run.md")) {
			latestMD = filepath.Join(latestRun, "run.md")
		}
	}
	return map[string]any{
		"slug":                slug,
		"exists":              fileExists(monitorProfilePath(slug)),
		"monitor_dir":         dir,
		"profile_path":        monitorProfilePath(slug),
		"schedule_path":       nullableStringIfExists(monitorSchedulePath(slug)),
		"schedule_json_path":  nullableStringIfExists(monitorScheduleJSONPath(slug)),
		"run_count":           len(runs),
		"latest_run_dir":      nullableString(latestRun),
		"latest_run_markdown": nullableString(latestMD),
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

func scheduleInstructions(slug, runtime, every string) string {
	return fmt.Sprintf("Every %s, run `newsjack monitor run %s` inside %s, then use the installed newsjack-detector skill to complete LLM analysis and rerender run.md.", every, shellQuote(slug), runtime)
}

func renderScheduleMarkdown(payload map[string]any) string {
	return fmt.Sprintf(`# Newsjack Agent Schedule

- Monitor: %s
- Runtime: %s
- Every: %s
- System cron installed: false

%s

This is an agent-harness schedule contract. Do not install this monitor in
system cron for v1; the run must trigger the agent harness so the LLM analysis
step can operate on the detector artifacts.
`, payload["slug"], payload["runtime"], payload["every"], payload["instructions"])
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
