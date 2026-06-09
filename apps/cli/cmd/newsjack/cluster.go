package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type clusterOptions struct {
	TitleOverlap float64 // overlap-coefficient threshold on significant title tokens
	MinShared    int     // minimum shared significant title tokens to merge
	DropStale    bool
	WindowHours  float64
	StaleMaxBand string // highest story-size band still eligible for stale pre-gate drop
}

// cmdCluster collapses same-story signals into clusters before the expensive
// story-origin retrieval pass runs. Only one representative per cluster is
// carried forward in "signals"; the other pickups are recorded in
// "clustered_duplicates" so retrieval and freshness gating run once per story,
// not once per syndicated copy. With --drop-stale it also deterministically
// pre-gates low-value, clearly-old signals so they never reach retrieval.
func cmdCluster(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("cluster", flag.ContinueOnError)
	fs.SetOutput(stderr)
	candidatesPath := fs.String("candidates", "", "Relevant candidates JSON")
	outputPath := fs.String("output", "", "Output path")
	titleOverlap := fs.Float64("title-overlap", 0.6, "Overlap-coefficient threshold on significant title tokens to treat two signals as the same story")
	minShared := fs.Int("min-shared-tokens", 2, "Minimum shared significant title tokens required to merge two signals")
	dropStale := fs.Bool("drop-stale", false, "Deterministically pre-gate low-story-size signals whose detector decay is well outside the window")
	windowHours := fs.Float64("window-hours", 24.0, "Freshness window in hours (used only to label the stale pre-gate)")
	staleMaxBand := fs.String("stale-max-band", "moderate", "Highest story-size band eligible for the stale pre-gate drop (low|moderate|high|major); larger stories always get researched")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *candidatesPath == "" {
		return fail(stderr, errors.New("--candidates is required"))
	}
	candidates, err := readJSONMap(*candidatesPath)
	if err != nil {
		return fail(stderr, err)
	}
	output := clusterCandidates(candidates, clusterOptions{
		TitleOverlap: *titleOverlap,
		MinShared:    *minShared,
		DropStale:    *dropStale,
		WindowHours:  *windowHours,
		StaleMaxBand: strings.ToLower(strings.TrimSpace(*staleMaxBand)),
	})
	data := marshalJSON(output)
	if *outputPath != "" {
		if err := os.WriteFile(expandPath(*outputPath), data, 0o644); err != nil {
			return fail(stderr, err)
		}
	} else {
		stdout.Write(data)
	}
	return 0
}

var storySizeBandRank = map[string]int{"unknown": 0, "low": 1, "moderate": 2, "high": 3, "major": 4}

// staleDecayBuckets are detector decay buckets that are unambiguously older than
// a 24h window. We never auto-drop "unknown" — only verifiably old buckets.
var staleDecayBuckets = map[string]bool{"week": true, "month": true}

type clusterGroup struct {
	signals   []map[string]any
	urlKeys   map[string]bool
	seedTitle map[string]bool // significant title tokens of the seed signal
}

func clusterCandidates(candidates map[string]any, opts clusterOptions) map[string]any {
	if opts.MinShared <= 0 {
		opts.MinShared = 2
	}
	signals := signalSlice(candidates["signals"])
	// Stable input order so clustering is deterministic regardless of map iteration.
	sort.SliceStable(signals, func(i, j int) bool {
		return stringValue(signals[i]["id"]) < stringValue(signals[j]["id"])
	})

	var groups []*clusterGroup
	for _, signal := range signals {
		urlKeys := signalURLKeys(signal)
		titleToks := significantTitleTokens(stringValue(signal["title"]))
		placed := false
		for _, g := range groups {
			oc, inter := overlapCoefficient(titleToks, g.seedTitle)
			if shareURLKey(urlKeys, g.urlKeys) || (oc >= opts.TitleOverlap && inter >= opts.MinShared) {
				g.signals = append(g.signals, signal)
				for k := range urlKeys {
					g.urlKeys[k] = true
				}
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, &clusterGroup{
				signals:   []map[string]any{signal},
				urlKeys:   cloneStringSet(urlKeys),
				seedTitle: titleToks,
			})
		}
	}

	representatives := []map[string]any{}
	duplicates := []map[string]any{}
	preGated := []map[string]any{}
	bandCounts := map[string]int{}
	for idx, g := range groups {
		rep := pickRepresentative(g.signals)
		clusterID := stringValue(rep["id"])
		memberIDs := make([]string, 0, len(g.signals))
		for _, s := range g.signals {
			if id := stringValue(s["id"]); id != "" && id != clusterID {
				memberIDs = append(memberIDs, id)
			}
		}
		sort.Strings(memberIDs)
		clusterMeta := map[string]any{
			"cluster_id":      clusterID,
			"cluster_index":   idx,
			"cluster_size":    len(g.signals),
			"role":            "representative",
			"member_count":    len(memberIDs),
			"member_ids":      memberIDs,
			"duplicate_count": len(memberIDs),
		}
		band := signalStorySizeBandValue(rep)
		bandCounts[band]++
		if opts.DropStale && staleEligible(rep, band, opts.StaleMaxBand) {
			s := summarySignal(rep)
			s["cluster"] = clusterMeta
			s["pre_gate_reason"] = "stale_low_value"
			s["pre_gate_rationale"] = fmt.Sprintf("decay_bucket=%s is outside the %gh window and story_size band=%s is at or below the stale-max-band; skipped expensive story-origin retrieval", signalDecayBucket(rep), opts.WindowHours, band)
			preGated = append(preGated, s)
		} else {
			withCluster := cloneMap(rep)
			withCluster["cluster"] = clusterMeta
			representatives = append(representatives, withCluster)
		}
		for _, s := range g.signals {
			if stringValue(s["id"]) == clusterID {
				continue
			}
			dup := summarySignal(s)
			dup["cluster_id"] = clusterID
			dup["representative_id"] = clusterID
			duplicates = append(duplicates, dup)
		}
	}

	return map[string]any{
		"version":      1,
		"generated_at": time.Now().UTC().Format(time.RFC3339Nano),
		"monitor":      valueOrEmptyMap(candidates["monitor"]),
		"signals":      representatives,
		"clustering": map[string]any{
			"input_signal_count":   len(signals),
			"cluster_count":        len(groups),
			"representative_count": len(representatives),
			"duplicate_count":      len(duplicates),
			"pre_gated_count":      len(preGated),
			"title_overlap":        opts.TitleOverlap,
			"min_shared_tokens":    opts.MinShared,
			"drop_stale":           opts.DropStale,
			"stale_max_band":       opts.StaleMaxBand,
			"story_size_bands":     sortedCountMap(bandCounts),
		},
		"clustered_duplicates": duplicates,
		"pre_gated_stale":      preGated,
		"coarse_relevance":     candidates["coarse_relevance"],
		"detector_diagnostics": valueOrEmptyMap(candidates["detector_diagnostics"]),
		"source_errors":        valueOrEmptyMap(candidates["source_errors"]),
	}
}

// pickRepresentative chooses the signal that should carry the story forward:
// highest queue priority, then largest story size, then lowest id for stability.
func pickRepresentative(signals []map[string]any) map[string]any {
	best := signals[0]
	bestQ := queuePriority(best)
	bestS, _ := signalStorySizeScore(best)
	for _, s := range signals[1:] {
		q := queuePriority(s)
		size, _ := signalStorySizeScore(s)
		switch {
		case q > bestQ:
			best, bestQ, bestS = s, q, size
		case q == bestQ && size > bestS:
			best, bestQ, bestS = s, q, size
		case q == bestQ && size == bestS && stringValue(s["id"]) < stringValue(best["id"]):
			best = s
		}
	}
	return best
}

func staleEligible(signal map[string]any, band, staleMaxBand string) bool {
	if !staleDecayBuckets[signalDecayBucket(signal)] {
		return false
	}
	maxRank, ok := storySizeBandRank[staleMaxBand]
	if !ok {
		maxRank = storySizeBandRank["moderate"]
	}
	return storySizeBandRank[band] <= maxRank
}

func signalDecayBucket(signal map[string]any) string {
	return strings.ToLower(stringValue(valueOrEmptyMap(signal["features"])["decay_bucket"]))
}

func signalStorySizeBandValue(signal map[string]any) string {
	if band := strings.ToLower(stringValue(valueOrEmptyMap(signal["story_size"])["band"])); band != "" {
		if band == "unknown" {
			if hintBand := strings.ToLower(stringValue(valueOrEmptyMap(valueOrEmptyMap(signal["story_size"])["attention_hint"])["band"])); hintBand != "" {
				return hintBand
			}
		}
		return band
	}
	if score, ok := signalStorySizeScore(signal); ok {
		return storySizeBand(score)
	}
	return "unknown"
}

func signalURLKeys(signal map[string]any) map[string]bool {
	out := map[string]bool{}
	for _, raw := range anySlice(signal["evidence"]) {
		if m, ok := raw.(map[string]any); ok {
			if k := normalizedURLKey(stringValue(m["url"])); k != "" {
				out[k] = true
			}
		}
	}
	return out
}

// clusterStopwords are common headline filler that should not count as a shared
// "anchor" between two same-story headlines.
var clusterStopwords = stringSet(strings.Fields("the a an of to in on for and or with as at by from is are be new this that its has have will into over after amid out get said says how why what can also more their our your you not but its it's about up down off launches launch unveils announces reveals brings makes adds gets sets first amid"))

// significantTitleTokens returns the distinctive lowercase tokens of a title:
// length >= 3, not a stopword. Cross-outlet headlines about the same event share
// their key nouns (entities, products, actions) even when the rest differs.
func significantTitleTokens(title string) map[string]bool {
	out := map[string]bool{}
	for tok := range tokens(title) {
		if len(tok) >= 3 && !clusterStopwords[tok] {
			out[tok] = true
		}
	}
	return out
}

func shareURLKey(a, b map[string]bool) bool {
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}

// overlapCoefficient returns |a∩b| / min(|a|,|b|) and the intersection size.
// Unlike Jaccard it is not diluted by one headline carrying many extra words,
// so a short shared core of entities still scores high.
func overlapCoefficient(a, b map[string]bool) (float64, int) {
	if len(a) == 0 || len(b) == 0 {
		return 0, 0
	}
	inter := 0
	for t := range a {
		if b[t] {
			inter++
		}
	}
	min := len(a)
	if len(b) < min {
		min = len(b)
	}
	return float64(inter) / float64(min), inter
}

func cloneStringSet(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for k := range in {
		out[k] = true
	}
	return out
}
