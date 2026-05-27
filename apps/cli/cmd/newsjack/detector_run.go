package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func detectorRun(opts detectorOptions, stdout io.Writer) error {
	profile := defaultProfile()
	var err error
	if opts.ProfilePath != "" {
		profile, err = profileFromFile(opts.ProfilePath)
		if err != nil {
			return err
		}
	}
	queries := []string{}
	if !opts.FeedOnly {
		queries = buildQueries(opts, profile)
	}
	feedURLs, err := buildFeedURLs(opts, profile)
	if err != nil {
		return err
	}
	config := configFromEnv()
	requestedSources := []string{}
	if !opts.FeedOnly {
		requestedSources, err = requestedSourcesFor(opts, profile)
		if err != nil {
			return err
		}
	}
	trendRequested := contains(requestedSources, "x_trends") && !opts.FeedOnly
	if len(queries) == 0 && len(feedURLs) == 0 && !trendRequested {
		return errors.New("Provide a query, --topic, --major-feeds, --feed-url, or --profile with topics/competitors.")
	}
	querySources := querySources(requestedSources)
	sources := []string{}
	if len(queries) > 0 {
		if opts.Mock {
			sources = querySources
		} else {
			sources = availableSources(config, querySources)
		}
		if len(sources) == 0 {
			return errors.New("No requested sources are available. Configure MEDIALYST_API_KEY or X_BEARER_TOKEN, choose RSS/public sources, or rerun with --mock.")
		}
	}
	now := time.Now().UTC()
	var allSignals []map[string]any
	var seenURLsToMark []string
	sourceErrors := map[string]any{}
	evidenceSourceCounts := map[string]int{}
	hygieneRejections := map[string]int{}
	noteItems := func(items []evidenceItem) {
		for _, item := range items {
			evidenceSourceCounts[item.Source]++
		}
	}
	processItems := func(query string, items []evidenceItem, errors map[string]string) error {
		items = filterItemsByAge(items, now, opts.MaxAgeHours)
		var rejected map[string]int
		items, rejected = filterItemsByHygiene(items, !opts.NoHygieneFilter)
		mergeCounts(hygieneRejections, rejected)
		noteItems(items)
		if len(errors) > 0 {
			sourceErrors[query] = errors
		}
		clusters := clusterItems(items)
		var urls []string
		for _, cluster := range clusters {
			urls = append(urls, cluster.urls()...)
		}
		seenURLsToMark = append(seenURLsToMark, urls...)
		seen, err := seenStatus(urls, opts.Store)
		if err != nil {
			return err
		}
		for _, cluster := range clusters {
			signal := scoreSignal(cluster, profile, seen, now, opts)
			signal["query"] = query
			if !(opts.NewOnly && signalIsSeen(signal)) {
				allSignals = append(allSignals, signal)
			}
		}
		return nil
	}
	for _, query := range queries {
		items, errors := collectQuery(query, sources, config, opts, now)
		if err := processItems(query, items, errors); err != nil {
			return err
		}
	}
	if len(feedURLs) > 0 {
		items, errors := collectFeeds(feedURLs, opts.Depth, opts.Mock, now)
		if err := processItems("major_news_feed", items, errors); err != nil {
			return err
		}
	}
	if trendRequested {
		items, errors := collectXTrends(profile, config, opts, now)
		if err := processItems("x_trends", items, errors); err != nil {
			return err
		}
	}
	laneCaps := parseLaneCaps(opts.LaneCaps)
	signals := selectSignals(allSignals, opts.Limit, laneCaps, opts.MinQueuePriority, opts.MinMajorNews)
	var runID any
	runID = nil
	if opts.Save {
		id, err := recordRun(opts.MonitorName, profile.publicDict(), queries, signals, seenURLsToMark, opts.Store)
		if err != nil {
			return err
		}
		runID = id
	}
	payload := map[string]any{
		"monitor": map[string]any{
			"name":              nullableString(opts.MonitorName),
			"generated_at":      now.Format(time.RFC3339Nano),
			"profile":           profile.publicDict(),
			"queries":           nonNilStrings(queries),
			"feed_urls":         nonNilStrings(feedURLs),
			"sources_requested": requestedSources,
			"sources_used":      append(append([]string{}, sources...), append(boolSlice(trendRequested, "x_trends"), boolSlice(len(feedURLs) > 0, "major_feed")...)...),
			"lookback_days":     opts.LookbackDays,
			"max_age_hours":     opts.MaxAgeHours,
			"new_only":          opts.NewOnly,
			"depth":             opts.Depth,
			"mock":              opts.Mock,
		},
		"signals": signals,
		"diagnostics": map[string]any{
			"evidence_by_source":    evidenceSourceCounts,
			"hygiene_rejections":    hygieneRejections,
			"signals_by_lane":       countByLanes(allSignals),
			"emitted_by_lane":       countByLanes(signals),
			"lane_caps":             laneCaps,
			"selection":             map[string]any{"mode": map[bool]string{true: "lane_caps", false: "mechanical_floor"}[laneCaps != nil], "limit": opts.Limit, "min_queue_priority": opts.MinQueuePriority, "min_major_news": opts.MinMajorNews},
			"total_scored_signals":  len(allSignals),
			"total_emitted_signals": len(signals),
		},
		"source_errors": sourceErrors,
		"store": map[string]any{
			"saved":  opts.Save,
			"run_id": runID,
			"path":   storePathForOutput(opts.Store, opts.Save),
		},
	}
	if opts.IncludeAllScored {
		selected := map[string]bool{}
		for _, signal := range signals {
			if id := stringValue(signal["id"]); id != "" {
				selected[id] = true
			}
		}
		var dropped []string
		for _, signal := range allSignals {
			id := stringValue(signal["id"])
			if id != "" && !selected[id] {
				dropped = append(dropped, id)
			}
		}
		payload["debug"] = map[string]any{"all_scored_signals": allSignals, "dropped_signal_ids": nonNilStrings(dropped), "include_all_scored": true}
	}
	if opts.Emit == "brief" {
		fmt.Fprint(stdout, detectorBrief(payload))
	} else {
		writeJSON(stdout, payload)
	}
	return nil
}

func buildQueries(opts detectorOptions, profile monitorProfile) []string {
	var queries []string
	queries = append(queries, opts.Topics...)
	if len(opts.Query) > 0 {
		queries = append(queries, strings.TrimSpace(strings.Join(opts.Query, " ")))
	}
	queries = append(queries, profile.queryTerms()...)
	return dedupeStrings(queries)
}

var defaultMajorFeeds = []string{"https://www.techmeme.com/feed.xml"}

func buildFeedURLs(opts detectorOptions, profile monitorProfile) ([]string, error) {
	var feeds []string
	if !opts.NoProfileFeeds {
		feeds = append(feeds, profile.FeedURLs...)
	}
	if opts.MajorFeeds {
		envFeeds := envMajorFeeds()
		if len(envFeeds) > 0 {
			feeds = append(feeds, envFeeds...)
		} else if len(profile.FeedURLs) == 0 {
			feeds = append(feeds, defaultMajorFeeds...)
		}
	}
	feeds = append(feeds, opts.FeedURLs...)
	for _, file := range opts.FeedFiles {
		read, err := readFeedURLs(file)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, read...)
	}
	return dedupeStrings(feeds), nil
}

func envMajorFeeds() []string {
	raw := os.Getenv("NEWSJACK_MAJOR_FEEDS")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return splitFields(raw)
}

func readFeedURLs(path string) ([]string, error) {
	data, err := os.ReadFile(expandPath(path))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out, nil
}

func requestedSourcesFor(opts detectorOptions, profile monitorProfile) ([]string, error) {
	sources, err := parseSources(opts.Sources)
	if err != nil {
		return nil, err
	}
	if truthy(profile.XNews["enabled"], true) && !opts.NoXNews && !contains(sources, "x_news") {
		sources = append(sources, "x_news")
	}
	trendsMode := strings.ToLower(stringValue(profile.XTrends["mode"]))
	if trendsMode != "" && trendsMode != "none" && trendsMode != "off" && trendsMode != "false" && !opts.NoXTrends && !contains(sources, "x_trends") {
		sources = append(sources, "x_trends")
	}
	if opts.NoXNews {
		sources = removeString(sources, "x_news")
	}
	if opts.NoXTrends {
		sources = removeString(sources, "x_trends")
	}
	return sources, nil
}

func querySources(sources []string) []string {
	var out []string
	for _, s := range sources {
		if s != "x_trends" {
			out = append(out, s)
		}
	}
	return out
}

var defaultSources = []string{"news_search", "x_news", "x"}
var allSources = stringSet([]string{"news_search", "x_news", "x", "x_trends", "reddit", "hackernews"})

func parseSources(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return append([]string{}, defaultSources...), nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		key := strings.ToLower(strings.TrimSpace(part))
		switch key {
		case "":
			continue
		case "hn":
			key = "hackernews"
		case "news":
			key = "news_search"
		case "twitter", "x_posts":
			key = "x"
		}
		if !allSources[key] {
			return nil, fmt.Errorf("unsupported source for v0: %s", part)
		}
		if !contains(out, key) {
			out = append(out, key)
		}
	}
	return out, nil
}

func configFromEnv() map[string]string {
	fileEnv := envFileValues()
	get := func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fileEnv[key]
	}
	medialystKey, _ := loadAPIKey()
	if medialystKey == "" {
		medialystKey = get("MEDIALYST_API_KEY")
	}
	return map[string]string{
		"MEDIALYST_API_KEY":        medialystKey,
		"MEDIALYST_API_BASE":       get("MEDIALYST_API_BASE"),
		"MEDIALYST_NEWS_PATH":      get("MEDIALYST_NEWS_PATH"),
		"TWITTER_BEARER_TOKEN":     get("TWITTER_BEARER_TOKEN"),
		"X_BEARER_TOKEN":           get("X_BEARER_TOKEN"),
		"X_API_BEARER_TOKEN":       get("X_API_BEARER_TOKEN"),
		"TWITTER_API_BEARER_TOKEN": get("TWITTER_API_BEARER_TOKEN"),
	}
}

func envFileValues() map[string]string {
	if os.Getenv("NEWSJACK_IGNORE_DOTENV") == "1" {
		return map[string]string{}
	}
	paths := []string{}
	if root, err := newsjackRoot(); err == nil {
		paths = append(paths, filepath.Join(root, ".env"))
	}
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".env"))
	}
	paths = append(paths, filepath.Join(newsjackHome(), ".env"))
	out := map[string]string{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, raw := range strings.Split(string(data), "\n") {
			line := strings.TrimSpace(raw)
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			k = strings.TrimSpace(k)
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if k != "" {
				out[k] = v
			}
		}
	}
	return out
}

func availableSources(config map[string]string, requested []string) []string {
	var out []string
	bearer := bearerToken(config) != ""
	for _, source := range requested {
		switch {
		case source == "news_search" && config["MEDIALYST_API_KEY"] != "":
			out = append(out, source)
		case source == "x" && bearer:
			out = append(out, source)
		case source == "x_news" && bearer:
			out = append(out, source)
		case source == "x_trends" && bearer:
			out = append(out, source)
		case source == "reddit" || source == "hackernews":
			out = append(out, source)
		}
	}
	return out
}

func bearerToken(config map[string]string) string {
	for _, key := range []string{"TWITTER_BEARER_TOKEN", "X_BEARER_TOKEN", "X_API_BEARER_TOKEN", "TWITTER_API_BEARER_TOKEN"} {
		if config[key] != "" {
			return config[key]
		}
	}
	return ""
}

func mockItems(query string, now time.Time) []evidenceItem {
	today := now.Format("2006-01-02")
	h1 := sha1.Sum([]byte(query))
	h2 := sha1.Sum([]byte(query + "x"))
	return []evidenceItem{
		{Source: "news_search", Title: "Regulators open inquiry tied to " + query, URL: "https://example.com/news/" + hex.EncodeToString(h1[:])[:8], Container: "Example News", PublishedAt: today, Excerpt: "Officials are examining claims and compliance practices around " + query + ".", Engagement: map[string]any{}, Metadata: map[string]any{}},
		{Source: "x", Title: "Experts are reacting to " + query, URL: "https://x.com/example/status/" + hex.EncodeToString(h2[:])[:8], Author: "example", Container: "x.com", PublishedAt: today, Excerpt: "Thread: the " + query + " inquiry is moving faster than vendors expected.", Engagement: map[string]any{"likes": 120, "reposts": 22, "replies": 9}, Metadata: map[string]any{}},
	}
}

func mockFeedItems(now time.Time) []evidenceItem {
	published := now.Format(time.RFC3339Nano)
	return []evidenceItem{
		{Source: "major_feed", Title: "Salesforce launches free AI customer service agents for startups", URL: "https://example.com/major/salesforce-ai-agents", Container: "Example Major Feed", PublishedAt: published, Excerpt: "A major CRM vendor is targeting startup and SMB customer-support workflows with free AI agents.", Engagement: map[string]any{}, Metadata: map[string]any{"feed_title": "Example Major Feed", "feed_url": "mock://major-feed", "feed_position": 1}},
		{Source: "major_feed", Title: "Pentagon launches task force for safe deployment of AI tools", URL: "https://example.com/major/pentagon-ai-task-force", Container: "Example Major Feed", PublishedAt: published, Excerpt: "The Pentagon is studying how to deploy leading AI tools across sensitive government workflows.", Engagement: map[string]any{}, Metadata: map[string]any{"feed_title": "Example Major Feed", "feed_url": "mock://major-feed", "feed_position": 2}},
	}
}

func collectQuery(query string, sources []string, config map[string]string, opts detectorOptions, now time.Time) ([]evidenceItem, map[string]string) {
	if opts.Mock {
		return mockItems(query, now), map[string]string{}
	}
	from, to := dateRange(opts.LookbackDays)
	errors := map[string]string{}
	var items []evidenceItem
	for _, source := range sources {
		rawItems, err := collectSource(source, query, from, to, opts.Depth, config)
		if err != "" {
			errors[source] = err
		}
		for _, raw := range rawItems {
			item := evidenceFromMap(raw)
			if item.Source != "news_search" && item.Source != "x_news" && jaccard(query, item.text()) < 0.08 {
				continue
			}
			if item.Title != "" || item.Excerpt != "" {
				items = append(items, item)
			}
		}
	}
	return items, errors
}

func collectFeeds(feedURLs []string, depth string, mock bool, now time.Time) ([]evidenceItem, map[string]string) {
	if mock {
		return mockFeedItems(now), map[string]string{}
	}
	limit := map[string]int{"quick": 15, "default": 30, "deep": 60}[depth]
	var items []evidenceItem
	errors := map[string]string{}
	for _, feed := range feedURLs {
		raw, errText := collectFeed(feed, limit)
		if errText != "" {
			errors[feed] = errText
		}
		for _, r := range raw {
			item := evidenceFromMap(r)
			if item.Title != "" || item.Excerpt != "" {
				items = append(items, item)
			}
		}
	}
	return items, errors
}

func collectXTrends(profile monitorProfile, config map[string]string, opts detectorOptions, now time.Time) ([]evidenceItem, map[string]string) {
	mode := strings.ToLower(stringValue(profile.XTrends["mode"]))
	if mode == "" || mode == "none" || mode == "off" || mode == "false" {
		return nil, nil
	}
	if opts.Mock {
		return []evidenceItem{{
			Source:      "x_trends",
			Title:       "Meta Cuts 8,000 Jobs to Focus on AI Future",
			URL:         "https://x.com/search?q=Meta%20Cuts%208000%20Jobs&f=live",
			Author:      "x-trends",
			Container:   "x.com/trends",
			PublishedAt: now.Format(time.RFC3339Nano),
			Excerpt:     "Personalized X trend, 8.7K posts, trending for 5 hours.",
			Engagement:  map[string]any{"score": 500},
			Metadata:    map[string]any{"x_signal_type": "trend", "x_trend_mode": mode, "x_trend_post_count": "8.7K posts"},
		}}, nil
	}
	raw, errText := collectXTrendsRaw(profile.XTrends, opts.Depth, bearerToken(config))
	var items []evidenceItem
	for _, r := range raw {
		item := evidenceFromMap(mapXTrend(r))
		if item.Title != "" || item.Excerpt != "" {
			items = append(items, item)
		}
	}
	if errText != "" {
		return items, map[string]string{"x_trends": errText}
	}
	return items, nil
}

func collectSource(source, query, fromDate, toDate, depth string, config map[string]string) (items []map[string]any, errText string) {
	defer func() {
		if value := recover(); value != nil {
			items = nil
			errText = fmt.Sprintf("panic: %v", value)
		}
	}()
	switch source {
	case "news_search":
		items, err := searchNews(query, fromDate, toDate, limitForDepth(depth), config)
		return items, err
	case "x":
		response := searchX(query, depth, bearerToken(config))
		if err := stringValue(response["error"]); err != "" {
			return nil, err
		}
		counts := recentCountSummary(query, bearerToken(config))
		parsed := parseXResponse(response, query, counts)
		var out []map[string]any
		for _, item := range parsed {
			if keepXItem(item) {
				out = append(out, mapX(item))
			}
		}
		return out, ""
	case "x_news":
		response := searchXNews(query, depth, lookbackHours(fromDate, toDate), bearerToken(config))
		if err := stringValue(response["error"]); err != "" {
			return nil, err
		}
		var out []map[string]any
		for _, item := range parseXNewsResponse(response, query) {
			out = append(out, mapXNews(item))
		}
		return out, ""
	case "reddit":
		var out []map[string]any
		for _, item := range searchRedditPublic(query, fromDate, toDate, depth) {
			out = append(out, mapReddit(item))
		}
		return out, ""
	case "hackernews":
		response, errText := searchHackerNews(query, fromDate, toDate, depth)
		if errText != "" {
			return nil, errText
		}
		var out []map[string]any
		for _, item := range parseHackerNewsResponse(response, query) {
			out = append(out, mapHackerNews(item))
		}
		return out, ""
	default:
		return nil, "Unsupported source: " + source
	}
}

func limitForDepth(depth string) int {
	return map[string]int{"quick": 10, "default": 25, "deep": 50}[depth]
}

func dateRange(days int) (string, string) {
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -maxInt(days-1, 0))
	return from.Format("2006-01-02"), now.Format("2006-01-02")
}

func detectorDiagnose(sourcesRaw, store string, stdout, stderr io.Writer) int {
	config := configFromEnv()
	requested, err := parseSources(sourcesRaw)
	if err != nil {
		return fail(stderr, err)
	}
	available := availableSources(config, requested)
	payload := map[string]any{
		"sources_requested":      requested,
		"sources_available":      available,
		"news_search_configured": config["MEDIALYST_API_KEY"] != "",
		"x_news_available":       contains(availableSources(config, []string{"x_news"}), "x_news"),
		"x_trends_available":     contains(availableSources(config, []string{"x_trends"}), "x_trends"),
		"x_api_configured":       bearerToken(config) != "",
		"store_path":             dbPathFromEnv(store),
	}
	writeJSON(stdout, payload)
	return 0
}

func parseLaneCaps(raw string) map[string]int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	out := map[string]int{}
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		if n < 0 {
			n = 0
		}
		out[strings.TrimSpace(key)] = n
	}
	return out
}

func selectSignals(all []map[string]any, limit int, laneCaps map[string]int, minQueuePriority, minMajorNews float64) []map[string]any {
	sortedSignals := dedupeSignalsByURL(sortSignalsByQueue(all))
	if laneCaps == nil {
		var selected []map[string]any
		for _, signal := range sortedSignals {
			if passesSelectionFloor(signal, minQueuePriority, minMajorNews) {
				selected = append(selected, signal)
			}
		}
		if limit > 0 && len(selected) > limit {
			return selected[:limit]
		}
		return selected
	}
	var lanes []string
	for lane := range laneCaps {
		lanes = append(lanes, lane)
	}
	sort.Strings(lanes)
	var selected []map[string]any
	selectedIDs := map[string]bool{}
	for _, lane := range lanes {
		count := 0
		for _, signal := range sortedSignals {
			if signalLaneValue(signal) != lane || count >= laneCaps[lane] {
				continue
			}
			id := stringValue(signal["id"])
			if selectedIDs[id] {
				continue
			}
			selected = append(selected, signal)
			selectedIDs[id] = true
			count++
			if limit > 0 && len(selected) >= limit {
				return sortSignalsByQueue(selected)
			}
		}
	}
	for _, signal := range sortedSignals {
		id := stringValue(signal["id"])
		if selectedIDs[id] {
			continue
		}
		if _, ok := laneCaps[signalLaneValue(signal)]; ok {
			continue
		}
		selected = append(selected, signal)
		selectedIDs[id] = true
		if limit > 0 && len(selected) >= limit {
			break
		}
	}
	return sortSignalsByQueue(selected)
}

func passesSelectionFloor(signal map[string]any, minQueuePriority, minMajorNews float64) bool {
	mech, _ := signal["mechanical_scores"].(map[string]any)
	if queuePriority(signal) >= minQueuePriority {
		return true
	}
	return signalLaneValue(signal) == "major_news" && floatValue(mech["major_news"]) >= minMajorNews
}

func queuePriority(signal map[string]any) float64 {
	routing, _ := signal["routing"].(map[string]any)
	return floatValue(routing["queue_priority"])
}

func signalLaneValue(signal map[string]any) string {
	routing, _ := signal["routing"].(map[string]any)
	lane := stringValue(routing["lane"])
	if lane == "" {
		return "unknown"
	}
	return lane
}

func sortSignalsByQueue(signals []map[string]any) []map[string]any {
	out := append([]map[string]any{}, signals...)
	sort.SliceStable(out, func(i, j int) bool { return queuePriority(out[i]) > queuePriority(out[j]) })
	return out
}

func dedupeSignalsByURL(signals []map[string]any) []map[string]any {
	seenIDs := map[string]bool{}
	seenURLs := map[string]bool{}
	var out []map[string]any
	for _, signal := range signals {
		id := stringValue(signal["id"])
		urls := evidenceURLs(signal)
		overlap := false
		for _, u := range urls {
			if seenURLs[u] {
				overlap = true
				break
			}
		}
		if seenIDs[id] || (len(urls) > 0 && overlap) {
			continue
		}
		seenIDs[id] = true
		for _, u := range urls {
			seenURLs[u] = true
		}
		out = append(out, signal)
	}
	return out
}

func evidenceURLs(signal map[string]any) []string {
	var out []string
	for _, item := range anySlice(signal["evidence"]) {
		if m, ok := item.(map[string]any); ok {
			if u := stringValue(m["url"]); u != "" {
				out = append(out, u)
			}
		}
	}
	return out
}

func countByLanes(signals []map[string]any) map[string]int {
	out := map[string]int{}
	for _, signal := range signals {
		out[signalLaneValue(signal)]++
	}
	return sortedCountMap(out)
}

func signalIsSeen(signal map[string]any) bool {
	features, _ := signal["features"].(map[string]any)
	v, _ := features["seen_before"].(bool)
	return v
}

func detectorBrief(payload map[string]any) string {
	var lines []string
	lines = append(lines, "newsjack monitor", "")
	diagnostics, _ := payload["diagnostics"].(map[string]any)
	if len(diagnostics) > 0 {
		lines = append(lines, fmt.Sprintf("scored=%v emitted=%v", diagnostics["total_scored_signals"], diagnostics["total_emitted_signals"]))
		if m, ok := diagnostics["evidence_by_source"].(map[string]int); ok && len(m) > 0 {
			lines = append(lines, "evidence_by_source="+joinCounts(m))
		} else if m, ok := diagnostics["evidence_by_source"].(map[string]any); ok && len(m) > 0 {
			lines = append(lines, "evidence_by_source="+joinAnyCounts(m))
		}
		if m, ok := diagnostics["hygiene_rejections"].(map[string]int); ok && len(m) > 0 {
			lines = append(lines, "hygiene_rejections="+joinCounts(m))
		}
		if m, ok := diagnostics["emitted_by_lane"].(map[string]int); ok && len(m) > 0 {
			lines = append(lines, "emitted_by_lane="+joinCounts(m))
		}
		lines = append(lines, "")
	}
	for i, signal := range signalSlice(payload["signals"]) {
		features, _ := signal["features"].(map[string]any)
		routing, _ := signal["routing"].(map[string]any)
		mech, _ := signal["mechanical_scores"].(map[string]any)
		lines = append(lines, fmt.Sprintf("%d. %s (%s, %s, %s)", i+1, signal["title"], routing["lane"], features["decay_bucket"], strings.Join(toStringSlice(signal["sources"]), ", ")))
		lines = append(lines, fmt.Sprintf("   mechanical: queue_priority=%v, profile_match=%v, major_news=%v, momentum=%v", routing["queue_priority"], mech["profile_match"], mech["major_news"], mech["momentum"]))
		for _, link := range briefEvidenceLinks(signal, 3) {
			lines = append(lines, "   evidence: "+link)
		}
		if flags := anySlice(features["safety_flags"]); len(flags) > 0 {
			lines = append(lines, fmt.Sprintf("   safety_flags=%d", len(flags)))
		}
	}
	if len(signalSlice(payload["signals"])) == 0 {
		lines = append(lines, "No signals returned.")
	}
	return strings.Join(lines, "\n") + "\n"
}

func briefEvidenceLinks(signal map[string]any, limit int) []string {
	var out []string
	seen := map[string]bool{}
	for _, item := range anySlice(signal["evidence"]) {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		u := stringValue(m["url"])
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		source := stringValue(m["source"])
		title := firstString(m["title"], m["container"])
		if title != "" {
			out = append(out, fmt.Sprintf("%s: %s - %s", source, truncate(title, 110), u))
		} else {
			out = append(out, fmt.Sprintf("%s: %s", source, u))
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}
