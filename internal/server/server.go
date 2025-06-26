package server

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/ylchen07/mcp-python/internal/config"
	"github.com/ylchen07/mcp-python/internal/executor"
	"github.com/ylchen07/mcp-python/internal/tools"
)

func NewMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		config.ServerName,
		config.ServerVersion,
	)

	pythonExecutor := executor.NewPythonExecutor()
	pythonTool := tools.NewPythonTool(pythonExecutor)

	bashExecutor := executor.NewBashExecutor()
	bashTool := tools.NewBashTool(bashExecutor)

	mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
	mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)

	return mcpServer
}

func RunStdio(mcpServer *server.MCPServer) error {
	return server.ServeStdio(mcpServer)
}

func RunSSE(mcpServer *server.MCPServer) error {
	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL(config.SSEHost))
	log.Printf("Starting SSE server on localhost:8080")
	return sseServer.Start(config.SSEPort)
}

func RunHTTP(mcpServer *server.MCPServer) error {
	httpServer := server.NewStreamableHTTPServer(mcpServer)
	log.Printf("Starting HTTP server on localhost:8081")
	return httpServer.Start(config.HTTPPort)
}