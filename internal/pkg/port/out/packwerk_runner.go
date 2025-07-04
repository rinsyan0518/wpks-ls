package out

import (
	"context"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type PackwerkRunner interface {
	RunCheck(context context.Context, rootPath string, path string) ([]domain.Violation, error)
	RunCheckAll(context context.Context, rootPath string) ([]domain.Violation, error)
}
