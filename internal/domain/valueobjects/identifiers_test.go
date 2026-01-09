package valueobjects

import (
	"testing"
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
			got, err := NewConversationID(tt.input)
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
	id := GenerateConversationID()
	if id.IsEmpty() {
		t.Error("GenerateConversationID() returned empty ID")
	}
	if len(id.String()) != 36 { // UUID format: 8-4-4-4-12
		t.Errorf("GenerateConversationID() returned invalid length: %d", len(id.String()))
	}
}

func TestConversationID_Equals(t *testing.T) {
	id1, _ := NewConversationID("123e4567-e89b-12d3-a456-426614174000")
	id2, _ := NewConversationID("123e4567-e89b-12d3-a456-426614174000")
	id3, _ := NewConversationID("223e4567-e89b-12d3-a456-426614174000")

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
			got, err := NewToolID(tt.input)
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
			got, err := NewResourceID(tt.input)
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
	// Generate a new session ID
	id := GenerateSessionID()

	if id.IsEmpty() {
		t.Error("GenerateSessionID() returned empty ID")
	}

	// Create from string
	id2, err := NewSessionID(id.String())
	if err != nil {
		t.Errorf("NewSessionID() from valid string failed: %v", err)
	}

	if !id.Equals(id2) {
		t.Error("Expected IDs to be equal")
	}
}

func TestRequestID(t *testing.T) {
	// Generate a new request ID
	id := GenerateRequestID()

	if id.IsEmpty() {
		t.Error("GenerateRequestID() returned empty ID")
	}

	// Create from string
	id2, err := NewRequestID(id.String())
	if err != nil {
		t.Errorf("NewRequestID() from valid string failed: %v", err)
	}

	if !id.Equals(id2) {
		t.Error("Expected IDs to be equal")
	}
}
