package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/pkg/mcp"
)

func TestNewTextContent(t *testing.T) {
	c := mcp.NewTextContent("hello")
	assert.Equal(t, "text", c.Type)
	assert.Equal(t, "hello", c.Text)
}

func TestNewImageContent(t *testing.T) {
	c := mcp.NewImageContent("base64data", "image/png")
	assert.Equal(t, "image", c.Type)
	assert.Equal(t, "base64data", c.Data)
	assert.Equal(t, "image/png", c.MimeType)
}

func TestNewResourceContent(t *testing.T) {
	c := mcp.NewResourceContent(&mcp.EmbeddedResource{URI: "file:///test", Text: "content"})
	assert.Equal(t, "resource", c.Type)
	assert.Equal(t, "file:///test", c.Resource.URI)
}

func TestInitializeParams(t *testing.T) {
	p := mcp.InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    mcp.ClientCapability{Sampling: &mcp.SamplingCapability{}},
		ClientInfo:      mcp.ClientInfo{Name: "test", Version: "1.0"},
	}
	data, err := json.Marshal(p)
	require.NoError(t, err)
	assert.Contains(t, string(data), "2024-11-05")
}

func TestInitializeResult(t *testing.T) {
	r := mcp.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    mcp.ServerCapability{Tools: &mcp.ToolsCapability{ListChanged: true}},
		ServerInfo:      mcp.ServerInfo{Name: "Test", Version: "1.0"},
	}
	assert.Equal(t, "2024-11-05", r.ProtocolVersion)
}

func TestToolType(t *testing.T) {
	tool := mcp.Tool{
		Name:        "test",
		Description: "A test tool",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}
	data, err := json.Marshal(tool)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test")
}

func TestToolListResult(t *testing.T) {
	cursor := "next"
	r := mcp.ToolListResult{
		Tools:      []mcp.Tool{{Name: "t1"}},
		NextCursor: &cursor,
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	assert.Contains(t, string(data), "nextCursor")
}

func TestCallToolParams(t *testing.T) {
	p := mcp.CallToolParams{
		Name:      "my_tool",
		Arguments: map[string]interface{}{"key": "val"},
	}
	assert.Equal(t, "my_tool", p.Name)
}

func TestCallToolResult(t *testing.T) {
	r := mcp.CallToolResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent("result")},
		IsError: false,
	}
	assert.False(t, r.IsError)
	assert.Len(t, r.Content, 1)
}

func TestResourceType(t *testing.T) {
	r := mcp.Resource{URI: "file:///test", Name: "Test", MimeType: "text/plain"}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	assert.Contains(t, string(data), "file:///test")
}

func TestResourceReadParams(t *testing.T) {
	p := mcp.ResourceReadParams{URI: "file:///test"}
	assert.Equal(t, "file:///test", p.URI)
}

func TestResourceReadResult(t *testing.T) {
	r := mcp.ResourceReadResult{
		Contents: []mcp.ResourceContent{{URI: "file:///test", Text: "hello"}},
	}
	assert.Len(t, r.Contents, 1)
}

func TestResourceTemplate(t *testing.T) {
	rt := mcp.ResourceTemplate{URITemplate: "file:///{path}", Name: "File"}
	data, err := json.Marshal(rt)
	require.NoError(t, err)
	assert.Contains(t, string(data), "file:///{path}")
}

func TestPromptType(t *testing.T) {
	p := mcp.Prompt{
		Name:        "test_prompt",
		Description: "Test",
		Arguments:   []mcp.PromptArgument{{Name: "input", Required: true}},
	}
	data, err := json.Marshal(p)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test_prompt")
}

func TestGetPromptParams(t *testing.T) {
	p := mcp.GetPromptParams{Name: "prompt", Arguments: map[string]interface{}{"k": "v"}}
	assert.Equal(t, "prompt", p.Name)
}

func TestGetPromptResult(t *testing.T) {
	r := mcp.GetPromptResult{
		Description: "desc",
		Messages:    []mcp.PromptMessage{{Role: "user", Content: mcp.NewTextContent("hi")}},
	}
	assert.Len(t, r.Messages, 1)
}

func TestSetLevelParams(t *testing.T) {
	p := mcp.SetLevelParams{Level: mcp.LogLevelDebug}
	assert.Equal(t, mcp.LogLevelDebug, p.Level)
}

func TestLogMessageParams(t *testing.T) {
	p := mcp.LogMessageParams{Level: mcp.LogLevelInfo, Logger: "test", Data: "msg"}
	assert.Equal(t, "test", p.Logger)
}

func TestSamplingMessage(t *testing.T) {
	sm := mcp.SamplingMessage{Role: "user", Content: mcp.NewTextContent("hi")}
	assert.Equal(t, "user", sm.Role)
}

func TestCreateMessageParams(t *testing.T) {
	p := mcp.CreateMessageParams{
		Messages:  []mcp.SamplingMessage{{Role: "user", Content: mcp.NewTextContent("hi")}},
		MaxTokens: 100,
	}
	assert.Equal(t, 100, p.MaxTokens)
}

func TestModelPreferences(t *testing.T) {
	mp := mcp.ModelPreferences{
		Hints:                []mcp.ModelHint{{Name: "claude-opus-4-7"}},
		CostPriority:         float64Ptr(0.5),
		SpeedPriority:        float64Ptr(0.8),
		IntelligencePriority: float64Ptr(0.9),
	}
	assert.Len(t, mp.Hints, 1)
}

func TestCreateMessageResult(t *testing.T) {
	r := mcp.CreateMessageResult{Role: "assistant", Model: "claude-opus-4-7", StopReason: "end_turn"}
	assert.Equal(t, "assistant", r.Role)
}

func TestProgressParams(t *testing.T) {
	p := mcp.ProgressParams{ProgressToken: "tok", Progress: 0.5, Total: 1.0}
	assert.Equal(t, "tok", p.ProgressToken)
}

func TestCancelledParams(t *testing.T) {
	p := mcp.CancelledParams{RequestID: 42, Reason: "timeout"}
	assert.Equal(t, 42, p.RequestID)
}

func TestContentBlock(t *testing.T) {
	cb := mcp.ContentBlock{
		Type:        "text",
		Text:        "hello",
		Annotations: map[string]interface{}{"audience": "user"},
	}
	data, err := json.Marshal(cb)
	require.NoError(t, err)
	assert.Contains(t, string(data), "hello")
}

func TestEmbeddedResource(t *testing.T) {
	er := mcp.EmbeddedResource{URI: "file:///test", MimeType: "text/plain", Text: "content"}
	data, err := json.Marshal(er)
	require.NoError(t, err)
	assert.Contains(t, string(data), "file:///test")
}

func TestPaginatedParams(t *testing.T) {
	p := mcp.PaginatedParams{Cursor: "abc"}
	assert.Equal(t, "abc", p.Cursor)
}

func float64Ptr(v float64) *float64 {
	return &v
}
