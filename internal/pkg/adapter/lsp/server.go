package lsp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/task"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

const (
	serverName    = "wpks-ls"
	serverVersion = "0.0.1"
)

// DiagnoseType represents the type of diagnosis to perform
type DiagnoseType int

const (
	DiagnoseFile DiagnoseType = iota // Diagnose a single file
	DiagnoseAll                      // Diagnose all files
)

// Message represents a message containing glsp.Context and URI
type Message struct {
	GLSPContext *glsp.Context
	URI         string
	Type        DiagnoseType
	// Additional fields can be added here as needed
}

// Server represents a minimal LSP server.
type Server struct {
	diagnoseFile in.DiagnoseFile
	configure    in.Configure
	messageQueue task.Broker[Message]
}

func NewServer(diagnoseFile in.DiagnoseFile, configure in.Configure) *Server {
	messageQueue := task.NewMessageBroker[Message]()

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
	s.messageQueue.RegisterTopic(
		"diagnose",
		s.handleDiagnose,
		task.WithQueueSize(100),
		task.WithBatchConfig(10, 100*time.Millisecond),
	)
}

// handleDiagnose processes diagnosis operations for multiple messages with batch processing
func (s *Server) handleDiagnose(ctx context.Context, msgs []Message) {
	if len(msgs) == 0 {
		return
	}
	lastMsg := msgs[len(msgs)-1]
	notifier := NewContextNotifier(lastMsg.GLSPContext)

	token := uuid.New().String()
	NotifyServerWindowWorkDoneProgressCreate(notifier, token)

	hasAll := false
	for _, msg := range msgs {
		if msg.Type == DiagnoseAll {
			hasAll = true
			break
		}
	}

	var allResults map[string][]domain.Diagnostic
	var err error

	if hasAll {
		NotifyBeginProgress(notifier, token, "Diagnosing all files...", false)
		NotifyReportProgress(notifier, token, "Diagnosing...", 25)

		allResults, err = s.diagnoseFile.DiagnoseAll(ctx)
	} else {
		NotifyBeginProgress(notifier, token, "Diagnosing files...", false)
		NotifyReportProgress(notifier, token, "Diagnosing...", 25)

		// Collect unique URIs to avoid duplicate processing
		uriSet := make(map[string]struct{})
		for _, msg := range msgs {
			if msg.Type == DiagnoseFile {
				uriSet[msg.URI] = struct{}{}
			}
		}

		uris := make([]string, 0, len(uriSet))
		for uri := range uriSet {
			uris = append(uris, uri)
		}

		allResults, err = s.diagnoseFile.Diagnose(ctx, uris...)
	}

	if err != nil {
		NotifyErrorLogMessage(notifier, "Error during diagnosis: "+err.Error())
	} else {
		NotifyReportProgress(notifier, token, "Diagnosing...", 75)

		for _, msg := range msgs {
			msgNotifier := NewContextNotifier(msg.GLSPContext)
			for uri, diagnostics := range allResults {
				NotifyPublishDiagnostics(msgNotifier, uri, diagnostics)
			}
		}
	}

	NotifyReportProgress(notifier, token, "Diagnosing...", 100)
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
	message := Message{
		GLSPContext: ctx,
		URI:         "", // Not applicable for "diagnose all"
		Type:        DiagnoseAll,
	}
	s.messageQueue.Enqueue("diagnose", message)

	return nil
}

func (s *Server) onDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	// Run diagnostics for the opened file with progress notification
	message := Message{
		GLSPContext: ctx,
		URI:         uri,
		Type:        DiagnoseFile,
	}
	s.messageQueue.Enqueue("diagnose", message)
	return nil
}

func (s *Server) onDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	// Run diagnostics for the saved file with progress notification
	message := Message{
		GLSPContext: ctx,
		URI:         uri,
		Type:        DiagnoseFile,
	}
	s.messageQueue.Enqueue("diagnose", message)
	return nil
}

func (s *Server) onDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	return nil
}
