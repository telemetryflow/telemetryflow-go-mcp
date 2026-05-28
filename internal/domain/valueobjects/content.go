// Package valueobjects contains immutable, self-validating value objects
//
// TelemetryFlow GO MCP Server - Community Enterprise Observability Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
// Open Source Software built by Telemetri Data Indonesia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// Model represents an LLM model identifier
type Model string

// Anthropic Claude
const (
	ModelClaudeOpus47     Model = "claude-opus-4-7"
	ModelClaudeOpus47Fast Model = "claude-opus-4-7-fast"
	ModelClaudeOpus46     Model = "claude-opus-4-6"
	ModelClaudeOpus46Fast Model = "claude-opus-4-6-fast"
	ModelClaudeSonnet46   Model = "claude-sonnet-4-6"
	ModelClaudeOpus45     Model = "claude-opus-4-5"
	ModelClaudeSonnet45   Model = "claude-sonnet-4-5-20250929"
	ModelClaudeHaiku45    Model = "claude-haiku-4-5"
	ModelClaudeHaiku45Oct Model = "claude-haiku-4-5-20251001"
	ModelClaudeSonnet4    Model = "claude-sonnet-4-20250514"
	ModelClaudeMythosPrev Model = "claude-mythos-preview"
)

// Google Gemini
const (
	ModelGemini35Flash       Model = "gemini-3.5-flash"
	ModelGemini31FlashLite   Model = "gemini-3.1-flash-lite"
	ModelGemini31ProPreview  Model = "gemini-3.1-pro-preview"
	ModelGemini3FlashPreview Model = "gemini-3-flash-preview"
	ModelGemini25Pro         Model = "gemini-2.5-pro"
	ModelGemini25Flash       Model = "gemini-2.5-flash"
	ModelGemini25FlashLite   Model = "gemini-2.5-flash-lite"
	ModelGemini20Flash       Model = "gemini-2.0-flash"
	ModelGemini20FlashLite   Model = "gemini-2.0-flash-lite"
	ModelGemini15Pro         Model = "gemini-1.5-pro"
)

// OpenAI
const (
	ModelGPT55Pro  Model = "gpt-5.5-pro"
	ModelGPT55     Model = "gpt-5.5"
	ModelGPT54Pro  Model = "gpt-5.4-pro"
	ModelGPT54     Model = "gpt-5.4"
	ModelGPT54Mini Model = "gpt-5.4-mini"
	ModelGPT54Nano Model = "gpt-5.4-nano"
	ModelGPT53Chat Model = "gpt-5.3-chat"
	ModelGPT5      Model = "gpt-5"
	ModelGPT41     Model = "gpt-4.1"
	ModelO3        Model = "o3"
)

// DeepSeek
const (
	ModelDeepSeekV4Pro       Model = "deepseek-v4-pro"
	ModelDeepSeekV4Flash     Model = "deepseek-v4-flash"
	ModelDeepSeekV32Speciale Model = "deepseek-v3.2-speciale"
	ModelDeepSeekChat        Model = "deepseek-chat"
	ModelDeepSeekV32         Model = "deepseek-v3.2"
	ModelDeepSeekV32Exp      Model = "deepseek-v3.2-exp"
	ModelDeepSeekV31Terminus Model = "deepseek-v3.1-terminus"
	ModelDeepSeekChatV31     Model = "deepseek-chat-v3.1"
	ModelDeepSeekR10528      Model = "deepseek-r1-0528"
	ModelDeepSeekReasoner    Model = "deepseek-reasoner"
)

// Alibaba Qwen
const (
	ModelQwen36MaxPreview Model = "qwen3.6-max-preview"
	ModelQwen36Plus       Model = "qwen3.6-plus"
	ModelQwen36Flash      Model = "qwen3.6-flash"
	ModelQwen3635BA3B     Model = "qwen3.6-35b-a3b"
	ModelQwen3627B        Model = "qwen3.6-27b"
	ModelQwen35Plus       Model = "qwen3.5-plus"
	ModelQwen359B         Model = "qwen3.5-9b"
	ModelQwen3535BA3B     Model = "qwen3.5-35b-a3b"
	ModelQwen3527B        Model = "qwen3.5-27b"
	ModelQwen35122BA10B   Model = "qwen3.5-122b-a10b"
)

// Mistral AI
const (
	ModelMistralMedium35      Model = "mistral-medium-3-5"
	ModelMistralSmall4        Model = "mistral-small-2603"
	ModelMistralLarge3        Model = "mistral-large-2512"
	ModelMistralDevstral2     Model = "devstral-2512"
	ModelMistralMinistral314B Model = "ministral-14b-2512"
	ModelMistralMinistral38B  Model = "ministral-8b-2512"
	ModelMistralMinistral33B  Model = "ministral-3b-2512"
	ModelMistralMedium31      Model = "mistral-medium-2508"
	ModelMistralCodestral     Model = "codestral-2508"
	ModelMistralLarge21       Model = "mistral-large-2411"
)

// xAI Grok
const (
	ModelGrok43              Model = "grok-4.3"
	ModelGrok420MultiAgent   Model = "grok-4.20-multi-agent"
	ModelGrok420Reasoning    Model = "grok-4.20-0309-reasoning"
	ModelGrok420NonReasoning Model = "grok-4.20-0309-non-reasoning"
	ModelGrok41FastReasoning Model = "grok-4-1-fast-reasoning"
	ModelGrok41FastNonReason Model = "grok-4-1-fast-non-reasoning"
	ModelGrok3               Model = "grok-3"
	ModelGrok3Mini           Model = "grok-3-mini"
	ModelGrok2               Model = "grok-2"
	ModelGrok2Mini           Model = "grok-2-mini"
)

// Kimi / Moonshot AI
const (
	ModelKimiK26            Model = "kimi-k2.6"
	ModelKimiK25            Model = "kimi-k2.5"
	ModelKimiK2Thinking     Model = "kimi-k2-thinking"
	ModelKimiK20905         Model = "kimi-k2-0905"
	ModelKimiK2TurboPreview Model = "kimi-k2-turbo-preview"
	ModelKimiK2             Model = "kimi-k2"
	ModelMoonshotV1128K     Model = "moonshot-v1-128k"
	ModelMoonshotV132K      Model = "moonshot-v1-32k"
	ModelMoonshotV18K       Model = "moonshot-v1-8k"
	ModelMoonshotV1Auto     Model = "moonshot-v1-auto"
)

// Zhipu GLM
const (
	ModelGLM51      Model = "glm-5.1"
	ModelGLM5Turbo  Model = "glm-5-turbo"
	ModelGLM5       Model = "glm-5"
	ModelGLM47Flash Model = "glm-4.7-flash"
	ModelGLM47      Model = "glm-4.7"
	ModelGLM46      Model = "glm-4.6"
	ModelGLM45      Model = "glm-4.5"
	ModelGLM45Air   Model = "glm-4.5-air"
	ModelGLM4Flash  Model = "glm-4-flash"
	ModelGLM4       Model = "glm-4"
)

// Xiaomi MiMo
const (
	ModelMiMoV25Pro  Model = "mimo-v2.5-pro"
	ModelMiMoV25     Model = "mimo-v2.5"
	ModelMiMoV2Omni  Model = "mimo-v2-omni"
	ModelMiMoV2Pro   Model = "mimo-v2-pro"
	ModelMiMoV2Flash Model = "mimo-v2-flash"
	ModelMiMoV2TTS   Model = "mimo-v2-tts"
	ModelMiMo7B      Model = "mimo-7b"
	ModelMiMoVL7B    Model = "mimo-vl-7b"
	ModelMiMoV25Lite Model = "mimo-v2.5-lite"
	ModelMiMo7B0321  Model = "mimo-7b-0321"
)

// DefaultModel is the default model to use
const DefaultModel = ModelClaudeOpus47

// IsValid checks if the model is a known TFO-Platform model
func (m Model) IsValid() bool {
	switch m {
	case ModelClaudeOpus47, ModelClaudeOpus47Fast, ModelClaudeOpus46, ModelClaudeOpus46Fast,
		ModelClaudeSonnet46, ModelClaudeOpus45, ModelClaudeSonnet45, ModelClaudeHaiku45,
		ModelClaudeHaiku45Oct, ModelClaudeSonnet4, ModelClaudeMythosPrev,
		ModelGemini35Flash, ModelGemini31FlashLite, ModelGemini31ProPreview,
		ModelGemini3FlashPreview, ModelGemini25Pro, ModelGemini25Flash,
		ModelGemini25FlashLite, ModelGemini20Flash, ModelGemini20FlashLite, ModelGemini15Pro,
		ModelGPT55Pro, ModelGPT55, ModelGPT54Pro, ModelGPT54, ModelGPT54Mini,
		ModelGPT54Nano, ModelGPT53Chat, ModelGPT5, ModelGPT41, ModelO3,
		ModelDeepSeekV4Pro, ModelDeepSeekV4Flash, ModelDeepSeekV32Speciale,
		ModelDeepSeekChat, ModelDeepSeekV32, ModelDeepSeekV32Exp,
		ModelDeepSeekV31Terminus, ModelDeepSeekChatV31,
		ModelDeepSeekR10528, ModelDeepSeekReasoner,
		ModelQwen36MaxPreview, ModelQwen36Plus, ModelQwen36Flash,
		ModelQwen3635BA3B, ModelQwen3627B, ModelQwen35Plus, ModelQwen359B,
		ModelQwen3535BA3B, ModelQwen3527B, ModelQwen35122BA10B,
		ModelMistralMedium35, ModelMistralSmall4, ModelMistralLarge3,
		ModelMistralDevstral2, ModelMistralMinistral314B, ModelMistralMinistral38B,
		ModelMistralMinistral33B, ModelMistralMedium31, ModelMistralCodestral, ModelMistralLarge21,
		ModelGrok43, ModelGrok420MultiAgent, ModelGrok420Reasoning, ModelGrok420NonReasoning,
		ModelGrok41FastReasoning, ModelGrok41FastNonReason,
		ModelGrok3, ModelGrok3Mini, ModelGrok2, ModelGrok2Mini,
		ModelKimiK26, ModelKimiK25, ModelKimiK2Thinking, ModelKimiK20905,
		ModelKimiK2TurboPreview, ModelKimiK2,
		ModelMoonshotV1128K, ModelMoonshotV132K, ModelMoonshotV18K, ModelMoonshotV1Auto,
		ModelGLM51, ModelGLM5Turbo, ModelGLM5, ModelGLM47Flash, ModelGLM47,
		ModelGLM46, ModelGLM45, ModelGLM45Air, ModelGLM4Flash, ModelGLM4,
		ModelMiMoV25Pro, ModelMiMoV25, ModelMiMoV2Omni, ModelMiMoV2Pro,
		ModelMiMoV2Flash, ModelMiMoV2TTS, ModelMiMo7B, ModelMiMoVL7B,
		ModelMiMoV25Lite, ModelMiMo7B0321:
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
