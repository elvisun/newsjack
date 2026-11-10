package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

// The Claude plugin ships two manifests whose versions must stay in lockstep:
// plugins/newsjack/.claude-plugin/plugin.json and the plugin entry in
// .claude-plugin/marketplace.json.
// The release workflow separately enforces that they match the release tag
// (see .github/workflows/release.yml, "Verify plugin version matches tag").
// This test guards the cheaper invariant — the two manifests never drift from
// each other — so a half-done version bump fails in CI before a tag is cut.
func TestPluginManifestVersionsMatch(t *testing.T) {
	root := repoRootForTest(t)

	pluginPath := filepath.Join(root, "plugins", "newsjack", ".claude-plugin", "plugin.json")
	plugin := readJSONForTest(t, pluginPath)
	pluginVer, _ := plugin["version"].(string)
	if pluginVer == "" {
		t.Fatal("plugins/newsjack/.claude-plugin/plugin.json has no version")
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

func TestHostedPluginMCPMatchesRootWithoutHeadersHelper(t *testing.T) {
	root := repoRootForTest(t)
	expected := readJSONForTest(t, filepath.Join(root, ".mcp.json"))
	packaged := readJSONForTest(t, filepath.Join(root, "plugins", "newsjack", ".mcp.json"))

	expectedServers, ok := expected["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal(".mcp.json has no mcpServers object")
	}
	for name, raw := range expectedServers {
		server, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf(".mcp.json server %q is not an object", name)
		}
		delete(server, "headersHelper")
	}

	packagedServers, ok := packaged["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("plugins/newsjack/.mcp.json has no mcpServers object")
	}
	for name, raw := range packagedServers {
		server, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("packaged .mcp.json server %q is not an object", name)
		}
		if _, found := server["headersHelper"]; found {
			t.Fatalf("packaged .mcp.json server %q contains headersHelper", name)
		}
	}

	if !reflect.DeepEqual(packaged, expected) {
		t.Fatal("plugins/newsjack/.mcp.json must match .mcp.json with every headersHelper removed")
	}
}
