// Package server provides MCP server initialization and transport management
// for running the mcp-executor with stdio, SSE, and HTTP transport modes.
package server

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/ylchen07/mcp-executor/internal/config"
	"github.com/ylchen07/mcp-executor/internal/executor"
	"github.com/ylchen07/mcp-executor/internal/logger"
	"github.com/ylchen07/mcp-executor/internal/tools"
)

func NewMCPServer() *server.MCPServer {
	logger.Debug("Creating new MCP server")
	mcpServer := server.NewMCPServer(
		config.ServerName,
		config.ServerVersion,
	)

	logger.Debug("Initializing Python executor and tool")
	pythonExecutor := executor.NewPythonExecutor()
	pythonTool := tools.NewPythonTool(pythonExecutor)

	logger.Debug("Initializing Bash executor and tool")
	bashExecutor := executor.NewBashExecutor()
	bashTool := tools.NewBashTool(bashExecutor)

	logger.Debug("Registering tools with MCP server")
	mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
	mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)

	logger.Debug("MCP server initialization complete")
	return mcpServer
}

func RunStdio(mcpServer *server.MCPServer) error {
	logger.Debug("Starting stdio server")
	return server.ServeStdio(mcpServer)
}

func RunSSE(mcpServer *server.MCPServer) error {
	logger.Debug("Setting up SSE server")
	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL(config.SSEHost))
	logger.Verbose("Starting SSE server on localhost:8080")
	return sseServer.Start(config.SSEPort)
}

func RunHTTP(mcpServer *server.MCPServer) error {
	logger.Debug("Setting up HTTP server")
	httpServer := server.NewStreamableHTTPServer(mcpServer)
	logger.Verbose("Starting HTTP server on localhost:8081")
	return httpServer.Start(config.HTTPPort)
}
