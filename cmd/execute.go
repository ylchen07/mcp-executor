package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ylchen07/mcp-python/internal/server"
)

var (
	// Global flags
	verbose bool
	version = "dev" // Will be set during build
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcp-executor",
	Short: "MCP server for Python and Bash execution",
	Long: `mcp-python is an MCP (Model Context Protocol) server that provides
both Python and Bash execution capabilities in isolated Docker containers.

It supports multiple transport modes: stdio (default), SSE, and HTTP.`,
	Version: version,
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the MCP server with the specified transport mode.

The server provides two main tools:
- execute-python: Run Python code in Docker containers with Playwright support
- execute-bash: Run bash scripts in isolated Ubuntu containers`,
	Run: func(cmd *cobra.Command, args []string) {
		mcpServer := server.NewMCPServer()

		var err error
		mode, _ := cmd.Flags().GetString("mode")

		switch mode {
		case "http":
			if verbose {
				fmt.Println("Starting MCP server in HTTP mode on port 8081")
			}
			err = server.RunHTTP(mcpServer)
		case "sse":
			if verbose {
				fmt.Println("Starting MCP server in SSE mode on port 8080")
			}
			err = server.RunSSE(mcpServer)
		default:
			if verbose {
				fmt.Println("Starting MCP server in stdio mode")
			}
			err = server.RunStdio(mcpServer)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of mcp-python`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-python version %s\n", version)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Serve command flags
	serveCmd.Flags().StringP("mode", "m", "stdio", "Transport mode: stdio, sse, or http")

	// Add commands to root
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// If no arguments provided, default to serve command
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "serve")
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
