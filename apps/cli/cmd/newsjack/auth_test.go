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
		if _, statErr := os.Stat(credentialsPath()); statErr != nil {
			t.Fatalf("credentials not written: %v", statErr)
		}
		assertOwnerOnlyFile(t, credentialsPath())

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

func TestAuthSetStoresOptionalAPIs(t *testing.T) {
	withTempEnv(t, map[string]string{
		"HOME":                     t.TempDir(),
		"NEWSJACK_HOME":            "",
		"MEDIALYST_API_KEY":        "",
		"X_BEARER_TOKEN":           "",
		"TWITTER_BEARER_TOKEN":     "",
		"X_API_BEARER_TOKEN":       "",
		"TWITTER_API_BEARER_TOKEN": "",
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

		var out, errBuf bytes.Buffer
		testKey := "mlst_" + strings.Repeat("a", 12)
		code := runCLI([]string{"auth", "set", "--medialyst-key", testKey, "--x-bearer-token", "x-token"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("auth set code=%d stderr=%s stdout=%s", code, errBuf.String(), out.String())
		}
		if !strings.Contains(out.String(), "live news search, media database, find journalists") {
			t.Fatalf("auth set should explain Medialyst usage:\n%s", out.String())
		}

		key, keySource := loadAPIKey()
		if key != testKey || !strings.HasPrefix(keySource, "credentials:") {
			t.Fatalf("medialyst key/source=%q/%q", key, keySource)
		}
		xToken, xSource := loadXBearerToken()
		if xToken != "x-token" || !strings.Contains(xSource, newsjackEnvPath()+":X_BEARER_TOKEN") {
			t.Fatalf("x token/source=%q/%q", xToken, xSource)
		}
		envBody, err := os.ReadFile(newsjackEnvPath())
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(envBody), `X_BEARER_TOKEN="x-token"`) {
			t.Fatalf("X token was not saved to newsjack env:\n%s", envBody)
		}
		assertOwnerOnlyFile(t, newsjackEnvPath())

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"auth", "status"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("auth status code=%d stderr=%s", code, errBuf.String())
		}
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil {
			t.Fatalf("invalid auth status JSON: %s", out.String())
		}
		if status["medialyst_configured"] != true || status["x_api_configured"] != true {
			t.Fatalf("auth status should report both APIs configured: %s", out.String())
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
