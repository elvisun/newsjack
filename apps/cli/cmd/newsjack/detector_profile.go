package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type monitorProfile struct {
	Company      string
	Description  string
	Topics       []string
	Competitors  []string
	SearchTerms  []string
	FeedURLs     []string
	XNews        map[string]any
	XTrends      map[string]any
	Spokespeople []string
	Standing     []string
	Exclusions   []string
	Raw          map[string]any
}

func defaultProfile() monitorProfile {
	return monitorProfile{
		Topics:       []string{},
		Competitors:  []string{},
		SearchTerms:  []string{},
		FeedURLs:     []string{},
		Spokespeople: []string{},
		Standing:     []string{},
		Exclusions:   []string{},
		XNews:        map[string]any{"enabled": true},
		XTrends:      map[string]any{"mode": "none", "woeids": []any{}, "locations": []any{}},
	}
}

func profileFromFile(path string) (monitorProfile, error) {
	data, err := os.ReadFile(expandPath(path))
	if err != nil {
		return monitorProfile{}, err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return monitorProfile{}, err
	}
	return profileFromMap(payload), nil
}

func profileFromMap(payload map[string]any) monitorProfile {
	p := defaultProfile()
	p.Raw = payload
	p.Company = firstString(payload["company"], payload["client"])
	p.Description = stringValue(payload["description"])
	p.Topics = stringListValue(firstValue(payload, "topics", "beats"))
	p.Competitors = stringListValue(payload["competitors"])
	p.SearchTerms = stringListValue(firstValue(payload, "search_terms", "queries"))
	p.FeedURLs = stringListValue(firstValue(payload, "feed_urls", "feeds", "rss_feeds"))
	p.XNews = dictValue(payload["x_news"], map[string]any{"enabled": true})
	p.XTrends = dictValue(payload["x_trends"], map[string]any{"mode": "none", "woeids": []any{}, "locations": []any{}})
	p.Spokespeople = stringListValue(firstValue(payload, "spokespeople", "experts"))
	p.Standing = stringListValue(firstValue(payload, "standing", "expertise"))
	p.Exclusions = stringListValue(firstValue(payload, "exclusions", "do_not_newsjack"))
	return p
}

func (p monitorProfile) queryTerms() []string {
	if len(p.SearchTerms) > 0 {
		return dedupeStrings(p.SearchTerms)
	}
	var out []string
	out = append(out, p.Topics...)
	for _, term := range p.Competitors {
		out = append(out, competitorQuery(term))
	}
	return dedupeStrings(out)
}

func (p monitorProfile) matchText() string {
	var parts []string
	parts = append(parts, p.Company, p.Description)
	parts = append(parts, p.Topics...)
	parts = append(parts, p.Competitors...)
	parts = append(parts, p.FeedURLs...)
	parts = append(parts, p.Spokespeople...)
	parts = append(parts, p.Standing...)
	return strings.TrimSpace(strings.Join(parts, " "))
}

func (p monitorProfile) publicDict() map[string]any {
	return map[string]any{
		"company":      nullableString(p.Company),
		"topics":       p.Topics,
		"competitors":  p.Competitors,
		"search_terms": p.SearchTerms,
		"feed_urls":    p.FeedURLs,
		"x_news":       p.XNews,
		"x_trends":     p.XTrends,
		"spokespeople": p.Spokespeople,
		"standing":     p.Standing,
		"exclusions":   p.Exclusions,
	}
}

func firstValue(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func firstString(values ...any) string {
	for _, v := range values {
		if s := stringValue(v); s != "" {
			return s
		}
	}
	return ""
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	return s
}

func dictValue(v any, def map[string]any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		out := map[string]any{}
		for k, v := range m {
			out[k] = v
		}
		return out
	}
	out := map[string]any{}
	for k, v := range def {
		out[k] = v
	}
	return out
}

func stringListValue(v any) []string {
	if v == nil {
		return emptyStrings()
	}
	switch x := v.(type) {
	case string:
		parts := regexp.MustCompile(`[,;\n]`).Split(x, -1)
		var out []string
		for _, part := range parts {
			if s := strings.TrimSpace(part); s != "" {
				out = append(out, s)
			}
		}
		return nonNilStrings(out)
	case []any:
		var out []string
		for _, item := range x {
			if m, ok := item.(map[string]any); ok {
				keys := make([]string, 0, len(m))
				for k := range m {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					if s := stringValue(m[k]); s != "" {
						out = append(out, s)
					}
				}
				continue
			}
			if s := stringValue(item); s != "" {
				out = append(out, s)
			}
		}
		return nonNilStrings(out)
	case []string:
		var out []string
		for _, item := range x {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}
		return nonNilStrings(out)
	default:
		if s := stringValue(v); s != "" {
			return []string{s}
		}
		return emptyStrings()
	}
}

func competitorQuery(value string) string {
	term := strings.TrimSpace(value)
	if term == "" || strings.HasPrefix(term, `"`) || !strings.Contains(term, " ") {
		return term
	}
	return `"` + strings.ReplaceAll(term, `"`, `\"`) + `"`
}
