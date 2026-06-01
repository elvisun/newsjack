package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDoctorReportsMissingXBearerToken(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":                     t.TempDir(),
		"NEWSJACK_HOME":            "",
		"NEWSJACK_ROOT":            repo,
		"NEWSJACK_IGNORE_DOTENV":   "1",
		"MEDIALYST_API_KEY":        "",
		"X_BEARER_TOKEN":           "",
		"TWITTER_BEARER_TOKEN":     "",
		"X_API_BEARER_TOKEN":       "",
		"TWITTER_API_BEARER_TOKEN": "",
		"PATH":                     t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"doctor", "--json"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("doctor code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid doctor JSON: %s", out.String())
		}
		auth := valueOrEmptyMap(payload["auth"])
		if auth["x_api_configured"] != false {
			t.Fatalf("x_api_configured=%v, want false", auth["x_api_configured"])
		}
		sources := valueOrEmptyMap(payload["sources"])
		if sources["x_news"] != false || sources["x_trends"] != false || sources["x"] != false {
			t.Fatalf("x sources should be unavailable without bearer token: %#v", sources)
		}
		if !strings.Contains(out.String(), "X_BEARER_TOKEN") {
			t.Fatalf("doctor should include actionable X warning:\n%s", out.String())
		}
	})
}

func TestDoctorReportsAvailableXLane(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":                     t.TempDir(),
		"NEWSJACK_HOME":            "",
		"NEWSJACK_ROOT":            repo,
		"NEWSJACK_IGNORE_DOTENV":   "1",
		"MEDIALYST_API_KEY":        "",
		"X_BEARER_TOKEN":           "x-token",
		"TWITTER_BEARER_TOKEN":     "",
		"X_API_BEARER_TOKEN":       "",
		"TWITTER_API_BEARER_TOKEN": "",
		"PATH":                     t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"doctor", "--json"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("doctor code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid doctor JSON: %s", out.String())
		}
		auth := valueOrEmptyMap(payload["auth"])
		if auth["x_api_configured"] != true {
			t.Fatalf("x_api_configured=%v, want true", auth["x_api_configured"])
		}
		sources := valueOrEmptyMap(payload["sources"])
		if sources["x_news"] != true || sources["x_trends"] != true || sources["x"] != true {
			t.Fatalf("x sources should be available with bearer token: %#v", sources)
		}
	})
}

func TestDoctorDefaultOutputIsHumanReadable(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":                     t.TempDir(),
		"NEWSJACK_HOME":            "",
		"NEWSJACK_ROOT":            repo,
		"NEWSJACK_IGNORE_DOTENV":   "1",
		"MEDIALYST_API_KEY":        "",
		"X_BEARER_TOKEN":           "",
		"TWITTER_BEARER_TOKEN":     "",
		"X_API_BEARER_TOKEN":       "",
		"TWITTER_API_BEARER_TOKEN": "",
		"PATH":                     t.TempDir(),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"doctor"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("doctor code=%d stderr=%s", code, errBuf.String())
		}
		text := out.String()
		if strings.HasPrefix(strings.TrimSpace(text), "{") {
			t.Fatalf("doctor default output should not be JSON:\n%s", text)
		}
		for _, want := range []string{"DOCTOR", "AUTH", "SOURCES", "RUNTIMES", "WARNINGS", "X_BEARER_TOKEN"} {
			if !strings.Contains(text, want) {
				t.Fatalf("doctor output missing %q:\n%s", want, text)
			}
		}
	})
}
