package main

import (
	"path/filepath"
	"strings"
)

type commandInvocation struct {
	Command string
	Args    []string
}

func distributionName() string {
	if npmDistribution() {
		return "npm"
	}
	return "managed"
}

func npmDistribution() bool {
	root := strings.TrimSpace(getenv("NEWSJACK_ROOT", ""))
	return strings.EqualFold(strings.TrimSpace(getenv("NEWSJACK_DISTRIBUTION", "")), "npm") ||
		(root != "" && fileExists(filepath.Join(root, ".newsjack-npm")))
}

func newsjackCLIInvocation() commandInvocation {
	if npmDistribution() {
		return commandInvocation{Command: "newsjack"}
	}
	return commandInvocation{Command: filepath.Join(newsjackHome(), "bin", "newsjack")}
}

func (inv commandInvocation) With(args ...string) []string {
	out := append([]string{}, inv.Args...)
	return append(out, args...)
}

func (inv commandInvocation) CommandLine(args ...string) []string {
	out := []string{inv.Command}
	out = append(out, inv.With(args...)...)
	return out
}

func (inv commandInvocation) Display(args ...string) string {
	return strings.Join(inv.CommandLine(args...), " ")
}
