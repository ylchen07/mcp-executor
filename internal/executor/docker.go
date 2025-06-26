package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
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
	cmdArgs := []string{
		"run",
		"--rm",
		"-i",
		d.config.Image,
	}
	shArgs := []string{}

	if len(dependencies) > 0 {
		shArgs = append(shArgs, d.config.InstallCmd...)
		shArgs = append(shArgs, dependencies...)
		shArgs = append(shArgs, "&&")
	}

	shArgs = append(shArgs, d.config.ExecuteCmd...)
	cmdArgs = append(cmdArgs, "sh", "-c", strings.Join(shArgs, " "))

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s exited with code %d: %s", d.config.ExecutorName, exitError.ExitCode(), string(exitError.Stderr))
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	return string(out), nil
}
