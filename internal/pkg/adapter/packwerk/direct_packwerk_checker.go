package packwerk

import (
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type DirectPackwerkChecker struct{}

func (DirectPackwerkChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	packwerkPath, packwerkErr := exec.LookPath("packwerk")
	if packwerkErr != nil {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.Command(packwerkPath, "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}
