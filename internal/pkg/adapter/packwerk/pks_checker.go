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

func (c *PksChecker) RunCheck(context context.Context, rootPath string, paths ...string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"pks"}
	}

	if len(paths) == 0 {
		return []domain.Violation{}, nil
	}

	// pks can accept multiple paths
	args := []string{"-e", "check", "--"}
	args = append(args, paths...)

	cmd := exec.CommandContext(context, "pks", args...)
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
