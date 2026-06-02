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
	uiCommand(w, "version", "print the installed version", "")
	uiCommand(w, "path", "print the install root", "")
	uiCommand(w, "doctor", "run a system health check", "[--json]")
	uiCommand(w, "setup", "guided install: runtimes, auth, skills", "[--yes]")
	uiCommand(w, "install", "install the skill bundle", "[--source DIR]")
	uiCommand(w, "skills [list]", "manage installed skills", "")
	uiCommand(w, "runtimes detect", "detect supported agent runtimes", "")
	uiCommand(w, "login", "save an optional Medialyst API key", "[--key KEY]")
	uiCommand(w, "auth status|set|logout", "inspect, save, or revoke API credentials", "")
	uiCommand(w, "monitor init|test|run...", "manage newsjacking monitors", "")
	uiCommand(w, "detector run|recent...", "angle detection over recent stories", "")
	uiCommand(w, "mcp setup|status", "configure the MCP bridge", "")
	uiCommand(w, "update", "pull the latest skill bundle", "")
	fmt.Fprintln(w)
	uiSection(w, "api setup")
	uiCommand(w, "auth set-medialyst", "save Medialyst for live news search, media database, find journalists", "--key KEY")
	uiCommand(w, "auth set-x", "save X bearer token for X News, trends, and post search", "--bearer-token TOKEN")
	uiKV(w, "Medialyst key", medialystAPIKeyURL)
	uiKV(w, "X bearer token", xAPIKeyURL)
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
	case "auth":
		printAuthHelp(w)
		return true
	case "doctor":
		uiProduct(w, "doctor", "checks install health and prints concrete recovery commands.")
		fmt.Fprintln(w)
		uiSection(w, "usage")
		fmt.Fprintln(w, "  newsjack doctor [--json]")
		fmt.Fprintln(w)
		uiSection(w, "api recovery")
		uiCommand(w, "auth set-medialyst", "save Medialyst API key", "--key <mlst_...>")
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
		uiCommand(w, "auth set-medialyst", "save Medialyst without rerunning setup", "--key <mlst_...>")
		uiCommand(w, "auth set-x", "save X without rerunning setup", "--bearer-token <token>")
		return true
	default:
		return false
	}
}

func printAuthHelp(w io.Writer) {
	uiProduct(w, "auth", "inspect and configure optional API credentials.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack auth status")
	fmt.Fprintln(w, "  newsjack auth set --medialyst-key <mlst_...> --x-bearer-token <token>")
	fmt.Fprintln(w, "  newsjack auth set-medialyst --key <mlst_...>")
	fmt.Fprintln(w, "  newsjack auth set-x --bearer-token <token>")
	fmt.Fprintln(w)
	uiSection(w, "optional apis")
	uiKV(w, "Medialyst", "live news search, media database, find journalists")
	uiKV(w, "get key", medialystAPIKeyURL)
	uiKV(w, "save key", "newsjack auth set-medialyst --key <mlst_...>")
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
