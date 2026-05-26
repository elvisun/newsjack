package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func httpGetRaw(rawURL string, headers map[string]string, timeout time.Duration) (string, error) {
	resp, err := httpGetRawResponse(rawURL, headers, timeout)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(resp.Body, 300))
	}
	return resp.Body, nil
}

type httpRawResponse struct {
	StatusCode int
	Header     http.Header
	Body       string
}

func httpGetRawResponse(rawURL string, headers map[string]string, timeout time.Duration) (httpRawResponse, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return httpRawResponse{}, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	setDefaultUserAgent(req)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return httpRawResponse{}, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return httpRawResponse{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: string(data)}, nil
}

func setDefaultUserAgent(req *http.Request) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", fmt.Sprintf("Mozilla/5.0 (compatible; newsjack.sh/%s; +https://newsjack.sh)", version))
	}
}
