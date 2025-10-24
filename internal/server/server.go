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

func NewMCPServer(executionMode string) *server.MCPServer {
	logger.Debug("Creating new MCP server with execution mode: %s", executionMode)
	mcpServer := server.NewMCPServer(
		config.ServerName,
		config.ServerVersion,
	)

	// Create executors based on execution mode
	var pythonExecutor executor.Executor
	var bashExecutor executor.Executor

	switch executionMode {
	case "docker":
		logger.Debug("Using Docker executors")
		pythonExecutor = executor.NewPythonExecutor()
		bashExecutor = executor.NewBashExecutor()
	case "subprocess":
		logger.Debug("Using subprocess executors")
		pythonExecutor = executor.NewSubprocessPythonExecutor()
		bashExecutor = executor.NewSubprocessBashExecutor()
	default:
		logger.Debug("Unknown execution mode '%s', defaulting to subprocess", executionMode)
		pythonExecutor = executor.NewSubprocessPythonExecutor()
		bashExecutor = executor.NewSubprocessBashExecutor()
	}

	logger.Debug("Initializing Python tool with %T", pythonExecutor)
	pythonTool := tools.NewPythonTool(pythonExecutor)

	logger.Debug("Initializing Bash tool with %T", bashExecutor)
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
