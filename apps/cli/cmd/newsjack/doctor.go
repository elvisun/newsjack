package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	medialystCredentialNotConfigured = "not_configured"
	medialystCredentialOK            = "ok"
	medialystCredentialInvalid       = "invalid"
	medialystCredentialUnreachable   = "unreachable"
)

type medialystCredentialCheck struct {
	State  string
	Detail string
}

var doctorMedialystCredentialCheck = checkMedialystCredential

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func cmdDoctor(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "Emit doctor status as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	root, rootErr := newsjackRoot()
	state := readInstallStateOrDefault()
	medialystStatus := loadMedialystAuthStatus()
	medialystCheck := medialystCredentialCheck{State: medialystCredentialNotConfigured}
	if medialystStatus.Configured {
		medialystCheck = doctorMedialystCredentialCheck()
	}
	config := configFromEnv()
	xConfigured := bearerToken(config) != ""
	typesafeKey, typesafeSource := loadTypeSafeAPIKey()
	available := availableSources(config, []string{"news_search", "x_news", "x", "x_trends", "reddit", "hackernews"})
	warnings := doctorWarnings(rootErr, medialystCheck, xConfigured)
	actions := doctorActions(rootErr, medialystCheck, xConfigured)
	cli := newsjackCLIInvocation()
	payload := map[string]any{
		"version":       version,
		"newsjack_home": newsjackHome(),
		"newsjack_root": root,
		"root_ok":       rootErr == nil,
		"install": map[string]any{
			"distribution":   distributionName(),
			"cli_command":    cli.CommandLine(),
			"cli_installed":  npmDistribution() || fileExists(installedBinaryPath()),
			"cli_version":    version,
			"bundle_version": installedVersion(),
			"skills_mode":    state.SkillsMode,
			"repo":           state.Repo,
			"runtimes":       state.Runtimes,
			"runtimes_raw":   state.RuntimesRaw,
		},
		"auth": map[string]any{
			"medialyst_configured":         medialystStatus.Configured,
			"medialyst_status":             medialystCheck.State,
			"medialyst_status_detail":      nullableString(medialystCheck.Detail),
			"medialyst_oauth_configured":   medialystStatus.OAuthConfigured,
			"medialyst_api_key_configured": medialystStatus.APIKeyConfigured,
			"source":                       nullableString(medialystStatus.Source),
			"auth_type":                    nullableString(medialystStatus.Kind),
			"x_api_configured":             xConfigured,
			"typesafe_configured":          typesafeKey != "",
			"typesafe_source":              nullableString(typesafeSource),
		},
		"sources": map[string]any{
			"news_search": contains(available, "news_search"),
			"x_news":      contains(available, "x_news"),
			"x":           contains(available, "x"),
			"x_trends":    contains(available, "x_trends"),
			"reddit":      contains(available, "reddit"),
			"hackernews":  contains(available, "hackernews"),
		},
		"warnings": warnings,
		"actions":  actions,
	}
	payload["runtimes"] = runtimeStatus()
	if *jsonOut {
		writeJSON(stdout, payload)
		return 0
	}
	printDoctor(stdout, root, rootErr, medialystCheck, medialystStatus.Source, xConfigured, typesafeKey != "", typesafeSource, available, warnings, actions)
	return 0
}

func printDoctor(w io.Writer, root string, rootErr error, medialystCheck medialystCredentialCheck, medialystSource string, xConfigured bool, typesafeConfigured bool, typesafeSource string, available []string, warnings []string, actions []map[string]string) {
	uiProduct(w, "doctor", "system health check")
	fmt.Fprintln(w)
	uiSection(w, "paths")
	uiKV(w, "distribution", distributionName())
	uiKV(w, "home", newsjackHome())
	if rootErr != nil {
		uiKV(w, "install root", "missing: "+rootErr.Error())
	} else {
		uiKV(w, "install root", root)
	}

	fmt.Fprintln(w)
	uiSection(w, "auth")
	uiKV(w, "Medialyst", doctorMedialystAuthStatus(medialystCheck, medialystSource))
	uiKV(w, "X API", doctorStatus(xConfigured))
	uiKV(w, "TypeSafe (Jev)", doctorOptionalStatus(typesafeConfigured, typesafeSource, "newsjack auth set-typesafe --key <key>"))

	fmt.Fprintln(w)
	uiSection(w, "sources")
	for _, source := range []string{"news_search", "x_news", "x", "x_trends", "reddit", "hackernews"} {
		uiKV(w, source, doctorStatus(contains(available, source)))
	}

	fmt.Fprintln(w)
	uiSection(w, "runtimes")
	for _, rt := range runtimeTargets {
		status := doctorStatus(runtimeDetected(rt))
		uiKV(w, runtimeLabel(rt.Key), status+"  "+targetDir(rt))
	}

	fmt.Fprintln(w)
	if len(warnings) == 0 {
		uiSuccess(w, "doctor found no actionable problems.")
		return
	}
	uiSection(w, "warnings")
	for _, warning := range warnings {
		uiWarn(w, "%s", warning)
	}
	if len(actions) > 0 {
		fmt.Fprintln(w)
		uiSection(w, "next actions")
		for _, action := range actions {
			if label := action["label"]; label != "" {
				uiKV(w, label, action["command"])
			}
			if usedFor := action["used_for"]; usedFor != "" {
				uiKV(w, "used for", usedFor)
			}
			if getKey := action["get_key_url"]; getKey != "" {
				uiKV(w, "get key", getKey)
			}
			if fallback := action["fallback"]; fallback != "" {
				uiKV(w, "fallback", fallback)
			}
		}
	}
}

func checkMedialystCredential() medialystCredentialCheck {
	_, err := medialystAPIRequest(http.MethodGet, "/v1/credits/balance", nil, nil, nil, 10*time.Second)
	if err == nil {
		return medialystCredentialCheck{State: medialystCredentialOK}
	}
	var apiErr *medialystAPIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusUnauthorized:
			return medialystCredentialCheck{State: medialystCredentialInvalid, Detail: "authentication was rejected"}
		case http.StatusForbidden:
			return medialystCredentialCheck{State: medialystCredentialOK, Detail: "credential is valid but lacks the balance scope"}
		}
	}
	if isOAuthErrorCode(err, "invalid_grant") {
		return medialystCredentialCheck{State: medialystCredentialInvalid, Detail: "OAuth login is expired or revoked"}
	}
	var refreshTransportErr *oauthRefreshTransportError
	if errors.As(err, &refreshTransportErr) {
		return medialystCredentialCheck{State: medialystCredentialUnreachable, Detail: err.Error()}
	}
	if storedOAuthRequiresLogin() {
		return medialystCredentialCheck{State: medialystCredentialInvalid, Detail: "OAuth login has no usable refresh token"}
	}
	return medialystCredentialCheck{State: medialystCredentialUnreachable, Detail: err.Error()}
}

func storedOAuthRequiresLogin() bool {
	if key, _ := loadAPIKey(); key != "" {
		return false
	}
	creds, ok, err := readCredentialsFile()
	if err != nil || !ok || creds.Medialyst.OAuth == nil {
		return false
	}
	oauth := creds.Medialyst.OAuth
	return strings.TrimSpace(oauth.AccessToken) != "" &&
		strings.TrimSpace(oauth.RefreshToken) == "" &&
		oauthTokenExpired(oauth.ExpiresAt)
}

func doctorMedialystAuthStatus(check medialystCredentialCheck, source string) string {
	status := check.State
	if status == medialystCredentialNotConfigured {
		return "missing"
	}
	if source != "" {
		status += "  " + source
	}
	return status
}

func doctorAuthStatus(configured bool, source string) string {
	if !configured {
		return "missing"
	}
	if source == "" {
		return "ok"
	}
	return "ok  " + source
}

// doctorOptionalStatus renders an optional integration: missing is a hint,
// not a warning, because the pipeline works without it.
func doctorOptionalStatus(configured bool, source, command string) string {
	if configured {
		return doctorAuthStatus(true, source)
	}
	return "not configured (optional)  " + command
}

func doctorStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "missing"
}

func doctorWarnings(rootErr error, medialystCheck medialystCredentialCheck, xConfigured bool) []string {
	var warnings []string
	if rootErr != nil {
		warnings = append(warnings, rootErr.Error())
	}
	switch medialystCheck.State {
	case medialystCredentialNotConfigured:
		warnings = append(warnings, "Medialyst is not connected; live news search, journalist enrichment, and media list research will be unavailable. Run: newsjack login. For API-key automation, use: newsjack auth set-medialyst --key <mlst_...>.")
	case medialystCredentialInvalid:
		warnings = append(warnings, "Medialyst credentials are invalid. Run: newsjack login. For unattended runs, use: newsjack auth set-medialyst --key <mlst_...>.")
	case medialystCredentialUnreachable:
		warnings = append(warnings, "Medialyst could not be reached, so the saved credential was not marked invalid. Check the network and run newsjack doctor again.")
	}
	if !xConfigured {
		warnings = append(warnings, "X bearer token is not configured; x_news, x_trends, and X post search will be unavailable. Run: newsjack auth set-x --bearer-token <token>. This writes X_BEARER_TOKEN to ~/.newsjack/.env.")
	}
	return warnings
}

func doctorActions(rootErr error, medialystCheck medialystCredentialCheck, xConfigured bool) []map[string]string {
	var actions []map[string]string
	if rootErr != nil {
		actions = append(actions, map[string]string{
			"id":      "reinstall",
			"label":   "Reinstall Newsjack",
			"command": reinstallCommand(),
		})
	}
	if medialystCheck.State == medialystCredentialNotConfigured || medialystCheck.State == medialystCredentialInvalid {
		label := "Connect Medialyst (Optional)"
		if medialystCheck.State == medialystCredentialInvalid {
			label = "Reconnect Medialyst"
		}
		actions = append(actions, map[string]string{
			"id":          "configure_medialyst",
			"label":       label,
			"used_for":    "live news search, journalist enrichment, and media list research",
			"get_key_url": medialystAPIKeyURL,
			"command":     "newsjack login",
			"fallback":    "newsjack auth set-medialyst --key <mlst_...>",
		})
	}
	if !xConfigured {
		actions = append(actions, map[string]string{
			"id":          "configure_x",
			"label":       "Configure X API (Optional)",
			"used_for":    "X News, X trends, and X post search",
			"get_key_url": xAPIKeyURL,
			"command":     "newsjack auth set-x --bearer-token <token>",
			"writes":      "~/.newsjack/.env:X_BEARER_TOKEN",
		})
	}
	return actions
}

func reinstallCommand() string {
	if npmDistribution() {
		return "npm i -g newsjack"
	}
	if goos() == "windows" {
		return "newsjack setup"
	}
	return "curl -fsSL https://newsjack.sh | bash"
}

func runtimeStatus() map[string]any {
	out := map[string]any{}
	for _, rt := range runtimeTargets {
		out[rt.Key] = map[string]any{
			"detected":   runtimeDetected(rt),
			"skills_dir": targetDir(rt),
		}
	}
	return out
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
