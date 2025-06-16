package app

import (
	"os/exec"
	"strings"
)

// RunPackwerkCheck runs 'packwerk check' for the given file URI and returns the output.
func RunPackwerkCheck(uri string) (string, error) {
	// Convert file URI to path (strip 'file://')
	path := strings.TrimPrefix(uri, "file://")
	cmd := exec.Command("packwerk", "check", path)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
