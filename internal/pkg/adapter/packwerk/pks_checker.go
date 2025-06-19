package packwerk

import (
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type PksChecker struct{}

func (PksChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	pksPath, pksErr := exec.LookPath("pks")
	if pksErr != nil {
		return nil, CommandNotFoundError{"pks"}
	}
	cmd := exec.Command(pksPath, "-e", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}
