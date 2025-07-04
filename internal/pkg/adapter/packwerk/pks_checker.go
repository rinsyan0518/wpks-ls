package packwerk

import (
	"context"
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type PksChecker struct{}

func NewPksChecker() *PksChecker {
	return &PksChecker{}
}

func (c *PksChecker) IsAvailable(rootPath string) bool {
	_, pksErr := exec.LookPath("pks")
	return pksErr == nil
}

func (c *PksChecker) RunCheck(context context.Context, rootPath, path string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"pks"}
	}
	cmd := exec.CommandContext(context, "pks", "-e", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

func (c *PksChecker) RunCheckAll(context context.Context, rootPath string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"pks"}
	}
	cmd := exec.CommandContext(context, "pks", "-e", "check")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

var _ CheckerCommand = &PksChecker{}
