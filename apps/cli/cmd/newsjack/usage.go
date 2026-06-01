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
	uiCommand(w, "auth status|headers|logout", "inspect or revoke credentials", "")
	uiCommand(w, "monitor init|test|run...", "manage newsjacking monitors", "")
	uiCommand(w, "detector run|recent...", "angle detection over recent stories", "")
	uiCommand(w, "run-summary", "write deterministic run metadata as JSON", "INPUT [--output FILE]")
	uiCommand(w, "mcp setup|status", "configure the MCP bridge", "")
	uiCommand(w, "update", "pull the latest skill bundle", "")
	fmt.Fprintln(w)
	uiSection(w, "learn")
	uiCommand(w, "newsjack help <command>", "detail on a single command", "")
	uiCommand(w, "newsjack doctor", "fastest way to know it works", "")
	fmt.Fprintln(w)
	uiNote(w, "start with: newsjack setup")
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
