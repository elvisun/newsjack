package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type capturedAPIRequest struct {
	Method string
	Path   string
	Query  string
	Auth   string
	Body   map[string]any
}

func runWithMockMedialyst(t *testing.T, handler func(w http.ResponseWriter, r *http.Request), run func(baseURL string) (int, string, string)) (int, string, string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)
	var code int
	var stdout, stderr string
	withTempEnv(t, map[string]string{
		"HOME":                        t.TempDir(),
		"NEWSJACK_HOME":               "",
		"NEWSJACK_MEDIALYST_API_BASE": server.URL,
		"MEDIALYST_API_BASE":          "",
		"MEDIALYST_API_KEY":           "mlst_test_key",
		"NEWSJACK_NO_AUTO_UPDATE":     "1",
		"NEWSJACK_IGNORE_DOTENV":      "1",
	}, func() {
		code, stdout, stderr = run(server.URL)
	})
	return code, stdout, stderr
}

func decodeCapturedRequest(t *testing.T, r *http.Request) capturedAPIRequest {
	t.Helper()
	var body map[string]any
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("request body is not JSON: %v", err)
		}
	}
	return capturedAPIRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Auth:   r.Header.Get("Authorization"),
		Body:   body,
	}
}

func TestNewsSearchCallsPublicRESTAPI(t *testing.T) {
	var got capturedAPIRequest
	code, stdout, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		got = decodeCapturedRequest(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"news":[{"title":"AI funding","link":"https://example.com"}]}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"news", "search", "--query", "AI funding", "--page", "2", "--tbs", "qdr:m", "--limit", "7"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("news search code=%d stderr=%s", code, stderr)
	}
	if got.Method != http.MethodPost || got.Path != "/v1/news/search" {
		t.Fatalf("request = %s %s", got.Method, got.Path)
	}
	if got.Auth != "Bearer mlst_test_key" {
		t.Fatalf("auth header = %q", got.Auth)
	}
	if got.Body["q"] != "AI funding" || got.Body["tbs"] != "qdr:m" || got.Body["num"] != float64(7) {
		t.Fatalf("unexpected body: %#v", got.Body)
	}
	if !strings.Contains(stdout, `"news"`) {
		t.Fatalf("stdout should contain API response JSON: %s", stdout)
	}
}

func TestNewsSearchBareJSONFlagExplainsOutputDefault(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"news", "search", "--query", "AI funding", "--json"}, &out, &errBuf)
	if code == 0 {
		t.Fatalf("news search --json without body should fail")
	}
	if !strings.Contains(errBuf.String(), "prints JSON by default") || !strings.Contains(errBuf.String(), "--json expects an exact JSON request body") {
		t.Fatalf("stderr should explain --json semantics, got: %s", errBuf.String())
	}
}

func TestJournalistsEnrichUsesPR1024Shape(t *testing.T) {
	var got capturedAPIRequest
	code, _, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		got = decodeCapturedRequest(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"object":"journalist_enrichment_batch","status":"complete","journalists":[],"research":[],"unresolved":[],"usage":{"estimated_credits":1,"credits_used":0,"not_enriched":0}}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"journalists", "enrich", "--url", "https://example.com/story", "--pitch", "developer AI observability", "--include-recent", "5"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("journalists enrich code=%d stderr=%s", code, stderr)
	}
	if got.Method != http.MethodPost || got.Path != "/v1/journalists/enrich" {
		t.Fatalf("request = %s %s", got.Method, got.Path)
	}
	from, ok := got.Body["from"].([]any)
	if !ok || len(from) != 1 {
		t.Fatalf("from payload = %#v", got.Body["from"])
	}
	source, _ := from[0].(map[string]any)
	if source["type"] != "article_url" || source["url"] != "https://example.com/story" {
		t.Fatalf("source payload = %#v", source)
	}
	if fit, _ := got.Body["fit_context"].(map[string]any); fit["pitch"] != "developer AI observability" {
		t.Fatalf("fit_context = %#v", got.Body["fit_context"])
	}
	options, _ := got.Body["options"].(map[string]any)
	if options["wait"] != true || options["include_recent"] != float64(5) {
		t.Fatalf("options = %#v", options)
	}
}

func TestJournalistsEnrichWaitRejectsMultipleURLs(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runCLI([]string{
		"journalists", "enrich",
		"--url", "https://example.com/one",
		"--url", "https://example.com/two",
		"--wait",
	}, &out, &errBuf)
	if code == 0 {
		t.Fatalf("journalists enrich should reject multi-url foreground waits")
	}
	if !strings.Contains(errBuf.String(), "one --url at a time") {
		t.Fatalf("stderr should explain bounded foreground waits, got: %s", errBuf.String())
	}
}

func TestJournalistsEnrichRejectsLongForegroundTimeouts(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runCLI([]string{
		"journalists", "enrich",
		"--url", "https://example.com/story",
		"--wait",
		"--poll-timeout-ms", "90000",
	}, &out, &errBuf)
	if code == 0 {
		t.Fatalf("journalists enrich should reject long foreground wait budget")
	}
	if !strings.Contains(errBuf.String(), "capped at 45000") {
		t.Fatalf("stderr should explain foreground wait cap, got: %s", errBuf.String())
	}
}

func TestJournalistsEnrichRejectsLongAPITimeout(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runCLI([]string{
		"journalists", "enrich",
		"--url", "https://example.com/story",
		"--timeout-ms", "90000",
	}, &out, &errBuf)
	if code == 0 {
		t.Fatalf("journalists enrich should reject API timeout above public cap")
	}
	if !strings.Contains(errBuf.String(), "capped by the public API at 30000") {
		t.Fatalf("stderr should explain API timeout cap, got: %s", errBuf.String())
	}
}

func TestJournalistsEnrichWaitPollsJobUntilComplete(t *testing.T) {
	var paths []string
	code, stdout, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch len(paths) {
		case 1:
			if r.Method != http.MethodPost || r.URL.Path != "/v1/journalists/enrich" {
				t.Fatalf("initial request = %s %s", r.Method, r.URL.Path)
			}
			w.Write([]byte(`{"id":"jej_123","object":"journalist_enrichment_job","status":"processing"}`))
		case 2:
			if r.Method != http.MethodGet || r.URL.Path != "/v1/journalist-enrichment-jobs/jej_123" {
				t.Fatalf("poll request = %s %s", r.Method, r.URL.Path)
			}
			w.Write([]byte(`{"id":"jej_123","object":"journalist_enrichment_job","status":"processing"}`))
		default:
			if r.Method != http.MethodGet || r.URL.Path != "/v1/journalist-enrichment-jobs/jej_123" {
				t.Fatalf("poll request = %s %s", r.Method, r.URL.Path)
			}
			w.Write([]byte(`{"id":"jej_123","object":"journalist_enrichment_job","status":"complete","result":{"journalists":[{"name":"Ada Reporter"}]}}`))
		}
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"journalists", "enrich", "--url", "https://example.com/story", "--wait", "--poll-timeout-ms", "200", "--poll-interval-ms", "1"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("journalists enrich code=%d stderr=%s", code, stderr)
	}
	if len(paths) != 3 {
		t.Fatalf("expected initial request plus two polls, got %v", paths)
	}
	if !strings.Contains(stdout, `"status": "complete"`) || !strings.Contains(stdout, "Ada Reporter") {
		t.Fatalf("stdout should contain completed job payload:\n%s", stdout)
	}
}

func TestJournalistsEnrichWaitDoesNotPollAfterForegroundBudget(t *testing.T) {
	var paths []string
	code, stdout, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if len(paths) == 1 {
			time.Sleep(30 * time.Millisecond)
			w.Write([]byte(`{"id":"jej_slow","object":"journalist_enrichment_job","status":"processing"}`))
			return
		}
		w.Write([]byte(`{"id":"jej_slow","object":"journalist_enrichment_job","status":"complete","result":{"journalists":[{"name":"Late Poll"}]}}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"journalists", "enrich", "--url", "https://example.com/story", "--wait", "--poll-timeout-ms", "20", "--poll-interval-ms", "1"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("journalists enrich code=%d stderr=%s", code, stderr)
	}
	if len(paths) != 1 {
		t.Fatalf("expected only initial request after budget was consumed, got %v", paths)
	}
	if !strings.Contains(stdout, `"status": "processing"`) || strings.Contains(stdout, "Late Poll") {
		t.Fatalf("stdout should contain the initial processing payload only:\n%s", stdout)
	}
}

func TestJournalistsEnrichNoWaitDoesNotPoll(t *testing.T) {
	var paths []string
	code, stdout, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"jej_123","object":"journalist_enrichment_job","status":"processing"}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"journalists", "enrich", "--url", "https://example.com/story", "--wait=false", "--poll-interval-ms", "1"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("journalists enrich code=%d stderr=%s", code, stderr)
	}
	if len(paths) != 1 {
		t.Fatalf("expected no polling, got %v", paths)
	}
	if !strings.Contains(stdout, `"status": "processing"`) {
		t.Fatalf("stdout should contain initial job payload:\n%s", stdout)
	}
}

func TestJournalistsEnrichNoWaitAllowsCandidateBatch(t *testing.T) {
	var got capturedAPIRequest
	code, _, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		got = decodeCapturedRequest(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"jej_batch","object":"journalist_enrichment_job","status":"processing"}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{
			"journalists", "enrich",
			"--url", "https://example.com/one",
			"--url", "https://example.com/two",
			"--pitch", "screen regional fintech candidates",
			"--wait=false",
		}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("journalists enrich batch code=%d stderr=%s", code, stderr)
	}
	if got.Method != http.MethodPost || got.Path != "/v1/journalists/enrich" {
		t.Fatalf("request = %s %s", got.Method, got.Path)
	}
	sources, ok := got.Body["from"].([]any)
	if !ok || len(sources) != 2 {
		t.Fatalf("from payload = %#v", got.Body["from"])
	}
	options, _ := got.Body["options"].(map[string]any)
	if options["wait"] != false {
		t.Fatalf("options = %#v", options)
	}
}

func TestCreditsDefaultsToBalance(t *testing.T) {
	var got capturedAPIRequest
	code, stdout, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, r *http.Request) {
		got = capturedAPIRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Auth:   r.Header.Get("Authorization"),
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"credits":{"org_available_balance":42}}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"credits"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 0 {
		t.Fatalf("credits code=%d stderr=%s", code, stderr)
	}
	if got.Method != http.MethodGet || got.Path != "/v1/credits/balance" {
		t.Fatalf("request = %s %s", got.Method, got.Path)
	}
	if !strings.Contains(stdout, "org_available_balance") {
		t.Fatalf("stdout should contain balance payload: %s", stdout)
	}
}

func TestMedialystAPIErrorMessageIncludesRequestID(t *testing.T) {
	code, _, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"FORBIDDEN","code":"FORBIDDEN","message":"Insufficient scope","request_id":"req_123"}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"credits", "balance"}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 1 {
		t.Fatalf("forbidden request code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{"Insufficient scope", "req_123", "HTTP 403"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q:\n%s", want, stderr)
		}
	}
}

func TestMedialystAPIErrorMessageUsesNestedIssue(t *testing.T) {
	code, _, stderr := runWithMockMedialyst(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":{"issues":[{"message":"Invalid discriminator value"}],"name":"ZodError"}}`))
	}, func(_ string) (int, string, string) {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"news", "search", "--json", `{"q":123}`}, &out, &errBuf)
		return code, out.String(), errBuf.String()
	})
	if code != 1 {
		t.Fatalf("bad request code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "HTTP 400 HTTP_400: Invalid discriminator value") {
		t.Fatalf("stderr should include nested issue message without Go map formatting:\n%s", stderr)
	}
}

func TestMediaListsCommandIsNotDispatched(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"media-lists", "list"}, &out, &errBuf)
	if code == 0 {
		t.Fatalf("media-lists command should not be dispatched")
	}
	if !strings.Contains(errBuf.String(), "unknown command: media-lists") {
		t.Fatalf("stderr should explain command removal:\n%s", errBuf.String())
	}
}

func TestMedialystAPIMissingKeyPrintsLoginHint(t *testing.T) {
	withTempEnv(t, map[string]string{
		"HOME":                    t.TempDir(),
		"NEWSJACK_HOME":           "",
		"MEDIALYST_API_KEY":       "",
		"NEWSJACK_IGNORE_DOTENV":  "1",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"news", "search", "--query", "AI"}, &out, &errBuf)
		if code != 1 {
			t.Fatalf("missing key code=%d stdout=%s stderr=%s", code, out.String(), errBuf.String())
		}
		if !strings.Contains(errBuf.String(), "newsjack login") {
			t.Fatalf("stderr should include login hint:\n%s", errBuf.String())
		}
	})
}
