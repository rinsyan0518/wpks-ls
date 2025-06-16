package adapter

import (
	"os/exec"
	"strings"

	portout "github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type PackwerkRunnerImpl struct{}

func (PackwerkRunnerImpl) RunCheck(uri string) (string, error) {
	path := strings.TrimPrefix(uri, "file://")
	cmd := exec.Command("packwerk", "check", path)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var _ portout.PackwerkRunner = (*PackwerkRunnerImpl)(nil)
