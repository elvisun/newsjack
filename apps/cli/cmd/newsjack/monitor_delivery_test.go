package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestMonitorSlackDeliveryLifecycleKeepsWebhookSecret(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		createMonitorProfileForDeliveryTest(t, "acme")
		webhook := "https://hooks.slack.com/services/T000/B000/super-secret-token"

		var out, errBuf bytes.Buffer
		code := runCLIWithIO(
			[]string{"monitor", "delivery", "set-slack", "acme"},
			strings.NewReader(webhook+"\n"),
			&out,
			&errBuf,
		)
		if code != 0 {
			t.Fatalf("set-slack code=%d stderr=%s", code, errBuf.String())
		}
		if strings.Contains(out.String()+errBuf.String(), webhook) || strings.Contains(out.String()+errBuf.String(), "super-secret-token") {
			t.Fatalf("set-slack leaked the webhook: stdout=%s stderr=%s", out.String(), errBuf.String())
		}

		configPath := monitorDeliveryConfigPath("acme")
		assertOwnerOnlyFile(t, configPath)
		configBody, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(configBody), webhook) || !strings.Contains(string(configBody), `"notify_on": "every_run"`) {
			t.Fatalf("saved delivery config mismatch:\n%s", configBody)
		}
		profileBody, err := os.ReadFile(monitorProfilePath("acme"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(profileBody), "slack") || strings.Contains(string(profileBody), webhook) {
			t.Fatalf("profile should not contain delivery credentials:\n%s", profileBody)
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"monitor", "delivery", "status", "acme"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("delivery status code=%d stderr=%s", code, errBuf.String())
		}
		var status map[string]any
		if err := json.Unmarshal(out.Bytes(), &status); err != nil {
			t.Fatalf("invalid status JSON: %v\n%s", err, out.String())
		}
		if status["slack_configured"] != true || status["notify_on"] != slackNotifyEveryRun {
			t.Fatalf("unexpected delivery status: %#v", status)
		}
		if strings.Contains(out.String(), webhook) || strings.Contains(out.String(), "super-secret-token") {
			t.Fatalf("delivery status leaked the webhook: %s", out.String())
		}

		out.Reset()
		errBuf.Reset()
		code = runCLI([]string{"monitor", "delivery", "remove-slack", "acme"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("remove-slack code=%d stderr=%s", code, errBuf.String())
		}
		if fileExists(configPath) {
			t.Fatal("remove-slack left delivery.json behind")
		}
	})
}

func TestNormalizeSlackNotifyOnSupportsQuieterOverride(t *testing.T) {
	for input, want := range map[string]string{
		"every-run":   slackNotifyEveryRun,
		"every_run":   slackNotifyEveryRun,
		"pitch-ready": slackNotifyPitchReady,
		"pitch_ready": slackNotifyPitchReady,
	} {
		got, err := normalizeSlackNotifyOn(input)
		if err != nil {
			t.Fatalf("normalizeSlackNotifyOn(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("normalizeSlackNotifyOn(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestMonitorSlackDeliveryIsAbsentForExistingMonitorsByDefault(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		createMonitorProfileForDeliveryTest(t, "legacy-monitor")
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{"monitor", "delivery", "status", "legacy-monitor"}, &out, &errBuf); code != 0 {
			t.Fatalf("delivery status code=%d stderr=%s", code, errBuf.String())
		}
		var status map[string]any
		if err := json.Unmarshal(out.Bytes(), &status); err != nil {
			t.Fatal(err)
		}
		if status["slack_configured"] != false || status["notify_on"] != nil {
			t.Fatalf("legacy monitor should remain unconfigured: %#v", status)
		}
		if fileExists(monitorDeliveryConfigPath("legacy-monitor")) {
			t.Fatal("status should not create delivery state")
		}
	})
}

func TestMonitorSlackDeliveryRejectsNonSlackWebhookURLs(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		createMonitorProfileForDeliveryTest(t, "acme")
		for _, webhook := range []string{
			"http://hooks.slack.com/services/T/B/secret",
			"https://hooks.slack.com.evil.example/services/T/B/secret",
			"https://hooks.slack.com/services/T/B/secret?copy=1",
			"https://hooks.slack.com/services/T/B",
			"https://user@hooks.slack.com/services/T/B/secret",
		} {
			t.Run(webhook, func(t *testing.T) {
				var out, errBuf bytes.Buffer
				code := runCLIWithIO(
					[]string{"monitor", "delivery", "set-slack", "acme"},
					strings.NewReader(webhook+"\n"),
					&out,
					&errBuf,
				)
				if code == 0 {
					t.Fatalf("set-slack accepted %q", webhook)
				}
				if strings.Contains(out.String()+errBuf.String(), webhook) {
					t.Fatalf("validation error leaked webhook %q: %s", webhook, errBuf.String())
				}
				if fileExists(monitorDeliveryConfigPath("acme")) {
					t.Fatal("invalid webhook was saved")
				}
			})
		}
	})
}

func TestMonitorSlackDeliverySendsExactMessageAndDeduplicatesRun(t *testing.T) {
	home := t.TempDir()
	webhook := "https://hooks.slack.com/services/T000/B000/super-secret-token"
	message := "*Newsjack — Acme*\n1 pitch-ready · <https://example.com/story|Source>\n"
	var requests int
	var gotBody, gotContentType string
	originalClient := slackDeliveryHTTPClient
	slackDeliveryHTTPClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(body)
		gotContentType = request.Header.Get("Content-Type")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	t.Cleanup(func() { slackDeliveryHTTPClient = originalClient })

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		createMonitorProfileForDeliveryTest(t, "acme")
		if _, err := writeMonitorDeliveryConfig("acme", monitorDeliveryConfig{Slack: &slackDeliveryConfig{
			WebhookURL: webhook,
			NotifyOn:   slackNotifyPitchReady,
		}}); err != nil {
			t.Fatal(err)
		}
		messagePath := filepath.Join(t.TempDir(), "slack.md")
		if err := os.WriteFile(messagePath, []byte(message), 0o644); err != nil {
			t.Fatal(err)
		}

		args := []string{"monitor", "delivery", "send", "acme", "--message-file", messagePath, "--run-id", "20260914T120000Z"}
		var out, errBuf bytes.Buffer
		if code := runCLI(args, &out, &errBuf); code != 0 {
			t.Fatalf("delivery send code=%d stderr=%s", code, errBuf.String())
		}
		if requests != 1 {
			t.Fatalf("requests=%d, want 1", requests)
		}
		if gotContentType != "application/json" {
			t.Fatalf("Content-Type=%q", gotContentType)
		}
		var payload map[string]string
		if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
			t.Fatalf("invalid Slack JSON: %v\n%s", err, gotBody)
		}
		if payload["text"] != message {
			t.Fatalf("Slack text=%q, want %q", payload["text"], message)
		}
		if strings.Contains(out.String()+errBuf.String(), webhook) || strings.Contains(out.String()+errBuf.String(), "super-secret-token") {
			t.Fatal("delivery output leaked webhook")
		}

		receiptPath := monitorSlackReceiptPath("acme", "20260914T120000Z")
		assertOwnerOnlyFile(t, receiptPath)
		out.Reset()
		errBuf.Reset()
		if code := runCLI(args, &out, &errBuf); code != 0 {
			t.Fatalf("duplicate delivery code=%d stderr=%s", code, errBuf.String())
		}
		if requests != 1 {
			t.Fatalf("duplicate delivery made another request; requests=%d", requests)
		}
		var duplicate map[string]any
		if err := json.Unmarshal(out.Bytes(), &duplicate); err != nil {
			t.Fatal(err)
		}
		if duplicate["duplicate"] != true || duplicate["sent"] != false {
			t.Fatalf("unexpected duplicate response: %#v", duplicate)
		}
	})
}

func TestSlackWebhookUsesRealHTTPJSONContractAndDoesNotFollowRedirects(t *testing.T) {
	var gotMethod, gotContentType, gotText string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		gotMethod = request.Method
		gotContentType = request.Header.Get("Content-Type")
		var payload map[string]string
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		gotText = payload["text"]
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(server.Close)

	originalClient := slackDeliveryHTTPClient
	slackDeliveryHTTPClient = &http.Client{
		Timeout: originalClient.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	t.Cleanup(func() { slackDeliveryHTTPClient = originalClient })

	if err := postSlackWebhookRequest(server.URL, "hello from Newsjack"); err != nil {
		t.Fatalf("Slack-compatible HTTP exchange failed: %v", err)
	}
	if gotMethod != http.MethodPost || gotContentType != "application/json" || gotText != "hello from Newsjack" {
		t.Fatalf("request contract method=%q content-type=%q text=%q", gotMethod, gotContentType, gotText)
	}

	redirectTargetHit := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirectTargetHit = true
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(target.Close)
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		http.Redirect(w, request, target.URL, http.StatusFound)
	}))
	t.Cleanup(redirect.Close)
	if err := postSlackWebhookRequest(redirect.URL, "do not forward"); err == nil {
		t.Fatal("redirect response should fail")
	}
	if redirectTargetHit {
		t.Fatal("Slack delivery followed a redirect")
	}
}

func TestMonitorSlackDeliveryTestPostsSafeMessage(t *testing.T) {
	home := t.TempDir()
	webhook := "https://hooks.slack.com/services/T000/B000/super-secret-token"
	var gotText string
	originalClient := slackDeliveryHTTPClient
	slackDeliveryHTTPClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var payload map[string]string
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		gotText = payload["text"]
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	t.Cleanup(func() { slackDeliveryHTTPClient = originalClient })

	withTempEnv(t, map[string]string{
		"HOME":                    home,
		"NEWSJACK_HOME":           "",
		"NEWSJACK_NO_AUTO_UPDATE": "1",
	}, func() {
		createMonitorProfileForDeliveryTest(t, "acme")
		if _, err := writeMonitorDeliveryConfig("acme", monitorDeliveryConfig{Slack: &slackDeliveryConfig{
			WebhookURL: webhook,
			NotifyOn:   slackNotifyPitchReady,
		}}); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{"monitor", "delivery", "test", "acme"}, &out, &errBuf); code != 0 {
			t.Fatalf("delivery test code=%d stderr=%s", code, errBuf.String())
		}
		if gotText != "Newsjack Slack delivery is connected for `acme`." {
			t.Fatalf("test message=%q", gotText)
		}
	})
}

func createMonitorProfileForDeliveryTest(t *testing.T, slug string) {
	t.Helper()
	if err := os.MkdirAll(monitorDir(slug), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(monitorProfilePath(slug), []byte(`{"company":"Acme"}`), 0o644); err != nil {
		t.Fatal(err)
	}
}
