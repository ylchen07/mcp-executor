// Package tools provides MCP tool implementations for executing Perl code
// in isolated Docker containers with support for dynamic module installation.
package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ylchen07/mcp-executor/internal/executor"
	"github.com/ylchen07/mcp-executor/internal/logger"
)

type PerlTool struct {
	executor executor.Executor
}

func NewPerlTool(exec executor.Executor) *PerlTool {
	return &PerlTool{
		executor: exec,
	}
}

func (p *PerlTool) CreateTool() mcp.Tool {
	description := `Execute Perl code in an isolated Docker container.
External modules can be dynamically installed via CPAN. Use this tool when you need real-time information or require external Perl modules.
Only output printed to stdout or stderr is returned so ALWAYS use print statements!
Note: Code runs in ephemeral containers - modules and state do NOT persist between executions.`

	return mcp.NewTool(
		"execute-perl",
		mcp.WithDescription(description),
		mcp.WithString(
			"code",
			mcp.Description("The Perl code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"modules",
			mcp.Description(`Comma-separated list of Perl modules to install (e.g., 'LWP::UserAgent,JSON,DBI').
Modules are installed automatically via cpanm before code execution.`),
		),
		mcp.WithString(
			"env",
			mcp.Description(`Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true').
These will be available to your Perl code.`),
		),
	)
}

func (p *PerlTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Perl tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Perl tool execution failed: missing code argument")
		return mcp.NewToolResultError("Missing or invalid code argument"), nil
	}

	var modules []string
	if modulesStr := request.GetString("modules", ""); modulesStr != "" {
		modules = strings.Split(modulesStr, ",")
		logger.Debug("Perl modules requested: %v", modules)
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
		logger.Debug("Perl environment variables: %v", envVars)
	}

	output, err := p.executor.Execute(ctx, code, modules, envVars)
	if err != nil {
		logger.Debug("Perl execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Perl execution completed successfully")
	return mcp.NewToolResultText(output), nil
}

// SubprocessPerlTool executes Perl code on the host system without module installation support
type SubprocessPerlTool struct {
	executor executor.Executor
}

func NewSubprocessPerlTool(exec executor.Executor) *SubprocessPerlTool {
	return &SubprocessPerlTool{
		executor: exec,
	}
}

func (p *SubprocessPerlTool) CreateTool() mcp.Tool {
	description := `Execute Perl code directly on the host system. Only standard library and pre-installed modules are available.
Use this tool when you need real-time information and don't require external dependencies.
Only output printed to stdout or stderr is returned so ALWAYS use print statements!
Note: Code runs on the host system with user permissions.`

	return mcp.NewTool(
		"execute-perl",
		mcp.WithDescription(description),
		mcp.WithString(
			"code",
			mcp.Description("The Perl code to execute"),
			mcp.Required(),
		),
		mcp.WithString(
			"env",
			mcp.Description(`Comma-separated list of environment variables in KEY=VALUE format (e.g., 'API_KEY=secret,DEBUG=true').
These will be available to your Perl code.`),
		),
	)
}

func (p *SubprocessPerlTool) HandleExecution(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	logger.Debug("Subprocess Perl tool execution requested")

	code, err := request.RequireString("code")
	if err != nil {
		logger.Debug("Subprocess Perl tool execution failed: missing code argument")
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
		logger.Debug("Subprocess Perl environment variables: %v", envVars)
	}

	// No module installation for subprocess mode - pass empty slice
	output, err := p.executor.Execute(ctx, code, nil, envVars)
	if err != nil {
		logger.Debug("Subprocess Perl execution failed: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	logger.Debug("Subprocess Perl execution completed successfully")
	return mcp.NewToolResultText(output), nil
}
