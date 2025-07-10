package out

import "github.com/rinsyan0518/wpks-ls/internal/pkg/domain"

type WorkspaceRepository interface {
	Save(workspace *domain.Workspace) error
	GetWorkspace() (*domain.Workspace, error)
}
