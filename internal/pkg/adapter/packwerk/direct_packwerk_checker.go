package packwerk

import (
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

func (c *DirectPackwerkChecker) RunCheck(rootPath, path string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.Command("packwerk", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

func (c *DirectPackwerkChecker) RunCheckAll(rootPath string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.Command("packwerk", "check")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

var _ CheckerCommand = &DirectPackwerkChecker{}
