package packwerk

import (
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type Runner struct{}

func (Runner) RunCheck(rootPath string, path string) (*domain.CheckResult, error) {
	cmd := exec.Command("bundle", "exec", "packwerk", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

var _ out.PackwerkRunner = (*Runner)(nil)
