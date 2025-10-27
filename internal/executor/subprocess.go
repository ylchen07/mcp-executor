// Package executor implements subprocess-based code execution for Python and Bash
// running directly on the host machine without containerization.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// TypeScriptSubprocessExecutor is a specialized executor for TypeScript using ts-node
type TypeScriptSubprocessExecutor struct{}

func NewSubprocessTypeScriptExecutor() *TypeScriptSubprocessExecutor {
	return &TypeScriptSubprocessExecutor{}
}

func (t *TypeScriptSubprocessExecutor) Execute(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
	logger.Debug("Starting typescript-subprocess execution")

	if len(dependencies) > 0 {
		logger.Debug("Skipping dependency installation for typescript-subprocess (not supported in subprocess mode)")
	}

	// Create a temporary directory for the TypeScript file
	tmpDir, err := os.MkdirTemp("", "mcp-ts-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write code to a temporary .ts file
	tmpFile := filepath.Join(tmpDir, "index.ts")
	if err := os.WriteFile(tmpFile, []byte(code), 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %v", err)
	}

	logger.Verbose("Executing TypeScript code in subprocess")
	logger.Debug("Code to execute:\n%s", code)

	// Execute with ts-node (falls back to tsx, then npx tsx if not available)
	var cmd *exec.Cmd
	if _, err := exec.LookPath("ts-node"); err == nil {
		cmd = exec.CommandContext(ctx, "ts-node", tmpFile)
	} else if _, err := exec.LookPath("tsx"); err == nil {
		cmd = exec.CommandContext(ctx, "tsx", tmpFile)
	} else if _, err := exec.LookPath("npx"); err == nil {
		cmd = exec.CommandContext(ctx, "npx", "tsx", tmpFile)
	} else {
		return "", fmt.Errorf("neither ts-node, tsx, nor npx found on system - please install one to run TypeScript")
	}

	// Set environment variables
	cmd.Env = os.Environ() // Start with current environment
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Execution failed: %v", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("typescript-subprocess exited with code %d: %s", exitError.ExitCode(), string(out))
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	logger.Debug("Execution completed successfully, output length: %d bytes", len(out))
	return string(out), nil
}

// GoSubprocessExecutor is a specialized executor for Go that uses temporary files
type GoSubprocessExecutor struct{}

func NewSubprocessGoExecutor() *GoSubprocessExecutor {
	return &GoSubprocessExecutor{}
}

func (g *GoSubprocessExecutor) Execute(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
	logger.Debug("Starting go-subprocess execution")

	if len(dependencies) > 0 {
		logger.Debug("Skipping dependency installation for go-subprocess (not supported in subprocess mode)")
	}

	// Create a temporary directory for the Go file
	tmpDir, err := os.MkdirTemp("", "mcp-go-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write code to a temporary .go file
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %v", err)
	}

	logger.Verbose("Executing Go code in subprocess")
	logger.Debug("Code to execute:\n%s", code)

	// Execute with go run
	cmd := exec.CommandContext(ctx, "go", "run", tmpFile)

	// Set environment variables
	cmd.Env = os.Environ() // Start with current environment
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Execution failed: %v", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("go-subprocess exited with code %d: %s", exitError.ExitCode(), string(out))
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	logger.Debug("Execution completed successfully, output length: %d bytes", len(out))
	return string(out), nil
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
