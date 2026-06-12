package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type bridgeRequestLog struct {
	mu       sync.Mutex
	auth     []string
	sessions []string
}

func (l *bridgeRequestLog) record(r *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.auth = append(l.auth, r.Header.Get("Authorization"))
	l.sessions = append(l.sessions, r.Header.Get("Mcp-Session-Id"))
}

func newMockMCPServer(t *testing.T, log *bridgeRequestLog) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		if r.Header.Get("Authorization") != "Bearer mlst_test_key" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"bad token"}`)
			return
		}
		body, err := json.RawMessage(nil), error(nil)
		raw := make([]byte, 0, 1024)
		buf := bytes.NewBuffer(raw)
		if _, err = buf.ReadFrom(r.Body); err != nil {
			t.Errorf("mock server read: %v", err)
		}
		body = buf.Bytes()
		var envelope struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			t.Errorf("mock server got non-JSON body: %s", body)
		}
		switch envelope.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "session-123")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"jsonrpc": "2.0",
				"id": %s,
				"result": {"protocolVersion": "2025-03-26", "capabilities": {}, "serverInfo": {"name": "mock-medialyst"}}
			}`, envelope.ID)
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
		case "tools/list":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{\"tools\":[{\"name\":\"mock_search\"}]}}\n\n", envelope.ID)
		default:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{}}`, envelope.ID)
		}
	}))
}

func runBridge(t *testing.T, endpoint, key, input string) (int, string, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	bridge := newMCPBridge(endpoint, key, &out, &errBuf)
	code := bridge.run(strings.NewReader(input))
	return code, out.String(), errBuf.String()
}

func bridgeResponseByID(t *testing.T, stdout string, id int) map[string]any {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var msg map[string]any
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Fatalf("bridge stdout line is not JSON: %q", line)
		}
		if got, ok := msg["id"].(float64); ok && int(got) == id {
			return msg
		}
	}
	t.Fatalf("no response with id=%d in bridge stdout:\n%s", id, stdout)
	return nil
}

func TestMCPBridgeInitializeAndToolListRoundTrip(t *testing.T) {
	log := &bridgeRequestLog{}
	server := newMockMCPServer(t, log)
	defer server.Close()

	// Sequential messages mirror the MCP handshake order; the harness waits
	// for each response before sending the next message.
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}` + "\n"
	code, stdout, stderr := runBridge(t, server.URL, "mlst_test_key", input)
	if code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, stderr)
	}
	initResp := bridgeResponseByID(t, stdout, 1)
	result, _ := initResp["result"].(map[string]any)
	if result == nil || result["protocolVersion"] != "2025-03-26" {
		t.Fatalf("initialize result missing protocol version: %s", stdout)
	}

	input = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}` + "\n" +
		`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	code, stdout, stderr = runBridge(t, server.URL, "mlst_test_key", input)
	if code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, stderr)
	}
	toolsResp := bridgeResponseByID(t, stdout, 2)
	if !strings.Contains(stdout, "mock_search") {
		t.Fatalf("SSE tool response not relayed: %s", stdout)
	}
	if toolsResp["error"] != nil {
		t.Fatalf("tools/list relayed an error: %s", stdout)
	}

	log.mu.Lock()
	defer log.mu.Unlock()
	if len(log.auth) == 0 {
		t.Fatal("mock server saw no requests")
	}
	for i, auth := range log.auth {
		if auth != "Bearer mlst_test_key" {
			t.Fatalf("request %d missing bearer auth: %q", i, auth)
		}
	}
}

func TestMCPBridgeRejectedKeyFailsWithLoginHint(t *testing.T) {
	log := &bridgeRequestLog{}
	server := newMockMCPServer(t, log)
	defer server.Close()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	code, stdout, stderr := runBridge(t, server.URL, "mlst_wrong_key", input)
	if code == 0 {
		t.Fatalf("bridge should exit nonzero on rejected key, stdout=%s", stdout)
	}
	if !strings.Contains(stderr, "login") {
		t.Fatalf("stderr should include the re-auth hint:\n%s", stderr)
	}
	resp := bridgeResponseByID(t, stdout, 1)
	if resp["error"] == nil {
		t.Fatalf("client should receive a JSON-RPC error instead of hanging: %s", stdout)
	}
}

func TestMCPBridgeServerErrorAnswersClientWithoutHanging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	input := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{}}` + "\n"
	code, stdout, stderr := runBridge(t, server.URL, "mlst_test_key", input)
	if code != 0 {
		t.Fatalf("HTTP 502 should not be fatal, exit=%d stderr=%s", code, stderr)
	}
	resp := bridgeResponseByID(t, stdout, 7)
	errPayload, _ := resp["error"].(map[string]any)
	if errPayload == nil || !strings.Contains(stringValue(errPayload["message"]), "502") {
		t.Fatalf("client should see the upstream status: %s", stdout)
	}
}

func TestMCPBridgeUnreachableServerAnswersClient(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":3,"method":"initialize","params":{}}` + "\n"
	code, stdout, stderr := runBridge(t, "http://127.0.0.1:1", "mlst_test_key", input)
	if code != 0 {
		t.Fatalf("connection failure should not crash the bridge loop, exit=%d", code)
	}
	if !strings.Contains(stderr, "request failed") {
		t.Fatalf("stderr should describe the connection failure:\n%s", stderr)
	}
	resp := bridgeResponseByID(t, stdout, 3)
	if resp["error"] == nil {
		t.Fatalf("client should receive a JSON-RPC error: %s", stdout)
	}
}

func TestMCPBridgeSessionHeaderPropagates(t *testing.T) {
	log := &bridgeRequestLog{}
	server := newMockMCPServer(t, log)
	defer server.Close()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	if code, _, stderr := runBridge(t, server.URL, "mlst_test_key", input); code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, stderr)
	}

	var out, errBuf bytes.Buffer
	bridge := newMCPBridge(server.URL, "mlst_test_key", &out, &errBuf)
	if code := bridge.run(strings.NewReader(input)); code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, errBuf.String())
	}
	if code := bridge.run(strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")); code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, errBuf.String())
	}

	log.mu.Lock()
	defer log.mu.Unlock()
	last := log.sessions[len(log.sessions)-1]
	if last != "session-123" {
		t.Fatalf("follow-up request should carry the negotiated session id, got %q (all: %v)", last, log.sessions)
	}
}

func TestCmdMCPBridgeLoadsKeyFromCredentialsFile(t *testing.T) {
	log := &bridgeRequestLog{}
	server := newMockMCPServer(t, log)
	defer server.Close()

	home := t.TempDir()
	credentials := filepath.Join(home, ".newsjack", "credentials.json")
	if err := os.MkdirAll(filepath.Dir(credentials), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(credentials, []byte(`{"medialyst":{"api_key":"mlst_test_key"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	withTempEnv(t, map[string]string{
		"HOME":                       home,
		"NEWSJACK_HOME":              "",
		"NEWSJACK_IGNORE_DOTENV":     "1",
		"MEDIALYST_API_KEY":          "",
		"NEWSJACK_MEDIALYST_MCP_URL": server.URL,
	}, func() {
		key, source := loadAPIKey()
		if key != "mlst_test_key" || !strings.HasPrefix(source, "credentials:") {
			t.Fatalf("key/source = %q/%q, want credentials file", key, source)
		}
		if medialystMCPEndpoint() != server.URL {
			t.Fatalf("endpoint override not applied: %s", medialystMCPEndpoint())
		}
	})
}

func TestCmdMCPBridgeWithoutKeyPrintsLoginHint(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                   home,
		"NEWSJACK_HOME":          "",
		"NEWSJACK_IGNORE_DOTENV": "1",
		"MEDIALYST_API_KEY":      "",
	}, func() {
		var out, errBuf bytes.Buffer
		if code := cmdMCPBridge(nil, &out, &errBuf); code != 1 {
			t.Fatalf("missing key should exit 1, got %d", code)
		}
		if !strings.Contains(errBuf.String(), "login") {
			t.Fatalf("stderr should point at login:\n%s", errBuf.String())
		}
	})
}
