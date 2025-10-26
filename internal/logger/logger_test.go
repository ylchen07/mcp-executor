package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSetVerbose(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "enable verbose",
			enabled: true,
		},
		{
			name:    "disable verbose",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVerbose(tt.enabled)
			if verboseEnabled != tt.enabled {
				t.Errorf("SetVerbose(%v) failed, verboseEnabled = %v", tt.enabled, verboseEnabled)
			}
		})
	}
}

func TestIsVerbose(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	tests := []struct {
		name     string
		setValue bool
		want     bool
	}{
		{
			name:     "verbose enabled returns true",
			setValue: true,
			want:     true,
		},
		{
			name:     "verbose disabled returns false",
			setValue: false,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVerbose(tt.setValue)
			got := IsVerbose()
			if got != tt.want {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerboseState(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	// Default should be false
	verboseEnabled = false
	if IsVerbose() {
		t.Error("IsVerbose() should default to false")
	}

	// Toggle to true
	SetVerbose(true)
	if !IsVerbose() {
		t.Error("IsVerbose() should be true after SetVerbose(true)")
	}

	// Toggle back to false
	SetVerbose(false)
	if IsVerbose() {
		t.Error("IsVerbose() should be false after SetVerbose(false)")
	}
}

func TestVerbosePrint(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	tests := []struct {
		name           string
		verboseEnabled bool
		format         string
		args           []interface{}
		wantOutput     bool
	}{
		{
			name:           "verbose enabled outputs message",
			verboseEnabled: true,
			format:         "test message %s",
			args:           []interface{}{"arg"},
			wantOutput:     true,
		},
		{
			name:           "verbose disabled outputs nothing",
			verboseEnabled: false,
			format:         "test message %s",
			args:           []interface{}{"arg"},
			wantOutput:     false,
		},
		{
			name:           "verbose enabled with no args",
			verboseEnabled: true,
			format:         "simple message",
			args:           []interface{}{},
			wantOutput:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			SetVerbose(tt.verboseEnabled)
			VerbosePrint(tt.format, tt.args...)

			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe writer: %v", err)
			}
			os.Stdout = old

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to copy pipe output: %v", err)
			}
			output := buf.String()

			if tt.wantOutput {
				if output == "" {
					t.Error("Expected output but got none")
				}
				// Check if the format string is in the output
				expectedContent := strings.Split(tt.format, "%")[0]
				if !strings.Contains(output, expectedContent) {
					t.Errorf("Output %q should contain %q", output, expectedContent)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no output but got: %q", output)
				}
			}
		})
	}
}

func TestVerbose(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	// These tests verify the function doesn't panic
	// We can't easily capture stderr without significant setup
	tests := []struct {
		name           string
		verboseEnabled bool
		format         string
		args           []interface{}
	}{
		{
			name:           "verbose enabled",
			verboseEnabled: true,
			format:         "debug message %d",
			args:           []interface{}{42},
		},
		{
			name:           "verbose disabled",
			verboseEnabled: false,
			format:         "debug message %d",
			args:           []interface{}{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Verbose() panicked: %v", r)
				}
			}()

			SetVerbose(tt.verboseEnabled)
			Verbose(tt.format, tt.args...)
		})
	}
}

func TestDebug(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	tests := []struct {
		name           string
		verboseEnabled bool
		format         string
		args           []interface{}
	}{
		{
			name:           "debug with verbose enabled",
			verboseEnabled: true,
			format:         "debug: %s",
			args:           []interface{}{"test"},
		},
		{
			name:           "debug with verbose disabled",
			verboseEnabled: false,
			format:         "debug: %s",
			args:           []interface{}{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Debug() panicked: %v", r)
				}
			}()

			SetVerbose(tt.verboseEnabled)
			Debug(tt.format, tt.args...)
		})
	}
}

func TestInfo(t *testing.T) {
	// Info should always output regardless of verbose setting
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	tests := []struct {
		name           string
		verboseEnabled bool
		format         string
		args           []interface{}
	}{
		{
			name:           "info with verbose enabled",
			verboseEnabled: true,
			format:         "info: %s",
			args:           []interface{}{"test"},
		},
		{
			name:           "info with verbose disabled",
			verboseEnabled: false,
			format:         "info: %s",
			args:           []interface{}{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Info() panicked: %v", r)
				}
			}()

			SetVerbose(tt.verboseEnabled)
			Info(tt.format, tt.args...)
		})
	}
}

func TestError(t *testing.T) {
	// Error should always output regardless of verbose setting
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	tests := []struct {
		name           string
		verboseEnabled bool
		format         string
		args           []interface{}
	}{
		{
			name:           "error with verbose enabled",
			verboseEnabled: true,
			format:         "error: %s",
			args:           []interface{}{"test"},
		},
		{
			name:           "error with verbose disabled",
			verboseEnabled: false,
			format:         "error: %s",
			args:           []interface{}{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Error() panicked: %v", r)
				}
			}()

			SetVerbose(tt.verboseEnabled)
			Error(tt.format, tt.args...)
		})
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Verify logger is initialized
	if logger == nil {
		t.Error("Logger should be initialized")
	}
}

func TestMultipleSetVerboseCalls(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	// Test multiple toggles
	SetVerbose(true)
	if !IsVerbose() {
		t.Error("First SetVerbose(true) failed")
	}

	SetVerbose(false)
	if IsVerbose() {
		t.Error("SetVerbose(false) failed")
	}

	SetVerbose(true)
	if !IsVerbose() {
		t.Error("Second SetVerbose(true) failed")
	}

	SetVerbose(true)
	if !IsVerbose() {
		t.Error("Calling SetVerbose(true) twice should keep it true")
	}

	SetVerbose(false)
	SetVerbose(false)
	if IsVerbose() {
		t.Error("Calling SetVerbose(false) twice should keep it false")
	}
}

func TestVerbosePrintWithFormatting(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetVerbose(true)
	VerbosePrint("Number: %d, String: %s, Bool: %v", 42, "test", true)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe writer: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy pipe output: %v", err)
	}
	output := buf.String()

	expectedParts := []string{"Number:", "42", "String:", "test", "Bool:", "true"}
	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output %q should contain %q", output, part)
		}
	}
}

func TestLogFunctionsWithComplexFormatting(t *testing.T) {
	// Save original state
	originalState := verboseEnabled
	defer func() {
		verboseEnabled = originalState
	}()

	SetVerbose(true)

	// These shouldn't panic with complex formatting
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Log functions panicked with complex formatting: %v", r)
		}
	}()

	type testStruct struct {
		Name  string
		Value int
	}

	testData := testStruct{Name: "test", Value: 123}

	Verbose("Struct: %+v, Type: %T", testData, testData)
	Debug("Debug with struct: %#v", testData)
	Info("Info with multiple args: %s %d %v", "string", 42, true)
	Error("Error with struct: %v", testData)
}

// Example demonstrates how to use the logger package
func ExampleSetVerbose() {
	// Enable verbose logging
	SetVerbose(true)
	fmt.Println(IsVerbose())

	// Disable verbose logging
	SetVerbose(false)
	fmt.Println(IsVerbose())

	// Output:
	// true
	// false
}
