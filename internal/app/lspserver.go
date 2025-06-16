package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// LSPServer represents a minimal LSP server.
type LSPServer struct{}

// Start runs the LSP server loop.
func (s *LSPServer) Start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		headers := make(map[string]string)
		// Read headers
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				headers[parts[0]] = parts[1]
			}
		}
		// Read content
		clen := headers["Content-Length"]
		if clen == "" {
			continue
		}
		var length int
		fmt.Sscanf(clen, "%d", &length)
		content := make([]byte, length)
		_, err := io.ReadFull(reader, content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading content: %v\n", err)
			return
		}
		// Parse JSON-RPC
		var msg map[string]interface{}
		if err := json.Unmarshal(content, &msg); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			continue
		}
		// Handle initialize
		if method, ok := msg["method"].(string); ok && method == "initialize" {
			id := msg["id"]
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"capabilities": map[string]interface{}{},
				},
			}
			respBytes, _ := json.Marshal(resp)
			fmt.Printf("Content-Length: %d\r\n\r\n%s", len(respBytes), respBytes)
			continue
		}
		// Handle textDocument/didOpen
		if method, ok := msg["method"].(string); ok && method == "textDocument/didOpen" {
			params, _ := msg["params"].(map[string]interface{})
			if td, ok := params["textDocument"].(map[string]interface{}); ok {
				uri, _ := td["uri"].(string)
				fmt.Fprintf(os.Stderr, "didOpen received for: %s\n", uri)
				// Run Packwerk diagnostics and print as JSON
				output, err := RunPackwerkCheck(uri)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Packwerk error: %v\n", err)
				}
				violations := ParsePackwerkOutput(output)
				diagnostics := make([]Diagnostic, 0, len(violations))
				for _, v := range violations {
					diagnostics = append(diagnostics, Diagnostic{
						Range: Range{
							Start: Position{Line: v.Line - 1, Character: v.Column - 1},
							End:   Position{Line: v.Line - 1, Character: v.Column - 1},
						},
						Severity: SeverityError,
						Source:   "packwerk",
						Message:  v.Message,
					})
				}
				b, _ := json.Marshal(diagnostics)
				fmt.Fprintf(os.Stderr, "diagnostics: %s\n", b)
			}
			continue
		}
		// Handle textDocument/didChange
		if method, ok := msg["method"].(string); ok && method == "textDocument/didChange" {
			params, _ := msg["params"].(map[string]interface{})
			if td, ok := params["textDocument"].(map[string]interface{}); ok {
				uri, _ := td["uri"].(string)
				fmt.Fprintf(os.Stderr, "didChange received for: %s\n", uri)
				// Run Packwerk diagnostics and print as JSON
				output, err := RunPackwerkCheck(uri)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Packwerk error: %v\n", err)
				}
				violations := ParsePackwerkOutput(output)
				diagnostics := make([]Diagnostic, 0, len(violations))
				for _, v := range violations {
					diagnostics = append(diagnostics, Diagnostic{
						Range: Range{
							Start: Position{Line: v.Line - 1, Character: v.Column - 1},
							End:   Position{Line: v.Line - 1, Character: v.Column - 1},
						},
						Severity: SeverityError,
						Source:   "packwerk",
						Message:  v.Message,
					})
				}
				b, _ := json.Marshal(diagnostics)
				fmt.Fprintf(os.Stderr, "diagnostics: %s\n", b)
			}
			continue
		}
	}
}
