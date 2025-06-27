// Package executor implements Docker-based code execution for Python and Bash
// with support for dynamic dependency installation and isolated environments.
package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ylchen07/mcp-executor/internal/logger"
)

type ExecutorConfig struct {
	Image        string
	InstallCmd   []string
	ExecuteCmd   []string
	ExecutorName string
}

type DockerExecutor struct {
	config ExecutorConfig
}

func NewPythonExecutor() *DockerExecutor {
	return &DockerExecutor{
		config: ExecutorConfig{
			Image:        "mcr.microsoft.com/playwright/python:v1.53.0-noble",
			InstallCmd:   []string{"python", "-m", "pip", "install", "--quiet"},
			ExecuteCmd:   []string{"python"},
			ExecutorName: "python",
		},
	}
}

func NewBashExecutor() *DockerExecutor {
	return &DockerExecutor{
		config: ExecutorConfig{
			Image:        "ubuntu:22.04",
			InstallCmd:   []string{"apt-get", "update", "-qq", "&&", "apt-get", "install", "-y", "-qq"},
			ExecuteCmd:   []string{"bash"},
			ExecutorName: "bash",
		},
	}
}

func (d *DockerExecutor) Execute(ctx context.Context, code string, dependencies []string) (string, error) {
	logger.Debug("Starting %s execution", d.config.ExecutorName)

	cmdArgs := []string{
		"run",
		"--rm",
		"-i",
		d.config.Image,
	}
	shArgs := []string{}

	if len(dependencies) > 0 {
		logger.Debug("Installing dependencies: %v", dependencies)
		shArgs = append(shArgs, d.config.InstallCmd...)
		shArgs = append(shArgs, dependencies...)
		shArgs = append(shArgs, "&&")
	}

	shArgs = append(shArgs, d.config.ExecuteCmd...)
	cmdArgs = append(cmdArgs, "sh", "-c", strings.Join(shArgs, " "))

	logger.Verbose("Executing Docker command: docker %s", strings.Join(cmdArgs, " "))
	logger.Debug("Code to execute:\n%s", code)

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.Output()
	if err != nil {
		logger.Debug("Execution failed: %v", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s exited with code %d: %s", d.config.ExecutorName, exitError.ExitCode(), string(exitError.Stderr))
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	logger.Debug("Execution completed successfully, output length: %d bytes", len(out))
	return string(out), nil
}
