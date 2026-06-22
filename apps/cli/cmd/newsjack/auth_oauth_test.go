package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func withOAuthHooks(t *testing.T, open func(string) error, sleep func(time.Duration), now func() time.Time) {
	t.Helper()
	oldOpen := openBrowserURL
	oldSleep := oauthSleep
	oldNow := oauthNow
	if open != nil {
		openBrowserURL = open
	}
	if sleep != nil {
		oauthSleep = sleep
	}
	if now != nil {
		oauthNow = now
	}
	t.Cleanup(func() {
		openBrowserURL = oldOpen
		oauthSleep = oldSleep
		oauthNow = oldNow
	})
}

func TestLoginDeviceFlowSendsClientAndStoresOAuth(t *testing.T) {
	home := t.TempDir()
	var deviceForm, tokenForm map[string]string
	var opened []string
	var sleeps []time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/oauth/device_authorization":
			if r.Method != http.MethodPost {
				t.Fatalf("device method=%s", r.Method)
			}
			if !strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
				t.Fatalf("device content-type=%q", r.Header.Get("Content-Type"))
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			deviceForm = map[string]string{
				"client_id": r.Form.Get("client_id"),
				"scope":     r.Form.Get("scope"),
			}
			w.Write([]byte(`{
				"device_code":"mcp_dc_test",
				"user_code":"ABCD-EFGH",
				"verification_uri":"` + serverDeviceURL(r, "/device") + `",
				"verification_uri_complete":"` + serverDeviceURL(r, "/device?user_code=ABCD-EFGH") + `",
				"expires_in":600,
				"interval":2
			}`))
		case "/api/oauth/token":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			tokenForm = map[string]string{
				"grant_type":  r.Form.Get("grant_type"),
				"client_id":   r.Form.Get("client_id"),
				"device_code": r.Form.Get("device_code"),
			}
			w.Write([]byte(`{
				"access_token":"mcp_at_test",
				"token_type":"Bearer",
				"expires_in":3600,
				"refresh_token":"mcp_rt_test",
				"scope":"news:search media_lists:manage"
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	withOAuthHooks(t,
		func(rawURL string) error {
			opened = append(opened, rawURL)
			return errors.New("browser unavailable")
		},
		func(d time.Duration) { sleeps = append(sleeps, d) },
		func() time.Time { return time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC) },
	)

	withTempEnv(t, map[string]string{
		"HOME":                         home,
		"NEWSJACK_HOME":                "",
		"NEWSJACK_IGNORE_DOTENV":       "1",
		"NEWSJACK_NO_AUTO_UPDATE":      "1",
		"MEDIALYST_API_KEY":            "",
		"NEWSJACK_MEDIALYST_AUTH_BASE": "",
		"MEDIALYST_AUTH_BASE":          "",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"login", "--base-url", server.URL}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("login code=%d stdout=%s stderr=%s", code, out.String(), errBuf.String())
		}
		if deviceForm["client_id"] != medialystOAuthClientID || deviceForm["scope"] != medialystOAuthDefaultScope {
			t.Fatalf("device form=%#v", deviceForm)
		}
		if tokenForm["grant_type"] != deviceGrantType || tokenForm["client_id"] != medialystOAuthClientID || tokenForm["device_code"] != "mcp_dc_test" {
			t.Fatalf("token form=%#v", tokenForm)
		}
		if len(opened) != 1 || !strings.Contains(opened[0], "/device?user_code=ABCD-EFGH") {
			t.Fatalf("opened=%v", opened)
		}
		if len(sleeps) != 1 || sleeps[0] != 2*time.Second {
			t.Fatalf("sleeps=%v", sleeps)
		}
		text := out.String()
		if !strings.Contains(text, "Could not open a browser automatically") ||
			!strings.Contains(text, "/device") ||
			!strings.Contains(text, "ABCD-EFGH") {
			t.Fatalf("stdout should include fallback URL and user code:\n%s", text)
		}
		creds, ok, err := readCredentialsFile()
		if err != nil || !ok {
			t.Fatalf("read credentials ok=%v err=%v", ok, err)
		}
		if creds.Medialyst.OAuth == nil ||
			creds.Medialyst.OAuth.AccessToken != "mcp_at_test" ||
			creds.Medialyst.OAuth.RefreshToken != "mcp_rt_test" ||
			creds.Medialyst.OAuth.ClientID != medialystOAuthClientID ||
			creds.Medialyst.Source != medialystOAuthSource {
			t.Fatalf("unexpected credentials: %#v", creds.Medialyst)
		}
		assertOwnerOnlyFile(t, credentialsPath())
	})
}

func TestDevicePollHandlesPendingSlowDownAndTerminalErrors(t *testing.T) {
	var sleeps []time.Duration
	withOAuthHooks(t, nil, func(d time.Duration) { sleeps = append(sleeps, d) }, func() time.Time {
		return time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	})
	tokenCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/oauth/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		tokenCalls++
		switch tokenCalls {
		case 1:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"authorization_pending"}`))
		case 2:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"slow_down"}`))
		default:
			w.Write([]byte(`{"access_token":"mcp_at_new","token_type":"Bearer","expires_in":3600,"refresh_token":"mcp_rt_new","scope":"news:search media_lists:manage"}`))
		}
	}))
	defer server.Close()
	token, err := pollMedialystDeviceToken(server.URL, oauthDeviceResponse{DeviceCode: "mcp_dc_test", ExpiresIn: 600, Interval: 2}, 10*time.Minute)
	if err != nil {
		t.Fatalf("poll returned error: %v", err)
	}
	if token.AccessToken != "mcp_at_new" || token.RefreshToken != "mcp_rt_new" {
		t.Fatalf("token=%#v", token)
	}
	wantSleeps := []time.Duration{2 * time.Second, 2 * time.Second, 7 * time.Second}
	if len(sleeps) != len(wantSleeps) {
		t.Fatalf("sleeps=%v, want %v", sleeps, wantSleeps)
	}
	for i := range wantSleeps {
		if sleeps[i] != wantSleeps[i] {
			t.Fatalf("sleeps=%v, want %v", sleeps, wantSleeps)
		}
	}

	for _, tc := range []struct {
		code string
		want string
	}{
		{"expired_token", "expired"},
		{"access_denied", "denied"},
		{"invalid_grant", "no longer valid"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"` + tc.code + `"}`))
			}))
			defer errServer.Close()
			_, err := pollMedialystDeviceToken(errServer.URL, oauthDeviceResponse{DeviceCode: "mcp_dc_test", ExpiresIn: 600, Interval: 1}, time.Minute)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v, want %q", err, tc.want)
			}
		})
	}
}

func TestRefreshRotatesSavedOAuthToken(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	withOAuthHooks(t, nil, func(time.Duration) {}, func() time.Time { return now })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/oauth/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != refreshGrantType ||
			r.Form.Get("client_id") != medialystOAuthClientID ||
			r.Form.Get("refresh_token") != "mcp_rt_old" {
			t.Fatalf("refresh form=%#v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"mcp_at_new","token_type":"Bearer","expires_in":3600,"refresh_token":"mcp_rt_new","scope":"news:search media_lists:manage"}`))
	}))
	defer server.Close()

	withTempEnv(t, map[string]string{
		"HOME":                         home,
		"NEWSJACK_HOME":                "",
		"NEWSJACK_IGNORE_DOTENV":       "1",
		"MEDIALYST_API_KEY":            "",
		"NEWSJACK_MEDIALYST_AUTH_BASE": server.URL,
	}, func() {
		if _, err := writeCredentialsFile(credentialsFile{Medialyst: medialystCredentials{
			APIKey: "mlst_" + strings.Repeat("a", 12),
			OAuth: &medialystOAuthCredentials{
				AccessToken:  "mcp_at_old",
				RefreshToken: "mcp_rt_old",
				TokenType:    "Bearer",
				ExpiresAt:    now.Add(-time.Hour).Format(time.RFC3339),
				Scope:        medialystOAuthDefaultScope,
				ClientID:     medialystOAuthClientID,
			},
			CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
			Source:    medialystOAuthSource,
		}}); err != nil {
			t.Fatal(err)
		}
		cred, err := loadMedialystBearerCredential()
		if err != nil {
			t.Fatalf("load bearer: %v", err)
		}
		if cred.Token != "mcp_at_new" || cred.Kind != "oauth" {
			t.Fatalf("cred=%#v", cred)
		}
		creds, _, err := readCredentialsFile()
		if err != nil {
			t.Fatal(err)
		}
		if creds.Medialyst.APIKey == "" || creds.Medialyst.OAuth.RefreshToken != "mcp_rt_new" {
			t.Fatalf("credentials after refresh=%#v", creds.Medialyst)
		}
	})
}

func TestMedialystAPIRefreshesOAuthAfterUnauthorizedAndRetries(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	withOAuthHooks(t, nil, func(time.Duration) {}, func() time.Time { return now })
	var authHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/news/search":
			authHeaders = append(authHeaders, r.Header.Get("Authorization"))
			if r.Header.Get("Authorization") == "Bearer mcp_at_old" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"message":"expired"}}`))
				return
			}
			if r.Header.Get("Authorization") != "Bearer mcp_at_new" {
				t.Fatalf("unexpected auth header %q", r.Header.Get("Authorization"))
			}
			w.Write([]byte(`{"news":[{"title":"AI funding","link":"https://example.com"}]}`))
		case "/api/oauth/token":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("grant_type") != refreshGrantType || r.Form.Get("refresh_token") != "mcp_rt_old" {
				t.Fatalf("refresh form=%#v", r.Form)
			}
			w.Write([]byte(`{"access_token":"mcp_at_new","token_type":"Bearer","expires_in":3600,"refresh_token":"mcp_rt_new","scope":"news:search media_lists:manage"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	withTempEnv(t, map[string]string{
		"HOME":                         home,
		"NEWSJACK_HOME":                "",
		"NEWSJACK_IGNORE_DOTENV":       "1",
		"MEDIALYST_API_KEY":            "",
		"NEWSJACK_MEDIALYST_API_BASE":  server.URL,
		"NEWSJACK_MEDIALYST_AUTH_BASE": server.URL,
	}, func() {
		if _, err := writeCredentialsFile(credentialsFile{Medialyst: medialystCredentials{
			OAuth: &medialystOAuthCredentials{
				AccessToken:  "mcp_at_old",
				RefreshToken: "mcp_rt_old",
				TokenType:    "Bearer",
				ExpiresAt:    now.Add(time.Hour).Format(time.RFC3339),
				Scope:        medialystOAuthDefaultScope,
				ClientID:     medialystOAuthClientID,
			},
			CreatedAt: now.Add(-time.Hour).Format(time.RFC3339),
			Source:    medialystOAuthSource,
		}}); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"news", "search", "--query", "AI funding"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("news search code=%d stdout=%s stderr=%s", code, out.String(), errBuf.String())
		}
		if len(authHeaders) != 2 || authHeaders[0] != "Bearer mcp_at_old" || authHeaders[1] != "Bearer mcp_at_new" {
			t.Fatalf("auth headers=%v", authHeaders)
		}
		creds, _, err := readCredentialsFile()
		if err != nil {
			t.Fatal(err)
		}
		if creds.Medialyst.OAuth.RefreshToken != "mcp_rt_new" {
			t.Fatalf("refresh token was not rotated: %#v", creds.Medialyst.OAuth)
		}
	})
}

func TestAuthHeadersPrefersOAuthBeforeEnvironmentAPIKey(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	withOAuthHooks(t, nil, nil, func() time.Time { return now })
	withTempEnv(t, map[string]string{
		"HOME":                   home,
		"NEWSJACK_HOME":          "",
		"NEWSJACK_IGNORE_DOTENV": "1",
		"MEDIALYST_API_KEY":      "",
	}, func() {
		if _, err := writeOAuthCredentials(oauthTokenResponse{
			AccessToken:  "mcp_at_saved",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			RefreshToken: "mcp_rt_saved",
			Scope:        medialystOAuthDefaultScope,
		}); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"auth", "headers"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("headers code=%d stderr=%s", code, errBuf.String())
		}
		if out.String() != "{\"Authorization\":\"Bearer mcp_at_saved\"}\n" {
			t.Fatalf("headers=%q", out.String())
		}
	})

	withTempEnv(t, map[string]string{
		"HOME":                   home,
		"NEWSJACK_HOME":          "",
		"NEWSJACK_IGNORE_DOTENV": "1",
		"MEDIALYST_API_KEY":      "mlst_env_key_12345",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"auth", "headers"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("headers code=%d stderr=%s", code, errBuf.String())
		}
		if out.String() != "{\"Authorization\":\"Bearer mcp_at_saved\"}\n" {
			t.Fatalf("headers=%q", out.String())
		}
	})
}

func TestOAuthCredentialsJSONShape(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	withOAuthHooks(t, nil, nil, func() time.Time { return now })
	withTempEnv(t, map[string]string{
		"HOME":              home,
		"NEWSJACK_HOME":     "",
		"MEDIALYST_API_KEY": "",
	}, func() {
		if code := saveMedialystAPIKey("mlst_"+strings.Repeat("a", 12), &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
			t.Fatalf("save API key code=%d", code)
		}
		if _, err := writeOAuthCredentials(oauthTokenResponse{
			AccessToken:  "mcp_at_saved",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			RefreshToken: "mcp_rt_saved",
			Scope:        medialystOAuthDefaultScope,
		}); err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(credentialsPath())
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		med := valueOrEmptyMap(payload["medialyst"])
		oauth := valueOrEmptyMap(med["oauth"])
		if med["api_key"] == "" || oauth["access_token"] != "mcp_at_saved" || oauth["client_id"] != medialystOAuthClientID {
			t.Fatalf("credentials JSON=%s", body)
		}
		if code := saveMedialystAPIKey("mlst_"+strings.Repeat("b", 12), &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
			t.Fatalf("save second API key code=%d", code)
		}
		creds, _, err := readCredentialsFile()
		if err != nil {
			t.Fatal(err)
		}
		if creds.Medialyst.APIKey != "mlst_"+strings.Repeat("b", 12) ||
			creds.Medialyst.OAuth == nil ||
			creds.Medialyst.OAuth.AccessToken != "mcp_at_saved" {
			t.Fatalf("API-key save should preserve OAuth: %#v", creds.Medialyst)
		}
	})
}

func serverDeviceURL(r *http.Request, path string) string {
	return "http://" + r.Host + path
}
