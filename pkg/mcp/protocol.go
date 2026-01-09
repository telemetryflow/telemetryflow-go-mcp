// Package mcp provides Model Context Protocol types and utilities
package mcp

import (
	"encoding/json"
	"fmt"
)

// Protocol version
const (
	ProtocolVersion = "2024-11-05"
	JSONRPCVersion  = "2.0"
)

// JSON-RPC Error Codes
const (
	ParseError          = -32700
	InvalidRequest      = -32600
	MethodNotFound      = -32601
	InvalidParams       = -32602
	InternalError       = -32603
	ServerError         = -32000
	SessionNotFound     = -32001
	ToolNotFound        = -32002
	ResourceNotFound    = -32003
	PromptNotFound      = -32004
	InvalidSessionState = -32005
)

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Notification represents a JSON-RPC notification (no ID)
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// NewError creates a new MCP error
func NewError(code int, message string, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewParseError creates a parse error
func NewParseError(message string) *Error {
	return NewError(ParseError, message, nil)
}

// NewInvalidRequestError creates an invalid request error
func NewInvalidRequestError(message string) *Error {
	return NewError(InvalidRequest, message, nil)
}

// NewMethodNotFoundError creates a method not found error
func NewMethodNotFoundError(method string) *Error {
	return NewError(MethodNotFound, fmt.Sprintf("method not found: %s", method), nil)
}

// NewInvalidParamsError creates an invalid params error
func NewInvalidParamsError(message string) *Error {
	return NewError(InvalidParams, message, nil)
}

// NewInternalError creates an internal error
func NewInternalError(message string) *Error {
	return NewError(InternalError, message, nil)
}

// NewSessionNotFoundError creates a session not found error
func NewSessionNotFoundError(sessionID string) *Error {
	return NewError(SessionNotFound, fmt.Sprintf("session not found: %s", sessionID), nil)
}

// NewToolNotFoundError creates a tool not found error
func NewToolNotFoundError(toolName string) *Error {
	return NewError(ToolNotFound, fmt.Sprintf("tool not found: %s", toolName), nil)
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(uri string) *Error {
	return NewError(ResourceNotFound, fmt.Sprintf("resource not found: %s", uri), nil)
}

// NewPromptNotFoundError creates a prompt not found error
func NewPromptNotFoundError(promptName string) *Error {
	return NewError(PromptNotFound, fmt.Sprintf("prompt not found: %s", promptName), nil)
}

// NewResponse creates a successful response
func NewResponse(id interface{}, result interface{}) *Response {
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(id interface{}, err *Error) *Response {
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error:   err,
	}
}

// NewNotification creates a notification
func NewNotification(method string, params interface{}) (*Notification, error) {
	var rawParams json.RawMessage
	if params != nil {
		var err error
		rawParams, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}
	return &Notification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  rawParams,
	}, nil
}
