package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const medialystAPIDefaultBase = "https://medialyst.ai/api"
const maxJournalistEnrichPollTimeout = 45 * time.Second

type repeatedString []string

func (v *repeatedString) String() string {
	return strings.Join(*v, ",")
}

func (v *repeatedString) Set(value string) error {
	value = strings.TrimSpace(value)
	if value != "" {
		*v = append(*v, value)
	}
	return nil
}

func restHelpRequested(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return true
	}
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

type medialystAPIResponse struct {
	StatusCode int
	Headers    http.Header
	Payload    any
	Body       string
}

type medialystAPIError struct {
	StatusCode int
	Payload    any
	Body       string
}

func (e *medialystAPIError) Error() string {
	message := strings.TrimSpace(e.Body)
	code := fmt.Sprintf("HTTP_%d", e.StatusCode)
	requestID := ""
	if record, ok := e.Payload.(map[string]any); ok {
		code = firstAPIString(record["code"], apiNestedValue(record["error"], "code"), apiNestedValue(record["error"], "error"), record["error"], code)
		message = firstAPIString(record["message"], apiNestedValue(record["error"], "message"), firstAPIIssueMessage(record["issues"]), firstAPIIssueMessage(apiNestedValue(record["error"], "issues")), message)
		requestID = firstAPIString(record["request_id"], apiNestedValue(record["error"], "request_id"))
	}
	if message == "" {
		message = http.StatusText(e.StatusCode)
	}
	if requestID != "" {
		return fmt.Sprintf("HTTP %d %s: %s (request_id: %s)", e.StatusCode, code, message, requestID)
	}
	return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, code, message)
}

func firstAPIString(values ...any) string {
	for _, v := range values {
		if s, ok := v.(string); ok {
			if strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func apiNestedValue(v any, key string) any {
	if record, ok := v.(map[string]any); ok {
		return record[key]
	}
	return nil
}

func firstAPIIssueMessage(v any) string {
	issues, ok := v.([]any)
	if !ok || len(issues) == 0 {
		return ""
	}
	if issue, ok := issues[0].(map[string]any); ok {
		return firstAPIString(issue["message"])
	}
	return ""
}

func medialystAPIBase() string {
	base := strings.TrimSpace(os.Getenv("NEWSJACK_MEDIALYST_API_BASE"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("MEDIALYST_API_BASE"))
	}
	if base == "" {
		base = medialystAPIDefaultBase
	}
	return strings.TrimRight(base, "/")
}

func medialystAPIRequest(method, path string, query url.Values, body any, headers map[string]string, timeout time.Duration) (*medialystAPIResponse, error) {
	key, source := loadAPIKey()
	if key == "" {
		return nil, errors.New("Medialyst API key not found. Run: " + newsjackCLIInvocation().Display("login"))
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	rawURL := medialystAPIBase() + path
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		if strings.TrimSpace(v) != "" {
			req.Header.Set(k, v)
		}
	}
	if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "Loaded Medialyst API key from %s\n", source)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	bodyText := string(data)
	var payload any
	if len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &payload); err != nil {
			if resp.StatusCode >= 400 {
				return nil, &medialystAPIError{StatusCode: resp.StatusCode, Body: bodyText}
			}
			return nil, fmt.Errorf("Medialyst returned non-JSON response: %s", truncate(bodyText, 300))
		}
	}
	apiResp := &medialystAPIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Payload:    payload,
		Body:       bodyText,
	}
	if resp.StatusCode >= 400 {
		return nil, &medialystAPIError{StatusCode: resp.StatusCode, Payload: payload, Body: bodyText}
	}
	return apiResp, nil
}

func runMedialystJSON(stdout, stderr io.Writer, method, path string, query url.Values, body any, headers map[string]string, timeout time.Duration) int {
	resp, err := medialystAPIRequest(method, path, query, body, headers, timeout)
	if err != nil {
		return failf(stderr, "Medialyst API request failed: %v", err)
	}
	writeMedialystPayload(stdout, resp.Payload)
	return 0
}

func writeMedialystPayload(stdout io.Writer, payload any) {
	if payload == nil {
		fmt.Fprintln(stdout, "{}")
		return
	}
	writeJSON(stdout, payload)
}

func durationFromMillis(value int, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return time.Duration(value) * time.Millisecond
}

func durationMillisFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return 0
	}
}

func journalistEnrichRequestTimeout(body any) time.Duration {
	timeout := 30 * time.Second
	if record, ok := body.(map[string]any); ok {
		if options, ok := record["options"].(map[string]any); ok {
			if value := durationMillisFromAny(options["timeout_ms"]); value > 0 {
				timeout = time.Duration(value) * time.Millisecond
			}
		}
	}
	return timeout + 15*time.Second
}

func apiPayloadString(payload any, key string) string {
	if record, ok := payload.(map[string]any); ok {
		return firstAPIString(record[key])
	}
	return ""
}

func enrichmentJobTerminal(payload any) bool {
	status := strings.ToLower(apiPayloadString(payload, "status"))
	switch status {
	case "", "queued", "processing", "running", "pending":
		return false
	default:
		return true
	}
}

func waitForJournalistEnrichmentJob(jobID string, initial any, timeout, interval time.Duration) (any, error) {
	if strings.TrimSpace(jobID) == "" || timeout <= 0 || enrichmentJobTerminal(initial) {
		return initial, nil
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	deadline := time.Now().Add(timeout)
	last := initial
	path := "/v1/journalist-enrichment-jobs/" + url.PathEscape(jobID)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return last, nil
		}
		if remaining < interval {
			time.Sleep(remaining)
		} else {
			time.Sleep(interval)
		}
		remaining = time.Until(deadline)
		if remaining <= 0 {
			return last, nil
		}
		requestTimeout := remaining
		if requestTimeout > 45*time.Second {
			requestTimeout = 45 * time.Second
		}
		resp, err := medialystAPIRequest(http.MethodGet, path, nil, nil, nil, requestTimeout)
		if err != nil {
			return nil, err
		}
		last = resp.Payload
		if enrichmentJobTerminal(last) {
			return last, nil
		}
	}
}

func runJournalistEnrich(stdout, stderr io.Writer, body any, headers map[string]string, shouldWait bool, pollTimeout, pollInterval time.Duration) int {
	started := time.Now()
	resp, err := medialystAPIRequest(http.MethodPost, "/v1/journalists/enrich", nil, body, headers, journalistEnrichRequestTimeout(body))
	if err != nil {
		return failf(stderr, "Medialyst API request failed: %v", err)
	}
	payload := resp.Payload
	if shouldWait {
		if jobID := apiPayloadString(payload, "id"); jobID != "" && !enrichmentJobTerminal(payload) {
			remaining := pollTimeout - time.Since(started)
			if remaining > 0 {
				payload, err = waitForJournalistEnrichmentJob(jobID, payload, remaining, pollInterval)
			}
			if err != nil {
				return failf(stderr, "Medialyst API request failed: %v", err)
			}
		}
	}
	writeMedialystPayload(stdout, payload)
	return 0
}

func runJournalistEnrichJob(stdout, stderr io.Writer, jobID string, shouldWait bool, pollTimeout, pollInterval time.Duration) int {
	path := "/v1/journalist-enrichment-jobs/" + url.PathEscape(jobID)
	resp, err := medialystAPIRequest(http.MethodGet, path, nil, nil, nil, 45*time.Second)
	if err != nil {
		return failf(stderr, "Medialyst API request failed: %v", err)
	}
	payload := resp.Payload
	if shouldWait {
		payload, err = waitForJournalistEnrichmentJob(jobID, payload, pollTimeout, pollInterval)
		if err != nil {
			return failf(stderr, "Medialyst API request failed: %v", err)
		}
	}
	writeMedialystPayload(stdout, payload)
	return 0
}

func parseJSONFlag(inline, file, label string) (any, bool, error) {
	if strings.TrimSpace(inline) != "" && strings.TrimSpace(file) != "" {
		return nil, false, fmt.Errorf("use only one of --json or --json-file for %s", label)
	}
	var data []byte
	var err error
	switch {
	case strings.TrimSpace(inline) != "":
		data = []byte(inline)
	case strings.TrimSpace(file) != "":
		if file == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(expandPath(file))
		}
		if err != nil {
			return nil, false, err
		}
	default:
		return nil, false, nil
	}
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, false, fmt.Errorf("invalid JSON for %s: %w", label, err)
	}
	return payload, true, nil
}

func bareJSONFlag(args []string) bool {
	for i, arg := range args {
		if arg == "--json" || arg == "-json" {
			return i == len(args)-1 || strings.HasPrefix(args[i+1], "-")
		}
		if arg == "--json=" || arg == "-json=" {
			return true
		}
	}
	return false
}

func bareJSONFlagError(label string) error {
	return fmt.Errorf("%s prints JSON by default; --json expects an exact JSON request body, for example --json '{\"q\":\"AI observability\"}'. Omit --json for normal output", label)
}

func addQueryIfSet(values url.Values, key, value string) {
	if strings.TrimSpace(value) != "" {
		values.Set(key, strings.TrimSpace(value))
	}
}

func addQueryIntIfPositive(values url.Values, key string, value int) {
	if value > 0 {
		values.Set(key, strconv.Itoa(value))
	}
}

func addQueryBoolIfTrue(values url.Values, key string, value bool) {
	if value {
		values.Set(key, "true")
	}
}

func cmdCredits(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printCreditsHelp(stdout)
		return 0
	}
	if len(args) == 0 {
		return runMedialystJSON(stdout, stderr, http.MethodGet, "/v1/credits/balance", nil, nil, nil, 30*time.Second)
	}
	switch args[0] {
	case "balance":
		fs := flag.NewFlagSet("credits balance", flag.ContinueOnError)
		fs.SetOutput(stderr)
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		return runMedialystJSON(stdout, stderr, http.MethodGet, "/v1/credits/balance", nil, nil, nil, 30*time.Second)
	default:
		return failf(stderr, "unknown credits command: %s", args[0])
	}
}

func cmdNews(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printNewsHelp(stdout)
		return 0
	}
	switch args[0] {
	case "search":
		return cmdNewsSearch(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown news command: %s", args[0])
	}
}

func cmdNewsSearch(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("news search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	query := fs.String("query", "", "News search query")
	q := fs.String("q", "", "Alias for --query")
	gl := fs.String("gl", "us", "Country code")
	hl := fs.String("hl", "en", "Interface language")
	page := fs.Int("page", 1, "1-based result page")
	limit := fs.Int("limit", 0, "Result limit, sent to the API as num")
	num := fs.Int("num", 0, "Exact API result limit")
	tbs := fs.String("tbs", "", "Serper time filter, such as qdr:m")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if bareJSONFlag(args) {
		return fail(stderr, bareJSONFlagError("news search"))
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "news search"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/news/search", nil, body, nil, 45*time.Second)
	}
	searchQuery := strings.TrimSpace(firstString(*query, *q))
	if searchQuery == "" {
		return fail(stderr, errors.New("usage: newsjack news search --query <query>"))
	}
	body := map[string]any{
		"q":  searchQuery,
		"gl": strings.TrimSpace(*gl),
		"hl": strings.TrimSpace(*hl),
	}
	if *page > 0 {
		body["page"] = *page
	}
	if *limit > 0 {
		body["num"] = *limit
	}
	if *num > 0 {
		body["num"] = *num
	}
	if strings.TrimSpace(*tbs) != "" {
		body["tbs"] = strings.TrimSpace(*tbs)
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/news/search", nil, body, nil, 45*time.Second)
}

func cmdJournalists(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printJournalistsHelp(stdout)
		return 0
	}
	switch args[0] {
	case "enrich":
		return cmdJournalistsEnrich(args[1:], stdout, stderr)
	case "enrich-job", "job":
		return cmdJournalistsEnrichJob(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown journalists command: %s", args[0])
	}
}

func cmdJournalistsEnrich(args []string, stdout, stderr io.Writer) int {
	if restHelpRequested(args) {
		printJournalistsHelp(stdout)
		return 0
	}
	fs := flag.NewFlagSet("journalists enrich", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var urls repeatedString
	fs.Var(&urls, "url", "Source article URL; repeat for multiple URLs")
	pitch := fs.String("pitch", "", "Optional pitch context for fit scoring")
	includeRecent := fs.Int("include-recent", 10, "Recent article count: 0 or 3-20")
	wait := fs.Bool("wait", true, "Wait briefly for a completed result")
	timeoutMS := fs.Int("timeout-ms", 30000, "Wait timeout in milliseconds, max 30000")
	pollTimeoutMS := fs.Int("poll-timeout-ms", 45000, "Total foreground wait budget when --wait is true")
	pollIntervalMS := fs.Int("poll-interval-ms", 3000, "CLI job polling interval while waiting")
	externalID := fs.String("external-id", "", "Optional external idempotency key")
	idempotencyKey := fs.String("idempotency-key", "", "Idempotency-Key header")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	headers := map[string]string{}
	if strings.TrimSpace(*idempotencyKey) != "" {
		headers["Idempotency-Key"] = strings.TrimSpace(*idempotencyKey)
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "journalists enrich"); err != nil {
		return fail(stderr, err)
	} else if ok {
		pollTimeout := durationFromMillis(*pollTimeoutMS, maxJournalistEnrichPollTimeout)
		if pollTimeout > maxJournalistEnrichPollTimeout {
			return fail(stderr, fmt.Errorf("journalists enrich --poll-timeout-ms is capped at 45000 for foreground CLI runs; got %d", *pollTimeoutMS))
		}
		return runJournalistEnrich(stdout, stderr, body, headers, *wait, pollTimeout, durationFromMillis(*pollIntervalMS, 3*time.Second))
	}
	if len(urls) == 0 {
		return fail(stderr, errors.New("usage: newsjack journalists enrich --url <article-url> [--pitch <pitch>]"))
	}
	if *wait && len(urls) > 1 {
		return fail(stderr, errors.New("journalists enrich --wait accepts one --url at a time to keep foreground runs bounded; repeat the command for selected anchor articles, or use --wait=false for a batch job"))
	}
	if *timeoutMS > 30000 {
		return fail(stderr, fmt.Errorf("journalists enrich --timeout-ms is capped by the public API at 30000; got %d", *timeoutMS))
	}
	pollTimeout := durationFromMillis(*pollTimeoutMS, maxJournalistEnrichPollTimeout)
	if pollTimeout > maxJournalistEnrichPollTimeout {
		return fail(stderr, fmt.Errorf("journalists enrich --poll-timeout-ms is capped at 45000 for foreground CLI runs; got %d", *pollTimeoutMS))
	}
	sources := make([]any, 0, len(urls))
	for _, rawURL := range urls {
		sources = append(sources, map[string]any{"type": "article_url", "url": rawURL})
	}
	options := map[string]any{
		"include_recent": *includeRecent,
		"wait":           *wait,
		"timeout_ms":     *timeoutMS,
	}
	if strings.TrimSpace(*externalID) != "" {
		options["external_id"] = strings.TrimSpace(*externalID)
	}
	body := map[string]any{
		"from":    sources,
		"options": options,
	}
	if strings.TrimSpace(*pitch) != "" {
		body["fit_context"] = map[string]any{"pitch": strings.TrimSpace(*pitch)}
	}
	return runJournalistEnrich(stdout, stderr, body, headers, *wait, pollTimeout, durationFromMillis(*pollIntervalMS, 3*time.Second))
}

func cmdJournalistsEnrichJob(args []string, stdout, stderr io.Writer) int {
	if restHelpRequested(args) {
		printJournalistsHelp(stdout)
		return 0
	}
	args = reorderIntermixedFlags(args, map[string]bool{"timeout-ms": true, "poll-timeout-ms": true, "poll-interval-ms": true})
	fs := flag.NewFlagSet("journalists enrich-job", flag.ContinueOnError)
	fs.SetOutput(stderr)
	wait := fs.Bool("wait", false, "Wait briefly for progress")
	timeoutMS := fs.Int("timeout-ms", 45000, "CLI job polling timeout in milliseconds")
	pollTimeoutMS := fs.Int("poll-timeout-ms", 0, "Alias for --timeout-ms")
	pollIntervalMS := fs.Int("poll-interval-ms", 3000, "CLI job polling interval while waiting")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack journalists enrich-job <job-id> [--wait]"))
	}
	if *pollTimeoutMS > 0 {
		*timeoutMS = *pollTimeoutMS
	}
	return runJournalistEnrichJob(stdout, stderr, fs.Arg(0), *wait, durationFromMillis(*timeoutMS, 45*time.Second), durationFromMillis(*pollIntervalMS, 3*time.Second))
}

func cmdMediaLists(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printMediaListsHelp(stdout)
		return 0
	}
	if len(args) > 1 && restHelpRequested(args[1:]) {
		printMediaListsHelp(stdout)
		return 0
	}
	switch args[0] {
	case "create":
		return cmdMediaListsCreate(args[1:], stdout, stderr)
	case "create-async":
		return cmdMediaListsCreateAsync(args[1:], stdout, stderr)
	case "job":
		return cmdMediaListsJob(args[1:], stdout, stderr)
	case "list":
		return cmdMediaListsList(args[1:], stdout, stderr)
	case "get":
		return cmdMediaListsGet(args[1:], stdout, stderr)
	case "inspect":
		return cmdMediaListsInspect(args[1:], stdout, stderr)
	case "full-values", "read-full-values":
		return cmdMediaListsFullValues(args[1:], stdout, stderr)
	case "preview-column-render":
		return cmdMediaListsPreviewColumnRender(args[1:], stdout, stderr)
	case "action", "apply-action":
		return cmdMediaListsAction(args[1:], stdout, stderr)
	case "add-urls", "add-url":
		return cmdMediaListsAddURLs(args[1:], stdout, stderr)
	case "add-keywords", "add-keyword":
		return cmdMediaListsAddKeywords(args[1:], stdout, stderr)
	case "share":
		return cmdMediaListsShare(args[1:], stdout, stderr)
	case "delete", "rm":
		return cmdMediaListsDelete(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown media-lists command: %s", args[0])
	}
}

func cmdMediaListsCreate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("media-lists create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	name := fs.String("name", "", "Media list name")
	description := fs.String("description", "", "Optional description")
	var urls repeatedString
	var keywords repeatedString
	fs.Var(&urls, "url", "Article URL source; repeat for multiple URLs")
	fs.Var(&keywords, "keyword", "Keyword source; repeat for multiple keywords")
	sourcePrompt := fs.String("source-prompt", "", "Prompt source for server-side planning")
	empty := fs.Bool("empty", false, "Create an empty list")
	limit := fs.Int("limit", 10, "Keyword or prompt source article limit")
	country := fs.String("country", "us", "Two-letter country for keyword or prompt source")
	dateRange := fs.String("date-range", "anytime", "Search date range: anytime,h,d,w,m,y")
	templateID := fs.String("template-id", "", "Optional saved template id")
	bare := fs.Bool("bare", false, "Pass template_id:null for a bare list")
	runInitial := fs.Bool("run-initial-enrichment", false, "Run workflow columns after creation")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists create"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists", nil, body, nil, 2*time.Minute)
	}
	if strings.TrimSpace(*name) == "" {
		return fail(stderr, errors.New("usage: newsjack media-lists create --name <name> (--url URL...|--keyword KW...|--source-prompt TEXT|--empty)"))
	}
	sourceCount := 0
	if len(urls) > 0 {
		sourceCount++
	}
	if len(keywords) > 0 {
		sourceCount++
	}
	if strings.TrimSpace(*sourcePrompt) != "" {
		sourceCount++
	}
	if *empty {
		sourceCount++
	}
	if sourceCount != 1 {
		return fail(stderr, errors.New("choose exactly one media list source: --url, --keyword, --source-prompt, or --empty"))
	}
	body := map[string]any{
		"name": strings.TrimSpace(*name),
	}
	if strings.TrimSpace(*description) != "" {
		body["description"] = strings.TrimSpace(*description)
	}
	if strings.TrimSpace(*templateID) != "" {
		body["template_id"] = strings.TrimSpace(*templateID)
	} else if *bare {
		body["template_id"] = nil
	}
	if *runInitial {
		body["run_initial_enrichment"] = true
	}
	switch {
	case len(urls) > 0:
		body["source"] = map[string]any{"type": "urls", "urls": []string(urls)}
	case len(keywords) > 0:
		body["source"] = map[string]any{
			"type":       "keywords",
			"keywords":   []string(keywords),
			"limit":      *limit,
			"country":    strings.TrimSpace(*country),
			"date_range": strings.TrimSpace(*dateRange),
		}
	case strings.TrimSpace(*sourcePrompt) != "":
		body["source"] = map[string]any{
			"type":       "prompt",
			"prompt":     strings.TrimSpace(*sourcePrompt),
			"limit":      *limit,
			"country":    strings.TrimSpace(*country),
			"date_range": strings.TrimSpace(*dateRange),
		}
	default:
		body["source"] = map[string]any{"type": "empty"}
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists", nil, body, nil, 2*time.Minute)
}

func cmdMediaListsCreateAsync(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("media-lists create-async", flag.ContinueOnError)
	fs.SetOutput(stderr)
	prompt := fs.String("prompt", "", "Media-list planning prompt")
	maxArticles := fs.Int("max-articles", 25, "Maximum articles")
	idempotencyKey := fs.String("idempotency-key", "", "Idempotency-Key header")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	headers := map[string]string{}
	if strings.TrimSpace(*idempotencyKey) != "" {
		headers["Idempotency-Key"] = strings.TrimSpace(*idempotencyKey)
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists create-async"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists:create-async", nil, body, headers, 45*time.Second)
	}
	if strings.TrimSpace(*prompt) == "" {
		return fail(stderr, errors.New("usage: newsjack media-lists create-async --prompt <prompt> [--max-articles N]"))
	}
	body := map[string]any{"prompt": strings.TrimSpace(*prompt), "max_articles": *maxArticles}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists:create-async", nil, body, headers, 45*time.Second)
}

func cmdMediaListsJob(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"limit": true, "cursor": true})
	fs := flag.NewFlagSet("media-lists job", flag.ContinueOnError)
	fs.SetOutput(stderr)
	includeResults := fs.Bool("include-results", false, "Include rows when available")
	limit := fs.Int("limit", 25, "Rows limit")
	cursor := fs.String("cursor", "", "Pagination cursor")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists job <job-id> [--include-results]"))
	}
	query := url.Values{}
	if *includeResults {
		query.Set("include", "results")
		addQueryIntIfPositive(query, "limit", *limit)
		addQueryIfSet(query, "cursor", *cursor)
	}
	path := "/v1/jobs/" + url.PathEscape(fs.Arg(0))
	return runMedialystJSON(stdout, stderr, http.MethodGet, path, query, nil, nil, 45*time.Second)
}

func cmdMediaListsList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("media-lists list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	limit := fs.Int("limit", 24, "Results limit")
	cursor := fs.Int("cursor", 0, "Pagination cursor")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	query := url.Values{}
	addQueryIntIfPositive(query, "limit", *limit)
	if *cursor > 0 {
		query.Set("cursor", strconv.Itoa(*cursor))
	}
	return runMedialystJSON(stdout, stderr, http.MethodGet, "/v1/media-lists", query, nil, nil, 45*time.Second)
}

func cmdMediaListsGet(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"row-detail": true, "limit": true, "cursor": true})
	fs := flag.NewFlagSet("media-lists get", flag.ContinueOnError)
	fs.SetOutput(stderr)
	includeRows := fs.Bool("include-rows", false, "Include rows")
	includeSchema := fs.Bool("include-schema", false, "Include schema")
	rowDetail := fs.String("row-detail", "preview", "Row detail: preview or full")
	limit := fs.Int("limit", 25, "Rows limit")
	cursor := fs.Int("cursor", 0, "Pagination cursor")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists get <media-list-id> [--include-rows]"))
	}
	query := url.Values{}
	addQueryBoolIfTrue(query, "include_rows", *includeRows)
	addQueryBoolIfTrue(query, "include_schema", *includeSchema)
	addQueryIfSet(query, "row_detail", *rowDetail)
	addQueryIntIfPositive(query, "limit", *limit)
	if *cursor > 0 {
		query.Set("cursor", strconv.Itoa(*cursor))
	}
	path := "/v1/media-lists/" + url.PathEscape(fs.Arg(0))
	return runMedialystJSON(stdout, stderr, http.MethodGet, path, query, nil, nil, 45*time.Second)
}

func cmdMediaListsInspect(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{
		"view": true, "mode": true, "view-id": true, "limit": true, "offset": true,
		"cell-detail": true, "include-rows": true, "row-limit": true, "json": true, "json-file": true,
	})
	fs := flag.NewFlagSet("media-lists inspect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	view := fs.String("view", "rows", "View: rows or enrichment_health")
	mode := fs.String("mode", "row_window_preview", "Rows preview mode")
	viewID := fs.String("view-id", "", "Optional saved view id")
	limit := fs.Int("limit", 10, "Rows limit")
	offset := fs.Int("offset", 0, "Rows offset")
	cellDetail := fs.String("cell-detail", "compact", "Cell detail: compact or full")
	includeRows := fs.String("include-rows", "blocked", "Health rows: none,blocked,failed,active,all")
	rowLimit := fs.Int("row-limit", 20, "Health row limit")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists inspect <media-list-id>"))
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists inspect"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/inspect", nil, body, nil, 45*time.Second)
	}
	body := map[string]any{"view": strings.TrimSpace(*view)}
	if body["view"] == "enrichment_health" {
		body["includeRows"] = strings.TrimSpace(*includeRows)
		body["rowLimit"] = *rowLimit
	} else {
		body["mode"] = strings.TrimSpace(*mode)
		body["limit"] = *limit
		body["offset"] = *offset
		body["cellDetail"] = strings.TrimSpace(*cellDetail)
		if strings.TrimSpace(*viewID) != "" {
			body["viewId"] = strings.TrimSpace(*viewID)
		}
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/inspect", nil, body, nil, 45*time.Second)
}

func cmdMediaListsFullValues(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"row-id": true, "column-id": true, "json": true, "json-file": true})
	fs := flag.NewFlagSet("media-lists full-values", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var rowIDs repeatedString
	var columnIDs repeatedString
	fs.Var(&rowIDs, "row-id", "Row id; repeat for multiple rows")
	fs.Var(&columnIDs, "column-id", "Column id; repeat for multiple columns")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists full-values <media-list-id> --row-id ROW --column-id COL"))
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists full-values"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/full-values", nil, body, nil, 45*time.Second)
	}
	if len(rowIDs) == 0 || len(columnIDs) == 0 {
		return fail(stderr, errors.New("full-values requires at least one --row-id and one --column-id"))
	}
	body := map[string]any{"rowIds": []string(rowIDs), "columnIds": []string(columnIDs)}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/full-values", nil, body, nil, 45*time.Second)
}

func cmdMediaListsPreviewColumnRender(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"row-id": true, "column-id": true, "json": true, "json-file": true})
	fs := flag.NewFlagSet("media-lists preview-column-render", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var rowIDs repeatedString
	fs.Var(&rowIDs, "row-id", "Row id; repeat for up to three rows")
	columnID := fs.String("column-id", "", "Existing column id")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists preview-column-render <media-list-id> --row-id ROW --column-id COL"))
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists preview-column-render"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/column-render-preview", nil, body, nil, 45*time.Second)
	}
	if len(rowIDs) == 0 || strings.TrimSpace(*columnID) == "" {
		return fail(stderr, errors.New("preview-column-render requires --row-id and --column-id, or use --json"))
	}
	body := map[string]any{"rowIds": []string(rowIDs), "columnId": strings.TrimSpace(*columnID)}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/column-render-preview", nil, body, nil, 45*time.Second)
}

func cmdMediaListsAction(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"json": true, "json-file": true})
	fs := flag.NewFlagSet("media-lists action", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonInline := fs.String("json", "", "Exact MediaListAction JSON body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists action <media-list-id> --json '{\"action\":\"...\"}'"))
	}
	body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists action")
	if err != nil {
		return fail(stderr, err)
	}
	if !ok {
		return fail(stderr, errors.New("media-lists action requires --json or --json-file"))
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/actions", nil, body, nil, 45*time.Second)
}

func cmdMediaListsAddURLs(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"url": true, "json": true, "json-file": true})
	fs := flag.NewFlagSet("media-lists add-urls", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var urls repeatedString
	fs.Var(&urls, "url", "Article URL; repeat for multiple URLs")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists add-urls <media-list-id> --url URL [--url URL...]"))
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists add-urls"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/actions", nil, body, nil, 45*time.Second)
	}
	if len(urls) == 0 {
		return fail(stderr, errors.New("media-lists add-urls requires at least one --url"))
	}
	body := map[string]any{"action": "add_articles_by_urls", "urls": []string(urls)}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/actions", nil, body, nil, 45*time.Second)
}

func cmdMediaListsAddKeywords(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{
		"keyword": true, "limit": true, "country": true, "date-range": true, "json": true, "json-file": true,
	})
	fs := flag.NewFlagSet("media-lists add-keywords", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var keywords repeatedString
	fs.Var(&keywords, "keyword", "Search keyword; repeat for multiple keyword themes")
	limit := fs.Int("limit", 10, "Article limit per keyword")
	country := fs.String("country", "us", "Two-letter country")
	dateRange := fs.String("date-range", "m", "Search date range: h,d,w,m,y,anytime")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists add-keywords <media-list-id> --keyword KW [--keyword KW...]"))
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists add-keywords"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/actions", nil, body, nil, 45*time.Second)
	}
	if len(keywords) == 0 {
		return fail(stderr, errors.New("media-lists add-keywords requires at least one --keyword"))
	}
	body := map[string]any{
		"action":    "add_articles_by_keywords",
		"keywords":  []string(keywords),
		"limit":     *limit,
		"country":   strings.TrimSpace(*country),
		"dateRange": strings.TrimSpace(*dateRange),
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/actions", nil, body, nil, 45*time.Second)
}

func cmdMediaListsShare(args []string, stdout, stderr io.Writer) int {
	args = reorderIntermixedFlags(args, map[string]bool{"label": true, "expires-at": true, "view-id": true, "json": true, "json-file": true})
	fs := flag.NewFlagSet("media-lists share", flag.ContinueOnError)
	fs.SetOutput(stderr)
	label := fs.String("label", "", "Optional share label")
	expiresAt := fs.String("expires-at", "", "Optional ISO datetime")
	viewID := fs.String("view-id", "", "Optional view id")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists share <media-list-id> [--view-id VIEW]"))
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "media-lists share"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/shares", nil, body, nil, 45*time.Second)
	}
	body := map[string]any{}
	if strings.TrimSpace(*label) != "" {
		body["label"] = strings.TrimSpace(*label)
	}
	if strings.TrimSpace(*expiresAt) != "" {
		body["expires_at"] = strings.TrimSpace(*expiresAt)
	}
	if strings.TrimSpace(*viewID) != "" {
		body["view_id"] = strings.TrimSpace(*viewID)
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/media-lists/"+url.PathEscape(fs.Arg(0))+"/shares", nil, body, nil, 45*time.Second)
}

func cmdMediaListsDelete(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("media-lists delete", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack media-lists delete <media-list-id>"))
	}
	return runMedialystJSON(stdout, stderr, http.MethodDelete, "/v1/media-lists/"+url.PathEscape(fs.Arg(0)), nil, nil, nil, 45*time.Second)
}
