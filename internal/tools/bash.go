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

type BashTool struct {
	executor executor.Executor
}

func NewBashTool(exec executor.Executor) *BashTool {
	return &BashTool{
		executor: exec,
	}
}

func (b *BashTool) CreateTool() mcp.Tool {
	return mcp.NewTool(
		"execute-bash",
		mcp.WithDescription(
			"Execute bash/shell commands in an isolated Docker container (Ubuntu 22.04). System packages can be dynamically installed. Use this tool when you need to run shell commands, system utilities, or require specific command-line tools. Only output printed to stdout or stderr is returned so make sure commands produce output! Note: Code runs in ephemeral containers - files and state do NOT persist between executions.",
		),
		mcp.WithString(
			"script",
			mcp.Description("The bash script or commands to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"packages",
			mcp.Description(
				"Comma-separated list of Ubuntu packages to install (e.g., 'curl,jq,git'). Packages are installed automatically via apt-get before script execution.",
			),
		),
		mcp.WithString(
			"env",
			mcp.Description(
				"Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true'). These will be available to your bash script.",
			),
		),
	)
}

func (b *BashTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Bash tool execution requested")

	script, err := request.RequireString("script")
	if err != nil {
		logger.Debug("Bash tool execution failed: missing script argument")
		return mcp.NewToolResultError("Missing or invalid script argument"), nil
	}

	var packages []string
	if packagesStr := request.GetString("packages", ""); packagesStr != "" {
		packages = strings.Split(packagesStr, ",")
		// Clean up package names (trim whitespace)
		for i, pkg := range packages {
			packages[i] = strings.TrimSpace(pkg)
		}
		logger.Debug("Bash packages requested: %v", packages)
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
		logger.Debug("Bash environment variables: %v", envVars)
	}

	output, err := b.executor.Execute(ctx, script, packages, envVars)
	if err != nil {
		logger.Debug("Bash execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Bash execution completed successfully")
	return mcp.NewToolResultText(output), nil
}

// SubprocessBashTool executes bash commands on the host system without package installation support
type SubprocessBashTool struct {
	executor executor.Executor
}

func NewSubprocessBashTool(exec executor.Executor) *SubprocessBashTool {
	return &SubprocessBashTool{
		executor: exec,
	}
}

func (b *SubprocessBashTool) CreateTool() mcp.Tool {
	return mcp.NewTool(
		"execute-bash",
		mcp.WithDescription(
			"Execute bash/shell commands directly on the host system. Only pre-installed system utilities are available. Use this tool when you need to run shell commands or interact with the host filesystem. Only output printed to stdout or stderr is returned so make sure commands produce output! Note: Code runs on the host system with user permissions.",
		),
		mcp.WithString(
			"script",
			mcp.Description("The bash script or commands to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"env",
			mcp.Description(
				"Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true'). These will be available to your bash script.",
			),
		),
	)
}

func (b *SubprocessBashTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Subprocess Bash tool execution requested")

	script, err := request.RequireString("script")
	if err != nil {
		logger.Debug("Subprocess Bash tool execution failed: missing script argument")
		return mcp.NewToolResultError("Missing or invalid script argument"), nil
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
		logger.Debug("Subprocess Bash environment variables: %v", envVars)
	}

	// No package installation for subprocess mode - pass empty slice
	output, err := b.executor.Execute(ctx, script, nil, envVars)
	if err != nil {
		logger.Debug("Subprocess Bash execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Subprocess Bash execution completed successfully")
	return mcp.NewToolResultText(output), nil
}
