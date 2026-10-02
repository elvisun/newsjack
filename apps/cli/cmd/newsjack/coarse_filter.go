package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// coarse-filter runs the relevance-coarse-filter pass through Jev, TypeSafe
// AI's typed-decision model, instead of an LLM worker. The rubric in
// skills/relevance-coarse-filter/SKILL.md is translated into the embedded
// question set below; the command builds one request per signal, applies
// deterministic post-rules so typed answers cannot contradict the skill's hard
// rules, and writes coarse_relevance_decisions.json in the exact shape
// filter-apply consumes. The LLM worker path is unchanged and remains the
// default when no TypeSafe key is configured.

//go:embed coarse_filter_questions.json
var defaultCoarseQuestionsJSON []byte

const (
	typesafeDefaultBaseURL = "https://api.typesafe.ai"
	typesafeDefaultModel   = "jev-latest"
	typesafeSystemOnePath  = "/v1/systemone"
	typesafeAPIKeyURL      = "https://typesafe.ai"
	// List price per million input tokens on 2026-09-18; output is free.
	typesafeInputPricePerMTok = 0.042

	coarseMaxEvidence       = 5
	coarseMaxExcerptChars   = 600
	coarseDefaultConcurreny = 8
	coarseMaxFailureRate    = 0.2
	coarseUncertainReject   = 0.55
	coarseBridgeThreshold   = 0.5
	coarseSafetyThreshold   = 0.7
	coarsePromoThreshold    = 0.7
)

var (
	coarseJunkReasons = stringSet([]string{"keyword_collision", "not_news", "owned_docs_or_product_page", "seo_landing_page", "competitor_or_promotional", "low_reach_x_post", "safety_risk", "duplicate", "off_beat", "no_profile_bridge"})
	coarseKeepReasons = stringSet([]string{"relevant_news", "plausible_client_bridge"})
	// coarseRetryBase is the first backoff step on 429/5xx; tests shrink it.
	coarseRetryBase = 500 * time.Millisecond
	coarseAttempts  = 4
)

type coarseFilterOptions struct {
	Engine         string
	CandidatesPath string
	ProfilePath    string
	OutputPath     string
	QuestionsPath  string
	Model          string
	Concurrency    int
	Timeout        time.Duration
	PrintQuestions bool
	DryRun         bool
}

// coarseSignal is the normalized, judgable surface of one candidate. It is
// built from either a full candidates.json signal or a harness chunk signal
// (which flattens routing/story_size and names the id signal_id).
type coarseSignal struct {
	ID             string
	State          map[string]any
	EvidenceURLs   []string
	ProfileMatches []string
	Promotional    bool
}

type coarseCallResult struct {
	Answers      map[string]any
	InputTokens  int
	OutputTokens int
	Model        string
	LatencyMS    int64
	Err          error
}

func cmdCoarseFilter(args []string, stdout, stderr io.Writer) int {
	var opts coarseFilterOptions
	fs := flag.NewFlagSet("coarse-filter", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&opts.Engine, "engine", "jev", "Coarse-filter engine. Only jev (TypeSafe AI) is supported.")
	fs.StringVar(&opts.CandidatesPath, "candidates", "", "Detector candidates.json or a harness coarse chunk file")
	fs.StringVar(&opts.ProfilePath, "profile", "", "Monitor profile JSON; overrides the profile embedded in --candidates")
	fs.StringVar(&opts.OutputPath, "output", "", "Output path for coarse_relevance_decisions.json (default stdout)")
	fs.StringVar(&opts.QuestionsPath, "questions", "", "Question set JSON to send instead of the embedded default")
	fs.StringVar(&opts.Model, "model", typesafeDefaultModel, "TypeSafe model id")
	fs.IntVar(&opts.Concurrency, "concurrency", coarseDefaultConcurreny, "Parallel calls")
	fs.DurationVar(&opts.Timeout, "timeout", 30*time.Second, "Per-call HTTP timeout")
	fs.BoolVar(&opts.PrintQuestions, "print-questions", false, "Print the question set and exit")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Build requests and print token/cost estimates without calling the API")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if opts.Engine != "jev" {
		return failf(stderr, "unsupported --engine %q: only jev is available", opts.Engine)
	}
	questions, err := loadCoarseQuestions(opts.QuestionsPath)
	if err != nil {
		return fail(stderr, err)
	}
	if opts.PrintQuestions {
		writeJSON(stdout, questions)
		return 0
	}
	if opts.CandidatesPath == "" {
		return fail(stderr, errors.New("--candidates is required"))
	}
	candidates, err := readJSONMap(opts.CandidatesPath)
	if err != nil {
		return fail(stderr, err)
	}
	profile, err := coarseProfile(candidates, opts.ProfilePath)
	if err != nil {
		return fail(stderr, err)
	}
	signals := coarseSignals(candidates)
	if len(signals) == 0 {
		return fail(stderr, errors.New("no signals found in --candidates"))
	}
	clientState := coarseClientState(profile)

	if opts.DryRun {
		writeJSON(stdout, coarseDryRun(signals, clientState, questions, opts.Model))
		return 0
	}

	apiKey, keySource := loadTypeSafeAPIKey()
	if apiKey == "" {
		return failf(stderr, "TypeSafe API key not configured; Jev coarse filtering is unavailable. Run: newsjack auth set-typesafe --key <key> (or set %s). Get a key at %s. The LLM worker path in skills/relevance-coarse-filter still works without it.", envTypeSafeKey, typesafeAPIKeyURL)
	}
	if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
		fmt.Fprintf(stderr, "Loaded TypeSafe credentials from %s\n", keySource)
	}
	client := typesafeClient{BaseURL: typesafeBaseURL(), APIKey: apiKey, Model: opts.Model, Timeout: opts.Timeout}

	started := time.Now()
	results := make([]coarseCallResult, len(signals))
	runCoarsePool(len(signals), maxInt(1, opts.Concurrency), func(i int) {
		results[i] = client.judge(map[string]any{"client": clientState, "signal": signals[i].State}, questions)
	})
	elapsed := time.Since(started)

	decisions := make([]map[string]any, 0, len(signals))
	engine := map[string]any{
		"name": "jev", "model": opts.Model, "base_url": client.BaseURL,
		"questions_sha256": sha256Hex(marshalJSON(questions)),
		"signals":          len(signals), "calls": len(signals), "concurrency": maxInt(1, opts.Concurrency),
	}
	failures, inputTokens, outputTokens := 0, 0, 0
	var latencies []int64
	for i, signal := range signals {
		res := results[i]
		if res.Err != nil {
			failures++
			decisions = append(decisions, coarseFailureDecision(signal, res.Err))
			continue
		}
		inputTokens += res.InputTokens
		outputTokens += res.OutputTokens
		latencies = append(latencies, res.LatencyMS)
		if res.Model != "" {
			engine["model"] = res.Model
		}
		decisions = append(decisions, coarseDecisionFromAnswers(signal, res.Answers))
	}
	engine["failures"] = failures
	engine["input_tokens"] = inputTokens
	engine["output_tokens"] = outputTokens
	engine["est_cost_usd"] = roundN(float64(inputTokens)/1e6*typesafeInputPricePerMTok, 4)
	engine["pricing_note"] = fmt.Sprintf("list price $%.3f per million input tokens, output free (2026-09-18)", typesafeInputPricePerMTok)
	engine["elapsed_ms"] = elapsed.Milliseconds()
	if len(latencies) > 0 {
		sort.Slice(latencies, func(a, b int) bool { return latencies[a] < latencies[b] })
		engine["latency_ms_p50"] = latencies[len(latencies)/2]
		engine["latency_ms_max"] = latencies[len(latencies)-1]
	}
	engine["decision_counts"] = coarseDecisionCounts(decisions)

	output := map[string]any{
		"version":      1,
		"generated_at": time.Now().UTC().Format(time.RFC3339Nano),
		"engine":       engine,
		"decisions":    decisions,
	}
	data := marshalJSON(output)
	if opts.OutputPath != "" {
		if err := os.WriteFile(expandPath(opts.OutputPath), data, 0o644); err != nil {
			return fail(stderr, err)
		}
	} else {
		stdout.Write(data)
	}
	if float64(failures) > coarseMaxFailureRate*float64(len(signals)) {
		return failf(stderr, "jev coarse filter: %d of %d calls failed (over %.0f%%); output written with failed signals kept as monitor_only. Fall back to the LLM worker path for this run.", failures, len(signals), coarseMaxFailureRate*100)
	}
	if failures > 0 {
		warn(stderr, "jev coarse filter: %d of %d calls failed; those signals were kept as monitor_only", failures, len(signals))
	}
	return 0
}

func loadCoarseQuestions(path string) (map[string]any, error) {
	data := defaultCoarseQuestionsJSON
	if path != "" {
		var err error
		data, err = os.ReadFile(expandPath(path))
		if err != nil {
			return nil, err
		}
	}
	var questions map[string]any
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("question set is not a JSON object: %w", err)
	}
	for _, key := range []string{"decision", "reason"} {
		q := valueOrEmptyMap(questions[key])
		if stringValue(q["type"]) != "choice" {
			return nil, fmt.Errorf("question set must define a %q choice question", key)
		}
	}
	return questions, nil
}

func coarseProfile(candidates map[string]any, profilePath string) (monitorProfile, error) {
	if profilePath != "" {
		raw, err := readJSONMap(profilePath)
		if err != nil {
			return monitorProfile{}, err
		}
		return profileFromMap(raw), nil
	}
	if raw := valueOrEmptyMap(valueOrEmptyMap(candidates["monitor"])["profile"]); len(raw) > 0 {
		return profileFromMap(raw), nil
	}
	if raw := valueOrEmptyMap(candidates["profile_context"]); len(raw) > 0 {
		return profileFromMap(raw), nil
	}
	return monitorProfile{}, errors.New("no profile found: expected monitor.profile or profile_context in --candidates, or pass --profile")
}

func coarseClientState(profile monitorProfile) map[string]any {
	return map[string]any{
		"company":      profile.Company,
		"description":  profile.Description,
		"website":      stringValue(profile.Raw["website"]),
		"topics":       nonNilStrings(profile.Topics),
		"competitors":  nonNilStrings(profile.Competitors),
		"standing":     nonNilStrings(profile.Standing),
		"search_terms": nonNilStrings(profile.SearchTerms),
		"exclusions":   nonNilStrings(profile.Exclusions),
	}
}

func coarseSignals(candidates map[string]any) []coarseSignal {
	var out []coarseSignal
	for _, signal := range signalSlice(candidates["signals"]) {
		cs := coarseSignalFromMap(signal)
		if cs.ID == "" {
			continue
		}
		out = append(out, cs)
	}
	return out
}

func coarseSignalFromMap(signal map[string]any) coarseSignal {
	features := valueOrEmptyMap(signal["features"])
	routing := valueOrEmptyMap(signal["routing"])
	storySize := valueOrEmptyMap(signal["story_size"])
	id := strings.TrimSpace(firstString(signal["id"], signal["signal_id"]))
	matches := dedupeStrings(append(toStringSlice(signal["profile_matches"]), toStringSlice(features["profile_matches"])...))
	safety := dedupeStrings(append(toStringSlice(signal["safety_flags"]), toStringSlice(features["safety_flags"])...))
	band := firstString(storySize["band"], signal["story_size_band"], "unknown")
	sizeState := map[string]any{"band": band}
	if hint := valueOrEmptyMap(storySize["attention_hint"]); len(hint) > 0 {
		sizeState["attention_hint"] = hint
	}

	var evidence []map[string]any
	var urls []string
	promotional := false
	for _, raw := range anySlice(signal["evidence"]) {
		ev, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		metadata := valueOrEmptyMap(ev["metadata"])
		pubType := firstString(ev["publication_type"], metadata["publication_type"])
		if isPromotionalPublication(map[string]any{"publication_type": pubType}) {
			promotional = true
		}
		if u := stringValue(ev["url"]); u != "" {
			urls = append(urls, u)
		}
		if len(evidence) >= coarseMaxEvidence {
			continue
		}
		item := map[string]any{
			"title":        stringValue(ev["title"]),
			"url":          stringValue(ev["url"]),
			"container":    stringValue(ev["container"]),
			"source":       stringValue(ev["source"]),
			"excerpt":      truncate(stringValue(ev["excerpt"]), coarseMaxExcerptChars),
			"published_at": stringValue(ev["published_at"]),
		}
		if pubType != "" {
			item["publication_type"] = pubType
		}
		if followers := firstNonNil(ev["x_author_followers"], metadata["x_author_followers"]); followers != nil {
			item["x_author_followers"] = followers
		}
		evidence = append(evidence, item)
	}
	if evidence == nil {
		evidence = []map[string]any{}
	}
	state := map[string]any{
		"id":              id,
		"title":           stringValue(signal["title"]),
		"query":           stringValue(signal["query"]),
		"lane":            firstString(routing["lane"], signal["lane"]),
		"sources":         nonNilStrings(toStringSlice(signal["sources"])),
		"profile_matches": nonNilStrings(matches),
		"safety_flags":    nonNilStrings(safety),
		"story_size":      sizeState,
		"evidence":        evidence,
	}
	return coarseSignal{ID: id, State: state, EvidenceURLs: firstN(dedupeStrings(urls), coarseMaxEvidence), ProfileMatches: matches, Promotional: promotional}
}

func coarseDryRun(signals []coarseSignal, clientState map[string]any, questions map[string]any, model string) map[string]any {
	estTokens := 0
	var sample map[string]any
	for i, signal := range signals {
		body := map[string]any{"model": model, "state": map[string]any{"client": clientState, "signal": signal.State}, "questions": questions}
		estTokens += len(marshalJSON(body)) / 4
		if i == 0 {
			sample = body
		}
	}
	return map[string]any{
		"engine":           "jev",
		"model":            model,
		"signals":          len(signals),
		"calls":            len(signals),
		"est_input_tokens": estTokens,
		"est_cost_usd":     roundN(float64(estTokens)/1e6*typesafeInputPricePerMTok, 4),
		"pricing_note":     fmt.Sprintf("list price $%.3f per million input tokens, output free (2026-09-18); tokens estimated at 4 bytes each", typesafeInputPricePerMTok),
		"sample_request":   sample,
	}
}

func runCoarsePool(n, workers int, fn func(i int)) {
	var wg sync.WaitGroup
	next := make(chan int)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				fn(i)
			}
		}()
	}
	for i := 0; i < n; i++ {
		next <- i
	}
	close(next)
	wg.Wait()
}

// --- TypeSafe client ---------------------------------------------------------

type typesafeClient struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func typesafeBaseURL() string {
	return strings.TrimRight(getenv(envTypeSafeBaseURL, typesafeDefaultBaseURL), "/")
}

func (c typesafeClient) judge(state map[string]any, questions map[string]any) coarseCallResult {
	body := map[string]any{"model": c.Model, "state": state, "questions": questions}
	headers := map[string]string{"Authorization": "Bearer " + c.APIKey}
	started := time.Now()
	var lastErr error
	for attempt := 0; attempt < coarseAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(coarseRetryBase * time.Duration(1<<uint(attempt-1)))
		}
		payload, err := httpJSON("POST", c.BaseURL+typesafeSystemOnePath, headers, body, c.Timeout)
		if err != nil {
			lastErr = err
			if typesafeRetryable(err) {
				continue
			}
			return coarseCallResult{Err: err}
		}
		answers := valueOrEmptyMap(payload["answers"])
		if len(answers) == 0 {
			return coarseCallResult{Err: errors.New("TypeSafe response has no answers")}
		}
		usage := valueOrEmptyMap(payload["usage"])
		return coarseCallResult{
			Answers:      answers,
			InputTokens:  intValue(usage["input_tokens"], 0),
			OutputTokens: intValue(usage["output_tokens"], 0),
			Model:        stringValue(payload["model"]),
			LatencyMS:    time.Since(started).Milliseconds(),
		}
	}
	if lastErr == nil {
		lastErr = errors.New("TypeSafe: gave up")
	}
	return coarseCallResult{Err: lastErr}
}

// typesafeRetryable treats rate limits, upstream errors, and transport
// failures as retryable; 4xx other than 429 is a request problem.
func typesafeRetryable(err error) bool {
	msg := err.Error()
	if !strings.HasPrefix(msg, "HTTP ") {
		return true
	}
	var code int
	if _, scanErr := fmt.Sscanf(msg, "HTTP %d", &code); scanErr != nil {
		return true
	}
	return code == 429 || code >= 500
}

// --- Post-rules --------------------------------------------------------------

type coarseChoice struct {
	Choice        string
	Probabilities map[string]float64
	Confidence    float64
	HasConfidence bool
}

func coarseChoiceAnswer(answers map[string]any, key string) coarseChoice {
	raw := valueOrEmptyMap(answers[key])
	out := coarseChoice{Choice: stringValue(raw["choice"]), Probabilities: map[string]float64{}}
	for k, v := range valueOrEmptyMap(raw["probabilities"]) {
		if f, ok := numberValue(v); ok {
			out.Probabilities[k] = f
		}
	}
	if f, ok := numberValue(raw["confidence"]); ok {
		out.Confidence, out.HasConfidence = f, true
	} else if p, ok := out.Probabilities[out.Choice]; ok {
		out.Confidence, out.HasConfidence = p, true
	}
	if out.Choice == "" {
		out.Choice = coarseTopChoice(out.Probabilities, nil)
	}
	return out
}

func coarseNoul(answers map[string]any, key string) (float64, bool) {
	raw := valueOrEmptyMap(answers[key])
	return numberValue(raw["noul"])
}

// coarseTopChoice returns the highest-probability option, restricted to
// allowed when allowed is non-nil. Ties break alphabetically for determinism.
func coarseTopChoice(probs map[string]float64, allowed map[string]bool) string {
	best, bestP := "", -1.0
	keys := make([]string, 0, len(probs))
	for k := range probs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if allowed != nil && !allowed[k] {
			continue
		}
		if probs[k] > bestP {
			best, bestP = k, probs[k]
		}
	}
	return best
}

func coarseReasonsFor(decision string) map[string]bool {
	switch decision {
	case "reject":
		return coarseJunkReasons
	case "keep":
		return coarseKeepReasons
	default:
		all := map[string]bool{"major_news_no_bridge": true}
		for k := range coarseJunkReasons {
			all[k] = true
		}
		for k := range coarseKeepReasons {
			all[k] = true
		}
		return all
	}
}

func coarseDefaultReason(decision string) string {
	switch decision {
	case "reject":
		return "off_beat"
	case "keep":
		return "relevant_news"
	default:
		return "plausible_client_bridge"
	}
}

// coarseDecisionFromAnswers turns one Jev answer set into a decision in the
// relevance-coarse-filter contract. Post-rules run in a fixed order so the
// typed answers cannot contradict the skill's hard rules; each rule that
// fires is recorded in post_rules. Big-story recall is intentionally left to
// filter-apply, which already upgrades any reject on a high/major signal.
func coarseDecisionFromAnswers(signal coarseSignal, answers map[string]any) map[string]any {
	var rules []string
	decisionAns := coarseChoiceAnswer(answers, "decision")
	reasonAns := coarseChoiceAnswer(answers, "reason")
	decision := decisionAns.Choice
	if !allowedDecisions[decision] {
		if top := coarseTopChoice(decisionAns.Probabilities, allowedDecisions); top != "" {
			decision = top
		} else {
			decision = "monitor_only"
		}
		rules = append(rules, "decision_normalized")
	}
	reason := reasonAns.Choice
	if allowed := coarseReasonsFor(decision); !allowed[reason] {
		if top := coarseTopChoice(reasonAns.Probabilities, allowed); top != "" {
			reason = top
		} else {
			reason = coarseDefaultReason(decision)
		}
		rules = append(rules, "reason_consistency")
	}

	isNews, hasIsNews := coarseNoul(answers, "is_news")
	bridge, hasBridge := coarseNoul(answers, "profile_bridge")
	promo, hasPromo := coarseNoul(answers, "promotional")
	safety, hasSafety := coarseNoul(answers, "safety_risk")

	bridged := len(signal.ProfileMatches) > 0 || (hasBridge && bridge >= coarseBridgeThreshold)
	if bridged && decision == "reject" {
		decision = "monitor_only"
		if reason == "no_profile_bridge" || reason == "major_news_no_bridge" {
			reason = "plausible_client_bridge"
		}
		rules = append(rules, "profile_bridge_floor")
	}
	if hasSafety && safety >= coarseSafetyThreshold {
		reason = "safety_risk"
		if decision == "keep" {
			decision = "monitor_only"
		}
		rules = append(rules, "safety_floor")
	}
	// Promotional or owned content is never rejected and never sent on as a
	// clean keep: the skill parks it at monitor_only with
	// competitor_or_promotional so standing triage can gate it.
	promotional := signal.Promotional || (hasPromo && promo >= coarsePromoThreshold) || reasonAns.Choice == "competitor_or_promotional"
	if promotional && decision != "monitor_only" {
		decision = "monitor_only"
		reason = "competitor_or_promotional"
		rules = append(rules, "promotional_floor")
	}
	confidence := coarseConfidenceLabel(decisionAns)
	if decision == "reject" && decisionAns.HasConfidence && decisionAns.Confidence < coarseUncertainReject {
		decision = "monitor_only"
		confidence = "low"
		rules = append(rules, "uncertainty_floor")
	}
	if len(rules) > 0 && confidence == "high" {
		confidence = "medium"
	}

	var rationale strings.Builder
	fmt.Fprintf(&rationale, "Jev: %s", decisionAns.Choice)
	if decisionAns.HasConfidence {
		fmt.Fprintf(&rationale, " (%.2f)", decisionAns.Confidence)
	}
	fmt.Fprintf(&rationale, "; %s", reasonAns.Choice)
	if p, ok := reasonAns.Probabilities[reasonAns.Choice]; ok {
		fmt.Fprintf(&rationale, " (%.2f)", p)
	}
	rationale.WriteString(".")
	var nouls []string
	if hasIsNews {
		nouls = append(nouls, fmt.Sprintf("is_news %.2f", isNews))
	}
	if hasBridge {
		nouls = append(nouls, fmt.Sprintf("profile_bridge %.2f", bridge))
	}
	if hasPromo {
		nouls = append(nouls, fmt.Sprintf("promotional %.2f", promo))
	}
	if hasSafety {
		nouls = append(nouls, fmt.Sprintf("safety_risk %.2f", safety))
	}
	if len(nouls) > 0 {
		rationale.WriteString(" " + strings.Join(nouls, ", ") + ".")
	}
	if len(rules) > 0 {
		rationale.WriteString(" Post-rules: " + strings.Join(rules, ", ") + ".")
	}

	var basis []string
	if len(signal.ProfileMatches) > 0 {
		basis = append(basis, "Profile matched: "+strings.Join(firstN(signal.ProfileMatches, 5), ", ")+".")
	} else if hasBridge && bridge >= coarseBridgeThreshold {
		basis = append(basis, fmt.Sprintf("Jev found a profile bridge (%.2f).", bridge))
	} else if hasBridge {
		basis = append(basis, fmt.Sprintf("No profile entity matched (bridge %.2f).", bridge))
	}
	if hasIsNews && isNews < 0.5 {
		basis = append(basis, fmt.Sprintf("Unlikely to be reported news (%.2f).", isNews))
	}
	if signal.Promotional {
		basis = append(basis, "Evidence metadata marks this as promotional or newswire content.")
	}
	if len(basis) == 0 {
		basis = append(basis, "Typed judgment only; no prose rationale is available from this engine.")
	}

	out := map[string]any{
		"signal_id":       signal.ID,
		"decision":        decision,
		"reason":          reason,
		"rationale":       rationale.String(),
		"confidence":      confidence,
		"evidence_urls":   nonNilStrings(signal.EvidenceURLs),
		"relevance_basis": strings.Join(basis, " "),
		"engine":          "jev",
		"answers":         answers,
		"post_rules":      nonNilStrings(rules),
	}
	return out
}

func coarseConfidenceLabel(ans coarseChoice) string {
	if !ans.HasConfidence {
		return "medium"
	}
	switch {
	case ans.Confidence >= 0.8:
		return "high"
	case ans.Confidence >= 0.6:
		return "medium"
	default:
		return "low"
	}
}

func coarseFailureDecision(signal coarseSignal, err error) map[string]any {
	return map[string]any{
		"signal_id":       signal.ID,
		"decision":        "monitor_only",
		"reason":          "plausible_client_bridge",
		"rationale":       "Jev call failed: " + cleanError(err.Error()) + "; kept for review.",
		"confidence":      "low",
		"evidence_urls":   nonNilStrings(signal.EvidenceURLs),
		"relevance_basis": "Engine error; no judgment was made. Kept so the expensive pass can decide.",
		"engine":          "jev",
		"error":           cleanError(err.Error()),
		"post_rules":      []string{"engine_failure"},
	}
}

func coarseDecisionCounts(decisions []map[string]any) map[string]int {
	counts := map[string]int{}
	for _, d := range decisions {
		counts[stringValue(d["decision"])]++
	}
	return sortedCountMap(counts)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func printCoarseFilterHelp(w io.Writer) {
	uiProduct(w, "coarse-filter", "run the relevance-coarse-filter pass through Jev (TypeSafe AI) instead of an LLM worker.")
	fmt.Fprintln(w)
	uiSection(w, "usage")
	fmt.Fprintln(w, "  newsjack coarse-filter --engine jev --candidates candidates.json --output coarse_relevance_decisions.json")
	fmt.Fprintln(w, "  newsjack coarse-filter --candidates coarse_chunk.1.json --dry-run")
	fmt.Fprintln(w, "  newsjack coarse-filter --print-questions")
	fmt.Fprintln(w)
	uiSection(w, "options")
	uiCommand(w, "--engine", "engine name; only jev is available", "jev")
	uiCommand(w, "--candidates", "detector candidates.json or a harness coarse chunk ({profile_context, signals})", "F")
	uiCommand(w, "--profile", "monitor profile JSON; overrides the profile inside --candidates", "F")
	uiCommand(w, "--output", "write coarse_relevance_decisions.json here instead of stdout", "F")
	uiCommand(w, "--questions", "question set JSON to send instead of the embedded default", "F")
	uiCommand(w, "--model", "TypeSafe model id", typesafeDefaultModel)
	uiCommand(w, "--concurrency", "parallel calls", fmt.Sprint(coarseDefaultConcurreny))
	uiCommand(w, "--dry-run", "build every request and print token and cost estimates without calling the API", "")
	uiCommand(w, "--print-questions", "print the question set the rubric is translated into", "")
	fmt.Fprintln(w)
	uiSection(w, "behavior")
	uiKV(w, "contract", "same decisions file the LLM worker path produces; feed it to filter-apply unchanged")
	uiKV(w, "one call per signal", "six typed questions: decision, reason, is_news, profile_bridge, promotional, safety_risk")
	uiKV(w, "post-rules", "reason consistency, profile-bridge floor, safety floor, promotional floor, uncertainty floor; big-story recall stays in filter-apply")
	uiKV(w, "failures", "a failed call keeps the signal as monitor_only; more than 20% failures exits non-zero so the skill falls back to LLM workers")
	uiKV(w, "credentials", "TYPESAFE_API_KEY via env, .env, or: newsjack auth set-typesafe --key <key>")
	uiKV(w, "base URL", envTypeSafeBaseURL+" overrides "+typesafeDefaultBaseURL)
	uiNote(w, "The rubric in skills/relevance-coarse-filter/SKILL.md is the source of truth; the question set is its typed translation.")
}
