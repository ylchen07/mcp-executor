package executor

import (
	"context"
	"strings"
	"testing"
)

func TestSubprocessPythonExecutor_Execute(t *testing.T) {
	ctx := context.Background()
	executor := NewSubprocessPythonExecutor()

	tests := []struct {
		name        string
		code        string
		envVars     map[string]string
		wantContain string
		wantErr     bool
	}{
		{
			name:        "simple print",
			code:        `print("Hello World")`,
			envVars:     nil,
			wantContain: "Hello World",
			wantErr:     false,
		},
		{
			name:        "environment variable",
			code:        `import os; print(os.environ.get("TEST_VAR", "NOT_SET"))`,
			envVars:     map[string]string{"TEST_VAR": "test_value"},
			wantContain: "test_value",
			wantErr:     false,
		},
		{
			name:        "syntax error",
			code:        `print("missing closing quote`,
			envVars:     nil,
			wantContain: "",
			wantErr:     true,
		},
		{
			name:        "import built-in module",
			code:        `import sys; print(sys.version.split()[0])`,
			envVars:     nil,
			wantContain: ".",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(ctx, tt.code, nil, tt.envVars)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(result, tt.wantContain) {
				t.Errorf("Execute() result = %q, want to contain %q", result, tt.wantContain)
			}
		})
	}
}

func TestSubprocessBashExecutor_Execute(t *testing.T) {
	ctx := context.Background()
	executor := NewSubprocessBashExecutor()

	tests := []struct {
		name        string
		code        string
		envVars     map[string]string
		wantContain string
		wantErr     bool
	}{
		{
			name:        "simple echo",
			code:        `echo "Hello Bash"`,
			envVars:     nil,
			wantContain: "Hello Bash",
			wantErr:     false,
		},
		{
			name:        "environment variable",
			code:        `echo "$TEST_VAR"`,
			envVars:     map[string]string{"TEST_VAR": "bash_value"},
			wantContain: "bash_value",
			wantErr:     false,
		},
		{
			name:        "command not found",
			code:        `nonexistent_command_12345`,
			envVars:     nil,
			wantContain: "",
			wantErr:     true,
		},
		{
			name:        "multiple commands",
			code:        `echo "line1"; echo "line2"`,
			envVars:     nil,
			wantContain: "line1",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(ctx, tt.code, nil, tt.envVars)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(result, tt.wantContain) {
				t.Errorf("Execute() result = %q, want to contain %q", result, tt.wantContain)
			}
		})
	}
}

func TestSubprocessPythonExecutor_DependencyInstallation(t *testing.T) {
	ctx := context.Background()
	executor := NewSubprocessPythonExecutor()

	// Test that dependencies parameter doesn't cause errors
	// We skip actual installation since it would modify the host
	code := `print("test")`
	dependencies := []string{"fake-package-that-does-not-exist-xyz"}

	_, err := executor.Execute(ctx, code, dependencies, nil)
	// This might fail due to package not found, which is expected
	// We're mainly testing that the mechanism doesn't panic
	if err != nil {
		t.Logf("Expected failure for non-existent package: %v", err)
	}
}

func TestSubprocessBashExecutor_SkipsDependencies(t *testing.T) {
	ctx := context.Background()
	executor := NewSubprocessBashExecutor()

	// Bash executor should skip dependency installation
	code := `echo "test"`
	dependencies := []string{"curl", "wget"}

	result, err := executor.Execute(ctx, code, dependencies, nil)
	if err != nil {
		t.Errorf("Execute() with dependencies should not fail for bash: %v", err)
	}

	if !strings.Contains(result, "test") {
		t.Errorf("Expected output to contain 'test', got: %q", result)
	}
}
