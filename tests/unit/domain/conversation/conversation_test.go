// Package conversation_test provides unit tests for the conversation aggregate.
//
// TelemetryFlow MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package conversation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	"github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

func TestNewConversation(t *testing.T) {
	t.Run("should create conversation with unique ID", func(t *testing.T) {
		sessionID, _ := vo.NewSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		require.NotNil(t, conv)
		assert.NotEmpty(t, conv.ID().String())
	})

	t.Run("should create conversation with specified session ID", func(t *testing.T) {
		sessionID, _ := vo.NewSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.Equal(t, sessionID, conv.SessionID())
	})

	t.Run("should create conversation with specified model", func(t *testing.T) {
		sessionID, _ := vo.NewSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Opus)
		assert.Equal(t, vo.ModelClaude4Opus, conv.Model())
	})

	t.Run("should create conversation in active status", func(t *testing.T) {
		sessionID, _ := vo.NewSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.Equal(t, vo.ConversationStatusActive, conv.Status())
	})

	t.Run("should have empty message list initially", func(t *testing.T) {
		sessionID, _ := vo.NewSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.Empty(t, conv.Messages())
	})

	t.Run("should generate unique IDs for different conversations", func(t *testing.T) {
		sessionID, _ := vo.NewSessionID()
		conv1 := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		conv2 := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.NotEqual(t, conv1.ID().String(), conv2.ID().String())
	})
}

func TestConversationMessages(t *testing.T) {
	t.Run("should add user message", func(t *testing.T) {
		conv := createTestConversation(t)
		content := createTextContent("Hello, Claude!")

		err := conv.AddMessage(vo.RoleUser, content)
		require.NoError(t, err)
		assert.Len(t, conv.Messages(), 1)
	})

	t.Run("should add assistant message", func(t *testing.T) {
		conv := createTestConversation(t)
		content := createTextContent("Hello! How can I help you?")

		err := conv.AddMessage(vo.RoleAssistant, content)
		require.NoError(t, err)
		assert.Len(t, conv.Messages(), 1)
	})

	t.Run("should add multiple messages in order", func(t *testing.T) {
		conv := createTestConversation(t)

		err := conv.AddMessage(vo.RoleUser, createTextContent("First"))
		require.NoError(t, err)
		err = conv.AddMessage(vo.RoleAssistant, createTextContent("Second"))
		require.NoError(t, err)
		err = conv.AddMessage(vo.RoleUser, createTextContent("Third"))
		require.NoError(t, err)

		messages := conv.Messages()
		assert.Len(t, messages, 3)
		assert.Equal(t, vo.RoleUser, messages[0].Role())
		assert.Equal(t, vo.RoleAssistant, messages[1].Role())
		assert.Equal(t, vo.RoleUser, messages[2].Role())
	})

	t.Run("should not add message to closed conversation", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.Close()
		require.NoError(t, err)

		err = conv.AddMessage(vo.RoleUser, createTextContent("Should fail"))
		assert.Error(t, err)
	})

	t.Run("should track message count", func(t *testing.T) {
		conv := createTestConversation(t)

		for i := 0; i < 10; i++ {
			_ = conv.AddMessage(vo.RoleUser, createTextContent("Message"))
		}

		assert.Equal(t, 10, conv.MessageCount())
	})

	t.Run("should get last message", func(t *testing.T) {
		conv := createTestConversation(t)

		_ = conv.AddMessage(vo.RoleUser, createTextContent("First"))
		_ = conv.AddMessage(vo.RoleAssistant, createTextContent("Last"))

		lastMsg := conv.LastMessage()
		require.NotNil(t, lastMsg)
		assert.Equal(t, vo.RoleAssistant, lastMsg.Role())
	})

	t.Run("should return nil for last message on empty conversation", func(t *testing.T) {
		conv := createTestConversation(t)
		lastMsg := conv.LastMessage()
		assert.Nil(t, lastMsg)
	})
}

func TestConversationSystemPrompt(t *testing.T) {
	t.Run("should set system prompt", func(t *testing.T) {
		conv := createTestConversation(t)
		systemPrompt := "You are a helpful assistant."

		err := conv.SetSystemPrompt(systemPrompt)
		require.NoError(t, err)
		assert.Equal(t, systemPrompt, conv.SystemPrompt())
	})

	t.Run("should update system prompt", func(t *testing.T) {
		conv := createTestConversation(t)

		_ = conv.SetSystemPrompt("First prompt")
		err := conv.SetSystemPrompt("Updated prompt")
		require.NoError(t, err)
		assert.Equal(t, "Updated prompt", conv.SystemPrompt())
	})

	t.Run("should not set system prompt on closed conversation", func(t *testing.T) {
		conv := createTestConversation(t)
		_ = conv.Close()

		err := conv.SetSystemPrompt("Should fail")
		assert.Error(t, err)
	})
}

func TestConversationSettings(t *testing.T) {
	t.Run("should have default max tokens", func(t *testing.T) {
		conv := createTestConversation(t)
		assert.Greater(t, conv.MaxTokens(), 0)
	})

	t.Run("should set max tokens", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.SetMaxTokens(2048)
		require.NoError(t, err)
		assert.Equal(t, 2048, conv.MaxTokens())
	})

	t.Run("should not set invalid max tokens", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.SetMaxTokens(-100)
		assert.Error(t, err)
	})

	t.Run("should have default temperature", func(t *testing.T) {
		conv := createTestConversation(t)
		assert.InDelta(t, 1.0, conv.Temperature(), 0.01)
	})

	t.Run("should set temperature", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.SetTemperature(0.7)
		require.NoError(t, err)
		assert.InDelta(t, 0.7, conv.Temperature(), 0.01)
	})

	t.Run("should not set temperature out of range", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.SetTemperature(2.5)
		assert.Error(t, err)
	})

	t.Run("should set top P", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.SetTopP(0.9)
		require.NoError(t, err)
		assert.InDelta(t, 0.9, conv.TopP(), 0.01)
	})

	t.Run("should set top K", func(t *testing.T) {
		conv := createTestConversation(t)
		err := conv.SetTopK(40)
		require.NoError(t, err)
		assert.Equal(t, 40, conv.TopK())
	})
}

func TestConversationClose(t *testing.T) {
	t.Run("should close active conversation", func(t *testing.T) {
		conv := createTestConversation(t)

		err := conv.Close()
		require.NoError(t, err)
		assert.Equal(t, vo.ConversationStatusClosed, conv.Status())
	})

	t.Run("should set closed time", func(t *testing.T) {
		conv := createTestConversation(t)
		beforeClose := time.Now()

		err := conv.Close()
		require.NoError(t, err)

		closedAt := conv.ClosedAt()
		require.NotNil(t, closedAt)
		assert.True(t, closedAt.After(beforeClose) || closedAt.Equal(beforeClose))
	})

	t.Run("should not close already closed conversation", func(t *testing.T) {
		conv := createTestConversation(t)

		err := conv.Close()
		require.NoError(t, err)

		err = conv.Close()
		assert.Error(t, err)
	})

	t.Run("should preserve messages after close", func(t *testing.T) {
		conv := createTestConversation(t)
		_ = conv.AddMessage(vo.RoleUser, createTextContent("Test message"))

		err := conv.Close()
		require.NoError(t, err)

		assert.Len(t, conv.Messages(), 1)
	})
}

func TestConversationTools(t *testing.T) {
	t.Run("should register available tool", func(t *testing.T) {
		conv := createTestConversation(t)
		tool := createTestTool(t, "test_tool", "Test description")

		err := conv.RegisterTool(tool)
		require.NoError(t, err)
		assert.Len(t, conv.AvailableTools(), 1)
	})

	t.Run("should get registered tool by name", func(t *testing.T) {
		conv := createTestConversation(t)
		tool := createTestTool(t, "my_tool", "My description")

		err := conv.RegisterTool(tool)
		require.NoError(t, err)

		retrievedTool := conv.GetTool("my_tool")
		assert.NotNil(t, retrievedTool)
	})

	t.Run("should return nil for non-existent tool", func(t *testing.T) {
		conv := createTestConversation(t)
		tool := conv.GetTool("non_existent")
		assert.Nil(t, tool)
	})

	t.Run("should not register duplicate tool", func(t *testing.T) {
		conv := createTestConversation(t)
		tool1 := createTestTool(t, "dup_tool", "First")
		tool2 := createTestTool(t, "dup_tool", "Second")

		err := conv.RegisterTool(tool1)
		require.NoError(t, err)

		err = conv.RegisterTool(tool2)
		assert.Error(t, err)
	})
}

func TestConversationMetadata(t *testing.T) {
	t.Run("should set metadata", func(t *testing.T) {
		conv := createTestConversation(t)
		metadata := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		conv.SetMetadata(metadata)
		assert.Equal(t, metadata, conv.Metadata())
	})

	t.Run("should get metadata value", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetMetadata(map[string]interface{}{
			"test_key": "test_value",
		})

		value := conv.GetMetadataValue("test_key")
		assert.Equal(t, "test_value", value)
	})

	t.Run("should return nil for non-existent metadata key", func(t *testing.T) {
		conv := createTestConversation(t)
		value := conv.GetMetadataValue("non_existent")
		assert.Nil(t, value)
	})
}

func TestConversationCreatedAt(t *testing.T) {
	t.Run("should set created time", func(t *testing.T) {
		beforeCreate := time.Now()
		conv := createTestConversation(t)
		afterCreate := time.Now()

		createdAt := conv.CreatedAt()
		assert.True(t, createdAt.After(beforeCreate) || createdAt.Equal(beforeCreate))
		assert.True(t, createdAt.Before(afterCreate) || createdAt.Equal(afterCreate))
	})
}

func TestConversationModels(t *testing.T) {
	models := []vo.Model{
		vo.ModelClaude4Opus,
		vo.ModelClaude4Sonnet,
		vo.ModelClaude37Sonnet,
		vo.ModelClaude35Sonnet,
		vo.ModelClaude35Haiku,
	}

	for _, model := range models {
		t.Run("should create conversation with model "+string(model), func(t *testing.T) {
			sessionID, _ := vo.NewSessionID()
			conv := aggregates.NewConversation(sessionID, model)
			assert.Equal(t, model, conv.Model())
		})
	}
}

func TestConversationTokenTracking(t *testing.T) {
	t.Run("should track total tokens", func(t *testing.T) {
		conv := createTestConversation(t)

		// Add messages with token counts
		msg := entities.NewMessage(vo.RoleUser, createTextContent("Hello"))
		msg.SetTokenCount(5)
		conv.AddMessageEntity(msg)

		msg2 := entities.NewMessage(vo.RoleAssistant, createTextContent("Hi there!"))
		msg2.SetTokenCount(3)
		conv.AddMessageEntity(msg2)

		assert.Equal(t, 8, conv.TotalTokens())
	})
}

// Helper functions

func createTestConversation(t *testing.T) *aggregates.Conversation {
	t.Helper()
	sessionID, err := vo.NewSessionID()
	require.NoError(t, err)
	return aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
}

func createTextContent(text string) []entities.Content {
	return []entities.Content{
		entities.NewTextContent(text),
	}
}

func createTestTool(t *testing.T, name, description string) *entities.Tool {
	t.Helper()
	toolName, err := vo.NewToolName(name)
	require.NoError(t, err)
	toolDesc, err := vo.NewToolDescription(description)
	require.NoError(t, err)
	tool, err := entities.NewTool(toolName, toolDesc, nil)
	require.NoError(t, err)
	return tool
}

// Benchmarks

func BenchmarkNewConversation(b *testing.B) {
	sessionID, _ := vo.NewSessionID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
	}
}

func BenchmarkConversationAddMessage(b *testing.B) {
	sessionID, _ := vo.NewSessionID()
	conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
	content := createTextContent("Benchmark message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.AddMessage(vo.RoleUser, content)
	}
}

func BenchmarkConversationGetMessages(b *testing.B) {
	sessionID, _ := vo.NewSessionID()
	conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)

	// Add 100 messages
	for i := 0; i < 100; i++ {
		content := createTextContent("Message")
		_ = conv.AddMessage(vo.RoleUser, content)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.Messages()
	}
}
