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

func TestProductRendererUsesBannerArt(t *testing.T) {
	var out bytes.Buffer
	uiProduct(&out, "doctor", "system health check")
	text := out.String()
	for _, want := range []string{bannerArt, productTagline, "DOCTOR", "system health check"} {
		if !strings.Contains(text, want) {
			t.Fatalf("product renderer missing %q:\n%s", want, text)
		}
	}
}

func TestBareCommandPrintsUsageLikeHelp(t *testing.T) {
	var bareOut, bareErr, helpOut, helpErr bytes.Buffer
	if code := runCLI(nil, &bareOut, &bareErr); code != 0 {
		t.Fatalf("bare command exit code=%d", code)
	}
	if code := runCLI([]string{"help"}, &helpOut, &helpErr); code != 0 {
		t.Fatalf("help exit code=%d", code)
	}
	if bareOut.String() != helpOut.String() {
		t.Fatalf("bare command output should match help:\nbare=%q\nhelp=%q", bareOut.String(), helpOut.String())
	}
	if strings.Contains(bareOut.String(), "banner") {
		t.Fatalf("usage should not list banner command:\n%s", bareOut.String())
	}
	if bareErr.Len() != 0 || helpErr.Len() != 0 {
		t.Fatalf("bare/help should not write stderr: bare=%q help=%q", bareErr.String(), helpErr.String())
	}
}

func TestBannerCommandIsRemoved(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := runCLI([]string{"banner"}, &out, &errBuf); code == 0 {
		t.Fatalf("banner should not be a command")
	}
	if !strings.Contains(errBuf.String(), "unknown command: banner") {
		t.Fatalf("banner should fail as unknown command:\n%s", errBuf.String())
	}
}
