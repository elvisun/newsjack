package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	os.Exit(runCLIWithIO(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	return runCLIWithIO(args, strings.NewReader(""), stdout, stderr)
}

func runCLIWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cmd := "help"
	if len(args) > 0 {
		cmd = args[0]
	}
	if code, handled := maybeAutoUpdate(args, stderr); handled {
		return code
	}
	switch cmd {
	case "help", "--help", "-h":
		if len(args) > 1 {
			if printCommandHelp(stdout, args[1]) {
				return 0
			}
			return failf(stderr, "unknown help topic: %s", args[1])
		}
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
			case "status":
				return cmdSkillsStatus(args[2:], stdout, stderr)
			case "--help", "-h":
				fmt.Fprintln(stdout, "Usage: newsjack skills list|install|status")
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
		return cmdSetup(args[1:], stdin, stdout, stderr)
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
	case "coverage":
		return cmdCoverage(args[1:], stdout, stderr)
	case "filter-apply":
		return cmdFilterApply(args[1:], stdout, stderr)
	case "cluster":
		return cmdCluster(args[1:], stdout, stderr)
	case "origin-apply":
		return cmdOriginApply(args[1:], stdout, stderr)
	case "run-summary":
		return cmdRunSummary(args[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return failf(stderr, "unknown command: %s", cmd)
	}
}
