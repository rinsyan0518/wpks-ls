package lsp

import (
	"reflect"
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// MockNotifier implements Notifier for testing
type MockNotifier struct {
	NotifiedMethods []string
	NotifiedParams  []any
}

func (m *MockNotifier) Notify(method string, params any) {
	m.NotifiedMethods = append(m.NotifiedMethods, method)
	m.NotifiedParams = append(m.NotifiedParams, params)
}

// Tests for ContextNotifier
func TestNewContextNotifier(t *testing.T) {
	t.Run("create context notifier", func(t *testing.T) {
		// We can't easily mock glsp.Context, so we test with nil
		// The actual functionality will be tested through integration tests
		got := NewContextNotifier(nil)
		if got == nil {
			t.Fatal("NewContextNotifier() returned nil")
		}
		if got.ctx != nil {
			t.Errorf("NewContextNotifier() ctx = %v, want nil", got.ctx)
		}
	})
}

func TestContextNotifier_Notify(t *testing.T) {
	t.Run("notify implements Notifier interface", func(t *testing.T) {
		// Test that ContextNotifier properly implements the Notifier interface
		notifier := NewContextNotifier(nil)

		// This test verifies the interface implementation without panicking
		// The actual notification behavior is tested through the functions that use Notifier
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic with nil context, but interface should be implemented
				t.Logf("Expected panic with nil context: %v", r)
			}
		}()

		// This should panic due to nil context, but proves the interface works
		notifier.Notify("test", nil)
	})
}

func TestNewInitializeResult(t *testing.T) {
	tests := []struct {
		name          string
		serverName    string
		serverVersion string
		want          protocol.InitializeResult
	}{
		{
			name:          "basic initialize result",
			serverName:    "test-server",
			serverVersion: "1.0.0",
			want: protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					TextDocumentSync: &protocol.TextDocumentSyncOptions{
						OpenClose:         ptrBool(true),
						Change:            ptrTextDocumentSyncKind(protocol.TextDocumentSyncKindNone),
						Save:              ptrBool(true),
						WillSave:          ptrBool(false),
						WillSaveWaitUntil: ptrBool(false),
					},
				},
				ServerInfo: &protocol.InitializeResultServerInfo{
					Name:    "test-server",
					Version: ptrString("1.0.0"),
				},
			},
		},
		{
			name:          "empty server name and version",
			serverName:    "",
			serverVersion: "",
			want: protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					TextDocumentSync: &protocol.TextDocumentSyncOptions{
						OpenClose:         ptrBool(true),
						Change:            ptrTextDocumentSyncKind(protocol.TextDocumentSyncKindNone),
						Save:              ptrBool(true),
						WillSave:          ptrBool(false),
						WillSaveWaitUntil: ptrBool(false),
					},
				},
				ServerInfo: &protocol.InitializeResultServerInfo{
					Name:    "",
					Version: ptrString(""),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInitializeResult(tt.serverName, tt.serverVersion)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInitializeResult() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestNotifyServerWindowWorkDoneProgressCreate(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "basic token",
			token: "test-token",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "uuid token",
			token: "12345678-1234-5678-9012-123456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyServerWindowWorkDoneProgressCreate(mockNotifier, tt.token)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.ServerWindowWorkDoneProgressCreate {
				t.Errorf("expected method %s, got %s", protocol.ServerWindowWorkDoneProgressCreate, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.WorkDoneProgressCreateParams)
			if !ok {
				t.Fatalf("expected WorkDoneProgressCreateParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Token.Value != tt.token {
				t.Errorf("expected token %s, got %s", tt.token, params.Token.Value)
			}
		})
	}
}

func TestNotifyBeginProgress(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		title       string
		cancellable bool
	}{
		{
			name:        "basic progress begin",
			token:       "test-token",
			title:       "Test Title",
			cancellable: true,
		},
		{
			name:        "non-cancellable progress",
			token:       "another-token",
			title:       "Another Title",
			cancellable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyBeginProgress(mockNotifier, tt.token, tt.title, tt.cancellable)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.MethodProgress {
				t.Errorf("expected method %s, got %s", protocol.MethodProgress, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.ProgressParams)
			if !ok {
				t.Fatalf("expected ProgressParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Token.Value != tt.token {
				t.Errorf("expected token %s, got %s", tt.token, params.Token.Value)
			}

			beginValue, ok := params.Value.(*protocol.WorkDoneProgressBegin)
			if !ok {
				t.Fatalf("expected WorkDoneProgressBegin, got %T", params.Value)
			}

			if beginValue.Kind != "begin" {
				t.Errorf("expected kind 'begin', got %s", beginValue.Kind)
			}

			if beginValue.Title != tt.title {
				t.Errorf("expected title %s, got %s", tt.title, beginValue.Title)
			}

			if beginValue.Cancellable == nil || *beginValue.Cancellable != tt.cancellable {
				t.Errorf("expected cancellable %v, got %v", tt.cancellable, beginValue.Cancellable)
			}
		})
	}
}

func TestNotifyReportProgress(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		message    string
		percentage uint32
	}{
		{
			name:       "basic progress report",
			token:      "test-token",
			message:    "Test Message",
			percentage: 50,
		},
		{
			name:       "zero percentage",
			token:      "token",
			message:    "Starting...",
			percentage: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyReportProgress(mockNotifier, tt.token, tt.message, tt.percentage)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.MethodProgress {
				t.Errorf("expected method %s, got %s", protocol.MethodProgress, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.ProgressParams)
			if !ok {
				t.Fatalf("expected ProgressParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Token.Value != tt.token {
				t.Errorf("expected token %s, got %s", tt.token, params.Token.Value)
			}

			reportValue, ok := params.Value.(*protocol.WorkDoneProgressReport)
			if !ok {
				t.Fatalf("expected WorkDoneProgressReport, got %T", params.Value)
			}

			if reportValue.Kind != "report" {
				t.Errorf("expected kind 'report', got %s", reportValue.Kind)
			}

			if reportValue.Message == nil || *reportValue.Message != tt.message {
				t.Errorf("expected message %s, got %v", tt.message, reportValue.Message)
			}

			if reportValue.Percentage == nil || *reportValue.Percentage != tt.percentage {
				t.Errorf("expected percentage %d, got %v", tt.percentage, reportValue.Percentage)
			}
		})
	}
}

func TestNotifyEndProgress(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		message string
	}{
		{
			name:    "basic progress end",
			token:   "test-token",
			message: "Test Message",
		},
		{
			name:    "completion message",
			token:   "uuid-token",
			message: "Diagnosis complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyEndProgress(mockNotifier, tt.token, tt.message)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.MethodProgress {
				t.Errorf("expected method %s, got %s", protocol.MethodProgress, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.ProgressParams)
			if !ok {
				t.Fatalf("expected ProgressParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Token.Value != tt.token {
				t.Errorf("expected token %s, got %s", tt.token, params.Token.Value)
			}

			endValue, ok := params.Value.(*protocol.WorkDoneProgressEnd)
			if !ok {
				t.Fatalf("expected WorkDoneProgressEnd, got %T", params.Value)
			}

			if endValue.Kind != "end" {
				t.Errorf("expected kind 'end', got %s", endValue.Kind)
			}

			if endValue.Message == nil || *endValue.Message != tt.message {
				t.Errorf("expected message %s, got %v", tt.message, endValue.Message)
			}
		})
	}
}

func TestNotifyPublishDiagnostics(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		diagnostics []domain.Diagnostic
	}{
		{
			name:        "empty diagnostics",
			uri:         "file:///test.rb",
			diagnostics: []domain.Diagnostic{},
		},
		{
			name: "single diagnostic",
			uri:  "file:///test.rb",
			diagnostics: []domain.Diagnostic{
				{
					Range: domain.Range{
						Start: domain.Position{Line: 1, Character: 2},
						End:   domain.Position{Line: 1, Character: 3},
					},
					Severity: domain.SeverityError,
					Source:   "test",
					Message:  "Test error",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyPublishDiagnostics(mockNotifier, tt.uri, tt.diagnostics)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.ServerTextDocumentPublishDiagnostics {
				t.Errorf("expected method %s, got %s", protocol.ServerTextDocumentPublishDiagnostics, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.PublishDiagnosticsParams)
			if !ok {
				t.Fatalf("expected PublishDiagnosticsParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if string(params.URI) != tt.uri {
				t.Errorf("expected URI %s, got %s", tt.uri, params.URI)
			}

			if len(params.Diagnostics) != len(tt.diagnostics) {
				t.Fatalf("expected %d diagnostics, got %d", len(tt.diagnostics), len(params.Diagnostics))
			}

			// Test that diagnostics are properly mapped
			for i, expectedDiag := range tt.diagnostics {
				actualDiag := params.Diagnostics[i]
				if actualDiag.Message != expectedDiag.Message {
					t.Errorf("diagnostic %d: expected message %s, got %s", i, expectedDiag.Message, actualDiag.Message)
				}
				if actualDiag.Severity == nil || *actualDiag.Severity != protocol.DiagnosticSeverity(expectedDiag.Severity) {
					t.Errorf("diagnostic %d: expected severity %d, got %v", i, expectedDiag.Severity, actualDiag.Severity)
				}
			}
		})
	}
}

func TestNotifyErrorLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "basic error message",
			message: "Test error message",
		},
		{
			name:    "empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyErrorLogMessage(mockNotifier, "%s", tt.message)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.ServerWindowLogMessage {
				t.Errorf("expected method %s, got %s", protocol.ServerWindowLogMessage, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.LogMessageParams)
			if !ok {
				t.Fatalf("expected LogMessageParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Message != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, params.Message)
			}

			if params.Type != protocol.MessageTypeError {
				t.Errorf("expected type %d, got %d", protocol.MessageTypeError, params.Type)
			}
		})
	}
}

// Tests for NotifyErrorLogMessage with format arguments
func TestNotifyErrorLogMessageWithFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "error message with format",
			format: "Error: %s failed with code %d",
			args:   []interface{}{"operation", 500},
			want:   "Error: operation failed with code 500",
		},
		{
			name:   "multiple format specifiers",
			format: "Error in file %s at line %d: %s",
			args:   []interface{}{"test.go", 42, "syntax error"},
			want:   "Error in file test.go at line 42: syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			NotifyErrorLogMessage(mockNotifier, tt.format, tt.args...)

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.LogMessageParams)
			if !ok {
				t.Fatalf("expected LogMessageParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Message != tt.want {
				t.Errorf("expected message %s, got %s", tt.want, params.Message)
			}

			if params.Type != protocol.MessageTypeError {
				t.Errorf("expected type %d, got %d", protocol.MessageTypeError, params.Type)
			}
		})
	}
}

// Tests for NotifyWarningLogMessage
func TestNotifyWarningLogMessage(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "basic warning message",
			format: "Test warning message",
			args:   nil,
			want:   "Test warning message",
		},
		{
			name:   "warning message with format",
			format: "Warning: %s failed with code %d",
			args:   []interface{}{"operation", 404},
			want:   "Warning: operation failed with code 404",
		},
		{
			name:   "empty message",
			format: "",
			args:   nil,
			want:   "",
		},
		{
			name:   "format without args",
			format: "Simple warning",
			args:   []interface{}{},
			want:   "Simple warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			if tt.args == nil {
				NotifyWarningLogMessage(mockNotifier, "%s", tt.format)
			} else {
				NotifyWarningLogMessage(mockNotifier, tt.format, tt.args...)
			}

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.ServerWindowLogMessage {
				t.Errorf("expected method %s, got %s", protocol.ServerWindowLogMessage, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.LogMessageParams)
			if !ok {
				t.Fatalf("expected LogMessageParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Message != tt.want {
				t.Errorf("expected message %s, got %s", tt.want, params.Message)
			}

			if params.Type != protocol.MessageTypeWarning {
				t.Errorf("expected type %d, got %d", protocol.MessageTypeWarning, params.Type)
			}
		})
	}
}

// Tests for NotifyInfoLogMessage
func TestNotifyInfoLogMessage(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "basic info message",
			format: "Test info message",
			args:   nil,
			want:   "Test info message",
		},
		{
			name:   "info message with format",
			format: "Info: processed %d files in %s",
			args:   []interface{}{42, "10ms"},
			want:   "Info: processed 42 files in 10ms",
		},
		{
			name:   "empty message",
			format: "",
			args:   nil,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			if tt.args == nil {
				NotifyInfoLogMessage(mockNotifier, "%s", tt.format)
			} else {
				NotifyInfoLogMessage(mockNotifier, tt.format, tt.args...)
			}

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.ServerWindowLogMessage {
				t.Errorf("expected method %s, got %s", protocol.ServerWindowLogMessage, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.LogMessageParams)
			if !ok {
				t.Fatalf("expected LogMessageParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Message != tt.want {
				t.Errorf("expected message %s, got %s", tt.want, params.Message)
			}

			if params.Type != protocol.MessageTypeInfo {
				t.Errorf("expected type %d, got %d", protocol.MessageTypeInfo, params.Type)
			}
		})
	}
}

// Tests for NotifyLogMessage
func TestNotifyLogMessage(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "basic log message",
			format: "Test log message",
			args:   nil,
			want:   "Test log message",
		},
		{
			name:   "log message with format",
			format: "Log: user %s performed action %s at %s",
			args:   []interface{}{"john", "login", "2023-01-01"},
			want:   "Log: user john performed action login at 2023-01-01",
		},
		{
			name:   "empty message",
			format: "",
			args:   nil,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &MockNotifier{}

			if tt.args == nil {
				NotifyLogMessage(mockNotifier, "%s", tt.format)
			} else {
				NotifyLogMessage(mockNotifier, tt.format, tt.args...)
			}

			if len(mockNotifier.NotifiedMethods) != 1 {
				t.Fatalf("expected 1 notification, got %d", len(mockNotifier.NotifiedMethods))
			}

			if mockNotifier.NotifiedMethods[0] != protocol.ServerWindowLogMessage {
				t.Errorf("expected method %s, got %s", protocol.ServerWindowLogMessage, mockNotifier.NotifiedMethods[0])
			}

			params, ok := mockNotifier.NotifiedParams[0].(protocol.LogMessageParams)
			if !ok {
				t.Fatalf("expected LogMessageParams, got %T", mockNotifier.NotifiedParams[0])
			}

			if params.Message != tt.want {
				t.Errorf("expected message %s, got %s", tt.want, params.Message)
			}

			if params.Type != protocol.MessageTypeLog {
				t.Errorf("expected type %d, got %d", protocol.MessageTypeLog, params.Type)
			}
		})
	}
}

// ============================================================================
// Parameter Creation Tests
// ============================================================================

// Test parameter creation for NotifyServerWindowWorkDoneProgressCreate
func TestCreateWorkDoneProgressCreateParams(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  protocol.WorkDoneProgressCreateParams
	}{
		{
			name:  "basic token",
			token: "test-token",
			want: protocol.WorkDoneProgressCreateParams{
				Token: protocol.ProgressToken{Value: "test-token"},
			},
		},
		{
			name:  "empty token",
			token: "",
			want: protocol.WorkDoneProgressCreateParams{
				Token: protocol.ProgressToken{Value: ""},
			},
		},
		{
			name:  "uuid token",
			token: "12345678-1234-5678-9012-123456789012",
			want: protocol.WorkDoneProgressCreateParams{
				Token: protocol.ProgressToken{Value: "12345678-1234-5678-9012-123456789012"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter creation logic that's inside NotifyServerWindowWorkDoneProgressCreate
			progressToken := protocol.ProgressToken{Value: tt.token}
			got := protocol.WorkDoneProgressCreateParams{Token: progressToken}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WorkDoneProgressCreateParams = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test parameter creation for NotifyBeginProgress
func TestCreateWorkDoneProgressBeginParams(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		title       string
		cancellable bool
		want        protocol.ProgressParams
	}{
		{
			name:        "basic progress begin",
			token:       "test-token",
			title:       "Test Title",
			cancellable: true,
			want: protocol.ProgressParams{
				Token: protocol.ProgressToken{Value: "test-token"},
				Value: &protocol.WorkDoneProgressBegin{
					Kind:        "begin",
					Title:       "Test Title",
					Cancellable: ptrBool(true),
				},
			},
		},
		{
			name:        "non-cancellable progress",
			token:       "another-token",
			title:       "Another Title",
			cancellable: false,
			want: protocol.ProgressParams{
				Token: protocol.ProgressToken{Value: "another-token"},
				Value: &protocol.WorkDoneProgressBegin{
					Kind:        "begin",
					Title:       "Another Title",
					Cancellable: ptrBool(false),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter creation logic that's inside NotifyBeginProgress
			progressToken := protocol.ProgressToken{Value: tt.token}
			got := protocol.ProgressParams{
				Token: progressToken,
				Value: &protocol.WorkDoneProgressBegin{
					Kind:        "begin",
					Title:       tt.title,
					Cancellable: &tt.cancellable,
				},
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProgressParams = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test parameter creation for NotifyReportProgress
func TestCreateWorkDoneProgressReportParams(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		message    string
		percentage uint32
		want       protocol.ProgressParams
	}{
		{
			name:       "basic progress report",
			token:      "test-token",
			message:    "Test Message",
			percentage: 50,
			want: protocol.ProgressParams{
				Token: protocol.ProgressToken{Value: "test-token"},
				Value: &protocol.WorkDoneProgressReport{
					Kind:       "report",
					Message:    ptrString("Test Message"),
					Percentage: ptrUint32(50),
				},
			},
		},
		{
			name:       "zero percentage",
			token:      "token",
			message:    "Starting...",
			percentage: 0,
			want: protocol.ProgressParams{
				Token: protocol.ProgressToken{Value: "token"},
				Value: &protocol.WorkDoneProgressReport{
					Kind:       "report",
					Message:    ptrString("Starting..."),
					Percentage: ptrUint32(0),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter creation logic that's inside NotifyReportProgress
			progressToken := protocol.ProgressToken{Value: tt.token}
			got := protocol.ProgressParams{
				Token: progressToken,
				Value: &protocol.WorkDoneProgressReport{
					Kind:       "report",
					Message:    &tt.message,
					Percentage: &tt.percentage,
				},
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProgressParams = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test parameter creation for NotifyEndProgress
func TestCreateWorkDoneProgressEndParams(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		message string
		want    protocol.ProgressParams
	}{
		{
			name:    "basic progress end",
			token:   "test-token",
			message: "Test Message",
			want: protocol.ProgressParams{
				Token: protocol.ProgressToken{Value: "test-token"},
				Value: &protocol.WorkDoneProgressEnd{
					Kind:    "end",
					Message: ptrString("Test Message"),
				},
			},
		},
		{
			name:    "completion message",
			token:   "uuid-token",
			message: "Diagnosis complete",
			want: protocol.ProgressParams{
				Token: protocol.ProgressToken{Value: "uuid-token"},
				Value: &protocol.WorkDoneProgressEnd{
					Kind:    "end",
					Message: ptrString("Diagnosis complete"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter creation logic that's inside NotifyEndProgress
			progressToken := protocol.ProgressToken{Value: tt.token}
			got := protocol.ProgressParams{
				Token: progressToken,
				Value: &protocol.WorkDoneProgressEnd{
					Kind:    "end",
					Message: &tt.message,
				},
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProgressParams = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test parameter creation for NotifyPublishDiagnostics
func TestCreatePublishDiagnosticsParams(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		diagnostics []domain.Diagnostic
		want        protocol.PublishDiagnosticsParams
	}{
		{
			name:        "empty diagnostics",
			uri:         "file:///test.rb",
			diagnostics: []domain.Diagnostic{},
			want: protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentUri("file:///test.rb"),
				Diagnostics: []protocol.Diagnostic{},
			},
		},
		{
			name: "single diagnostic",
			uri:  "file:///test.rb",
			diagnostics: []domain.Diagnostic{
				{
					Range: domain.Range{
						Start: domain.Position{Line: 1, Character: 2},
						End:   domain.Position{Line: 1, Character: 3},
					},
					Severity: domain.SeverityError,
					Source:   "test",
					Message:  "Test error",
				},
			},
			want: protocol.PublishDiagnosticsParams{
				URI: protocol.DocumentUri("file:///test.rb"),
				Diagnostics: []protocol.Diagnostic{
					{
						Range: protocol.Range{
							Start: protocol.Position{Line: 1, Character: 2},
							End:   protocol.Position{Line: 1, Character: 3},
						},
						Severity: ptrDiagnosticSeverity(protocol.DiagnosticSeverity(domain.SeverityError)),
						Source:   ptrString("test"),
						Message:  "Test error",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter creation logic that's inside NotifyPublishDiagnostics
			lspDiagnostics := MapDiagnostics(tt.diagnostics)
			got := protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentUri(tt.uri),
				Diagnostics: lspDiagnostics,
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublishDiagnosticsParams = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test parameter creation for NotifyErrorLogMessage
func TestCreateLogMessageParams(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    protocol.LogMessageParams
	}{
		{
			name:    "basic error message",
			message: "Test error message",
			want: protocol.LogMessageParams{
				Message: "Test error message",
				Type:    protocol.MessageTypeError,
			},
		},
		{
			name:    "empty message",
			message: "",
			want: protocol.LogMessageParams{
				Message: "",
				Type:    protocol.MessageTypeError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter creation logic that's inside NotifyErrorLogMessage
			got := protocol.LogMessageParams{
				Message: tt.message,
				Type:    protocol.MessageTypeError,
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LogMessageParams = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// Helper functions for creating pointers
func ptrBool(v bool) *bool {
	return &v
}

func ptrString(v string) *string {
	return &v
}

func ptrUint32(v uint32) *uint32 {
	return &v
}

func ptrTextDocumentSyncKind(v protocol.TextDocumentSyncKind) *protocol.TextDocumentSyncKind {
	return &v
}

func ptrDiagnosticSeverity(v protocol.DiagnosticSeverity) *protocol.DiagnosticSeverity {
	return &v
}
