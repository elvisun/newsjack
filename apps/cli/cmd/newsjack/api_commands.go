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
	cred, err := loadMedialystBearerCredential()
	if err != nil {
		return nil, err
	}
	resp, err := medialystAPIRequestWithBearer(cred, method, path, query, body, headers, timeout)
	if apiErr, ok := err.(*medialystAPIError); ok && apiErr.StatusCode == http.StatusUnauthorized && cred.Kind == "oauth" {
		refreshed, refreshErr := refreshStoredMedialystOAuth()
		if refreshErr == nil {
			return medialystAPIRequestWithBearer(refreshed, method, path, query, body, headers, timeout)
		}
		return nil, fmt.Errorf("%w; OAuth refresh also failed: %v. Run: newsjack login", err, refreshErr)
	}
	return resp, err
}

func medialystAPIRequestWithBearer(cred medialystBearerCredential, method, path string, query url.Values, body any, headers map[string]string, timeout time.Duration) (*medialystAPIResponse, error) {
	reader, err := cloneRequestBody(body)
	if err != nil {
		return nil, err
	}

	rawURL := medialystAPIBase() + path
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cred.Token)
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
		fmt.Fprintf(os.Stderr, "Loaded Medialyst %s credentials from %s\n", cred.Kind, cred.Source)
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

func cmdPRCalendar(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printPRCalendarHelp(stdout)
		return 0
	}
	switch args[0] {
	case "query", "search":
		return cmdPRCalendarQuery(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown pr-calendar command: %s", args[0])
	}
}

func cmdPRCalendarQuery(args []string, stdout, stderr io.Writer) int {
	if restHelpRequested(args) {
		printPRCalendarHelp(stdout)
		return 0
	}
	fs := flag.NewFlagSet("pr-calendar query", flag.ContinueOnError)
	fs.SetOutput(stderr)
	query := fs.String("query", "", "Calendar query")
	q := fs.String("q", "", "Alias for --query")
	from := fs.String("from", "", "Start date YYYY-MM-DD, sent as start_date")
	startDate := fs.String("start-date", "", "Alias for --from")
	to := fs.String("to", "", "End date YYYY-MM-DD, sent as end_date")
	endDate := fs.String("end-date", "", "Alias for --to")
	pitchStartBefore := fs.String("pitch-start-before", "", "Only include moments whose pitch window opens before YYYY-MM-DD")
	pitchDeadlineAfter := fs.String("pitch-deadline-after", "", "Only include moments whose pitch deadline is after YYYY-MM-DD")
	limit := fs.Int("limit", 25, "Result limit, max 100")
	cursor := fs.String("cursor", "", "Pagination cursor from the previous response")
	var industries repeatedString
	var types repeatedString
	var tags repeatedString
	var audiences repeatedString
	var countryCodes repeatedString
	fs.Var(&industries, "industry", "Industry filter; repeat or comma-separate")
	fs.Var(&industries, "industries", "Alias for --industry")
	fs.Var(&types, "type", "Moment type filter; repeat or comma-separate")
	fs.Var(&types, "types", "Alias for --type")
	fs.Var(&tags, "tag", "Tag filter; repeat or comma-separate")
	fs.Var(&tags, "tags", "Alias for --tag")
	fs.Var(&audiences, "audience", "Audience filter; repeat or comma-separate")
	fs.Var(&audiences, "audiences", "Alias for --audience")
	fs.Var(&countryCodes, "country-code", "Country code filter; repeat or comma-separate")
	fs.Var(&countryCodes, "country-codes", "Alias for --country-code")
	jsonInline := fs.String("json", "", "Exact JSON request body")
	jsonFile := fs.String("json-file", "", "Read exact JSON request body from file, or - for stdin")
	if bareJSONFlag(args) {
		return fail(stderr, bareJSONFlagError("pr-calendar query"))
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if body, ok, err := parseJSONFlag(*jsonInline, *jsonFile, "pr-calendar query"); err != nil {
		return fail(stderr, err)
	} else if ok {
		return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/pr-calendar/query", nil, body, nil, 45*time.Second)
	}
	body := map[string]any{}
	if value := strings.TrimSpace(firstString(*query, *q)); value != "" {
		body["q"] = value
	}
	if values := cleanCommaStrings(industries); len(values) > 0 {
		body["industries"] = values
	}
	if values := cleanCommaStrings(types); len(values) > 0 {
		body["types"] = values
	}
	if values := cleanCommaStrings(tags); len(values) > 0 {
		body["tags"] = values
	}
	if values := cleanCommaStrings(audiences); len(values) > 0 {
		body["audiences"] = values
	}
	if values := cleanCommaStrings(countryCodes); len(values) > 0 {
		body["country_codes"] = values
	}
	if value := strings.TrimSpace(firstString(*from, *startDate)); value != "" {
		body["start_date"] = value
	}
	if value := strings.TrimSpace(firstString(*to, *endDate)); value != "" {
		body["end_date"] = value
	}
	if strings.TrimSpace(*pitchStartBefore) != "" {
		body["pitch_start_before"] = strings.TrimSpace(*pitchStartBefore)
	}
	if strings.TrimSpace(*pitchDeadlineAfter) != "" {
		body["pitch_deadline_after"] = strings.TrimSpace(*pitchDeadlineAfter)
	}
	if *limit > 0 {
		body["limit"] = *limit
	}
	if strings.TrimSpace(*cursor) != "" {
		body["cursor"] = strings.TrimSpace(*cursor)
	}
	return runMedialystJSON(stdout, stderr, http.MethodPost, "/v1/pr-calendar/query", nil, body, nil, 45*time.Second)
}

func cleanCommaStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			out = append(out, part)
		}
	}
	return out
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
