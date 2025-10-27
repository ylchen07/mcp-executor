// Package config provides centralized configuration constants
// for server identity, ports, transport endpoints, and Docker images.
package config

const (
	ServerName    = "mcp-executor"
	ServerVersion = "1.0.0"
	SSEPort       = ":8080"
	SSEHost       = "http://localhost:8080"
	HTTPPort      = ":8081"
	HTTPHost      = "http://localhost:8081"

	// Docker images for code execution
	PythonDockerImage     = "mcr.microsoft.com/playwright/python:v1.53.0-noble"
	BashDockerImage       = "ubuntu:22.04"
	TypeScriptDockerImage = "node:22-alpine"
	GoDockerImage         = "golang:1.23"
)
