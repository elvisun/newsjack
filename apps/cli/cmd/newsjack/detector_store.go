package main

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func dbPathFromEnv(override string) string {
	if override != "" {
		return expandPath(override)
	}
	if v := os.Getenv("NEWSJACK_STORE"); v != "" {
		return expandPath(v)
	}
	return filepath.Join(homeDir(), ".local", "share", "newsjack", "monitor.db")
}

func initDB(override string) error {
	path := dbPathFromEnv(override)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA busy_timeout=5000;
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
CREATE TABLE IF NOT EXISTS seen_urls (url TEXT PRIMARY KEY, first_seen TEXT NOT NULL, last_seen TEXT NOT NULL, sighting_count INTEGER NOT NULL DEFAULT 1);
CREATE TABLE IF NOT EXISTS monitor_runs (id INTEGER PRIMARY KEY, monitor_name TEXT, profile_json TEXT, query_json TEXT NOT NULL, generated_at TEXT NOT NULL, signal_count INTEGER NOT NULL DEFAULT 0, created_at TEXT DEFAULT (datetime('now')));
CREATE TABLE IF NOT EXISTS signal_snapshots (id INTEGER PRIMARY KEY, run_id INTEGER REFERENCES monitor_runs(id) ON DELETE CASCADE, signal_id TEXT NOT NULL, title TEXT NOT NULL, rank_score REAL NOT NULL, payload_json TEXT NOT NULL, created_at TEXT DEFAULT (datetime('now')));
CREATE INDEX IF NOT EXISTS idx_signal_snapshots_run ON signal_snapshots(run_id);
CREATE INDEX IF NOT EXISTS idx_signal_snapshots_rank ON signal_snapshots(rank_score DESC);`)
	return err
}

func openDB(override string) (*sql.DB, error) {
	if err := initDB(override); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPathFromEnv(override))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func seenStatus(urls []string, override string) (map[string]map[string]any, error) {
	if len(urls) == 0 {
		return map[string]map[string]any{}, nil
	}
	db, err := openDB(override)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	unique := dedupeStrings(urls)
	placeholders := strings.TrimRight(strings.Repeat("?,", len(unique)), ",")
	args := make([]any, len(unique))
	for i, u := range unique {
		args[i] = u
	}
	rows, err := db.Query("SELECT url, first_seen, last_seen, sighting_count FROM seen_urls WHERE url IN ("+placeholders+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]map[string]any{}
	for rows.Next() {
		var u, first, last string
		var count int
		if err := rows.Scan(&u, &first, &last, &count); err != nil {
			return nil, err
		}
		out[u] = map[string]any{"first_seen": first, "last_seen": last, "sighting_count": count}
	}
	return out, rows.Err()
}

func recordRun(monitorName string, profile map[string]any, queries []string, signals []map[string]any, seenURLs []string, override string) (int64, error) {
	db, err := openDB(override)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	profileJSON, _ := json.Marshal(profile)
	queryJSON, _ := json.Marshal(queries)
	result, err := tx.Exec("INSERT INTO monitor_runs (monitor_name, profile_json, query_json, generated_at, signal_count) VALUES (?, ?, ?, ?, ?)", nullSQLString(monitorName), string(profileJSON), string(queryJSON), now, len(signals))
	if err != nil {
		return 0, err
	}
	runID, _ := result.LastInsertId()
	for _, signal := range signals {
		payload, _ := json.Marshal(signal)
		if _, err := tx.Exec("INSERT INTO signal_snapshots (run_id, signal_id, title, rank_score, payload_json) VALUES (?, ?, ?, ?, ?)", runID, signal["id"], signal["title"], queuePriority(signal), string(payload)); err != nil {
			return 0, err
		}
	}
	for _, u := range dedupeStrings(seenURLs) {
		if _, err := tx.Exec(`INSERT INTO seen_urls (url, first_seen, last_seen, sighting_count) VALUES (?, ?, ?, 1)
ON CONFLICT(url) DO UPDATE SET last_seen = excluded.last_seen, sighting_count = seen_urls.sighting_count + 1`, u, now, now); err != nil {
			return 0, err
		}
	}
	return runID, tx.Commit()
}

func recentRuns(limit int, override string) ([]map[string]any, error) {
	db, err := openDB(override)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT id, monitor_name, query_json, generated_at, signal_count FROM monitor_runs ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, count int
		var name sql.NullString
		var queryJSON, generated string
		if err := rows.Scan(&id, &name, &queryJSON, &generated, &count); err != nil {
			return nil, err
		}
		var queries []string
		_ = json.Unmarshal([]byte(queryJSON), &queries)
		out = append(out, map[string]any{"id": id, "monitor_name": nullableSQLString(name), "queries": queries, "generated_at": generated, "signal_count": count})
	}
	return out, rows.Err()
}

func storePathForOutput(override string, saved bool) any {
	if !saved {
		return nil
	}
	return dbPathFromEnv(override)
}

func nullSQLString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullableSQLString(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}
