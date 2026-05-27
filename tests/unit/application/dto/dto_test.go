package dto_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/dto"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestSessionDTO(t *testing.T) {
	now := time.Now()
	d := dto.SessionDTO{
		ID:            "session-123",
		State:         "ready",
		ClientName:    "TestClient",
		ClientVersion: "1.0.0",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	assert.Equal(t, "session-123", d.ID)
	assert.Equal(t, "ready", d.State)
	data, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Contains(t, string(data), "session-123")
}

func TestConversationDTO(t *testing.T) {
	now := time.Now()
	d := dto.ConversationDTO{
		ID:        "conv-123",
		SessionID: "sess-123",
		Model:     "claude-opus-4-7",
		IsActive:  true,
		Messages:  []dto.MessageDTO{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	assert.True(t, d.IsActive)
	assert.Empty(t, d.Messages)
}

func TestMessageDTO(t *testing.T) {
	d := dto.MessageDTO{
		ID:      "msg-123",
		Role:    "user",
		Content: []dto.ContentBlockDTO{{Type: "text", Text: "hello"}},
	}
	assert.Equal(t, "user", d.Role)
	assert.Len(t, d.Content, 1)
}

func TestToolDTO(t *testing.T) {
	d := dto.ToolDTO{
		Name:        "my_tool",
		Description: "A tool",
		InputSchema: map[string]interface{}{"type": "object"},
	}
	assert.Equal(t, "my_tool", d.Name)
}

func TestResourceDTO(t *testing.T) {
	d := dto.ResourceDTO{
		URI:         "file:///test",
		Name:        "Test",
		Description: "A resource",
		MimeType:    "text/plain",
	}
	assert.Equal(t, "file:///test", d.URI)
}

func TestPromptDTO(t *testing.T) {
	d := dto.PromptDTO{
		Name:        "test_prompt",
		Description: "A prompt",
		Arguments: []dto.ArgumentDTO{
			{Name: "arg1", Description: "first", Required: true},
		},
	}
	assert.Len(t, d.Arguments, 1)
	assert.True(t, d.Arguments[0].Required)
}

func TestInitializeRequestDTO(t *testing.T) {
	req := dto.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: dto.ClientCapabilitiesDTO{
			Experimental: map[string]interface{}{},
			Sampling:     &dto.SamplingCapabilityDTO{},
			Roots:        &dto.RootsCapabilityDTO{ListChanged: true},
		},
		ClientInfo: dto.ClientInfoDTO{Name: "test", Version: "1.0"},
	}
	assert.Equal(t, "2024-11-05", req.ProtocolVersion)
	assert.True(t, req.Capabilities.Roots.ListChanged)
}

func TestInitializeResultDTO(t *testing.T) {
	res := dto.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: dto.ServerCapabilitiesDTO{
			Tools:     &dto.ToolsCapabilityDTO{ListChanged: true},
			Resources: &dto.ResourcesCapabilityDTO{Subscribe: true},
			Prompts:   &dto.PromptsCapabilityDTO{},
			Logging:   &dto.LoggingCapabilityDTO{},
		},
		ServerInfo: dto.ServerInfoDTO{Name: "TestServer", Version: "1.0"},
	}
	assert.Equal(t, "TestServer", res.ServerInfo.Name)
}

func TestCallToolRequestDTO(t *testing.T) {
	req := dto.CallToolRequest{
		Name:      "my_tool",
		Arguments: map[string]interface{}{"key": "val"},
	}
	assert.Equal(t, "my_tool", req.Name)
}

func TestCallToolResultDTO(t *testing.T) {
	res := dto.CallToolResult{
		Content: []dto.ContentBlockDTO{{Type: "text", Text: "result"}},
		IsError: false,
	}
	assert.False(t, res.IsError)
}

func TestToModel(t *testing.T) {
	m := dto.ToModel("claude-opus-4-7")
	assert.Equal(t, vo.Model("claude-opus-4-7"), m)
}

func TestToRole(t *testing.T) {
	r := dto.ToRole("user")
	assert.Equal(t, vo.Role("user"), r)
}

func TestDTOJSONSerialization(t *testing.T) {
	t.Run("SessionDTO marshals and unmarshals", func(t *testing.T) {
		original := dto.SessionDTO{
			ID: "test", State: "ready", ClientName: "c",
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)
		var decoded dto.SessionDTO
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, original.ID, decoded.ID)
		assert.Equal(t, original.State, decoded.State)
	})

	t.Run("ToolDTO marshals with input schema", func(t *testing.T) {
		original := dto.ToolDTO{
			Name:        "tool1",
			Description: "desc",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)
		assert.Contains(t, string(data), "tool1")
	})
}
