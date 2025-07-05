package packwerk

import (
	"context"
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

func (c *BundlePackwerkChecker) RunCheck(context context.Context, rootPath string, paths ...string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bundle"}
	}

	if len(paths) == 0 {
		return []domain.Violation{}, nil
	}

	// bundle exec packwerk can accept multiple paths
	args := []string{"exec", "packwerk", "check", "--offenses-formatter=default", "--"}
	args = append(args, paths...)

	cmd := exec.CommandContext(context, "bundle", args...)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

func (c *BundlePackwerkChecker) RunCheckAll(context context.Context, rootPath string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bundle"}
	}
	cmd := exec.CommandContext(context, "bundle", "exec", "packwerk", "check", "--offenses-formatter=default")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

var _ CheckerCommand = &BundlePackwerkChecker{}
