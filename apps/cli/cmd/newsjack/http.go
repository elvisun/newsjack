package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// maxHTTPBodyBytes caps how much of any HTTP response body newsjack will read
// (feeds, manifests, API replies). Larger bodies fail instead of exhausting memory.
const maxHTTPBodyBytes = 10 << 20 // 10 MiB

var errBodyTooLarge = errors.New("response body exceeds size limit")

// readBodyLimited reads at most limit bytes and fails if the body is larger.
func readBodyLimited(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w (> %d bytes)", errBodyTooLarge, limit)
	}
	return data, nil
}

// allowPrivateURLs disables the SSRF guard for local development and tests
// (feeds served from 127.0.0.1, etc.). Never set it on a server.
func allowPrivateURLs() bool {
	return os.Getenv("NEWSJACK_ALLOW_PRIVATE_URLS") == "1"
}

// isDisallowedIP reports whether an address must not be contacted on behalf of a
// user-supplied URL: loopback, link-local (incl. cloud metadata 169.254.169.254),
// RFC 1918 / ULA private ranges, CGNAT, multicast, and unspecified addresses.
func isDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() || ip.IsPrivate() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		// 100.64.0.0/10 (carrier-grade NAT) and 0.0.0.0/8
		if ip4[0] == 100 && ip4[1]&0xc0 == 64 {
			return true
		}
		if ip4[0] == 0 {
			return true
		}
	}
	return false
}

// externalHTTPClient returns a client for URLs that come from user data (feed URLs
// in monitor profiles). It resolves the host itself, refuses non-public addresses,
// and dials the vetted IP directly, so redirects and DNS tricks cannot reach
// internal services or cloud metadata endpoints. NEWSJACK_ALLOW_PRIVATE_URLS=1
// turns the guard off for local development.
func externalHTTPClient(timeout time.Duration) *http.Client {
	if allowPrivateURLs() {
		return &http.Client{Timeout: timeout}
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
		ForceAttemptHTTP2:   true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses for %s", host)
			}
			for _, ip := range ips {
				if isDisallowedIP(ip.IP) {
					return nil, fmt.Errorf("refusing to connect to non-public address %s for host %s", ip.IP, host)
				}
			}
			var lastErr error
			for _, ip := range ips {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

func httpJSON(method, rawURL string, headers map[string]string, body any, timeout time.Duration) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := readBodyLimited(resp.Body, maxHTTPBodyBytes)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// httpGetRaw fetches operator-configured URLs (release manifests, API bases).
func httpGetRaw(rawURL string, headers map[string]string, timeout time.Duration) (string, error) {
	return httpGetRawWith(&http.Client{Timeout: timeout}, rawURL, headers)
}

// httpGetRawExternal fetches URLs that originate from user data (profile feed URLs):
// same as httpGetRaw but through the SSRF-guarded client.
func httpGetRawExternal(rawURL string, headers map[string]string, timeout time.Duration) (string, error) {
	return httpGetRawWith(externalHTTPClient(timeout), rawURL, headers)
}

func httpGetRawWith(client *http.Client, rawURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := readBodyLimited(resp.Body, maxHTTPBodyBytes)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	return string(data), nil
}
