package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	cmd := "help"
	if len(args) > 0 {
		cmd = args[0]
	}
	if code, handled := maybeAutoUpdate(args, stderr); handled {
		return code
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
	case "setup":
		return cmdSetup(args[1:], stdout, stderr)
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
	case "monitor":
		return cmdMonitor(args[1:], stdout, stderr)
	case "filter-apply":
		return cmdFilterApply(args[1:], stdout, stderr)
	case "origin-apply":
		return cmdOriginApply(args[1:], stdout, stderr)
	case "summarize-run":
		return cmdSummarizeRun(args[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return failf(stderr, "unknown command: %s", cmd)
	}
}
