package packwerk

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type Runner struct{}

func (Runner) RunCheck(rootPath string, path string) (*domain.CheckResult, error) {
	// Skip diagnostics if packwerk.yml does not exist in the workspace root
	if _, err := os.Stat(filepath.Join(rootPath, "packwerk.yml")); err != nil {
		if os.IsNotExist(err) {
			return domain.NewCheckResult(""), nil
		}
		return nil, err
	}

	// Prefer bundle exec packwerk if available in the bundle context
	bundlePath, bundleErr := exec.LookPath("bundle")
	if bundleErr == nil {
		showCmd := exec.Command(bundlePath, "show", "packwerk")
		showCmd.Dir = rootPath
		if err := showCmd.Run(); err == nil {
			cmd := exec.Command(bundlePath, "exec", "packwerk", "check", "--", path)
			cmd.Dir = rootPath
			out, _ := cmd.Output()
			return domain.NewCheckResult(string(out)), nil
		}
	}

	// Fallback: try packwerk directly
	packwerkPath, packwerkErr := exec.LookPath("packwerk")
	if packwerkErr == nil {
		cmd := exec.Command(packwerkPath, "check", "--", path)
		cmd.Dir = rootPath
		out, _ := cmd.Output()
		return domain.NewCheckResult(string(out)), nil
	}

	return nil, errors.New("packwerk not found")
}

var _ out.PackwerkRunner = (*Runner)(nil)
