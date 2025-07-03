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

func (c *BinPackwerkChecker) RunCheck(rootPath, path string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bin/packwerk"}
	}
	packwerkPath := filepath.Join(rootPath, "bin", "packwerk")
	cmd := exec.Command(packwerkPath, "check", "--offenses-formatter=default", "--", path)
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

func (c *BinPackwerkChecker) RunCheckAll(rootPath string) ([]domain.Violation, error) {
	if !c.IsAvailable(rootPath) {
		return nil, CommandNotFoundError{"bin/packwerk"}
	}
	packwerkPath := filepath.Join(rootPath, "bin", "packwerk")
	cmd := exec.Command(packwerkPath, "check", "--offenses-formatter=default")
	cmd.Dir = rootPath
	out, _ := cmd.Output()
	return NewPackwerkOutput(string(out)).Parse(), nil
}

var _ CheckerCommand = &BinPackwerkChecker{}
