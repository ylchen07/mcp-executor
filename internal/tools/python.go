// Package tools provides MCP tool implementations for executing Python and Bash code
// in isolated Docker containers with support for dynamic module/package installation.
package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ylchen07/mcp-executor/internal/executor"
	"github.com/ylchen07/mcp-executor/internal/logger"
)

type PythonTool struct {
	executor executor.Executor
}

func NewPythonTool(exec executor.Executor) *PythonTool {
	return &PythonTool{
		executor: exec,
	}
}

func (p *PythonTool) CreateTool() mcp.Tool {
	return mcp.NewTool(
		"execute-python",
		mcp.WithDescription(
			"Execute Python code in an isolated Docker container. Playwright and headless browsers are pre-installed for web scraping. External modules can be dynamically installed. Use this tool when you need real-time information or require external Python packages. Only output printed to stdout or stderr is returned so ALWAYS use print statements! Note: Code runs in ephemeral containers - modules and state do NOT persist between executions.",
		),
		mcp.WithString(
			"code",
			mcp.Description("The Python code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"modules",
			mcp.Description(
				"Comma-separated list of Python modules to install (e.g., 'requests,beautifulsoup4,pandas'). Modules are installed automatically via pip before code execution.",
			),
		),
		mcp.WithString(
			"env",
			mcp.Description(
				"Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true'). These will be available to your Python code.",
			),
		),
	)
}

func (p *PythonTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Python tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Python tool execution failed: missing code argument")
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	var modules []string
	if modulesStr := request.GetString("modules", ""); modulesStr != "" {
		modules = strings.Split(modulesStr, ",")
		logger.Debug("Python modules requested: %v", modules)
	}

	// Parse environment variables
	envVars := make(map[string]string)
	if envStr := request.GetString("env", ""); envStr != "" {
		envPairs := strings.Split(envStr, ",")
		for _, pair := range envPairs {
			pair = strings.TrimSpace(pair)
			if equalIndex := strings.Index(pair, "="); equalIndex > 0 {
				key := strings.TrimSpace(pair[:equalIndex])
				value := strings.TrimSpace(pair[equalIndex+1:])
				envVars[key] = value
			}
		}
		logger.Debug("Python environment variables: %v", envVars)
	}

	output, err := p.executor.Execute(ctx, code, modules, envVars)
	if err != nil {
		logger.Debug("Python execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Python execution completed successfully")
	return mcp.NewToolResultText(output), nil
}

// SubprocessPythonTool executes Python code on the host system without module installation support
type SubprocessPythonTool struct {
	executor executor.Executor
}

func NewSubprocessPythonTool(exec executor.Executor) *SubprocessPythonTool {
	return &SubprocessPythonTool{
		executor: exec,
	}
}

func (p *SubprocessPythonTool) CreateTool() mcp.Tool {
	return mcp.NewTool(
		"execute-python",
		mcp.WithDescription(
			"Execute Python code directly on the host system. Only standard library and pre-installed packages are available. Use this tool when you need real-time information and don't require external dependencies. Only output printed to stdout or stderr is returned so ALWAYS use print statements! Note: Code runs on the host system with user permissions.",
		),
		mcp.WithString(
			"code",
			mcp.Description("The Python code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"env",
			mcp.Description(
				"Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true'). These will be available to your Python code.",
			),
		),
	)
}

func (p *SubprocessPythonTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Subprocess Python tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Subprocess Python tool execution failed: missing code argument")
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	// Parse environment variables
	envVars := make(map[string]string)
	if envStr := request.GetString("env", ""); envStr != "" {
		envPairs := strings.Split(envStr, ",")
		for _, pair := range envPairs {
			pair = strings.TrimSpace(pair)
			if equalIndex := strings.Index(pair, "="); equalIndex > 0 {
				key := strings.TrimSpace(pair[:equalIndex])
				value := strings.TrimSpace(pair[equalIndex+1:])
				envVars[key] = value
			}
		}
		logger.Debug("Subprocess Python environment variables: %v", envVars)
	}

	// No module installation for subprocess mode - pass empty slice
	output, err := p.executor.Execute(ctx, code, nil, envVars)
	if err != nil {
		logger.Debug("Subprocess Python execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Subprocess Python execution completed successfully")
	return mcp.NewToolResultText(output), nil
}
