package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewBashTool(t *testing.T) {
	mockExec := &mockExecutor{}
	tool := NewBashTool(mockExec)

	if tool == nil {
		t.Fatal("NewBashTool() returned nil")
	}

	if tool.executor == nil {
		t.Error("NewBashTool() executor should not be nil")
	}
}

func TestBashTool_CreateTool(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewBashTool(mockExec)

	tool := bashTool.CreateTool()

	if tool.Name != "execute-bash" {
		t.Errorf("Tool name = %q, want %q", tool.Name, "execute-bash")
	}

	if tool.Description == "" {
		t.Error("Tool description should not be empty")
	}

	// Verify required parameters
	if tool.InputSchema.Properties == nil {
		t.Fatal("Tool should have input schema properties")
	}

	// Check that 'script' parameter exists and is required
	scriptSchema, hasScript := tool.InputSchema.Properties["script"]
	if !hasScript {
		t.Error("Tool should have 'script' parameter")
	}
	if scriptSchema == nil {
		t.Error("Script parameter schema should not be nil")
	}

	// Check optional parameters exist
	if _, hasPackages := tool.InputSchema.Properties["packages"]; !hasPackages {
		t.Error("Tool should have 'packages' parameter")
	}

	if _, hasEnv := tool.InputSchema.Properties["env"]; !hasEnv {
		t.Error("Tool should have 'env' parameter")
	}
}

func TestBashTool_HandleExecution(t *testing.T) {
	tests := []struct {
		name          string
		params        map[string]interface{}
		mockOutput    string
		mockError     error
		wantErr       bool
		wantResult    string
		checkPackages []string
		checkEnvVars  map[string]string
	}{
		{
			name: "simple script execution",
			params: map[string]interface{}{
				"script": `echo "hello"`,
			},
			mockOutput:    "hello\n",
			mockError:     nil,
			wantErr:       false,
			wantResult:    "hello",
			checkPackages: nil,
		},
		{
			name: "with single package",
			params: map[string]interface{}{
				"script":   `curl --version`,
				"packages": "curl",
			},
			mockOutput:    "success",
			mockError:     nil,
			wantErr:       false,
			wantResult:    "success",
			checkPackages: []string{"curl"},
		},
		{
			name: "with multiple packages",
			params: map[string]interface{}{
				"script":   `curl --version && wget --version`,
				"packages": "curl,wget,jq",
			},
			mockOutput:    "success",
			mockError:     nil,
			wantErr:       false,
			wantResult:    "success",
			checkPackages: []string{"curl", "wget", "jq"},
		},
		{
			name: "with packages containing spaces",
			params: map[string]interface{}{
				"script":   `curl --version`,
				"packages": "curl , wget , jq",
			},
			mockOutput:    "success",
			mockError:     nil,
			wantErr:       false,
			wantResult:    "success",
			checkPackages: []string{"curl", "wget", "jq"},
		},
		{
			name: "with single env var",
			params: map[string]interface{}{
				"script": `echo "$API_KEY"`,
				"env":    "API_KEY=secret123",
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
				"script": `echo "$API_KEY:$DEBUG:$PORT"`,
				"env":    "API_KEY=secret123,DEBUG=true,PORT=8080",
			},
			mockOutput: "secret123:true:8080",
			mockError:  nil,
			wantErr:    false,
			wantResult: "secret123:true:8080",
			checkEnvVars: map[string]string{
				"API_KEY": "secret123",
				"DEBUG":   "true",
				"PORT":    "8080",
			},
		},
		{
			name: "with env vars containing spaces",
			params: map[string]interface{}{
				"script": `echo "$VAR1"`,
				"env":    "VAR1=value1 , VAR2=value2 , VAR3=value3",
			},
			mockOutput: "value1",
			mockError:  nil,
			wantErr:    false,
			wantResult: "value1",
			checkEnvVars: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
				"VAR3": "value3",
			},
		},
		{
			name: "with env var containing equals sign in value",
			params: map[string]interface{}{
				"script": `echo "$CONNECTION_STRING"`,
				"env":    "CONNECTION_STRING=server=localhost;user=admin",
			},
			mockOutput: "server=localhost;user=admin",
			mockError:  nil,
			wantErr:    false,
			wantResult: "server=localhost;user=admin",
			checkEnvVars: map[string]string{
				"CONNECTION_STRING": "server=localhost;user=admin",
			},
		},
		{
			name: "with packages and env vars",
			params: map[string]interface{}{
				"script":   `curl --version`,
				"packages": "curl,wget",
				"env":      "API_KEY=secret,DEBUG=true",
			},
			mockOutput:    "success",
			mockError:     nil,
			wantErr:       false,
			wantResult:    "success",
			checkPackages: []string{"curl", "wget"},
			checkEnvVars: map[string]string{
				"API_KEY": "secret",
				"DEBUG":   "true",
			},
		},
		{
			name: "empty script parameter",
			params: map[string]interface{}{
				"script": "",
			},
			mockOutput: "",
			mockError:  nil,
			wantErr:    false,
			wantResult: "",
		},
		{
			name: "multiline script",
			params: map[string]interface{}{
				"script": "#!/bin/bash\necho 'line1'\necho 'line2'\necho 'line3'",
			},
			mockOutput: "line1\nline2\nline3\n",
			mockError:  nil,
			wantErr:    false,
			wantResult: "line1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				executeFunc: func(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}

			bashTool := NewBashTool(mockExec)
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "execute-bash",
					Arguments: tt.params,
				},
			}

			result, err := bashTool.HandleExecution(context.Background(), request)

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

			// Check that packages were passed correctly
			if tt.checkPackages != nil {
				if len(mockExec.lastDeps) != len(tt.checkPackages) {
					t.Errorf("Packages count = %d, want %d", len(mockExec.lastDeps), len(tt.checkPackages))
				} else {
					for i, pkg := range tt.checkPackages {
						if mockExec.lastDeps[i] != pkg {
							t.Errorf("Package[%d] = %q, want %q", i, mockExec.lastDeps[i], pkg)
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

func TestBashTool_HandleExecution_MissingScript(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewBashTool(mockExec)

	// Request without 'script' parameter
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "execute-bash",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := bashTool.HandleExecution(context.Background(), request)
	// Should not return an error (handled gracefully)
	if err != nil {
		t.Errorf("HandleExecution() should not return error for missing script, got: %v", err)
	}

	// Should return an error result
	if result == nil {
		t.Fatal("HandleExecution() should return a result")
	}

	if !result.IsError {
		t.Error("HandleExecution() result should be an error when script is missing")
	}
}

func TestBashTool_HandleExecution_ExecutorError(t *testing.T) {
	mockExec := &mockExecutor{
		executeFunc: func(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
			return "", &ExecutorError{Message: "execution failed"}
		},
	}

	bashTool := NewBashTool(mockExec)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-bash",
			Arguments: map[string]interface{}{
				"script": `echo "test"`,
			},
		},
	}

	result, err := bashTool.HandleExecution(context.Background(), request)
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

func TestBashTool_HandleExecution_EmptyPackages(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewBashTool(mockExec)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-bash",
			Arguments: map[string]interface{}{
				"script":   `echo "test"`,
				"packages": "",
			},
		},
	}

	_, err := bashTool.HandleExecution(context.Background(), request)
	if err != nil {
		t.Errorf("HandleExecution() should handle empty packages string, got error: %v", err)
	}

	// Should have no packages
	if len(mockExec.lastDeps) != 0 {
		t.Errorf("Expected no packages for empty packages string, got: %v", mockExec.lastDeps)
	}
}

func TestBashTool_HandleExecution_EmptyEnv(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewBashTool(mockExec)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-bash",
			Arguments: map[string]interface{}{
				"script": `echo "test"`,
				"env":    "",
			},
		},
	}

	_, err := bashTool.HandleExecution(context.Background(), request)
	if err != nil {
		t.Errorf("HandleExecution() should handle empty env string, got error: %v", err)
	}

	// Should have no env vars
	if len(mockExec.lastEnvVars) != 0 {
		t.Errorf("Expected no env vars for empty env string, got: %v", mockExec.lastEnvVars)
	}
}

func TestBashTool_HandleExecution_ComplexEnvVarParsing(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewBashTool(mockExec)

	tests := []struct {
		name         string
		envString    string
		expectedVars map[string]string
	}{
		{
			name:      "simple key=value",
			envString: "KEY=value",
			expectedVars: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:      "value with equals sign",
			envString: "DB_URL=postgres://user:pass@localhost:5432/db",
			expectedVars: map[string]string{
				"DB_URL": "postgres://user:pass@localhost:5432/db",
			},
		},
		{
			name:      "empty value",
			envString: "EMPTY=",
			expectedVars: map[string]string{
				"EMPTY": "",
			},
		},
		{
			name:      "value with commas",
			envString: "TAGS=tag1;tag2;tag3,OWNER=admin",
			expectedVars: map[string]string{
				"TAGS":  "tag1;tag2;tag3",
				"OWNER": "admin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "execute-bash",
					Arguments: map[string]interface{}{
						"script": `echo "test"`,
						"env":    tt.envString,
					},
				},
			}

			_, err := bashTool.HandleExecution(context.Background(), request)
			if err != nil {
				t.Errorf("HandleExecution() error = %v", err)
				return
			}

			for key, expectedValue := range tt.expectedVars {
				if actualValue, ok := mockExec.lastEnvVars[key]; !ok {
					t.Errorf("EnvVar %q not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("EnvVar[%q] = %q, want %q", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestBashTool_HandleExecution_PackagesParsing(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewBashTool(mockExec)

	tests := []struct {
		name             string
		packagesString   string
		expectedPackages []string
	}{
		{
			name:             "single package",
			packagesString:   "curl",
			expectedPackages: []string{"curl"},
		},
		{
			name:             "multiple packages",
			packagesString:   "curl,wget,jq",
			expectedPackages: []string{"curl", "wget", "jq"},
		},
		{
			name:             "packages with spaces",
			packagesString:   "curl , wget , jq",
			expectedPackages: []string{"curl", "wget", "jq"},
		},
		{
			name:             "packages with extra spaces",
			packagesString:   "  curl  ,  wget  ,  jq  ",
			expectedPackages: []string{"curl", "wget", "jq"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "execute-bash",
					Arguments: map[string]interface{}{
						"script":   `echo "test"`,
						"packages": tt.packagesString,
					},
				},
			}

			_, err := bashTool.HandleExecution(context.Background(), request)
			if err != nil {
				t.Errorf("HandleExecution() error = %v", err)
				return
			}

			if len(mockExec.lastDeps) != len(tt.expectedPackages) {
				t.Errorf("Packages count = %d, want %d", len(mockExec.lastDeps), len(tt.expectedPackages))
			}

			for i, expectedPkg := range tt.expectedPackages {
				if i >= len(mockExec.lastDeps) {
					t.Errorf("Missing package at index %d", i)
					continue
				}
				if mockExec.lastDeps[i] != expectedPkg {
					t.Errorf("Package[%d] = %q, want %q", i, mockExec.lastDeps[i], expectedPkg)
				}
			}
		})
	}
}

// Tests for SubprocessBashTool

func TestNewSubprocessBashTool(t *testing.T) {
	mockExec := &mockExecutor{}
	tool := NewSubprocessBashTool(mockExec)

	if tool == nil {
		t.Fatal("NewSubprocessBashTool() returned nil")
	}

	if tool.executor == nil {
		t.Error("NewSubprocessBashTool() executor should not be nil")
	}
}

func TestSubprocessBashTool_CreateTool(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewSubprocessBashTool(mockExec)

	tool := bashTool.CreateTool()

	if tool.Name != "execute-bash" {
		t.Errorf("Tool name = %q, want %q", tool.Name, "execute-bash")
	}

	if tool.Description == "" {
		t.Error("Tool description should not be empty")
	}

	// Verify description mentions host system
	if !strings.Contains(tool.Description, "host system") {
		t.Error("Tool description should mention 'host system'")
	}

	// Verify required parameters
	if tool.InputSchema.Properties == nil {
		t.Fatal("Tool should have input schema properties")
	}

	// Check that 'script' parameter exists
	scriptSchema, hasScript := tool.InputSchema.Properties["script"]
	if !hasScript {
		t.Error("Tool should have 'script' parameter")
	}
	if scriptSchema == nil {
		t.Error("Script parameter schema should not be nil")
	}

	// CRITICAL: Check that 'packages' parameter does NOT exist
	if _, hasPackages := tool.InputSchema.Properties["packages"]; hasPackages {
		t.Error("SubprocessBashTool should NOT have 'packages' parameter (no apt-get install allowed)")
	}

	// Check that 'env' parameter exists
	if _, hasEnv := tool.InputSchema.Properties["env"]; !hasEnv {
		t.Error("Tool should have 'env' parameter")
	}
}

func TestSubprocessBashTool_HandleExecution(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		mockOutput   string
		mockError    error
		wantErr      bool
		wantResult   string
		checkEnvVars map[string]string
	}{
		{
			name: "simple script execution",
			params: map[string]interface{}{
				"script": `echo "hello"`,
			},
			mockOutput: "hello\n",
			mockError:  nil,
			wantErr:    false,
			wantResult: "hello",
		},
		{
			name: "with environment variables",
			params: map[string]interface{}{
				"script": `echo $API_KEY`,
				"env":    "API_KEY=secret123,DEBUG=true",
			},
			mockOutput: "secret123",
			mockError:  nil,
			wantErr:    false,
			wantResult: "secret123",
			checkEnvVars: map[string]string{
				"API_KEY": "secret123",
				"DEBUG":   "true",
			},
		},
		{
			name: "multiline script",
			params: map[string]interface{}{
				"script": "#!/bin/bash\necho 'line1'\necho 'line2'",
			},
			mockOutput: "line1\nline2\n",
			mockError:  nil,
			wantErr:    false,
			wantResult: "line1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				executeFunc: func(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}

			bashTool := NewSubprocessBashTool(mockExec)
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "execute-bash",
					Arguments: tt.params,
				},
			}

			result, err := bashTool.HandleExecution(context.Background(), request)

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

			// CRITICAL: Verify dependencies are ALWAYS nil for subprocess mode
			if mockExec.lastDeps != nil {
				t.Errorf("SubprocessBashTool should always pass nil dependencies, got: %v", mockExec.lastDeps)
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

func TestSubprocessBashTool_NoDependencies(t *testing.T) {
	mockExec := &mockExecutor{}
	bashTool := NewSubprocessBashTool(mockExec)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "execute-bash",
			Arguments: map[string]interface{}{
				"script": `echo "test"`,
			},
		},
	}

	_, err := bashTool.HandleExecution(context.Background(), request)
	if err != nil {
		t.Errorf("HandleExecution() should succeed, got error: %v", err)
	}

	// CRITICAL: Verify that nil is passed for dependencies (no apt-get install)
	if mockExec.lastDeps != nil {
		t.Errorf("SubprocessBashTool must pass nil dependencies to prevent apt-get install, got: %v", mockExec.lastDeps)
	}
}
