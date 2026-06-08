package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

const productTagline = "the operating system for agentic PR"

const bannerArt = `███╗   ██╗███████╗██╗    ██╗███████╗     ██╗ █████╗  ██████╗██╗  ██╗   ███████╗██╗  ██╗
████╗  ██║██╔════╝██║    ██║██╔════╝     ██║██╔══██╗██╔════╝██║ ██╔╝   ██╔════╝██║  ██║
██╔██╗ ██║█████╗  ██║ █╗ ██║███████╗     ██║███████║██║     █████╔╝    ███████╗███████║
██║╚██╗██║██╔══╝  ██║███╗██║╚════██║██   ██║██╔══██║██║     ██╔═██╗    ╚════██║██╔══██║
██║ ╚████║███████╗╚███╔███╔╝███████║╚█████╔╝██║  ██║╚██████╗██║  ██╗██╗███████║██║  ██║
╚═╝  ╚═══╝╚══════╝ ╚══╝╚══╝ ╚══════╝ ╚════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝╚══════╝╚═╝  ╚═╝`

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiCyan    = "\x1b[36m"
	ansiMagenta = "\x1b[38;5;209m"
)

func uiProduct(w io.Writer, section, note string) {
	uiBanner(w)
	fmt.Fprintln(w, uiPaint(w, ansiDim, productTagline))
	if section != "" {
		fmt.Fprintln(w)
		uiSection(w, section)
	}
	if note != "" {
		uiNote(w, "%s", note)
	}
}

func uiBanner(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, uiPaint(w, ansiMagenta+ansiBold, bannerArt))
}

func uiSection(w io.Writer, title string) {
	fmt.Fprintln(w, uiPaint(w, ansiMagenta+ansiBold, strings.ToUpper(title)))
}

func uiSectionExact(w io.Writer, title string) {
	fmt.Fprintln(w, uiPaint(w, ansiMagenta+ansiBold, title))
}

func uiKV(w io.Writer, key, value string) {
	fmt.Fprintf(w, "  %s %s\n", uiPaint(w, ansiDim, fmt.Sprintf("%-22s", key)), value)
}

func uiCommand(w io.Writer, command, description, flags string) {
	if flags != "" {
		flags = "  " + uiPaint(w, ansiDim, flags)
	}
	fmt.Fprintf(w, "  %s %s%s\n", uiPaint(w, ansiCyan, fmt.Sprintf("%-28s", command)), description, flags)
}

func uiInfo(w io.Writer, format string, args ...any) {
	uiTagged(w, ansiBlue, "[info]", format, args...)
}

func uiSuccess(w io.Writer, format string, args ...any) {
	uiTagged(w, ansiGreen, "[success]", format, args...)
}

func uiWarn(w io.Writer, format string, args ...any) {
	uiTagged(w, ansiYellow, "[warn]", format, args...)
}

func uiError(w io.Writer, format string, args ...any) {
	uiTagged(w, ansiRed, "[error]", format, args...)
}

func uiNote(w io.Writer, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(w, "%s %s\n", uiPaint(w, ansiDim, "-"), uiPaint(w, ansiDim, message))
}

func uiQuestion(w io.Writer) string {
	return uiPaint(w, ansiGreen+ansiBold, "?")
}

func uiTagged(w io.Writer, color, tag, format string, args ...any) {
	fmt.Fprintf(w, "%s "+format+"\n", append([]any{uiPaint(w, color, tag)}, args...)...)
}

func uiPaint(w io.Writer, color, value string) string {
	if !uiColorEnabled(w) || color == "" {
		return value
	}
	return color + value + ansiReset
}

func uiColorEnabled(w io.Writer) bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("NEWSJACK_COLOR")))
	if os.Getenv("NO_COLOR") != "" && mode != "always" {
		return false
	}
	switch mode {
	case "always", "1", "true", "yes", "on":
		return true
	case "never", "0", "false", "no", "off":
		return false
	}
	if strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))) == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
