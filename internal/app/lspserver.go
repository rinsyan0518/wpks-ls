package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter"
	portout "github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/usecase"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

// LSPServer represents a minimal LSP server.
type LSPServer struct {
	PackwerkRunner       portout.PackwerkRunner
	DiagnosticsPublisher portout.DiagnosticsPublisher
}

func NewLSPServer(runner portout.PackwerkRunner, publisher portout.DiagnosticsPublisher) *LSPServer {
	return &LSPServer{
		PackwerkRunner:       runner,
		DiagnosticsPublisher: publisher,
	}
}

// Start runs the LSP server loop.
func (s *LSPServer) Start() {
	handler := protocol.Handler{
		Initialize: func(ctx *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
			return protocol.InitializeResult{Capabilities: protocol.ServerCapabilities{}}, nil
		},
		TextDocumentDidOpen: func(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
			uri := params.TextDocument.URI
			fmt.Fprintf(os.Stderr, "didOpen received for: %s\n", uri)
			output, err := s.PackwerkRunner.RunCheck(string(uri))
			violations := adapter.ParsePackwerkOutput(output)
			diagnostics := usecase.GenerateDiagnostics(violations, err)
			b, _ := json.Marshal(diagnostics)
			fmt.Fprintf(os.Stderr, "diagnostics: %s\n", b)
			s.DiagnosticsPublisher.Publish(string(uri), diagnostics)
			return nil
		},
		TextDocumentDidChange: func(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
			uri := params.TextDocument.URI
			fmt.Fprintf(os.Stderr, "didChange received for: %s\n", uri)
			output, err := s.PackwerkRunner.RunCheck(string(uri))
			violations := adapter.ParsePackwerkOutput(output)
			diagnostics := usecase.GenerateDiagnostics(violations, err)
			b, _ := json.Marshal(diagnostics)
			fmt.Fprintf(os.Stderr, "diagnostics: %s\n", b)
			s.DiagnosticsPublisher.Publish(string(uri), diagnostics)
			return nil
		},
	}
	ls := server.NewServer(&handler, "wpks-ls", false)
	ls.RunStdio()
}
