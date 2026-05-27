// Package valueobjects_test contains unit tests for domain value objects
package valueobjects_test

import (
	"strings"
	"testing"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestContentType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		contentType vo.ContentType
		want        bool
	}{
		{"text is valid", vo.ContentTypeText, true},
		{"image is valid", vo.ContentTypeImage, true},
		{"tool_use is valid", vo.ContentTypeToolUse, true},
		{"tool_result is valid", vo.ContentTypeToolResult, true},
		{"invalid type", vo.ContentType("invalid"), false},
		{"empty type", vo.ContentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.contentType.IsValid(); got != tt.want {
				t.Errorf("ContentType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentType_String(t *testing.T) {
	tests := []struct {
		contentType vo.ContentType
		want        string
	}{
		{vo.ContentTypeText, "text"},
		{vo.ContentTypeImage, "image"},
		{vo.ContentTypeToolUse, "tool_use"},
		{vo.ContentTypeToolResult, "tool_result"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.contentType.String(); got != tt.want {
				t.Errorf("ContentType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role vo.Role
		want bool
	}{
		{"user is valid", vo.RoleUser, true},
		{"assistant is valid", vo.RoleAssistant, true},
		{"system is valid", vo.RoleSystem, true},
		{"invalid role", vo.Role("invalid"), false},
		{"empty role", vo.Role(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("Role.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_String(t *testing.T) {
	tests := []struct {
		role vo.Role
		want string
	}{
		{vo.RoleUser, "user"},
		{vo.RoleAssistant, "assistant"},
		{vo.RoleSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		model vo.Model
		want  bool
	}{
		{"claude-opus-4-7 is valid", vo.ModelClaudeOpus47, true},
		{"claude-sonnet-4 is valid", vo.ModelClaudeSonnet4, true},
		{"claude-sonnet-4-6 is valid", vo.ModelClaudeSonnet46, true},
		{"claude-sonnet-4-5 is valid", vo.ModelClaudeSonnet45, true},
		{"claude-sonnet-4-5 (v2) is valid", vo.ModelClaudeSonnet45, true},
		{"claude-haiku-4-5 is valid", vo.ModelClaudeHaiku45, true},
		{"claude-opus-4-5 is valid", vo.ModelClaudeOpus45, true},
		{"claude-sonnet-4 (v3) is valid", vo.ModelClaudeSonnet4, true},
		{"claude-haiku-4-5 (v3) is valid", vo.ModelClaudeHaiku45, true},
		{"invalid model", vo.Model("invalid-model"), false},
		{"empty model", vo.Model(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.IsValid(); got != tt.want {
				t.Errorf("Model.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	model := vo.ModelClaudeSonnet4
	if got := model.String(); got != "claude-sonnet-4-20250514" {
		t.Errorf("Model.String() = %v, want claude-sonnet-4-20250514", got)
	}
}

func TestNewTextContent(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"valid text", "Hello, World!", nil},
		{"valid unicode", "こんにちは世界", nil},
		{"empty string", "", vo.ErrEmptyContent},
		{"whitespace only", "   ", vo.ErrEmptyContent},
		{"text with whitespace", "  Hello  ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := vo.NewTextContent(tt.value)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewTextContent() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewTextContent() unexpected error = %v", err)
				return
			}
			if content.String() != tt.value {
				t.Errorf("TextContent.String() = %v, want %v", content.String(), tt.value)
			}
		})
	}
}

func TestTextContent_TooLong(t *testing.T) {
	longText := strings.Repeat("a", vo.MaxTextContentLength+1)
	_, err := vo.NewTextContent(longText)
	if err != vo.ErrContentTooLong {
		t.Errorf("NewTextContent() error = %v, want ErrContentTooLong", err)
	}
}

func TestTextContent_Methods(t *testing.T) {
	content, err := vo.NewTextContent("Hello, World!")
	if err != nil {
		t.Fatalf("NewTextContent() error = %v", err)
	}

	// Test IsEmpty
	if content.IsEmpty() {
		t.Error("TextContent.IsEmpty() should return false for non-empty content")
	}

	// Test Length
	if content.Length() != 13 {
		t.Errorf("TextContent.Length() = %v, want 13", content.Length())
	}

	// Test Truncate
	truncated := content.Truncate(5)
	if truncated != "Hello..." {
		t.Errorf("TextContent.Truncate(5) = %v, want Hello...", truncated)
	}

	// Test Truncate with larger limit
	notTruncated := content.Truncate(100)
	if notTruncated != "Hello, World!" {
		t.Errorf("TextContent.Truncate(100) = %v, want Hello, World!", notTruncated)
	}
}

func TestNewSystemPrompt(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"valid prompt", "You are a helpful assistant", nil},
		{"empty prompt", "", nil}, // Empty is valid for system prompt
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := vo.NewSystemPrompt(tt.value)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewSystemPrompt() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewSystemPrompt() unexpected error = %v", err)
				return
			}
			if prompt.String() != tt.value {
				t.Errorf("SystemPrompt.String() = %v, want %v", prompt.String(), tt.value)
			}
		})
	}
}

func TestSystemPrompt_TooLong(t *testing.T) {
	longPrompt := strings.Repeat("a", vo.MaxSystemPromptLength+1)
	_, err := vo.NewSystemPrompt(longPrompt)
	if err != vo.ErrContentTooLong {
		t.Errorf("NewSystemPrompt() error = %v, want ErrContentTooLong", err)
	}
}

func TestSystemPrompt_IsEmpty(t *testing.T) {
	emptyPrompt, _ := vo.NewSystemPrompt("")
	if !emptyPrompt.IsEmpty() {
		t.Error("SystemPrompt.IsEmpty() should return true for empty prompt")
	}

	nonEmptyPrompt, _ := vo.NewSystemPrompt("Hello")
	if nonEmptyPrompt.IsEmpty() {
		t.Error("SystemPrompt.IsEmpty() should return false for non-empty prompt")
	}
}

func TestNewMimeType(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"json", "application/json", "application/json"},
		{"plain text", "text/plain", "text/plain"},
		{"empty defaults to plain text", "", vo.MimeTypePlainText},
		{"whitespace defaults to plain text", "   ", vo.MimeTypePlainText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, err := vo.NewMimeType(tt.value)
			if err != nil {
				t.Errorf("NewMimeType() error = %v", err)
				return
			}
			if mimeType.String() != tt.expected {
				t.Errorf("MimeType.String() = %v, want %v", mimeType.String(), tt.expected)
			}
		})
	}
}

func TestMimeType_IsText(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{vo.MimeTypePlainText, true},
		{vo.MimeTypeMarkdown, true},
		{vo.MimeTypeHTML, true},
		{vo.MimeTypeJSON, true},
		{vo.MimeTypeXML, true},
		{vo.MimeTypePNG, false},
		{vo.MimeTypeJPEG, false},
		{vo.MimeTypePDF, false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			mimeType, _ := vo.NewMimeType(tt.mimeType)
			if got := mimeType.IsText(); got != tt.want {
				t.Errorf("MimeType.IsText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMimeType_IsImage(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{vo.MimeTypePNG, true},
		{vo.MimeTypeJPEG, true},
		{vo.MimeTypeGIF, true},
		{vo.MimeTypeWebP, true},
		{vo.MimeTypePlainText, false},
		{vo.MimeTypeJSON, false},
		{vo.MimeTypePDF, false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			mimeType, _ := vo.NewMimeType(tt.mimeType)
			if got := mimeType.IsImage(); got != tt.want {
				t.Errorf("MimeType.IsImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		name, err := vo.NewToolName("my_tool")
		if err != nil {
			t.Fatalf("NewToolName() failed: %v", err)
		}
		if name.String() != "my_tool" {
			t.Errorf("ToolName.String() = %v, want my_tool", name.String())
		}
		if name.IsEmpty() {
			t.Error("ToolName.IsEmpty() should be false")
		}
	})

	t.Run("equals", func(t *testing.T) {
		n1, _ := vo.NewToolName("tool1")
		n2, _ := vo.NewToolName("tool1")
		n3, _ := vo.NewToolName("tool2")
		if !n1.Equals(n2) {
			t.Error("Expected equal")
		}
		if n1.Equals(n3) {
			t.Error("Expected not equal")
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := vo.NewToolName("")
		if err == nil {
			t.Error("Expected error for empty tool name")
		}
	})

	t.Run("invalid_pattern", func(t *testing.T) {
		_, err := vo.NewToolName("123invalid")
		if err == nil {
			t.Error("Expected error for invalid tool name pattern")
		}
	})
}

func TestToolDescription(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		desc, err := vo.NewToolDescription("A test tool")
		if err != nil {
			t.Fatalf("NewToolDescription() failed: %v", err)
		}
		if desc.String() != "A test tool" {
			t.Errorf("ToolDescription.String() = %v, want 'A test tool'", desc.String())
		}
		if desc.IsEmpty() {
			t.Error("ToolDescription.IsEmpty() should be false")
		}
	})

	t.Run("empty description", func(t *testing.T) {
		desc, err := vo.NewToolDescription("")
		if err != nil {
			t.Fatalf("NewToolDescription('') failed: %v", err)
		}
		if !desc.IsEmpty() {
			t.Error("ToolDescription.IsEmpty() should be true")
		}
	})

	t.Run("too long", func(t *testing.T) {
		longDesc := strings.Repeat("a", vo.MaxToolDescriptionLength+1)
		_, err := vo.NewToolDescription(longDesc)
		if err != vo.ErrContentTooLong {
			t.Errorf("Expected ErrContentTooLong, got %v", err)
		}
	})
}

func TestResourceURI(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		uri, err := vo.NewResourceURI("file:///path/to/resource")
		if err != nil {
			t.Fatalf("NewResourceURI() failed: %v", err)
		}
		if uri.String() != "file:///path/to/resource" {
			t.Errorf("ResourceURI.String() = %v", uri.String())
		}
		if uri.IsEmpty() {
			t.Error("ResourceURI.IsEmpty() should be false")
		}
	})

	t.Run("equals", func(t *testing.T) {
		u1, _ := vo.NewResourceURI("file:///a")
		u2, _ := vo.NewResourceURI("file:///a")
		u3, _ := vo.NewResourceURI("file:///b")
		if !u1.Equals(u2) {
			t.Error("Expected equal")
		}
		if u1.Equals(u3) {
			t.Error("Expected not equal")
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := vo.NewResourceURI("")
		if err == nil {
			t.Error("Expected error for empty URI")
		}
	})

	t.Run("invalid_pattern", func(t *testing.T) {
		_, err := vo.NewResourceURI("no-scheme")
		if err == nil {
			t.Error("Expected error for URI without scheme")
		}
	})
}

func TestDefaultModel(t *testing.T) {
	if vo.DefaultModel != vo.ModelClaudeOpus47 {
		t.Errorf("DefaultModel = %v, want %v", vo.DefaultModel, vo.ModelClaudeOpus47)
	}
}
