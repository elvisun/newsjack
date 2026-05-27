package main

import (
	"fmt"
	"io"
)

func printUsage(w io.Writer) {
	fmt.Fprint(w, `newsjack

Usage:
  newsjack help
  newsjack version
  newsjack path
  newsjack doctor
  newsjack setup [--json]
  newsjack install [--source DIR]
  newsjack skills [list]
  newsjack skills install [--source DIR]
  newsjack runtimes detect
  newsjack login [--key KEY]
  newsjack auth status|headers|logout
  newsjack monitor init|test|run|schedule|status|open ...
  newsjack detector run|diagnose|recent ...
  newsjack filter-apply ...
  newsjack origin-apply ...
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
