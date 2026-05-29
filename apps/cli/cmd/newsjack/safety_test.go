package main

import "testing"

func TestSafetyFlagsTragedyTerms(t *testing.T) {
	flagged := []string{
		"Crews in Dallas search for missing people after massive apartment fire kills 3",
		"Second death confirmed after blast in Washington state, and no survivors expected",
		"School shooting leaves several dead",
		"Three killed in plane crash",
		"Manslaughter charge filed after fatal stabbing",
	}
	for _, title := range flagged {
		if got := safetyFlags(title, nil); len(got) == 0 {
			t.Errorf("expected a hard safety flag for %q, got none", title)
		}
	}

	clean := []string{
		"Connecticut Law Bans Location Data Sales, Targets Data Brokers",
		"UiPath Drives Financial Growth and Stock Gains via Agentic AI",
		"This award-winning Vancouver coffee roaster treats beans like fine wine",
		"AI knowledge base startup raises Series A",
	}
	for _, title := range clean {
		if got := safetyFlags(title, nil); len(got) != 0 {
			t.Errorf("did not expect a safety flag for %q, got %v", title, got)
		}
	}
}

func TestSafetyFlagsWordBoundaries(t *testing.T) {
	// "manslaughter" must not match the substring "laughter"; word boundaries
	// keep the literal-violence terms from firing on innocuous words.
	if got := safetyFlags("The room erupted in laughter", nil); len(got) != 0 {
		t.Errorf("substring 'laughter' should not flag, got %v", got)
	}
	if got := safetyFlags("Assassin's Creed laughter reel", nil); len(got) != 0 {
		t.Errorf("unexpected flag, got %v", got)
	}
}

func TestSafetyFlagsExclusions(t *testing.T) {
	flags := safetyFlags("A normal headline about Acme", []string{"Acme"})
	found := false
	for _, f := range flags {
		if f["type"] == "profile_exclusion" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a profile_exclusion flag, got %v", flags)
	}
}
