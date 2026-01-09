// Package claude provides Claude API client utilities
package claude

import (
	"context"
	"fmt"
	"time"
)

// Model constants
const (
	ModelOpus4         = "claude-opus-4-20250514"
	ModelSonnet4       = "claude-sonnet-4-20250514"
	ModelSonnet35      = "claude-3-5-sonnet-20241022"
	ModelHaiku35       = "claude-3-5-haiku-20241022"
	DefaultModel       = ModelSonnet4
	DefaultMaxTokens   = 4096
	DefaultTemperature = 0.7
)

// Role constants
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ContentType constants
const (
	ContentTypeText  = "text"
	ContentTypeImage = "image"
)

// StopReason constants
const (
	StopReasonEndTurn      = "end_turn"
	StopReasonMaxTokens    = "max_tokens"
	StopReasonToolUse      = "tool_use"
	StopReasonStopSequence = "stop_sequence"
)

// Config holds Claude API configuration
type Config struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	TopK        int
	Timeout     time.Duration
	RetryConfig *RetryConfig
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultConfig returns default configuration
func DefaultConfig(apiKey string) *Config {
	return &Config{
		APIKey:      apiKey,
		BaseURL:     "https://api.anthropic.com",
		Model:       DefaultModel,
		MaxTokens:   DefaultMaxTokens,
		Temperature: DefaultTemperature,
		TopP:        0.9,
		TopK:        40,
		Timeout:     30 * time.Second,
		RetryConfig: &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
	}
}

// Message represents a conversation message
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block in a message
type ContentBlock struct {
	Type      string       `json:"type"`
	Text      string       `json:"text,omitempty"`
	Source    *ImageSource `json:"source,omitempty"`
	ID        string       `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Input     interface{}  `json:"input,omitempty"`
	ToolUseID string       `json:"tool_use_id,omitempty"`
	Content   string       `json:"content,omitempty"`
}

// ImageSource represents an image source
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// Tool represents a tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"input_schema"`
}

// CreateMessageRequest represents a message creation request
type CreateMessageRequest struct {
	Model         string      `json:"model"`
	Messages      []Message   `json:"messages"`
	MaxTokens     int         `json:"max_tokens"`
	System        string      `json:"system,omitempty"`
	Temperature   *float64    `json:"temperature,omitempty"`
	TopP          *float64    `json:"top_p,omitempty"`
	TopK          *int        `json:"top_k,omitempty"`
	StopSequences []string    `json:"stop_sequences,omitempty"`
	Stream        bool        `json:"stream,omitempty"`
	Tools         []Tool      `json:"tools,omitempty"`
	ToolChoice    *ToolChoice `json:"tool_choice,omitempty"`
	Metadata      *Metadata   `json:"metadata,omitempty"`
}

// ToolChoice represents tool choice configuration
type ToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Metadata represents request metadata
type Metadata struct {
	UserID string `json:"user_id,omitempty"`
}

// CreateMessageResponse represents a message creation response
type CreateMessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// Usage represents token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Client interface for Claude API operations
type Client interface {
	// CreateMessage creates a new message
	CreateMessage(ctx context.Context, req *CreateMessageRequest) (*CreateMessageResponse, error)
	// CreateMessageStream creates a streaming message
	CreateMessageStream(ctx context.Context, req *CreateMessageRequest) (<-chan StreamEvent, error)
	// CountTokens counts tokens for messages
	CountTokens(ctx context.Context, messages []Message, system string) (int, error)
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type    string
	Message *CreateMessageResponse
	Delta   *ContentBlock
	Error   error
}

// NewTextMessage creates a text message
func NewTextMessage(role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentBlock{
			{
				Type: ContentTypeText,
				Text: text,
			},
		},
	}
}

// NewUserMessage creates a user message
func NewUserMessage(text string) Message {
	return NewTextMessage(RoleUser, text)
}

// NewAssistantMessage creates an assistant message
func NewAssistantMessage(text string) Message {
	return NewTextMessage(RoleAssistant, text)
}

// NewToolResultMessage creates a tool result message
func NewToolResultMessage(toolUseID, content string) Message {
	return Message{
		Role: RoleUser,
		Content: []ContentBlock{
			{
				Type:      "tool_result",
				ToolUseID: toolUseID,
				Content:   content,
			},
		},
	}
}

// ExtractText extracts text from content blocks
func ExtractText(content []ContentBlock) string {
	var result string
	for _, block := range content {
		if block.Type == ContentTypeText {
			result += block.Text
		}
	}
	return result
}

// HasToolUse checks if response contains tool use
func HasToolUse(content []ContentBlock) bool {
	for _, block := range content {
		if block.Type == "tool_use" {
			return true
		}
	}
	return false
}

// GetToolUseBlocks returns tool use blocks from content
func GetToolUseBlocks(content []ContentBlock) []ContentBlock {
	var toolUses []ContentBlock
	for _, block := range content {
		if block.Type == "tool_use" {
			toolUses = append(toolUses, block)
		}
	}
	return toolUses
}

// ValidateModel validates if model is supported
func ValidateModel(model string) error {
	switch model {
	case ModelOpus4, ModelSonnet4, ModelSonnet35, ModelHaiku35:
		return nil
	default:
		return fmt.Errorf("unsupported model: %s", model)
	}
}

// GetModelInfo returns information about a model
func GetModelInfo(model string) (maxTokens int, contextWindow int) {
	switch model {
	case ModelOpus4:
		return 32768, 200000
	case ModelSonnet4:
		return 16384, 200000
	case ModelSonnet35:
		return 8192, 200000
	case ModelHaiku35:
		return 8192, 200000
	default:
		return DefaultMaxTokens, 100000
	}
}
