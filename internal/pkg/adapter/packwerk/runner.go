package packwerk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type CheckerCommand interface {
	RunCheck(rootPath, path string) (*domain.CheckResult, error)
}

type CommandNotFoundError struct {
	Cmd string
}

func (e CommandNotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Cmd)
}

func IsCommandNotFoundError(err error) bool {
	_, ok := err.(CommandNotFoundError)
	return ok
}

type Runner struct {
	checkers []CheckerCommand
}

func NewRunnerWithDefaultCheckers() *Runner {
	return NewRunner(
		PksChecker{},
		BinPackwerkChecker{},
		BundlePackwerkChecker{},
		DirectPackwerkChecker{},
	)
}

func NewRunner(checkers ...CheckerCommand) *Runner {
	return &Runner{checkers: checkers}
}

func (r *Runner) RunCheck(rootPath string, path string) (*domain.CheckResult, error) {
	// Skip diagnostics if packwerk.yml does not exist in the workspace root
	if _, err := os.Stat(filepath.Join(rootPath, "packwerk.yml")); err != nil {
		if os.IsNotExist(err) {
			return domain.NewCheckResult(""), nil
		}
		return nil, err
	}

	var lastErr error
	for _, checker := range r.checkers {
		result, err := checker.RunCheck(rootPath, path)
		if err == nil {
			return result, nil
		}
		if IsCommandNotFoundError(err) {
			continue // skip this checker
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("no checker command succeeded")
}

var _ out.PackwerkRunner = (*Runner)(nil)
