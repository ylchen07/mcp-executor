package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// SystemCheckPrompt generates a bash script to gather comprehensive host system information.
// This prompt is only available in subprocess execution mode to ensure accurate host system info.
type SystemCheckPrompt struct{}

// NewSystemCheckPrompt creates a new SystemCheckPrompt instance.
func NewSystemCheckPrompt() *SystemCheckPrompt {
	return &SystemCheckPrompt{}
}

// CreatePrompt defines the MCP prompt schema with optional detail_level argument.
func (p *SystemCheckPrompt) CreatePrompt() mcp.Prompt {
	return mcp.NewPrompt(
		"system-check",
		mcp.WithPromptDescription(
			"Gather comprehensive system information from the host machine including OS details, CPU, memory, disk usage, network interfaces, and running processes. Only available in subprocess execution mode.",
		),
		mcp.WithArgument(
			"detail_level",
			mcp.ArgumentDescription("Level of detail: 'basic' (default), 'detailed', or 'full'. Basic includes OS, CPU, memory, disk. Detailed adds network, processes, uptime. Full adds all filesystems, kernel params, environment."),
		),
	)
}

// HandlePrompt processes the prompt request and returns a formatted message with the bash script.
func (p *SystemCheckPrompt) HandlePrompt(
	ctx context.Context,
	request mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	// Parse detail level argument (default to "basic")
	detailLevel := "basic"
	if request.Params.Arguments != nil {
		if level, ok := request.Params.Arguments["detail_level"]; ok && level != "" {
			// Validate detail level
			switch strings.ToLower(level) {
			case "basic", "detailed", "full":
				detailLevel = strings.ToLower(level)
			default:
				detailLevel = "basic" // Fallback to basic for invalid values
			}
		}
	}

	// Generate the appropriate bash script
	script := generateSystemCheckScript(detailLevel)

	// Create the prompt message with instructions and script
	message := fmt.Sprintf(
		"I'll help you gather system information at the '%s' detail level.\n\n"+
			"⚠️  **Important**: This prompt is designed for subprocess execution mode to gather accurate host system information. "+
			"In Docker mode, you would only see container information, not the host system.\n\n"+
			"Execute this bash script using the execute-bash tool:\n\n"+
			"```bash\n%s\n```\n\n"+
			"This will provide:\n%s",
		detailLevel,
		script,
		getDetailLevelDescription(detailLevel),
	)

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(
			mcp.RoleAssistant,
			mcp.NewTextContent(message),
		),
	}

	return mcp.NewGetPromptResult(
		fmt.Sprintf("System check script (%s level)", detailLevel),
		messages,
	), nil
}

// generateSystemCheckScript creates a bash script based on the requested detail level.
func generateSystemCheckScript(level string) string {
	var script strings.Builder

	// All levels include basic information
	script.WriteString("#!/bin/bash\n")
	script.WriteString("echo '=== System Information ==='\n")
	script.WriteString("echo ''\n\n")

	// Basic level: OS, CPU, Memory, Disk
	script.WriteString("echo '--- Operating System ---'\n")
	script.WriteString("if command -v lsb_release &> /dev/null; then\n")
	script.WriteString("  lsb_release -a 2>/dev/null\n")
	script.WriteString("elif [ -f /etc/os-release ]; then\n")
	script.WriteString("  cat /etc/os-release\n")
	script.WriteString("else\n")
	script.WriteString("  uname -a\n")
	script.WriteString("fi\n")
	script.WriteString("echo ''\n\n")

	script.WriteString("echo '--- CPU Information ---'\n")
	script.WriteString("if command -v lscpu &> /dev/null; then\n")
	script.WriteString("  lscpu | grep -E 'Model name|Architecture|CPU\\(s\\)|Thread|Core|Socket'\n")
	script.WriteString("else\n")
	script.WriteString("  echo 'CPU(s):' $(nproc 2>/dev/null || grep -c ^processor /proc/cpuinfo)\n")
	script.WriteString("  grep 'model name' /proc/cpuinfo | head -n1 | cut -d':' -f2 | xargs\n")
	script.WriteString("fi\n")
	script.WriteString("echo ''\n\n")

	script.WriteString("echo '--- Memory Usage ---'\n")
	script.WriteString("if command -v free &> /dev/null; then\n")
	script.WriteString("  free -h\n")
	script.WriteString("else\n")
	script.WriteString("  cat /proc/meminfo | grep -E 'MemTotal|MemFree|MemAvailable'\n")
	script.WriteString("fi\n")
	script.WriteString("echo ''\n\n")

	script.WriteString("echo '--- Disk Usage (Root) ---'\n")
	script.WriteString("df -h / 2>/dev/null || echo 'df command not available'\n")
	script.WriteString("echo ''\n")

	// Detailed level adds: network, processes, uptime
	if level == "detailed" || level == "full" {
		script.WriteString("\necho '--- System Uptime ---'\n")
		script.WriteString("uptime\n")
		script.WriteString("echo ''\n\n")

		script.WriteString("echo '--- Network Interfaces ---'\n")
		script.WriteString("if command -v ip &> /dev/null; then\n")
		script.WriteString("  ip -brief addr show\n")
		script.WriteString("elif command -v ifconfig &> /dev/null; then\n")
		script.WriteString("  ifconfig | grep -E 'inet |ether '\n")
		script.WriteString("else\n")
		script.WriteString("  cat /proc/net/dev | grep ':' | awk '{print $1}'\n")
		script.WriteString("fi\n")
		script.WriteString("echo ''\n\n")

		script.WriteString("echo '--- Top 10 Processes by Memory ---'\n")
		script.WriteString("ps aux --sort=-%mem | head -n 11\n")
		script.WriteString("echo ''\n\n")

		script.WriteString("echo '--- Process Count ---'\n")
		script.WriteString("echo \"Total processes: $(ps aux | wc -l)\"\n")
		script.WriteString("echo ''\n")
	}

	// Full level adds: all filesystems, kernel params, environment, users
	if level == "full" {
		script.WriteString("\necho '--- All Mounted Filesystems ---'\n")
		script.WriteString("df -h 2>/dev/null || echo 'df command not available'\n")
		script.WriteString("echo ''\n\n")

		script.WriteString("echo '--- Kernel Parameters (sample) ---'\n")
		script.WriteString("if command -v sysctl &> /dev/null; then\n")
		script.WriteString("  sysctl -a 2>/dev/null | head -n 20\n")
		script.WriteString("  echo '... (showing first 20 parameters)'\n")
		script.WriteString("else\n")
		script.WriteString("  echo 'sysctl command not available'\n")
		script.WriteString("fi\n")
		script.WriteString("echo ''\n\n")

		script.WriteString("echo '--- Logged-in Users ---'\n")
		script.WriteString("who 2>/dev/null || w 2>/dev/null || echo 'User information not available'\n")
		script.WriteString("echo ''\n\n")

		script.WriteString("echo '--- Environment Variables (non-sensitive sample) ---'\n")
		script.WriteString("env | grep -E '^(PATH|HOME|USER|SHELL|LANG|TERM)=' | sort\n")
		script.WriteString("echo ''\n")
	}

	script.WriteString("\necho '=== System Check Complete ==='\n")

	return script.String()
}

// getDetailLevelDescription returns a human-readable description of what each level includes.
func getDetailLevelDescription(level string) string {
	switch level {
	case "basic":
		return "• OS name and version\n• CPU model and core count\n• Memory usage (total/used/free)\n• Root disk usage"
	case "detailed":
		return "• Everything in 'basic'\n• System uptime and load averages\n• Network interfaces and IP addresses\n• Top 10 processes by memory usage\n• Total process count"
	case "full":
		return "• Everything in 'detailed'\n• All mounted filesystems\n• Kernel parameters (sample)\n• Logged-in users\n• Environment variables (non-sensitive)"
	default:
		return ""
	}
}
