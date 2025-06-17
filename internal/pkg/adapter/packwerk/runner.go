package packwerk

import (
	"os/exec"
	"strings"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type Runner struct{}

func (Runner) RunCheck(uri string) (*domain.CheckResult, error) {
	path := strings.TrimPrefix(uri, "file://")
	cmd := exec.Command("packwerk", "check", path)
	out, err := cmd.CombinedOutput()
	return domain.NewCheckResult(string(out)), err
}

var _ out.PackwerkRunner = (*Runner)(nil)
