// Package server provides MCP server initialization and transport management
// for running the mcp-executor with stdio, SSE, and HTTP transport modes.
package server

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/ylchen07/mcp-executor/internal/config"
	"github.com/ylchen07/mcp-executor/internal/executor"
	"github.com/ylchen07/mcp-executor/internal/logger"
	"github.com/ylchen07/mcp-executor/internal/prompts"
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
		typescriptExecutor := executor.NewTypeScriptExecutor()
		goExecutor := executor.NewGoExecutor()

		logger.Debug("Initializing Docker Python tool with module installation support")
		pythonTool := tools.NewPythonTool(pythonExecutor)

		logger.Debug("Initializing Docker Bash tool with package installation support")
		bashTool := tools.NewBashTool(bashExecutor)

		logger.Debug("Initializing Docker TypeScript tool with package installation support")
		typescriptTool := tools.NewTypeScriptTool(typescriptExecutor)

		logger.Debug("Initializing Docker Go tool with package installation support")
		goTool := tools.NewGoTool(goExecutor)

		logger.Debug("Registering Docker tools with MCP server")
		mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
		mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)
		mcpServer.AddTool(typescriptTool.CreateTool(), typescriptTool.HandleExecution)
		mcpServer.AddTool(goTool.CreateTool(), goTool.HandleExecution)

	case "subprocess":
		logger.Debug("Using subprocess executors (no dependency installation)")
		pythonExecutor := executor.NewSubprocessPythonExecutor()
		bashExecutor := executor.NewSubprocessBashExecutor()
		typescriptExecutor := executor.NewSubprocessTypeScriptExecutor()
		goExecutor := executor.NewSubprocessGoExecutor()

		logger.Debug("Initializing subprocess Python tool (no module installation)")
		pythonTool := tools.NewSubprocessPythonTool(pythonExecutor)

		logger.Debug("Initializing subprocess Bash tool (no package installation)")
		bashTool := tools.NewSubprocessBashTool(bashExecutor)

		logger.Debug("Initializing subprocess TypeScript tool (no package installation)")
		typescriptTool := tools.NewSubprocessTypeScriptTool(typescriptExecutor)

		logger.Debug("Initializing subprocess Go tool (no package installation)")
		goTool := tools.NewSubprocessGoTool(goExecutor)

		logger.Debug("Registering subprocess tools with MCP server")
		mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
		mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)
		mcpServer.AddTool(typescriptTool.CreateTool(), typescriptTool.HandleExecution)
		mcpServer.AddTool(goTool.CreateTool(), goTool.HandleExecution)

	default:
		logger.Debug("Unknown execution mode '%s', defaulting to subprocess", executionMode)
		pythonExecutor := executor.NewSubprocessPythonExecutor()
		bashExecutor := executor.NewSubprocessBashExecutor()
		typescriptExecutor := executor.NewSubprocessTypeScriptExecutor()
		goExecutor := executor.NewSubprocessGoExecutor()

		pythonTool := tools.NewSubprocessPythonTool(pythonExecutor)
		bashTool := tools.NewSubprocessBashTool(bashExecutor)
		typescriptTool := tools.NewSubprocessTypeScriptTool(typescriptExecutor)
		goTool := tools.NewSubprocessGoTool(goExecutor)

		mcpServer.AddTool(pythonTool.CreateTool(), pythonTool.HandleExecution)
		mcpServer.AddTool(bashTool.CreateTool(), bashTool.HandleExecution)
		mcpServer.AddTool(typescriptTool.CreateTool(), typescriptTool.HandleExecution)
		mcpServer.AddTool(goTool.CreateTool(), goTool.HandleExecution)
	}

	// Register prompts based on execution mode
	registerPrompts(mcpServer, executionMode)

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

// registerPrompts registers prompts to the MCP server based on execution mode.
// Some prompts are only available in specific execution modes:
// - subprocess: system-check (host system information)
// - docker: (future prompts that require container isolation)
// - all modes: (future universal prompts)
func registerPrompts(mcpServer *server.MCPServer, executionMode string) {
	logger.Debug("Registering prompts for execution mode: %s", executionMode)

	switch executionMode {
	case "subprocess", "": // Empty string is default/unknown mode (defaults to subprocess)
		logger.Debug("Registering subprocess-mode prompts")

		// System check - only works in subprocess mode for host system info
		systemCheckPrompt := prompts.NewSystemCheckPrompt()
		mcpServer.AddPrompt(
			systemCheckPrompt.CreatePrompt(),
			systemCheckPrompt.HandlePrompt,
		)
		logger.Debug("Registered system-check prompt")

	case "docker":
		logger.Debug("No prompts registered for Docker mode (container-only context)")
		// Future: Add Docker-specific prompts here
		// Example: prompts for exploring container capabilities, installed packages, etc.
	}

	// Future: Register prompts that work in ALL execution modes
	// Example:
	// logger.Debug("Registering universal prompts")
	// helpPrompt := prompts.NewHelpPrompt()
	// mcpServer.AddPrompt(helpPrompt.CreatePrompt(), helpPrompt.HandlePrompt())
}
