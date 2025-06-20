package packwerk

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type BinPackwerkChecker struct{}

func (c BinPackwerkChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	packwerkPath := filepath.Join(rootPath, "bin", "packwerk")
	if _, err := os.Stat(packwerkPath); os.IsNotExist(err) {
		return nil, CommandNotFoundError{"bin/packwerk"}
	}

	cmd := exec.Command(packwerkPath, "check", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}
