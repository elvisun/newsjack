package main

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
)

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
	key, source := loadAPIKey()
	config := configFromEnv()
	xConfigured := bearerToken(config) != ""
	available := availableSources(config, []string{"news_search", "x_news", "x", "x_trends", "reddit", "hackernews"})
	warnings := doctorWarnings(rootErr, key != "", xConfigured)
	actions := doctorActions(rootErr, key != "", xConfigured)
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
			"install_mcp":    state.InstallMCP,
		},
		"auth": map[string]any{
			"medialyst_configured": key != "",
			"source":               nullableString(source),
			"x_api_configured":     xConfigured,
		},
		"mcp_bridge": map[string]any{
			"transport": "native",
			"endpoint":  medialystMCPEndpoint(),
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
	printDoctor(stdout, root, rootErr, key != "", source, xConfigured, available, warnings, actions)
	return 0
}

func printDoctor(w io.Writer, root string, rootErr error, medialystConfigured bool, medialystSource string, xConfigured bool, available []string, warnings []string, actions []map[string]string) {
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
	uiKV(w, "Medialyst", doctorAuthStatus(medialystConfigured, medialystSource))
	uiKV(w, "X API", doctorStatus(xConfigured))

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
		}
	}
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

func doctorStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "missing"
}

func doctorWarnings(rootErr error, medialystConfigured, xConfigured bool) []string {
	var warnings []string
	if rootErr != nil {
		warnings = append(warnings, rootErr.Error())
	}
	if !medialystConfigured {
		warnings = append(warnings, "Medialyst API key is not configured; live news search, media database, and journalist discovery will be unavailable. Get one at "+medialystAPIKeyURL+", then run: newsjack auth set-medialyst --key <mlst_...>.")
	}
	if !xConfigured {
		warnings = append(warnings, "X bearer token is not configured; x_news, x_trends, and X post search will be unavailable. Run: newsjack auth set-x --bearer-token <token>. This writes X_BEARER_TOKEN to ~/.newsjack/.env.")
	}
	return warnings
}

func doctorActions(rootErr error, medialystConfigured, xConfigured bool) []map[string]string {
	var actions []map[string]string
	if rootErr != nil {
		actions = append(actions, map[string]string{
			"id":      "reinstall",
			"label":   "Reinstall Newsjack",
			"command": reinstallCommand(),
		})
	}
	if !medialystConfigured {
		actions = append(actions, map[string]string{
			"id":          "configure_medialyst",
			"label":       "Configure Medialyst API (Optional)",
			"used_for":    "live news search, media database, find journalists",
			"get_key_url": medialystAPIKeyURL,
			"command":     "newsjack auth set-medialyst --key <mlst_...>",
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
