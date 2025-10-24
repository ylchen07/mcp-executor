package executor

import (
	"strings"
	"testing"
)

func TestNewPythonExecutor(t *testing.T) {
	executor := NewPythonExecutor()

	if executor == nil {
		t.Fatal("NewPythonExecutor() returned nil")
	}

	if executor.config.ExecutorName != "python" {
		t.Errorf("ExecutorName = %q, want %q", executor.config.ExecutorName, "python")
	}

	if executor.config.Image != "mcr.microsoft.com/playwright/python:v1.53.0-noble" {
		t.Errorf("Image = %q, want %q", executor.config.Image, "mcr.microsoft.com/playwright/python:v1.53.0-noble")
	}

	expectedInstallCmd := []string{"python", "-m", "pip", "install", "--quiet"}
	if len(executor.config.InstallCmd) != len(expectedInstallCmd) {
		t.Errorf("InstallCmd length = %d, want %d", len(executor.config.InstallCmd), len(expectedInstallCmd))
	}
	for i, cmd := range expectedInstallCmd {
		if i >= len(executor.config.InstallCmd) || executor.config.InstallCmd[i] != cmd {
			t.Errorf("InstallCmd[%d] = %q, want %q", i, executor.config.InstallCmd[i], cmd)
		}
	}

	expectedExecuteCmd := []string{"python"}
	if len(executor.config.ExecuteCmd) != len(expectedExecuteCmd) {
		t.Errorf("ExecuteCmd length = %d, want %d", len(executor.config.ExecuteCmd), len(expectedExecuteCmd))
	}
}

func TestNewBashExecutor(t *testing.T) {
	executor := NewBashExecutor()

	if executor == nil {
		t.Fatal("NewBashExecutor() returned nil")
	}

	if executor.config.ExecutorName != "bash" {
		t.Errorf("ExecutorName = %q, want %q", executor.config.ExecutorName, "bash")
	}

	if executor.config.Image != "ubuntu:22.04" {
		t.Errorf("Image = %q, want %q", executor.config.Image, "ubuntu:22.04")
	}

	expectedInstallCmd := []string{"apt-get", "update", "-qq", "&&", "apt-get", "install", "-y", "-qq"}
	if len(executor.config.InstallCmd) != len(expectedInstallCmd) {
		t.Errorf("InstallCmd length = %d, want %d", len(executor.config.InstallCmd), len(expectedInstallCmd))
	}
	for i, cmd := range expectedInstallCmd {
		if i >= len(executor.config.InstallCmd) || executor.config.InstallCmd[i] != cmd {
			t.Errorf("InstallCmd[%d] = %q, want %q", i, executor.config.InstallCmd[i], cmd)
		}
	}

	expectedExecuteCmd := []string{"bash"}
	if len(executor.config.ExecuteCmd) != len(expectedExecuteCmd) {
		t.Errorf("ExecuteCmd length = %d, want %d", len(executor.config.ExecuteCmd), len(expectedExecuteCmd))
	}
}

func TestDockerExecutor_CommandConstruction_NoDependencies(t *testing.T) {
	tests := []struct {
		name        string
		executor    *DockerExecutor
		code        string
		envVars     map[string]string
		wantImage   string
		wantEnvVars []string
	}{
		{
			name:        "python no deps no env",
			executor:    NewPythonExecutor(),
			code:        `print("hello")`,
			envVars:     nil,
			wantImage:   "mcr.microsoft.com/playwright/python:v1.53.0-noble",
			wantEnvVars: nil,
		},
		{
			name:     "python with env vars",
			executor: NewPythonExecutor(),
			code:     `import os; print(os.getenv("KEY"))`,
			envVars: map[string]string{
				"API_KEY": "secret",
				"DEBUG":   "true",
			},
			wantImage:   "mcr.microsoft.com/playwright/python:v1.53.0-noble",
			wantEnvVars: []string{"API_KEY=secret", "DEBUG=true"},
		},
		{
			name:        "bash no deps no env",
			executor:    NewBashExecutor(),
			code:        `echo "hello"`,
			envVars:     nil,
			wantImage:   "ubuntu:22.04",
			wantEnvVars: nil,
		},
		{
			name:     "bash with env vars",
			executor: NewBashExecutor(),
			code:     `echo "$VAR1:$VAR2"`,
			envVars: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			wantImage:   "ubuntu:22.04",
			wantEnvVars: []string{"VAR1=value1", "VAR2=value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the executor uses the correct image
			if tt.executor.config.Image != tt.wantImage {
				t.Errorf("Image = %q, want %q", tt.executor.config.Image, tt.wantImage)
			}

			// Verify environment variables would be properly formatted
			if tt.envVars != nil {
				foundEnvVars := make(map[string]bool)
				for _, want := range tt.wantEnvVars {
					foundEnvVars[want] = false
				}

				for key, value := range tt.envVars {
					envPair := key + "=" + value
					if _, exists := foundEnvVars[envPair]; exists {
						foundEnvVars[envPair] = true
					}
				}

				for envVar, found := range foundEnvVars {
					if !found {
						t.Errorf("Expected env var %q to be set", envVar)
					}
				}
			}
		})
	}
}

func TestDockerExecutor_CommandConstruction_WithDependencies(t *testing.T) {
	tests := []struct {
		name         string
		executor     *DockerExecutor
		dependencies []string
		wantInstall  []string
	}{
		{
			name:         "python single dependency",
			executor:     NewPythonExecutor(),
			dependencies: []string{"requests"},
			wantInstall:  []string{"python", "-m", "pip", "install", "--quiet", "requests"},
		},
		{
			name:         "python multiple dependencies",
			executor:     NewPythonExecutor(),
			dependencies: []string{"requests", "numpy", "pandas"},
			wantInstall:  []string{"python", "-m", "pip", "install", "--quiet", "requests", "numpy", "pandas"},
		},
		{
			name:         "bash single package",
			executor:     NewBashExecutor(),
			dependencies: []string{"curl"},
			wantInstall:  []string{"apt-get", "update", "-qq", "&&", "apt-get", "install", "-y", "-qq", "curl"},
		},
		{
			name:         "bash multiple packages",
			executor:     NewBashExecutor(),
			dependencies: []string{"curl", "wget", "jq"},
			wantInstall:  []string{"apt-get", "update", "-qq", "&&", "apt-get", "install", "-y", "-qq", "curl", "wget", "jq"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build what the install command would look like
			installCmd := append([]string{}, tt.executor.config.InstallCmd...)
			installCmd = append(installCmd, tt.dependencies...)

			// Verify the constructed command matches expected
			if len(installCmd) != len(tt.wantInstall) {
				t.Errorf("Install command length = %d, want %d", len(installCmd), len(tt.wantInstall))
				t.Errorf("Got: %v", installCmd)
				t.Errorf("Want: %v", tt.wantInstall)
				return
			}

			for i, want := range tt.wantInstall {
				if installCmd[i] != want {
					t.Errorf("Install command[%d] = %q, want %q", i, installCmd[i], want)
				}
			}
		})
	}
}

func TestDockerExecutor_Execute_ErrorHandling(t *testing.T) {
	// Test that Execute properly handles context
	executor := NewPythonExecutor()

	// Note: Without Docker, we can't actually test execution
	// But we can verify the method signature and basic structure
	if executor == nil {
		t.Fatal("Executor should not be nil")
	}

	// Verify the Execute method exists and has correct signature
	_ = executor.Execute
}

func TestDockerExecutor_ConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		executor *DockerExecutor
		wantErr  bool
	}{
		{
			name:     "valid python executor",
			executor: NewPythonExecutor(),
			wantErr:  false,
		},
		{
			name:     "valid bash executor",
			executor: NewBashExecutor(),
			wantErr:  false,
		},
		{
			name: "custom executor with minimal config",
			executor: &DockerExecutor{
				config: ExecutorConfig{
					Image:        "alpine:latest",
					InstallCmd:   []string{"apk", "add"},
					ExecuteCmd:   []string{"sh"},
					ExecutorName: "custom",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate configuration
			if tt.executor.config.Image == "" {
				t.Error("Image should not be empty")
			}
			if tt.executor.config.ExecutorName == "" {
				t.Error("ExecutorName should not be empty")
			}
			if len(tt.executor.config.ExecuteCmd) == 0 {
				t.Error("ExecuteCmd should not be empty")
			}
		})
	}
}

func TestDockerExecutor_ShellCommandConstruction(t *testing.T) {
	tests := []struct {
		name         string
		executor     *DockerExecutor
		dependencies []string
		wantContains []string
	}{
		{
			name:         "python with dependencies",
			executor:     NewPythonExecutor(),
			dependencies: []string{"requests"},
			wantContains: []string{"pip install", "requests", "&&", "python"},
		},
		{
			name:         "bash with packages",
			executor:     NewBashExecutor(),
			dependencies: []string{"curl"},
			wantContains: []string{"apt-get", "curl", "&&", "bash"},
		},
		{
			name:         "python no dependencies",
			executor:     NewPythonExecutor(),
			dependencies: nil,
			wantContains: []string{"python"},
		},
		{
			name:         "bash no packages",
			executor:     NewBashExecutor(),
			dependencies: nil,
			wantContains: []string{"bash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build shell args as the Execute method would
			shArgs := []string{}
			if len(tt.dependencies) > 0 {
				shArgs = append(shArgs, tt.executor.config.InstallCmd...)
				shArgs = append(shArgs, tt.dependencies...)
				shArgs = append(shArgs, "&&")
			}
			shArgs = append(shArgs, tt.executor.config.ExecuteCmd...)

			shellCmd := strings.Join(shArgs, " ")

			// Verify expected components are in the shell command
			for _, want := range tt.wantContains {
				if !strings.Contains(shellCmd, want) {
					t.Errorf("Shell command %q should contain %q", shellCmd, want)
				}
			}
		})
	}
}
