package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// mockExecutor implements the executor.Executor interface for testing
type mockExecutor struct {
	executeFunc func(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error)
	lastCode    string
	lastDeps    []string
	lastEnvVars map[string]string
}

func (m *mockExecutor) Execute(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
	m.lastCode = code
	m.lastDeps = dependencies
	m.lastEnvVars = envVars

	if m.executeFunc != nil {
		return m.executeFunc(ctx, code, dependencies, envVars)
	}
	return "mock output", nil
}

func TestNewPythonTool(t *testing.T) {
	mockExec := &mockExecutor{}
	tool := NewPythonTool(mockExec)

	if tool == nil {
		t.Fatal("NewPythonTool() returned nil")
	}

	if tool.executor == nil {
		t.Error("NewPythonTool() executor should not be nil")
	}
}

func TestPythonTool_CreateTool(t *testing.T) {
	mockExec := &mockExecutor{}
	pythonTool := NewPythonTool(mockExec)

	tool := pythonTool.CreateTool()

	if tool.Name != "execute-python" {
		t.Errorf("Tool name = %q, want %q", tool.Name, "execute-python")
	}

	if tool.Description == "" {
		t.Error("Tool description should not be empty")
	}

	// Verify required parameters
	if tool.InputSchema.Properties == nil {
		t.Fatal("Tool should have input schema properties")
	}

	// Check that 'code' parameter exists and is required
	codeSchema, hasCode := tool.InputSchema.Properties["code"]
	if !hasCode {
		t.Error("Tool should have 'code' parameter")
	}
	if codeSchema == nil {
		t.Error("Code parameter schema should not be nil")
	}

	// Check optional parameters exist
	if _, hasModules := tool.InputSchema.Properties["modules"]; !hasModules {
		t.Error("Tool should have 'modules' parameter")
	}

	if _, hasEnv := tool.InputSchema.Properties["env"]; !hasEnv {
		t.Error("Tool should have 'env' parameter")
	}
}

func TestPythonTool_HandleExecution(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		mockOutput   string
		mockError    error
		wantErr      bool
		wantResult   string
		checkDeps    []string
		checkEnvVars map[string]string
	}{
		{
			name: "simple code execution",
			params: map[string]interface{}{
				"code": `print("hello")`,
			},
			mockOutput: "hello\n",
			mockError:  nil,
			wantErr:    false,
			wantResult: "hello",
			checkDeps:  nil,
		},
		{
			name: "with single module",
			params: map[string]interface{}{
				"code":    `import requests`,
				"modules": "requests",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkDeps:  []string{"requests"},
		},
		{
			name: "with multiple modules",
			params: map[string]interface{}{
				"code":    `import requests, numpy`,
				"modules": "requests,numpy,pandas",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkDeps:  []string{"requests", "numpy", "pandas"},
		},
		{
			name: "with modules containing spaces",
			params: map[string]interface{}{
				"code":    `import requests`,
				"modules": "requests , numpy , pandas",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkDeps:  []string{"requests ", " numpy ", " pandas"},
		},
		{
			name: "with single env var",
			params: map[string]interface{}{
				"code": `import os; print(os.getenv("API_KEY"))`,
				"env":  "API_KEY=secret123",
			},
			mockOutput: "secret123",
			mockError:  nil,
			wantErr:    false,
			wantResult: "secret123",
			checkEnvVars: map[string]string{
				"API_KEY": "secret123",
			},
		},
		{
			name: "with multiple env vars",
			params: map[string]interface{}{
				"code": `import os`,
				"env":  "API_KEY=secret123,DEBUG=true,PORT=8080",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkEnvVars: map[string]string{
				"API_KEY": "secret123",
				"DEBUG":   "true",
				"PORT":    "8080",
			},
		},
		{
			name: "with env vars containing spaces",
			params: map[string]interface{}{
				"code": `import os`,
				"env":  "API_KEY=secret123 , DEBUG=true , PORT=8080",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkEnvVars: map[string]string{
				"API_KEY": "secret123",
				"DEBUG":   "true",
				"PORT":    "8080",
			},
		},
		{
			name: "with env var containing equals sign in value",
			params: map[string]interface{}{
				"code": `import os`,
				"env":  "CONNECTION_STRING=server=localhost;user=admin",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkEnvVars: map[string]string{
				"CONNECTION_STRING": "server=localhost;user=admin",
			},
		},
		{
			name: "with modules and env vars",
			params: map[string]interface{}{
				"code":    `import requests`,
				"modules": "requests,numpy",
				"env":     "API_KEY=secret,DEBUG=true",
			},
			mockOutput: "success",
			mockError:  nil,
			wantErr:    false,
			wantResult: "success",
			checkDeps:  []string{"requests", "numpy"},
			checkEnvVars: map[string]string{
				"API_KEY": "secret",
				"DEBUG":   "true",
			},
		},
		{
			name: "empty code parameter",
			params: map[string]interface{}{
				"code": "",
			},
			mockOutput: "",
			mockError:  nil,
			wantErr:    false,
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				executeFunc: func(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}

			pythonTool := NewPythonTool(mockExec)
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "execute-python",
					Arguments: tt.params,
				},
			}

			result, err := pythonTool.HandleExecution(context.Background(), request)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleExecution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil && result.Content != nil {
				if len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(mcp.TextContent); ok {
						if !strings.Contains(textContent.Text, tt.wantResult) {
							t.Errorf("HandleExecution() result = %q, want to contain %q", textContent.Text, tt.wantResult)
						}
					}
				}
			}

			// Check that dependencies were passed correctly
			if tt.checkDeps != nil {
				if len(mockExec.lastDeps) != len(tt.checkDeps) {
					t.Errorf("Dependencies count = %d, want %d", len(mockExec.lastDeps), len(tt.checkDeps))
				} else {
					for i, dep := range tt.checkDeps {
						if mockExec.lastDeps[i] != dep {
							t.Errorf("Dependency[%d] = %q, want %q", i, mockExec.lastDeps[i], dep)
						}
					}
				}
			}

			// Check that environment variables were passed correctly
			if tt.checkEnvVars != nil {
				if len(mockExec.lastEnvVars) != len(tt.checkEnvVars) {
					t.Errorf("EnvVars count = %d, want %d", len(mockExec.lastEnvVars), len(tt.checkEnvVars))
				}
				for key, expectedValue := range tt.checkEnvVars {
					if actualValue, ok := mockExec.lastEnvVars[key]; !ok {
						t.Errorf("EnvVar %q not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("EnvVar[%q] = %q, want %q", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestPythonTool_HandleExecution_MissingCode(t *testing.T) {
	mockExec := &mockExecutor{}
	pythonTool := NewPythonTool(mockExec)

	// Request without 'code' parameter
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "execute-python",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := pythonTool.HandleExecution(context.Background(), request)

	// Should not return an error (handled gracefully)
	if err != nil {
		t.Errorf("HandleExecution() should not return error for missing code, got: %v", err)
	}

	// Should return an error result
	if result == nil {
		t.Fatal("HandleExecution() should return a result")
	}

	if !result.IsError {
		t.Error("HandleExecution() result should be an error when code is missing")
	}
}

func TestPythonTool_HandleExecution_ExecutorError(t *testing.T) {
	mockExec := &mockExecutor{
		executeFunc: func(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
			return "", &ExecutorError{Message: "execution failed"}
		},
	}

	pythonTool := NewPythonTool(mockExec)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-python",
			Arguments: map[string]interface{}{
				"code": `print("test")`,
			},
		},
	}

	result, err := pythonTool.HandleExecution(context.Background(), request)

	if err != nil {
		t.Errorf("HandleExecution() should not return error, errors should be in result, got: %v", err)
	}

	if result == nil {
		t.Fatal("HandleExecution() should return a result")
	}

	if !result.IsError {
		t.Error("HandleExecution() result should be an error when executor fails")
	}
}

func TestPythonTool_HandleExecution_EmptyModules(t *testing.T) {
	mockExec := &mockExecutor{}
	pythonTool := NewPythonTool(mockExec)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-python",
			Arguments: map[string]interface{}{
				"code":    `print("test")`,
				"modules": "",
			},
		},
	}

	_, err := pythonTool.HandleExecution(context.Background(), request)
	if err != nil {
		t.Errorf("HandleExecution() should handle empty modules string, got error: %v", err)
	}

	// Should have no dependencies
	if len(mockExec.lastDeps) != 0 {
		t.Errorf("Expected no dependencies for empty modules string, got: %v", mockExec.lastDeps)
	}
}

func TestPythonTool_HandleExecution_EmptyEnv(t *testing.T) {
	mockExec := &mockExecutor{}
	pythonTool := NewPythonTool(mockExec)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-python",
			Arguments: map[string]interface{}{
				"code": `print("test")`,
				"env":  "",
			},
		},
	}

	_, err := pythonTool.HandleExecution(context.Background(), request)
	if err != nil {
		t.Errorf("HandleExecution() should handle empty env string, got error: %v", err)
	}

	// Should have no env vars
	if len(mockExec.lastEnvVars) != 0 {
		t.Errorf("Expected no env vars for empty env string, got: %v", mockExec.lastEnvVars)
	}
}

// ExecutorError is a simple error type for testing
type ExecutorError struct {
	Message string
}

func (e ExecutorError) Error() string {
	return e.Message
}
