package packwerk

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type BinPackwerkChecker struct{}

func NewBinPackwerkChecker() *BinPackwerkChecker {
	return &BinPackwerkChecker{}
}

func (c *BinPackwerkChecker) IsAvailable(rootPath string) bool {
	packwerkPath := filepath.Join(rootPath, "bin", "packwerk")
	if _, err := os.Stat(packwerkPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *BinPackwerkChecker) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bin/packwerk"}
	}
	packwerkPath := filepath.Join(rootPath, "bin", "packwerk")
	cmd := exec.Command(packwerkPath, "check", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

func (c *BinPackwerkChecker) RunCheckAll(rootPath string) (*domain.CheckResult, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bin/packwerk"}
	}
	packwerkPath := filepath.Join(rootPath, "bin", "packwerk")
	cmd := exec.Command(packwerkPath, "check")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return domain.NewCheckResult(string(out)), nil
}

var _ CheckerCommand = &BinPackwerkChecker{}
