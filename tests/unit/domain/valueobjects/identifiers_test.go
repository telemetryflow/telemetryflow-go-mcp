package valueobjects_test

import (
	"strings"
	"testing"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewConversationID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid uuid",
			input:   "123e4567-e89b-12d3-a456-426614174000",
			wantErr: false,
		},
		{
			name:    "invalid uuid",
			input:   "invalid-uuid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vo.NewConversationID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConversationID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.input {
				t.Errorf("NewConversationID() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestGenerateConversationID(t *testing.T) {
	id := vo.GenerateConversationID()
	if id.IsEmpty() {
		t.Error("GenerateConversationID() returned empty ID")
	}
	if len(id.String()) != 36 {
		t.Errorf("GenerateConversationID() returned invalid length: %d", len(id.String()))
	}
}

func TestConversationID_Equals(t *testing.T) {
	id1, _ := vo.NewConversationID("123e4567-e89b-12d3-a456-426614174000")
	id2, _ := vo.NewConversationID("123e4567-e89b-12d3-a456-426614174000")
	id3, _ := vo.NewConversationID("223e4567-e89b-12d3-a456-426614174000")

	if !id1.Equals(id2) {
		t.Error("Expected equal IDs to match")
	}
	if id1.Equals(id3) {
		t.Error("Expected different IDs to not match")
	}
}

func TestNewToolID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid tool name",
			input:   "my_tool",
			wantErr: false,
		},
		{
			name:    "valid with hyphen",
			input:   "my-tool",
			wantErr: false,
		},
		{
			name:    "valid alphanumeric",
			input:   "tool123",
			wantErr: false,
		},
		{
			name:    "invalid starts with number",
			input:   "123tool",
			wantErr: true,
		},
		{
			name:    "invalid special characters",
			input:   "tool@name",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "too long (65 chars)",
			input:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vo.NewToolID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewToolID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.input {
				t.Errorf("NewToolID() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestNewResourceID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid file URI",
			input:   "file:///path/to/file",
			wantErr: false,
		},
		{
			name:    "valid custom scheme",
			input:   "tfo://resources/config",
			wantErr: false,
		},
		{
			name:    "invalid no scheme",
			input:   "path/to/file",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vo.NewResourceID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.input {
				t.Errorf("NewResourceID() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestSessionID(t *testing.T) {
	id := vo.GenerateSessionID()

	if id.IsEmpty() {
		t.Error("GenerateSessionID() returned empty ID")
	}

	id2, err := vo.NewSessionID(id.String())
	if err != nil {
		t.Errorf("NewSessionID() from valid string failed: %v", err)
	}

	if !id.Equals(id2) {
		t.Error("Expected IDs to be equal")
	}
}

func TestRequestID(t *testing.T) {
	id := vo.GenerateRequestID()

	if id.IsEmpty() {
		t.Error("GenerateRequestID() returned empty ID")
	}

	id2, err := vo.NewRequestID(id.String())
	if err != nil {
		t.Errorf("NewRequestID() from valid string failed: %v", err)
	}

	if !id.Equals(id2) {
		t.Error("Expected IDs to be equal")
	}
}

func TestMessageID(t *testing.T) {
	t.Run("generate and parse", func(t *testing.T) {
		id := vo.GenerateMessageID()
		if id.IsEmpty() {
			t.Error("GenerateMessageID() returned empty ID")
		}
		parsed, err := vo.NewMessageID(id.String())
		if err != nil {
			t.Fatalf("NewMessageID() failed: %v", err)
		}
		if !id.Equals(parsed) {
			t.Error("Expected IDs to be equal")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := vo.NewMessageID("not-a-uuid")
		if err == nil {
			t.Error("Expected error for invalid message ID")
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := vo.NewMessageID("")
		if err == nil {
			t.Error("Expected error for empty message ID")
		}
	})
}

func TestPromptID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		id, err := vo.NewPromptID("my_prompt")
		if err != nil {
			t.Fatalf("NewPromptID() failed: %v", err)
		}
		if id.String() != "my_prompt" {
			t.Errorf("PromptID.String() = %v, want my_prompt", id.String())
		}
		if id.IsEmpty() {
			t.Error("PromptID.IsEmpty() should be false")
		}
	})

	t.Run("equals", func(t *testing.T) {
		id1, _ := vo.NewPromptID("test")
		id2, _ := vo.NewPromptID("test")
		id3, _ := vo.NewPromptID("other")
		if !id1.Equals(id2) {
			t.Error("Expected equal")
		}
		if id1.Equals(id3) {
			t.Error("Expected not equal")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := vo.NewPromptID("")
		if err == nil {
			t.Error("Expected error for empty prompt ID")
		}
		_, err = vo.NewPromptID("1invalid")
		if err == nil {
			t.Error("Expected error for prompt ID starting with number")
		}
	})
}

func TestToolID_IsEmpty_Equals(t *testing.T) {
	id1, _ := vo.NewToolID("my_tool")
	id2, _ := vo.NewToolID("my_tool")
	id3, _ := vo.NewToolID("other_tool")

	if id1.IsEmpty() {
		t.Error("ToolID.IsEmpty() should be false")
	}
	if !id1.Equals(id2) {
		t.Error("Expected equal")
	}
	if id1.Equals(id3) {
		t.Error("Expected not equal")
	}
}

func TestResourceID_IsEmpty_Equals(t *testing.T) {
	id1, _ := vo.NewResourceID("file:///path")
	id2, _ := vo.NewResourceID("file:///path")
	id3, _ := vo.NewResourceID("tfo://other")

	if id1.IsEmpty() {
		t.Error("ResourceID.IsEmpty() should be false")
	}
	if !id1.Equals(id2) {
		t.Error("Expected equal")
	}
	if id1.Equals(id3) {
		t.Error("Expected not equal")
	}
}

func TestNewRequestID_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty string", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
		{name: "valid id", input: "req-123", wantErr: false},
		{name: "trimmed whitespace", input: "  req-456  ", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := vo.NewRequestID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequestID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				trimmed := strings.TrimSpace(tt.input)
				if id.String() != trimmed {
					t.Errorf("expected %s, got %s", trimmed, id.String())
				}
			}
		})
	}
}

func TestRequestID_IsEmpty_Equals(t *testing.T) {
	id1, _ := vo.NewRequestID("req-1")
	id2, _ := vo.NewRequestID("req-1")
	id3, _ := vo.NewRequestID("req-2")

	emptyID := vo.RequestID{}
	if !emptyID.IsEmpty() {
		t.Error("expected empty RequestID to be empty")
	}
	if id1.IsEmpty() {
		t.Error("expected non-empty RequestID")
	}
	if !id1.Equals(id2) {
		t.Error("expected equal")
	}
	if id1.Equals(id3) {
		t.Error("expected not equal")
	}
}

func TestConversationID_IsEmpty(t *testing.T) {
	emptyID := vo.ConversationID{}
	if !emptyID.IsEmpty() {
		t.Error("expected empty ConversationID")
	}
}

func TestSessionID_InvalidUUID(t *testing.T) {
	_, err := vo.NewSessionID("not-a-uuid")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestSessionID_EmptyString(t *testing.T) {
	_, err := vo.NewSessionID("")
	if err == nil {
		t.Error("expected error for empty session ID")
	}
}

func TestSessionID_IsEmpty(t *testing.T) {
	emptyID := vo.SessionID{}
	if !emptyID.IsEmpty() {
		t.Error("expected empty SessionID")
	}
}

func TestSessionID_Equals(t *testing.T) {
	id1 := vo.GenerateSessionID()
	id2, _ := vo.NewSessionID(id1.String())
	id3 := vo.GenerateSessionID()

	if !id1.Equals(id2) {
		t.Error("expected equal")
	}
	if id1.Equals(id3) {
		t.Error("expected not equal")
	}
}

func TestMessageID_IsEmpty(t *testing.T) {
	emptyID := vo.MessageID{}
	if !emptyID.IsEmpty() {
		t.Error("expected empty MessageID")
	}
}

func TestMessageID_Equals(t *testing.T) {
	id1 := vo.GenerateMessageID()
	id2, _ := vo.NewMessageID(id1.String())
	id3 := vo.GenerateMessageID()

	if !id1.Equals(id2) {
		t.Error("expected equal")
	}
	if id1.Equals(id3) {
		t.Error("expected not equal")
	}
}

func TestToolID_EmptyString(t *testing.T) {
	_, err := vo.NewToolID("")
	if err == nil {
		t.Error("expected error for empty tool ID")
	}
}

func TestConversationID_TrimmedWhitespace(t *testing.T) {
	id, err := vo.NewConversationID("  123e4567-e89b-12d3-a456-426614174000  ")
	if err != nil {
		t.Fatalf("NewConversationID: %v", err)
	}
	if id.String() != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("expected trimmed UUID, got %s", id.String())
	}
}

func TestSessionID_TrimmedWhitespace(t *testing.T) {
	uuid := vo.GenerateSessionID().String()
	id, err := vo.NewSessionID("  " + uuid + "  ")
	if err != nil {
		t.Fatalf("NewSessionID: %v", err)
	}
	if id.String() != uuid {
		t.Errorf("expected trimmed UUID, got %s", id.String())
	}
}

func TestMessageID_TrimmedWhitespace(t *testing.T) {
	uuid := vo.GenerateMessageID().String()
	id, err := vo.NewMessageID("  " + uuid + "  ")
	if err != nil {
		t.Fatalf("NewMessageID: %v", err)
	}
	if id.String() != uuid {
		t.Errorf("expected trimmed UUID, got %s", id.String())
	}
}
