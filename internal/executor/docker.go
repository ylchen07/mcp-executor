package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"github.com/ylchen07/mcp-python/internal/config"
)

type DockerExecutor struct{}

func NewDockerExecutor() *DockerExecutor {
	return &DockerExecutor{}
}

func (d *DockerExecutor) Execute(ctx context.Context, code string, modules []string) (string, error) {
	cmdArgs := []string{
		"run",
		"--rm",
		"-i",
		config.DockerImage,
	}
	shArgs := []string{}

	if len(modules) > 0 {
		shArgs = append(shArgs, "python", "-m", "pip", "install", "--quiet")
		shArgs = append(shArgs, modules...)
		shArgs = append(shArgs, "&&")
	}

	shArgs = append(shArgs, "python")
	cmdArgs = append(cmdArgs, "sh", "-c", strings.Join(shArgs, " "))

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("python exited with code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	return string(out), nil
}

