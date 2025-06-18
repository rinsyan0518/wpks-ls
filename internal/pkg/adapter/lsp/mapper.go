package lsp

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func MapDiagnostics(diags []domain.Diagnostic) []protocol.Diagnostic {
	lspDiagnostics := make([]protocol.Diagnostic, 0, len(diags))
	for _, d := range diags {
		severity := protocol.DiagnosticSeverity(d.Severity)
		lspDiagnostics = append(lspDiagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: d.Range.Start.Line, Character: d.Range.Start.Character},
				End:   protocol.Position{Line: d.Range.End.Line, Character: d.Range.End.Character},
			},
			Severity: &severity,
			Source:   &d.Source,
			Message:  d.Message,
		})
	}
	return lspDiagnostics
}
