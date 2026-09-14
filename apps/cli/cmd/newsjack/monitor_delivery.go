package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	slackNotifyPitchReady = "pitch_ready"
	slackNotifyEveryRun   = "every_run"
	maxSlackMessageRunes  = 40000
	maxSlackWebhookBytes  = 8192
)

type monitorDeliveryConfig struct {
	Slack *slackDeliveryConfig `json:"slack,omitempty"`
}

type slackDeliveryConfig struct {
	WebhookURL   string `json:"webhook_url"`
	NotifyOn     string `json:"notify_on"`
	ConfiguredAt string `json:"configured_at"`
}

type slackDeliveryReceipt struct {
	RunID         string `json:"run_id"`
	MessageSHA256 string `json:"message_sha256"`
	SentAt        string `json:"sent_at"`
}

var slackDeliveryHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func cmdMonitorDelivery(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, errors.New("usage: newsjack monitor delivery set-slack|status|test|send|remove-slack"))
	}
	switch args[0] {
	case "set-slack":
		return cmdMonitorDeliverySetSlack(args[1:], stdin, stdout, stderr)
	case "status":
		return cmdMonitorDeliveryStatus(args[1:], stdout, stderr)
	case "test":
		return cmdMonitorDeliveryTest(args[1:], stdout, stderr)
	case "send":
		return cmdMonitorDeliverySend(args[1:], stdout, stderr)
	case "remove-slack":
		return cmdMonitorDeliveryRemoveSlack(args[1:], stdout, stderr)
	default:
		return failf(stderr, "unknown monitor delivery command: %s", args[0])
	}
}

func cmdMonitorDeliverySetSlack(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor delivery set-slack", flag.ContinueOnError)
	fs.SetOutput(stderr)
	notifyOnRaw := fs.String("notify-on", "every-run", "Notification policy: every-run or pitch-ready")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"notify-on"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor delivery set-slack <slug> [--notify-on every-run|pitch-ready]"))
	}
	slug := slugify(fs.Arg(0))
	if err := requireMonitor(slug); err != nil {
		return fail(stderr, err)
	}
	notifyOn, err := normalizeSlackNotifyOn(*notifyOnRaw)
	if err != nil {
		return fail(stderr, err)
	}
	webhookURL, err := readSlackWebhookURL(stdin, stderr)
	if err != nil {
		return fail(stderr, err)
	}
	if err := validateSlackWebhookURL(webhookURL); err != nil {
		return fail(stderr, err)
	}
	config := monitorDeliveryConfig{Slack: &slackDeliveryConfig{
		WebhookURL:   webhookURL,
		NotifyOn:     notifyOn,
		ConfiguredAt: time.Now().UTC().Format(time.RFC3339Nano),
	}}
	path, err := writeMonitorDeliveryConfig(slug, config)
	if err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, map[string]any{
		"slug":             slug,
		"slack_configured": true,
		"notify_on":        notifyOn,
		"credentials_path": path,
	})
	return 0
}

func cmdMonitorDeliveryStatus(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor delivery status <slug>"))
	}
	slug := slugify(args[0])
	if err := requireMonitor(slug); err != nil {
		return fail(stderr, err)
	}
	config, configured, err := readMonitorDeliveryConfig(slug)
	if err != nil {
		return fail(stderr, err)
	}
	payload := map[string]any{
		"slug":             slug,
		"slack_configured": configured,
		"notify_on":        nil,
	}
	if configured {
		payload["notify_on"] = config.Slack.NotifyOn
	}
	writeJSON(stdout, payload)
	return 0
}

func cmdMonitorDeliveryTest(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor delivery test <slug>"))
	}
	slug := slugify(args[0])
	config, err := configuredSlackDelivery(slug)
	if err != nil {
		return fail(stderr, err)
	}
	message := fmt.Sprintf("Newsjack Slack delivery is connected for `%s`.", slug)
	if err := postSlackWebhook(config.WebhookURL, message); err != nil {
		return fail(stderr, err)
	}
	writeJSON(stdout, map[string]any{
		"slug":   slug,
		"tested": true,
		"sent":   true,
	})
	return 0
}

func cmdMonitorDeliverySend(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("monitor delivery send", flag.ContinueOnError)
	fs.SetOutput(stderr)
	messagePathRaw := fs.String("message-file", "", "Path to the agent-rendered Slack message")
	runIDRaw := fs.String("run-id", "", "Stable run identifier used to prevent duplicate delivery")
	force := fs.Bool("force", false, "Send again even when this run already has a receipt")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"message-file", "run-id"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 || strings.TrimSpace(*messagePathRaw) == "" || strings.TrimSpace(*runIDRaw) == "" {
		return fail(stderr, errors.New("usage: newsjack monitor delivery send <slug> --message-file <path> --run-id <id> [--force]"))
	}
	slug := slugify(fs.Arg(0))
	config, err := configuredSlackDelivery(slug)
	if err != nil {
		return fail(stderr, err)
	}
	runID := strings.TrimSpace(*runIDRaw)
	receiptPath := monitorSlackReceiptPath(slug, runID)
	if fileExists(receiptPath) && !*force {
		writeJSON(stdout, map[string]any{
			"slug":      slug,
			"run_id":    runID,
			"sent":      false,
			"duplicate": true,
		})
		return 0
	}
	messagePath := expandPath(*messagePathRaw)
	messageFile, err := os.Open(messagePath)
	if err != nil {
		return fail(stderr, err)
	}
	defer messageFile.Close()
	messageBytes, err := io.ReadAll(io.LimitReader(messageFile, maxSlackMessageRunes*utf8.UTFMax+1))
	if err != nil {
		return fail(stderr, err)
	}
	if len(messageBytes) > maxSlackMessageRunes*utf8.UTFMax {
		return failf(stderr, "Slack message exceeds %d characters", maxSlackMessageRunes)
	}
	message := string(messageBytes)
	if strings.TrimSpace(message) == "" {
		return fail(stderr, errors.New("Slack message file is empty"))
	}
	if !utf8.ValidString(message) {
		return fail(stderr, errors.New("Slack message file must be valid UTF-8"))
	}
	if utf8.RuneCountInString(message) > maxSlackMessageRunes {
		return failf(stderr, "Slack message exceeds %d characters", maxSlackMessageRunes)
	}
	if err := postSlackWebhook(config.WebhookURL, message); err != nil {
		return fail(stderr, err)
	}
	receipt := slackDeliveryReceipt{
		RunID:         runID,
		MessageSHA256: sha256String(message),
		SentAt:        time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := writeSlackDeliveryReceipt(receiptPath, receipt); err != nil {
		return failf(stderr, "Slack accepted the message but the local delivery receipt could not be saved: %v", err)
	}
	writeJSON(stdout, map[string]any{
		"slug":      slug,
		"run_id":    runID,
		"sent":      true,
		"duplicate": false,
	})
	return 0
}

func cmdMonitorDeliveryRemoveSlack(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return fail(stderr, errors.New("usage: newsjack monitor delivery remove-slack <slug>"))
	}
	slug := slugify(args[0])
	if err := requireMonitor(slug); err != nil {
		return fail(stderr, err)
	}
	path := monitorDeliveryConfigPath(slug)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fail(stderr, err)
	}
	writeJSON(stdout, map[string]any{
		"slug":             slug,
		"slack_configured": false,
		"removed":          true,
	})
	return 0
}

func requireMonitor(slug string) error {
	if !fileExists(monitorProfilePath(slug)) {
		return fmt.Errorf("monitor profile not found: %s", slug)
	}
	return nil
}

func normalizeSlackNotifyOn(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pitch-ready", "pitch_ready":
		return slackNotifyPitchReady, nil
	case "every-run", "every_run":
		return slackNotifyEveryRun, nil
	default:
		return "", errors.New("--notify-on must be every-run or pitch-ready")
	}
}

func readSlackWebhookURL(stdin io.Reader, stderr io.Writer) (string, error) {
	fmt.Fprint(stderr, "Slack incoming webhook URL: ")
	if file, ok := stdin.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		secret, err := term.ReadPassword(int(file.Fd()))
		fmt.Fprintln(stderr)
		if err != nil {
			return "", errors.New("could not read Slack webhook URL")
		}
		value := strings.TrimSpace(string(secret))
		if value == "" {
			return "", errors.New("Slack webhook URL is required")
		}
		return value, nil
	}
	secret, err := io.ReadAll(io.LimitReader(stdin, maxSlackWebhookBytes+1))
	if err != nil {
		return "", errors.New("could not read Slack webhook URL")
	}
	if len(secret) > maxSlackWebhookBytes {
		return "", errors.New("Slack webhook URL is too long")
	}
	value := strings.TrimSpace(string(secret))
	if value == "" {
		return "", errors.New("Slack webhook URL is required on standard input")
	}
	return value, nil
}

func validateSlackWebhookURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return errors.New("invalid Slack incoming webhook URL")
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme != "https" || (host != "hooks.slack.com" && host != "hooks.slack-gov.com") || parsed.Port() != "" {
		return errors.New("webhook must be an HTTPS Slack incoming webhook URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("invalid Slack incoming webhook URL")
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(parsed.EscapedPath(), "/services/"), "/"), "/")
	if !strings.HasPrefix(parsed.EscapedPath(), "/services/") || len(parts) != 3 {
		return errors.New("invalid Slack incoming webhook URL")
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return errors.New("invalid Slack incoming webhook URL")
		}
	}
	return nil
}

func configuredSlackDelivery(slug string) (*slackDeliveryConfig, error) {
	if err := requireMonitor(slug); err != nil {
		return nil, err
	}
	config, configured, err := readMonitorDeliveryConfig(slug)
	if err != nil {
		return nil, err
	}
	if !configured {
		return nil, fmt.Errorf("Slack delivery is not configured for %s", slug)
	}
	if err := validateSlackWebhookURL(config.Slack.WebhookURL); err != nil {
		return nil, errors.New("saved Slack delivery configuration is invalid; run set-slack again")
	}
	return config.Slack, nil
}

func readMonitorDeliveryConfig(slug string) (monitorDeliveryConfig, bool, error) {
	path := monitorDeliveryConfigPath(slug)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return monitorDeliveryConfig{}, false, nil
		}
		return monitorDeliveryConfig{}, false, err
	}
	var config monitorDeliveryConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return monitorDeliveryConfig{}, false, fmt.Errorf("invalid delivery configuration at %s", path)
	}
	if config.Slack == nil || strings.TrimSpace(config.Slack.WebhookURL) == "" {
		return config, false, nil
	}
	if _, err := normalizeSlackNotifyOn(config.Slack.NotifyOn); err != nil {
		return monitorDeliveryConfig{}, false, fmt.Errorf("invalid delivery configuration at %s", path)
	}
	return config, true, nil
}

func writeMonitorDeliveryConfig(slug string, config monitorDeliveryConfig) (string, error) {
	path := monitorDeliveryConfigPath(slug)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func postSlackWebhook(webhookURL, message string) error {
	if err := validateSlackWebhookURL(webhookURL); err != nil {
		return err
	}
	return postSlackWebhookRequest(webhookURL, message)
}

func postSlackWebhookRequest(webhookURL, message string) error {
	payload, err := json.Marshal(map[string]string{"text": message})
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return errors.New("could not create Slack delivery request")
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := slackDeliveryHTTPClient.Do(request)
	if err != nil {
		return errors.New("Slack delivery request failed")
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	if response.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "ok" {
		return fmt.Errorf("Slack rejected the message with HTTP %d", response.StatusCode)
	}
	return nil
}

func monitorDeliveryConfigPath(slug string) string {
	return filepath.Join(monitorDir(slug), "delivery.json")
}

func monitorSlackReceiptPath(slug, runID string) string {
	return filepath.Join(monitorDir(slug), "delivery-receipts", "slack-"+sha256String(runID)+".json")
}

func writeSlackDeliveryReceipt(path string, receipt slackDeliveryReceipt) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func sha256String(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
