package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func withDoctorMedialystCheck(t *testing.T, check func() medialystCredentialCheck, fn func()) {
	t.Helper()
	old := doctorMedialystCredentialCheck
	doctorMedialystCredentialCheck = check
	defer func() { doctorMedialystCredentialCheck = old }()
	fn()
}

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
		actions := anySlice(payload["actions"])
		if len(actions) != 2 {
			t.Fatalf("doctor JSON should include two API actions: %s", out.String())
		}
		medialystAction := valueOrEmptyMap(actions[0])
		xAction := valueOrEmptyMap(actions[1])
		if medialystAction["label"] != "Connect Medialyst (Optional)" ||
			medialystAction["command"] != "newsjack login" ||
			medialystAction["fallback"] != "newsjack auth set-medialyst --key <mlst_...>" ||
			medialystAction["get_key_url"] != "https://medialyst.ai/app/onboarding/developer" ||
			medialystAction["used_for"] != "live news search, journalist enrichment, and media list research" {
			t.Fatalf("unexpected Medialyst doctor action: %#v", medialystAction)
		}
		if xAction["label"] != "Configure X API (Optional)" ||
			xAction["command"] != "newsjack auth set-x --bearer-token <token>" ||
			xAction["writes"] != "~/.newsjack/.env:X_BEARER_TOKEN" {
			t.Fatalf("unexpected X doctor action: %#v", xAction)
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
		for _, want := range []string{
			"DOCTOR",
			"AUTH",
			"SOURCES",
			"RUNTIMES",
			"WARNINGS",
			"NEXT ACTIONS",
			"X_BEARER_TOKEN",
			"newsjack login",
			"newsjack auth set-medialyst --key <mlst_...>",
			"newsjack auth set-x --bearer-token <token>",
			"https://medialyst.ai/app/onboarding/developer",
		} {
			if !strings.Contains(text, want) {
				t.Fatalf("doctor output missing %q:\n%s", want, text)
			}
		}
	})
}

func TestDoctorReportsMedialystCredentialStates(t *testing.T) {
	repo := repoRootForTest(t)
	for _, tc := range []struct {
		state       string
		detail      string
		wantWarning string
	}{
		{state: medialystCredentialOK, wantWarning: ""},
		{state: medialystCredentialInvalid, detail: "authentication was rejected", wantWarning: "credentials are invalid"},
		{state: medialystCredentialUnreachable, detail: "network unavailable", wantWarning: "could not be reached"},
	} {
		t.Run(tc.state, func(t *testing.T) {
			withTempEnv(t, map[string]string{
				"HOME":                   t.TempDir(),
				"NEWSJACK_HOME":          "",
				"NEWSJACK_ROOT":          repo,
				"NEWSJACK_IGNORE_DOTENV": "1",
				"MEDIALYST_API_KEY":      "mlst_test_key_12345",
				"X_BEARER_TOKEN":         "x-token",
				"PATH":                   t.TempDir(),
			}, func() {
				withDoctorMedialystCheck(t, func() medialystCredentialCheck {
					return medialystCredentialCheck{State: tc.state, Detail: tc.detail}
				}, func() {
					var out, errBuf bytes.Buffer
					if code := runCLI([]string{"doctor", "--json"}, &out, &errBuf); code != 0 {
						t.Fatalf("doctor code=%d stderr=%s", code, errBuf.String())
					}
					var payload map[string]any
					if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
						t.Fatal(err)
					}
					auth := valueOrEmptyMap(payload["auth"])
					if auth["medialyst_status"] != tc.state {
						t.Fatalf("medialyst_status=%v, want %s", auth["medialyst_status"], tc.state)
					}
					if tc.detail != "" && auth["medialyst_status_detail"] != tc.detail {
						t.Fatalf("medialyst_status_detail=%v, want %s", auth["medialyst_status_detail"], tc.detail)
					}
					if tc.wantWarning != "" && !strings.Contains(strings.ToLower(out.String()), tc.wantWarning) {
						t.Fatalf("doctor JSON missing warning %q: %s", tc.wantWarning, out.String())
					}

					out.Reset()
					if code := runCLI([]string{"doctor"}, &out, &errBuf); code != 0 {
						t.Fatalf("doctor code=%d stderr=%s", code, errBuf.String())
					}
					if !strings.Contains(out.String(), "Medialyst") || !strings.Contains(out.String(), tc.state) {
						t.Fatalf("human doctor output missing state %q:\n%s", tc.state, out.String())
					}
				})
			})
		})
	}
}

func TestCheckMedialystCredentialUsesBalanceEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name       string
		statusCode int
		want       string
	}{
		{name: "ok", statusCode: http.StatusOK, want: medialystCredentialOK},
		{name: "invalid", statusCode: http.StatusUnauthorized, want: medialystCredentialInvalid},
		{name: "forbidden_is_alive", statusCode: http.StatusForbidden, want: medialystCredentialOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath, gotAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath, gotAuth = r.Method, r.URL.Path, r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(`{"credits":{"org_available_balance":1}}`))
			}))
			defer server.Close()
			withTempEnv(t, map[string]string{
				"HOME":                        t.TempDir(),
				"NEWSJACK_HOME":               "",
				"NEWSJACK_IGNORE_DOTENV":      "1",
				"MEDIALYST_API_KEY":           "mlst_test_key_12345",
				"NEWSJACK_MEDIALYST_API_BASE": server.URL,
			}, func() {
				check := checkMedialystCredential()
				if check.State != tc.want {
					t.Fatalf("state=%s detail=%q, want %s", check.State, check.Detail, tc.want)
				}
			})
			if gotMethod != http.MethodGet || gotPath != "/v1/credits/balance" || gotAuth != "Bearer mlst_test_key_12345" {
				t.Fatalf("request=%s %s auth=%q", gotMethod, gotPath, gotAuth)
			}
		})
	}
}

func TestCheckMedialystCredentialReportsInvalidGrantAndUnreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/oauth/token" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh token expired"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	serverURL := server.URL
	server.Close()

	withTempEnv(t, map[string]string{
		"HOME":                        t.TempDir(),
		"NEWSJACK_HOME":               "",
		"NEWSJACK_IGNORE_DOTENV":      "1",
		"MEDIALYST_API_KEY":           "mlst_test_key_12345",
		"NEWSJACK_MEDIALYST_API_BASE": serverURL,
	}, func() {
		if check := checkMedialystCredential(); check.State != medialystCredentialUnreachable {
			t.Fatalf("closed server state=%s detail=%q", check.State, check.Detail)
		}
	})

	invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh token expired"}`))
	}))
	defer invalidServer.Close()
	withTempEnv(t, map[string]string{
		"HOME":                        t.TempDir(),
		"NEWSJACK_HOME":               "",
		"NEWSJACK_IGNORE_DOTENV":      "1",
		"MEDIALYST_API_KEY":           "",
		"NEWSJACK_MEDIALYST_API_BASE": invalidServer.URL,
	}, func() {
		if _, err := writeCredentialsFile(credentialsFile{Medialyst: medialystCredentials{OAuth: &medialystOAuthCredentials{
			AccessToken:  "mcp_at_old",
			RefreshToken: "mcp_rt_old",
			ExpiresAt:    time.Now().Add(-time.Hour).Format(time.RFC3339),
			BaseURL:      invalidServer.URL,
		}}}); err != nil {
			t.Fatal(err)
		}
		if check := checkMedialystCredential(); check.State != medialystCredentialInvalid {
			t.Fatalf("invalid_grant state=%s detail=%q", check.State, check.Detail)
		}
	})
}
