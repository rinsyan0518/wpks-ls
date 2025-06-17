package main

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/lsp"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/packwerk"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/usecase"
)

func main() {
	diagnoseFile := usecase.NewDiagnoseFile(packwerk.Runner{})
	server := lsp.NewServer(diagnoseFile)
	server.Start()
}
