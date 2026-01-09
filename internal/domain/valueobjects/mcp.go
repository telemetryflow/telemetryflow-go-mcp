// Package valueobjects contains MCP-specific value objects
package valueobjects

import (
	"errors"
	"strings"
)

// MCP validation errors
var (
	ErrInvalidJSONRPCVersion = errors.New("invalid JSON-RPC version")
	ErrInvalidMethod         = errors.New("invalid method")
	ErrInvalidCapability     = errors.New("invalid capability")
)

// JSONRPCVersion represents the JSON-RPC version
type JSONRPCVersion string

const (
	JSONRPC20 JSONRPCVersion = "2.0"
)

// IsValid checks if the version is valid
func (v JSONRPCVersion) IsValid() bool {
	return v == JSONRPC20
}

// String returns the string representation
func (v JSONRPCVersion) String() string {
	return string(v)
}

// MCPMethod represents an MCP method
type MCPMethod string

// MCP Protocol Methods
const (
	// Lifecycle methods
	MethodInitialize  MCPMethod = "initialize"
	MethodInitialized MCPMethod = "notifications/initialized"
	MethodPing        MCPMethod = "ping"
	MethodShutdown    MCPMethod = "shutdown"

	// Tool methods
	MethodToolsList MCPMethod = "tools/list"
	MethodToolsCall MCPMethod = "tools/call"

	// Resource methods
	MethodResourcesList        MCPMethod = "resources/list"
	MethodResourcesRead        MCPMethod = "resources/read"
	MethodResourcesSubscribe   MCPMethod = "resources/subscribe"
	MethodResourcesUnsubscribe MCPMethod = "resources/unsubscribe"

	// Prompt methods
	MethodPromptsList MCPMethod = "prompts/list"
	MethodPromptsGet  MCPMethod = "prompts/get"

	// Completion methods
	MethodCompletionComplete MCPMethod = "completion/complete"

	// Logging methods
	MethodLoggingSetLevel MCPMethod = "logging/setLevel"

	// Notification methods
	MethodNotificationsCancelled            MCPMethod = "notifications/cancelled"
	MethodNotificationsProgress             MCPMethod = "notifications/progress"
	MethodNotificationsMessage              MCPMethod = "notifications/message"
	MethodNotificationsResourcesUpdated     MCPMethod = "notifications/resources/updated"
	MethodNotificationsResourcesListChanged MCPMethod = "notifications/resources/list_changed"
	MethodNotificationsToolsListChanged     MCPMethod = "notifications/tools/list_changed"
	MethodNotificationsPromptsListChanged   MCPMethod = "notifications/prompts/list_changed"
)

// IsValid checks if the method is valid
func (m MCPMethod) IsValid() bool {
	switch m {
	case MethodInitialize, MethodInitialized, MethodPing, MethodShutdown,
		MethodToolsList, MethodToolsCall,
		MethodResourcesList, MethodResourcesRead, MethodResourcesSubscribe, MethodResourcesUnsubscribe,
		MethodPromptsList, MethodPromptsGet,
		MethodCompletionComplete, MethodLoggingSetLevel,
		MethodNotificationsCancelled, MethodNotificationsProgress, MethodNotificationsMessage,
		MethodNotificationsResourcesUpdated, MethodNotificationsResourcesListChanged,
		MethodNotificationsToolsListChanged, MethodNotificationsPromptsListChanged:
		return true
	}
	return false
}

// String returns the string representation
func (m MCPMethod) String() string {
	return string(m)
}

// IsNotification checks if the method is a notification
func (m MCPMethod) IsNotification() bool {
	return strings.HasPrefix(string(m), "notifications/")
}

// MCPCapability represents an MCP capability
type MCPCapability string

// MCP Capabilities
const (
	CapabilityTools        MCPCapability = "tools"
	CapabilityResources    MCPCapability = "resources"
	CapabilityPrompts      MCPCapability = "prompts"
	CapabilityLogging      MCPCapability = "logging"
	CapabilitySampling     MCPCapability = "sampling"
	CapabilityRoots        MCPCapability = "roots"
	CapabilityExperimental MCPCapability = "experimental"
)

// IsValid checks if the capability is valid
func (c MCPCapability) IsValid() bool {
	switch c {
	case CapabilityTools, CapabilityResources, CapabilityPrompts,
		CapabilityLogging, CapabilitySampling, CapabilityRoots, CapabilityExperimental:
		return true
	}
	return false
}

// String returns the string representation
func (c MCPCapability) String() string {
	return string(c)
}

// MCPLogLevel represents an MCP log level
type MCPLogLevel string

// MCP Log Levels
const (
	LogLevelDebug     MCPLogLevel = "debug"
	LogLevelInfo      MCPLogLevel = "info"
	LogLevelNotice    MCPLogLevel = "notice"
	LogLevelWarning   MCPLogLevel = "warning"
	LogLevelError     MCPLogLevel = "error"
	LogLevelCritical  MCPLogLevel = "critical"
	LogLevelAlert     MCPLogLevel = "alert"
	LogLevelEmergency MCPLogLevel = "emergency"
)

// IsValid checks if the log level is valid
func (l MCPLogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelNotice, LogLevelWarning,
		LogLevelError, LogLevelCritical, LogLevelAlert, LogLevelEmergency:
		return true
	}
	return false
}

// String returns the string representation
func (l MCPLogLevel) String() string {
	return string(l)
}

// Severity returns the numeric severity (higher = more severe)
func (l MCPLogLevel) Severity() int {
	switch l {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelNotice:
		return 2
	case LogLevelWarning:
		return 3
	case LogLevelError:
		return 4
	case LogLevelCritical:
		return 5
	case LogLevelAlert:
		return 6
	case LogLevelEmergency:
		return 7
	}
	return 0
}

// MCPProtocolVersion represents the MCP protocol version
type MCPProtocolVersion struct {
	value string
}

// Current MCP protocol version
const CurrentMCPProtocolVersion = "2024-11-05"

// NewMCPProtocolVersion creates a new MCPProtocolVersion
func NewMCPProtocolVersion(value string) MCPProtocolVersion {
	if value == "" {
		value = CurrentMCPProtocolVersion
	}
	return MCPProtocolVersion{value: value}
}

// String returns the string representation
func (v MCPProtocolVersion) String() string {
	return v.value
}

// IsLatest checks if this is the latest protocol version
func (v MCPProtocolVersion) IsLatest() bool {
	return v.value == CurrentMCPProtocolVersion
}

// MCPErrorCode represents an MCP error code
type MCPErrorCode int

// MCP Error Codes (JSON-RPC 2.0 standard + MCP specific)
const (
	// Standard JSON-RPC errors
	ErrorCodeParseError     MCPErrorCode = -32700
	ErrorCodeInvalidRequest MCPErrorCode = -32600
	ErrorCodeMethodNotFound MCPErrorCode = -32601
	ErrorCodeInvalidParams  MCPErrorCode = -32602
	ErrorCodeInternalError  MCPErrorCode = -32603

	// MCP specific errors (-32000 to -32099)
	ErrorCodeToolNotFound       MCPErrorCode = -32001
	ErrorCodeResourceNotFound   MCPErrorCode = -32002
	ErrorCodePromptNotFound     MCPErrorCode = -32003
	ErrorCodeToolExecutionError MCPErrorCode = -32004
	ErrorCodeResourceReadError  MCPErrorCode = -32005
	ErrorCodeUnauthorized       MCPErrorCode = -32006
	ErrorCodeRateLimited        MCPErrorCode = -32007
	ErrorCodeTimeout            MCPErrorCode = -32008
	ErrorCodeCancelled          MCPErrorCode = -32009
)

// IsStandardError checks if the error is a standard JSON-RPC error
func (e MCPErrorCode) IsStandardError() bool {
	return e <= -32600 && e >= -32700
}

// IsMCPError checks if the error is an MCP-specific error
func (e MCPErrorCode) IsMCPError() bool {
	return e >= -32099 && e <= -32000
}

// Message returns the default error message for the code
func (e MCPErrorCode) Message() string {
	switch e {
	case ErrorCodeParseError:
		return "Parse error"
	case ErrorCodeInvalidRequest:
		return "Invalid Request"
	case ErrorCodeMethodNotFound:
		return "Method not found"
	case ErrorCodeInvalidParams:
		return "Invalid params"
	case ErrorCodeInternalError:
		return "Internal error"
	case ErrorCodeToolNotFound:
		return "Tool not found"
	case ErrorCodeResourceNotFound:
		return "Resource not found"
	case ErrorCodePromptNotFound:
		return "Prompt not found"
	case ErrorCodeToolExecutionError:
		return "Tool execution error"
	case ErrorCodeResourceReadError:
		return "Resource read error"
	case ErrorCodeUnauthorized:
		return "Unauthorized"
	case ErrorCodeRateLimited:
		return "Rate limited"
	case ErrorCodeTimeout:
		return "Request timeout"
	case ErrorCodeCancelled:
		return "Request cancelled"
	}
	return "Unknown error"
}
