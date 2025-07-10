package usecase

import (
	"context"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

const (
	packwerkSource = "packwerk"
)

type DiagnoseFile struct {
	workspaceRepository out.WorkspaceRepository
	packwerkRunner      out.PackwerkRunner
}

func NewDiagnoseFile(workspaceRepository out.WorkspaceRepository, packwerkRunner out.PackwerkRunner) *DiagnoseFile {
	return &DiagnoseFile{workspaceRepository: workspaceRepository, packwerkRunner: packwerkRunner}
}

func (d *DiagnoseFile) Diagnose(context context.Context, uris ...string) (map[string][]domain.Diagnostic, error) {
	if len(uris) == 0 {
		return map[string][]domain.Diagnostic{}, nil
	}

	workspace, err := d.workspaceRepository.GetWorkspace()
	if err != nil {
		return nil, err
	}

	// Convert URIs to paths
	paths := make([]string, len(uris))
	for i, uri := range uris {
		paths[i] = workspace.StripRootUri(uri)
	}

	// Run check for all paths at once
	violations, err := d.packwerkRunner.RunCheck(context, workspace.RootPath, paths...)
	if err != nil {
		return nil, err
	}

	// Group violations by file URI
	allDiagnostics := make(map[string][]domain.Diagnostic)
	for _, v := range violations {
		fileUri := workspace.BuildFileUri(v.File)
		diagnostic := domain.Diagnostic{
			Range: domain.Range{
				Start: domain.Position{Line: v.Line - 1, Character: v.Character},
				End:   domain.Position{Line: v.Line - 1, Character: v.Character + 1},
			},
			Severity: domain.SeverityError,
			Source:   packwerkSource,
			Message:  v.Message,
		}
		allDiagnostics[fileUri] = append(allDiagnostics[fileUri], diagnostic)
	}

	return allDiagnostics, nil
}

func (d *DiagnoseFile) DiagnoseAll(context context.Context) (map[string][]domain.Diagnostic, error) {
	workspace, err := d.workspaceRepository.GetWorkspace()
	if err != nil {
		return nil, err
	}

	violations, err := d.packwerkRunner.RunCheckAll(
		context,
		workspace.RootPath,
	)
	if err != nil {
		return nil, err
	}

	diagnosticsByFile := make(map[string][]domain.Diagnostic)

	for _, v := range violations {
		fileUri := workspace.BuildFileUri(v.File)
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
