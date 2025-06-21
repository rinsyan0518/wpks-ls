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
func (s *Server) Start() error {
	handler := protocol.Handler{
		Initialize:          s.onInitialize,
		Initialized:         s.onInitialized,
		TextDocumentDidOpen: s.onDidOpen,
		TextDocumentDidSave: s.onDidSave,
	}
	ls := server.NewServer(&handler, "wpks-ls", false)
	return ls.RunStdio()
}

func (s *Server) onInitialize(ctx *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
	checkAllOnInitialized := false
	if params.InitializationOptions != nil {
		if options, ok := params.InitializationOptions.(map[string]any); ok {
			if checkAll, ok := options["checkAllOnInitialized"].(bool); ok {
				checkAllOnInitialized = checkAll
			}
		}
	}

	err := s.configure.Configure(*params.RootURI, *params.RootPath, checkAllOnInitialized)
	if err != nil {
		return nil, err
	}

	openClose := true
	change := protocol.TextDocumentSyncKindIncremental
	save := true

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
	diagnostics, err := s.diagnoseFile.DiagnoseAll()
	if err != nil {
		return err
	}

	for uri, diagnostics := range diagnostics {
		lspDiagnostics := MapDiagnostics(diagnostics)
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: lspDiagnostics,
		})
	}

	return nil
}

func (s *Server) onDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	diagnostics, err := s.diagnoseFile.Diagnose(string(uri))
	if err != nil {
		return err
	}

	lspDiagnostics := MapDiagnostics(diagnostics)

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

	lspDiagnostics := MapDiagnostics(diagnostics)

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: lspDiagnostics,
	})

	return nil
}
