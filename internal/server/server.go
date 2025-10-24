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

	switch executionMode {
	case "docker":
		logger.Debug("Using Docker executors with full tool capabilities")
		pythonExecutor := executor.NewPythonExecutor()
		bashExecutor := executor.NewBashExecutor()

		logger.Debug("Initializing Docker Python tool with module installation support")
		pythonTool := tools.NewPythonTool(pythonExecutor)

		logger.Debug("Initializing Docker Bash tool with package installation support")
		bashTool := tools.NewBashTool(bashExecutor)

		logger.Debug("Registering Docker tools with MCP server")
		mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
		mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)

	case "subprocess":
		logger.Debug("Using subprocess executors (no dependency installation)")
		pythonExecutor := executor.NewSubprocessPythonExecutor()
		bashExecutor := executor.NewSubprocessBashExecutor()

		logger.Debug("Initializing subprocess Python tool (no module installation)")
		pythonTool := tools.NewSubprocessPythonTool(pythonExecutor)

		logger.Debug("Initializing subprocess Bash tool (no package installation)")
		bashTool := tools.NewSubprocessBashTool(bashExecutor)

		logger.Debug("Registering subprocess tools with MCP server")
		mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
		mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)

	default:
		logger.Debug("Unknown execution mode '%s', defaulting to subprocess", executionMode)
		pythonExecutor := executor.NewSubprocessPythonExecutor()
		bashExecutor := executor.NewSubprocessBashExecutor()

		pythonTool := tools.NewSubprocessPythonTool(pythonExecutor)
		bashTool := tools.NewSubprocessBashTool(bashExecutor)

		mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
		mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)
	}

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
