// Command mock-mcp is a minimal streamable-HTTP MCP server for harness
// smoke tests of `newsjack mcp-bridge`. It is not product code: it exists so
// CI can prove the native bridge handshake works with no live Medialyst
// credentials and no Node runtime on the machine.
//
// Usage: go run . --addr 127.0.0.1:8970 --key mock-key
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8970", "listen address")
	key := flag.String("key", "mock-key", "bearer token the bridge must present")
	flag.Parse()

	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+*key {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"missing or wrong bearer token"}`)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var envelope struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch envelope.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "mock-session")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-03-26","capabilities":{"tools":{}},"serverInfo":{"name":"mock-medialyst","version":"0.0.1"}}}`, envelope.ID)
		case "tools/list":
			// Answer over SSE to exercise the bridge's event-stream path.
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{\"tools\":[{\"name\":\"mock_search\",\"description\":\"harness smoke tool\",\"inputSchema\":{\"type\":\"object\"}}]}}\n\n", envelope.ID)
		default:
			if len(envelope.ID) == 0 || string(envelope.ID) == "null" {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{}}`, envelope.ID)
		}
	})

	log.Printf("mock-mcp listening on http://%s/mcp", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
