package main

import (
	"github.com/rinsyan0518/wpks-ls/internal/app"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter"
)

func main() {
	runner := adapter.PackwerkRunnerImpl{}
	publisher := adapter.DiagnosticsPublisherImpl{}
	server := app.NewLSPServer(runner, publisher)
	server.Start()
}
