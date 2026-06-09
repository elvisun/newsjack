package main

const (
	storyRecallHighMinStorySize        = 0.45
	storyRecallHighMinProfileMatch     = 0.01
	storyRecallModerateMinStorySize    = 0.22
	storyRecallModerateMinProfileMatch = 0.025
	storyRecallHighQueueFloor          = 46.0
	storyRecallModerateQueueFloor      = 42.0
)

func storyRecallQueueFloor(storySize float64, hasStorySize bool, profileMatch float64) (float64, bool) {
	if !hasStorySize {
		return 0, false
	}
	switch {
	case storySize >= storyRecallHighMinStorySize && profileMatch >= storyRecallHighMinProfileMatch:
		return storyRecallHighQueueFloor, true
	case storySize >= storyRecallModerateMinStorySize && profileMatch >= storyRecallModerateMinProfileMatch:
		return storyRecallModerateQueueFloor, true
	default:
		return 0, false
	}
}

func storyRecallSelectionPass(signal map[string]any) bool {
	storySize, ok := signalStorySizeScore(signal)
	if !ok {
		return false
	}
	_, ok = storyRecallQueueFloor(storySize, true, signalProfileMatchForRecall(signal))
	return ok
}

func signalStorySizeScore(signal map[string]any) (float64, bool) {
	if score, ok := numberValue(valueOrEmptyMap(signal["mechanical_scores"])["story_size"]); ok {
		return normalizeStorySizeScore(score)
	}
	if score, ok := numberValue(valueOrEmptyMap(signal["story_size"])["score"]); ok {
		return normalizeStorySizeScore(score)
	}
	if score, ok := numberValue(valueOrEmptyMap(valueOrEmptyMap(signal["story_size"])["attention_hint"])["score"]); ok {
		return normalizeStorySizeScore(score)
	}
	return 0, false
}

func normalizeStorySizeScore(score float64) (float64, bool) {
	if score < 0 {
		return 0, false
	}
	if score > 1 {
		score = score / 100.0
	}
	return score, true
}

func signalProfileMatchForRecall(signal map[string]any) float64 {
	score := floatValue(valueOrEmptyMap(signal["mechanical_scores"])["profile_match"])
	if len(signalProfileMatches(signal)) > 0 && score < storyRecallModerateMinProfileMatch {
		return storyRecallModerateMinProfileMatch
	}
	return score
}

func signalProfileMatches(signal map[string]any) []any {
	if matches := valueOrEmptyArray(signal["profile_matches"]); len(matches) > 0 {
		return matches
	}
	return valueOrEmptyArray(valueOrEmptyMap(signal["features"])["profile_matches"])
}
