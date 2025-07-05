package in

import (
	"context"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type DiagnoseFile interface {
	Diagnose(context context.Context, uris ...string) (map[string][]domain.Diagnostic, error)
	DiagnoseAll(context context.Context) (map[string][]domain.Diagnostic, error)
}
