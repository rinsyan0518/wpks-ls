package lsp

import (
	"context"

	"github.com/google/uuid"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
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
	jobQueue     shared.KeyedJobQueue
}

func NewServer(diagnoseFile in.DiagnoseFile, configure in.Configure, jobQueue shared.KeyedJobQueue) *Server {
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
	s.jobQueue.Start(context.Background())
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

	return NewInitializeResult(serverName, serverVersion), nil
}

func (s *Server) onShutdown(ctx *glsp.Context) error {
	s.jobQueue.Close()
	return nil
}

func (s *Server) onInitialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	// Run diagnostics for all files with progress notification
	s.runWithDiagnoseProgress(
		ctx,
		"diagnoseAll",
		"Diagnosing all files...",
		func(ctx context.Context) (map[string][]domain.Diagnostic, error) {
			return s.diagnoseFile.DiagnoseAll(ctx)
		},
	)

	return nil
}

func (s *Server) onDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	// Run diagnostics for the opened file with progress notification
	s.runWithDiagnoseProgress(
		ctx,
		uri,
		"Diagnosing file...",
		func(ctx context.Context) (map[string][]domain.Diagnostic, error) {
			return s.diagnoseFile.Diagnose(ctx, uri)
		},
	)
	return nil
}

func (s *Server) onDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	// Run diagnostics for the saved file with progress notification
	s.runWithDiagnoseProgress(
		ctx,
		uri,
		"Diagnosing file...",
		func(ctx context.Context) (map[string][]domain.Diagnostic, error) {
			return s.diagnoseFile.Diagnose(ctx, uri)
		},
	)
	return nil
}

func (s *Server) onDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	return nil
}

// runWithDiagnoseProgress executes diagnostics with LSP progress notification.
// This function is specialized for running diagnose operations with progress reporting.
func (s *Server) runWithDiagnoseProgress(
	glspCtx *glsp.Context,
	key string,
	title string,
	diagnoseFunc func(ctx context.Context) (map[string][]domain.Diagnostic, error),
) {
	s.jobQueue.Enqueue(key, func(ctx context.Context) {
		// Create a notifier from the glsp context
		notifier := NewContextNotifier(glspCtx)

		// Generate a unique progress token for this operation
		token := uuid.New().String()
		// Request the client to create a progress token
		NotifyServerWindowWorkDoneProgressCreate(notifier, token)

		// Notify the client that the progress has begun
		NotifyBeginProgress(notifier, token, title, false)

		// Optionally notify intermediate progress (e.g., 50%)
		NotifyReportProgress(notifier, token, "Diagnosing...", 50)

		// Run the actual diagnose function
		results, err := diagnoseFunc(ctx)
		if err == nil {
			for uri, diagnostics := range results {
				NotifyPublishDiagnostics(notifier, uri, diagnostics)
			}
		} else {
			NotifyErrorLogMessage(notifier, "Error diagnosing file: "+err.Error())
		}

		// Notify the client that the progress has ended
		NotifyEndProgress(notifier, token, "Diagnosis complete")
	})
}
