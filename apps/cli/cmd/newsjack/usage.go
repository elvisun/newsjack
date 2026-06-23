package main

import (
	"fmt"
	"io"
)

func printUsage(w io.Writer) {
	uiProduct(w, "", "")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack <command> [flags]")
	fmt.Fprintln(w)
	uiSection(w, "commands")
	uiCommand(w, "help", "show this screen", "")
	uiCommand(w, "usage", "show this screen", "")
	uiCommand(w, "version", "print the installed version", "")
	uiCommand(w, "path", "print the install root", "")
	uiCommand(w, "doctor", "run a system health check", "[--json]")
	uiCommand(w, "setup", "guided install: runtimes, auth, skills", "[--yes]")
	uiCommand(w, "install", "install the skill bundle", "[--source DIR]")
	uiCommand(w, "skills [list]", "manage installed skills", "")
	uiCommand(w, "runtimes detect", "detect supported agent runtimes", "")
	uiCommand(w, "login", "connect Medialyst with browser OAuth", "[--no-browser]")
	uiCommand(w, "auth status|set|logout", "inspect, save, or revoke API credentials", "")
	uiCommand(w, "credits [balance]", "show Medialyst credit balance", "")
	uiCommand(w, "news search", "search current news through Medialyst", "--query Q")
	uiCommand(w, "journalists enrich", "enrich journalists from article URLs", "--url URL [--pitch TEXT]")
	uiCommand(w, "monitor init|test|run...", "manage newsjacking monitors", "")
	uiCommand(w, "coverage list|init|check...", "manage coverage trackers", "")
	uiCommand(w, "detector run|recent...", "angle detection over recent stories", "")
	uiCommand(w, "update", "pull the latest skill bundle", "")
	fmt.Fprintln(w)
	uiSection(w, "api setup")
	uiCommand(w, "login", "recommended Medialyst browser login for live news search and journalist enrichment", "")
	uiCommand(w, "auth set-medialyst", "API-key fallback for CI or automation", "--key KEY")
	uiCommand(w, "auth set-x", "save X bearer token for X News, trends, and post search", "--bearer-token TOKEN")
	uiKV(w, "Medialyst login", "newsjack login")
	uiKV(w, "Medialyst API key", medialystAPIKeyURL)
	uiKV(w, "X bearer token", xAPIKeyURL)
	uiNote(w, "Medialyst REST commands prefer saved OAuth, then API keys from ~/.newsjack/credentials.json or MEDIALYST_API_KEY.")
	fmt.Fprintln(w)
	uiSection(w, "pipeline")
	uiCommand(w, "filter-apply", "apply coarse-relevance decisions to candidates", "--candidates F --decisions F")
	uiCommand(w, "cluster", "collapse same-story pickups before retrieval", "--candidates F [--drop-stale]")
	uiCommand(w, "origin-apply", "apply the deterministic freshness gate", "--candidates F --origins F")
	uiCommand(w, "run-summary", "write deterministic run metadata as JSON", "INPUT [--output FILE]")
	fmt.Fprintln(w)
	uiSection(w, "learn")
	uiCommand(w, "newsjack help <command>", "detail on a single command", "")
	uiCommand(w, "newsjack doctor", "fastest way to know it works", "")
	fmt.Fprintln(w)
	uiNote(w, "start with: newsjack setup")
}

func printCommandHelp(w io.Writer, command string) bool {
	switch command {
	case "login":
		uiProduct(w, "login", "connect Medialyst with browser OAuth.")
		fmt.Fprintln(w)
		uiSection(w, "usage")
		fmt.Fprintln(w, "  newsjack login [--no-browser]")
		fmt.Fprintln(w, "  newsjack login --key <mlst_...>    # API-key fallback")
		fmt.Fprintln(w)
		uiSection(w, "flow")
		uiKV(w, "client", medialystOAuthClientID)
		uiKV(w, "scopes", medialystOAuthDefaultScope)
		uiKV(w, "what users do", "open the printed Medialyst link, approve newsjack CLI, then return to the agent")
		uiKV(w, "storage", "~/.newsjack/credentials.json")
		uiNote(w, "Agents should use this path for interactive setup. API keys remain supported for CI and automation.")
		return true
	case "auth":
		printAuthHelp(w)
		return true
	case "install":
		uiProduct(w, "install", "install skills into agent runtimes.")
		fmt.Fprintln(w)
		uiSection(w, "usage")
		fmt.Fprintln(w, "  newsjack install [--source <bundle-dir>] [--runtimes auto|all|none|codex,claude,openclaw,hermes] [--force]")
		fmt.Fprintln(w)
		uiSection(w, "options")
		uiCommand(w, "--source", "install from a local bundle dir; prebuilt bundles are adopted as the managed install", "<dir>")
		uiCommand(w, "--runtimes", "target runtimes; auto detects installed agent CLIs", "claude")
		uiCommand(w, "--force", "overwrite user-owned skill directories too", "")
		uiNote(w, "release overrides: NEWSJACK_VERSION pins the bundle tag; NEWSJACK_RELEASE_BASE points at any URL serving release assets.")
		return true
	case "skills":
		uiProduct(w, "skills", "inspect and manage installed runtime skills.")
		fmt.Fprintln(w)
		uiSection(w, "usage")
		fmt.Fprintln(w, "  newsjack skills list|install|status")
		fmt.Fprintln(w)
		uiCommand(w, "skills list", "list skills in the installed bundle", "")
		uiCommand(w, "skills install", "same as newsjack install", "")
		uiCommand(w, "skills status", "skills install health as JSON", "")
		return true
	case "doctor":
		uiProduct(w, "doctor", "checks install health and prints concrete recovery commands.")
		fmt.Fprintln(w)
		uiSection(w, "usage")
		fmt.Fprintln(w, "  newsjack doctor [--json]")
		fmt.Fprintln(w)
		uiSection(w, "api recovery")
		uiCommand(w, "login", "connect Medialyst with browser OAuth", "")
		uiCommand(w, "auth set-medialyst", "API-key fallback for CI or automation", "--key <mlst_...>")
		uiCommand(w, "auth set-x", "save X API bearer token", "--bearer-token <token>")
		uiNote(w, "doctor --json includes the same actions for agents that need machine-readable recovery steps.")
		return true
	case "setup":
		uiProduct(w, "setup", "guided install: runtimes, optional APIs, skills, and agent launch.")
		fmt.Fprintln(w)
		uiSection(w, "usage")
		fmt.Fprintln(w, "  newsjack setup [--schedule-runtime codex] [--skip-credentials] [--no-launch]")
		fmt.Fprintln(w)
		uiSection(w, "optional api commands")
		uiCommand(w, "login", "connect Medialyst without rerunning setup", "")
		uiCommand(w, "auth set-medialyst", "API-key fallback without rerunning setup", "--key <mlst_...>")
		uiCommand(w, "auth set-x", "save X without rerunning setup", "--bearer-token <token>")
		return true
	case "detector":
		printDetectorHelp(w)
		return true
	case "coverage":
		printCoverageHelp(w)
		return true
	case "credits":
		printCreditsHelp(w)
		return true
	case "news", "news search":
		printNewsHelp(w)
		return true
	case "journalists", "journalists enrich", "journalists enrich-job":
		printJournalistsHelp(w)
		return true
	default:
		return false
	}
}

func printCreditsHelp(w io.Writer) {
	uiProduct(w, "credits", "thin Medialyst REST wrapper for account credits.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack credits")
	fmt.Fprintln(w, "  newsjack credits balance")
	fmt.Fprintln(w)
	uiKV(w, "endpoint", "GET /api/v1/credits/balance")
}

func printNewsHelp(w io.Writer) {
	uiProduct(w, "news", "thin Medialyst REST wrapper for news search.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack news search --query \"AI infrastructure startup funding\" [--page 1] [--limit 10] [--tbs qdr:m]")
	fmt.Fprintln(w, "  newsjack news search --json '{\"q\":\"AI infrastructure\",\"gl\":\"us\",\"page\":1}'")
	fmt.Fprintln(w)
	uiSection(w, "mapping")
	uiKV(w, "news search", "POST /api/v1/news/search")
	uiNote(w, "Use --json or --json-file to send the exact API request body.")
}

func printJournalistsHelp(w io.Writer) {
	uiProduct(w, "journalists", "thin Medialyst REST wrapper for journalist enrichment.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack journalists enrich --url https://example.com/story --pitch \"why this journalist fits\" [--wait]")
	fmt.Fprintln(w, "  newsjack journalists enrich --json-file request.json")
	fmt.Fprintln(w, "  newsjack journalists enrich-job <job-id> [--wait]    # later revisit only")
	fmt.Fprintln(w)
	uiSection(w, "wait behavior")
	uiKV(w, "--wait", "POSTs with API wait=true, then polls the returned job only within the remaining foreground budget")
	uiKV(w, "--poll-timeout-ms", "total foreground wait budget for enrich jobs; default 45000, max 45000")
	uiKV(w, "--poll-interval-ms", "CLI polling interval; default 3000")
	fmt.Fprintln(w)
	uiSection(w, "mapping")
	uiKV(w, "enrich", "POST /api/v1/journalists/enrich")
	uiKV(w, "enrich-job", "GET /api/v1/journalist-enrichment-jobs/{jobId}")
	uiNote(w, "PR1024's enrich endpoint currently supports article_url sources. Use --json for the exact documented request shape.")
	uiNote(w, "With convenience flags, --wait accepts one --url at a time. Use --wait=false or --json for intentional API-shaped batch jobs.")
	uiNote(w, "Use enrich for a small number of high-confidence article URLs; do not batch-enrich broad news-search results.")
	uiNote(w, "If a job is still processing after the bounded wait, keep the job ID and revisit later instead of calling enrich-job immediately or writing your own polling loop.")
}

func printDetectorHelp(w io.Writer) {
	uiProduct(w, "detector", "collects news evidence and emits JSON candidates for agent judgment.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack detector run --profile profile.json --save")
	fmt.Fprintln(w, "  newsjack detector run --profile profile.json --mock")
	fmt.Fprintln(w, "  newsjack detector diagnose")
	fmt.Fprintln(w, "  newsjack detector recent")
	fmt.Fprintln(w)
	uiSection(w, "learn")
	uiCommand(w, "newsjack detector run --help", "authoritative run flags and defaults", "")
	uiCommand(w, "newsjack doctor", "credential and source health", "[--json]")
	fmt.Fprintln(w)
	uiSection(w, "sources")
	uiKV(w, "news_search", "primary live news search")
	uiKV(w, "x_news", "X story clusters; auto-included when enabled in profile")
	uiKV(w, "x", "raw X post search with reach filters")
	uiKV(w, "x_trends", "personalized or location trends from profile config")
	uiKV(w, "major_feed", "RSS/Atom feeds from profile or flags")
	uiKV(w, "reddit/hackernews", "optional v0 sources")
	fmt.Fprintln(w)
	uiSection(w, "profiles")
	uiKV(w, "feed_urls", "included unless --no-profile-feeds is set")
	uiKV(w, "search_terms", "if present, retrieval uses these instead of raw topics + competitors")
	uiKV(w, "topics", "broad profile meaning and matching context")
	uiKV(w, "competitors", "matching context; add to search_terms when they must drive retrieval")
	uiNote(w, "Use --topic only for an explicit one-off topic. Routine monitor runs should rely on the profile.")
}

func printCoverageHelp(w io.Writer) {
	uiProduct(w, "coverage", "stores keyword coverage tracker config and LLM alert state.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack coverage list")
	fmt.Fprintln(w, "  newsjack coverage init <slug> --config tracker.json")
	fmt.Fprintln(w, "  newsjack coverage status <slug>")
	fmt.Fprintln(w, "  newsjack coverage open <slug>")
	fmt.Fprintln(w, "  newsjack coverage check <slug> --input candidates.json")
	fmt.Fprintln(w, "  newsjack coverage record <slug> --input decisions.json")
	fmt.Fprintln(w)
	uiSection(w, "shape")
	uiKV(w, "list/init/status/open", "storage helpers for coverage-tracker-setup and coverage-tracker")
	uiKV(w, "check", "marks candidates already classified in SQLite so the LLM can skip repeats")
	uiKV(w, "record", "persists LLM-classified articles and suppresses repeat alerts")
	uiKV(w, "scheduling", "owned by Claude, Codex, Hermes, OpenClaw, or another agent harness")
	uiNote(w, "The coverage-tracker skill owns news-search, dedupe judgment, feature detection, and alert rendering.")
}

func printAuthHelp(w io.Writer) {
	uiProduct(w, "auth", "inspect and configure optional API credentials.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack login [--no-browser]")
	fmt.Fprintln(w, "  newsjack auth status")
	fmt.Fprintln(w, "  newsjack auth set --medialyst-key <mlst_...> --x-bearer-token <token>")
	fmt.Fprintln(w, "  newsjack auth set-medialyst --key <mlst_...>")
	fmt.Fprintln(w, "  newsjack auth set-x --bearer-token <token>")
	fmt.Fprintln(w)
	uiSection(w, "optional apis")
	uiKV(w, "Medialyst", "live news search and journalist enrichment")
	uiKV(w, "recommended login", "newsjack login")
	uiKV(w, "login behavior", "prints a Medialyst approval link, opens the browser when possible, and stores OAuth")
	uiKV(w, "OAuth storage", "~/.newsjack/credentials.json")
	uiKV(w, "API-key fallback", "newsjack auth set-medialyst --key <mlst_...>")
	uiKV(w, "get API key", medialystAPIKeyURL)
	uiKV(w, "API-key storage", "~/.newsjack/credentials.json or MEDIALYST_API_KEY")
	uiKV(w, "X API", "X News, X trends, and X post search")
	uiKV(w, "get token", xAPIKeyURL)
	uiKV(w, "save token", "newsjack auth set-x --bearer-token <token>")
}

func fail(w io.Writer, err error) int {
	uiError(w, "%v", err)
	return 1
}

func failf(w io.Writer, format string, args ...any) int {
	return fail(w, fmt.Errorf(format, args...))
}

func warn(w io.Writer, format string, args ...any) {
	uiWarn(w, format, args...)
}

func logf(w io.Writer, format string, args ...any) {
	uiInfo(w, format, args...)
}

func successf(w io.Writer, format string, args ...any) {
	uiSuccess(w, format, args...)
}
