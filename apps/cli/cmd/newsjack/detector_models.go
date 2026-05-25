package main

import (
	"sort"
	"strings"
)

type evidenceItem struct {
	Source      string
	Title       string
	URL         string
	Excerpt     string
	Author      string
	Container   string
	PublishedAt string
	Engagement  map[string]any
	Metadata    map[string]any
}

func evidenceFromMap(m map[string]any) evidenceItem {
	return evidenceItem{
		Source:      firstString(m["source"], "unknown"),
		Title:       strings.TrimSpace(firstString(m["title"], m["excerpt"])),
		URL:         strings.TrimSpace(stringValue(m["url"])),
		Excerpt:     strings.TrimSpace(stringValue(m["excerpt"])),
		Author:      stringValue(m["author"]),
		Container:   stringValue(m["container"]),
		PublishedAt: stringValue(m["published_at"]),
		Engagement:  dictValue(m["engagement"], map[string]any{}),
		Metadata:    dictValue(m["metadata"], map[string]any{}),
	}
}

func (e evidenceItem) text() string {
	return strings.Join(nonEmpty(e.Title, e.Excerpt, e.Author, e.Container), " ")
}

func (e evidenceItem) publicDict() map[string]any {
	payload := map[string]any{
		"source":       e.Source,
		"title":        e.Title,
		"url":          e.URL,
		"author":       nullableString(e.Author),
		"container":    nullableString(e.Container),
		"published_at": nullableString(e.PublishedAt),
		"excerpt":      truncate(e.Excerpt, 500),
		"engagement":   e.Engagement,
	}
	if len(e.Metadata) > 0 {
		payload["metadata"] = e.Metadata
	}
	return payload
}

type signalCluster struct {
	Evidence []evidenceItem
}

func (c signalCluster) title() string {
	for _, item := range c.Evidence {
		if item.Source == "news_search" && item.Title != "" {
			return item.Title
		}
	}
	if len(c.Evidence) > 0 {
		return c.Evidence[0].Title
	}
	return ""
}

func (c signalCluster) text() string {
	var parts []string
	for _, item := range c.Evidence {
		parts = append(parts, item.text())
	}
	return strings.Join(parts, " ")
}

func (c signalCluster) sources() []string {
	set := map[string]bool{}
	for _, item := range c.Evidence {
		set[item.Source] = true
	}
	var out []string
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (c signalCluster) urls() []string {
	var out []string
	for _, item := range c.Evidence {
		if item.URL != "" {
			out = append(out, item.URL)
		}
	}
	return out
}
