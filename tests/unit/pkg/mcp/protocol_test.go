package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/telemetryflow/telemetryflow-go-mcp/pkg/mcp"
)

func TestRequest_Marshaling(t *testing.T) {
	req := mcp.Request{
		JSONRPC: mcp.JSONRPCVersion,
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded mcp.Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.JSONRPC != mcp.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", mcp.JSONRPCVersion, decoded.JSONRPC)
	}

	if decoded.Method != "initialize" {
		t.Errorf("Expected method 'initialize', got %s", decoded.Method)
	}
}

func TestResponse_Success(t *testing.T) {
	result := map[string]interface{}{
		"protocolVersion": mcp.ProtocolVersion,
		"serverInfo": map[string]string{
			"name":    "TelemetryFlow-MCP",
			"version": "1.2.0",
		},
	}

	resp := mcp.NewResponse(1, result)

	if resp.JSONRPC != mcp.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", mcp.JSONRPCVersion, resp.JSONRPC)
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
	err := mcp.NewInternalError("something went wrong")
	resp := mcp.NewErrorResponse(1, err)

	if resp.JSONRPC != mcp.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", mcp.JSONRPCVersion, resp.JSONRPC)
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

	if resp.Error.Code != mcp.InternalError {
		t.Errorf("Expected error code %d, got %d", mcp.InternalError, resp.Error.Code)
	}
}

func TestNotification_Creation(t *testing.T) {
	params := map[string]string{"key": "value"}
	notif, err := mcp.NewNotification("notifications/message", params)

	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if notif.JSONRPC != mcp.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", mcp.JSONRPCVersion, notif.JSONRPC)
	}

	if notif.Method != "notifications/message" {
		t.Errorf("Expected method 'notifications/message', got %s", notif.Method)
	}
}

func TestNotification_NilParams(t *testing.T) {
	notif, err := mcp.NewNotification("notifications/cancelled", nil)

	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if notif.Params != nil {
		t.Error("Params should be nil")
	}
}

func TestError_Interface(t *testing.T) {
	err := mcp.NewError(mcp.InvalidParams, "invalid parameter", nil)

	var e error = err
	if e.Error() != "MCP error -32602: invalid parameter" {
		t.Errorf("Unexpected error string: %s", e.Error())
	}
}

func TestNewParseError(t *testing.T) {
	err := mcp.NewParseError("invalid JSON")

	if err.Code != mcp.ParseError {
		t.Errorf("Expected code %d, got %d", mcp.ParseError, err.Code)
	}

	if err.Message != "invalid JSON" {
		t.Errorf("Expected message 'invalid JSON', got %s", err.Message)
	}
}

func TestNewInvalidRequestError(t *testing.T) {
	err := mcp.NewInvalidRequestError("missing method")

	if err.Code != mcp.InvalidRequest {
		t.Errorf("Expected code %d, got %d", mcp.InvalidRequest, err.Code)
	}
}

func TestNewMethodNotFoundError(t *testing.T) {
	err := mcp.NewMethodNotFoundError("unknown/method")

	if err.Code != mcp.MethodNotFound {
		t.Errorf("Expected code %d, got %d", mcp.MethodNotFound, err.Code)
	}

	if err.Message != "method not found: unknown/method" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewInvalidParamsError(t *testing.T) {
	err := mcp.NewInvalidParamsError("missing required field")

	if err.Code != mcp.InvalidParams {
		t.Errorf("Expected code %d, got %d", mcp.InvalidParams, err.Code)
	}
}

func TestNewInternalError(t *testing.T) {
	err := mcp.NewInternalError("database connection failed")

	if err.Code != mcp.InternalError {
		t.Errorf("Expected code %d, got %d", mcp.InternalError, err.Code)
	}
}

func TestNewSessionNotFoundError(t *testing.T) {
	err := mcp.NewSessionNotFoundError("abc-123")

	if err.Code != mcp.SessionNotFound {
		t.Errorf("Expected code %d, got %d", mcp.SessionNotFound, err.Code)
	}

	if err.Message != "session not found: abc-123" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewToolNotFoundError(t *testing.T) {
	err := mcp.NewToolNotFoundError("unknown_tool")

	if err.Code != mcp.ToolNotFound {
		t.Errorf("Expected code %d, got %d", mcp.ToolNotFound, err.Code)
	}

	if err.Message != "tool not found: unknown_tool" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewResourceNotFoundError(t *testing.T) {
	err := mcp.NewResourceNotFoundError("file:///missing")

	if err.Code != mcp.ResourceNotFound {
		t.Errorf("Expected code %d, got %d", mcp.ResourceNotFound, err.Code)
	}

	if err.Message != "resource not found: file:///missing" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestNewPromptNotFoundError(t *testing.T) {
	err := mcp.NewPromptNotFoundError("unknown_prompt")

	if err.Code != mcp.PromptNotFound {
		t.Errorf("Expected code %d, got %d", mcp.PromptNotFound, err.Code)
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
	err := mcp.NewError(mcp.InvalidParams, "validation failed", data)

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
	if mcp.ProtocolVersion != "2024-11-05" {
		t.Errorf("Unexpected ProtocolVersion: %s", mcp.ProtocolVersion)
	}

	if mcp.JSONRPCVersion != "2.0" {
		t.Errorf("Unexpected JSONRPCVersion: %s", mcp.JSONRPCVersion)
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"ParseError", mcp.ParseError, -32700},
		{"InvalidRequest", mcp.InvalidRequest, -32600},
		{"MethodNotFound", mcp.MethodNotFound, -32601},
		{"InvalidParams", mcp.InvalidParams, -32602},
		{"InternalError", mcp.InternalError, -32603},
		{"ServerError", mcp.ServerError, -32000},
		{"SessionNotFound", mcp.SessionNotFound, -32001},
		{"ToolNotFound", mcp.ToolNotFound, -32002},
		{"ResourceNotFound", mcp.ResourceNotFound, -32003},
		{"PromptNotFound", mcp.PromptNotFound, -32004},
		{"InvalidSessionState", mcp.InvalidSessionState, -32005},
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
		_ = mcp.NewResponse(i, result)
	}
}

func BenchmarkNewNotification(b *testing.B) {
	params := map[string]string{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mcp.NewNotification("test/method", params)
	}
}

func BenchmarkError_Error(b *testing.B) {
	err := mcp.NewInternalError("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
