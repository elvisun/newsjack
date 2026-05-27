package main

import (
	"io"
	"os/exec"
)

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func cmdDoctor(_ []string, stdout, _ io.Writer) int {
	root, rootErr := newsjackRoot()
	key, source := loadAPIKey()
	payload := map[string]any{
		"version":       version,
		"newsjack_home": newsjackHome(),
		"newsjack_root": root,
		"root_ok":       rootErr == nil,
		"auth": map[string]any{
			"medialyst_configured": key != "",
			"source":               nullableString(source),
		},
		"dependencies": map[string]any{
			"npx": commandAvailable("npx"),
		},
	}
	payload["runtimes"] = runtimeStatus()
	writeJSON(stdout, payload)
	return 0
}

func runtimeStatus() map[string]any {
	out := map[string]any{}
	for _, rt := range runtimeTargets {
		out[rt.Key] = map[string]any{
			"detected":   runtimeDetected(rt),
			"skills_dir": targetDir(rt),
		}
	}
	return out
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
