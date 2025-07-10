package usecase

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type CreateWorkspace struct {
	workspaceRepository out.WorkspaceRepository
}

func NewCreateWorkspace(workspaceRepository out.WorkspaceRepository) *CreateWorkspace {
	return &CreateWorkspace{workspaceRepository: workspaceRepository}
}

func (c *CreateWorkspace) Create(rootUri string, rootPath string) error {
	workspace := domain.NewWorkspace(rootUri, rootPath)
	return c.workspaceRepository.Save(workspace)
}

var _ in.CreateWorkspace = (*CreateWorkspace)(nil)
