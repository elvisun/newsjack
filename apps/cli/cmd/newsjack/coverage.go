package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func cmdCoverage(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printCoverageHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return cmdCoverageList(args[1:], stdout, stderr)
	case "init":
		return cmdCoverageInit(args[1:], stdout, stderr)
	case "status":
		return cmdCoverageStatus(args[1:], stdout, stderr)
	case "open":
		return cmdCoverageOpen(args[1:], stdout, stderr)
	case "check":
		return cmdCoverageCheck(args[1:], stdout, stderr)
	case "record":
		return cmdCoverageRecord(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown coverage command: %s", args[0])
	}
}

func cmdCoverageList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("coverage list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		return fail(stderr, errors.New("usage: newsjack coverage list"))
	}
	writeJSON(stdout, coverageList())
	return 0
}

func cmdCoverageInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("coverage init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "Coverage tracker JSON to save")
	force := fs.Bool("force", false, "Overwrite an existing coverage tracker config")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"config"}))); err != nil {
		return 2
	}
	if *configPath == "" {
		return fail(stderr, errors.New("usage: newsjack coverage init [slug] --config tracker.json"))
	}
	payload, err := readJSONMap(*configPath)
	if err != nil {
		return fail(stderr, err)
	}
	slug := ""
	if fs.NArg() > 0 {
		slug = fs.Arg(0)
	}
	if slug == "" {
		slug = slugify(firstString(payload["name"], "coverage"))
	}
	slug = slugify(slug)
	if slug == "" {
		return fail(stderr, errors.New("coverage slug is empty"))
	}
	if err := validateCoverageConfig(payload); err != nil {
		return fail(stderr, err)
	}
	dir := coverageDir(slug)
	dest := coverageConfigPath(slug)
	if fileExists(dest) && !*force {
		return failf(stderr, "coverage tracker already exists: %s", slug)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(dest, marshalJSON(payload), 0o644); err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, map[string]any{
		"slug":          slug,
		"coverage_dir":  dir,
		"config_path":   dest,
		"keyword_count": len(coverageKeywords(payload)),
	})
	return 0
}

func cmdCoverageStatus(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack coverage status <slug>"))
	}
	payload := coverageStatus(args[0])
	writeJSON(stdout, payload)
	if !truthy(payload["exists"], false) {
		return 1
	}
	return 0
}

func cmdCoverageOpen(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack coverage open <slug>"))
	}
	status := coverageStatus(args[0])
	if !truthy(status["exists"], false) {
		return failf(stderr, "coverage tracker not found: %s", args[0])
	}
	if latest := stringValue(status["latest_run_dir"]); latest != "" {
		fmt.Fprintln(stdout, latest)
		return 0
	}
	fmt.Fprintln(stdout, coverageDir(args[0]))
	return 0
}

func cmdCoverageCheck(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("coverage check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("input", "", "JSON file of candidate coverage items")
	store := fs.String("store", "", "Override SQLite store path")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"input", "store"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *input == "" {
		return fail(stderr, errors.New("usage: newsjack coverage check <slug> --input candidates.json"))
	}
	slug := slugify(fs.Arg(0))
	if !fileExists(coverageConfigPath(slug)) {
		return failf(stderr, "coverage tracker not found: %s", slug)
	}
	items, err := readCoverageDecisionItems(*input)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := checkCoverageItems(slug, items, *store)
	if err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, result)
	return 0
}

func cmdCoverageRecord(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("coverage record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("input", "", "JSON file of LLM-classified coverage items")
	runDir := fs.String("run-dir", "", "Run directory for provenance")
	store := fs.String("store", "", "Override SQLite store path")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"input", "run-dir", "store"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *input == "" {
		return fail(stderr, errors.New("usage: newsjack coverage record <slug> --input decisions.json"))
	}
	slug := slugify(fs.Arg(0))
	if !fileExists(coverageConfigPath(slug)) {
		return failf(stderr, "coverage tracker not found: %s", slug)
	}
	items, err := readCoverageDecisionItems(*input)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := recordCoverageDecisions(slug, items, *runDir, *store)
	if err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, result)
	return 0
}

func validateCoverageConfig(payload map[string]any) error {
	if len(coverageKeywords(payload)) == 0 {
		return errors.New("coverage tracker config must include at least one keyword")
	}
	for i, kw := range coverageKeywords(payload) {
		if strings.TrimSpace(stringValue(kw["keyword"])) == "" {
			return fmt.Errorf("coverage keyword %d is missing keyword", i+1)
		}
		if strings.TrimSpace(stringValue(kw["means"])) == "" {
			return fmt.Errorf("coverage keyword %d is missing means", i+1)
		}
	}
	return nil
}

func coverageKeywords(payload map[string]any) []map[string]any {
	var out []map[string]any
	for _, raw := range anySlice(payload["keywords"]) {
		switch v := raw.(type) {
		case map[string]any:
			out = append(out, v)
		case string:
			if strings.TrimSpace(v) != "" {
				out = append(out, map[string]any{"keyword": strings.TrimSpace(v), "means": strings.TrimSpace(v)})
			}
		}
	}
	return out
}

func readCoverageDecisionItems(path string) ([]map[string]any, error) {
	payload, err := readJSONAny(path)
	if err != nil {
		return nil, err
	}
	switch v := payload.(type) {
	case []any:
		return mapSlice(v), nil
	case map[string]any:
		for _, key := range []string{"items", "decisions", "articles"} {
			if raw, ok := v[key]; ok {
				return mapSlice(raw), nil
			}
		}
		return nil, fmt.Errorf("%s must contain items, decisions, or articles", path)
	default:
		return nil, fmt.Errorf("%s must contain coverage decision JSON", path)
	}
}

func checkCoverageItems(slug string, items []map[string]any, override string) (map[string]any, error) {
	db, err := openDB(override)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	knownItems := []map[string]any{}
	unknownItems := []map[string]any{}
	for _, item := range items {
		annotated := cloneMap(item)
		normalizedURL := normalizeCoverageURL(firstString(item["canonical_url"], item["url"], item["link"]))
		articleID := coverageArticleID(normalizedURL, item)
		annotated["article_id"] = articleID
		annotated["canonical_url"] = normalizedURL
		decision, err := latestCoverageDecision(db, slug, articleID)
		if err != nil {
			return nil, err
		}
		if decision == nil {
			annotated["known"] = false
			unknownItems = append(unknownItems, annotated)
			continue
		}
		annotated["known"] = true
		annotated["prior_decision"] = decision
		if alert, err := coverageAlertRecord(db, slug, articleID); err != nil {
			return nil, err
		} else if alert != nil {
			annotated["prior_alert"] = alert
		}
		knownItems = append(knownItems, annotated)
	}
	return map[string]any{
		"slug":          slug,
		"store_path":    dbPathFromEnv(override),
		"checked_items": len(items),
		"known_count":   len(knownItems),
		"unknown_count": len(unknownItems),
		"known_items":   knownItems,
		"unknown_items": unknownItems,
	}, nil
}

func latestCoverageDecision(db *sql.DB, slug, articleID string) (map[string]any, error) {
	var verdict string
	var alert int
	var confidence, rationale, runDir sql.NullString
	var decidedAt string
	err := db.QueryRow(`SELECT verdict, confidence, alert, rationale, run_dir, decided_at
FROM coverage_decisions
WHERE tracker_slug = ? AND article_id = ?
ORDER BY decided_at DESC, id DESC
LIMIT 1`, slug, articleID).Scan(&verdict, &confidence, &alert, &rationale, &runDir, &decidedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"verdict":    verdict,
		"confidence": nullableSQLString(confidence),
		"alert":      alert != 0,
		"rationale":  nullableSQLString(rationale),
		"run_dir":    nullableSQLString(runDir),
		"decided_at": decidedAt,
	}, nil
}

func coverageAlertRecord(db *sql.DB, slug, articleID string) (map[string]any, error) {
	var keyword sql.NullString
	var firstAlertedAt, lastAlertedAt string
	var alertCount int
	err := db.QueryRow(`SELECT keyword, first_alerted_at, last_alerted_at, alert_count
FROM coverage_alerts
WHERE tracker_slug = ? AND article_id = ?`, slug, articleID).Scan(&keyword, &firstAlertedAt, &lastAlertedAt, &alertCount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"keyword":          nullableSQLString(keyword),
		"first_alerted_at": firstAlertedAt,
		"last_alerted_at":  lastAlertedAt,
		"alert_count":      alertCount,
	}, nil
}

func recordCoverageDecisions(slug string, items []map[string]any, runDir, override string) (map[string]any, error) {
	db, err := openDB(override)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	newAlerts := []map[string]any{}
	articleIDs := map[string]bool{}
	alertable := 0
	for _, item := range items {
		normalizedURL := normalizeCoverageURL(firstString(item["canonical_url"], item["url"], item["link"]))
		articleID := coverageArticleID(normalizedURL, item)
		articleIDs[articleID] = true
		payloadJSON, _ := json.Marshal(item)
		title := firstString(item["title"], item["headline"])
		outlet := firstString(item["outlet"], item["source"], item["container"], item["publication"])
		author := firstString(item["author"], item["byline"])
		published := firstString(item["published_at"], item["publishedAt"], item["date"])
		if _, err := tx.Exec(`INSERT INTO coverage_articles (id, canonical_url, title, outlet, author, published_at, first_seen, last_seen, sighting_count, payload_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?)
ON CONFLICT(id) DO UPDATE SET last_seen = excluded.last_seen, sighting_count = coverage_articles.sighting_count + 1, title = excluded.title, outlet = excluded.outlet, author = excluded.author, published_at = excluded.published_at, payload_json = excluded.payload_json`,
			articleID, normalizedURL, nullableString(title), nullableString(outlet), nullableString(author), nullableString(published), now, now, string(payloadJSON)); err != nil {
			return nil, err
		}
		keyword := firstString(item["keyword"], item["keyword_id"])
		verdict := firstString(item["verdict"], "unknown")
		confidence := firstString(item["confidence"], item["certainty"])
		rationale := firstString(item["rationale"], item["reason"], item["why"])
		alert := coverageItemAlerts(item)
		if alert {
			alertable++
		}
		if _, err := tx.Exec(`INSERT INTO coverage_decisions (tracker_slug, article_id, keyword, verdict, confidence, alert, rationale, run_dir, decided_at, payload_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			slug, articleID, keyword, verdict, nullableString(confidence), boolInt(alert), nullableString(rationale), nullableString(expandPath(runDir)), now, string(payloadJSON)); err != nil {
			return nil, err
		}
		if !alert {
			continue
		}
		inserted, err := insertCoverageAlert(tx, slug, articleID, keyword, now)
		if err != nil {
			return nil, err
		}
		if inserted {
			itemCopy := cloneMap(item)
			itemCopy["article_id"] = articleID
			itemCopy["canonical_url"] = normalizedURL
			itemCopy["new_alert"] = true
			newAlerts = append(newAlerts, itemCopy)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{
		"slug":            slug,
		"store_path":      dbPathFromEnv(override),
		"recorded_items":  len(items),
		"unique_articles": len(articleIDs),
		"alertable_items": alertable,
		"new_alerts":      newAlerts,
		"new_alert_count": len(newAlerts),
	}, nil
}

func insertCoverageAlert(tx *sql.Tx, slug, articleID, keyword, now string) (bool, error) {
	var existing string
	err := tx.QueryRow("SELECT first_alerted_at FROM coverage_alerts WHERE tracker_slug = ? AND article_id = ?", slug, articleID).Scan(&existing)
	if err == nil {
		_, updateErr := tx.Exec("UPDATE coverage_alerts SET last_alerted_at = ?, alert_count = alert_count + 1 WHERE tracker_slug = ? AND article_id = ?", now, slug, articleID)
		return false, updateErr
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	_, err = tx.Exec("INSERT INTO coverage_alerts (tracker_slug, article_id, keyword, first_alerted_at, last_alerted_at, alert_count) VALUES (?, ?, ?, ?, ?, 1)", slug, articleID, keyword, now, now)
	return true, err
}

func coverageItemAlerts(item map[string]any) bool {
	if v, ok := item["alert"]; ok {
		return truthy(v, false)
	}
	verdict := strings.ToLower(strings.TrimSpace(firstString(item["verdict"], item["classification"])))
	return verdict == "real_feature" || verdict == "feature"
}

func normalizeCoverageURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return raw
	}
	parsed.Fragment = ""
	q := parsed.Query()
	for key := range q {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" || lower == "mc_cid" || lower == "mc_eid" {
			q.Del(key)
		}
	}
	parsed.RawQuery = q.Encode()
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.EscapedPath(), "/")
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String()
}

func coverageArticleID(normalizedURL string, item map[string]any) string {
	basis := normalizedURL
	if basis == "" {
		basis = strings.ToLower(strings.Trim(strings.Join([]string{
			firstString(item["title"], item["headline"]),
			firstString(item["outlet"], item["source"], item["container"], item["publication"]),
			firstString(item["published_at"], item["publishedAt"], item["date"]),
		}, "|"), "|"))
	}
	if basis == "" {
		data, _ := json.Marshal(item)
		basis = string(data)
	}
	sum := sha1.Sum([]byte(basis))
	return hex.EncodeToString(sum[:])
}

func coverageStatus(slug string) map[string]any {
	slug = slugify(slug)
	dir := coverageDir(slug)
	configPath := coverageConfigPath(slug)
	runs := coverageRuns(slug)
	latestRun := ""
	if len(runs) > 0 {
		latestRun = runs[len(runs)-1]
	}
	keywordCount := 0
	name := ""
	if fileExists(configPath) {
		if payload, err := readJSONMap(configPath); err == nil {
			keywordCount = len(coverageKeywords(payload))
			name = stringValue(payload["name"])
		}
	}
	articleCount, alertCount := coverageStoredCounts(slug, "")
	return map[string]any{
		"slug":            slug,
		"name":            nullableString(name),
		"exists":          fileExists(configPath),
		"coverage_dir":    dir,
		"config_path":     configPath,
		"keyword_count":   keywordCount,
		"run_count":       len(runs),
		"latest_run_dir":  nullableString(latestRun),
		"store_path":      dbPathFromEnv(""),
		"stored_articles": articleCount,
		"stored_alerts":   alertCount,
	}
}

func coverageList() []map[string]any {
	entries, err := os.ReadDir(coverageHomeDir())
	if err != nil {
		return []map[string]any{}
	}
	var out []map[string]any
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := slugify(entry.Name())
		if !fileExists(coverageConfigPath(slug)) {
			continue
		}
		out = append(out, coverageStatus(slug))
	}
	sort.Slice(out, func(i, j int) bool {
		return stringValue(out[i]["slug"]) < stringValue(out[j]["slug"])
	})
	return out
}

func coverageStoredCounts(slug, override string) (int, int) {
	db, err := openDB(override)
	if err != nil {
		return 0, 0
	}
	defer db.Close()
	var articles, alerts int
	_ = db.QueryRow("SELECT COUNT(DISTINCT article_id) FROM coverage_decisions WHERE tracker_slug = ?", slug).Scan(&articles)
	_ = db.QueryRow("SELECT COUNT(*) FROM coverage_alerts WHERE tracker_slug = ?", slug).Scan(&alerts)
	return articles, alerts
}

func coverageRuns(slug string) []string {
	root := filepath.Join(coverageDir(slug), "runs")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var runs []string
	for _, entry := range entries {
		if entry.IsDir() {
			runs = append(runs, filepath.Join(root, entry.Name()))
		}
	}
	sort.Strings(runs)
	return runs
}

func coverageHomeDir() string { return filepath.Join(newsjackHome(), "coverage") }

func coverageDir(slug string) string { return filepath.Join(coverageHomeDir(), slugify(slug)) }

func coverageConfigPath(slug string) string { return filepath.Join(coverageDir(slug), "tracker.json") }

func coverageRunDir(slug string) string {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	dir := filepath.Join(coverageDir(slug), "runs", stamp)
	if !dirExists(dir) {
		return dir
	}
	return filepath.Join(coverageDir(slug), "runs", time.Now().UTC().Format("20060102T150405.000000000Z"))
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
