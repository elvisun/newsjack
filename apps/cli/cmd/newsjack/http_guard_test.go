package main

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsDisallowedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "127.8.8.8", "::1",
		"10.0.0.1", "172.16.0.1", "172.31.255.255", "192.168.1.1", // RFC 1918
		"169.254.169.254", "169.254.0.1", "fe80::1", // link-local incl. cloud metadata
		"fc00::1", "fd12::1", // ULA
		"100.64.0.1", "100.127.255.254", // CGNAT
		"0.0.0.0", "0.1.2.3", "::",
		"224.0.0.1", "ff02::1",
	}
	for _, s := range blocked {
		if !isDisallowedIP(net.ParseIP(s)) {
			t.Errorf("%s should be disallowed", s)
		}
	}
	allowed := []string{"8.8.8.8", "1.1.1.1", "140.82.112.3", "172.32.0.1", "100.128.0.1", "2606:4700:4700::1111"}
	for _, s := range allowed {
		if isDisallowedIP(net.ParseIP(s)) {
			t.Errorf("%s should be allowed", s)
		}
	}
	if !isDisallowedIP(nil) {
		t.Error("nil IP should be disallowed")
	}
}

func TestExternalClientRefusesLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<rss/>"))
	}))
	defer srv.Close()

	withTempEnv(t, map[string]string{"NEWSJACK_ALLOW_PRIVATE_URLS": ""}, func() {
		_, err := httpGetRawExternal(srv.URL, nil, 5*time.Second)
		if err == nil {
			t.Fatal("expected the SSRF guard to refuse a loopback URL")
		}
		if !strings.Contains(err.Error(), "non-public address") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// The guard is off only with the explicit development override.
	withTempEnv(t, map[string]string{"NEWSJACK_ALLOW_PRIVATE_URLS": "1"}, func() {
		body, err := httpGetRawExternal(srv.URL, nil, 5*time.Second)
		if err != nil {
			t.Fatalf("override should allow loopback: %v", err)
		}
		if body != "<rss/>" {
			t.Fatalf("unexpected body %q", body)
		}
	})
}

func TestExternalClientIgnoresEnvironmentProxy(t *testing.T) {
	// A proxy in the environment must not be used to reach a blocked destination.
	var proxied int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&proxied, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()
	withTempEnv(t, map[string]string{"NEWSJACK_ALLOW_PRIVATE_URLS": "", "HTTP_PROXY": proxy.URL, "http_proxy": proxy.URL, "HTTPS_PROXY": proxy.URL, "https_proxy": proxy.URL, "NO_PROXY": "", "no_proxy": ""}, func() {
		_, err := httpGetRawExternal("http://10.0.0.1/feed.xml", nil, 5*time.Second)
		if err == nil || !strings.Contains(err.Error(), "non-public address") {
			t.Fatalf("expected the guard to refuse 10.0.0.1 directly, got %v", err)
		}
	})
	if n := atomic.LoadInt32(&proxied); n != 0 {
		t.Fatalf("guarded client used the environment proxy (%d requests)", n)
	}
}

func TestCollectFeedRefusesPrivateURL(t *testing.T) {
	withTempEnv(t, map[string]string{"NEWSJACK_ALLOW_PRIVATE_URLS": ""}, func() {
		items, errText := collectFeed("http://169.254.169.254/latest/meta-data/", 5)
		if len(items) != 0 || errText == "" {
			t.Fatalf("expected a guarded error, got items=%d err=%q", len(items), errText)
		}
	})
}

func TestReadBodyLimited(t *testing.T) {
	data, err := readBodyLimited(strings.NewReader("hello"), 5)
	if err != nil || string(data) != "hello" {
		t.Fatalf("exact-size body should pass: %v %q", err, data)
	}
	_, err = readBodyLimited(strings.NewReader("hello!"), 5)
	if !errors.Is(err, errBodyTooLarge) {
		t.Fatalf("oversize body should fail with errBodyTooLarge, got %v", err)
	}
}

func TestHTTPGetRawCapsBodySize(t *testing.T) {
	big := strings.Repeat("x", maxHTTPBodyBytes+1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(big))
	}))
	defer srv.Close()
	_, err := httpGetRaw(srv.URL, nil, 10*time.Second)
	if !errors.Is(err, errBodyTooLarge) {
		t.Fatalf("expected body-size error, got %v", err)
	}
}
