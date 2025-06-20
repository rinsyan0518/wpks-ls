package packwerk

import (
	"os/exec"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type BundlePackwerkChecker struct{}

func NewBundlePackwerkChecker() *BundlePackwerkChecker {
	return &BundlePackwerkChecker{}
}

func (c *BundlePackwerkChecker) IsAvailable(rootPath string) bool {
	bundlePath, bundleErr := exec.LookPath("bundle")
	if bundleErr != nil {
		return false
	}
	showCmd := exec.Command(bundlePath, "show", "packwerk")
	showCmd.Dir = rootPath
	if err := showCmd.Run(); err != nil {
		return false
	}
	return true
}

func (c *BundlePackwerkChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bundle"}
	}

	cmd := exec.Command("bundle", "exec", "packwerk", "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

func (c *BundlePackwerkChecker) RunCheckAll(rootPath string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bundle"}
	}
	cmd := exec.Command("bundle", "exec", "packwerk", "check")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

var _ CheckerCommand = &BundlePackwerkChecker{}
