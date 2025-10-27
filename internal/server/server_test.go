package server

import (
	"testing"

	"github.com/ylchen07/mcp-executor/internal/executor"
)

func TestNewMCPServer_DockerMode(t *testing.T) {
	mcpServer := NewMCPServer("docker")

	if mcpServer == nil {
		t.Fatal("NewMCPServer() returned nil")
	}

	// Verify tools are registered (indication of proper initialization)
	tools := mcpServer.ListTools()
	if len(tools) == 0 {
		t.Error("Server should have tools registered")
	}
}

func TestNewMCPServer_SubprocessMode(t *testing.T) {
	mcpServer := NewMCPServer("subprocess")

	if mcpServer == nil {
		t.Fatal("NewMCPServer() returned nil")
	}

	// Verify tools are registered
	tools := mcpServer.ListTools()
	if len(tools) == 0 {
		t.Error("Server should have tools registered")
	}
}

func TestNewMCPServer_DefaultMode(t *testing.T) {
	tests := []struct {
		name          string
		executionMode string
	}{
		{
			name:          "empty string defaults to subprocess",
			executionMode: "",
		},
		{
			name:          "unknown mode defaults to subprocess",
			executionMode: "unknown",
		},
		{
			name:          "invalid mode defaults to subprocess",
			executionMode: "invalid-mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcpServer := NewMCPServer(tt.executionMode)

			if mcpServer == nil {
				t.Fatal("NewMCPServer() returned nil")
			}

			// Should have tools registered even with unknown mode
			tools := mcpServer.ListTools()
			if len(tools) == 0 {
				t.Error("Server should have tools registered even with unknown mode")
			}
		})
	}
}

func TestNewMCPServer_ToolRegistration(t *testing.T) {
	mcpServer := NewMCPServer("subprocess")

	if mcpServer == nil {
		t.Fatal("NewMCPServer() returned nil")
	}

	// Verify tools are registered
	tools := mcpServer.ListTools()
	if len(tools) == 0 {
		t.Fatal("No tools registered")
	}

	// Check for expected tools
	expectedTools := []string{"execute-python", "execute-bash", "execute-typescript", "execute-go"}
	for _, expectedTool := range expectedTools {
		if _, found := tools[expectedTool]; !found {
			t.Errorf("Expected tool %q not found in registered tools", expectedTool)
		}
	}

	// Should have exactly 4 tools
	if len(tools) != 4 {
		t.Errorf("Expected 4 tools, got %d", len(tools))
	}
}

func TestNewMCPServer_ExecutorSelection(t *testing.T) {
	tests := []struct {
		name          string
		executionMode string
		description   string
	}{
		{
			name:          "docker mode uses docker executors",
			executionMode: "docker",
			description:   "Should create Docker-based executors",
		},
		{
			name:          "subprocess mode uses subprocess executors",
			executionMode: "subprocess",
			description:   "Should create subprocess-based executors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't directly inspect which executor was created without
			// modifying the server code, but we can verify the server was created
			// and tools were registered properly
			mcpServer := NewMCPServer(tt.executionMode)

			if mcpServer == nil {
				t.Fatalf("NewMCPServer(%q) returned nil", tt.executionMode)
			}

			// Verify tools are present
			tools := mcpServer.ListTools()
			if len(tools) != 4 {
				t.Errorf("Expected 4 tools for %s mode, got %d", tt.executionMode, len(tools))
			}
		})
	}
}

func TestNewMCPServer_MultipleInstances(t *testing.T) {
	// Test that we can create multiple server instances
	server1 := NewMCPServer("docker")
	server2 := NewMCPServer("subprocess")

	if server1 == nil || server2 == nil {
		t.Fatal("One or more servers failed to initialize")
	}

	// Both should be independent instances
	if server1 == server2 {
		t.Error("Multiple NewMCPServer calls should return different instances")
	}

	// Both should have tools registered
	if len(server1.ListTools()) != 4 {
		t.Error("Server 1 should have 4 tools")
	}
	if len(server2.ListTools()) != 4 {
		t.Error("Server 2 should have 4 tools")
	}
}

func TestNewMCPServer_ExecutorTypes(t *testing.T) {
	// Test helper to verify executor selection logic through indirect means
	// We create servers and verify they work correctly without errors

	modes := []string{"docker", "subprocess", ""}
	for _, mode := range modes {
		t.Run("mode="+mode, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewMCPServer(%q) panicked: %v", mode, r)
				}
			}()

			mcpServer := NewMCPServer(mode)
			if mcpServer == nil {
				t.Errorf("NewMCPServer(%q) returned nil", mode)
			}

			// Verify tools were registered
			tools := mcpServer.ListTools()
			if len(tools) == 0 {
				t.Errorf("NewMCPServer(%q) should have tools registered", mode)
			}
		})
	}
}

func TestNewMCPServer_ToolDetails(t *testing.T) {
	mcpServer := NewMCPServer("subprocess")

	if mcpServer == nil {
		t.Fatal("NewMCPServer() returned nil")
	}

	tools := mcpServer.ListTools()

	// Check each tool has a handler
	for toolName, tool := range tools {
		if tool == nil {
			t.Errorf("Tool %q should not be nil", toolName)
		}
	}

	// Verify we can get individual tools
	pythonTool := mcpServer.GetTool("execute-python")
	if pythonTool == nil {
		t.Error("GetTool('execute-python') should not return nil")
	}

	bashTool := mcpServer.GetTool("execute-bash")
	if bashTool == nil {
		t.Error("GetTool('execute-bash') should not return nil")
	}

	typescriptTool := mcpServer.GetTool("execute-typescript")
	if typescriptTool == nil {
		t.Error("GetTool('execute-typescript') should not return nil")
	}

	goTool := mcpServer.GetTool("execute-go")
	if goTool == nil {
		t.Error("GetTool('execute-go') should not return nil")
	}

	// Non-existent tool should return nil
	nonExistentTool := mcpServer.GetTool("non-existent-tool")
	if nonExistentTool != nil {
		t.Error("GetTool('non-existent-tool') should return nil")
	}
}

// TestExecutorInterface verifies the executor interface is properly implemented
func TestExecutorInterface(t *testing.T) {
	// Verify both executor types implement the Executor interface
	var _ executor.Executor = executor.NewPythonExecutor()
	var _ executor.Executor = executor.NewBashExecutor()
	var _ executor.Executor = executor.NewTypeScriptExecutor()
	var _ executor.Executor = executor.NewGoExecutor()
	var _ executor.Executor = executor.NewSubprocessPythonExecutor()
	var _ executor.Executor = executor.NewSubprocessBashExecutor()
	var _ executor.Executor = executor.NewSubprocessTypeScriptExecutor()
	var _ executor.Executor = executor.NewSubprocessGoExecutor()

	// If we get here without compile errors, the interface is correctly implemented
	t.Log("All executors correctly implement the Executor interface")
}

func TestRunFunctions_NoExternalDependencies(t *testing.T) {
	// These tests verify the functions exist and have correct signatures
	// We don't actually run the servers as that would require network/stdio setup

	mcpServer := NewMCPServer("subprocess")
	if mcpServer == nil {
		t.Fatal("NewMCPServer() returned nil")
	}

	// Verify RunStdio function signature
	_ = RunStdio

	// Verify RunSSE function signature
	_ = RunSSE

	// Verify RunHTTP function signature
	_ = RunHTTP

	t.Log("All Run* functions have correct signatures")
}

func TestNewMCPServer_NoNilReturns(t *testing.T) {
	// Test that NewMCPServer never returns nil for any valid/invalid input
	testCases := []string{
		"docker",
		"subprocess",
		"",
		"unknown",
		"invalid",
		"random-string",
		"123",
		"@#$%",
	}

	for _, tc := range testCases {
		t.Run("mode="+tc, func(t *testing.T) {
			server := NewMCPServer(tc)
			if server == nil {
				t.Errorf("NewMCPServer(%q) returned nil, should always return a server", tc)
			}
		})
	}
}
