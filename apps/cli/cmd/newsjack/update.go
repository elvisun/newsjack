package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const autoUpdateGuard = "NEWSJACK_AUTO_UPDATE_RUNNING"

func cmdUpdate(_ []string, stdout, stderr io.Writer) int {
	if err := runHostedInstaller(stdout, stderr); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func maybeAutoUpdate(args []string, stderr io.Writer) (int, bool) {
	if !shouldAutoUpdate(args) {
		return 0, false
	}
	latest, err := latestChannelVersion()
	if err != nil || latest == "" {
		return 0, false
	}
	current := strings.TrimSpace(readInstalledVersion())
	if current == latest {
		return 0, false
	}

	if current == "" {
		logf(stderr, "auto-updating to %s", latest)
	} else {
		logf(stderr, "auto-updating from %s to %s", current, latest)
	}

	if err := runHostedInstaller(os.Stderr, os.Stderr); err != nil {
		warn(stderr, "auto-update failed: %v", err)
		return 0, false
	}

	code := runInstalledBinary(args)
	return code, true
}

func shouldAutoUpdate(args []string) bool {
	if autoUpdateDisabled() {
		return false
	}
	if os.Getenv(autoUpdateGuard) == "1" {
		return false
	}
	if !runningInstalledBinary() {
		return false
	}
	cmd := "help"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "help", "--help", "-h", "version", "--version", "install", "update", "auth", "mcp-bridge":
		return false
	}
	return true
}

func autoUpdateDisabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NEWSJACK_AUTO_UPDATE"))) {
	case "0", "false", "off", "no":
		return true
	}
	return os.Getenv("NEWSJACK_NO_AUTO_UPDATE") == "1"
}

func runningInstalledBinary() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	want := filepath.Join(newsjackHome(), "bin", "newsjack")
	exeReal, err := filepath.EvalSymlinks(exe)
	if err != nil {
		exeReal = exe
	}
	wantReal, err := filepath.EvalSymlinks(want)
	if err != nil {
		wantReal = want
	}
	return exeReal == wantReal
}

func readInstalledVersion() string {
	data, err := os.ReadFile(filepath.Join(newsjackHome(), "newsjack", "VERSION"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func latestChannelVersion() (string, error) {
	base := strings.TrimRight(getenv("NEWSJACK_DIST_BASE", defaultDistBase), "/")
	channel := strings.TrimSpace(getenv("NEWSJACK_CHANNEL", "main"))
	if channel == "" {
		channel = "main"
	}
	raw, err := httpGetRaw(base+"/channels/"+channel+".txt", nil, 2*time.Second)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

func runHostedInstaller(stdout, stderr io.Writer) error {
	shell := "sh"
	url := getenv("NEWSJACK_INSTALL_URL", defaultInstallURL)
	var cmd *exec.Cmd
	if curl, err := exec.LookPath("curl"); err == nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q -fsSL %q | sh", curl, url))
	} else if wget, err := exec.LookPath("wget"); err == nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q -qO- %q | sh", wget, url))
	} else {
		return errors.New("curl or wget is required")
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), autoUpdateGuard+"=1")
	return cmd.Run()
}

func runInstalledBinary(args []string) int {
	bin := filepath.Join(newsjackHome(), "bin", "newsjack")
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), autoUpdateGuard+"=1")
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		uiError(os.Stderr, "could not run updated binary: %v", err)
		return 1
	}
	return 0
}
