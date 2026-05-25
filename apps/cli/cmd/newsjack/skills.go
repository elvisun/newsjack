package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func cmdUpdate(_ []string, _ io.Writer, stderr io.Writer) int {
	shell := "sh"
	var cmd *exec.Cmd
	if curl, err := exec.LookPath("curl"); err == nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q -fsSL https://newsjack.sh/install.sh | sh", curl))
	} else if wget, err := exec.LookPath("wget"); err == nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q -qO- https://newsjack.sh/install.sh | sh", wget))
	} else {
		return fail(stderr, errors.New("curl or wget is required"))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fail(stderr, err)
	}
	return 0
}
