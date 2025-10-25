package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewSystemCheckPrompt(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	if prompt == nil {
		t.Fatal("NewSystemCheckPrompt() returned nil")
	}
}

func TestSystemCheckPrompt_CreatePrompt(t *testing.T) {
	prompt := NewSystemCheckPrompt()
	mcpPrompt := prompt.CreatePrompt()

	// Verify prompt name
	if mcpPrompt.Name != "system-check" {
		t.Errorf("Prompt name = %q, want %q", mcpPrompt.Name, "system-check")
	}

	// Verify description exists
	if mcpPrompt.Description == "" {
		t.Error("Prompt description should not be empty")
	}

	// Verify description mentions subprocess mode
	if !strings.Contains(mcpPrompt.Description, "subprocess") {
		t.Error("Prompt description should mention 'subprocess' execution mode")
	}

	// Verify arguments are defined
	if len(mcpPrompt.Arguments) == 0 {
		t.Fatal("Prompt should have arguments defined")
	}

	// Verify detail_level argument exists
	foundDetailLevel := false
	for _, arg := range mcpPrompt.Arguments {
		if arg.Name == "detail_level" {
			foundDetailLevel = true
			if arg.Description == "" {
				t.Error("detail_level argument should have a description")
			}
			// Verify it's optional (not required)
			if arg.Required {
				t.Error("detail_level argument should be optional (not required)")
			}
		}
	}

	if !foundDetailLevel {
		t.Error("Prompt should have 'detail_level' argument")
	}
}

func TestSystemCheckPrompt_HandlePrompt_Basic(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "system-check",
			Arguments: map[string]string{
				"detail_level": "basic",
			},
		},
	}

	result, err := prompt.HandlePrompt(context.Background(), request)

	if err != nil {
		t.Fatalf("HandlePrompt() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("HandlePrompt() returned nil result")
	}

	// Verify description
	if !strings.Contains(result.Description, "basic") {
		t.Errorf("Result description should mention 'basic' level, got: %s", result.Description)
	}

	// Verify messages
	if len(result.Messages) == 0 {
		t.Fatal("Result should contain at least one message")
	}

	// Get the message content
	message := result.Messages[0]
	if message.Role != mcp.RoleAssistant {
		t.Errorf("Message role = %v, want %v", message.Role, mcp.RoleAssistant)
	}

	// Extract text content
	textContent, ok := message.Content.(mcp.TextContent)
	if !ok {
		t.Fatal("Message content should be TextContent")
	}

	messageText := textContent.Text

	// Verify message contains key information
	expectedContents := []string{
		"basic",
		"bash",
		"execute-bash",
		"Operating System",
		"CPU Information",
		"Memory Usage",
		"Disk Usage",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(messageText, expected) {
			t.Errorf("Message should contain %q, got: %s", expected, messageText)
		}
	}

	// Verify warning about subprocess mode
	if !strings.Contains(messageText, "subprocess") {
		t.Error("Message should contain warning about subprocess execution mode")
	}

	// Verify basic level should NOT contain detailed/full features
	unwantedContents := []string{
		"Network Interfaces",
		"Top 10 Processes",
		"Kernel Parameters",
		"Environment Variables",
	}

	for _, unwanted := range unwantedContents {
		if strings.Contains(messageText, unwanted) {
			t.Errorf("Basic level message should NOT contain %q", unwanted)
		}
	}
}

func TestSystemCheckPrompt_HandlePrompt_Detailed(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "system-check",
			Arguments: map[string]string{
				"detail_level": "detailed",
			},
		},
	}

	result, err := prompt.HandlePrompt(context.Background(), request)

	if err != nil {
		t.Fatalf("HandlePrompt() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("HandlePrompt() returned nil result")
	}

	// Verify description mentions detailed level
	if !strings.Contains(result.Description, "detailed") {
		t.Errorf("Result description should mention 'detailed' level, got: %s", result.Description)
	}

	// Get message text
	textContent, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatal("Message content should be TextContent")
	}
	messageText := textContent.Text

	// Verify detailed level includes basic + additional features
	expectedContents := []string{
		"detailed",
		"Operating System",
		"CPU Information",
		"Memory Usage",
		"Disk Usage",
		"System Uptime",
		"Network Interfaces",
		"Top 10 Processes",
		"Process Count",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(messageText, expected) {
			t.Errorf("Detailed level message should contain %q", expected)
		}
	}

	// Verify detailed level should NOT contain full-level features
	unwantedContents := []string{
		"All Mounted Filesystems",
		"Kernel Parameters",
		"Environment Variables",
		"Logged-in Users",
	}

	for _, unwanted := range unwantedContents {
		if strings.Contains(messageText, unwanted) {
			t.Errorf("Detailed level message should NOT contain %q", unwanted)
		}
	}
}

func TestSystemCheckPrompt_HandlePrompt_Full(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "system-check",
			Arguments: map[string]string{
				"detail_level": "full",
			},
		},
	}

	result, err := prompt.HandlePrompt(context.Background(), request)

	if err != nil {
		t.Fatalf("HandlePrompt() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("HandlePrompt() returned nil result")
	}

	// Verify description mentions full level
	if !strings.Contains(result.Description, "full") {
		t.Errorf("Result description should mention 'full' level, got: %s", result.Description)
	}

	// Get message text
	textContent, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatal("Message content should be TextContent")
	}
	messageText := textContent.Text

	// Verify full level includes everything
	expectedContents := []string{
		"full",
		"Operating System",
		"CPU Information",
		"Memory Usage",
		"Disk Usage",
		"System Uptime",
		"Network Interfaces",
		"Top 10 Processes",
		"All Mounted Filesystems",
		"Kernel Parameters",
		"Logged-in Users",
		"Environment Variables",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(messageText, expected) {
			t.Errorf("Full level message should contain %q", expected)
		}
	}
}

func TestSystemCheckPrompt_HandlePrompt_NoArguments(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	// Request without arguments should default to basic
	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      "system-check",
			Arguments: nil,
		},
	}

	result, err := prompt.HandlePrompt(context.Background(), request)

	if err != nil {
		t.Fatalf("HandlePrompt() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("HandlePrompt() returned nil result")
	}

	// Should default to basic level
	if !strings.Contains(result.Description, "basic") {
		t.Errorf("Result description should default to 'basic' level, got: %s", result.Description)
	}

	// Get message text
	textContent, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatal("Message content should be TextContent")
	}
	messageText := textContent.Text

	// Verify basic level features
	if !strings.Contains(messageText, "basic") {
		t.Error("Message should mention 'basic' detail level")
	}

	// Should NOT contain detailed/full features
	if strings.Contains(messageText, "Network Interfaces") {
		t.Error("Default (basic) level should NOT contain Network Interfaces")
	}
}

func TestSystemCheckPrompt_HandlePrompt_EmptyArgument(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "system-check",
			Arguments: map[string]string{
				"detail_level": "",
			},
		},
	}

	result, err := prompt.HandlePrompt(context.Background(), request)

	if err != nil {
		t.Fatalf("HandlePrompt() error = %v, want nil", err)
	}

	// Empty string should default to basic
	if !strings.Contains(result.Description, "basic") {
		t.Errorf("Empty detail_level should default to 'basic', got: %s", result.Description)
	}
}

func TestSystemCheckPrompt_HandlePrompt_InvalidArgument(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "system-check",
			Arguments: map[string]string{
				"detail_level": "invalid_level",
			},
		},
	}

	result, err := prompt.HandlePrompt(context.Background(), request)

	if err != nil {
		t.Fatalf("HandlePrompt() error = %v, want nil", err)
	}

	// Invalid detail level should fallback to basic
	if !strings.Contains(result.Description, "basic") {
		t.Errorf("Invalid detail_level should fallback to 'basic', got: %s", result.Description)
	}
}

func TestSystemCheckPrompt_HandlePrompt_CaseInsensitive(t *testing.T) {
	prompt := NewSystemCheckPrompt()

	testCases := []struct {
		input    string
		expected string
	}{
		{"BASIC", "basic"},
		{"Basic", "basic"},
		{"DETAILED", "detailed"},
		{"Detailed", "detailed"},
		{"FULL", "full"},
		{"Full", "full"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			request := mcp.GetPromptRequest{
				Params: mcp.GetPromptParams{
					Name: "system-check",
					Arguments: map[string]string{
						"detail_level": tc.input,
					},
				},
			}

			result, err := prompt.HandlePrompt(context.Background(), request)

			if err != nil {
				t.Fatalf("HandlePrompt() error = %v, want nil", err)
			}

			// Should normalize to lowercase expected level
			if !strings.Contains(result.Description, tc.expected) {
				t.Errorf("Case-insensitive input %q should be treated as %q, got: %s",
					tc.input, tc.expected, result.Description)
			}
		})
	}
}

func TestGenerateSystemCheckScript_Basic(t *testing.T) {
	script := generateSystemCheckScript("basic")

	// Verify script starts with shebang
	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Error("Script should start with #!/bin/bash shebang")
	}

	// Verify basic sections are present
	expectedSections := []string{
		"Operating System",
		"CPU Information",
		"Memory Usage",
		"Disk Usage",
		"System Check Complete",
	}

	for _, section := range expectedSections {
		if !strings.Contains(script, section) {
			t.Errorf("Basic script should contain section %q", section)
		}
	}

	// Verify detailed sections are NOT present
	unwantedSections := []string{
		"System Uptime",
		"Network Interfaces",
		"Top 10 Processes",
	}

	for _, section := range unwantedSections {
		if strings.Contains(script, section) {
			t.Errorf("Basic script should NOT contain section %q", section)
		}
	}
}

func TestGenerateSystemCheckScript_Detailed(t *testing.T) {
	script := generateSystemCheckScript("detailed")

	// Verify basic + detailed sections are present
	expectedSections := []string{
		"Operating System",
		"CPU Information",
		"Memory Usage",
		"Disk Usage",
		"System Uptime",
		"Network Interfaces",
		"Top 10 Processes",
		"Process Count",
	}

	for _, section := range expectedSections {
		if !strings.Contains(script, section) {
			t.Errorf("Detailed script should contain section %q", section)
		}
	}

	// Verify full-only sections are NOT present
	unwantedSections := []string{
		"All Mounted Filesystems",
		"Kernel Parameters",
		"Logged-in Users",
	}

	for _, section := range unwantedSections {
		if strings.Contains(script, section) {
			t.Errorf("Detailed script should NOT contain section %q", section)
		}
	}
}

func TestGenerateSystemCheckScript_Full(t *testing.T) {
	script := generateSystemCheckScript("full")

	// Verify all sections are present
	expectedSections := []string{
		"Operating System",
		"CPU Information",
		"Memory Usage",
		"Disk Usage",
		"System Uptime",
		"Network Interfaces",
		"Top 10 Processes",
		"All Mounted Filesystems",
		"Kernel Parameters",
		"Logged-in Users",
		"Environment Variables",
	}

	for _, section := range expectedSections {
		if !strings.Contains(script, section) {
			t.Errorf("Full script should contain section %q", section)
		}
	}

	// Verify script includes fallback commands for missing utilities
	expectedFallbacks := []string{
		"command -v",
		"&> /dev/null",
		"2>/dev/null",
	}

	for _, fallback := range expectedFallbacks {
		if !strings.Contains(script, fallback) {
			t.Errorf("Script should include fallback pattern %q", fallback)
		}
	}
}

func TestGetDetailLevelDescription(t *testing.T) {
	testCases := []struct {
		level       string
		shouldMatch []string
	}{
		{
			level: "basic",
			shouldMatch: []string{
				"OS",
				"CPU",
				"Memory",
				"disk",
			},
		},
		{
			level: "detailed",
			shouldMatch: []string{
				"basic",
				"uptime",
				"Network",
				"processes",
			},
		},
		{
			level: "full",
			shouldMatch: []string{
				"detailed",
				"filesystems",
				"Kernel",
				"Environment",
			},
		},
		{
			level:       "invalid",
			shouldMatch: nil, // Should return empty string
		},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			description := getDetailLevelDescription(tc.level)

			if tc.shouldMatch == nil {
				if description != "" {
					t.Errorf("Invalid level should return empty string, got: %s", description)
				}
				return
			}

			for _, expected := range tc.shouldMatch {
				if !strings.Contains(description, expected) {
					t.Errorf("Description for level %q should contain %q, got: %s",
						tc.level, expected, description)
				}
			}
		})
	}
}
