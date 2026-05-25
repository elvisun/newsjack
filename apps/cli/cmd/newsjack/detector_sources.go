package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func searchNews(query, fromDate, toDate string, limit int, config map[string]string) ([]map[string]any, string) {
	apiKey := config["MEDIALYST_API_KEY"]
	if apiKey == "" {
		return nil, ""
	}
	base := config["MEDIALYST_API_BASE"]
	if base == "" {
		base = "https://medialyst.ai/api"
	}
	path := config["MEDIALYST_NEWS_PATH"]
	if path == "" {
		path = "/v1/news/search"
	}
	body := map[string]any{"q": query, "num": limit, "gl": "us", "hl": "en", "tbs": tbsForRange(fromDate, toDate)}
	payload, err := httpJSON("POST", strings.TrimRight(base, "/")+path, map[string]string{"Authorization": "Bearer " + apiKey}, body, 30*time.Second)
	if err != nil {
		return nil, err.Error()
	}
	return parseNewsResponse(payload), ""
}

func parseNewsResponse(payload map[string]any) []map[string]any {
	rawItems := firstArray(payload, "items", "results", "news", "organic")
	var items []map[string]any
	for i, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := firstAny(item, "title", "headline", "name")
		link := firstAny(item, "url", "link")
		snippet := firstAny(item, "snippet", "summary", "description", "content")
		source := firstAny(item, "source", "publication", "publisher", "site")
		if sm, ok := source.(map[string]any); ok {
			source = firstAny(sm, "name", "domain")
		}
		published := firstAny(item, "published_at", "publishedAt", "published", "date", "created_at")
		if stringValue(title) == "" && stringValue(snippet) == "" {
			continue
		}
		id := firstString(firstAny(item, "id", "uuid"), fmt.Sprintf("ML%d", i+1))
		items = append(items, map[string]any{
			"id":           id,
			"source":       "news_search",
			"title":        strings.TrimSpace(firstString(title, snippet)),
			"url":          strings.TrimSpace(stringValue(link)),
			"author":       nullableString(stringValue(firstAny(item, "author", "byline"))),
			"container":    nullableString(strings.TrimSpace(stringValue(source))),
			"published_at": nullableString(normalizeLooseDate(stringValue(published))),
			"excerpt":      strings.TrimSpace(stringValue(snippet)),
			"engagement":   map[string]any{},
			"metadata":     map[string]any{"raw_source": source},
		})
	}
	return items
}

func collectFeed(urlOrPath string, limit int) ([]map[string]any, string) {
	var text string
	if regexp.MustCompile(`(?i)^https?://`).MatchString(urlOrPath) {
		resp, err := httpGetRaw(urlOrPath, map[string]string{"Accept": "application/rss+xml, application/atom+xml, text/xml, */*"}, 20*time.Second)
		if err != nil {
			return nil, fmt.Sprintf("%T: %v", err, err)
		}
		text = resp
	} else {
		data, err := os.ReadFile(expandPath(urlOrPath))
		if err != nil {
			return nil, fmt.Sprintf("%T: %v", err, err)
		}
		text = string(data)
	}
	items, err := parseFeed(text, urlOrPath, limit)
	if err != nil {
		return nil, fmt.Sprintf("%T: %v", err, err)
	}
	return items, ""
}

func parseFeed(xmlText, feedURL string, limit int) ([]map[string]any, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlText))
	var stack []string
	var current map[string]string
	var inItem bool
	feedTitle := feedURL
	var items []map[string]any
	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			stack = append(stack, name)
			if name == "item" || name == "entry" {
				inItem = true
				current = map[string]string{}
			}
			if inItem && name == "link" {
				for _, attr := range t.Attr {
					if strings.ToLower(attr.Name.Local) == "href" {
						current["link"] = attr.Value
					}
				}
			}
		case xml.EndElement:
			name := strings.ToLower(t.Name.Local)
			if (name == "item" || name == "entry") && inItem {
				position := len(items) + 1
				if item := feedItemDict(current, feedTitle, feedURL, position); item != nil {
					items = append(items, item)
					if len(items) >= limit {
						return items, nil
					}
				}
				inItem = false
				current = nil
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text == "" || len(stack) == 0 {
				continue
			}
			name := stack[len(stack)-1]
			if inItem {
				if current[name] == "" {
					current[name] = text
				} else {
					current[name] += " " + text
				}
			} else if name == "title" && feedTitle == feedURL {
				feedTitle = cleanText(text)
			}
		}
	}
	return items, nil
}

func feedItemDict(m map[string]string, feedTitle, feedURL string, position int) map[string]any {
	title := cleanText(m["title"])
	link := firstString(m["link"], m["guid"], m["id"])
	excerpt := cleanText(firstString(m["description"], m["summary"], m["content"]))
	published := normalizeLooseDate(firstString(m["pubDate"], m["published"], m["updated"]))
	container := cleanText(firstString(m["source"], feedTitle))
	guid := firstString(m["guid"], m["id"])
	if title == "" && excerpt == "" {
		return nil
	}
	if title == "" {
		title = truncate(excerpt, 120)
	}
	id := firstString(guid, link, fmt.Sprintf("%s#%d", feedURL, position))
	return map[string]any{
		"id":           id,
		"source":       "major_feed",
		"title":        title,
		"url":          strings.TrimSpace(link),
		"author":       nil,
		"container":    firstString(container, feedTitle),
		"published_at": nullableString(published),
		"excerpt":      excerpt,
		"engagement":   map[string]any{},
		"metadata":     map[string]any{"feed_title": feedTitle, "feed_url": feedURL, "feed_position": position},
	}
}

func cleanText(value string) string {
	text := html.UnescapeString(value)
	text = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`(?s)<[^>]+>`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func xurlAvailable() bool {
	ctx, cancel := contextWithTimeout(10 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xurl", "whoami")
	out, err := cmd.Output()
	return err == nil && strings.Contains(string(out), `"username"`)
}

func searchX(query, depth string) map[string]any {
	maxResults := maxInt(10, minInt(100, map[string]int{"quick": 10, "default": 30, "deep": 60}[depth]))
	normalized := normalizeXQuery(query)
	params := url.Values{}
	params.Set("query", normalized)
	params.Set("max_results", strconv.Itoa(maxResults))
	params.Set("sort_order", "relevancy")
	params.Set("tweet.fields", "created_at,public_metrics,author_id,conversation_id,referenced_tweets,lang,possibly_sensitive")
	params.Set("expansions", "author_id")
	params.Set("user.fields", "username,name,verified,is_identity_verified,public_metrics")
	response := xurlGet("/2/tweets/search/recent?" + params.Encode())
	if response["error"] == nil {
		response["_newsjack_query"] = normalized
		return response
	}
	fallback := xurlSearchShortcut(query, maxResults)
	if fallback["error"] != nil {
		return response
	}
	fallback["_newsjack_query"] = query
	return fallback
}

func recentCountSummary(query, bearer string) map[string]any {
	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(-24 * time.Hour)
	params := url.Values{}
	params.Set("query", normalizeXQuery(query))
	params.Set("granularity", "hour")
	params.Set("start_time", isoZ(start))
	params.Set("end_time", isoZ(now))
	path := "/2/tweets/counts/recent?" + params.Encode()
	var response map[string]any
	if bearer != "" {
		response = apiGet(path, bearer)
	} else {
		response = xurlGetAuth(path, "app")
	}
	if response["error"] != nil {
		return nil
	}
	return summarizeCounts(response)
}

func searchXNews(query, depth string, maxAgeHours int, bearer string) map[string]any {
	maxResults := map[string]int{"quick": 5, "default": 10, "deep": 20}[depth]
	params := url.Values{}
	params.Set("query", strings.TrimSpace(query))
	params.Set("max_results", strconv.Itoa(maxResults))
	params.Set("max_age_hours", strconv.Itoa(maxAgeHours))
	params.Set("news.fields", "id,name,summary,hook,contexts,cluster_posts_results,updated_at,keywords,category")
	path := "/2/news/search?" + params.Encode()
	if bearer != "" {
		return apiGet(path, bearer)
	}
	return xurlGet(path)
}

func parseXNewsResponse(response map[string]any, topic string) []map[string]any {
	var items []map[string]any
	for i, raw := range anySlice(response["data"]) {
		story, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(stringValue(story["name"]))
		hook := strings.TrimSpace(stringValue(story["hook"]))
		summary := strings.TrimSpace(stringValue(story["summary"]))
		if name == "" && hook == "" && summary == "" {
			continue
		}
		postIDs := storyPostIDs(story)
		text := strings.Join(nonEmpty(name, hook, summary), " ")
		engagement := map[string]any{"score": minInt(500, maxInt(1, len(postIDs))*10)}
		if len(postIDs) > 0 {
			engagement["comments"] = len(postIDs)
		}
		items = append(items, map[string]any{
			"id":            fmt.Sprintf("XNEWS%d", i+1),
			"title":         firstString(name, truncate(hook, 120)),
			"text":          truncate(text, 1200),
			"url":           "https://x.com/search?q=" + url.QueryEscape(firstString(name, topic)) + "&f=live",
			"author_handle": "x-news",
			"date":          story["updated_at"],
			"engagement":    engagement,
			"metadata": map[string]any{
				"x_signal_type":             "story_cluster",
				"x_news_id":                 firstAny(story, "id", "rest_id"),
				"x_news_category":           story["category"],
				"x_news_keywords":           valueOrEmptyArray(story["keywords"]),
				"x_news_contexts":           valueOrEmptyMap(story["contexts"]),
				"x_news_cluster_post_ids":   postIDs,
				"x_news_cluster_post_count": len(postIDs),
				"x_news_disclaimer":         story["disclaimer"],
			},
			"why_relevant": "X News story cluster",
			"relevance":    tokenOverlapRelevance(topic, text),
		})
	}
	return items
}

func collectXTrendsRaw(config map[string]any, depth, bearer string) ([]map[string]any, string) {
	mode := strings.ToLower(stringValue(config["mode"]))
	if mode == "personalized" {
		resp := xurlGet("/2/users/personalized_trends?personalized_trend.fields=trend_name%2Cpost_count%2Ccategory%2Ctrending_since")
		return parseTrendsResponse(resp, mode, "", ""), stringValue(resp["error"])
	}
	if mode == "location" {
		if bearer == "" {
			return nil, "x_trends location mode requires TWITTER_BEARER_TOKEN or X_BEARER_TOKEN"
		}
		maxResults := map[string]int{"quick": 10, "default": 20, "deep": 50}[depth]
		locations := stringListValue(config["locations"])
		var items []map[string]any
		var errors []string
		for i, raw := range anySlice(config["woeids"]) {
			woeid := stringValue(raw)
			resp := apiGet("/2/trends/by/woeid/"+url.PathEscape(woeid)+"?max_trends="+strconv.Itoa(maxResults), bearer)
			if err := stringValue(resp["error"]); err != "" {
				errors = append(errors, woeid+": "+err)
				continue
			}
			location := woeid
			if i < len(locations) {
				location = locations[i]
			}
			items = append(items, parseTrendsResponse(resp, mode, woeid, location)...)
		}
		return items, strings.Join(errors, "; ")
	}
	return nil, "Unsupported x_trends mode: " + mode
}

func parseTrendsResponse(response map[string]any, mode, woeid, location string) []map[string]any {
	if response["error"] != nil {
		return nil
	}
	var items []map[string]any
	for i, raw := range anySlice(response["data"]) {
		trend, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(stringValue(trend["trend_name"]))
		if name == "" {
			continue
		}
		countText := firstString(trend["post_count"], trend["tweet_count"])
		count := parseCountText(countText)
		context := strings.Join(nonEmpty(stringValue(trend["category"]), countText, stringValue(trend["trending_since"]), location), ", ")
		engagement := map[string]any{}
		if count > 0 {
			engagement["score"] = minInt(500, count)
		}
		items = append(items, map[string]any{
			"id":            fmt.Sprintf("XTREND%d", i+1),
			"title":         name,
			"text":          strings.TrimSpace(name + ". " + context),
			"url":           "https://x.com/search?q=" + url.QueryEscape(name) + "&f=live",
			"author_handle": "x-trends",
			"date":          time.Now().UTC().Truncate(time.Second).Format(time.RFC3339),
			"engagement":    engagement,
			"metadata": map[string]any{
				"x_signal_type":      "trend",
				"x_trend_mode":       mode,
				"x_trend_woeid":      nullableString(woeid),
				"x_trend_location":   nullableString(location),
				"x_trend_category":   trend["category"],
				"x_trend_post_count": countText,
				"x_trend_since":      trend["trending_since"],
			},
			"why_relevant": "X trend",
			"relevance":    0.5,
		})
	}
	return items
}

func parseXResponse(response map[string]any, topic string, counts map[string]any) []map[string]any {
	var items []map[string]any
	if counts != nil && truthy(counts["is_trending"], false) {
		items = append(items, trendItem(topic, counts))
	}
	data := anySlice(response["data"])
	if len(data) == 0 {
		return items
	}
	authors := map[string]map[string]any{}
	if includes, ok := response["includes"].(map[string]any); ok {
		for _, raw := range anySlice(includes["users"]) {
			if user, ok := raw.(map[string]any); ok {
				authors[stringValue(user["id"])] = user
			}
		}
	}
	for i, raw := range data {
		tweet, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		author := authors[stringValue(tweet["author_id"])]
		username := stringValue(author["username"])
		metrics := valueOrEmptyMap(tweet["public_metrics"])
		authorMetrics := valueOrEmptyMap(author["public_metrics"])
		engagement := map[string]any{}
		if len(metrics) > 0 {
			engagement = map[string]any{
				"likes":     intValue(metrics["like_count"], 0),
				"reposts":   intValue(metrics["retweet_count"], 0),
				"replies":   intValue(metrics["reply_count"], 0),
				"quotes":    intValue(metrics["quote_count"], 0),
				"bookmarks": intValue(metrics["bookmark_count"], 0),
				"views":     intValue(metrics["impression_count"], 0),
			}
		}
		verified := truthy(author["verified"], false) || truthy(author["is_identity_verified"], false)
		proof := socialProof(metrics, authorMetrics, verified)
		date := stringValue(tweet["created_at"])
		text := strings.TrimSpace(stringValue(tweet["text"]))
		u := ""
		if username != "" && stringValue(tweet["id"]) != "" {
			u = "https://x.com/" + username + "/status/" + stringValue(tweet["id"])
		}
		items = append(items, map[string]any{
			"id":            fmt.Sprintf("XURL%d", i+1),
			"text":          truncate(text, 500),
			"url":           u,
			"author_handle": username,
			"date":          nullableString(date),
			"engagement":    engagement,
			"metadata": map[string]any{
				"x_signal_type":      "post",
				"x_author_followers": intValue(authorMetrics["followers_count"], 0),
				"x_author_listed":    intValue(authorMetrics["listed_count"], 0),
				"x_author_verified":  verified,
				"x_low_reach":        len(proof) == 0,
				"x_social_proof":     proof,
				"x_query_counts":     counts,
			},
			"why_relevant": "",
			"relevance":    tokenOverlapRelevance(topic, text),
		})
	}
	return items
}

func keepXItem(item map[string]any) bool {
	metadata := valueOrEmptyMap(item["metadata"])
	if metadata["x_signal_type"] == "query_trend" {
		return true
	}
	return !truthy(metadata["x_low_reach"], false)
}

func mapX(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "x", "title": truncate(stringValue(item["text"]), 120), "url": item["url"], "author": item["author_handle"], "container": "x.com", "published_at": item["date"], "excerpt": item["text"], "engagement": valueOrEmptyMap(item["engagement"]), "metadata": valueOrEmptyMap(item["metadata"])}
}

func mapXNews(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "x_news", "title": firstString(item["title"], truncate(stringValue(item["text"]), 120)), "url": item["url"], "author": item["author_handle"], "container": "x.com/news", "published_at": item["date"], "excerpt": item["text"], "engagement": valueOrEmptyMap(item["engagement"]), "metadata": valueOrEmptyMap(item["metadata"])}
}

func mapXTrend(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "x_trends", "title": firstString(item["title"], truncate(stringValue(item["text"]), 120)), "url": item["url"], "author": item["author_handle"], "container": "x.com/trends", "published_at": item["date"], "excerpt": item["text"], "engagement": valueOrEmptyMap(item["engagement"]), "metadata": valueOrEmptyMap(item["metadata"])}
}

func mapReddit(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "reddit", "title": item["title"], "url": item["url"], "author": item["author"], "container": item["subreddit"], "published_at": item["date"], "excerpt": firstString(item["selftext"], item["body"], item["snippet"]), "engagement": valueOrEmptyMap(item["engagement"]), "metadata": map[string]any{"top_comments": valueOrEmptyArray(item["top_comments"])}}
}

func mapHackerNews(item map[string]any) map[string]any {
	return map[string]any{"id": item["id"], "source": "hackernews", "title": item["title"], "url": item["url"], "author": item["author"], "container": "news.ycombinator.com", "published_at": item["date"], "excerpt": firstString(item["text"], item["snippet"]), "engagement": valueOrEmptyMap(item["engagement"]), "metadata": map[string]any{}}
}

func xurlGet(path string) map[string]any { return xurlGetAuth(path, "") }

func xurlGetAuth(path, auth string) map[string]any {
	args := []string{}
	if auth != "" {
		args = append(args, "--auth", auth)
	}
	args = append(args, path)
	ctx, cancel := contextWithTimeout(30 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xurl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), contextDeadlineExceeded()) {
			return map[string]any{"error": "xurl request timed out (30s)"}
		}
		if errors.Is(err, exec.ErrNotFound) {
			return map[string]any{"error": "xurl not found in PATH"}
		}
		return map[string]any{"error": "xurl request failed: " + cleanError(string(out))}
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		return map[string]any{"error": "Invalid JSON from xurl: " + err.Error()}
	}
	return payload
}

func apiGet(path, bearer string) map[string]any {
	req, _ := http.NewRequest("GET", "https://api.x.com"+path, nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("%T: %v", err, err)}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if json.Unmarshal(data, &payload) != nil {
		return map[string]any{"error": "Invalid JSON from X API"}
	}
	if resp.StatusCode >= 400 {
		detail := firstString(payload["detail"], payload["title"], truncate(string(data), 300))
		if reason := stringValue(payload["reason"]); reason != "" {
			detail += " (" + reason + ")"
		}
		return map[string]any{"error": fmt.Sprintf("X API HTTP %d: %s", resp.StatusCode, detail)}
	}
	return payload
}

func xurlSearchShortcut(query string, maxResults int) map[string]any {
	ctx, cancel := contextWithTimeout(30 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xurl", "search", query, "-n", strconv.Itoa(maxResults))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{"error": "xurl search failed: " + cleanError(string(out))}
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		return map[string]any{"error": "Invalid JSON from xurl: " + err.Error()}
	}
	return payload
}

func normalizeXQuery(query string) string {
	out := strings.TrimSpace(query)
	var additions []string
	if !strings.Contains(out, "lang:") {
		additions = append(additions, "lang:en")
	}
	for _, op := range []string{"-is:retweet", "-is:reply", "-is:nullcast"} {
		if !strings.Contains(out, op) && !strings.Contains(out, op[1:]) {
			additions = append(additions, op)
		}
	}
	if len(additions) > 0 {
		out += " " + strings.Join(additions, " ")
	}
	return truncate(out, 512)
}

func socialProof(metrics, authorMetrics map[string]any, verified bool) []string {
	likes := intValue(metrics["like_count"], 0)
	reposts := intValue(metrics["retweet_count"], 0)
	replies := intValue(metrics["reply_count"], 0)
	quotes := intValue(metrics["quote_count"], 0)
	views := intValue(metrics["impression_count"], 0)
	followers := intValue(authorMetrics["followers_count"], 0)
	listed := intValue(authorMetrics["listed_count"], 0)
	total := likes + reposts + replies + quotes
	hasView := metrics["impression_count"] != nil
	minEngagement := envInt("NEWSJACK_X_MIN_ENGAGEMENT", 3)
	minFollowers := envInt("NEWSJACK_X_MIN_AUTHOR_FOLLOWERS", 2000)
	minViews := envInt("NEWSJACK_X_MIN_VIEWS", 1000)
	var proof []string
	if total >= minEngagement {
		proof = append(proof, "post_engagement")
	}
	if reposts > 0 || quotes > 0 {
		proof = append(proof, "reshared")
	}
	if views >= minViews {
		proof = append(proof, "views")
	}
	if followers >= minFollowers && !hasView {
		proof = append(proof, "author_followers")
	}
	if listed >= 25 && !hasView {
		proof = append(proof, "author_listed")
	}
	if verified && followers >= minFollowers && !hasView {
		proof = append(proof, "verified_author")
	}
	return proof
}

func summarizeCounts(response map[string]any) map[string]any {
	var counts []int
	for _, raw := range anySlice(response["data"]) {
		if m, ok := raw.(map[string]any); ok {
			counts = append(counts, intValue(m["tweet_count"], 0))
		}
	}
	total24 := sumInts(counts)
	recent6 := sumInts(lastInts(counts, 6))
	previous := counts
	if len(counts) > 6 {
		previous = counts[:len(counts)-6]
	} else {
		previous = nil
	}
	recentPerHour := 0.0
	if len(counts) > 0 {
		recentPerHour = float64(recent6) / 6.0
	}
	previousPerHour := 0.0
	if len(previous) > 0 {
		previousPerHour = float64(sumInts(previous)) / float64(len(previous))
	}
	velocity := (recentPerHour + 1.0) / (previousPerHour + 1.0)
	min24 := envInt("NEWSJACK_X_TREND_MIN_24H", 25)
	min6 := envInt("NEWSJACK_X_TREND_MIN_6H", 8)
	minVelocity := envFloat("NEWSJACK_X_TREND_MIN_VELOCITY", 2.0)
	return map[string]any{"total_24h": total24, "recent_6h": recent6, "previous_hourly_avg": roundN(previousPerHour, 2), "velocity": roundN(velocity, 2), "is_trending": total24 >= min24 && (recent6 >= min6 || velocity >= minVelocity), "bucket_count": len(counts)}
}

func trendItem(topic string, counts map[string]any) map[string]any {
	query := strings.TrimSpace(topic)
	text := fmt.Sprintf("X conversation volume for %q is elevated: %v posts in the last 6h, %v in the last 24h (velocity %vx).", query, counts["recent_6h"], counts["total_24h"], counts["velocity"])
	return map[string]any{
		"id":            "X-TREND",
		"text":          text,
		"url":           "https://x.com/search?q=" + url.QueryEscape(query) + "&f=live",
		"author_handle": "x-search",
		"date":          time.Now().UTC().Truncate(time.Second).Format(time.RFC3339),
		"engagement":    map[string]any{"score": minInt(500, intValue(counts["total_24h"], 0)), "comments": intValue(counts["recent_6h"], 0)},
		"metadata":      map[string]any{"x_signal_type": "query_trend", "x_social_proof": []string{"query_volume"}, "x_query_counts": counts},
		"why_relevant":  "X query-volume trend",
		"relevance":     0.7,
	}
}

func storyPostIDs(story map[string]any) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range anySlice(story["cluster_posts_results"]) {
		if m, ok := raw.(map[string]any); ok {
			id := firstString(m["post_id"], m["id"])
			if id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

func parseCountText(value string) int {
	text := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), ",", "")
	m := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([km]?)`).FindStringSubmatch(text)
	if len(m) == 0 {
		return 0
	}
	n, _ := strconv.ParseFloat(m[1], 64)
	if m[2] == "k" {
		n *= 1000
	} else if m[2] == "m" {
		n *= 1000000
	}
	return int(n)
}

func searchRedditPublic(query, fromDate, toDate, depth string) []map[string]any {
	limit := map[string]int{"quick": 10, "default": 25, "deep": 50}[depth]
	u := "https://www.reddit.com/search.json?q=" + url.QueryEscape(query) + "&sort=relevance&t=month&limit=" + strconv.Itoa(limit) + "&raw_json=1"
	payload, err := httpJSON("GET", u, map[string]string{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"}, nil, 15*time.Second)
	if err != nil {
		return nil
	}
	var posts []map[string]any
	data := valueOrEmptyMap(payload["data"])
	for _, childRaw := range anySlice(data["children"]) {
		child, ok := childRaw.(map[string]any)
		if !ok || child["kind"] != "t3" {
			continue
		}
		post := valueOrEmptyMap(child["data"])
		permalink := strings.TrimSpace(stringValue(post["permalink"]))
		if permalink == "" || !strings.Contains(permalink, "/comments/") {
			continue
		}
		score := intValue(post["score"], 0)
		comments := intValue(post["num_comments"], 0)
		date := ""
		if created := floatValue(post["created_utc"]); created > 0 {
			date = time.Unix(int64(created), 0).UTC().Format("2006-01-02")
		}
		if date != "" && (date < fromDate || date > toDate) {
			continue
		}
		posts = append(posts, map[string]any{
			"id":           fmt.Sprintf("R%d", len(posts)+1),
			"title":        strings.TrimSpace(stringValue(post["title"])),
			"url":          "https://www.reddit.com" + permalink,
			"score":        score,
			"num_comments": comments,
			"subreddit":    strings.TrimSpace(stringValue(post["subreddit"])),
			"author":       firstString(post["author"], "[deleted]"),
			"selftext":     truncate(stringValue(post["selftext"]), 500),
			"date":         nullableString(date),
			"engagement":   map[string]any{"score": score, "num_comments": comments, "upvote_ratio": post["upvote_ratio"]},
			"relevance":    roundN((math.Min(1.0, float64(score)/500.0)*0.6)+(math.Min(1.0, float64(comments)/200.0)*0.4), 3),
		})
	}
	sort.SliceStable(posts, func(i, j int) bool {
		return intValue(valueOrEmptyMap(posts[i]["engagement"])["score"], 0) > intValue(valueOrEmptyMap(posts[j]["engagement"])["score"], 0)
	})
	return posts
}

func searchHackerNews(query, fromDate, toDate, depth string) (map[string]any, string) {
	count := map[string]int{"quick": 15, "default": 30, "deep": 60}[depth]
	fromTS := dateToUnix(fromDate)
	toTS := dateToUnix(toDate) + 86400
	core := flattenQueryForAlgolia(query)
	params := url.Values{}
	params.Set("query", core)
	params.Set("tags", "story")
	params.Set("numericFilters", fmt.Sprintf("created_at_i>%d,created_at_i<%d,points>2", fromTS, toTS))
	params.Set("hitsPerPage", strconv.Itoa(count))
	toks := strings.Fields(core)
	if len(toks) > 1 {
		params.Set("optionalWords", strings.Join(toks[1:], " "))
	}
	payload, err := httpJSON("GET", "https://hn.algolia.com/api/v1/search?"+params.Encode(), nil, nil, 30*time.Second)
	if err != nil {
		return map[string]any{"hits": []any{}}, err.Error()
	}
	return payload, ""
}

func parseHackerNewsResponse(response map[string]any, query string) []map[string]any {
	hits := anySlice(response["hits"])
	var items []map[string]any
	for i, raw := range hits {
		hit, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := stringValue(hit["title"])
		if query != "" && !titleMatchesQuery(title, query) {
			continue
		}
		objectID := stringValue(hit["objectID"])
		points := intValue(hit["points"], 0)
		comments := intValue(hit["num_comments"], 0)
		date := ""
		if ts := intValue(hit["created_at_i"], 0); ts > 0 {
			date = time.Unix(int64(ts), 0).UTC().Format("2006-01-02")
		}
		rank := math.Max(0.3, 1.0-(float64(i)*0.02))
		engagementBoost := math.Min(0.2, math.Log1p(float64(points))/40)
		relevance := math.Min(1.0, 0.6*rank+0.4*tokenOverlapRelevance(query, title)+engagementBoost)
		items = append(items, map[string]any{"id": objectID, "title": title, "url": stringValue(hit["url"]), "hn_url": "https://news.ycombinator.com/item?id=" + objectID, "author": stringValue(hit["author"]), "date": nullableString(date), "engagement": map[string]any{"points": points, "comments": comments}, "relevance": roundN(relevance, 2), "why_relevant": "HN story about " + truncate(title, 60)})
	}
	return items
}

func titleMatchesQuery(title, query string) bool {
	check := strings.ToLower(regexp.MustCompile(`(?i)^(Tell HN|Show HN|Ask HN|Launch HN)\s*:\s*`).ReplaceAllString(title, ""))
	for _, word := range strings.Fields(flattenQueryForAlgolia(strings.ToLower(query))) {
		if regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`).MatchString(check) {
			return true
		}
	}
	return strings.TrimSpace(query) == ""
}

func flattenQueryForAlgolia(text string) string {
	return strings.Join(strings.Fields(strings.NewReplacer(",", " ", "-", " ").Replace(text)), " ")
}

func tokenOverlapRelevance(query, text string) float64 {
	qt, tt := tokens(query), tokens(text)
	if len(qt) == 0 || len(tt) == 0 {
		return 0.5
	}
	hits := 0
	for t := range qt {
		if tt[t] {
			hits++
		}
	}
	return math.Min(1.0, float64(hits)/float64(len(qt)))
}
