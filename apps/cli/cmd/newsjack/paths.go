package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}

func newsjackHome() string {
	if v := os.Getenv("NEWSJACK_HOME"); v != "" {
		return expandPath(v)
	}
	return filepath.Join(homeDir(), ".newsjack")
}

func newsjackRoot() (string, error) {
	if v := os.Getenv("NEWSJACK_ROOT"); v != "" {
		root := expandPath(v)
		if dirExists(filepath.Join(root, "skills")) {
			return root, nil
		}
		return "", fmt.Errorf("NEWSJACK_ROOT does not look like the newsjack repo: %s", root)
	}
	root := filepath.Join(newsjackHome(), "newsjack")
	if dirExists(filepath.Join(root, "skills")) {
		return root, nil
	}
	if cwd, err := os.Getwd(); err == nil {
		if repo, ok := findRepoRoot(cwd); ok {
			return repo, nil
		}
	}
	return "", fmt.Errorf("newsjack is not installed at %s; run: curl -fsSL newsjack.sh | bash", root)
}

func findRepoRoot(start string) (string, bool) {
	dir := start
	for {
		if fileExists(filepath.Join(dir, "skills", "newsjack-detector", "SKILL.md")) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func expandPath(path string) string {
	if path == "~" {
		return homeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir(), path[2:])
	}
	if path != "" && !filepath.IsAbs(path) {
		if base := os.Getenv("NEWSJACK_WORKDIR"); base != "" {
			return filepath.Join(base, path)
		}
	}
	return path
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
