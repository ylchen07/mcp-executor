// Package config provides centralized configuration constants
// for server identity, ports, and transport endpoints.
package config

const (
	ServerName    = "mcp-executor"
	ServerVersion = "1.0.0"
	SSEPort       = ":8080"
	SSEHost       = "http://localhost:8080"
	HTTPPort      = ":8081"
	HTTPHost      = "http://localhost:8081"
)
