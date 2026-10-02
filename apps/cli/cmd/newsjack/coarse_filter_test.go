package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func coarseTestSignal(id, title string, extra map[string]any) map[string]any {
	signal := map[string]any{
		"id":       id,
		"title":    title,
		"query":    "Manus",
		"sources":  []any{"news_search"},
		"routing":  map[string]any{"lane": "profile_relevance"},
		"features": map[string]any{"profile_matches": []any{}},
		"story_size": map[string]any{
			"band": "moderate",
		},
		"evidence": []any{map[string]any{
			"title":        title,
			"url":          "https://example.com/" + id,
			"container":    "Example",
			"source":       "news_search",
			"excerpt":      "Excerpt for " + title,
			"published_at": "2026-09-18T10:00:00Z",
			"metadata":     map[string]any{"publication_type": "news"},
		}},
	}
	for k, v := range extra {
		signal[k] = v
	}
	return signal
}

func coarseTestCandidates() map[string]any {
	return map[string]any{
		"monitor": map[string]any{"profile": map[string]any{
			"company":     "Simular",
			"description": "Computer-use agents",
			"topics":      []any{"computer-use agents"},
			"competitors": []any{"Manus"},
			"standing":    []any{"desktop automation"},
		}},
		"signals": []any{
			coarseTestSignal("s1", "Manus raises a new round", map[string]any{"features": map[string]any{"profile_matches": []any{"Manus"}}}),
			coarseTestSignal("s2", "Best desk lamps of 2026", nil),
			coarseTestSignal("s3", "Malformed answer signal", nil),
			coarseTestSignal("s4", "OpenAI ships a computer-use agent", nil),
			coarseTestSignal("s5", "Desktop automation benchmark results published", nil),
		},
	}
}

func choiceAnswer(choice string, confidence float64, probs map[string]float64) map[string]any {
	p := map[string]any{}
	for k, v := range probs {
		p[k] = v
	}
	return map[string]any{"choice": choice, "confidence": confidence, "probabilities": p}
}

func noulAnswer(v float64) map[string]any { return map[string]any{"noul": v} }

// fakeTypeSafe serves scripted answers keyed by signal id and counts calls.
func fakeTypeSafe(t *testing.T, script func(signalID string, attempt int) (int, map[string]any)) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	var mu sync.Mutex
	attempts := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if r.Header.Get("Authorization") != "Bearer test-typesafe-key" {
			w.WriteHeader(401)
			return
		}
		if r.URL.Path != typesafeSystemOnePath {
			w.WriteHeader(404)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			return
		}
		if len(valueOrEmptyMap(body["questions"])) == 0 || stringValue(body["model"]) == "" {
			w.WriteHeader(400)
			return
		}
		signalID := stringValue(valueOrEmptyMap(valueOrEmptyMap(body["state"])["signal"])["id"])
		mu.Lock()
		attempts[signalID]++
		attempt := attempts[signalID]
		mu.Unlock()
		status, payload := script(signalID, attempt)
		w.WriteHeader(status)
		if payload != nil {
			json.NewEncoder(w).Encode(payload)
		}
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func withFastRetries(t *testing.T) {
	t.Helper()
	oldBase := coarseRetryBase
	coarseRetryBase = time.Millisecond
	t.Cleanup(func() { coarseRetryBase = oldBase })
}

func TestCoarseFilterJevEndToEnd(t *testing.T) {
	withFastRetries(t)
	server, calls := fakeTypeSafe(t, func(id string, attempt int) (int, map[string]any) {
		usage := map[string]any{"input_tokens": 1000, "output_tokens": 0}
		switch id {
		case "s1": // confident reject on a signal with a profile match → floored to monitor_only
			return 200, map[string]any{"model": "jev-2026-09", "usage": usage, "answers": map[string]any{
				"decision": choiceAnswer("reject", 0.9, map[string]float64{"reject": 0.9, "monitor_only": 0.08, "keep": 0.02}),
				"reason":   choiceAnswer("keyword_collision", 0.7, map[string]float64{"keyword_collision": 0.7, "no_profile_bridge": 0.2}),
				"is_news":  noulAnswer(0.9), "profile_bridge": noulAnswer(0.2), "promotional": noulAnswer(0.1), "safety_risk": noulAnswer(0.01),
			}}
		case "s2": // rate-limited once, then a clean reject
			if attempt == 1 {
				return 429, map[string]any{"error": "slow down"}
			}
			return 200, map[string]any{"usage": usage, "answers": map[string]any{
				"decision": choiceAnswer("reject", 0.92, map[string]float64{"reject": 0.92}),
				"reason":   choiceAnswer("not_news", 0.8, map[string]float64{"not_news": 0.8}),
				"is_news":  noulAnswer(0.1), "profile_bridge": noulAnswer(0.05), "promotional": noulAnswer(0.2), "safety_risk": noulAnswer(0.0),
			}}
		case "s3": // malformed: no answers
			return 200, map[string]any{"usage": usage}
		default: // s4, s5 keep
			return 200, map[string]any{"usage": usage, "answers": map[string]any{
				"decision": choiceAnswer("keep", 0.85, map[string]float64{"keep": 0.85}),
				"reason":   choiceAnswer("plausible_client_bridge", 0.6, map[string]float64{"plausible_client_bridge": 0.6}),
				"is_news":  noulAnswer(0.95), "profile_bridge": noulAnswer(0.8), "promotional": noulAnswer(0.05), "safety_risk": noulAnswer(0.0),
			}}
		}
	})

	dir := t.TempDir()
	candidatesPath := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(candidatesPath, marshalJSON(coarseTestCandidates()), 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(dir, "coarse_relevance_decisions.json")

	withTempEnv(t, map[string]string{
		"HOME":                   t.TempDir(),
		"NEWSJACK_HOME":          "",
		"NEWSJACK_IGNORE_DOTENV": "1",
		envTypeSafeKey:           "test-typesafe-key",
		envTypeSafeBaseURL:       server.URL,
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"coarse-filter", "--engine", "jev", "--candidates", candidatesPath, "--output", outputPath, "--concurrency", "2"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("coarse-filter code=%d stderr=%s", code, errBuf.String())
		}
		if !strings.Contains(errBuf.String(), "1 of 5 calls failed") {
			t.Fatalf("expected a failure warning, got: %s", errBuf.String())
		}
	})
	if got := atomic.LoadInt32(calls); got != 6 {
		t.Fatalf("calls=%d, want 6 (5 signals + 1 retry)", got)
	}

	payload, err := readJSONMap(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	engine := valueOrEmptyMap(payload["engine"])
	if engine["name"] != "jev" || engine["model"] != "jev-2026-09" || intValue(engine["failures"], -1) != 1 || intValue(engine["input_tokens"], -1) != 4000 {
		t.Fatalf("unexpected engine block: %#v", engine)
	}
	if cost, _ := numberValue(engine["est_cost_usd"]); cost <= 0 {
		t.Fatalf("est_cost_usd should be positive: %#v", engine)
	}
	decisions := mapSlice(payload["decisions"])
	if len(decisions) != 5 {
		t.Fatalf("decisions=%d, want 5", len(decisions))
	}
	byID := map[string]map[string]any{}
	for _, d := range decisions {
		byID[stringValue(d["signal_id"])] = d
	}
	s1 := byID["s1"]
	if s1["decision"] != "monitor_only" || s1["reason"] != "keyword_collision" || !contains(toStringSlice(s1["post_rules"]), "profile_bridge_floor") {
		t.Fatalf("s1 should be floored to monitor_only by the profile bridge: %#v", s1)
	}
	if !strings.HasPrefix(stringValue(s1["rationale"]), "Jev: reject (0.90)") {
		t.Fatalf("rationale should disclose the engine and raw answer: %s", s1["rationale"])
	}
	s2 := byID["s2"]
	if s2["decision"] != "reject" || s2["reason"] != "not_news" || s2["confidence"] != "high" || len(toStringSlice(s2["post_rules"])) != 0 {
		t.Fatalf("s2 should be a clean high-confidence reject: %#v", s2)
	}
	if urls := toStringSlice(s2["evidence_urls"]); len(urls) != 1 || urls[0] != "https://example.com/s2" {
		t.Fatalf("evidence urls should be preserved: %#v", s2["evidence_urls"])
	}
	s3 := byID["s3"]
	if s3["decision"] != "monitor_only" || s3["confidence"] != "low" || stringValue(s3["error"]) == "" {
		t.Fatalf("s3 should be kept as a low-confidence failure: %#v", s3)
	}
	s4 := byID["s4"]
	if s4["decision"] != "keep" || s4["reason"] != "plausible_client_bridge" || s4["confidence"] != "high" {
		t.Fatalf("s4 should be kept: %#v", s4)
	}

	// The file must be consumable by filter-apply unchanged.
	applied, err := applyDecisions(coarseTestCandidates(), payload, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("filter-apply rejected the jev output: %v", err)
	}
	if got := len(signalSlice(applied["signals"])); got != 4 {
		t.Fatalf("filter-apply selected %d signals, want 4", got)
	}
}

func TestCoarseFilterFailureThresholdExitsNonZero(t *testing.T) {
	withFastRetries(t)
	server, calls := fakeTypeSafe(t, func(string, int) (int, map[string]any) { return 503, map[string]any{"error": "down"} })
	dir := t.TempDir()
	candidatesPath := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(candidatesPath, marshalJSON(coarseTestCandidates()), 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(dir, "decisions.json")
	withTempEnv(t, map[string]string{"HOME": t.TempDir(), "NEWSJACK_HOME": "", "NEWSJACK_IGNORE_DOTENV": "1", envTypeSafeKey: "test-typesafe-key", envTypeSafeBaseURL: server.URL}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"coarse-filter", "--candidates", candidatesPath, "--output", outputPath}, &out, &errBuf)
		if code == 0 {
			t.Fatalf("expected non-zero exit when every call fails; stderr=%s", errBuf.String())
		}
		if !strings.Contains(errBuf.String(), "Fall back to the LLM worker path") {
			t.Fatalf("stderr should tell the skill to fall back: %s", errBuf.String())
		}
	})
	if got := atomic.LoadInt32(calls); got != int32(5*coarseAttempts) {
		t.Fatalf("calls=%d, want %d (every signal retried)", got, 5*coarseAttempts)
	}
	payload, err := readJSONMap(outputPath)
	if err != nil {
		t.Fatalf("output should still be written for inspection: %v", err)
	}
	for _, d := range mapSlice(payload["decisions"]) {
		if d["decision"] != "monitor_only" {
			t.Fatalf("failed signals must be kept: %#v", d)
		}
	}
}

func TestCoarseFilterRequiresKeyAndSupportsDryRun(t *testing.T) {
	server, calls := fakeTypeSafe(t, func(string, int) (int, map[string]any) { return 500, nil })
	dir := t.TempDir()
	candidatesPath := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(candidatesPath, marshalJSON(coarseTestCandidates()), 0o644); err != nil {
		t.Fatal(err)
	}
	withTempEnv(t, map[string]string{"HOME": t.TempDir(), "NEWSJACK_HOME": "", "NEWSJACK_IGNORE_DOTENV": "1", envTypeSafeKey: "", envTypeSafeBaseURL: server.URL}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"coarse-filter", "--candidates", candidatesPath}, &out, &errBuf)
		if code == 0 || !strings.Contains(errBuf.String(), "newsjack auth set-typesafe --key") {
			t.Fatalf("missing key should fail with recovery command; code=%d stderr=%s", code, errBuf.String())
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"coarse-filter", "--candidates", candidatesPath, "--dry-run"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("dry-run code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("dry-run should print JSON: %s", out.String())
		}
		if intValue(payload["signals"], 0) != 5 || intValue(payload["est_input_tokens"], 0) <= 0 {
			t.Fatalf("dry-run estimate missing: %s", out.String())
		}
		sample := valueOrEmptyMap(payload["sample_request"])
		client := valueOrEmptyMap(valueOrEmptyMap(sample["state"])["client"])
		if client["company"] != "Simular" || len(valueOrEmptyMap(sample["questions"])) != 6 {
			t.Fatalf("sample request should carry the profile and six questions: %s", out.String())
		}

		out.Reset()
		code = runCLI([]string{"coarse-filter", "--print-questions"}, &out, &errBuf)
		if code != 0 || !strings.Contains(out.String(), `"no_profile_bridge"`) {
			t.Fatalf("print-questions should dump the embedded set; code=%d out=%s", code, out.String())
		}
	})
	if got := atomic.LoadInt32(calls); got != 0 {
		t.Fatalf("dry-run and print-questions must not call the API; calls=%d", got)
	}
}

func TestCoarseFilterAcceptsHarnessChunkShape(t *testing.T) {
	chunk := map[string]any{
		"profile":         "simular",
		"profile_context": map[string]any{"company": "Simular", "competitors": []any{"Manus"}},
		"signals": []any{map[string]any{
			"signal_id":       "abc",
			"title":           "Manus launches agent",
			"lane":            "profile_relevance",
			"story_size_band": "high",
			"profile_matches": []any{"Manus"},
			"evidence": []any{
				map[string]any{"url": "https://a.example/1", "excerpt": strings.Repeat("x", 900), "publication_type": "newswire"},
				map[string]any{"url": "https://a.example/2"}, map[string]any{"url": "https://a.example/3"},
				map[string]any{"url": "https://a.example/4"}, map[string]any{"url": "https://a.example/5"},
				map[string]any{"url": "https://a.example/6"},
			},
		}},
	}
	profile, err := coarseProfile(chunk, "")
	if err != nil || profile.Company != "Simular" {
		t.Fatalf("profile_context should be accepted: %v %#v", err, profile)
	}
	signals := coarseSignals(chunk)
	if len(signals) != 1 || signals[0].ID != "abc" {
		t.Fatalf("chunk signal not normalized: %#v", signals)
	}
	state := signals[0].State
	if state["lane"] != "profile_relevance" || valueOrEmptyMap(state["story_size"])["band"] != "high" {
		t.Fatalf("flattened chunk fields should map into state: %#v", state)
	}
	evidence := mapSlice(state["evidence"])
	if len(evidence) != coarseMaxEvidence {
		t.Fatalf("evidence should be capped at %d, got %d", coarseMaxEvidence, len(evidence))
	}
	if got := len(stringValue(evidence[0]["excerpt"])); got > coarseMaxExcerptChars+3 {
		t.Fatalf("excerpt should be truncated, got %d chars", got)
	}
	if evidence[0]["publication_type"] != "newswire" || !signals[0].Promotional {
		t.Fatalf("publication_type should be carried and flagged promotional: %#v", evidence[0])
	}
	if len(signals[0].EvidenceURLs) != coarseMaxEvidence {
		t.Fatalf("evidence urls should be capped: %#v", signals[0].EvidenceURLs)
	}
}

func TestCoarsePostRules(t *testing.T) {
	plain := coarseSignal{ID: "x", EvidenceURLs: []string{"https://e/1"}}
	answers := func(decision string, dConf float64, reason string, nouls map[string]float64) map[string]any {
		a := map[string]any{
			"decision": choiceAnswer(decision, dConf, map[string]float64{decision: dConf}),
			"reason":   choiceAnswer(reason, 0.6, map[string]float64{reason: 0.6, "off_beat": 0.3, "relevant_news": 0.1}),
		}
		for k, v := range nouls {
			a[k] = noulAnswer(v)
		}
		return a
	}
	cases := []struct {
		name     string
		signal   coarseSignal
		answers  map[string]any
		decision string
		reason   string
		conf     string
		rules    []string
	}{
		{"clean keep", plain, answers("keep", 0.9, "relevant_news", map[string]float64{"profile_bridge": 0.9}), "keep", "relevant_news", "high", nil},
		{"reason inconsistent with reject", plain, answers("reject", 0.9, "relevant_news", nil), "reject", "off_beat", "medium", []string{"reason_consistency"}},
		{"keep with junk reason", plain, answers("keep", 0.9, "keyword_collision", nil), "keep", "relevant_news", "medium", []string{"reason_consistency"}},
		{"bridge noul floors reject", plain, answers("reject", 0.9, "no_profile_bridge", map[string]float64{"profile_bridge": 0.6}), "monitor_only", "plausible_client_bridge", "medium", []string{"profile_bridge_floor"}},
		{"profile match floors reject", coarseSignal{ID: "x", ProfileMatches: []string{"Manus"}}, answers("reject", 0.9, "off_beat", map[string]float64{"profile_bridge": 0.1}), "monitor_only", "off_beat", "medium", []string{"profile_bridge_floor"}},
		{"safety floors keep", plain, answers("keep", 0.9, "relevant_news", map[string]float64{"safety_risk": 0.8}), "monitor_only", "safety_risk", "medium", []string{"safety_floor"}},
		{"promotional noul floors reject", plain, answers("reject", 0.9, "not_news", map[string]float64{"promotional": 0.75}), "monitor_only", "competitor_or_promotional", "medium", []string{"promotional_floor"}},
		{"promotional metadata floors reject", coarseSignal{ID: "x", Promotional: true}, answers("reject", 0.9, "not_news", nil), "monitor_only", "competitor_or_promotional", "medium", []string{"promotional_floor"}},
		{"promotional reason parks keep", plain, answers("keep", 0.9, "competitor_or_promotional", map[string]float64{"promotional": 0.86}), "monitor_only", "competitor_or_promotional", "medium", []string{"reason_consistency", "promotional_floor"}},
		{"mild promotional noul leaves keep alone", plain, answers("keep", 0.9, "relevant_news", map[string]float64{"promotional": 0.46}), "keep", "relevant_news", "high", nil},
		{"uncertain reject floors", plain, answers("reject", 0.5, "off_beat", nil), "monitor_only", "off_beat", "low", []string{"uncertainty_floor"}},
		{"confident reject passes", plain, answers("reject", 0.7, "seo_landing_page", map[string]float64{"is_news": 0.1}), "reject", "seo_landing_page", "medium", nil},
		{"unknown decision normalized", plain, map[string]any{"decision": map[string]any{"choice": "maybe", "probabilities": map[string]any{"keep": 0.3, "monitor_only": 0.5, "reject": 0.2}}, "reason": choiceAnswer("major_news_no_bridge", 0.5, nil)}, "monitor_only", "major_news_no_bridge", "medium", []string{"decision_normalized"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := coarseDecisionFromAnswers(tc.signal, tc.answers)
			if got["decision"] != tc.decision || got["reason"] != tc.reason || got["confidence"] != tc.conf {
				t.Fatalf("got decision=%v reason=%v confidence=%v, want %s/%s/%s (%s)", got["decision"], got["reason"], got["confidence"], tc.decision, tc.reason, tc.conf, got["rationale"])
			}
			rules := toStringSlice(got["post_rules"])
			if len(rules) != len(tc.rules) {
				t.Fatalf("post_rules=%v, want %v", rules, tc.rules)
			}
			for i := range rules {
				if rules[i] != tc.rules[i] {
					t.Fatalf("post_rules=%v, want %v", rules, tc.rules)
				}
			}
			if got["engine"] != "jev" || !strings.HasPrefix(stringValue(got["rationale"]), "Jev:") {
				t.Fatalf("decision must disclose the engine: %#v", got)
			}
		})
	}
}

func TestAuthSetTypeSafeAndDoctorReportIt(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":                     t.TempDir(),
		"NEWSJACK_HOME":            "",
		"NEWSJACK_ROOT":            repo,
		"MEDIALYST_API_KEY":        "",
		"X_BEARER_TOKEN":           "",
		"TWITTER_BEARER_TOKEN":     "",
		"X_API_BEARER_TOKEN":       "",
		"TWITTER_API_BEARER_TOKEN": "",
		envTypeSafeKey:             "",
		"PATH":                     t.TempDir(),
	}, func() {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(cwd) })
		if err := os.Chdir(t.TempDir()); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{"auth", "set-typesafe", "--key", "ts-secret"}, &out, &errBuf); code != 0 {
			t.Fatalf("auth set-typesafe code=%d stderr=%s", code, errBuf.String())
		}
		if !strings.Contains(out.String(), "Jev coarse filtering") {
			t.Fatalf("auth set-typesafe should explain usage: %s", out.String())
		}
		key, source := loadTypeSafeAPIKey()
		if key != "ts-secret" || !strings.HasPrefix(source, "dotenv:") {
			t.Fatalf("typesafe key/source=%q/%q", key, source)
		}
		assertOwnerOnlyFile(t, newsjackEnvPath())

		out.Reset()
		runCLI([]string{"auth", "status"}, &out, &errBuf)
		var status map[string]any
		if json.Unmarshal(out.Bytes(), &status) != nil || status["typesafe_configured"] != true {
			t.Fatalf("auth status should report typesafe configured: %s", out.String())
		}

		out.Reset()
		if code := runCLI([]string{"doctor", "--json"}, &out, &errBuf); code != 0 {
			t.Fatalf("doctor code=%d stderr=%s", code, errBuf.String())
		}
		var doctor map[string]any
		if json.Unmarshal(out.Bytes(), &doctor) != nil {
			t.Fatalf("invalid doctor JSON: %s", out.String())
		}
		if valueOrEmptyMap(doctor["auth"])["typesafe_configured"] != true {
			t.Fatalf("doctor should report typesafe configured: %s", out.String())
		}
		// Optional integration: never a warning or an action.
		for _, w := range toStringSlice(doctor["warnings"]) {
			if strings.Contains(strings.ToLower(w), "typesafe") {
				t.Fatalf("typesafe must not produce a doctor warning: %s", w)
			}
		}
		out.Reset()
		runCLI([]string{"doctor"}, &out, &errBuf)
		if !strings.Contains(out.String(), "TypeSafe (Jev)") {
			t.Fatalf("human doctor output should list TypeSafe: %s", out.String())
		}
	})
}
