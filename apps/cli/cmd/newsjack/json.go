package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func writeJSON(w io.Writer, payload any) {
	w.Write(marshalJSON(payload))
}

func writeJSONCompact(w io.Writer, payload any) {
	data, _ := json.Marshal(payload)
	w.Write(append(data, '\n'))
}

func marshalJSON(payload any) []byte {
	data, _ := json.MarshalIndent(payload, "", "  ")
	return append(data, '\n')
}

func readJSONMap(path string) (map[string]any, error) {
	payload, err := readJSONAny(path)
	if err != nil {
		return nil, err
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain a JSON object", path)
	}
	return m, nil
}

func readJSONAny(path string) (any, error) {
	data, err := os.ReadFile(expandPath(path))
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
