// Package tools provides MCP tool implementations for executing Go code
// in isolated Docker containers with support for dynamic package installation.
package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ylchen07/mcp-executor/internal/executor"
	"github.com/ylchen07/mcp-executor/internal/logger"
)

type GoTool struct {
	executor executor.Executor
}

func NewGoTool(exec executor.Executor) *GoTool {
	return &GoTool{
		executor: exec,
	}
}

func (g *GoTool) CreateTool() mcp.Tool {
	description := `Execute Go code in an isolated Docker container.
External packages can be dynamically installed via go get. Use this tool when you need real-time information or require external Go packages.
Only output printed to stdout or stderr is returned so ALWAYS use print/fmt.Println statements!
Note: Code runs in ephemeral containers - packages and state do NOT persist between executions.
Your code must include a main package and main function.`

	return mcp.NewTool(
		"execute-go",
		mcp.WithDescription(description),
		mcp.WithString(
			"code",
			mcp.Description("The Go code to execute (must include package main and func main)"),
			mcp.Required(),
		),
		mcp.WithString(
			"packages",
			mcp.Description(`Comma-separated list of Go packages to install (e.g., 'github.com/gorilla/mux,github.com/gin-gonic/gin').
Packages are installed automatically via go get before code execution.`),
		),
		mcp.WithString(
			"env",
			mcp.Description(`Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true').
These will be available to your Go code.`),
		),
	)
}

func (g *GoTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Go tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Go tool execution failed: missing code argument")
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	var packages []string
	if packagesStr := request.GetString("packages", ""); packagesStr != "" {
		packages = strings.Split(packagesStr, ",")
		logger.Debug("Go packages requested: %v", packages)
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
		logger.Debug("Go environment variables: %v", envVars)
	}

	output, err := g.executor.Execute(ctx, code, packages, envVars)
	if err != nil {
		logger.Debug("Go execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Go execution completed successfully")
	return mcp.NewToolResultText(output), nil
}

// SubprocessGoTool executes Go code on the host system without package installation support
type SubprocessGoTool struct {
	executor executor.Executor
}

func NewSubprocessGoTool(exec executor.Executor) *SubprocessGoTool {
	return &SubprocessGoTool{
		executor: exec,
	}
}

func (g *SubprocessGoTool) CreateTool() mcp.Tool {
	description := `Execute Go code directly on the host system. Only standard library and pre-installed packages are available.
Use this tool when you need real-time information and don't require external dependencies.
Only output printed to stdout or stderr is returned so ALWAYS use print/fmt.Println statements!
Note: Code runs on the host system with user permissions.
Your code must include a main package and main function.`

	return mcp.NewTool(
		"execute-go",
		mcp.WithDescription(description),
		mcp.WithString(
			"code",
			mcp.Description("The Go code to execute (must include package main and func main)"),
			mcp.Required(),
		),
		mcp.WithString(
			"env",
			mcp.Description(`Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true').
These will be available to your Go code.`),
		),
	)
}

func (g *SubprocessGoTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Subprocess Go tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Subprocess Go tool execution failed: missing code argument")
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
		logger.Debug("Subprocess Go environment variables: %v", envVars)
	}

	// No package installation for subprocess mode - pass empty slice
	output, err := g.executor.Execute(ctx, code, nil, envVars)
	if err != nil {
		logger.Debug("Subprocess Go execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Subprocess Go execution completed successfully")
	return mcp.NewToolResultText(output), nil
}
