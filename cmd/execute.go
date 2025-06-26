package cmd

import (
	"flag"
	"log"

	"github.com/ylchen07/mcp-python/internal/server"
)

func Execute() {
	sseMode := flag.Bool("sse", false, "Run in SSE mode instead of stdio mode")
	httpMode := flag.Bool("http", false, "Run in HTTP mode instead of stdio mode")
	flag.Parse()

	mcpServer := server.NewMCPServer()

	var err error
	if *httpMode {
		err = server.RunHTTP(mcpServer)
	} else if *sseMode {
		err = server.RunSSE(mcpServer)
	} else {
		err = server.RunStdio(mcpServer)
	}

	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}