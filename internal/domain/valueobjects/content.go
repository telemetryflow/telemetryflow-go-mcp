// Package valueobjects contains immutable, self-validating value objects
package valueobjects

import (
	"errors"
	"strings"
)

// Content validation errors
var (
	ErrEmptyContent       = errors.New("content cannot be empty")
	ErrContentTooLong     = errors.New("content exceeds maximum length")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrInvalidRole        = errors.New("invalid message role")
	ErrInvalidModel       = errors.New("invalid model identifier")
)

// ContentType represents the type of content in a message
type ContentType string

const (
	ContentTypeText       ContentType = "text"
	ContentTypeImage      ContentType = "image"
	ContentTypeToolUse    ContentType = "tool_use"
	ContentTypeToolResult ContentType = "tool_result"
)

// IsValid checks if the content type is valid
func (c ContentType) IsValid() bool {
	switch c {
	case ContentTypeText, ContentTypeImage, ContentTypeToolUse, ContentTypeToolResult:
		return true
	}
	return false
}

// String returns the string representation
func (c ContentType) String() string {
	return string(c)
}

// Role represents the role of a message participant
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	}
	return false
}

// String returns the string representation
func (r Role) String() string {
	return string(r)
}

// Model represents a Claude model identifier
type Model string

const (
	ModelClaude4Opus      Model = "claude-opus-4-20250514"
	ModelClaude4Sonnet    Model = "claude-sonnet-4-20250514"
	ModelClaude37Sonnet   Model = "claude-3-7-sonnet-20250219"
	ModelClaude35Sonnet   Model = "claude-3-5-sonnet-20241022"
	ModelClaude35SonnetV2 Model = "claude-3-5-sonnet-v2-20241022"
	ModelClaude35Haiku    Model = "claude-3-5-haiku-20241022"
	ModelClaude3Opus      Model = "claude-3-opus-20240229"
	ModelClaude3Sonnet    Model = "claude-3-sonnet-20240229"
	ModelClaude3Haiku     Model = "claude-3-haiku-20240307"
)

// DefaultModel is the default model to use
const DefaultModel = ModelClaude4Sonnet

// IsValid checks if the model is valid
func (m Model) IsValid() bool {
	switch m {
	case ModelClaude4Opus, ModelClaude4Sonnet, ModelClaude37Sonnet,
		ModelClaude35Sonnet, ModelClaude35SonnetV2, ModelClaude35Haiku,
		ModelClaude3Opus, ModelClaude3Sonnet, ModelClaude3Haiku:
		return true
	}
	return false
}

// String returns the string representation
func (m Model) String() string {
	return string(m)
}

// TextContent represents text content value object
type TextContent struct {
	value string
}

// MaxTextContentLength is the maximum allowed length for text content
const MaxTextContentLength = 1000000 // 1MB of text

// NewTextContent creates a new TextContent with validation
func NewTextContent(value string) (TextContent, error) {
	if strings.TrimSpace(value) == "" {
		return TextContent{}, ErrEmptyContent
	}
	if len(value) > MaxTextContentLength {
		return TextContent{}, ErrContentTooLong
	}
	return TextContent{value: value}, nil
}

// String returns the string representation
func (t TextContent) String() string {
	return t.value
}

// IsEmpty checks if the content is empty
func (t TextContent) IsEmpty() bool {
	return t.value == ""
}

// Length returns the length of the content
func (t TextContent) Length() int {
	return len(t.value)
}

// Truncate returns a truncated version of the content
func (t TextContent) Truncate(maxLen int) string {
	if len(t.value) <= maxLen {
		return t.value
	}
	return t.value[:maxLen] + "..."
}

// SystemPrompt represents a system prompt value object
type SystemPrompt struct {
	value string
}

// MaxSystemPromptLength is the maximum allowed length for system prompts
const MaxSystemPromptLength = 100000 // 100KB

// NewSystemPrompt creates a new SystemPrompt with validation
func NewSystemPrompt(value string) (SystemPrompt, error) {
	if len(value) > MaxSystemPromptLength {
		return SystemPrompt{}, ErrContentTooLong
	}
	return SystemPrompt{value: value}, nil
}

// String returns the string representation
func (s SystemPrompt) String() string {
	return s.value
}

// IsEmpty checks if the prompt is empty
func (s SystemPrompt) IsEmpty() bool {
	return s.value == ""
}

// ToolName represents a tool name value object
type ToolName struct {
	value string
}

// NewToolName creates a new ToolName with validation
func NewToolName(value string) (ToolName, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ToolName{}, ErrEmptyContent
	}
	if !toolIDPattern.MatchString(value) {
		return ToolName{}, ErrInvalidToolID
	}
	return ToolName{value: value}, nil
}

// String returns the string representation
func (t ToolName) String() string {
	return t.value
}

// IsEmpty checks if the name is empty
func (t ToolName) IsEmpty() bool {
	return t.value == ""
}

// Equals compares two ToolNames
func (t ToolName) Equals(other ToolName) bool {
	return t.value == other.value
}

// ToolDescription represents a tool description value object
type ToolDescription struct {
	value string
}

// MaxToolDescriptionLength is the maximum allowed length for tool descriptions
const MaxToolDescriptionLength = 10000

// NewToolDescription creates a new ToolDescription with validation
func NewToolDescription(value string) (ToolDescription, error) {
	if len(value) > MaxToolDescriptionLength {
		return ToolDescription{}, ErrContentTooLong
	}
	return ToolDescription{value: value}, nil
}

// String returns the string representation
func (t ToolDescription) String() string {
	return t.value
}

// IsEmpty checks if the description is empty
func (t ToolDescription) IsEmpty() bool {
	return t.value == ""
}

// ResourceURI represents a resource URI value object
type ResourceURI struct {
	value string
}

// NewResourceURI creates a new ResourceURI with validation
func NewResourceURI(value string) (ResourceURI, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ResourceURI{}, ErrEmptyContent
	}
	if !resourceIDPattern.MatchString(value) {
		return ResourceURI{}, ErrInvalidResourceID
	}
	return ResourceURI{value: value}, nil
}

// String returns the string representation
func (r ResourceURI) String() string {
	return r.value
}

// IsEmpty checks if the URI is empty
func (r ResourceURI) IsEmpty() bool {
	return r.value == ""
}

// Equals compares two ResourceURIs
func (r ResourceURI) Equals(other ResourceURI) bool {
	return r.value == other.value
}

// MimeType represents a MIME type value object
type MimeType struct {
	value string
}

// Common MIME types
const (
	MimeTypeJSON      = "application/json"
	MimeTypePlainText = "text/plain"
	MimeTypeMarkdown  = "text/markdown"
	MimeTypeHTML      = "text/html"
	MimeTypeXML       = "application/xml"
	MimeTypePNG       = "image/png"
	MimeTypeJPEG      = "image/jpeg"
	MimeTypeGIF       = "image/gif"
	MimeTypeWebP      = "image/webp"
	MimeTypePDF       = "application/pdf"
)

// NewMimeType creates a new MimeType with validation
func NewMimeType(value string) (MimeType, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return MimeType{value: MimeTypePlainText}, nil // Default to plain text
	}
	return MimeType{value: value}, nil
}

// String returns the string representation
func (m MimeType) String() string {
	return m.value
}

// IsEmpty checks if the MIME type is empty
func (m MimeType) IsEmpty() bool {
	return m.value == ""
}

// IsText checks if the MIME type is a text type
func (m MimeType) IsText() bool {
	return strings.HasPrefix(m.value, "text/") || m.value == MimeTypeJSON || m.value == MimeTypeXML
}

// IsImage checks if the MIME type is an image type
func (m MimeType) IsImage() bool {
	return strings.HasPrefix(m.value, "image/")
}
