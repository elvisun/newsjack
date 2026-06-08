package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	skillsModeManaged  = "managed"
	skillsModeExternal = "external"
)

type installState struct {
	Version     string   `json:"version"`
	Commit      string   `json:"commit"`
	Channel     string   `json:"channel"`
	Repo        string   `json:"repo"`
	InstallURL  string   `json:"install_url"`
	SkillsMode  string   `json:"skills_mode"`
	Runtimes    []string `json:"runtimes"`
	RuntimesRaw string   `json:"runtimes_raw"`
	InstallMCP  bool     `json:"install_mcp"`
	InstalledAt string   `json:"installed_at"`
}

func installStatePath() string {
	return filepath.Join(newsjackHome(), "install.json")
}

func readInstallState() (installState, error) {
	var state installState
	data, err := os.ReadFile(installStatePath())
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	return normalizeInstallState(state), nil
}

func readInstallStateOrDefault() installState {
	state, err := readInstallState()
	if err != nil {
		state = installState{}
	}
	return normalizeInstallState(state)
}

func normalizeInstallState(state installState) installState {
	emptyState := state.Version == "" && state.Commit == "" && state.InstalledAt == ""
	if state.Repo == "" {
		state.Repo = defaultRepo
	}
	if state.InstallURL == "" {
		state.InstallURL = defaultInstallURL
	}
	if state.Channel == "" {
		state.Channel = "stable"
	}
	switch state.SkillsMode {
	case skillsModeManaged, skillsModeExternal:
	default:
		state.SkillsMode = skillsModeManaged
	}
	if state.RuntimesRaw == "" && len(state.Runtimes) > 0 {
		state.RuntimesRaw = strings.Join(state.Runtimes, ",")
	}
	if emptyState {
		state.InstallMCP = true
	}
	return state
}

func installedVersion() string {
	if state, err := readInstallState(); err == nil && strings.TrimSpace(state.Version) != "" {
		return strings.TrimSpace(state.Version)
	}
	if installed := strings.TrimSpace(readInstalledVersion()); installed != "" {
		return installed
	}
	if root, err := newsjackRoot(); err == nil {
		if bundled := strings.TrimSpace(readSkillsManifestVersion(root)); bundled != "" {
			return bundled
		}
	}
	return version
}
