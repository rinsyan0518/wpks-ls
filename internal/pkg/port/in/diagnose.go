package in

import "github.com/rinsyan0518/wpks-ls/internal/pkg/domain"

type DiagnoseFile interface {
	Diagnose(uri string) ([]domain.Diagnostic, error)
}
