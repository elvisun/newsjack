package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestAuthStatusMissingAndLoginHeaders(t *testing.T) {
	withTempEnv(t, map[string]string{
		"HOME":              t.TempDir(),
		"NEWSJACK_HOME":     "",
		"MEDIALYST_API_KEY": "",
	}, func() {
		cwd, chdirErr := os.Getwd()
		if chdirErr != nil {
			t.Fatal(chdirErr)
		}
		t.Cleanup(func() {
			if err := os.Chdir(cwd); err != nil {
				t.Fatal(err)
			}
		})
		if err := os.Chdir(t.TempDir()); err != nil {
			t.Fatal(err)
		}

		var out, err bytes.Buffer
		code := runCLI([]string{"auth", "status"}, &out, &err)
		if code != 1 {
			t.Fatalf("status code=%d, want 1; stderr=%s", code, err.String())
		}
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil || status["configured"] != false {
			t.Fatalf("unexpected status JSON: %s", out.String())
		}

		out.Reset()
		err.Reset()
		testKey := "mlst_" + strings.Repeat("x", 12)
		code = runCLI([]string{"login", "--key", testKey}, &out, &err)
		if code != 0 {
			t.Fatalf("login code=%d stderr=%s", code, err.String())
		}
		info, statErr := os.Stat(credentialsPath())
		if statErr != nil {
			t.Fatalf("credentials not written: %v", statErr)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("credentials mode=%o, want 600", got)
		}

		out.Reset()
		err.Reset()
		code = runCLI([]string{"auth", "headers"}, &out, &err)
		if code != 0 {
			t.Fatalf("headers code=%d stderr=%s", code, err.String())
		}
		if out.String() != "{\"Authorization\":\"Bearer "+testKey+"\"}\n" {
			t.Fatalf("unexpected headers: %q", out.String())
		}
	})
}

func TestDetectorDiagnoseUsesSavedMedialystLogin(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":              t.TempDir(),
		"NEWSJACK_HOME":     "",
		"NEWSJACK_ROOT":     repo,
		"MEDIALYST_API_KEY": "",
	}, func() {
		cwd, chdirErr := os.Getwd()
		if chdirErr != nil {
			t.Fatal(chdirErr)
		}
		t.Cleanup(func() {
			if err := os.Chdir(cwd); err != nil {
				t.Fatal(err)
			}
		})
		if err := os.Chdir(t.TempDir()); err != nil {
			t.Fatal(err)
		}

		testKey := "mlst_" + strings.Repeat("x", 12)
		var out, err bytes.Buffer
		if code := runCLI([]string{"login", "--key", testKey}, &out, &err); code != 0 {
			t.Fatalf("login code=%d stderr=%s", code, err.String())
		}

		out.Reset()
		err.Reset()
		code := runCLI([]string{"detector", "diagnose", "--sources", "news_search"}, &out, &err)
		if code != 0 {
			t.Fatalf("diagnose code=%d stderr=%s", code, err.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid diagnose JSON: %s", out.String())
		}
		if payload["news_search_configured"] != true {
			t.Fatalf("news_search_configured=%v, want true; payload=%s", payload["news_search_configured"], out.String())
		}
		if !containsAnyString(anySlice(payload["sources_available"]), "news_search") {
			t.Fatalf("sources_available=%v, want news_search", payload["sources_available"])
		}
	})
}

func containsAnyString(items []any, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
