// Package executor implements subprocess-based code execution for Python and Bash
// running directly on the host machine without containerization.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ylchen07/mcp-executor/internal/logger"
)

type SubprocessConfig struct {
	Binary       string
	InstallCmd   []string
	ExecutorName string
}

type SubprocessExecutor struct {
	config SubprocessConfig
}

func NewSubprocessPythonExecutor() *SubprocessExecutor {
	return &SubprocessExecutor{
		config: SubprocessConfig{
			Binary:       "python3",
			InstallCmd:   nil, // No pip installation in subprocess mode for security
			ExecutorName: "python-subprocess",
		},
	}
}

func NewSubprocessBashExecutor() *SubprocessExecutor {
	return &SubprocessExecutor{
		config: SubprocessConfig{
			Binary:       "bash",
			InstallCmd:   nil, // Skip dependency installation for bash
			ExecutorName: "bash-subprocess",
		},
	}
}

func (s *SubprocessExecutor) Execute(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
	logger.Debug("Starting %s execution", s.config.ExecutorName)

	// Install dependencies if needed and install command is available
	if len(dependencies) > 0 && s.config.InstallCmd != nil {
		logger.Debug("Installing dependencies: %v", dependencies)
		if err := s.installDependencies(ctx, dependencies); err != nil {
			return "", fmt.Errorf("failed to install dependencies: %v", err)
		}
	} else if len(dependencies) > 0 && s.config.InstallCmd == nil {
		logger.Debug("Skipping dependency installation for %s (not supported in subprocess mode)", s.config.ExecutorName)
	}

	// Execute the code
	logger.Verbose("Executing %s code in subprocess", s.config.ExecutorName)
	logger.Debug("Code to execute:\n%s", code)

	cmd := exec.CommandContext(ctx, s.config.Binary)
	cmd.Stdin = strings.NewReader(code)

	// Set environment variables
	cmd.Env = os.Environ() // Start with current environment
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Execution failed: %v", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s exited with code %d: %s", s.config.ExecutorName, exitError.ExitCode(), string(out))
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	logger.Debug("Execution completed successfully, output length: %d bytes", len(out))
	return string(out), nil
}

func (s *SubprocessExecutor) installDependencies(ctx context.Context, dependencies []string) error {
	args := append(s.config.InstallCmd, dependencies...)
	logger.Verbose("Running: %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Dependency installation failed: %v\nOutput: %s", err, string(out))
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	logger.Debug("Dependencies installed successfully")
	return nil
}
