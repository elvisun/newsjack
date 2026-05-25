package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type detectorOptions struct {
	Command                         string
	Query                           []string
	Topics                          []string
	ProfilePath                     string
	Sources                         string
	MajorFeeds                      bool
	FeedURLs                        []string
	FeedFiles                       []string
	FeedOnly                        bool
	NoProfileFeeds                  bool
	NoXNews                         bool
	NoXTrends                       bool
	NoHygieneFilter                 bool
	Depth                           string
	LookbackDays                    int
	MaxAgeHours                     float64
	XNewsMinProfileMatch            float64
	XPostsMinProfileMatch           float64
	ProfileRelevanceMinProfileMatch float64
	MajorNewsMinProfileMatch        float64
	XTrendsMinProfileMatch          float64
	MinQueuePriority                float64
	MinMajorNews                    float64
	LaneCaps                        string
	NewOnly                         bool
	Limit                           int
	Mock                            bool
	Save                            bool
	Store                           string
	MonitorName                     string
	IncludeAllScored                bool
	Emit                            string
}

func cmdDetector(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{"run"}
	}
	commands := map[string]bool{"run": true, "diagnose": true, "recent": true}
	if !commands[args[0]] && args[0] != "-h" && args[0] != "--help" {
		args = append([]string{"run"}, args...)
	}
	switch args[0] {
	case "run":
		opts, code := parseDetectorRun(args[1:], stderr)
		if code != 0 {
			return code
		}
		if err := detectorRun(opts, stdout); err != nil {
			return fail(stderr, err)
		}
		return 0
	case "diagnose":
		fs := flag.NewFlagSet("detector diagnose", flag.ContinueOnError)
		fs.SetOutput(stderr)
		sources := fs.String("sources", "", "Comma-separated sources")
		store := fs.String("store", "", "SQLite store path")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		return detectorDiagnose(*sources, *store, stdout, stderr)
	case "recent":
		fs := flag.NewFlagSet("detector recent", flag.ContinueOnError)
		fs.SetOutput(stderr)
		limit := fs.Int("limit", 10, "Number of recent runs")
		store := fs.String("store", "", "SQLite store path")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		rows, err := recentRuns(*limit, *store)
		if err != nil {
			return fail(stderr, err)
		}
		writeJSON(stdout, rows)
		return 0
	default:
		return fail(stderr, errors.New("usage: newsjack detector run|diagnose|recent"))
	}
}

func parseDetectorRun(args []string, stderr io.Writer) (detectorOptions, int) {
	var topics, feedURLs, feedFiles stringList
	fs := flag.NewFlagSet("detector run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := detectorOptions{
		Command:                         "run",
		Depth:                           "quick",
		LookbackDays:                    1,
		MaxAgeHours:                     24.0,
		XNewsMinProfileMatch:            0.05,
		XPostsMinProfileMatch:           0.08,
		ProfileRelevanceMinProfileMatch: 0.05,
		MajorNewsMinProfileMatch:        0.05,
		XTrendsMinProfileMatch:          0.05,
		MinQueuePriority:                defaultMinQueuePriority,
		MinMajorNews:                    defaultMinMajorNews,
		Limit:                           20,
		Emit:                            "json",
	}
	fs.Var(&topics, "topic", "Topic to monitor. Repeatable.")
	fs.StringVar(&opts.ProfilePath, "profile", "", "Path to monitor profile JSON")
	fs.StringVar(&opts.Sources, "sources", "", "Comma-separated sources")
	fs.BoolVar(&opts.MajorFeeds, "major-feeds", false, "Include default curated major-news RSS feeds")
	fs.Var(&feedURLs, "feed-url", "RSS/Atom feed URL or local XML path. Repeatable.")
	fs.Var(&feedFiles, "feed-file", "File containing feed URLs. Repeatable.")
	fs.BoolVar(&opts.FeedOnly, "feed-only", false, "Skip query/profile searches")
	fs.BoolVar(&opts.NoProfileFeeds, "no-profile-feeds", false, "Do not include feed_urls from profile")
	fs.BoolVar(&opts.NoXNews, "no-x-news", false, "Do not auto-include x_news")
	fs.BoolVar(&opts.NoXTrends, "no-x-trends", false, "Do not include x_trends")
	fs.BoolVar(&opts.NoHygieneFilter, "no-hygiene-filter", false, "Disable deterministic hygiene filter")
	fs.StringVar(&opts.Depth, "depth", "quick", "quick, default, or deep")
	fs.IntVar(&opts.LookbackDays, "lookback-days", 1, "Lookback window in days")
	fs.Float64Var(&opts.MaxAgeHours, "max-age-hours", 24.0, "Drop items older than this many hours; 0 disables")
	fs.Float64Var(&opts.XNewsMinProfileMatch, "x-news-min-profile-match", 0.05, "Demote X News clusters below this profile-match score")
	fs.Float64Var(&opts.XPostsMinProfileMatch, "x-posts-min-profile-match", 0.08, "Demote raw X posts below this profile-match score")
	fs.Float64Var(&opts.ProfileRelevanceMinProfileMatch, "profile-relevance-min-profile-match", 0.05, "Demote profile query results below this score")
	fs.Float64Var(&opts.MajorNewsMinProfileMatch, "major-news-min-profile-match", 0.05, "Demote major-feed-only stories below this score")
	fs.Float64Var(&opts.XTrendsMinProfileMatch, "x-trends-min-profile-match", 0.05, "Demote X trends below this score")
	fs.Float64Var(&opts.MinQueuePriority, "min-queue-priority", defaultMinQueuePriority, "Mechanical queue priority floor")
	fs.Float64Var(&opts.MinMajorNews, "min-major-news", defaultMinMajorNews, "Major-news fallback floor")
	fs.StringVar(&opts.LaneCaps, "lane-caps", "", "Comma-separated per-lane output caps")
	fs.BoolVar(&opts.NewOnly, "new-only", false, "Suppress signals already seen in the monitor store")
	fs.IntVar(&opts.Limit, "limit", 20, "Maximum emitted signals")
	fs.BoolVar(&opts.Mock, "mock", false, "Use deterministic mock sources")
	fs.BoolVar(&opts.Save, "save", false, "Save run and seen URLs")
	fs.StringVar(&opts.Store, "store", "", "Override SQLite store path")
	fs.StringVar(&opts.MonitorName, "monitor-name", "", "Monitor name")
	fs.BoolVar(&opts.IncludeAllScored, "include-all-scored", false, "Include all scored signals in debug output")
	fs.StringVar(&opts.Emit, "emit", "json", "json or brief")
	valueFlags := stringSet([]string{
		"topic", "profile", "sources", "feed-url", "feed-file", "depth", "lookback-days", "max-age-hours",
		"x-news-min-profile-match", "x-posts-min-profile-match", "profile-relevance-min-profile-match",
		"major-news-min-profile-match", "x-trends-min-profile-match", "min-queue-priority", "min-major-news",
		"lane-caps", "limit", "store", "monitor-name", "emit",
	})
	if err := fs.Parse(reorderIntermixedFlags(args, valueFlags)); err != nil {
		return opts, 2
	}
	opts.Topics = topics
	opts.FeedURLs = feedURLs
	opts.FeedFiles = feedFiles
	opts.Query = fs.Args()
	if opts.Depth != "quick" && opts.Depth != "default" && opts.Depth != "deep" {
		fmt.Fprintln(stderr, "invalid --depth")
		return opts, 2
	}
	if opts.Emit != "json" && opts.Emit != "brief" {
		fmt.Fprintln(stderr, "invalid --emit")
		return opts, 2
	}
	return opts, 0
}
