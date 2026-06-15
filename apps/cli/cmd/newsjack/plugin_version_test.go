package main

import (
	"path/filepath"
	"testing"
)

// The Claude plugin ships two manifests whose versions must stay in lockstep:
// .claude-plugin/plugin.json and the plugin entry in .claude-plugin/marketplace.json.
// The release workflow separately enforces that they match the release tag
// (see .github/workflows/release.yml, "Verify plugin version matches tag").
// This test guards the cheaper invariant — the two manifests never drift from
// each other — so a half-done version bump fails in CI before a tag is cut.
func TestPluginManifestVersionsMatch(t *testing.T) {
	root := repoRootForTest(t)

	plugin := readJSONForTest(t, filepath.Join(root, ".claude-plugin", "plugin.json"))
	pluginVer, _ := plugin["version"].(string)
	if pluginVer == "" {
		t.Fatal(".claude-plugin/plugin.json has no version")
	}

	marketplace := readJSONForTest(t, filepath.Join(root, ".claude-plugin", "marketplace.json"))
	plugins, ok := marketplace["plugins"].([]any)
	if !ok || len(plugins) == 0 {
		t.Fatal(".claude-plugin/marketplace.json has no plugins")
	}

	found := false
	for _, raw := range plugins {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if name, _ := entry["name"].(string); name != "newsjack" {
			continue
		}
		found = true
		mktVer, _ := entry["version"].(string)
		if mktVer != pluginVer {
			t.Fatalf("plugin version drift: plugin.json=%q marketplace.json newsjack=%q — bump both together", pluginVer, mktVer)
		}
	}
	if !found {
		t.Fatal(".claude-plugin/marketplace.json has no plugin named newsjack")
	}
}
