package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUIColorHonorsForcedAndBufferedModes(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_COLOR": "", "NO_COLOR": ""}, func() {
		var out bytes.Buffer
		uiSuccess(&out, "ready")
		if strings.Contains(out.String(), "\x1b[") {
			t.Fatalf("buffered output should not contain ANSI color: %q", out.String())
		}
	})

	withTempEnv(t, map[string]string{"NEWSJACK_COLOR": "always", "NO_COLOR": ""}, func() {
		var out bytes.Buffer
		uiSuccess(&out, "ready")
		if !strings.Contains(out.String(), "\x1b[32m[success]\x1b[0m ready") {
			t.Fatalf("forced color missing from output: %q", out.String())
		}
	})

	withTempEnv(t, map[string]string{"NEWSJACK_COLOR": "", "NO_COLOR": "1"}, func() {
		var out bytes.Buffer
		uiSuccess(&out, "ready")
		if strings.Contains(out.String(), "\x1b[") {
			t.Fatalf("NO_COLOR should suppress ANSI color: %q", out.String())
		}
	})
}
