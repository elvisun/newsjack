package main

import (
	"fmt"
	"io"
	"strings"
)

const productTagline = "the operating system for agentic PR"

func uiProduct(w io.Writer, section, note string) {
	fmt.Fprintln(w, "newsjack")
	fmt.Fprintln(w, productTagline)
	if section != "" {
		fmt.Fprintln(w)
		uiSection(w, section)
	}
	if note != "" {
		uiNote(w, note)
	}
}

func uiSection(w io.Writer, title string) {
	fmt.Fprintln(w, strings.ToUpper(title))
}

func uiKV(w io.Writer, key, value string) {
	fmt.Fprintf(w, "  %-22s %s\n", key, value)
}

func uiCommand(w io.Writer, command, description, flags string) {
	if flags != "" {
		flags = "  " + flags
	}
	fmt.Fprintf(w, "  %-28s %s%s\n", command, description, flags)
}

func uiInfo(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[info] "+format+"\n", args...)
}

func uiSuccess(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[success] "+format+"\n", args...)
}

func uiWarn(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[warn] "+format+"\n", args...)
}

func uiError(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[error] "+format+"\n", args...)
}

func uiNote(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "- "+format+"\n", args...)
}
