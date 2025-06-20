package packwerk

import (
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

func (c *PksChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"pks"}
	}
	cmd := exec.Command("pks", "-e", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

func (c *PksChecker) RunCheckAll(rootPath string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"pks"}
	}
	cmd := exec.Command("pks", "-e", "check")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

var _ CheckerCommand = &PksChecker{}
