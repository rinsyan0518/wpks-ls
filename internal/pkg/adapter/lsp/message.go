package lsp

import (
	"fmt"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Notifier represents an interface that can send notifications
type Notifier interface {
	Notify(method string, params any)
}

// ContextNotifier wraps glsp.Context to implement the Notifier interface
type ContextNotifier struct {
	ctx *glsp.Context
}

// NewContextNotifier creates a new ContextNotifier
func NewContextNotifier(ctx *glsp.Context) *ContextNotifier {
	return &ContextNotifier{ctx: ctx}
}

// Notify implements the Notifier interface
func (c *ContextNotifier) Notify(method string, params any) {
	c.ctx.Notify(method, params)
}

func NewInitializeResult(serverName string, serverVersion string) protocol.InitializeResult {
	openClose := true
	change := protocol.TextDocumentSyncKindNone
	save := true
	willSave := false
	willSaveWaitUntil := false

	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose:         &openClose,
				Change:            &change,
				Save:              &save,
				WillSave:          &willSave,
				WillSaveWaitUntil: &willSaveWaitUntil,
			},
		},
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    serverName,
			Version: &serverVersion,
		},
	}
}

func NotifyServerWindowWorkDoneProgressCreate(notifier Notifier, token string) {
	progressToken := protocol.ProgressToken{Value: token}

	notifier.Notify(
		protocol.ServerWindowWorkDoneProgressCreate,
		protocol.WorkDoneProgressCreateParams{
			Token: progressToken,
		},
	)
}

func NotifyBeginProgress(notifier Notifier, token string, title string, cancellable bool) {
	progressToken := protocol.ProgressToken{Value: token}

	notifier.Notify(
		protocol.MethodProgress,
		protocol.ProgressParams{
			Token: progressToken,
			Value: &protocol.WorkDoneProgressBegin{
				Kind:        "begin",
				Title:       title,
				Cancellable: &cancellable,
			},
		},
	)
}

func NotifyReportProgress(notifier Notifier, token string, message string, percentage uint32) {
	progressToken := protocol.ProgressToken{Value: token}

	notifier.Notify(
		protocol.MethodProgress,
		protocol.ProgressParams{
			Token: progressToken,
			Value: &protocol.WorkDoneProgressReport{
				Kind:       "report",
				Message:    &message,
				Percentage: &percentage,
			},
		},
	)
}

func NotifyEndProgress(notifier Notifier, token string, message string) {
	progressToken := protocol.ProgressToken{Value: token}

	notifier.Notify(
		protocol.MethodProgress,
		protocol.ProgressParams{
			Token: progressToken,
			Value: &protocol.WorkDoneProgressEnd{
				Kind:    "end",
				Message: &message,
			},
		},
	)
}

func NotifyPublishDiagnostics(notifier Notifier, uri string, diagnostics []domain.Diagnostic) {
	lspDiagnostics := MapDiagnostics(diagnostics)

	notifier.Notify(
		protocol.ServerTextDocumentPublishDiagnostics,
		protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentUri(uri),
			Diagnostics: lspDiagnostics,
		},
	)
}

func NotifyErrorLogMessage(notifier Notifier, format string, args ...any) {
	notifier.Notify(
		protocol.ServerWindowLogMessage,
		protocol.LogMessageParams{
			Message: fmt.Sprintf(format, args...),
			Type:    protocol.MessageTypeError,
		},
	)
}

func NotifyWarningLogMessage(notifier Notifier, format string, args ...any) {
	notifier.Notify(
		protocol.ServerWindowLogMessage,
		protocol.LogMessageParams{
			Message: fmt.Sprintf(format, args...),
			Type:    protocol.MessageTypeWarning,
		},
	)
}

func NotifyInfoLogMessage(notifier Notifier, format string, args ...any) {
	notifier.Notify(
		protocol.ServerWindowLogMessage,
		protocol.LogMessageParams{
			Message: fmt.Sprintf(format, args...),
			Type:    protocol.MessageTypeInfo,
		},
	)
}

func NotifyLogMessage(notifier Notifier, format string, args ...any) {
	notifier.Notify(
		protocol.ServerWindowLogMessage,
		protocol.LogMessageParams{
			Message: fmt.Sprintf(format, args...),
			Type:    protocol.MessageTypeLog,
		},
	)
}
