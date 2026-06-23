package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func cmdLogin(args []string, stdout, stderr io.Writer) int {
	if restHelpRequested(args) {
		printCommandHelp(stdout, "login")
		return 0
	}
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(stderr)
	key := fs.String("key", "", "Save a Medialyst API key instead of using browser login")
	scope := fs.String("scope", medialystOAuthDefaultScope, "OAuth scopes for browser login")
	baseURL := fs.String("base-url", "", "Medialyst base URL for OAuth login")
	noBrowser := fs.Bool("no-browser", false, "Print the login URL without opening a browser")
	timeoutSeconds := fs.Int("timeout", 600, "OAuth login timeout in seconds")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	apiKey := strings.TrimSpace(*key)
	if apiKey != "" {
		if err := validateAPIKey(apiKey); err != nil {
			return failf(stderr, "invalid key: %v", err)
		}
		return saveMedialystAPIKey(apiKey, stdout, stderr)
	}
	return runMedialystDeviceLogin(oauthLoginOptions{
		BaseURL:   strings.TrimSpace(*baseURL),
		Scope:     strings.TrimSpace(*scope),
		NoBrowser: *noBrowser,
		Timeout:   time.Duration(*timeoutSeconds) * time.Second,
	}, stdout, stderr)
}

func cmdAuth(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printAuthHelp(stderr)
		return fail(stderr, errors.New("usage: newsjack auth status|set|set-medialyst|set-x|headers|logout"))
	}
	switch args[0] {
	case "--help", "-h", "help":
		printAuthHelp(stdout)
		return 0
	case "status":
		medialystStatus := loadMedialystAuthStatus()
		xToken, xSource := loadXBearerToken()
		payload := map[string]any{
			"configured":                    medialystStatus.Configured || xToken != "",
			"medialyst_configured":          medialystStatus.Configured,
			"medialyst_oauth_configured":    medialystStatus.OAuthConfigured,
			"medialyst_api_key_configured":  medialystStatus.APIKeyConfigured,
			"medialyst_source":              nullableString(medialystStatus.Source),
			"medialyst_auth_type":           nullableString(medialystStatus.Kind),
			"x_api_configured":              xToken != "",
			"x_bearer_token_source":         nullableString(xSource),
			"medialyst_get_key_url":         medialystAPIKeyURL,
			"medialyst_login_command":       "newsjack login",
			"medialyst_set_command":         "newsjack login",
			"medialyst_api_key_set_command": "newsjack auth set-medialyst --key <mlst_...>",
			"x_bearer_token_command":        "newsjack auth set-x --bearer-token <token>",
		}
		writeJSON(stdout, payload)
		if !medialystStatus.Configured && xToken == "" {
			return 1
		}
		return 0
	case "set":
		return cmdAuthSet(args[1:], stdout, stderr)
	case "set-medialyst":
		return cmdAuthSetMedialyst(args[1:], stdout, stderr)
	case "set-x":
		return cmdAuthSetX(args[1:], stdout, stderr)
	case "headers":
		cred, credErr := loadMedialystBearerCredential()
		if credErr != nil || cred.Token == "" {
			fmt.Fprintln(stderr, "Medialyst credentials not found. Run: newsjack login")
			fmt.Fprintln(stderr, "For API-key automation, run: newsjack auth set-medialyst --key <mlst_...>")
			return 1
		}
		writeJSONCompact(stdout, map[string]string{"Authorization": "Bearer " + cred.Token})
		if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
			fmt.Fprintf(stderr, "Loaded Medialyst %s credentials from %s\n", cred.Kind, cred.Source)
		}
		return 0
	case "logout":
		path := credentialsPath()
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintln(stdout, "No saved Medialyst credentials found.")
				return 0
			}
			return fail(stderr, err)
		}
		fmt.Fprintf(stdout, "Removed saved Medialyst credentials at %s\n", path)
		return 0
	default:
		return failf(stderr, "unknown auth command: %s", args[0])
	}
}

func cmdAuthSet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("auth set", flag.ContinueOnError)
	fs.SetOutput(stderr)
	medialystKey := fs.String("medialyst-key", "", "Medialyst API key")
	xBearerToken := fs.String("x-bearer-token", "", "X API bearer token")
	xToken := fs.String("x-token", "", "Alias for --x-bearer-token")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	medialystValue := strings.TrimSpace(*medialystKey)
	xValue := strings.TrimSpace(firstString(*xBearerToken, *xToken))
	if medialystValue == "" && xValue == "" {
		return fail(stderr, errors.New("usage: newsjack auth set [--medialyst-key KEY] [--x-bearer-token TOKEN]"))
	}
	if medialystValue != "" {
		if code := saveMedialystAPIKey(medialystValue, stdout, stderr); code != 0 {
			return code
		}
	}
	if xValue != "" {
		if code := saveXBearerToken(xValue, stdout, stderr); code != 0 {
			return code
		}
	}
	return 0
}

func cmdAuthSetMedialyst(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("auth set-medialyst", flag.ContinueOnError)
	fs.SetOutput(stderr)
	key := fs.String("key", "", "Medialyst API key")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	apiKey := strings.TrimSpace(*key)
	if apiKey == "" {
		return fail(stderr, errors.New("usage: newsjack auth set-medialyst --key <mlst_...>"))
	}
	return saveMedialystAPIKey(apiKey, stdout, stderr)
}

func cmdAuthSetX(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("auth set-x", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bearerToken := fs.String("bearer-token", "", "X API bearer token")
	token := fs.String("token", "", "Alias for --bearer-token")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	value := strings.TrimSpace(firstString(*bearerToken, *token))
	if value == "" {
		return fail(stderr, errors.New("usage: newsjack auth set-x --bearer-token <token>"))
	}
	return saveXBearerToken(value, stdout, stderr)
}

func saveMedialystAPIKey(apiKey string, stdout, stderr io.Writer) int {
	if err := validateAPIKey(apiKey); err != nil {
		return failf(stderr, "invalid key: %v", err)
	}
	path, err := writeCredentials(apiKey)
	if err != nil {
		return fail(stderr, err)
	}
	uiSuccess(stdout, "saved Medialyst credentials to %s", path)
	uiNote(stdout, "used for: live news search and journalist enrichment")
	uiNote(stdout, "newsjack REST commands now use the saved key without MEDIALYST_API_KEY exports.")
	return 0
}

func saveXBearerToken(token string, stdout, stderr io.Writer) int {
	if err := writeNewsjackEnv(map[string]string{envXBearerToken: token}); err != nil {
		return fail(stderr, err)
	}
	uiSuccess(stdout, "saved X API bearer token to %s", newsjackEnvPath())
	uiNote(stdout, "used for: X News, X trends, and X post search")
	return 0
}

func credentialsPath() string {
	return filepath.Join(newsjackHome(), "credentials.json")
}

func newsjackEnvPath() string {
	return filepath.Join(newsjackHome(), ".env")
}

func loadAPIKey() (string, string) {
	if v := strings.TrimSpace(os.Getenv(envMedialystKey)); v != "" {
		return v, "environment:" + envMedialystKey
	}
	path := credentialsPath()
	if data, err := os.ReadFile(path); err == nil {
		var payload map[string]any
		if json.Unmarshal(data, &payload) == nil {
			if med, ok := payload["medialyst"].(map[string]any); ok {
				if v, ok := med["api_key"].(string); ok && strings.TrimSpace(v) != "" {
					return strings.TrimSpace(v), "credentials:" + path
				}
			}
			if v, ok := payload[envMedialystKey].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), "credentials:" + path
			}
		}
	}
	for _, path := range candidateEnvPaths() {
		if v := readDotenvKey(path, envMedialystKey); v != "" {
			return v, "dotenv:" + path
		}
	}
	return "", ""
}

func loadXBearerToken() (string, string) {
	for _, key := range xBearerEnvKeys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v, "environment:" + key
		}
	}
	for _, path := range candidateEnvPaths() {
		for _, key := range xBearerEnvKeys {
			if v := readDotenvKey(path, key); v != "" {
				return v, "dotenv:" + path + ":" + key
			}
		}
	}
	return "", ""
}

func candidateEnvPaths() []string {
	if os.Getenv("NEWSJACK_IGNORE_DOTENV") == "1" {
		return nil
	}
	var out []string
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for {
			out = append(out, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	out = append(out, filepath.Join(newsjackHome(), ".env"))
	return out
}

func readDotenvKey(path, key string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		k, v, _ := strings.Cut(line, "=")
		if strings.TrimSpace(k) != key {
			continue
		}
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if v != "" {
			return v
		}
	}
	return ""
}

func validateAPIKey(key string) error {
	if !strings.HasPrefix(key, "mlst_") {
		return errors.New("Medialyst API keys should start with 'mlst_'")
	}
	if len(key) < 12 {
		return errors.New("Medialyst API key is too short")
	}
	return nil
}

func writeCredentials(apiKey string) (string, error) {
	creds, err := readCredentialsFileForWrite()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	creds.Medialyst.APIKey = apiKey
	creds.Medialyst.OAuth = nil
	if creds.Medialyst.CreatedAt == "" {
		creds.Medialyst.CreatedAt = now.Format(time.RFC3339Nano)
	}
	creds.Medialyst.Source = "newsjack-local-login"
	return writeCredentialsFile(creds)
}
