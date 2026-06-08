package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoverageInitStatusOpenAndRecord(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "tracker.json")
	if err := os.WriteFile(configPath, []byte(`{
  "name": "Profound",
  "lookback_days": 2,
  "keywords": [
    {"keyword": "profound", "means": "Profound, the AI search analytics company."}
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_STORE":          filepath.Join(home, "coverage-test.db"),
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"coverage", "init", "--config", configPath}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage init code=%d stderr=%s", code, errBuf.String())
		}
		var initPayload map[string]any
		if json.Unmarshal(out.Bytes(), &initPayload) != nil {
			t.Fatalf("invalid init JSON: %s", out.String())
		}
		if initPayload["slug"] != "profound" {
			t.Fatalf("slug=%v, want profound", initPayload["slug"])
		}
		configDest := filepath.Join(home, ".newsjack", "coverage", "profound", "tracker.json")
		if !fileExists(configDest) {
			t.Fatalf("tracker config was not saved: %s", configDest)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "status", "profound"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage status code=%d stderr=%s", code, errBuf.String())
		}
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil {
			t.Fatalf("invalid status JSON: %s", out.String())
		}
		if status["keyword_count"] != float64(1) || status["stored_alerts"] != float64(0) {
			t.Fatalf("unexpected initial status: %#v", status)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "list"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage list code=%d stderr=%s", code, errBuf.String())
		}
		var list []map[string]any
		if json.Unmarshal(out.Bytes(), &list) != nil {
			t.Fatalf("invalid list JSON: %s", out.String())
		}
		if len(list) != 1 || list[0]["slug"] != "profound" {
			t.Fatalf("unexpected coverage list: %#v", list)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "open", "profound"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage open code=%d stderr=%s", code, errBuf.String())
		}
		if got := strings.TrimSpace(out.String()); got != filepath.Join(home, ".newsjack", "coverage", "profound") {
			t.Fatalf("coverage open=%q", got)
		}

		runDir := filepath.Join(home, ".newsjack", "coverage", "profound", "runs", "20260605T120000Z")
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			t.Fatal(err)
		}

		decisionsPath := filepath.Join(t.TempDir(), "decisions.json")
		if err := os.WriteFile(decisionsPath, []byte(`{
  "items": [
    {
      "keyword": "profound",
      "title": "Profound launches coverage tracker",
      "url": "https://example.com/story?utm_source=test#section",
      "outlet": "Example News",
      "published_at": "2026-06-05T12:00:00Z",
      "verdict": "real_feature",
      "confidence": "high",
      "alert": true,
      "rationale": "This is about Profound the company."
    },
    {
      "keyword": "profound",
      "title": "A profound change in weather",
      "url": "https://example.com/weather",
      "outlet": "Example News",
      "published_at": "2026-06-05T12:05:00Z",
      "verdict": "wrong_entity",
      "confidence": "high",
      "alert": false,
      "rationale": "Generic adjective use."
    }
  ]
}`), 0o644); err != nil {
			t.Fatal(err)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "check", "profound", "--input", decisionsPath}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage check before record code=%d stderr=%s", code, errBuf.String())
		}
		var check map[string]any
		if json.Unmarshal(out.Bytes(), &check) != nil {
			t.Fatalf("invalid check JSON: %s", out.String())
		}
		if check["known_count"] != float64(0) || check["unknown_count"] != float64(2) {
			t.Fatalf("check before record should mark both unknown: %#v", check)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "record", "profound", "--input", decisionsPath, "--run-dir", runDir}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage record code=%d stderr=%s", code, errBuf.String())
		}
		var record map[string]any
		if json.Unmarshal(out.Bytes(), &record) != nil {
			t.Fatalf("invalid record JSON: %s", out.String())
		}
		if record["new_alert_count"] != float64(1) || record["recorded_items"] != float64(2) {
			t.Fatalf("unexpected record payload: %#v", record)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "record", "profound", "--input", decisionsPath, "--run-dir", runDir}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("second coverage record code=%d stderr=%s", code, errBuf.String())
		}
		if json.Unmarshal(out.Bytes(), &record) != nil {
			t.Fatalf("invalid second record JSON: %s", out.String())
		}
		if record["new_alert_count"] != float64(0) {
			t.Fatalf("repeat record should suppress alert: %#v", record)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "status", "profound"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage status after record code=%d stderr=%s", code, errBuf.String())
		}
		if json.Unmarshal(out.Bytes(), &status) != nil {
			t.Fatalf("invalid status JSON: %s", out.String())
		}
		if status["stored_articles"] != float64(2) || status["stored_alerts"] != float64(1) {
			t.Fatalf("status should include stored counts: %#v", status)
		}
		if stringValue(status["latest_run_dir"]) != runDir {
			t.Fatalf("latest_run_dir=%v, want %s", status["latest_run_dir"], runDir)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "open", "profound"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage open after alert code=%d stderr=%s", code, errBuf.String())
		}
		if got := strings.TrimSpace(out.String()); got != runDir {
			t.Fatalf("coverage open after run=%q, want %s", got, runDir)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coverage", "check", "profound", "--input", decisionsPath}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coverage check after record code=%d stderr=%s", code, errBuf.String())
		}
		if json.Unmarshal(out.Bytes(), &check) != nil {
			t.Fatalf("invalid check after record JSON: %s", out.String())
		}
		if check["known_count"] != float64(2) || check["unknown_count"] != float64(0) {
			t.Fatalf("check after record should mark both known: %#v", check)
		}
		known := mapSlice(check["known_items"])
		if len(known) != 2 {
			t.Fatalf("known_items length=%d, want 2: %#v", len(known), check)
		}
		firstDecision := valueOrEmptyMap(known[0]["prior_decision"])
		if firstDecision["verdict"] != "real_feature" || firstDecision["alert"] != true {
			t.Fatalf("first prior_decision mismatch: %#v", firstDecision)
		}
		if valueOrEmptyMap(known[0]["prior_alert"])["alert_count"] != float64(2) {
			t.Fatalf("first prior_alert should reflect repeated alert sighting: %#v", known[0])
		}
	})
}

func TestCoverageInitRejectsMissingMeaning(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "tracker.json")
	if err := os.WriteFile(configPath, []byte(`{"name":"Bad","keywords":[{"keyword":"ambiguous"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"coverage", "init", "bad", "--config", configPath}, &out, &errBuf)
		if code == 0 {
			t.Fatalf("coverage init should fail for missing means; stdout=%s", out.String())
		}
		if !strings.Contains(errBuf.String(), "missing means") {
			t.Fatalf("stderr should explain missing means:\n%s", errBuf.String())
		}
	})
}

func TestCoverageRecordAcceptsEmptyItems(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "tracker.json")
	if err := os.WriteFile(configPath, []byte(`{"name":"Empty","keywords":[{"keyword":"empty","means":"Empty, the example company."}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	decisionsPath := filepath.Join(t.TempDir(), "decisions.json")
	if err := os.WriteFile(decisionsPath, []byte(`{"items":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_ROOT":           repo,
		"NEWSJACK_NO_AUTO_UPDATE": "1",
		"NEWSJACK_STORE":          filepath.Join(home, "coverage-empty.db"),
	}, func() {
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{"coverage", "init", "empty", "--config", configPath}, &out, &errBuf); code != 0 {
			t.Fatalf("coverage init code=%d stderr=%s", code, errBuf.String())
		}
		out.Reset()
		errBuf.Reset()
		if code := runCLI([]string{"coverage", "record", "empty", "--input", decisionsPath}, &out, &errBuf); code != 0 {
			t.Fatalf("coverage record empty code=%d stderr=%s", code, errBuf.String())
		}
		var record map[string]any
		if json.Unmarshal(out.Bytes(), &record) != nil {
			t.Fatalf("invalid record JSON: %s", out.String())
		}
		if record["recorded_items"] != float64(0) || record["new_alert_count"] != float64(0) {
			t.Fatalf("empty record should be zero-count: %#v", record)
		}
		if alerts := anySlice(record["new_alerts"]); len(alerts) != 0 {
			t.Fatalf("empty record should return empty new_alerts: %#v", record)
		}
	})
}
