package inmemory

import (
	"errors"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type WorkspaceRepository struct {
	workspace *domain.Workspace
}

func NewWorkspaceRepository() *WorkspaceRepository {
	return &WorkspaceRepository{}
}

func (r *WorkspaceRepository) Save(workspace *domain.Workspace) error {
	r.workspace = workspace
	return nil
}

func (r *WorkspaceRepository) GetWorkspace() (*domain.Workspace, error) {
	if r.workspace == nil {
		return nil, errors.New("workspace not found")
	}
	return r.workspace, nil
}

var _ out.WorkspaceRepository = (*WorkspaceRepository)(nil)
