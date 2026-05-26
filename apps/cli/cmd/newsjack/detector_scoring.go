package main

import (
	"crypto/sha1"
	"encoding/hex"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var stopWords = stringSet([]string{
	"about", "after", "again", "against", "also", "and", "are", "because", "am", "an", "as", "at",
	"been", "being", "but", "can", "could", "did", "does", "doing", "for", "be", "by", "do", "go", "he", "if", "in", "is", "it", "me", "my",
	"no", "of", "on", "or", "so", "to", "up", "us", "we", "from", "had", "has", "have", "her", "here", "him", "his", "how",
	"into", "its", "just", "more", "new", "not", "now", "our", "out", "over", "per", "said", "say", "says", "she", "should", "than", "that",
	"the", "their", "them", "then", "there", "these", "they", "this", "those", "through", "too", "under", "was", "what", "when", "where",
	"which", "while", "who", "why", "will", "with", "would", "you", "your",
})

var tokenRe = regexp.MustCompile(`[a-z0-9][a-z0-9+._-]{1,}`)

func tokens(text string) map[string]bool {
	out := map[string]bool{}
	for _, token := range tokenRe.FindAllString(strings.ToLower(text), -1) {
		if !stopWords[token] {
			out[token] = true
		}
	}
	return out
}

func jaccard(left, right string) float64 {
	lt, rt := tokens(left), tokens(right)
	if len(lt) == 0 || len(rt) == 0 {
		return 0
	}
	inter := 0
	for t := range lt {
		if rt[t] {
			inter++
		}
	}
	union := len(lt)
	for t := range rt {
		if !lt[t] {
			union++
		}
	}
	return float64(inter) / float64(union)
}

func profileMatches(profile monitorProfile, text string) []string {
	lower := strings.ToLower(text)
	toks := tokens(text)
	var terms []string
	terms = append(terms, profile.Topics...)
	terms = append(terms, profile.Competitors...)
	terms = append(terms, profile.Standing...)
	terms = append(terms, profile.ProofAssets...)
	var matches []string
	seen := map[string]bool{}
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term != "" && termMatches(strings.ToLower(term), lower, toks) && !seen[term] {
			matches = append(matches, term)
			seen[term] = true
			if len(matches) >= 12 {
				break
			}
		}
	}
	return nonNilStrings(matches)
}

func profileMatchScore(profile monitorProfile, text string) float64 {
	context := profile.matchText()
	if context == "" {
		return 0.4
	}
	overlap := jaccard(context, text)
	phraseBonus := math.Min(0.4, 0.08*float64(len(profileMatches(profile, text))))
	return math.Min(1.0, overlap+phraseBonus)
}

func parseTime(value string) (time.Time, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, false
	}
	if strings.HasSuffix(raw, "Z") {
		raw = strings.TrimSuffix(raw, "Z") + "+00:00"
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), true
		}
	}
	if len(raw) >= 10 {
		if parsed, err := time.Parse("2006-01-02", raw[:10]); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func minAgeHours(cluster signalCluster, now time.Time) *float64 {
	var ages []float64
	for _, item := range cluster.Evidence {
		if parsed, ok := parseTime(item.PublishedAt); ok {
			age := now.Sub(parsed).Hours()
			if age < 0 {
				age = 0
			}
			ages = append(ages, age)
		}
	}
	if len(ages) == 0 {
		return nil
	}
	sort.Float64s(ages)
	return &ages[0]
}

func freshnessScore(age *float64, lookbackDays int) float64 {
	if age == nil {
		return 0.35
	}
	if *age <= 4 {
		return 1.0
	}
	if *age <= 24 {
		return 0.86
	}
	window := math.Max(24, float64(lookbackDays*24))
	return math.Max(0.05, 1.0-(*age/window))
}

func decayBucket(age *float64) string {
	if age == nil {
		return "unknown"
	}
	switch {
	case *age <= 1:
		return "30min"
	case *age <= 4:
		return "4hr"
	case *age <= 24:
		return "24hr"
	case *age <= 168:
		return "week"
	default:
		return "month"
	}
}

func filterItemsByAge(items []evidenceItem, now time.Time, maxAgeHours float64) []evidenceItem {
	if maxAgeHours <= 0 {
		return items
	}
	var out []evidenceItem
	for _, item := range items {
		parsed, ok := parseTime(item.PublishedAt)
		if !ok || math.Max(0, now.Sub(parsed).Hours()) <= maxAgeHours {
			out = append(out, item)
		}
	}
	return out
}

var socialSources = stringSet([]string{"x", "x_news", "x_trends", "reddit", "hackernews"})
var docsHostPrefixes = []string{"docs.", "doc.", "help.", "support.", "developer.", "developers."}
var docsPathParts = stringSet([]string{"docs", "documentation", "help", "support", "kb", "knowledge-base", "api", "api-reference", "reference", "manual", "guide", "guides", "tutorial", "tutorials", "connector", "connectors", "integration", "integrations"})
var productPathParts = stringSet([]string{"product", "products", "shop", "store", "cart", "checkout", "pricing", "plans", "marketplace", "app-store", "apps"})
var seoPathPatterns = mustRegexes([]string{`\bsell[-_]house[-_]fast\b`, `\bsell[-_]my[-_]house[-_]fast\b`, `\bcash[-_]house[-_]buyers?\b`, `\bwe[-_]buy[-_]houses?\b`, `\bquick[-_]house[-_]sale\b`, `\bbest[-_][a-z0-9-]+`})
var seoTitlePatterns = mustRegexes([]string{`\bbest\s+\d+\b`, `\bbest\s+[a-z0-9\s-]+\s+(tools|services|companies|platforms)\b`, `\bhow to sell (your )?house fast\b`, `\bcash house buyers?\b`})

func hygieneRejectionReason(item evidenceItem) string {
	if socialSources[item.Source] {
		return ""
	}
	parsed, _ := url.Parse(item.URL)
	host := strings.ToLower(parsed.Hostname())
	var parts []string
	for _, part := range strings.Split(strings.ToLower(parsed.Path), "/") {
		if strings.TrimSpace(part) != "" {
			parts = append(parts, strings.TrimSpace(part))
		}
	}
	pathText := strings.ToLower(parsed.Path)
	titleText := strings.ToLower(item.Title)
	combined := strings.Join([]string{titleText, strings.ToLower(item.Container), strings.ToLower(item.Excerpt)}, " ")
	for _, prefix := range docsHostPrefixes {
		if strings.HasPrefix(host, prefix) || strings.Contains(host, "readthedocs.io") {
			return "owned_docs_or_help"
		}
	}
	for _, part := range parts {
		if docsPathParts[part] {
			return "owned_docs_or_help"
		}
	}
	if regexp.MustCompile(`\b(documentation|api reference|developer docs|help center|support article)\b`).MatchString(combined) {
		return "owned_docs_or_help"
	}
	for _, part := range parts {
		if productPathParts[part] {
			return "product_or_ecommerce_page"
		}
	}
	if regexp.MustCompile(`\b(add to cart|buy now|pricing plans|product page|shop now)\b`).MatchString(combined) {
		return "product_or_ecommerce_page"
	}
	for _, re := range seoPathPatterns {
		if re.MatchString(pathText) {
			return "seo_landing_page"
		}
	}
	for _, re := range seoTitlePatterns {
		if re.MatchString(titleText) {
			return "seo_landing_page"
		}
	}
	return ""
}

func filterItemsByHygiene(items []evidenceItem, enabled bool) ([]evidenceItem, map[string]int) {
	if !enabled {
		return items, map[string]int{}
	}
	var out []evidenceItem
	counts := map[string]int{}
	for _, item := range items {
		if reason := hygieneRejectionReason(item); reason != "" {
			counts[reason]++
			continue
		}
		out = append(out, item)
	}
	return out, counts
}

var sourceQuality = map[string]float64{
	"major_feed":  0.88,
	"news_search": 0.95,
	"x":           0.70,
	"x_news":      0.86,
	"x_trends":    0.68,
	"reddit":      0.62,
	"hackernews":  0.72,
}

func sourceAgreementScore(sources []string) float64 {
	if len(sources) >= 3 {
		return 1.0
	}
	if len(sources) == 2 {
		return 0.78
	}
	if len(sources) == 1 && sources[0] == "major_feed" {
		return 0.62
	}
	if len(sources) == 1 && sources[0] == "news_search" {
		return 0.55
	}
	return 0.32
}

func sourceQualityScore(cluster signalCluster) float64 {
	if len(cluster.Evidence) == 0 {
		return 0
	}
	sum := 0.0
	for _, item := range cluster.Evidence {
		if q, ok := sourceQuality[item.Source]; ok {
			sum += q
		} else {
			sum += 0.5
		}
	}
	return sum / float64(len(cluster.Evidence))
}

func storySizeScore(cluster signalCluster, sourceQuality, majorNews, engagement float64) map[string]any {
	type outlet struct {
		domain         string
		score          float64
		trafficScore   *float64
		authorityScore *float64
		traffic        any
		authority      any
	}
	outletsByDomain := map[string]outlet{}
	for _, item := range cluster.Evidence {
		trafficScore, trafficRaw := publicationTrafficScore(item.Metadata)
		authorityScore, authorityRaw := publicationAuthorityScore(item.Metadata)
		if trafficScore == nil && authorityScore == nil {
			continue
		}
		score := publicationOutletScore(trafficScore, authorityScore)
		domain := publicationDomainKey(item)
		current, ok := outletsByDomain[domain]
		if !ok || score > current.score {
			outletsByDomain[domain] = outlet{
				domain:         domain,
				score:          score,
				trafficScore:   trafficScore,
				authorityScore: authorityScore,
				traffic:        trafficRaw,
				authority:      authorityRaw,
			}
		}
	}

	strongest := 0.0
	spreadWeight := 0.0
	knownTraffic := 0
	knownAuthority := 0
	var topOutlets []map[string]any
	for _, o := range outletsByDomain {
		if o.score > strongest {
			strongest = o.score
		}
		spreadWeight += math.Pow(o.score, 1.5)
		if o.trafficScore != nil {
			knownTraffic++
		}
		if o.authorityScore != nil {
			knownAuthority++
		}
		topOutlets = append(topOutlets, map[string]any{
			"domain":                            o.domain,
			"outlet_score":                      roundN(o.score, 3),
			"traffic_score":                     nullableRounded(o.trafficScore),
			"domain_authority_score":            nullableRounded(o.authorityScore),
			"estimated_monthly_organic_traffic": nullableNumberAny(o.traffic),
			"domain_authority":                  nullableNumberAny(o.authority),
		})
	}
	sort.SliceStable(topOutlets, func(i, j int) bool {
		return floatValue(topOutlets[i]["outlet_score"]) > floatValue(topOutlets[j]["outlet_score"])
	})

	basis := "publication_metadata"
	confidence := "medium"
	coverageSpread := 0.0
	score := 0.0
	if len(outletsByDomain) > 0 {
		coverageSpread = 1 - math.Exp(-spreadWeight/3.0)
		score = 0.60*strongest + 0.40*coverageSpread
		if knownTraffic == 0 || knownAuthority == 0 {
			confidence = "low"
		}
	} else {
		return map[string]any{
			"score":                 nil,
			"band":                  "unknown",
			"confidence":            "low",
			"basis":                 "publication_metadata_missing",
			"known_outlet_count":    0,
			"known_traffic_count":   0,
			"known_authority_count": 0,
			"top_outlets":           []map[string]any{},
			"components": map[string]any{
				"source_quality":  roundN(sourceQuality, 3),
				"major_news":      roundN(majorNews, 3),
				"social_momentum": roundN(engagement, 3),
			},
		}
	}
	score = clamp01(score)
	return map[string]any{
		"score":                 round1(100 * score),
		"band":                  storySizeBand(score),
		"confidence":            confidence,
		"basis":                 basis,
		"known_outlet_count":    len(outletsByDomain),
		"known_traffic_count":   knownTraffic,
		"known_authority_count": knownAuthority,
		"top_outlets":           firstN(topOutlets, 5),
		"components": map[string]any{
			"strongest_outlet": roundN(strongest, 3),
			"coverage_spread":  roundN(coverageSpread, 3),
			"source_quality":   roundN(sourceQuality, 3),
			"major_news":       roundN(majorNews, 3),
			"social_momentum":  roundN(engagement, 3),
		},
	}
}

func publicationOutletScore(trafficScore, authorityScore *float64) float64 {
	switch {
	case trafficScore != nil && authorityScore != nil:
		return clamp01(0.70*(*trafficScore) + 0.30*(*authorityScore))
	case trafficScore != nil:
		return clamp01(*trafficScore)
	case authorityScore != nil:
		return clamp01(*authorityScore)
	default:
		return 0
	}
}

func publicationTrafficScore(metadata map[string]any) (*float64, any) {
	raw := firstNonNil(metadata["estimated_monthly_organic_traffic"], metadata["monthly_organic_traffic"], metadata["traffic"])
	traffic, ok := numberValue(raw)
	if !ok || traffic <= 0 {
		return nil, raw
	}
	score := clamp01(math.Log10(math.Max(1, traffic)) / 8.0)
	return &score, raw
}

func publicationAuthorityScore(metadata map[string]any) (*float64, any) {
	raw := firstNonNil(metadata["domain_authority"], metadata["domainAuthority"], metadata["domain_rating"], metadata["domainRating"])
	authority, ok := numberValue(raw)
	if !ok || authority < 0 {
		return nil, raw
	}
	score := authority
	if score > 1 {
		score = score / 100.0
	}
	score = clamp01(score)
	return &score, raw
}

func publicationDomainKey(item evidenceItem) string {
	if u, err := url.Parse(item.URL); err == nil && u.Hostname() != "" {
		return strings.ToLower(u.Hostname())
	}
	if item.Container != "" {
		return strings.ToLower(item.Container)
	}
	return strings.ToLower(item.Source)
}

func nullableRounded(value *float64) any {
	if value == nil {
		return nil
	}
	return roundN(*value, 3)
}

func nullableNumberAny(value any) any {
	if f, ok := numberValue(value); ok {
		return f
	}
	return nil
}

func storySizeBand(score float64) string {
	switch {
	case score >= 0.75:
		return "major"
	case score >= 0.50:
		return "high"
	case score >= 0.25:
		return "moderate"
	default:
		return "low"
	}
}

func clamp01(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

var engagementFields = []string{"score", "num_comments", "comments", "likes", "reposts", "replies", "quotes", "bookmarks", "views", "points"}

func engagementScore(cluster signalCluster) float64 {
	total := 0.0
	for _, item := range cluster.Evidence {
		for _, field := range engagementFields {
			value := floatValue(item.Engagement[field])
			if value > 0 {
				total += math.Log1p(value)
			}
		}
	}
	return math.Min(1.0, total/24.0)
}

func noveltyScore(urls []string, seen map[string]map[string]any) float64 {
	if len(urls) == 0 {
		return 0.50
	}
	unseen := 0
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			unseen++
		}
	}
	if unseen == 0 {
		return 0.10
	}
	return float64(unseen) / float64(len(urls))
}

var majorNewsTerms = []string{"acquire", "acquired", "acquisition", "antitrust", "ban", "billion", "breach", "contract", "deal", "funding", "hack", "investigation", "ipo", "lawsuit", "launch", "launched", "layoffs", "merger", "outage", "probe", "regulation", "regulator", "ruling", "sec", "settlement", "shutdown", "sues", "valuation"}
var majorEntityTerms = []string{"amazon", "anthropic", "apple", "doj", "ftc", "google", "meta", "microsoft", "nvidia", "openai", "pentagon", "salesforce", "sec", "spacex", "tesla", "white house", "xai"}

func majorNewsScore(cluster signalCluster, age *float64) float64 {
	var feedItems []evidenceItem
	for _, item := range cluster.Evidence {
		if item.Source == "major_feed" {
			feedItems = append(feedItems, item)
		}
	}
	if len(feedItems) == 0 {
		return 0
	}
	best := 999
	for _, item := range feedItems {
		pos := intValue(item.Metadata["feed_position"], 999)
		if pos < best {
			best = pos
		}
	}
	positionScore := 0.42
	switch {
	case best <= 3:
		positionScore = 1.0
	case best <= 10:
		positionScore = 0.82
	case best <= 25:
		positionScore = 0.64
	}
	text := strings.ToLower(cluster.text())
	toks := tokens(text)
	stakeHits := 0
	for _, term := range majorNewsTerms {
		if termMatches(term, text, toks) {
			stakeHits++
		}
	}
	entityHits := 0
	for _, term := range majorEntityTerms {
		if termMatches(term, text, toks) {
			entityHits++
		}
	}
	stakeScore := math.Min(1.0, 0.22*float64(stakeHits))
	entityScore := math.Min(1.0, 0.18*float64(entityHits))
	freshness := freshnessScore(age, 7)
	return math.Min(1.0, 0.44*positionScore+0.24*freshness+0.20*stakeScore+0.12*entityScore)
}

func termMatches(term, text string, toks map[string]bool) bool {
	if strings.Contains(term, " ") {
		return strings.Contains(text, term)
	}
	return toks[term]
}

func signalLane(cluster signalCluster, majorNews, profileMatch float64, opts detectorOptions) string {
	sources := stringSet(cluster.sources())
	if sources["x_news"] {
		if profileMatch < opts.XNewsMinProfileMatch {
			return "x_news_unmatched"
		}
		return "x_news"
	}
	if sources["x_trends"] {
		if profileMatch < opts.XTrendsMinProfileMatch {
			return "x_trends_unmatched"
		}
		return "x_trends"
	}
	if len(sources) == 1 && sources["x"] {
		for _, item := range cluster.Evidence {
			if item.Metadata["x_signal_type"] == "query_trend" {
				if profileMatch < opts.XTrendsMinProfileMatch {
					return "x_trends_unmatched"
				}
				return "x_trends"
			}
		}
		if profileMatch < opts.XPostsMinProfileMatch {
			return "x_posts_weak"
		}
		return "x_posts"
	}
	if majorNews > 0 {
		if profileMatch < opts.MajorNewsMinProfileMatch {
			return "major_news_unmatched"
		}
		return "major_news"
	}
	if profileMatch < opts.ProfileRelevanceMinProfileMatch {
		return "profile_relevance_weak"
	}
	return "profile_relevance"
}

func scoreSignal(cluster signalCluster, profile monitorProfile, seen map[string]map[string]any, now time.Time, opts detectorOptions) map[string]any {
	text := cluster.text()
	sources := cluster.sources()
	urls := cluster.urls()
	age := minAgeHours(cluster, now)
	freshness := freshnessScore(age, opts.LookbackDays)
	novelty := noveltyScore(urls, seen)
	sourceAgreement := sourceAgreementScore(sources)
	sourceQuality := sourceQualityScore(cluster)
	engagement := engagementScore(cluster)
	profileMatch := profileMatchScore(profile, text)
	majorNews := majorNewsScore(cluster, age)
	storySize := storySizeScore(cluster, sourceQuality, majorNews, engagement)
	lane := signalLane(cluster, majorNews, profileMatch, opts)
	queue := 0.0
	switch lane {
	case "major_news":
		queue = round1(100 * (0.30*majorNews + 0.22*freshness + 0.14*novelty + 0.12*profileMatch + 0.12*sourceQuality + 0.10*sourceAgreement))
	case "major_news_unmatched":
		queue = round1(math.Min(39.9, 100*(0.18*majorNews+0.16*freshness+0.12*novelty+0.10*sourceQuality+0.08*sourceAgreement)))
	case "x_news", "x_trends":
		queue = round1(100 * (0.24*freshness + 0.20*profileMatch + 0.18*novelty + 0.16*sourceQuality + 0.12*sourceAgreement + 0.10*engagement))
	case "x_trends_unmatched":
		queue = round1(math.Min(39.9, 100*(0.16*freshness+0.14*novelty+0.12*sourceQuality+0.10*engagement)))
	case "x_news_unmatched", "profile_relevance_weak", "x_posts_weak":
		queue = round1(math.Min(39.9, 100*(0.18*freshness+0.16*novelty+0.12*sourceQuality+0.10*sourceAgreement+0.08*engagement)))
	case "x_posts":
		queue = round1(math.Min(64.0, 100*(0.22*freshness+0.20*engagement+0.18*profileMatch+0.16*novelty+0.14*sourceQuality+0.10*sourceAgreement)))
	default:
		queue = round1(100 * (0.24*freshness + 0.20*sourceAgreement + 0.18*novelty + 0.16*profileMatch + 0.12*sourceQuality + 0.10*engagement))
	}
	var evidence []map[string]any
	for i, item := range cluster.Evidence {
		if i >= 8 {
			break
		}
		evidence = append(evidence, item.publicDict())
	}
	features := map[string]any{
		"decay_bucket":    decayBucket(age),
		"source_count":    len(sources),
		"evidence_count":  len(cluster.Evidence),
		"seen_before":     len(urls) > 0 && allSeen(urls, seen),
		"seen_urls":       seenSubset(urls, seen),
		"profile_matches": profileMatches(profile, text),
		"safety_flags":    safetyFlags(text, profile.Exclusions),
	}
	if age != nil {
		features["age_hours"] = roundN(*age, 2)
	}
	mechanicalScores := map[string]any{
		"freshness":        roundN(freshness, 3),
		"source_agreement": roundN(sourceAgreement, 3),
		"novelty":          roundN(novelty, 3),
		"profile_match":    roundN(profileMatch, 3),
		"source_quality":   roundN(sourceQuality, 3),
		"momentum":         roundN(engagement, 3),
		"major_news":       roundN(majorNews, 3),
	}
	if score, ok := numberValue(storySize["score"]); ok {
		mechanicalScores["story_size"] = roundN(score/100.0, 3)
	}
	return map[string]any{
		"id":         signalID(cluster.title(), urls, text),
		"title":      cluster.title(),
		"sources":    sources,
		"evidence":   evidence,
		"features":   features,
		"story_size": storySize,
		"routing": map[string]any{
			"lane":           lane,
			"queue_priority": queue,
			"demoted":        strings.HasSuffix(lane, "_unmatched") || strings.HasSuffix(lane, "_weak"),
		},
		"mechanical_scores": mechanicalScores,
	}
}

func signalID(title string, urls []string, text string) string {
	basisParts := append([]string{title}, firstN(urls, 5)...)
	basis := strings.Join(basisParts, "|")
	if strings.Trim(basis, "|") == "" {
		basis = truncate(text, 400)
	}
	sum := sha1.Sum([]byte(basis))
	return hex.EncodeToString(sum[:])[:16]
}

func allSeen(urls []string, seen map[string]map[string]any) bool {
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			return false
		}
	}
	return true
}

func seenSubset(urls []string, seen map[string]map[string]any) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, u := range urls {
		if v, ok := seen[u]; ok {
			out[u] = v
		}
	}
	return out
}

var hardSafetyTerms = []string{"humanitarian crisis", "sexual violence", "missing child", "missing person", "mass shooting", "terror attack", "child abuse", "hate crime", "war crime", "earthquake", "genocide", "hostage", "assault", "bombing", "murder", "abuse", "rape"}

func safetyFlags(text string, exclusions []string) []map[string]string {
	lower := strings.ToLower(text)
	var flags []map[string]string
	for _, term := range hardSafetyTerms {
		if strings.Contains(lower, term) {
			flags = append(flags, map[string]string{"type": "hard_safety_term", "term": term, "note": "Review against tragedy and human-suffering newsjacking rules."})
		}
	}
	for _, term := range exclusions {
		if term != "" && strings.Contains(lower, strings.ToLower(term)) {
			flags = append(flags, map[string]string{"type": "profile_exclusion", "term": term, "note": "Matched a monitor-profile exclusion."})
		}
	}
	if flags == nil {
		return []map[string]string{}
	}
	return flags
}

func clusterItems(items []evidenceItem) []signalCluster {
	sort.SliceStable(items, func(i, j int) bool {
		ki, kj := 1, 1
		if items[i].Source == "news_search" {
			ki = 0
		}
		if items[j].Source == "news_search" {
			kj = 0
		}
		if ki != kj {
			return ki < kj
		}
		return items[i].PublishedAt < items[j].PublishedAt
	})
	var clusters []signalCluster
	for _, item := range items {
		if item.Title == "" && item.Excerpt == "" {
			continue
		}
		placed := false
		for i := range clusters {
			if item.URL != "" && contains(clusters[i].urls(), item.URL) {
				clusters[i].Evidence = append(clusters[i].Evidence, item)
				placed = true
				break
			}
			if jaccard(item.text(), clusters[i].text()) >= 0.32 {
				clusters[i].Evidence = append(clusters[i].Evidence, item)
				placed = true
				break
			}
		}
		if !placed {
			clusters = append(clusters, signalCluster{Evidence: []evidenceItem{item}})
		}
	}
	return clusters
}
