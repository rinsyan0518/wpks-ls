package lsp

import (
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
	// Run diagnostics for all files with progress notification
	s.runWithDiagnoseProgress(
		ctx,
		"diagnoseAll",
		"Diagnosing all files...",
		func() (map[string][]domain.Diagnostic, error) {
			return s.diagnoseFile.DiagnoseAll()
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
		func() (map[string][]domain.Diagnostic, error) {
			diags, err := s.diagnoseFile.Diagnose(uri)
			return map[string][]domain.Diagnostic{uri: diags}, err
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
		func() (map[string][]domain.Diagnostic, error) {
			diags, err := s.diagnoseFile.Diagnose(uri)
			return map[string][]domain.Diagnostic{uri: diags}, err
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
	ctx *glsp.Context,
	key string,
	title string,
	diagnoseFunc func() (map[string][]domain.Diagnostic, error),
) {
	s.jobQueue.Enqueue(key, func() {
		// Generate a unique progress token for this operation
		token := protocol.ProgressToken{Value: uuid.New().String()}
		// Request the client to create a progress token
		ctx.Notify(protocol.ServerWindowWorkDoneProgressCreate, protocol.WorkDoneProgressCreateParams{Token: token})

		// Notify the client that the progress has begun
		cancellable := false
		ctx.Notify(protocol.MethodProgress, &protocol.ProgressParams{
			Token: token,
			Value: &protocol.WorkDoneProgressBegin{
				Kind:        "begin",
				Title:       title,
				Cancellable: &cancellable,
			},
		})

		// Optionally notify intermediate progress (e.g., 50%)
		message := "Diagnosing..."
		percentage := protocol.UInteger(50)
		ctx.Notify(protocol.MethodProgress, &protocol.ProgressParams{
			Token: token,
			Value: &protocol.WorkDoneProgressReport{
				Kind:       "report",
				Message:    &message,
				Percentage: &percentage,
			},
		})

		// Run the actual diagnose function
		results, err := diagnoseFunc()
		if err == nil {
			for uri, diagnostics := range results {
				lspDiagnostics := MapDiagnostics(diagnostics)
				ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
					URI:         protocol.DocumentUri(uri),
					Diagnostics: lspDiagnostics,
				})
			}
		} else {
			ctx.Notify(protocol.ServerWindowLogMessage, &protocol.LogMessageParams{
				Type:    protocol.MessageTypeError,
				Message: "Error diagnosing file: " + err.Error(),
			})
		}

		// Notify the client that the progress has ended
		message = "Diagnosis complete"
		ctx.Notify(protocol.MethodProgress, &protocol.ProgressParams{
			Token: token,
			Value: &protocol.WorkDoneProgressEnd{
				Kind:    "end",
				Message: &message,
			},
		})
	})
}
