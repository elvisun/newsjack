package main

import (
	"encoding/json"
	"flag"
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

func cmdSkillsStatus(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("skills status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "Emit skill status as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	root, rootErr := newsjackRoot()
	state := readInstallStateOrDefault()
	payload := map[string]any{
		"newsjack_root": root,
		"root_ok":       rootErr == nil,
		"version":       installedVersion(),
		"skills_mode":   state.SkillsMode,
		"runtimes":      state.Runtimes,
		"runtimes_raw":  state.RuntimesRaw,
	}
	if rootErr == nil {
		names, err := skillNames(root)
		if err != nil {
			return fail(stderr, err)
		}
		payload["bundle_skills"] = names
		payload["bundle_skill_count"] = len(names)
		payload["skills_manifest_present"] = fileExists(filepath.Join(root, "skills-manifest.json"))
		payload["skills_manifest_version"] = readSkillsManifestVersion(root)
		payload["managed_targets"] = managedSkillTargets(state, names)
	} else {
		payload["bundle_skills"] = []string{}
		payload["bundle_skill_count"] = 0
		payload["skills_manifest_present"] = false
		payload["skills_manifest_version"] = ""
		payload["managed_targets"] = []map[string]any{}
	}

	if *jsonOut {
		writeJSON(stdout, payload)
		return 0
	}

	uiProduct(stdout, "skills", "installed skill status")
	uiKV(stdout, "mode", state.SkillsMode)
	uiKV(stdout, "version", installedVersion())
	if rootErr != nil {
		uiKV(stdout, "bundle", "missing: "+rootErr.Error())
		return 0
	}
	uiKV(stdout, "bundle", root)
	uiKV(stdout, "bundle skills", fmt.Sprintf("%v", payload["bundle_skill_count"]))
	if state.SkillsMode == skillsModeExternal {
		uiNote(stdout, "runtime skills are marketplace-owned; Newsjack updates the CLI and bundle only.")
	}
	return 0
}

func managedSkillTargets(state installState, names []string) []map[string]any {
	runtimes := state.RuntimesRaw
	if runtimes == "" {
		runtimes = "auto"
	}
	targets := selectedRuntimes(runtimes)
	out := []map[string]any{}
	for _, rt := range targets {
		dir := targetDir(rt)
		installed := 0
		missing := 0
		for _, name := range names {
			if fileExists(filepath.Join(dir, name, "SKILL.md")) {
				installed++
			} else {
				missing++
			}
		}
		out = append(out, map[string]any{
			"runtime":    rt.Key,
			"label":      rt.Label,
			"skills_dir": dir,
			"installed":  installed,
			"missing":    missing,
		})
	}
	return out
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

func readSkillsManifestVersion(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "skills-manifest.json"))
	if err != nil {
		return ""
	}
	var payload struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	return payload.Version
}
