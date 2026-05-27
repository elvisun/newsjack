package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLatestChannelVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dist/channels/main.txt" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("abc123\n"))
	}))
	defer server.Close()

	withTempEnv(t, map[string]string{
		"NEWSJACK_DIST_BASE": server.URL + "/dist",
		"NEWSJACK_CHANNEL":   "main",
	}, func() {
		got, err := latestChannelVersion()
		if err != nil {
			t.Fatal(err)
		}
		if got != "abc123" {
			t.Fatalf("latestChannelVersion() = %q, want abc123", got)
		}
	})
}

func TestReadInstalledVersion(t *testing.T) {
	home := t.TempDir()
	versionPath := filepath.Join(home, ".newsjack", "newsjack", "VERSION")
	if err := os.MkdirAll(filepath.Dir(versionPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(versionPath, []byte("abc123\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withTempEnv(t, map[string]string{"NEWSJACK_HOME": filepath.Join(home, ".newsjack")}, func() {
		if got := readInstalledVersion(); got != "abc123" {
			t.Fatalf("readInstalledVersion() = %q, want abc123", got)
		}
	})
}

func TestAutoUpdateDisabled(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_AUTO_UPDATE": "0"}, func() {
		if !autoUpdateDisabled() {
			t.Fatal("NEWSJACK_AUTO_UPDATE=0 should disable auto-update")
		}
	})
	withTempEnv(t, map[string]string{"NEWSJACK_NO_AUTO_UPDATE": "1"}, func() {
		if !autoUpdateDisabled() {
			t.Fatal("NEWSJACK_NO_AUTO_UPDATE=1 should disable auto-update")
		}
	})
}
