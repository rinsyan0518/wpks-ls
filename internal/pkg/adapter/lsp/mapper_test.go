package lsp

import (
	"reflect"
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestMapDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input []domain.Diagnostic
		want  []protocol.Diagnostic
	}{
		{
			name:  "empty input",
			input: nil,
			want:  []protocol.Diagnostic{},
		},
		{
			name: "single diagnostic",
			input: []domain.Diagnostic{
				{
					Range: domain.Range{
						Start: domain.Position{Line: 1, Character: 2},
						End:   domain.Position{Line: 3, Character: 4},
					},
					Severity: domain.SeverityWarning,
					Source:   "test",
					Message:  "msg",
				},
			},
			want: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 3, Character: 4},
					},
					Severity: Ptr(protocol.DiagnosticSeverity(domain.SeverityWarning)),
					Source:   Ptr("test"),
					Message:  "msg",
				},
			},
		},
		{
			name: "multiple diagnostics",
			input: []domain.Diagnostic{
				{
					Range: domain.Range{
						Start: domain.Position{Line: 1, Character: 1},
						End:   domain.Position{Line: 1, Character: 2},
					},
					Severity: domain.SeverityError,
					Source:   "a",
					Message:  "err",
				},
				{
					Range: domain.Range{
						Start: domain.Position{Line: 2, Character: 2},
						End:   domain.Position{Line: 2, Character: 3},
					},
					Severity: domain.SeverityHint,
					Source:   "b",
					Message:  "hint",
				},
			},
			want: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 1},
						End:   protocol.Position{Line: 1, Character: 2},
					},
					Severity: Ptr(protocol.DiagnosticSeverity(domain.SeverityError)),
					Source:   Ptr("a"),
					Message:  "err",
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 3},
					},
					Severity: Ptr(protocol.DiagnosticSeverity(domain.SeverityHint)),
					Source:   Ptr("b"),
					Message:  "hint",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapDiagnostics(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want %+v, got %+v", tt.want, got)
			}
		})
	}
}

func Ptr[T any](v T) *T {
	return &v
}
