package usecase

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type DiagnoseFile struct {
	packwerkRunner out.PackwerkRunner
}

func NewDiagnoseFile(packwerkRunner out.PackwerkRunner) *DiagnoseFile {
	return &DiagnoseFile{packwerkRunner: packwerkRunner}
}

func (d *DiagnoseFile) Diagnose(uri string) ([]domain.Diagnostic, error) {
	checkResult, err := d.packwerkRunner.RunCheck(uri)
	violations := checkResult.Parse()
	diagnostics := make([]domain.Diagnostic, 0, len(violations)+1)
	for _, v := range violations {
		diagnostics = append(diagnostics, domain.Diagnostic{
			Range: domain.Range{
				Start: domain.Position{Line: v.Line - 1, Character: v.Column - 1},
				End:   domain.Position{Line: v.Line - 1, Character: v.Column - 1},
			},
			Severity: domain.SeverityError,
			Source:   "packwerk",
			Message:  v.Message,
		})
	}
	if err != nil {
		diagnostics = append(diagnostics, domain.Diagnostic{
			Range: domain.Range{
				Start: domain.Position{Line: 0, Character: 0},
				End:   domain.Position{Line: 0, Character: 0},
			},
			Severity: domain.SeverityError,
			Source:   "packwerk",
			Message:  "Packwerk error: " + err.Error(),
		})
	}
	return diagnostics, err
}

var _ in.DiagnoseFile = (*DiagnoseFile)(nil)
