// Package cmd provides the command-line interface using Cobra framework
// for the mcp-executor application with support for multiple transport modes.
package cmd

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
	Long: `Start the MCP server with the specified transport mode.

The server provides two main tools:
- execute-python: Run Python code in Docker containers with Playwright support
- execute-bash: Run bash scripts in isolated Ubuntu containers`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set global verbose flag
		logger.SetVerbose(verbose)

		mcpServer := server.NewMCPServer()

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

	// Add serve command to root
	rootCmd.AddCommand(serveCmd)
}
