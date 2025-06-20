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
	IsAvailable(rootPath string) bool
	RunCheck(rootPath, path string) (*domain.CheckResult, error)
	RunCheckAll(rootPath string) (*domain.CheckResult, error)
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
		NewPksChecker(),
		NewBinPackwerkChecker(),
		NewBundlePackwerkChecker(),
		NewDirectPackwerkChecker(),
	)
}

func NewRunner(checkers ...CheckerCommand) *Runner {
	return &Runner{checkers: checkers}
}

func (r *Runner) IsAvailable(rootPath string) bool {
	if _, err := os.Stat(filepath.Join(rootPath, "packwerk.yml")); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func (r *Runner) RunCheck(rootPath string, path string) (*domain.CheckResult, error) {
	if !r.IsAvailable(rootPath) {
		return domain.NewCheckResult(""), nil
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

func (r *Runner) RunCheckAll(rootPath string) (*domain.CheckResult, error) {
	if !r.IsAvailable(rootPath) {
		return domain.NewCheckResult(""), nil
	}

	var lastErr error
	for _, checker := range r.checkers {
		result, err := checker.RunCheckAll(rootPath)
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
