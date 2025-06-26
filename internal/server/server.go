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

	dockerExecutor := executor.NewDockerExecutor()
	pythonTool := tools.NewPythonTool(dockerExecutor)

	mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)

	return mcpServer
}

func RunStdio(mcpServer *server.MCPServer) error {
	return server.ServeStdio(mcpServer)
}

func RunSSE(mcpServer *server.MCPServer) error {
	sseServer := server.NewSSEServer(mcpServer, config.SSEHost)
	log.Printf("Starting SSE server on localhost:8080")
	return sseServer.Start(config.SSEPort)
}