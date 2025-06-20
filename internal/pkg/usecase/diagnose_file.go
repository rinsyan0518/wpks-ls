package usecase

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

const (
	packwerkSource = "packwerk"
)

type DiagnoseFile struct {
	configurationRepository out.ConfigurationRepository
	packwerkRunner          out.PackwerkRunner
}

func NewDiagnoseFile(configurationRepository out.ConfigurationRepository, packwerkRunner out.PackwerkRunner) *DiagnoseFile {
	return &DiagnoseFile{configurationRepository: configurationRepository, packwerkRunner: packwerkRunner}
}

func (d *DiagnoseFile) Diagnose(uri string) ([]domain.Diagnostic, error) {
	configuration, err := d.configurationRepository.GetConfiguration()
	if err != nil {
		return nil, err
	}

	violations, err := d.packwerkRunner.RunCheck(
		configuration.RootPath,
		configuration.StripRootUri(uri),
	)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]domain.Diagnostic, 0, len(violations))
	for _, v := range violations {
		diagnostics = append(diagnostics, domain.Diagnostic{
			Range: domain.Range{
				Start: domain.Position{Line: v.Line - 1, Character: v.Character},
				End:   domain.Position{Line: v.Line - 1, Character: v.Character + 1},
			},
			Severity: domain.SeverityError,
			Source:   packwerkSource,
			Message:  v.Message,
		})
	}

	return diagnostics, nil
}

func (d *DiagnoseFile) DiagnoseAll() (map[string][]domain.Diagnostic, error) {
	configuration, err := d.configurationRepository.GetConfiguration()
	if err != nil {
		return nil, err
	}

	if !configuration.CheckAllOnInitialized {
		return map[string][]domain.Diagnostic{}, nil
	}

	violations, err := d.packwerkRunner.RunCheckAll(configuration.RootPath)
	if err != nil {
		return nil, err
	}

	diagnosticsByFile := make(map[string][]domain.Diagnostic)

	for _, v := range violations {
		fileUri := configuration.BuildFileUri(v.File)
		diagnostic := domain.Diagnostic{
			Range: domain.Range{
				Start: domain.Position{Line: v.Line - 1, Character: v.Character},
				End:   domain.Position{Line: v.Line - 1, Character: v.Character + 1},
			},
			Severity: domain.SeverityError,
			Source:   packwerkSource,
			Message:  v.Message,
		}

		diagnosticsByFile[fileUri] = append(diagnosticsByFile[fileUri], diagnostic)
	}

	return diagnosticsByFile, nil
}

var _ in.DiagnoseFile = (*DiagnoseFile)(nil)
