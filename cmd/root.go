package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	Long: `mcp-executor is an MCP (Model Context Protocol) server that provides
both Python and Bash execution capabilities in isolated Docker containers.

It supports multiple transport modes: stdio (default), SSE, and HTTP.`,
	Version: version,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
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

