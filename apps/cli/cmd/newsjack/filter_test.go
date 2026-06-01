package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestFilterApplyFixture(t *testing.T) {
	repo := repoRootForTest(t)
	var out, err bytes.Buffer
	code := runCLI([]string{
		"filter-apply",
		"--candidates", filepath.Join(repo, "fixtures/golden/go-port/detector/mock-query.json"),
		"--decisions", filepath.Join(repo, "fixtures/golden/go-port/filter/decisions.json"),
		"--include", "keep",
	}, &out, &err)
	if code != 0 {
		t.Fatalf("filter code=%d stderr=%s", code, err.String())
	}
	var payload map[string]any
	if json.Unmarshal(out.Bytes(), &payload) != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 2 {
		t.Fatalf("selected signals=%d, want 2", len(signals))
	}
	coarse := valueOrEmptyMap(payload["coarse_relevance"])
	if coarse["selected_count"].(float64) != 2 || coarse["rejected_count"].(float64) != 0 {
		t.Fatalf("unexpected coarse summary: %#v", coarse)
	}
}

func TestFilterApplyIgnoresOriginFields(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{},
		"signals": []any{
			map[string]any{"id": "fresh", "title": "Fresh story", "evidence": []any{map[string]any{"url": "https://example.com/fresh"}}},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "fresh",
				"decision":  "keep",
				"reason":    "relevant_news",
				"first_publication": map[string]any{
					"status":                          "fresh",
					"surfaced_article_published_at":   "2026-05-25T12:10:00Z",
					"first_public_at":                 "2026-05-25T12:00:00Z",
					"original_url":                    "https://example.com/original",
					"canonical_coverage_url":          "https://example.com/major-coverage",
					"canonical_coverage_source":       "Example Major",
					"canonical_coverage_published_at": "2026-05-25T12:05:00Z",
					"confidence":                      "high",
					"rationale":                       "Original source is within the cron window.",
					"evidence_urls":                   []any{"https://example.com/original", "https://example.com/major-coverage"},
					"same_story_basis":                "same official action and same named actors",
				},
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals, _ := payload["signals"].([]map[string]any)
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1", len(signals))
	}
	coarse := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if _, ok := coarse["first_publication"]; ok {
		t.Fatalf("coarse relevance preserved story-origin data: %#v", coarse)
	}
}

func TestFilterApplyAllowsPromptOwnedReasons(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{},
		"signals": []any{
			map[string]any{"id": "fresh", "title": "Fresh story"},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{"signal_id": "fresh", "decision": "keep", "reason": "new_prompt_owned_reason"},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions should allow prompt-owned reasons: %v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1", len(signals))
	}
	decision := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if decision["reason"] != "new_prompt_owned_reason" {
		t.Fatalf("reason=%v, want custom reason", decision["reason"])
	}
}

func TestFilterApplyGuardKeepsProfileMatchedNoBridgeReject(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{
			"profile": map[string]any{
				"company":     "Simular",
				"competitors": []any{"Manus"},
			},
		},
		"signals": []any{
			map[string]any{
				"id":              "manus",
				"title":           "China blocks Meta AI deal",
				"profile_matches": []any{"Manus"},
				"mechanical_scores": map[string]any{
					"profile_match": 0.12,
				},
				"evidence": []any{
					map[string]any{
						"title":   "China blocks Meta AI deal targeting Manus",
						"excerpt": "Regulators blocked Meta from acquiring Manus.",
						"url":     "https://example.com/manus",
					},
				},
			},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "manus",
				"decision":  "reject",
				"reason":    "no_profile_bridge",
				"rationale": "Meta AI does not match the profile.",
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1; payload=%#v", len(signals), payload)
	}
	decision := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if decision["decision"] != "monitor_only" || decision["reason"] != "plausible_client_bridge" {
		t.Fatalf("guarded decision mismatch: %#v", decision)
	}
	if decision["guardrail"] != "profile_match_recall_guard" {
		t.Fatalf("missing guardrail marker: %#v", decision)
	}
	coarse := valueOrEmptyMap(payload["coarse_relevance"])
	if coarse["guardrail_override_count"] != 1 {
		t.Fatalf("guardrail count=%v, want 1", coarse["guardrail_override_count"])
	}
}

func TestFilterApplyGuardKeepsLargeRemoteRelevantNoBridgeReject(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{
			"profile": map[string]any{
				"company": "Specright",
				"topics":  []any{"product safety compliance", "marketplace product compliance"},
			},
		},
		"signals": []any{
			map[string]any{
				"id":    "temu",
				"title": "Temu hit with EU fine over unsafe goods",
				"routing": map[string]any{
					"lane":           "profile_relevance_weak",
					"queue_priority": 46.0,
					"recall_guard":   "large_story_remote_relevance",
				},
				"mechanical_scores": map[string]any{
					"profile_match": 0.029,
					"story_size":    0.587,
				},
				"evidence": []any{
					map[string]any{
						"title":   "Temu hit with EU fine over unsafe goods",
						"excerpt": "Regulators say unsafe products were sold through the online marketplace.",
						"url":     "https://example.com/temu-fine",
					},
				},
			},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "temu",
				"decision":  "reject",
				"reason":    "major_news_no_bridge",
				"rationale": "Temu is not the client.",
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1; payload=%#v", len(signals), payload)
	}
	// story_size 0.587 → major band, so the big-story recall guard keeps it
	// directly (no profile bridge required): a big story is never hard-dropped.
	decision := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if decision["decision"] != "monitor_only" || decision["reason"] != "big_story_surfaced" {
		t.Fatalf("guarded decision mismatch: %#v", decision)
	}
	if decision["guardrail"] != "big_story_recall" {
		t.Fatalf("guardrail marker missing big_story_recall: %#v", decision)
	}
}

func TestFilterApplyGuardIgnoresDefaultProfileMatchWithoutProfileContext(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"profile": map[string]any{}},
		"signals": []any{
			map[string]any{
				"id": "generic",
				"mechanical_scores": map[string]any{
					"profile_match": 0.4,
				},
			},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "generic",
				"decision":  "reject",
				"reason":    "no_profile_bridge",
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	if got := len(signalSlice(payload["signals"])); got != 0 {
		t.Fatalf("selected signals=%d, want 0", got)
	}
	coarse := valueOrEmptyMap(payload["coarse_relevance"])
	if coarse["guardrail_override_count"] != 0 {
		t.Fatalf("guardrail count=%v, want 0", coarse["guardrail_override_count"])
	}
}

func TestFilterApplyNeverDropsBigStory(t *testing.T) {
	// A high/major story_size signal must survive coarse rejection regardless of
	// the worker's reason (here keyword_collision, which the profile-match guard
	// would NOT have rescued). It is downgraded to monitor_only, not dropped.
	candidates := map[string]any{
		"monitor": map[string]any{"profile": map[string]any{"company": "Property Saviour"}},
		"signals": []any{
			map[string]any{
				"id":         "opera",
				"title":      "John Cleese Opera House tickets out now",
				"story_size": map[string]any{"band": "high"},
				"evidence": []any{
					map[string]any{"url": "https://example.com/opera"},
				},
			},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "opera",
				"decision":  "reject",
				"reason":    "keyword_collision",
				"rationale": "Pure 'house' collision; not a property story.",
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1 (big story must survive); payload=%#v", len(signals), payload)
	}
	decision := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if decision["decision"] != "monitor_only" || decision["reason"] != "big_story_surfaced" {
		t.Fatalf("big-story decision mismatch: %#v", decision)
	}
	if decision["guardrail"] != "big_story_recall" {
		t.Fatalf("missing big_story_recall guardrail: %#v", decision)
	}
	if decision["weakness_flag"] != "keyword_collision" {
		t.Fatalf("weakness_flag not preserved: %#v", decision)
	}
	coarse := valueOrEmptyMap(payload["coarse_relevance"])
	if coarse["guardrail_override_count"] != 1 {
		t.Fatalf("guardrail count=%v, want 1", coarse["guardrail_override_count"])
	}
}

func TestFilterApplyNeverDropsWidelyCoveredUnknownBand(t *testing.T) {
	// A widely-covered story whose surfaced variant has no publication metadata
	// comes back band=unknown. Missing metadata must not strip big-story
	// protection: coverage across >=2 independent domains still counts as big.
	candidates := map[string]any{
		"monitor": map[string]any{"profile": map[string]any{"company": "Clearnym"}},
		"signals": []any{
			map[string]any{
				"id":         "spread",
				"title":      "Major data-broker enforcement action",
				"story_size": map[string]any{"band": "unknown"},
				"evidence": []any{
					map[string]any{"url": "https://www.reuters.com/legal/data-broker-action"},
					map[string]any{"url": "https://www.bloomberg.com/news/data-broker-action"},
				},
			},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{"signal_id": "spread", "decision": "reject", "reason": "no_profile_bridge"},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected=%d, want 1 (widely-covered unknown-band must survive)", len(signals))
	}
	d := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if d["guardrail"] != "big_story_recall" || d["decision"] != "monitor_only" {
		t.Fatalf("expected big_story_recall rescue, got: %#v", d)
	}
}

func TestFilterApplyDropsUnknownBandSingleSource(t *testing.T) {
	// Control: an unknown-band signal with a single source domain and no
	// cross-source agreement is NOT a big story — coarse may drop it.
	candidates := map[string]any{
		"monitor": map[string]any{"profile": map[string]any{"company": "Clearnym"}},
		"signals": []any{
			map[string]any{
				"id":         "lone",
				"title":      "A.I., Journalism and the Public Square",
				"story_size": map[string]any{"band": "unknown"},
				"evidence":   []any{map[string]any{"url": "https://www.nytco.com/press/ai-journalism/"}},
			},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{"signal_id": "lone", "decision": "reject", "reason": "off_beat"},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true, "monitor_only": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	if got := len(signalSlice(payload["signals"])); got != 0 {
		t.Fatalf("selected=%d, want 0 (lone unknown-band off-beat is droppable)", got)
	}
}

func TestOriginApplyComputesStaleDespiteWorkerFreshStatus(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-25T18:00:00Z"},
		"signals": []any{
			map[string]any{"id": "gm", "title": "General Motors settles lawsuit over selling customer driving data"},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":              "gm",
				"status":                 "fresh",
				"first_public_at":        "2026-05-08",
				"original_url":           "https://example.com/gm-original",
				"canonical_coverage_url": "https://example.com/gm-major",
				"same_story_basis":       "Same GM settlement and same lawsuit.",
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-05-25T18:00:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	if got := len(signalSlice(payload["signals"])); got != 0 {
		t.Fatalf("selected signals=%d, want 0", got)
	}
	rejected, _ := valueOrEmptyMap(payload["freshness_gate"])["rejected_signals"].([]map[string]any)
	if len(rejected) != 1 {
		t.Fatalf("rejected=%d, want 1", len(rejected))
	}
	gate := valueOrEmptyMap(rejected[0]["freshness_gate"])
	if gate["computed_status"] != "stale" {
		t.Fatalf("computed_status=%v, want stale; gate=%#v", gate["computed_status"], gate)
	}
	if gate["worker_status"] != "fresh" {
		t.Fatalf("worker_status=%v, want fresh", gate["worker_status"])
	}
}

func TestOriginApplySelectsFreshNewDevelopment(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-25T18:00:00Z"},
		"signals": []any{
			map[string]any{"id": "case", "title": "Regulator files new complaint"},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":          "case",
				"first_public_at":    "2026-05-08",
				"new_development_at": "2026-05-25T12:30:00Z",
				"new_development":    "New settlement was filed.",
				"timestamp_evidence": []any{
					map[string]any{
						"source":       "news_search",
						"url":          "https://reuters.com/legal/regulator-new-settlement",
						"published_at": "2026-05-25T12:30:00Z",
						"note":         "Reuters coverage of new settlement filing",
					},
				},
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-05-25T18:00:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1; freshness_gate=%#v", len(signals), payload["freshness_gate"])
	}
	gate := valueOrEmptyMap(signals[0]["freshness_gate"])
	if gate["computed_status"] != "fresh_new_development" {
		t.Fatalf("computed_status=%v, want fresh_new_development; gate=%#v", gate["computed_status"], gate)
	}
}

func TestOriginApplyPromotesCutoffDateWithPreciseTimestampEvidence(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-06-01T13:46:00Z"},
		"signals": []any{
			map[string]any{
				"id":    "scmp",
				"title": "5 new cafes and coffee shops to try in Hong Kong",
				"evidence": []any{
					map[string]any{"url": "https://www.scmp.com/lifestyle/food-drink/article/3355305/story?utm_source=x"},
				},
			},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":              "scmp",
				"first_public_at":        "2026-05-31",
				"original_url":           "https://www.scmp.com/lifestyle/food-drink/article/3355305/story",
				"canonical_coverage_url": "https://www.scmp.com/lifestyle/food-drink/article/3355305/story",
				"timestamp_evidence": []any{
					map[string]any{
						"source":       "page_meta",
						"url":          "https://www.scmp.com/lifestyle/food-drink/article/3355305/story",
						"published_at": "2026-05-31T22:46:35Z",
					},
					map[string]any{
						"source":       "news_search",
						"url":          "https://news.google.com/rss/articles/scmp-story",
						"published_at": "2026-05-31T22:46:35Z",
					},
				},
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-06-01T13:46:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1; gate=%#v", len(signals), payload["freshness_gate"])
	}
	gate := valueOrEmptyMap(signals[0]["freshness_gate"])
	if gate["computed_status"] != "fresh" {
		t.Fatalf("computed_status=%v, want fresh; gate=%#v", gate["computed_status"], gate)
	}
	if gate["basis_field"] != "timestamp_evidence.published_at" {
		t.Fatalf("basis_field=%v, want timestamp evidence; gate=%#v", gate["basis_field"], gate)
	}
}

func TestOriginApplyAcceptsPrimaryAndCanonicalTimestampEvidence(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-06-01T18:00:00Z"},
		"signals": []any{
			map[string]any{
				"id":    "datagrail",
				"title": "145 AI laws passed in 2025 and privacy teams aren’t catching a break",
				"evidence": []any{
					map[string]any{"url": "https://www.helpnetsecurity.com/2026/06/01/datagrail-ai-privacy-risks-report/"},
				},
			},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":              "datagrail",
				"first_public_at":        "2026-06-01",
				"original_url":           "https://www.datagrail.io/resources/interactive/data-privacy-trends-report-2026/",
				"canonical_coverage_url": "https://www.helpnetsecurity.com/2026/06/01/datagrail-ai-privacy-risks-report/",
				"timestamp_evidence": []any{
					map[string]any{
						"source":       "primary_source",
						"url":          "https://www.datagrail.io/resources/interactive/data-privacy-trends-report-2026/",
						"published_at": "2026-06-01",
					},
					map[string]any{
						"source":       "news_search",
						"url":          "https://www.helpnetsecurity.com/2026/06/01/datagrail-ai-privacy-risks-report/",
						"published_at": "2026-06-01",
					},
				},
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-06-01T18:00:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	if got := len(signalSlice(payload["signals"])); got != 1 {
		t.Fatalf("selected signals=%d, want 1; gate=%#v", got, payload["freshness_gate"])
	}
}

func TestOriginApplyDowngradesWhenWorkerCitesOnlySurfacedURL(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-27T19:00:00Z"},
		"signals": []any{
			map[string]any{
				"id":    "advocacy",
				"title": "Group applauds governor for signing privacy bill",
				"evidence": []any{
					map[string]any{
						"source": "news_search",
						"url":    "https://advocacy.example.org/press_release/applauds-signing",
					},
				},
			},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":              "advocacy",
				"first_public_at":        "2026-05-27T18:00:00Z",
				"original_url":           "https://advocacy.example.org/press_release/applauds-signing",
				"canonical_coverage_url": "https://advocacy.example.org/press_release/applauds-signing",
				"timestamp_evidence": []any{
					map[string]any{
						"source":       "page_meta",
						"url":          "https://advocacy.example.org/press_release/applauds-signing",
						"published_at": "2026-05-27T18:00:00Z",
						"note":         "self-citing press release page",
					},
				},
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-05-27T19:00:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	if got := len(signalSlice(payload["signals"])); got != 0 {
		t.Fatalf("selected signals=%d, want 0 (worker cited only surfaced URL)", got)
	}
	gate := valueOrEmptyMap(payload["freshness_gate"])
	rejected, _ := gate["rejected_signals"].([]map[string]any)
	if len(rejected) != 1 {
		t.Fatalf("rejected=%d, want 1; gate=%#v", len(rejected), gate)
	}
	rg := valueOrEmptyMap(rejected[0]["freshness_gate"])
	if rg["computed_status"] != "unverified_no_corroboration" {
		t.Fatalf("computed_status=%v, want unverified_no_corroboration; gate=%#v", rg["computed_status"], rg)
	}
	if rg["worker_status"] != "fresh" {
		t.Fatalf("worker_status=%v, want fresh (preserved); gate=%#v", rg["worker_status"], rg)
	}
}

func mustParseTestTime(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, ok := parseTime(raw)
	if !ok {
		t.Fatalf("could not parse %s", raw)
	}
	return parsed
}
