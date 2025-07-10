package main

import (
	"log"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/lsp"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/packwerk"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/usecase"
)

func main() {
	workspaceRepository := inmemory.NewWorkspaceRepository()
	diagnoseFile := usecase.NewDiagnoseFile(workspaceRepository, packwerk.NewRunnerWithDefaultCheckers())
	createWorkspace := usecase.NewCreateWorkspace(workspaceRepository)
	server := lsp.NewServer(diagnoseFile, createWorkspace)
	err := server.Start()
	if err != nil {
		log.Fatalf("failed to start LSP server: %v", err)
	}
}
