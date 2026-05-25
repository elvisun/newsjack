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
