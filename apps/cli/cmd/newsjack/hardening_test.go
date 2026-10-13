package main

import (
	"path/filepath"
	"testing"
)

func TestDateHelpersTolerateShortInput(t *testing.T) {
	// These previously sliced date[:10] and panicked on short strings.
	for _, in := range []string{"", "2026", "2026-08", "bad", "   "} {
		if got := dateToUnix(in); got != 0 {
			t.Errorf("dateToUnix(%q) = %d, want 0", in, got)
		}
		if got := lookbackHours(in, "2026-08-18"); got != 168 {
			t.Errorf("lookbackHours(%q, ...) = %d, want fallback 168", in, got)
		}
		if got := tbsForRange("2026-08-01", in); got != "qdr:w" {
			t.Errorf("tbsForRange(..., %q) = %q, want fallback qdr:w", in, got)
		}
	}
	if got := datePrefix("2026-08-18T12:34:56Z"); got != "2026-08-18" {
		t.Fatalf("datePrefix = %q", got)
	}
	if got := lookbackHours("2026-08-11", "2026-08-18T09:00:00Z"); got != 8*24 {
		t.Fatalf("lookbackHours = %d, want %d", got, 8*24)
	}
	if got := dateToUnix("2026-08-18T00:00:00Z"); got == 0 {
		t.Fatal("dateToUnix should parse a datetime prefix")
	}
}

func TestAutoUpdateForcedParsing(t *testing.T) {
	cases := map[string]bool{"": false, "0": false, "false": false, "1": true, "true": true, "ON": true, "yes": true}
	for value, want := range cases {
		withTempEnv(t, map[string]string{"NEWSJACK_AUTO_UPDATE": value}, func() {
			if got := autoUpdateForced(); got != want {
				t.Errorf("NEWSJACK_AUTO_UPDATE=%q forced=%v want %v", value, got, want)
			}
		})
	}
}

func TestShouldAutoUpdateSkipsNonInteractiveRuns(t *testing.T) {
	// go test never has a TTY on stdin, so implicit auto-update must be off here
	// regardless of the other preconditions.
	withTempEnv(t, map[string]string{"NEWSJACK_AUTO_UPDATE": "", "NEWSJACK_NO_AUTO_UPDATE": ""}, func() {
		if stdinIsTerminal() {
			t.Skip("stdin is a terminal; cannot exercise the non-interactive gate")
		}
		if shouldAutoUpdate([]string{"detector"}) {
			t.Fatal("non-interactive invocation must not auto-update")
		}
	})
}

func TestDetectorStoreRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "monitor.db")
	if err := initDB(dbPath); err != nil {
		t.Fatalf("initDB: %v", err)
	}
	signals := []map[string]any{{"id": "s1", "title": "Fogo de Chão opens in Costa Rica", "url": "https://example.com/a"}}
	seen := []string{"https://example.com/a", "https://example.com/b", "https://example.com/a"}
	runID, err := recordRun("adega", map[string]any{"topics": []string{"churrascaria"}}, []string{"q1"}, signals, seen, dbPath)
	if err != nil {
		t.Fatalf("recordRun: %v", err)
	}
	if runID <= 0 {
		t.Fatalf("unexpected run id %d", runID)
	}
	status, err := seenStatus([]string{"https://example.com/a", "https://example.com/zzz"}, dbPath)
	if err != nil {
		t.Fatalf("seenStatus: %v", err)
	}
	if _, ok := status["https://example.com/a"]; !ok {
		t.Fatal("recorded URL should be reported as seen")
	}
	if _, ok := status["https://example.com/zzz"]; ok {
		t.Fatal("unknown URL should not be reported as seen")
	}
	if got := status["https://example.com/a"]["sighting_count"]; got != 1 {
		t.Fatalf("duplicate seen URLs within one run must count once, got %v", got)
	}
	// Second run bumps the sighting count.
	if _, err := recordRun("adega", nil, nil, nil, []string{"https://example.com/a"}, dbPath); err != nil {
		t.Fatalf("recordRun #2: %v", err)
	}
	status, _ = seenStatus([]string{"https://example.com/a"}, dbPath)
	if got := status["https://example.com/a"]["sighting_count"]; got != 2 {
		t.Fatalf("sighting_count after second run = %v, want 2", got)
	}
	runs, err := recentRuns(5, dbPath)
	if err != nil {
		t.Fatalf("recentRuns: %v", err)
	}
	if len(runs) != 2 || runs[0]["signal_count"] != 0 || runs[1]["signal_count"] != 1 {
		t.Fatalf("unexpected recentRuns: %+v", runs)
	}
}
