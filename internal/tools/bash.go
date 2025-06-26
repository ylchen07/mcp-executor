package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ylchen07/mcp-python/internal/executor"
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
	script, ok := request.Params.Arguments["script"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid script argument"), nil
	}

	var packages []string
	if packagesStr, ok := request.Params.Arguments["packages"].(string); ok && packagesStr != "" {
		packages = strings.Split(packagesStr, ",")
		// Clean up package names (trim whitespace)
		for i, pkg := range packages {
			packages[i] = strings.TrimSpace(pkg)
		}
	}

	output, err := b.executor.Execute(ctx, script, packages)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(output), nil
}