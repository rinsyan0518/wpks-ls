package main

import (
	"log"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/lsp"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/packwerk"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/shared"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/usecase"
)

func main() {
	configurationRepository := inmemory.NewConfigurationRepository()
	jobQueue := shared.NewKeyedSerialJobQueue(100)
	diagnoseFile := usecase.NewDiagnoseFile(configurationRepository, packwerk.NewRunnerWithDefaultCheckers())
	configure := usecase.NewConfigure(configurationRepository)
	server := lsp.NewServer(diagnoseFile, configure, jobQueue)
	err := server.Start()
	if err != nil {
		log.Fatalf("failed to start LSP server: %v", err)
	}
}
