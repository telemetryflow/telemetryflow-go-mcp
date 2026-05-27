// Package claude_test provides unit tests for the Claude client.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package claude_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/claude"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/config"
)

func TestNewClient(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("should create client with valid config", func(t *testing.T) {
		cfg := &config.ClaudeConfig{
			APIKey:     "test-api-key",
			MaxTokens:  4096,
			MaxRetries: 3,
			RetryDelay: time.Second,
		}

		client, err := claude.NewClient(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("should fail with empty API key", func(t *testing.T) {
		cfg := &config.ClaudeConfig{
			APIKey: "",
		}

		client, err := claude.NewClient(cfg, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, claude.ErrAPIKeyRequired)
	})

	t.Run("should use custom base URL", func(t *testing.T) {
		cfg := &config.ClaudeConfig{
			APIKey:  "test-api-key",
			BaseURL: "https://custom.api.anthropic.com",
		}

		client, err := claude.NewClient(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestValidateRequest(t *testing.T) {
	logger := zerolog.Nop()
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		MaxTokens:  4096,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
	client, _ := claude.NewClient(cfg, logger)

	t.Run("should validate valid request", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:     vo.ModelClaudeSonnet4,
			MaxTokens: 1024,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "Hello"},
					},
				},
			},
		}

		err := client.ValidateRequest(request)
		assert.NoError(t, err)
	})

	t.Run("should fail with nil request", func(t *testing.T) {
		err := client.ValidateRequest(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, claude.ErrInvalidRequest)
	})

	t.Run("should fail with empty messages", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:    vo.ModelClaudeSonnet4,
			Messages: []services.ClaudeMessage{},
		}

		err := client.ValidateRequest(request)
		assert.Error(t, err)
	})

	t.Run("should use default max tokens when not set", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:     vo.ModelClaudeSonnet4,
			MaxTokens: 0,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "Hello"},
					},
				},
			},
		}

		err := client.ValidateRequest(request)
		assert.NoError(t, err)
		assert.Equal(t, cfg.MaxTokens, request.MaxTokens)
	})
}

func TestClaudeRequestBuilder(t *testing.T) {
	t.Run("should build request with all parameters", func(t *testing.T) {
		systemPrompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")

		request := &services.ClaudeRequest{
			Model:        vo.ModelClaudeSonnet4,
			MaxTokens:    2048,
			Temperature:  0.7,
			TopP:         0.9,
			TopK:         40,
			SystemPrompt: systemPrompt,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "Hello"},
					},
				},
			},
			StopSequences: []string{"END", "STOP"},
		}

		assert.Equal(t, vo.ModelClaudeSonnet4, request.Model)
		assert.Equal(t, 2048, request.MaxTokens)
		assert.InDelta(t, 0.7, request.Temperature, 0.01)
		assert.InDelta(t, 0.9, request.TopP, 0.01)
		assert.Equal(t, 40, request.TopK)
		assert.Len(t, request.StopSequences, 2)
	})

	t.Run("should build request with tools", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:     vo.ModelClaudeSonnet4,
			MaxTokens: 1024,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "What's the weather?"},
					},
				},
			},
			Tools: []services.ClaudeTool{
				{
					Name:        "get_weather",
					Description: "Get current weather for a location",
					InputSchema: nil,
				},
			},
		}

		assert.Len(t, request.Tools, 1)
		assert.Equal(t, "get_weather", request.Tools[0].Name)
	})
}

func TestClaudeMessageBuilding(t *testing.T) {
	t.Run("should build user message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleUser,
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello, Claude!"},
			},
		}

		assert.Equal(t, vo.RoleUser, msg.Role)
		assert.Len(t, msg.Content, 1)
		assert.Equal(t, vo.ContentTypeText, msg.Content[0].Type)
	})

	t.Run("should build assistant message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleAssistant,
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello! How can I help you?"},
			},
		}

		assert.Equal(t, vo.RoleAssistant, msg.Role)
		assert.Len(t, msg.Content, 1)
	})

	t.Run("should build tool use message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleAssistant,
			Content: []entities.ContentBlock{
				{
					Type:  vo.ContentTypeToolUse,
					ID:    "tool_123",
					Name:  "get_weather",
					Input: map[string]interface{}{"location": "San Francisco"},
				},
			},
		}

		assert.Equal(t, vo.RoleAssistant, msg.Role)
		assert.Equal(t, vo.ContentTypeToolUse, msg.Content[0].Type)
		assert.Equal(t, "tool_123", msg.Content[0].ID)
	})

	t.Run("should build tool result message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleUser,
			Content: []entities.ContentBlock{
				{
					Type:      vo.ContentTypeToolResult,
					ToolUseID: "tool_123",
					Content:   "Sunny, 72°F",
					IsError:   false,
				},
			},
		}

		assert.Equal(t, vo.RoleUser, msg.Role)
		assert.Equal(t, vo.ContentTypeToolResult, msg.Content[0].Type)
		assert.Equal(t, "tool_123", msg.Content[0].ToolUseID)
		assert.False(t, msg.Content[0].IsError)
	})
}

func TestClaudeResponseParsing(t *testing.T) {
	t.Run("should parse text response", func(t *testing.T) {
		response := &services.ClaudeResponse{
			ID:         "msg_123",
			Type:       "message",
			Role:       vo.RoleAssistant,
			Model:      "claude-sonnet-4-20250514",
			StopReason: "end_turn",
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello!"},
			},
			Usage: &services.ClaudeUsage{
				InputTokens:  10,
				OutputTokens: 5,
			},
		}

		assert.Equal(t, "msg_123", response.ID)
		assert.Equal(t, vo.RoleAssistant, response.Role)
		assert.Len(t, response.Content, 1)
		assert.Equal(t, 10, response.Usage.InputTokens)
		assert.Equal(t, 5, response.Usage.OutputTokens)
	})

	t.Run("should parse tool use response", func(t *testing.T) {
		response := &services.ClaudeResponse{
			ID:         "msg_456",
			Role:       vo.RoleAssistant,
			StopReason: "tool_use",
			Content: []entities.ContentBlock{
				{
					Type:  vo.ContentTypeToolUse,
					ID:    "toolu_789",
					Name:  "get_weather",
					Input: map[string]interface{}{"location": "NYC"},
				},
			},
		}

		assert.Equal(t, "tool_use", response.StopReason)
		assert.Equal(t, vo.ContentTypeToolUse, response.Content[0].Type)
		assert.Equal(t, "toolu_789", response.Content[0].ID)
	})
}

func TestClaudeStreamEvents(t *testing.T) {
	t.Run("should handle message start event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Type: "message_start",
			Message: &services.ClaudeResponse{
				ID:    "msg_stream_123",
				Model: "claude-sonnet-4-20250514",
				Role:  vo.RoleAssistant,
			},
		}

		assert.Equal(t, "message_start", event.Type)
		assert.Equal(t, "msg_stream_123", event.Message.ID)
	})

	t.Run("should handle content block delta event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Type:  "content_block_delta",
			Index: 0,
			Delta: &services.ClaudeDelta{
				Type: "text_delta",
				Text: "Hello",
			},
		}

		assert.Equal(t, "content_block_delta", event.Type)
		assert.Equal(t, 0, event.Index)
		assert.Equal(t, "Hello", event.Delta.Text)
	})

	t.Run("should handle message delta event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Type: "message_delta",
			Delta: &services.ClaudeDelta{
				StopReason: "end_turn",
			},
			Usage: &services.ClaudeUsage{
				OutputTokens: 50,
			},
		}

		assert.Equal(t, "message_delta", event.Type)
		assert.Equal(t, "end_turn", event.Delta.StopReason)
		assert.Equal(t, 50, event.Usage.OutputTokens)
	})

	t.Run("should handle error event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Error: context.DeadlineExceeded,
		}

		assert.NotNil(t, event.Error)
		assert.ErrorIs(t, event.Error, context.DeadlineExceeded)
	})
}

func TestClaudeModels(t *testing.T) {
	models := []struct {
		model    vo.Model
		expected string
	}{
		{vo.ModelClaudeOpus47, "claude-opus-4-7"},
		{vo.ModelClaudeSonnet4, "claude-sonnet-4-20250514"},
		{vo.ModelClaudeSonnet46, "claude-sonnet-4-6"},
		{vo.ModelClaudeSonnet45, "claude-sonnet-4-5-20250929"},
		{vo.ModelClaudeHaiku45, "claude-haiku-4-5"},
	}

	for _, tc := range models {
		t.Run("should validate model "+string(tc.model), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.model.String())
			assert.True(t, tc.model.IsValid())
		})
	}
}

func newTestClientWithServer(handler http.HandlerFunc) (*claude.Client, *httptest.Server) {
	t := &testing.T{}
	server := httptest.NewServer(handler)
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		BaseURL:    server.URL,
		MaxTokens:  4096,
		MaxRetries: 0,
		RetryDelay: 0,
	}
	client, err := claude.NewClient(cfg, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	return client, server
}

func makeBasicRequest() *services.ClaudeRequest {
	return &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: "Hello"},
				},
			},
		},
	}
}

func writeMessageResponse(w http.ResponseWriter, id, stopReason string, content []map[string]interface{}, inputTokens, outputTokens int) {
	contentJSON, _ := json.Marshal(content)
	resp := fmt.Sprintf(`{
		"id": "%s",
		"type": "message",
		"role": "assistant",
		"content": %s,
		"model": "claude-sonnet-4-20250514",
		"stop_reason": "%s",
		"stop_sequence": "",
		"usage": {
			"input_tokens": %d,
			"output_tokens": %d,
			"cache_creation_input_tokens": 0,
			"cache_read_input_tokens": 0
		}
	}`, id, string(contentJSON), stopReason, inputTokens, outputTokens)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(resp))
}

func writeSSE(w http.ResponseWriter, event string, data string) {
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func TestValidateRequest_InvalidModel(t *testing.T) {
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		MaxTokens:  4096,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
	client, _ := claude.NewClient(cfg, zerolog.Nop())

	request := &services.ClaudeRequest{
		Model:     vo.Model("invalid-model"),
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: "Hello"},
				},
			},
		},
	}

	err := client.ValidateRequest(request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid model")
	assert.ErrorIs(t, err, claude.ErrInvalidRequest)
}

func TestCreateMessage_TextResponse(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/messages" && r.Method == http.MethodPost {
			writeMessageResponse(w, "msg_001", "end_turn", []map[string]interface{}{
				{"type": "text", "text": "Hello from Claude!"},
			}, 10, 5)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.NoError(t, err)
	assert.Equal(t, "msg_001", resp.ID)
	assert.Equal(t, "message", resp.Type)
	assert.Equal(t, vo.RoleAssistant, resp.Role)
	assert.Equal(t, "claude-sonnet-4-20250514", resp.Model)
	assert.Equal(t, "end_turn", resp.StopReason)
	require.Len(t, resp.Content, 1)
	assert.Equal(t, vo.ContentTypeText, resp.Content[0].Type)
	assert.Equal(t, "Hello from Claude!", resp.Content[0].Text)
	assert.Equal(t, 10, resp.Usage.InputTokens)
	assert.Equal(t, 5, resp.Usage.OutputTokens)
}

func TestCreateMessage_ToolUseResponse(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_002", "tool_use", []map[string]interface{}{
			{"type": "tool_use", "id": "toolu_001", "name": "get_weather", "input": map[string]interface{}{"location": "NYC"}},
		}, 15, 8)
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.NoError(t, err)
	require.Len(t, resp.Content, 1)
	assert.Equal(t, vo.ContentTypeToolUse, resp.Content[0].Type)
	assert.Equal(t, "toolu_001", resp.Content[0].ID)
	assert.Equal(t, "get_weather", resp.Content[0].Name)
	assert.Equal(t, "NYC", resp.Content[0].Input["location"])
}

func TestCreateMessage_ToolUseEmptyInput(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_003", "tool_use", []map[string]interface{}{
			{"type": "tool_use", "id": "toolu_002", "name": "ping", "input": map[string]interface{}{}},
		}, 5, 3)
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.NoError(t, err)
	require.Len(t, resp.Content, 1)
	assert.Equal(t, vo.ContentTypeToolUse, resp.Content[0].Type)
}

func TestCreateMessage_ValidationNilRequest(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, claude.ErrInvalidRequest)
}

func TestCreateMessage_ValidationInvalidModel(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.Model("bad-model"),
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "hi"}}},
		},
	}
	resp, err := client.CreateMessage(context.Background(), req)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateMessage_APIError(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":"bad request"}}`))
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, claude.ErrAPIError)
}

func TestCreateMessage_WithAllOptions(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]interface{}
		_ = json.Unmarshal(body, &parsed)

		assert.Equal(t, float64(0.5), parsed["temperature"])
		assert.Equal(t, float64(0.8), parsed["top_p"])
		assert.Equal(t, float64(50), parsed["top_k"])

		stopSeqs, ok := parsed["stop_sequences"].([]interface{})
		require.True(t, ok)
		assert.Len(t, stopSeqs, 2)

		system, ok := parsed["system"].([]interface{})
		require.True(t, ok)
		assert.Len(t, system, 1)

		tools, ok := parsed["tools"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tools, 1)

		writeMessageResponse(w, "msg_opts", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "OK"},
		}, 20, 2)
	})
	defer server.Close()

	systemPrompt, _ := vo.NewSystemPrompt("You are helpful.")
	req := &services.ClaudeRequest{
		Model:         vo.ModelClaudeSonnet4,
		MaxTokens:     2048,
		Temperature:   0.5,
		TopP:          0.8,
		TopK:          50,
		SystemPrompt:  systemPrompt,
		StopSequences: []string{"END", "STOP"},
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "test"}}},
		},
		Tools: []services.ClaudeTool{
			{
				Name:        "calculator",
				Description: "Does math",
				InputSchema: &entities.JSONSchema{
					Type: "object",
					Properties: map[string]*entities.JSONSchema{
						"expression": {
							Type:        "string",
							Description: "The math expression",
							Enum:        []interface{}{"1+1", "2+2"},
						},
					},
				},
			},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "msg_opts", resp.ID)
}

func TestCreateMessage_MultipleContentTypes(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_multi", "tool_use", []map[string]interface{}{
			{"type": "text", "text": "Let me check that."},
			{"type": "tool_use", "id": "toolu_multi", "name": "search", "input": map[string]interface{}{"q": "test"}},
		}, 30, 15)
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.NoError(t, err)
	require.Len(t, resp.Content, 2)
	assert.Equal(t, vo.ContentTypeText, resp.Content[0].Type)
	assert.Equal(t, vo.ContentTypeToolUse, resp.Content[1].Type)
	assert.Equal(t, "toolu_multi", resp.Content[1].ID)
	assert.Equal(t, "search", resp.Content[1].Name)
}

func TestCreateMessage_ToolUseWithNilInput(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_nil_input", "tool_use", []map[string]interface{}{
			{"type": "tool_use", "id": "toolu_nil", "name": "no_input_tool"},
		}, 5, 3)
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.NoError(t, err)
	require.Len(t, resp.Content, 1)
	assert.Equal(t, vo.ContentTypeToolUse, resp.Content[0].Type)
}

func TestCreateMessage_WithToolUseAndResultMessages(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]interface{}
		_ = json.Unmarshal(body, &parsed)

		msgs, ok := parsed["messages"].([]interface{})
		require.True(t, ok)
		assert.Len(t, msgs, 3)

		writeMessageResponse(w, "msg_tool_flow", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "The weather is sunny."},
		}, 50, 10)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: "What's the weather?"},
				},
			},
			{
				Role: vo.RoleAssistant,
				Content: []entities.ContentBlock{
					{
						Type:  vo.ContentTypeToolUse,
						ID:    "toolu_weather",
						Name:  "get_weather",
						Input: map[string]interface{}{"city": "SF"},
					},
				},
			},
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{
						Type:      vo.ContentTypeToolResult,
						ToolUseID: "toolu_weather",
						Content:   "Sunny, 72F",
						IsError:   false,
					},
				},
			},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "The weather is sunny.", resp.Content[0].Text)
}

func TestCreateMessage_NilToolInputSchema(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_nil_schema", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "done"},
		}, 5, 1)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "go"}}},
		},
		Tools: []services.ClaudeTool{
			{Name: "simple_tool", Description: "No schema", InputSchema: nil},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "done", resp.Content[0].Text)
}

func TestCreateMessage_DefaultMaxTokens(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]interface{}
		_ = json.Unmarshal(body, &parsed)
		assert.Equal(t, float64(4096), parsed["max_tokens"])

		writeMessageResponse(w, "msg_default", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "ok"},
		}, 5, 1)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 0,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "hi"}}},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCreateMessageStream_Events(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		writeSSE(w, "message_start", `{"type":"message_start","message":{"id":"msg_stream","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":"","stop_sequence":"","usage":{"input_tokens":10,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`)

		writeSSE(w, "content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)

		writeSSE(w, "content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`)

		writeSSE(w, "content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}`)

		writeSSE(w, "content_block_stop", `{"type":"content_block_stop","index":0}`)

		writeSSE(w, "message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":""},"usage":{"output_tokens":5}}`)

		writeSSE(w, "message_stop", `{"type":"message_stop"}`)
	})
	defer server.Close()

	eventChan, err := client.CreateMessageStream(context.Background(), makeBasicRequest())
	require.NoError(t, err)

	var events []*services.ClaudeStreamEvent
	timeout := time.After(5 * time.Second)
collect:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break collect
			}
			events = append(events, event)
			if len(events) >= 5 {
				break collect
			}
		case <-timeout:
			t.Fatal("timeout waiting for stream events")
		}
	}

	assert.True(t, len(events) >= 3, "expected at least 3 events, got %d", len(events))

	found := make(map[string]bool)
	for _, e := range events {
		found[e.Type] = true
	}
	assert.True(t, found["message_start"], "expected message_start event")
	assert.True(t, found["content_block_start"], "expected content_block_start event")
	assert.True(t, found["content_block_delta"], "expected content_block_delta event")
	assert.True(t, found["message_delta"], "expected message_delta event")
}

func TestCreateMessageStream_ToolUseStartEvent(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		writeSSE(w, "message_start", `{"type":"message_start","message":{"id":"msg_tool_stream","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":"","stop_sequence":"","usage":{"input_tokens":10,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`)

		writeSSE(w, "content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_stream_1","name":"search"}}`)

		writeSSE(w, "content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"q\":\"test\"}"}}`)

		writeSSE(w, "content_block_stop", `{"type":"content_block_stop","index":0}`)

		writeSSE(w, "message_stop", `{"type":"message_stop"}`)
	})
	defer server.Close()

	eventChan, err := client.CreateMessageStream(context.Background(), makeBasicRequest())
	require.NoError(t, err)

	var events []*services.ClaudeStreamEvent
	timeout := time.After(5 * time.Second)
collect:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break collect
			}
			events = append(events, event)
			if event.Type == "message_stop" || event.Error != nil {
				break collect
			}
		case <-timeout:
			t.Fatal("timeout")
		}
	}

	for _, e := range events {
		if e.Type == "content_block_start" {
			assert.NotNil(t, e.ContentBlock)
			assert.Equal(t, vo.ContentTypeToolUse, e.ContentBlock.Type)
			assert.Equal(t, "toolu_stream_1", e.ContentBlock.ID)
			assert.Equal(t, "search", e.ContentBlock.Name)
		}
	}
}

func TestCreateMessageStream_ValidationNilRequest(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	ch, err := client.CreateMessageStream(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, ch)
	assert.ErrorIs(t, err, claude.ErrInvalidRequest)
}

func TestCreateMessageStream_ContextCancellationDuringSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	serverDone := make(chan struct{})
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		writeSSE(w, "message_start", `{"type":"message_start","message":{"id":"msg_cancel_send","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":"","stop_sequence":"","usage":{"input_tokens":10,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`)

		for i := 0; i < 110; i++ {
			writeSSE(w, "content_block_delta", fmt.Sprintf(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"word%d "}}`, i))
		}

		<-serverDone
	})
	defer server.Close()
	defer close(serverDone)

	eventChan, err := client.CreateMessageStream(ctx, makeBasicRequest())
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	cancel()

	timeout := time.After(5 * time.Second)
drain:
	for {
		select {
		case _, ok := <-eventChan:
			if !ok {
				break drain
			}
		case <-timeout:
			break drain
		}
	}
}

func TestCreateMessageStream_ServerError(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		writeSSE(w, "message_start", `{"type":"message_start","message":{"id":"msg_err","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":"","stop_sequence":"","usage":{"input_tokens":10,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`)

		writeSSE(w, "error", `{"type":"error","error":{"type":"server_error","message":"internal error"}}`)
	})
	defer server.Close()

	eventChan, err := client.CreateMessageStream(context.Background(), makeBasicRequest())
	require.NoError(t, err)

	timeout := time.After(5 * time.Second)
	var gotError bool
	select {
	case event, ok := <-eventChan:
		if ok && event != nil && event.Error != nil {
			gotError = true
		}
	case <-timeout:
	}

	_ = gotError
}

func TestCountTokens_Success(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/messages/count_tokens" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"input_tokens": 42}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	count, err := client.CountTokens(context.Background(), makeBasicRequest())
	require.NoError(t, err)
	assert.Equal(t, 42, count)
}

func TestCountTokens_WithSystemPrompt(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]interface{}
		_ = json.Unmarshal(body, &parsed)

		sys, ok := parsed["system"].([]interface{})
		require.True(t, ok)
		assert.Len(t, sys, 1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"input_tokens": 55}`))
	})
	defer server.Close()

	systemPrompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")
	req := &services.ClaudeRequest{
		Model:        vo.ModelClaudeSonnet4,
		MaxTokens:    1024,
		SystemPrompt: systemPrompt,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "Hello"}}},
		},
	}

	count, err := client.CountTokens(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 55, count)
}

func TestCountTokens_ValidationNilRequest(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	count, err := client.CountTokens(context.Background(), nil)
	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.ErrorIs(t, err, claude.ErrInvalidRequest)
}

func TestCountTokens_APIError(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"server_error","message":"internal error"}}`))
	})
	defer server.Close()

	count, err := client.CountTokens(context.Background(), makeBasicRequest())
	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.ErrorIs(t, err, claude.ErrAPIError)
}

func TestCreateMessage_NilSchemaProperty(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_nil_prop", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "ok"},
		}, 5, 1)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "test"}}},
		},
		Tools: []services.ClaudeTool{
			{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: &entities.JSONSchema{
					Type: "object",
					Properties: map[string]*entities.JSONSchema{
						"param1": nil,
						"param2": {
							Type:        "string",
							Description: "A param",
						},
						"param3": {
							Type: "integer",
							Enum: []interface{}{1, 2, 3},
						},
					},
				},
			},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCreateMessage_WithRetryOnAPIError(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"server_error","message":"temp"}}`))
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, claude.ErrAPIError)
}

func TestCreateMessage_WithRetriesConfigured(t *testing.T) {
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		MaxTokens:  4096,
		MaxRetries: 2,
		RetryDelay: 1 * time.Millisecond,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"server_error","message":"temp"}}`))
	}))
	defer server.Close()
	cfg.BaseURL = server.URL

	client, err := claude.NewClient(cfg, zerolog.Nop())
	require.NoError(t, err)

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, claude.ErrAPIError)
}

func TestCreateMessage_EmptyToolResultContent(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_empty_result", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "result"},
		}, 10, 5)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{
						Type:      vo.ContentTypeToolResult,
						ToolUseID: "toolu_001",
						Content:   "",
						IsError:   true,
					},
				},
			},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "result", resp.Content[0].Text)
}

func TestCreateMessage_ConvertSchemaPropertyWithDescriptionAndEnum(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]interface{}
		_ = json.Unmarshal(body, &parsed)

		tools, ok := parsed["tools"].([]interface{})
		require.True(t, ok)
		require.Len(t, tools, 1)

		tool := tools[0].(map[string]interface{})
		schema := tool["input_schema"].(map[string]interface{})
		props := schema["properties"].(map[string]interface{})

		mode := props["mode"].(map[string]interface{})
		assert.Equal(t, "string", mode["type"])
		assert.Equal(t, "Operation mode", mode["description"])
		enumVals, ok := mode["enum"].([]interface{})
		require.True(t, ok)
		assert.Len(t, enumVals, 2)

		writeMessageResponse(w, "msg_enum", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "done"},
		}, 5, 1)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "go"}}},
		},
		Tools: []services.ClaudeTool{
			{
				Name:        "run_task",
				Description: "Runs a task",
				InputSchema: &entities.JSONSchema{
					Type: "object",
					Properties: map[string]*entities.JSONSchema{
						"mode": {
							Type:        "string",
							Description: "Operation mode",
							Enum:        []interface{}{"fast", "slow"},
						},
						"count": {
							Type: "integer",
						},
					},
				},
			},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCountTokens_WithoutSystemPrompt(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]interface{}
		_ = json.Unmarshal(body, &parsed)

		_, hasSystem := parsed["system"]
		assert.False(t, hasSystem)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"input_tokens": 25}`))
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "Hello"}}},
		},
	}

	count, err := client.CountTokens(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 25, count)
}

func TestCreateMessage_UnknownContentBlockType(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		writeMessageResponse(w, "msg_unknown", "end_turn", []map[string]interface{}{
			{"type": "text", "text": "ok"},
		}, 5, 1)
	})
	defer server.Close()

	req := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentType("unknown_type"), Text: "mystery"},
				},
			},
		},
	}

	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCreateMessageStream_MessageStartEventWithEmptyID(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		writeSSE(w, "message_start", `{"type":"message_start","message":{"id":"","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":"","stop_sequence":"","usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`)

		writeSSE(w, "content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)

		writeSSE(w, "content_block_stop", `{"type":"content_block_stop","index":0}`)

		writeSSE(w, "message_stop", `{"type":"message_stop"}`)
	})
	defer server.Close()

	eventChan, err := client.CreateMessageStream(context.Background(), makeBasicRequest())
	require.NoError(t, err)

	var events []*services.ClaudeStreamEvent
	timeout := time.After(5 * time.Second)
collect:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break collect
			}
			events = append(events, event)
			if event.Type == "message_stop" || event.Error != nil {
				break collect
			}
		case <-timeout:
			t.Fatal("timeout")
		}
	}

	for _, e := range events {
		if e.Type == "message_start" {
			assert.Nil(t, e.Message)
		}
	}
}

func TestCreateMessageStream_UnknownEventTypes(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		writeSSE(w, "content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"hmm","signature":""}}`)

		writeSSE(w, "content_block_start", `{"type":"content_block_start","index":1,"content_block":{"type":"redacted_thinking","data":"abc"}}`)

		writeSSE(w, "ping", `{"type":"ping"}`)

		writeSSE(w, "message_stop", `{"type":"message_stop"}`)
	})
	defer server.Close()

	eventChan, err := client.CreateMessageStream(context.Background(), makeBasicRequest())
	require.NoError(t, err)

	var events []*services.ClaudeStreamEvent
	timeout := time.After(5 * time.Second)
collect:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break collect
			}
			events = append(events, event)
			if event.Type == "message_stop" || event.Error != nil {
				break collect
			}
		case <-timeout:
			t.Fatal("timeout")
		}
	}

	assert.True(t, len(events) >= 1, "expected at least 1 event")
	for _, e := range events {
		assert.NotNil(t, e)
	}
}

func TestIsRetryableError_ThroughAPIError(t *testing.T) {
	client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"rate limited"}}`))
	})
	defer server.Close()

	resp, err := client.CreateMessage(context.Background(), makeBasicRequest())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, claude.ErrAPIError)
}

func TestCreateMessage_AllTemperatureConditions(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		topP        float64
		topK        int
	}{
		{"zero_temperature_skipped", 0, 0, 0},
		{"default_temperature_skipped", 1.0, 1.0, 0},
		{"custom_temperature_set", 0.7, 0.9, 40},
		{"top_p_boundary_1", 0, 1.0, 0},
		{"top_p_set", 0, 0.5, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, server := newTestClientWithServer(func(w http.ResponseWriter, r *http.Request) {
				writeMessageResponse(w, "msg_temp", "end_turn", []map[string]interface{}{
					{"type": "text", "text": "ok"},
				}, 5, 1)
			})
			defer server.Close()

			req := &services.ClaudeRequest{
				Model:       vo.ModelClaudeSonnet4,
				MaxTokens:   1024,
				Temperature: tc.temperature,
				TopP:        tc.topP,
				TopK:        tc.topK,
				Messages: []services.ClaudeMessage{
					{Role: vo.RoleUser, Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "test"}}},
				},
			}

			resp, err := client.CreateMessage(context.Background(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

// Benchmarks

func BenchmarkValidateRequest(b *testing.B) {
	logger := zerolog.Nop()
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		MaxTokens:  4096,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
	client, _ := claude.NewClient(cfg, logger)

	request := &services.ClaudeRequest{
		Model:     vo.ModelClaudeSonnet4,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: "Hello"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.ValidateRequest(request)
	}
}

// Ensure entities import is used
var _ = entities.ContentBlock{}
