package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	medialystDefaultBase       = "https://medialyst.ai"
	medialystOAuthClientID     = "newsjack-cli"
	medialystOAuthDefaultScope = "news:search media_lists:manage"
	medialystOAuthSource       = "newsjack-oauth-device-flow"
	deviceGrantType            = "urn:ietf:params:oauth:grant-type:device_code"
	refreshGrantType           = "refresh_token"
)

var (
	openBrowserURL = openBrowserURLDefault
	oauthSleep     = time.Sleep
	oauthNow       = time.Now
)

type credentialsFile struct {
	Medialyst medialystCredentials `json:"medialyst,omitempty"`
}

type medialystCredentials struct {
	APIKey    string                     `json:"api_key,omitempty"`
	OAuth     *medialystOAuthCredentials `json:"oauth,omitempty"`
	CreatedAt string                     `json:"created_at,omitempty"`
	Source    string                     `json:"source,omitempty"`
}

type medialystOAuthCredentials struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
}

type medialystBearerCredential struct {
	Token  string
	Source string
	Kind   string
}

type medialystAuthStatus struct {
	Configured       bool
	OAuthConfigured  bool
	APIKeyConfigured bool
	Source           string
	Kind             string
}

type oauthLoginOptions struct {
	BaseURL   string
	Scope     string
	NoBrowser bool
	Timeout   time.Duration
}

type oauthDeviceResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func runMedialystDeviceLogin(opts oauthLoginOptions, stdout, stderr io.Writer) int {
	if strings.TrimSpace(opts.Scope) == "" {
		opts.Scope = medialystOAuthDefaultScope
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Minute
	}
	device, err := requestMedialystDeviceCode(opts.BaseURL, opts.Scope)
	if err != nil {
		return failf(stderr, "Medialyst login could not start: %v", err)
	}
	if err := printDeviceLoginInstructions(device, opts, stdout); err != nil {
		return fail(stderr, err)
	}
	token, err := pollMedialystDeviceToken(opts.BaseURL, device, opts.Timeout)
	if err != nil {
		return failf(stderr, "Medialyst login failed: %v", err)
	}
	path, err := writeOAuthCredentials(token)
	if err != nil {
		return fail(stderr, err)
	}
	uiSuccess(stdout, "authorized Medialyst and saved OAuth credentials to %s", path)
	uiNote(stdout, "newsjack REST commands now use this login automatically.")
	return 0
}

func printDeviceLoginInstructions(device oauthDeviceResponse, opts oauthLoginOptions, stdout io.Writer) error {
	completeURL := strings.TrimSpace(device.VerificationURIComplete)
	if completeURL == "" {
		completeURL = strings.TrimSpace(device.VerificationURI)
	}
	if completeURL == "" || strings.TrimSpace(device.UserCode) == "" {
		return errors.New("device authorization response did not include a verification URL and user code")
	}
	fmt.Fprintln(stdout)
	uiSectionExact(stdout, "Authorize Medialyst")
	uiKV(stdout, "open", completeURL)
	uiKV(stdout, "code", device.UserCode)
	uiNote(stdout, "Approve newsjack CLI in your browser. This device code expires in about %d minutes.", maxInt(device.ExpiresIn, 0)/60)
	if opts.NoBrowser {
		uiNote(stdout, "Browser launch skipped; open the link above to continue.")
		return nil
	}
	if err := openBrowserURL(completeURL); err != nil {
		uiNote(stdout, "Could not open a browser automatically. Open %s and enter %s.", device.VerificationURI, device.UserCode)
		return nil
	}
	uiNote(stdout, "Opened the Medialyst authorization page in your browser.")
	return nil
}

func requestMedialystDeviceCode(baseURL, scope string) (oauthDeviceResponse, error) {
	form := url.Values{}
	form.Set("client_id", medialystOAuthClientID)
	if strings.TrimSpace(scope) != "" {
		form.Set("scope", scope)
	}
	var device oauthDeviceResponse
	if err := postOAuthForm(medialystOAuthEndpoint(baseURL, "/api/oauth/device_authorization"), form, &device); err != nil {
		return oauthDeviceResponse{}, err
	}
	if strings.TrimSpace(device.DeviceCode) == "" {
		return oauthDeviceResponse{}, errors.New("device authorization response did not include device_code")
	}
	return device, nil
}

func pollMedialystDeviceToken(baseURL string, device oauthDeviceResponse, timeout time.Duration) (oauthTokenResponse, error) {
	interval := time.Duration(device.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	expiresIn := time.Duration(device.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = 10 * time.Minute
	}
	if timeout > 0 && timeout < expiresIn {
		expiresIn = timeout
	}
	deadline := oauthNow().Add(expiresIn)
	for {
		if oauthNow().After(deadline) {
			return oauthTokenResponse{}, errors.New("device code expired; run newsjack login again")
		}
		oauthSleep(interval)
		form := url.Values{}
		form.Set("grant_type", deviceGrantType)
		form.Set("client_id", medialystOAuthClientID)
		form.Set("device_code", device.DeviceCode)
		var token oauthTokenResponse
		err := postOAuthForm(medialystOAuthEndpoint(baseURL, "/api/oauth/token"), form, &token)
		if err == nil {
			if strings.TrimSpace(token.AccessToken) == "" || strings.TrimSpace(token.RefreshToken) == "" {
				return oauthTokenResponse{}, errors.New("token response did not include access_token and refresh_token")
			}
			return token, nil
		}
		var oauthErr *oauthEndpointError
		if !errors.As(err, &oauthErr) {
			return oauthTokenResponse{}, err
		}
		switch oauthErr.Code {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "expired_token":
			return oauthTokenResponse{}, errors.New("device code expired; run newsjack login again")
		case "access_denied":
			return oauthTokenResponse{}, errors.New("authorization was denied in the browser")
		case "invalid_grant":
			return oauthTokenResponse{}, errors.New("device code is no longer valid; run newsjack login again")
		default:
			return oauthTokenResponse{}, err
		}
	}
}

type oauthEndpointError struct {
	Status      int
	Code        string
	Description string
	Body        string
}

func (e *oauthEndpointError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Description)
	}
	if e.Code != "" {
		return e.Code
	}
	return fmt.Sprintf("HTTP %d: %s", e.Status, truncate(e.Body, 300))
}

func postOAuthForm(rawURL string, form url.Values, target any) error {
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var payload oauthErrorResponse
		_ = json.Unmarshal(data, &payload)
		return &oauthEndpointError{
			Status:      resp.StatusCode,
			Code:        strings.TrimSpace(payload.Error),
			Description: strings.TrimSpace(payload.ErrorDescription),
			Body:        string(data),
		}
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}

func refreshStoredMedialystOAuth() (medialystBearerCredential, error) {
	creds, ok, err := readCredentialsFile()
	if err != nil {
		return medialystBearerCredential{}, err
	}
	if !ok || creds.Medialyst.OAuth == nil || strings.TrimSpace(creds.Medialyst.OAuth.RefreshToken) == "" {
		return medialystBearerCredential{}, errors.New("no Medialyst OAuth refresh token is saved")
	}
	token, err := refreshMedialystOAuthToken("", creds.Medialyst.OAuth.RefreshToken)
	if err != nil {
		return medialystBearerCredential{}, err
	}
	if _, err := writeOAuthCredentials(token); err != nil {
		return medialystBearerCredential{}, err
	}
	return medialystBearerCredential{
		Token:  token.AccessToken,
		Source: "credentials:" + credentialsPath() + ":oauth",
		Kind:   "oauth",
	}, nil
}

func refreshMedialystOAuthToken(baseURL, refreshToken string) (oauthTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", refreshGrantType)
	form.Set("client_id", medialystOAuthClientID)
	form.Set("refresh_token", refreshToken)
	var token oauthTokenResponse
	if err := postOAuthForm(medialystOAuthEndpoint(baseURL, "/api/oauth/token"), form, &token); err != nil {
		return oauthTokenResponse{}, err
	}
	if strings.TrimSpace(token.AccessToken) == "" || strings.TrimSpace(token.RefreshToken) == "" {
		return oauthTokenResponse{}, errors.New("refresh response did not include access_token and refresh_token")
	}
	return token, nil
}

func loadMedialystBearerCredential() (medialystBearerCredential, error) {
	if cred, ok, err := loadStoredOAuthBearer(); err != nil {
		return medialystBearerCredential{}, err
	} else if ok {
		return cred, nil
	}
	if key, source := loadAPIKey(); key != "" {
		return medialystBearerCredential{Token: key, Source: source, Kind: "api_key"}, nil
	}
	return medialystBearerCredential{}, errors.New("Medialyst credentials not found. Run: newsjack login")
}

func loadStoredOAuthBearer() (medialystBearerCredential, bool, error) {
	creds, ok, err := readCredentialsFile()
	if err != nil || !ok || creds.Medialyst.OAuth == nil {
		return medialystBearerCredential{}, false, err
	}
	oauth := creds.Medialyst.OAuth
	if strings.TrimSpace(oauth.AccessToken) != "" && !oauthTokenExpired(oauth.ExpiresAt) {
		return medialystBearerCredential{
			Token:  strings.TrimSpace(oauth.AccessToken),
			Source: "credentials:" + credentialsPath() + ":oauth",
			Kind:   "oauth",
		}, true, nil
	}
	if strings.TrimSpace(oauth.RefreshToken) == "" {
		return medialystBearerCredential{}, false, nil
	}
	cred, err := refreshStoredMedialystOAuth()
	if err != nil {
		if key, source := loadAPIKey(); key != "" {
			return medialystBearerCredential{Token: key, Source: source, Kind: "api_key"}, true, nil
		}
		return medialystBearerCredential{}, false, err
	}
	return cred, true, nil
}

func oauthTokenExpired(expiresAt string) bool {
	if strings.TrimSpace(expiresAt) == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(expiresAt))
	if err != nil {
		if t, err = time.Parse(time.RFC3339Nano, strings.TrimSpace(expiresAt)); err != nil {
			return true
		}
	}
	return !oauthNow().UTC().Before(t.Add(-30 * time.Second))
}

func loadMedialystAuthStatus() medialystAuthStatus {
	status := medialystAuthStatus{}
	if creds, ok, err := readCredentialsFile(); err == nil && ok {
		if creds.Medialyst.OAuth != nil && (strings.TrimSpace(creds.Medialyst.OAuth.AccessToken) != "" || strings.TrimSpace(creds.Medialyst.OAuth.RefreshToken) != "") {
			status.Configured = true
			status.OAuthConfigured = true
			if status.Source == "" {
				status.Source = "credentials:" + credentialsPath() + ":oauth"
				status.Kind = "oauth"
			}
		}
		if strings.TrimSpace(creds.Medialyst.APIKey) != "" {
			status.Configured = true
			status.APIKeyConfigured = true
			if status.Source == "" {
				status.Source = "credentials:" + credentialsPath()
				status.Kind = "api_key"
			}
		}
	}
	if v := strings.TrimSpace(os.Getenv(envMedialystKey)); v != "" {
		status.Configured = true
		status.APIKeyConfigured = true
		if status.Source == "" {
			status.Source = "environment:" + envMedialystKey
			status.Kind = "api_key"
		}
	}
	if !status.APIKeyConfigured {
		for _, path := range candidateEnvPaths() {
			if v := readDotenvKey(path, envMedialystKey); v != "" {
				status.Configured = true
				status.APIKeyConfigured = true
				if status.Source == "" {
					status.Source = "dotenv:" + path
					status.Kind = "api_key"
				}
				break
			}
		}
	}
	return status
}

func writeOAuthCredentials(token oauthTokenResponse) (string, error) {
	creds, _, err := readCredentialsFile()
	if err != nil {
		return "", err
	}
	now := oauthNow().UTC()
	expiresIn := token.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	tokenType := strings.TrimSpace(token.TokenType)
	if tokenType == "" {
		tokenType = "Bearer"
	}
	creds.Medialyst.OAuth = &medialystOAuthCredentials{
		AccessToken:  strings.TrimSpace(token.AccessToken),
		RefreshToken: strings.TrimSpace(token.RefreshToken),
		TokenType:    tokenType,
		ExpiresAt:    now.Add(time.Duration(expiresIn) * time.Second).Format(time.RFC3339),
		Scope:        strings.TrimSpace(token.Scope),
		ClientID:     medialystOAuthClientID,
	}
	if creds.Medialyst.CreatedAt == "" {
		creds.Medialyst.CreatedAt = now.Format(time.RFC3339Nano)
	}
	creds.Medialyst.Source = medialystOAuthSource
	return writeCredentialsFile(creds)
}

func readCredentialsFile() (credentialsFile, bool, error) {
	path := credentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return credentialsFile{}, false, nil
		}
		return credentialsFile{}, false, err
	}
	var creds credentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return credentialsFile{}, false, err
	}
	return creds, true, nil
}

func writeCredentialsFile(creds credentialsFile) (string, error) {
	path := credentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	data, _ := json.MarshalIndent(creds, "", "  ")
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func medialystOAuthEndpoint(baseURL, path string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = strings.TrimSpace(os.Getenv("NEWSJACK_MEDIALYST_AUTH_BASE"))
	}
	if base == "" {
		base = strings.TrimSpace(os.Getenv("MEDIALYST_AUTH_BASE"))
	}
	if base == "" {
		base = strings.TrimSpace(os.Getenv("NEWSJACK_MEDIALYST_API_BASE"))
	}
	if base == "" {
		base = strings.TrimSpace(os.Getenv("MEDIALYST_API_BASE"))
	}
	if base == "" {
		base = medialystDefaultBase
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/api") {
		base = strings.TrimSuffix(base, "/api")
	}
	return base + path
}

func openBrowserURLDefault(rawURL string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", rawURL).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL).Start()
	default:
		return exec.Command("xdg-open", rawURL).Start()
	}
}

func cloneRequestBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
