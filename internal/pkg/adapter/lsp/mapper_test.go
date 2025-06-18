package lsp

import (
	"reflect"
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestMapDiagnostics_Empty(t *testing.T) {
	result := MapDiagnostics(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestMapDiagnostics_Single(t *testing.T) {
	d := domain.Diagnostic{
		Range: domain.Range{
			Start: domain.Position{Line: 1, Character: 2},
			End:   domain.Position{Line: 3, Character: 4},
		},
		Severity: domain.SeverityWarning,
		Source:   "test",
		Message:  "msg",
	}
	result := MapDiagnostics([]domain.Diagnostic{d})
	if len(result) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(result))
	}
	lsp := result[0]
	if lsp.Range.Start.Line != 1 || lsp.Range.End.Character != 4 {
		t.Errorf("unexpected range: %+v", lsp.Range)
	}
	if lsp.Severity == nil || *lsp.Severity != protocol.DiagnosticSeverity(domain.SeverityWarning) {
		t.Errorf("unexpected severity: %v", lsp.Severity)
	}
	if lsp.Source == nil || *lsp.Source != "test" {
		t.Errorf("unexpected source: %v", lsp.Source)
	}
	if lsp.Message != "msg" {
		t.Errorf("unexpected message: %v", lsp.Message)
	}
}

func TestMapDiagnostics_Multiple(t *testing.T) {
	ds := []domain.Diagnostic{
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
	}
	result := MapDiagnostics(ds)
	if len(result) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(result))
	}
	if *result[0].Severity != protocol.DiagnosticSeverity(domain.SeverityError) || *result[1].Severity != protocol.DiagnosticSeverity(domain.SeverityHint) {
		t.Errorf("unexpected severities: %v, %v", result[0].Severity, result[1].Severity)
	}
	if *result[0].Source != "a" || *result[1].Source != "b" {
		t.Errorf("unexpected sources: %v, %v", result[0].Source, result[1].Source)
	}
	if !reflect.DeepEqual(result[0].Range, protocol.Range{
		Start: protocol.Position{Line: 1, Character: 1},
		End:   protocol.Position{Line: 1, Character: 2},
	}) {
		t.Errorf("unexpected range: %+v", result[0].Range)
	}
}
