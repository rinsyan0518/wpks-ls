package out

import "github.com/rinsyan0518/wpks-ls/internal/pkg/domain"

type PackwerkRunner interface {
	RunCheck(rootPath string, path string) (*domain.CheckResult, error)
}
