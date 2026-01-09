// Package mcp provides Model Context Protocol types and utilities
package mcp

import "encoding/json"

// ==============================================================================
// Initialization Types
// ==============================================================================

// InitializeParams represents the initialize request parameters
type InitializeParams struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ClientCapability `json:"capabilities"`
	ClientInfo      ClientInfo       `json:"clientInfo"`
}

// InitializeResult represents the initialize response
type InitializeResult struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ServerCapability `json:"capabilities"`
	ServerInfo      ServerInfo       `json:"serverInfo"`
	Instructions    string           `json:"instructions,omitempty"`
}

// ClientInfo represents information about the MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapability represents client capabilities
type ClientCapability struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     *SamplingCapability    `json:"sampling,omitempty"`
	Roots        *RootsCapability       `json:"roots,omitempty"`
}

// ServerCapability represents server capabilities
type ServerCapability struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Completions  *CompletionsCapability `json:"completions,omitempty"`
}

// SamplingCapability represents sampling capability
type SamplingCapability struct{}

// RootsCapability represents roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct{}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// CompletionsCapability represents completions capability
type CompletionsCapability struct{}

// ==============================================================================
// Tool Types
// ==============================================================================

// Tool represents an MCP tool
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolListResult represents the tools/list response
type ToolListResult struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// CallToolParams represents the tools/call request parameters
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents the tools/call response
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ==============================================================================
// Resource Types
// ==============================================================================

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        *int64 `json:"size,omitempty"`
}

// ResourceListResult represents the resources/list response
type ResourceListResult struct {
	Resources  []Resource `json:"resources"`
	NextCursor *string    `json:"nextCursor,omitempty"`
}

// ResourceReadParams represents the resources/read request parameters
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResourceReadResult represents the resources/read response
type ResourceReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents resource content
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ResourceTemplate represents a resource template
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ==============================================================================
// Prompt Types
// ==============================================================================

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptListResult represents the prompts/list response
type PromptListResult struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor *string  `json:"nextCursor,omitempty"`
}

// GetPromptParams represents the prompts/get request parameters
type GetPromptParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// GetPromptResult represents the prompts/get response
type GetPromptResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// ==============================================================================
// Content Types
// ==============================================================================

// ContentBlock represents a block of content
type ContentBlock struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	Data        string                 `json:"data,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Resource    *EmbeddedResource      `json:"resource,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// EmbeddedResource represents an embedded resource in content
type EmbeddedResource struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// NewTextContent creates a text content block
func NewTextContent(text string) ContentBlock {
	return ContentBlock{
		Type: "text",
		Text: text,
	}
}

// NewImageContent creates an image content block
func NewImageContent(data, mimeType string) ContentBlock {
	return ContentBlock{
		Type:     "image",
		Data:     data,
		MimeType: mimeType,
	}
}

// NewResourceContent creates a resource content block
func NewResourceContent(resource *EmbeddedResource) ContentBlock {
	return ContentBlock{
		Type:     "resource",
		Resource: resource,
	}
}

// ==============================================================================
// Logging Types
// ==============================================================================

// LogLevel represents log levels
type LogLevel string

const (
	LogLevelDebug     LogLevel = "debug"
	LogLevelInfo      LogLevel = "info"
	LogLevelNotice    LogLevel = "notice"
	LogLevelWarning   LogLevel = "warning"
	LogLevelError     LogLevel = "error"
	LogLevelCritical  LogLevel = "critical"
	LogLevelAlert     LogLevel = "alert"
	LogLevelEmergency LogLevel = "emergency"
)

// SetLevelParams represents the logging/setLevel request parameters
type SetLevelParams struct {
	Level LogLevel `json:"level"`
}

// LogMessageParams represents a log message notification
type LogMessageParams struct {
	Level  LogLevel    `json:"level"`
	Logger string      `json:"logger,omitempty"`
	Data   interface{} `json:"data"`
}

// ==============================================================================
// Sampling Types
// ==============================================================================

// SamplingMessage represents a message for sampling
type SamplingMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// CreateMessageParams represents the sampling/createMessage request parameters
type CreateMessageParams struct {
	Messages         []SamplingMessage      `json:"messages"`
	ModelPreferences *ModelPreferences      `json:"modelPreferences,omitempty"`
	SystemPrompt     string                 `json:"systemPrompt,omitempty"`
	IncludeContext   string                 `json:"includeContext,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	MaxTokens        int                    `json:"maxTokens"`
	StopSequences    []string               `json:"stopSequences,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ModelPreferences represents model preferences for sampling
type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitempty"`
	CostPriority         *float64    `json:"costPriority,omitempty"`
	SpeedPriority        *float64    `json:"speedPriority,omitempty"`
	IntelligencePriority *float64    `json:"intelligencePriority,omitempty"`
}

// ModelHint represents a hint for model selection
type ModelHint struct {
	Name string `json:"name,omitempty"`
}

// CreateMessageResult represents the sampling/createMessage response
type CreateMessageResult struct {
	Role       string       `json:"role"`
	Content    ContentBlock `json:"content"`
	Model      string       `json:"model"`
	StopReason string       `json:"stopReason,omitempty"`
}

// ==============================================================================
// Pagination Types
// ==============================================================================

// PaginatedParams represents paginated request parameters
type PaginatedParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ==============================================================================
// Progress Types
// ==============================================================================

// ProgressParams represents progress notification parameters
type ProgressParams struct {
	ProgressToken string  `json:"progressToken"`
	Progress      float64 `json:"progress"`
	Total         float64 `json:"total,omitempty"`
}

// CancelledParams represents cancellation notification parameters
type CancelledParams struct {
	RequestID interface{} `json:"requestId"`
	Reason    string      `json:"reason,omitempty"`
}
