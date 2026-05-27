package main

import (
	"bytes"
	"io"
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

func TestBannerUsesProductRenderer(t *testing.T) {
	var out bytes.Buffer
	if code := runCLI([]string{"banner"}, &out, io.Discard); code != 0 {
		t.Fatalf("banner exit code=%d", code)
	}
	want := bannerArt + "\n"
	if out.String() != want {
		t.Fatalf("banner output=%q, want %q", out.String(), want)
	}
}

func TestBareCommandPrintsBanner(t *testing.T) {
	var out bytes.Buffer
	if code := runCLI(nil, &out, io.Discard); code != 0 {
		t.Fatalf("bare command exit code=%d", code)
	}
	if out.String() != bannerArt+"\n" {
		t.Fatalf("bare command output=%q", out.String())
	}
}
