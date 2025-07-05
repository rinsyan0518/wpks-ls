package packwerk

import (
	"context"
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type DirectPackwerkChecker struct{}

func NewDirectPackwerkChecker() *DirectPackwerkChecker {
	return &DirectPackwerkChecker{}
}

func (c *DirectPackwerkChecker) IsAvailable(rootPath string) bool {
	_, packwerkErr := exec.LookPath("packwerk")
	return packwerkErr == nil
}

func (c *DirectPackwerkChecker) RunCheck(context context.Context, rootPath string, paths ...string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"packwerk"}
	}

	if len(paths) == 0 {
		return []domain.Violation{}, nil
	}

	// packwerk can accept multiple paths
	args := []string{"check", "--offenses-formatter=default", "--"}
	args = append(args, paths...)

	cmd := exec.CommandContext(context, "packwerk", args...)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

func (c *DirectPackwerkChecker) RunCheckAll(context context.Context, rootPath string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.CommandContext(context, "packwerk", "check", "--offenses-formatter=default")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

var _ CheckerCommand = &DirectPackwerkChecker{}
