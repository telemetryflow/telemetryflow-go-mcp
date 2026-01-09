// Package mcp provides tests for Model Context Protocol types
package mcp

import (
	"encoding/json"
	"testing"
)

func TestRequest_Marshaling(t *testing.T) {
	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", JSONRPCVersion, decoded.JSONRPC)
	}

	if decoded.Method != "initialize" {
		t.Errorf("Expected method 'initialize', got %s", decoded.Method)
	}
}

func TestResponse_Success(t *testing.T) {
	result := map[string]interface{}{
		"protocolVersion": ProtocolVersion,
		"serverInfo": map[string]string{
			"name":    "TelemetryFlow-MCP",
			"version": "1.1.2",
		},
	}

	resp := NewResponse(1, result)

	if resp.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", JSONRPCVersion, resp.JSONRPC)
	}

	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %v", resp.ID)
	}

	if resp.Error != nil {
		t.Error("Error should be nil for success response")
	}

	if resp.Result == nil {
		t.Error("Result should not be nil")
	}
}

func TestResponse_Error(t *testing.T) {
	err := NewInternalError("something went wrong")
	resp := NewErrorResponse(1, err)

	if resp.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", JSONRPCVersion, resp.JSONRPC)
	}

	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %v", resp.ID)
	}

	if resp.Result != nil {
		t.Error("Result should be nil for error response")
	}

	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}

	if resp.Error.Code != InternalError {
		t.Errorf("Expected error code %d, got %d", InternalError, resp.Error.Code)
	}
}

func TestNotification_Creation(t *testing.T) {
	params := map[string]string{"key": "value"}
	notif, err := NewNotification("notifications/message", params)

	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if notif.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", JSONRPCVersion, notif.JSONRPC)
	}

	if notif.Method != "notifications/message" {
		t.Errorf("Expected method 'notifications/message', got %s", notif.Method)
	}
}

func TestNotification_NilParams(t *testing.T) {
	notif, err := NewNotification("notifications/cancelled", nil)

	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if notif.Params != nil {
		t.Error("Params should be nil")
	}
}

func TestError_Interface(t *testing.T) {
	err := NewError(InvalidParams, "invalid parameter", nil)

	// Test error interface
	var e error = err
	if e.Error() != "MCP error -32602: invalid parameter" {
		t.Errorf("Unexpected error string: %s", e.Error())
	}
}

func TestNewParseError(t *testing.T) {
	err := NewParseError("invalid JSON")

	if err.Code != ParseError {
		t.Errorf("Expected code %d, got %d", ParseError, err.Code)
	}

	if err.Message != "invalid JSON" {
		t.Errorf("Expected message 'invalid JSON', got %s", err.Message)
	}
}

func TestNewInvalidRequestError(t *testing.T) {
	err := NewInvalidRequestError("missing method")

	if err.Code != InvalidRequest {
		t.Errorf("Expected code %d, got %d", InvalidRequest, err.Code)
	}
}

func TestNewMethodNotFoundError(t *testing.T) {
	err := NewMethodNotFoundError("unknown/method")

	if err.Code != MethodNotFound {
		t.Errorf("Expected code %d, got %d", MethodNotFound, err.Code)
	}

	if err.Message != "method not found: unknown/method" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewInvalidParamsError(t *testing.T) {
	err := NewInvalidParamsError("missing required field")

	if err.Code != InvalidParams {
		t.Errorf("Expected code %d, got %d", InvalidParams, err.Code)
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("database connection failed")

	if err.Code != InternalError {
		t.Errorf("Expected code %d, got %d", InternalError, err.Code)
	}
}

func TestNewSessionNotFoundError(t *testing.T) {
	err := NewSessionNotFoundError("abc-123")

	if err.Code != SessionNotFound {
		t.Errorf("Expected code %d, got %d", SessionNotFound, err.Code)
	}

	if err.Message != "session not found: abc-123" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewToolNotFoundError(t *testing.T) {
	err := NewToolNotFoundError("unknown_tool")

	if err.Code != ToolNotFound {
		t.Errorf("Expected code %d, got %d", ToolNotFound, err.Code)
	}

	if err.Message != "tool not found: unknown_tool" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewResourceNotFoundError(t *testing.T) {
	err := NewResourceNotFoundError("file:///missing")

	if err.Code != ResourceNotFound {
		t.Errorf("Expected code %d, got %d", ResourceNotFound, err.Code)
	}

	if err.Message != "resource not found: file:///missing" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewPromptNotFoundError(t *testing.T) {
	err := NewPromptNotFoundError("unknown_prompt")

	if err.Code != PromptNotFound {
		t.Errorf("Expected code %d, got %d", PromptNotFound, err.Code)
	}

	if err.Message != "prompt not found: unknown_prompt" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestError_WithData(t *testing.T) {
	data := map[string]interface{}{
		"field":  "name",
		"reason": "too short",
	}
	err := NewError(InvalidParams, "validation failed", data)

	if err.Data == nil {
		t.Error("Data should not be nil")
	}

	dataMap, ok := err.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if dataMap["field"] != "name" {
		t.Errorf("Expected field 'name', got %v", dataMap["field"])
	}
}

func TestConstants(t *testing.T) {
	if ProtocolVersion != "2024-11-05" {
		t.Errorf("Unexpected ProtocolVersion: %s", ProtocolVersion)
	}

	if JSONRPCVersion != "2.0" {
		t.Errorf("Unexpected JSONRPCVersion: %s", JSONRPCVersion)
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"ParseError", ParseError, -32700},
		{"InvalidRequest", InvalidRequest, -32600},
		{"MethodNotFound", MethodNotFound, -32601},
		{"InvalidParams", InvalidParams, -32602},
		{"InternalError", InternalError, -32603},
		{"ServerError", ServerError, -32000},
		{"SessionNotFound", SessionNotFound, -32001},
		{"ToolNotFound", ToolNotFound, -32002},
		{"ResourceNotFound", ResourceNotFound, -32003},
		{"PromptNotFound", PromptNotFound, -32004},
		{"InvalidSessionState", InvalidSessionState, -32005},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected %s to be %d, got %d", tt.name, tt.expected, tt.code)
			}
		})
	}
}

func BenchmarkNewResponse(b *testing.B) {
	result := map[string]interface{}{
		"status": "ok",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResponse(i, result)
	}
}

func BenchmarkNewNotification(b *testing.B) {
	params := map[string]string{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewNotification("test/method", params)
	}
}

func BenchmarkError_Error(b *testing.B) {
	err := NewInternalError("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
