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

func (c *DirectPackwerkChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.Command("packwerk", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

func (c *DirectPackwerkChecker) RunCheckAll(rootPath string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.Command("packwerk", "check")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

var _ CheckerCommand = &DirectPackwerkChecker{}
