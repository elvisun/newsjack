package main

import (
	"os"
	"testing"
)

func repoRootForTest(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if root, ok := findRepoRoot(cwd); ok {
		return root
	}
	t.Fatal("repo root not found")
	return ""
}

func withTempEnv(t *testing.T, env map[string]string, fn func()) {
	t.Helper()
	old := map[string]string{}
	present := map[string]bool{}
	for key, value := range env {
		old[key], present[key] = os.LookupEnv(key)
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
	defer func() {
		for key := range env {
			if present[key] {
				os.Setenv(key, old[key])
			} else {
				os.Unsetenv(key)
			}
		}
	}()
	fn()
}
