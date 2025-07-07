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
	messageQueue shared.MessageJobQueue
}

func NewServer(diagnoseFile in.DiagnoseFile, configure in.Configure) *Server {
	messageQueue := shared.NewMessageSerialJobQueue(100)

	server := &Server{
		diagnoseFile: diagnoseFile,
		configure:    configure,
		messageQueue: messageQueue,
	}

	// Register handlers for different topics
	server.registerHandlers()

	return server
}

// registerHandlers registers handlers for different message topics
func (s *Server) registerHandlers() {
	// Unified handler for all diagnosis operations
	s.messageQueue.RegisterHandler("diagnose", s.handleDiagnose)
}

// handleDiagnose processes diagnosis operations based on message type
func (s *Server) handleDiagnose(ctx context.Context, msg shared.Message) {
	notifier := NewContextNotifier(msg.GLSPContext)

	// Generate unique progress token
	token := uuid.New().String()
	NotifyServerWindowWorkDoneProgressCreate(notifier, token)

	var title string
	var results map[string][]domain.Diagnostic
	var err error

	switch msg.Type {
	case shared.DiagnoseFile:
		title = "Diagnosing file..."
		NotifyBeginProgress(notifier, token, title, false)
		NotifyReportProgress(notifier, token, "Diagnosing...", 50)

		// Run diagnosis for single file
		results, err = s.diagnoseFile.Diagnose(ctx, msg.URI)

	case shared.DiagnoseAll:
		title = "Diagnosing all files..."
		NotifyBeginProgress(notifier, token, title, false)
		NotifyReportProgress(notifier, token, "Diagnosing...", 50)

		// Run diagnosis for all files
		results, err = s.diagnoseFile.DiagnoseAll(ctx)

	default:
		NotifyErrorLogMessage(notifier, "Unknown diagnosis type")
		return
	}

	// Handle results
	if err != nil {
		NotifyErrorLogMessage(notifier, "Error during diagnosis: "+err.Error())
	} else {
		for uri, diagnostics := range results {
			NotifyPublishDiagnostics(notifier, uri, diagnostics)
		}
	}

	NotifyEndProgress(notifier, token, "Diagnosis complete")
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
	s.messageQueue.Start(context.Background())
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
	s.messageQueue.Close()
	return nil
}

func (s *Server) onInitialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	// Run diagnostics for all files with progress notification
	message := shared.Message{
		GLSPContext: ctx,
		URI:         "", // Not applicable for "diagnose all"
		Type:        shared.DiagnoseAll,
	}
	s.messageQueue.Enqueue("diagnose", message)

	return nil
}

func (s *Server) onDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	// Run diagnostics for the opened file with progress notification
	message := shared.Message{
		GLSPContext: ctx,
		URI:         uri,
		Type:        shared.DiagnoseFile,
	}
	s.messageQueue.Enqueue("diagnose", message)
	return nil
}

func (s *Server) onDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	// Run diagnostics for the saved file with progress notification
	message := shared.Message{
		GLSPContext: ctx,
		URI:         uri,
		Type:        shared.DiagnoseFile,
	}
	s.messageQueue.Enqueue("diagnose", message)
	return nil
}

func (s *Server) onDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	return nil
}
