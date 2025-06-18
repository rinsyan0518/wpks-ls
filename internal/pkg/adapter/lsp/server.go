package lsp

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

// Server represents a minimal LSP server.
type Server struct {
	diagnoseFile in.DiagnoseFile
	configure    in.Configure
}

func NewServer(diagnoseFile in.DiagnoseFile, configure in.Configure) *Server {
	return &Server{
		diagnoseFile: diagnoseFile,
		configure:    configure,
	}
}

// Start runs the LSP server loop.
func (s *Server) Start() {
	handler := protocol.Handler{
		Initialize:          s.onInitialize,
		Initialized:         s.onInitialized,
		TextDocumentDidOpen: s.onDidOpen,
		TextDocumentDidSave: s.onDidSave,
	}
	ls := server.NewServer(&handler, "wpks-ls", false)
	ls.RunStdio()
}

func (s *Server) onInitialize(ctx *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
	openClose := true
	change := protocol.TextDocumentSyncKindIncremental
	save := true

	s.configure.Configure(*params.RootURI, *params.RootPath)

	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: &openClose,
				Change:    &change,
				Save:      &save,
			},
		},
	}, nil
}

func (s *Server) onInitialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) onDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	diagnostics, err := s.diagnoseFile.Diagnose(string(uri))
	if err != nil {
		return err
	}

	lspDiagnostics := make([]protocol.Diagnostic, 0, len(diagnostics))
	for _, d := range diagnostics {
		severity := protocol.DiagnosticSeverity(d.Severity)
		lspDiagnostics = append(lspDiagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: d.Range.Start.Line, Character: d.Range.Start.Character},
				End:   protocol.Position{Line: d.Range.End.Line, Character: d.Range.End.Character},
			},
			Severity: &severity,
			Source:   &d.Source,
			Message:  d.Message,
		})
	}

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: lspDiagnostics,
	})

	return nil
}

func (s *Server) onDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	uri := params.TextDocument.URI

	diagnostics, err := s.diagnoseFile.Diagnose(string(uri))
	if err != nil {
		return err
	}

	lspDiagnostics := make([]protocol.Diagnostic, 0, len(diagnostics))
	for _, d := range diagnostics {
		severity := protocol.DiagnosticSeverity(d.Severity)
		lspDiagnostics = append(lspDiagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: d.Range.Start.Line, Character: d.Range.Start.Character},
				End:   protocol.Position{Line: d.Range.End.Line, Character: d.Range.End.Character},
			},
			Severity: &severity,
			Source:   &d.Source,
			Message:  d.Message,
		})
	}

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: lspDiagnostics,
	})

	return nil
}
