package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func cmdMCPBridge(_ []string, _ io.Writer, stderr io.Writer) int {
	key, source := loadAPIKey()
	if key == "" {
		fmt.Fprintln(stderr, "Medialyst API key not found. Run: ~/.newsjack/bin/newsjack login")
		return 1
	}
	npx, err := exec.LookPath("npx")
	if err != nil {
		fmt.Fprintln(stderr, "npx is required for the Medialyst MCP stdio bridge. Install Node.js, then retry.")
		return 1
	}
	if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
		fmt.Fprintf(stderr, "Loaded Medialyst API key from %s\n", source)
	}
	env := os.Environ()
	env = setEnv(env, envMedialystKey, key)
	args := []string{npx, "-y", "mcp-remote", medialystMCPURL, "--header", "Authorization: Bearer ${" + envMedialystKey + "}"}
	if err := syscall.Exec(npx, args, env); err != nil {
		return fail(stderr, err)
	}
	return 127
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
