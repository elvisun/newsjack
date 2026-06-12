package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// mcpBridgeMaxMessageBytes bounds a single stdio JSON-RPC message.
const mcpBridgeMaxMessageBytes = 16 << 20

func medialystMCPEndpoint() string {
	return getenv("NEWSJACK_MEDIALYST_MCP_URL", medialystMCPURL)
}

func cmdMCPBridge(_ []string, stdout io.Writer, stderr io.Writer) int {
	key, source := loadAPIKey()
	if key == "" {
		fmt.Fprintf(stderr, "Medialyst API key not found. Run: %s\n", newsjackCLIInvocation().Display("login"))
		return 1
	}
	if os.Getenv("NEWSJACK_AUTH_DEBUG") != "" {
		fmt.Fprintf(stderr, "Loaded Medialyst API key from %s\n", source)
	}
	bridge := newMCPBridge(medialystMCPEndpoint(), key, stdout, stderr)
	return bridge.run(os.Stdin)
}

// mcpBridge proxies MCP stdio (newline-delimited JSON-RPC) from a local
// agent harness to the remote Medialyst streamable-HTTP MCP endpoint. The
// bearer token is attached per request so the key never lands in harness
// config files, and no external runtime (Node/npx) is required.
type mcpBridge struct {
	endpoint string
	key      string
	client   *http.Client
	stdout   io.Writer
	stderr   io.Writer

	writeMu  sync.Mutex
	stateMu  sync.Mutex
	session  string
	protocol string
}

func newMCPBridge(endpoint, key string, stdout, stderr io.Writer) *mcpBridge {
	return &mcpBridge{
		endpoint: endpoint,
		key:      key,
		client:   &http.Client{Timeout: 5 * time.Minute},
		stdout:   stdout,
		stderr:   stderr,
	}
}

// run processes messages strictly in order: a piped handshake must not let
// tools/list reach the server before the initialize response has populated
// the session and protocol headers, and a fatal condition (rejected key)
// must end the process promptly instead of blocking on stdin.
func (b *mcpBridge) run(stdin io.Reader) int {
	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 64*1024), mcpBridgeMaxMessageBytes)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		msg := append([]byte(nil), line...)
		if code := b.forward(msg); code != 0 {
			return code
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(b.stderr, "newsjack mcp-bridge: stdin: %v\n", err)
		return 1
	}
	return 0
}

// forward sends one client JSON-RPC message upstream and relays every
// resulting server message to stdout. A nonzero return marks a fatal
// bridge condition (for example rejected credentials), not a normal
// JSON-RPC level error.
func (b *mcpBridge) forward(msg []byte) int {
	req, err := http.NewRequest(http.MethodPost, b.endpoint, bytes.NewReader(msg))
	if err != nil {
		fmt.Fprintf(b.stderr, "newsjack mcp-bridge: %v\n", err)
		return 1
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+b.key)
	if session := b.sessionID(); session != "" {
		req.Header.Set("Mcp-Session-Id", session)
	}
	if protocol := b.protocolVersion(); protocol != "" {
		req.Header.Set("MCP-Protocol-Version", protocol)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		fmt.Fprintf(b.stderr, "newsjack mcp-bridge: Medialyst MCP request failed: %v\n", err)
		b.writeErrorResponse(msg, "Medialyst MCP request failed; check network connectivity")
		return 0
	}
	defer resp.Body.Close()

	if session := resp.Header.Get("Mcp-Session-Id"); session != "" {
		b.setSessionID(session)
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		fmt.Fprintf(b.stderr, "Medialyst rejected the API key (HTTP %d). Run: %s\n", resp.StatusCode, newsjackCLIInvocation().Display("login"))
		b.writeErrorResponse(msg, "Medialyst authentication failed; run: newsjack login")
		return 1
	case resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNoContent:
		_, _ = io.Copy(io.Discard, resp.Body)
		return 0
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		fmt.Fprintf(b.stderr, "newsjack mcp-bridge: HTTP %d from Medialyst MCP: %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
		b.writeErrorResponse(msg, fmt.Sprintf("Medialyst MCP returned HTTP %d", resp.StatusCode))
		return 0
	}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		b.relaySSE(resp.Body)
		return 0
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, mcpBridgeMaxMessageBytes))
	if err != nil {
		fmt.Fprintf(b.stderr, "newsjack mcp-bridge: reading Medialyst response: %v\n", err)
		b.writeErrorResponse(msg, "Medialyst MCP response could not be read")
		return 0
	}
	b.relayMessage(body)
	return 0
}

// relaySSE forwards each data event of a text/event-stream response.
func (b *mcpBridge) relaySSE(body io.Reader) {
	reader := bufio.NewReaderSize(body, 64*1024)
	var data []string
	flush := func() {
		if len(data) == 0 {
			return
		}
		b.relayMessage([]byte(strings.Join(data, "\n")))
		data = data[:0]
	}
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
		if err != nil {
			flush()
			return
		}
	}
}

// relayMessage writes one upstream JSON-RPC payload to stdout in
// newline-delimited form, compacting so embedded newlines cannot break
// the stdio framing.
func (b *mcpBridge) relayMessage(raw []byte) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		fmt.Fprintf(b.stderr, "newsjack mcp-bridge: dropping non-JSON upstream payload: %v\n", err)
		return
	}
	b.captureProtocolVersion(compact.Bytes())
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	_, _ = b.stdout.Write(append(compact.Bytes(), '\n'))
}

// writeErrorResponse answers a client request locally so the harness does
// not hang waiting on an upstream reply that will never come.
// Notifications (no id) get no response per JSON-RPC.
func (b *mcpBridge) writeErrorResponse(request []byte, message string) {
	var envelope struct {
		ID json.RawMessage `json:"id"`
	}
	if json.Unmarshal(request, &envelope) != nil || len(envelope.ID) == 0 || string(envelope.ID) == "null" {
		return
	}
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      json.RawMessage(envelope.ID),
		"error":   map[string]any{"code": -32603, "message": message},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	_, _ = b.stdout.Write(append(data, '\n'))
}

// captureProtocolVersion remembers the negotiated MCP protocol version
// from the initialize result so later HTTP requests can advertise it.
func (b *mcpBridge) captureProtocolVersion(message []byte) {
	if b.protocolVersion() != "" || !bytes.Contains(message, []byte("protocolVersion")) {
		return
	}
	var envelope struct {
		Result struct {
			ProtocolVersion string `json:"protocolVersion"`
		} `json:"result"`
	}
	if json.Unmarshal(message, &envelope) == nil && envelope.Result.ProtocolVersion != "" {
		b.setProtocolVersion(envelope.Result.ProtocolVersion)
	}
}

func (b *mcpBridge) sessionID() string {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return b.session
}

func (b *mcpBridge) setSessionID(session string) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	b.session = session
}

func (b *mcpBridge) protocolVersion() string {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return b.protocol
}

func (b *mcpBridge) setProtocolVersion(protocol string) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	b.protocol = protocol
}
