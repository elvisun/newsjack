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
	"time"
)

type bridgeRequestLog struct {
	mu        sync.Mutex
	auth      []string
	sessions  []string
	protocols []string
}

func (l *bridgeRequestLog) record(r *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.auth = append(l.auth, r.Header.Get("Authorization"))
	l.sessions = append(l.sessions, r.Header.Get("Mcp-Session-Id"))
	l.protocols = append(l.protocols, r.Header.Get("MCP-Protocol-Version"))
}

func newMockMCPServer(t *testing.T, log *bridgeRequestLog, bearerToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		if r.Header.Get("Authorization") != "Bearer "+bearerToken {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"bad token"}`)
			return
		}
		body := readRequestBody(t, r)
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

func readRequestBody(t *testing.T, r *http.Request) []byte {
	t.Helper()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		t.Errorf("mock server read: %v", err)
	}
	return buf.Bytes()
}

func runBridge(t *testing.T, endpoint, token, input string) (int, string, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	bridge := newMCPBridge(endpoint, medialystBearerCredential{Token: token, Kind: "oauth", Source: "test"}, &out, &errBuf)
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
	server := newMockMCPServer(t, log, "mcp_at_saved")
	defer server.Close()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}` + "\n"
	code, stdout, stderr := runBridge(t, server.URL, "mcp_at_saved", input)
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
	code, stdout, stderr = runBridge(t, server.URL, "mcp_at_saved", input)
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
		if auth != "Bearer mcp_at_saved" {
			t.Fatalf("request %d missing bearer auth: %q", i, auth)
		}
	}
}

func TestMCPBridgeRejectedCredentialsFailWithLoginHint(t *testing.T) {
	log := &bridgeRequestLog{}
	server := newMockMCPServer(t, log, "mcp_at_saved")
	defer server.Close()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	code, stdout, stderr := runBridge(t, server.URL, "mcp_at_wrong", input)
	if code == 0 {
		t.Fatalf("bridge should exit nonzero on rejected credentials, stdout=%s", stdout)
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
	code, stdout, stderr := runBridge(t, server.URL, "mcp_at_saved", input)
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
	code, stdout, stderr := runBridge(t, "http://127.0.0.1:1", "mcp_at_saved", input)
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
	server := newMockMCPServer(t, log, "mcp_at_saved")
	defer server.Close()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	var out, errBuf bytes.Buffer
	bridge := newMCPBridge(server.URL, medialystBearerCredential{Token: "mcp_at_saved", Kind: "oauth", Source: "test"}, &out, &errBuf)
	if code := bridge.run(strings.NewReader(input)); code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, errBuf.String())
	}
	if code := bridge.run(strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")); code != 0 {
		t.Fatalf("bridge exit=%d stderr=%s", code, errBuf.String())
	}

	log.mu.Lock()
	defer log.mu.Unlock()
	lastSession := log.sessions[len(log.sessions)-1]
	if lastSession != "session-123" {
		t.Fatalf("follow-up request should carry the negotiated session id, got %q (all: %v)", lastSession, log.sessions)
	}
	lastProtocol := log.protocols[len(log.protocols)-1]
	if lastProtocol != "2025-03-26" {
		t.Fatalf("follow-up request should carry the negotiated protocol, got %q (all: %v)", lastProtocol, log.protocols)
	}
}

func TestMCPBridgeCommandLoadsOAuthCredentialsFile(t *testing.T) {
	log := &bridgeRequestLog{}
	server := newMockMCPServer(t, log, "mcp_at_saved")
	defer server.Close()

	home := t.TempDir()
	writeTestOAuthCredentials(t, home, "mcp_at_saved", "mcp_rt_saved", "")

	withTempEnv(t, map[string]string{
		"HOME":                       home,
		"NEWSJACK_HOME":              "",
		"NEWSJACK_IGNORE_DOTENV":     "1",
		"MEDIALYST_API_KEY":          "",
		"NEWSJACK_MEDIALYST_MCP_URL": server.URL,
	}, func() {
		input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
		var out, errBuf bytes.Buffer
		code := runCLIWithIO([]string{"mcp-bridge"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("mcp-bridge exited %d: %s", code, errBuf.String())
		}
		bridgeResponseByID(t, out.String(), 1)
	})

	log.mu.Lock()
	defer log.mu.Unlock()
	if len(log.auth) != 1 || log.auth[0] != "Bearer mcp_at_saved" {
		t.Fatalf("mock server auth headers=%v", log.auth)
	}
}

func TestMCPBridgeRefreshesOAuthAfterUnauthorized(t *testing.T) {
	log := &bridgeRequestLog{}
	tokenRequests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		switch r.Header.Get("Authorization") {
		case "Bearer mcp_at_old":
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"expired"}`)
			return
		case "Bearer mcp_at_new":
		default:
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"bad token"}`)
			return
		}
		body := readRequestBody(t, r)
		var envelope struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			t.Errorf("mock server got non-JSON body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-03-26","capabilities":{}}}`, envelope.ID)
	})
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse refresh form: %v", err)
		}
		if r.Form.Get("grant_type") != refreshGrantType || r.Form.Get("refresh_token") != "mcp_rt_old" {
			t.Errorf("unexpected refresh form: %v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"mcp_at_new","token_type":"Bearer","expires_in":3600,"refresh_token":"mcp_rt_new","scope":"news:search media_lists:manage"}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	home := t.TempDir()
	writeTestOAuthCredentials(t, home, "mcp_at_old", "mcp_rt_old", server.URL)

	withTempEnv(t, map[string]string{
		"HOME":                       home,
		"NEWSJACK_HOME":              "",
		"NEWSJACK_IGNORE_DOTENV":     "1",
		"MEDIALYST_API_KEY":          "",
		"NEWSJACK_MEDIALYST_MCP_URL": server.URL + "/mcp",
	}, func() {
		input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
		var out, errBuf bytes.Buffer
		code := runCLIWithIO([]string{"mcp-bridge"}, strings.NewReader(input), &out, &errBuf)
		if code != 0 {
			t.Fatalf("mcp-bridge exited %d: %s", code, errBuf.String())
		}
		bridgeResponseByID(t, out.String(), 1)

		creds, ok, err := readCredentialsFile()
		if err != nil || !ok || creds.Medialyst.OAuth == nil {
			t.Fatalf("read refreshed credentials ok=%v err=%v creds=%#v", ok, err, creds)
		}
		if creds.Medialyst.OAuth.AccessToken != "mcp_at_new" || creds.Medialyst.OAuth.RefreshToken != "mcp_rt_new" {
			t.Fatalf("credentials were not refreshed: %#v", creds.Medialyst.OAuth)
		}
	})

	if tokenRequests != 1 {
		t.Fatalf("refresh endpoint calls=%d, want 1", tokenRequests)
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	if got := strings.Join(log.auth, ","); got != "Bearer mcp_at_old,Bearer mcp_at_new" {
		t.Fatalf("MCP auth sequence=%s", got)
	}
}

func TestMCPBridgeWithoutCredentialsPrintsLoginHint(t *testing.T) {
	home := t.TempDir()
	withTempEnv(t, map[string]string{
		"HOME":                   home,
		"NEWSJACK_HOME":          "",
		"NEWSJACK_IGNORE_DOTENV": "1",
		"MEDIALYST_API_KEY":      "",
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLIWithIO([]string{"mcp-bridge"}, strings.NewReader(""), &out, &errBuf)
		if code != 1 {
			t.Fatalf("missing credentials should exit 1, got %d", code)
		}
		if !strings.Contains(errBuf.String(), "newsjack login") {
			t.Fatalf("stderr should point at login:\n%s", errBuf.String())
		}
		if strings.Contains(errBuf.String(), "unknown command") {
			t.Fatalf("mcp-bridge command was not dispatched:\n%s", errBuf.String())
		}
	})
}

func writeTestOAuthCredentials(t *testing.T, home, accessToken, refreshToken, baseURL string) {
	t.Helper()
	credentials := filepath.Join(home, ".newsjack", "credentials.json")
	if err := os.MkdirAll(filepath.Dir(credentials), 0o700); err != nil {
		t.Fatal(err)
	}
	payload := fmt.Sprintf(`{
		"medialyst": {
			"oauth": {
				"access_token": %q,
				"refresh_token": %q,
				"token_type": "Bearer",
				"expires_at": %q,
				"scope": "news:search media_lists:manage",
				"client_id": %q,
				"base_url": %q
			},
			"source": %q
		}
	}`, accessToken, refreshToken, time.Now().Add(time.Hour).UTC().Format(time.RFC3339), medialystOAuthClientID, baseURL, medialystOAuthSource)
	if err := os.WriteFile(credentials, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
}
