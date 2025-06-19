package packwerk

import (
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type BundlePackwerkChecker struct{}

func (BundlePackwerkChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	bundlePath, bundleErr := exec.LookPath("bundle")
	if bundleErr != nil {
		return nil, CommandNotFoundError{"bundle"}
	}
	showCmd := exec.Command(bundlePath, "show", "packwerk")
	showCmd.Dir = rootPath
	if err := showCmd.Run(); err != nil {
		return nil, CommandNotFoundError{"packwerk"}
	}
	cmd := exec.Command(bundlePath, "exec", "packwerk", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}
