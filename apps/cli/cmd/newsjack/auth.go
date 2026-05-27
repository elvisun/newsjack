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
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(stderr)
	key := fs.String("key", "", "Medialyst API key")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	apiKey := strings.TrimSpace(*key)
	if apiKey == "" {
		fmt.Fprint(stderr, "Medialyst API key: ")
		var line string
		if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
			return fail(stderr, err)
		}
		apiKey = strings.TrimSpace(line)
	}
	if err := validateAPIKey(apiKey); err != nil {
		return failf(stderr, "invalid key: %v", err)
	}
	path, err := writeCredentials(apiKey)
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Saved Medialyst credentials to %s\n", path)
	fmt.Fprintln(stdout, "MCP-compatible runtimes can now use newsjack mcp-bridge without MEDIALYST_API_KEY exports.")
	return 0
}

func cmdAuth(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, errors.New("usage: newsjack auth status|headers|logout"))
	}
	switch args[0] {
	case "status":
		key, source := loadAPIKey()
		payload := map[string]any{"configured": key != "", "source": nullableString(source)}
		writeJSON(stdout, payload)
		if key == "" {
			return 1
		}
		return 0
	case "headers":
		key, source := loadAPIKey()
		if key == "" {
			fmt.Fprintln(stderr, "Medialyst API key not found. Run: ~/.newsjack/bin/newsjack login")
			return 1
		}
		writeJSONCompact(stdout, map[string]string{"Authorization": "Bearer " + key})
		if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
			fmt.Fprintf(stderr, "Loaded Medialyst API key from %s\n", source)
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

func credentialsPath() string {
	return filepath.Join(newsjackHome(), "credentials.json")
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
	path := credentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	payload := map[string]any{
		"medialyst": map[string]any{
			"api_key":    apiKey,
			"created_at": time.Now().UTC().Format(time.RFC3339Nano),
			"source":     "newsjack-local-login",
		},
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", err
	}
	return path, nil
}
