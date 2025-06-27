// Package main provides the entry point for the mcp-executor application,
// an MCP (Model Context Protocol) server that executes Python and Bash code
// in isolated Docker containers.
package main

import "github.com/ylchen07/mcp-executor/cmd"

func main() {
	cmd.Execute()
}
