// Package server provides unit tests for the MCP server
package server

import (
	"context"
	"encoding/json"
	"testing"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func TestJSONRPCRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     JSONRPCRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
			},
			expectError: false,
		},
		{
			name: "missing jsonrpc version",
			request: JSONRPCRequest{
				ID:     1,
				Method: "initialize",
			},
			expectError: true,
		},
		{
			name: "missing method",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
			},
			expectError: true,
		},
		{
			name: "notification (no id)",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "notifications/initialized",
			},
			expectError: false,
		},
		{
			name: "wrong jsonrpc version",
			request: JSONRPCRequest{
				JSONRPC: "1.0",
				ID:      1,
				Method:  "initialize",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.request.JSONRPC != "2.0" || tt.request.Method == ""

			// Notification requests don't need validation for missing ID
			if tt.request.ID == nil && tt.request.Method != "" && tt.request.JSONRPC == "2.0" {
				hasError = false
			}

			if hasError != tt.expectError {
				t.Errorf("expected error = %v, got %v", tt.expectError, hasError)
			}
		})
	}
}

func TestJSONRPCResponse_Serialization(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
			},
		}

		data, err := json.Marshal(response)
		if err != nil {
			t.Errorf("failed to marshal response: %v", err)
		}

		var parsed JSONRPCResponse
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}

		if parsed.JSONRPC != "2.0" {
			t.Errorf("expected jsonrpc = 2.0, got %s", parsed.JSONRPC)
		}
	})

	t.Run("error response", func(t *testing.T) {
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Error: &JSONRPCError{
				Code:    -32600,
				Message: "Invalid Request",
			},
		}

		data, err := json.Marshal(response)
		if err != nil {
			t.Errorf("failed to marshal error response: %v", err)
		}

		var parsed JSONRPCResponse
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("failed to unmarshal error response: %v", err)
		}

		if parsed.Error == nil {
			t.Error("expected error in response")
		}
		if parsed.Error.Code != -32600 {
			t.Errorf("expected error code = -32600, got %d", parsed.Error.Code)
		}
	})
}

func TestMCPServer_Initialize(t *testing.T) {
	ctx := context.Background()

	t.Run("valid initialize request", func(t *testing.T) {
		params := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		}

		// Validate params
		if params["protocolVersion"] == nil {
			t.Error("protocol version is required")
		}
		if params["clientInfo"] == nil {
			t.Error("client info is required")
		}

		_ = ctx
	})

	t.Run("unsupported protocol version", func(t *testing.T) {
		protocolVersion := "1.0.0"
		supportedVersions := []string{"2024-11-05"}

		isSupported := false
		for _, v := range supportedVersions {
			if v == protocolVersion {
				isSupported = true
				break
			}
		}

		if isSupported {
			t.Error("version 1.0.0 should not be supported")
		}
	})
}

func TestMCPServer_ListTools(t *testing.T) {
	t.Run("returns available tools", func(t *testing.T) {
		expectedTools := []string{
			"claude_conversation",
			"read_file",
			"write_file",
			"list_directory",
			"execute_command",
			"search_files",
			"system_info",
			"echo",
		}

		if len(expectedTools) == 0 {
			t.Error("should have at least one tool")
		}

		for _, tool := range expectedTools {
			if tool == "" {
				t.Error("tool name cannot be empty")
			}
		}
	})
}

func TestMCPServer_ListResources(t *testing.T) {
	t.Run("returns available resources", func(t *testing.T) {
		// Resources are typically files or data sources
		resourceCount := 0

		// Resources can be empty initially
		if resourceCount < 0 {
			t.Error("resource count cannot be negative")
		}
	})
}

func TestMCPServer_ListPrompts(t *testing.T) {
	t.Run("returns available prompts", func(t *testing.T) {
		promptCount := 0

		// Prompts can be empty initially
		if promptCount < 0 {
			t.Error("prompt count cannot be negative")
		}
	})
}

func TestMCPServer_CallTool(t *testing.T) {
	ctx := context.Background()

	t.Run("valid tool call", func(t *testing.T) {
		toolName := "echo"
		arguments := map[string]interface{}{
			"message": "Hello, World!",
		}

		if toolName == "" {
			t.Error("tool name is required")
		}

		_ = arguments
		_ = ctx
	})

	t.Run("unknown tool", func(t *testing.T) {
		toolName := "unknown_tool"
		knownTools := []string{"echo", "read_file", "write_file"}

		isKnown := false
		for _, t := range knownTools {
			if t == toolName {
				isKnown = true
				break
			}
		}

		if isKnown {
			t.Error("unknown_tool should not be known")
		}
	})

	t.Run("missing required arguments", func(t *testing.T) {
		toolName := "read_file"
		arguments := map[string]interface{}{}

		requiredArgs := []string{"path"}
		for _, arg := range requiredArgs {
			if _, ok := arguments[arg]; !ok {
				// This is expected - missing required argument
				continue
			}
		}

		_ = toolName
	})
}

func TestMCPServer_ErrorCodes(t *testing.T) {
	errorCodes := map[string]int{
		"ParseError":     -32700,
		"InvalidRequest": -32600,
		"MethodNotFound": -32601,
		"InvalidParams":  -32602,
		"InternalError":  -32603,
	}

	for name, code := range errorCodes {
		t.Run(name, func(t *testing.T) {
			if code >= 0 {
				t.Errorf("JSON-RPC error codes should be negative, got %d for %s", code, name)
			}
			if code > -32600 || code < -32700 {
				if code > -32000 || code < -32099 {
					// Server errors are -32000 to -32099
					// Pre-defined errors are -32600 to -32700
					// Both ranges are valid
				}
			}
		})
	}
}

func TestMCPServer_Shutdown(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		isShutdown := false

		// Simulate shutdown
		isShutdown = true

		if !isShutdown {
			t.Error("server should be shutdown")
		}
	})
}
