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
			"Execute bash/shell commands in an isolated Linux environment. Use this tool when you need to run shell commands, system utilities, or interact with the filesystem. Only output printed to stdout or stderr is returned so make sure commands produce output! Please note all code is run in an ephemeral container so files and state do NOT persist!",
		),
		mcp.WithString(
			"script",
			mcp.Description("The bash script or commands to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"packages",
			mcp.Description(
				"Comma-separated list of Ubuntu packages your script requires. If your script requires external tools you MUST pass them here! These will be installed automatically using apt-get.",
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

	output, err := b.executor.Execute(ctx, script, packages)
	if err != nil {
		logger.Debug("Bash execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Bash execution completed successfully")
	return mcp.NewToolResultText(output), nil
}
