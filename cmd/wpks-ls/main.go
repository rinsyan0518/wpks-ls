package main

import "github.com/rinsyan0518/wpks-ls/internal/app"

func main() {
	server := &app.LSPServer{}
	server.Start()
}
