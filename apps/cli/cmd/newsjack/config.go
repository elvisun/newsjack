package main

import ()

var version = "v0.1.0-dev"

const (
	defaultRepo             = "elvisun/newsjack"
	defaultRef              = "main"
	defaultInstallURL       = "https://newsjack.sh"
	defaultMinQueuePriority = 40.0
	defaultMinMajorNews     = 0.55
	envMedialystKey         = "MEDIALYST_API_KEY"
	envXBearerToken         = "X_BEARER_TOKEN"
	envTypeSafeKey          = "TYPESAFE_API_KEY"
	envTypeSafeBaseURL      = "NEWSJACK_TYPESAFE_BASE_URL"
)

var xBearerEnvKeys = []string{"TWITTER_BEARER_TOKEN", envXBearerToken, "X_API_BEARER_TOKEN", "TWITTER_API_BEARER_TOKEN"}
