package main

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/lsp"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/packwerk"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/usecase"
)

func main() {
	configurationRepository := inmemory.NewConfigurationRepository()
	diagnoseFile := usecase.NewDiagnoseFile(configurationRepository, packwerk.Runner{})
	configure := usecase.NewConfigure(configurationRepository)
	server := lsp.NewServer(diagnoseFile, configure)
	server.Start()
}
