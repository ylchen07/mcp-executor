// Package tools provides MCP tool implementations for executing TypeScript code
// in isolated Docker containers with support for dynamic package installation.
package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ylchen07/mcp-executor/internal/executor"
	"github.com/ylchen07/mcp-executor/internal/logger"
)

type TypeScriptTool struct {
	executor executor.Executor
}

func NewTypeScriptTool(exec executor.Executor) *TypeScriptTool {
	return &TypeScriptTool{
		executor: exec,
	}
}

func (t *TypeScriptTool) CreateTool() mcp.Tool {
	description := `Execute TypeScript code in an isolated Docker container with tsx runtime.
External packages can be dynamically installed via npm. Use this tool when you need real-time information or require external npm packages.
Only output printed to stdout or stderr is returned so ALWAYS use console.log() statements!
Note: Code runs in ephemeral containers - packages and state do NOT persist between executions.`

	return mcp.NewTool(
		"execute-typescript",
		mcp.WithDescription(description),
		mcp.WithString(
			"code",
			mcp.Description("The TypeScript code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"packages",
			mcp.Description(`Comma-separated list of npm packages to install (e.g., 'axios,lodash,date-fns').
Packages are installed automatically via npm before code execution.`),
		),
		mcp.WithString(
			"env",
			mcp.Description(`Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true').
These will be available to your TypeScript code.`),
		),
	)
}

func (t *TypeScriptTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("TypeScript tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("TypeScript tool execution failed: missing code argument")
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	var packages []string
	if packagesStr := request.GetString("packages", ""); packagesStr != "" {
		packages = strings.Split(packagesStr, ",")
		logger.Debug("TypeScript packages requested: %v", packages)
	}

	// Parse environment variables
	envVars := make(map[string]string)
	if envStr := request.GetString("env", ""); envStr != "" {
		envPairs := strings.SplitSeq(envStr, ",")
		for pair := range envPairs {
			pair = strings.TrimSpace(pair)
			if equalIndex := strings.Index(pair, "="); equalIndex > 0 {
				key := strings.TrimSpace(pair[:equalIndex])
				value := strings.TrimSpace(pair[equalIndex+1:])
				envVars[key] = value
			}
		}
		logger.Debug("TypeScript environment variables: %v", envVars)
	}

	output, err := t.executor.Execute(ctx, code, packages, envVars)
	if err != nil {
		logger.Debug("TypeScript execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("TypeScript execution completed successfully")
	return mcp.NewToolResultText(output), nil
}

// SubprocessTypeScriptTool executes TypeScript code on the host system without package installation support
type SubprocessTypeScriptTool struct {
	executor executor.Executor
}

func NewSubprocessTypeScriptTool(exec executor.Executor) *SubprocessTypeScriptTool {
	return &SubprocessTypeScriptTool{
		executor: exec,
	}
}

func (t *SubprocessTypeScriptTool) CreateTool() mcp.Tool {
	description := `Execute TypeScript code directly on the host system using ts-node or tsx. Only standard library and pre-installed packages are available.
Use this tool when you need real-time information and don't require external dependencies.
Only output printed to stdout or stderr is returned so ALWAYS use console.log() statements!
Note: Code runs on the host system with user permissions. Requires ts-node or tsx to be installed.`

	return mcp.NewTool(
		"execute-typescript",
		mcp.WithDescription(description),
		mcp.WithString(
			"code",
			mcp.Description("The TypeScript code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"env",
			mcp.Description(`Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true').
These will be available to your TypeScript code.`),
		),
	)
}

func (t *SubprocessTypeScriptTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Subprocess TypeScript tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Subprocess TypeScript tool execution failed: missing code argument")
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	// Parse environment variables
	envVars := make(map[string]string)
	if envStr := request.GetString("env", ""); envStr != "" {
		envPairs := strings.SplitSeq(envStr, ",")
		for pair := range envPairs {
			pair = strings.TrimSpace(pair)
			if equalIndex := strings.Index(pair, "="); equalIndex > 0 {
				key := strings.TrimSpace(pair[:equalIndex])
				value := strings.TrimSpace(pair[equalIndex+1:])
				envVars[key] = value
			}
		}
		logger.Debug("Subprocess TypeScript environment variables: %v", envVars)
	}

	// No package installation for subprocess mode - pass empty slice
	output, err := t.executor.Execute(ctx, code, nil, envVars)
	if err != nil {
		logger.Debug("Subprocess TypeScript execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Subprocess TypeScript execution completed successfully")
	return mcp.NewToolResultText(output), nil
}
