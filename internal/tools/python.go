package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ylchen07/mcp-executor/internal/executor"
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
			"Execute Python code in an isolated environment. Playwright and headless browser are available for web scraping. Use this tool when you need real-time information, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements! Please note all code is run in an ephemeral container so modules and code do NOT persist!",
		),
		mcp.WithString(
			"code",
			mcp.Description("The Python code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"modules",
			mcp.Description(
				"Comma-separated list of Python modules your code requires. If your code requires external modules you MUST pass them here! These will installed automatically.",
			),
		),
	)
}

func (p *PythonTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	var modules []string
	if modulesStr := request.GetString("modules", ""); modulesStr != "" {
		modules = strings.Split(modulesStr, ",")
	}

	output, err := p.executor.Execute(ctx, code, modules)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(output), nil
}