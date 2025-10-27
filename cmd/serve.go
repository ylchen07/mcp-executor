// Package main provides the command-line interface using Cobra framework
// for the mcp-executor application with support for multiple transport modes.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ylchen07/mcp-executor/internal/logger"
	"github.com/ylchen07/mcp-executor/internal/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the MCP server with the specified transport mode and execution mode.

The server provides four main tools:
- execute-python: Run Python code (subprocess mode by default, Docker optional)
- execute-bash: Run bash scripts (subprocess mode by default, Docker optional)
- execute-perl: Run Perl code (subprocess mode by default, Docker optional)
- execute-go: Run Go code (subprocess mode by default, Docker optional)

Execution modes:
- subprocess: Run code directly on host (default, faster, less isolated)
- docker: Run code in Docker containers (slower, fully isolated)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set global verbose flag
		logger.SetVerbose(verbose)

		executionMode, _ := cmd.Flags().GetString("execution-mode")
		mcpServer := server.NewMCPServer(executionMode)

		var err error
		mode, _ := cmd.Flags().GetString("mode")

		switch mode {
		case "http":
			logger.VerbosePrint("Starting MCP server in HTTP mode on port 8081")
			err = server.RunHTTP(mcpServer)
		case "sse":
			logger.VerbosePrint("Starting MCP server in SSE mode on port 8080")
			err = server.RunSSE(mcpServer)
		default:
			logger.VerbosePrint("Starting MCP server in stdio mode")
			err = server.RunStdio(mcpServer)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Serve command flags
	serveCmd.Flags().StringP("mode", "m", "stdio", "Transport mode: stdio, sse, or http")
	serveCmd.Flags().StringP("execution-mode", "e", "subprocess", "Execution mode: subprocess or docker")

	// Add serve command to root
	rootCmd.AddCommand(serveCmd)
}
