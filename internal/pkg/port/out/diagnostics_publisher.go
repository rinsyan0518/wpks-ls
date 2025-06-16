package portout

import "github.com/rinsyan0518/wpks-ls/internal/pkg/domain"

type DiagnosticsPublisher interface {
	Publish(uri string, diagnostics []domain.Diagnostic)
}
