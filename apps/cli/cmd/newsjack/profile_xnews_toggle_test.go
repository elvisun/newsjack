package main

import "testing"

func TestRequestedSourcesHonorsProfileXNewsDisabled(t *testing.T) {
	profile := monitorProfile{XNews: map[string]any{"enabled": false}}
	sources, err := requestedSourcesFor(detectorOptions{}, profile)
	if err != nil {
		t.Fatal(err)
	}
	if contains(sources, "x_news") {
		t.Fatalf("profile disables x_news but default sources still include it: %v", sources)
	}
	if !contains(sources, "news_search") {
		t.Fatalf("expected news_search to stay in default sources: %v", sources)
	}
}

func TestRequestedSourcesExplicitXNewsOverridesProfile(t *testing.T) {
	profile := monitorProfile{XNews: map[string]any{"enabled": false}}
	sources, err := requestedSourcesFor(detectorOptions{Sources: "news_search,x_news"}, profile)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(sources, "x_news") {
		t.Fatalf("an explicit --sources x_news should win over the profile: %v", sources)
	}
}

func TestRequestedSourcesDefaultsKeepXNewsWhenProfileSilent(t *testing.T) {
	sources, err := requestedSourcesFor(detectorOptions{}, monitorProfile{XNews: map[string]any{"enabled": true}})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(sources, "x_news") {
		t.Fatalf("x_news should stay on by default: %v", sources)
	}
}

func TestRequestedSourcesHonorsProfileXDisabled(t *testing.T) {
	profile := profileFromMap(map[string]any{
		"x_news": map[string]any{"enabled": false},
		"x":      map[string]any{"enabled": false},
	})
	sources, err := requestedSourcesFor(detectorOptions{}, profile)
	if err != nil {
		t.Fatal(err)
	}
	if contains(sources, "x") || contains(sources, "x_news") {
		t.Fatalf("profile disables both X sources but they are still requested: %v", sources)
	}
	if !contains(sources, "news_search") {
		t.Fatalf("expected news_search to stay: %v", sources)
	}
}
