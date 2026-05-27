package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func cmdSkillsList(_ []string, stdout, stderr io.Writer) int {
	root, err := newsjackRoot()
	if err != nil {
		return fail(stderr, err)
	}
	names, err := skillNames(root)
	if err != nil {
		return fail(stderr, err)
	}
	for _, name := range names {
		fmt.Fprintln(stdout, name)
	}
	return 0
}

func skillNames(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if fileExists(filepath.Join(root, "skills", entry.Name(), "SKILL.md")) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
