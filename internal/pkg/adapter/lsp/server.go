package lsp

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/shared"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

const (
	serverName    = "wpks-ls"
	serverVersion = "0.0.1"
)

// Server represents a minimal LSP server.
type Server struct {
	diagnoseFile in.DiagnoseFile
	configure    in.Configure
	jobQueue     shared.JobQueue
}

func NewServer(diagnoseFile in.DiagnoseFile, configure in.Configure, jobQueue shared.JobQueue) *Server {
	return &Server{
		diagnoseFile: diagnoseFile,
		configure:    configure,
		jobQueue:     jobQueue,
	}
}

// Start runs the LSP server loop.
func (s *Server) Start() error {
	handler := protocol.Handler{
		Initialize:           s.onInitialize,
		Initialized:          s.onInitialized,
		Shutdown:             s.onShutdown,
		TextDocumentDidOpen:  s.onDidOpen,
		TextDocumentDidSave:  s.onDidSave,
		TextDocumentDidClose: s.onDidClose,
	}
	ls := server.NewServer(&handler, serverName, false)
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

	serverVersion := serverVersion
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
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    serverName,
			Version: &serverVersion,
		},
	}, nil
}

func (s *Server) onShutdown(ctx *glsp.Context) error {
	s.jobQueue.Close()
	return nil
}

func (s *Server) onInitialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	done := make(chan struct{})
	s.jobQueue.Enqueue(func() {
		diagnostics, err := s.diagnoseFile.DiagnoseAll()
		if err == nil {
			for uri, diagnostics := range diagnostics {
				lspDiagnostics := MapDiagnostics(diagnostics)
				ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
					URI:         uri,
					Diagnostics: lspDiagnostics,
				})
			}
		}
		close(done)
	})
	<-done
	return nil
}

func (s *Server) onDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	done := make(chan struct{})
	uri := params.TextDocument.URI
	s.jobQueue.Enqueue(func() {
		diagnostics, err := s.diagnoseFile.Diagnose(string(uri))
		if err == nil {
			lspDiagnostics := MapDiagnostics(diagnostics)
			ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
				URI:         params.TextDocument.URI,
				Diagnostics: lspDiagnostics,
			})
		}
		close(done)
	})
	<-done
	return nil
}

func (s *Server) onDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	done := make(chan struct{})
	uri := params.TextDocument.URI
	s.jobQueue.Enqueue(func() {
		diagnostics, err := s.diagnoseFile.Diagnose(string(uri))
		if err == nil {
			lspDiagnostics := MapDiagnostics(diagnostics)
			ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
				URI:         params.TextDocument.URI,
				Diagnostics: lspDiagnostics,
			})
		}
		close(done)
	})
	<-done
	return nil
}

func (s *Server) onDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	return nil
}
