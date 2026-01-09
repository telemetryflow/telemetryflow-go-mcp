// Package session_test provides integration tests for session management.
//
// TelemetryFlow MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	"github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

func TestSessionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("complete session lifecycle", func(t *testing.T) {
		// Create session
		session := aggregates.NewSession()
		require.NotNil(t, session)
		assert.Equal(t, vo.SessionStateCreated, session.State())

		// Initialize session
		clientInfo := &aggregates.ClientInfo{
			Name:    "IntegrationTestClient",
			Version: "1.0.0",
		}
		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)
		assert.Equal(t, vo.SessionStateInitializing, session.State())

		// Mark ready
		session.MarkReady()
		assert.Equal(t, vo.SessionStateReady, session.State())

		// Register tools
		tool := createIntegrationTool(t, "test_tool", "Integration test tool")
		err = session.RegisterTool(tool)
		require.NoError(t, err)
		assert.Len(t, session.Tools(), 1)

		// Register resources
		uri, _ := vo.NewResourceURI("file:///integration/test")
		resource, _ := entities.NewResource(uri, "Integration Resource")
		err = session.RegisterResource(resource)
		require.NoError(t, err)
		assert.Len(t, session.Resources(), 1)

		// Create conversation
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)
		err = session.AddConversation(conv)
		require.NoError(t, err)
		assert.Len(t, session.Conversations(), 1)

		// Close conversation
		err = conv.Close()
		require.NoError(t, err)
		assert.Equal(t, vo.ConversationStatusClosed, conv.Status())

		// Close session
		err = session.Close()
		require.NoError(t, err)
		assert.Equal(t, vo.SessionStateClosed, session.State())
	})

	t.Run("session with multiple conversations", func(t *testing.T) {
		session := createReadySession(t)

		// Create multiple conversations
		models := []vo.Model{
			vo.ModelClaude4Opus,
			vo.ModelClaude4Sonnet,
			vo.ModelClaude35Sonnet,
		}

		for _, model := range models {
			conv := aggregates.NewConversation(session.ID(), model)
			err := session.AddConversation(conv)
			require.NoError(t, err)
		}

		assert.Len(t, session.Conversations(), 3)

		// Close all conversations
		for _, conv := range session.Conversations() {
			err := conv.Close()
			require.NoError(t, err)
		}

		// Close session
		err := session.Close()
		require.NoError(t, err)
	})

	t.Run("session with tool registration and unregistration", func(t *testing.T) {
		session := createReadySession(t)

		// Register multiple tools
		tools := []string{"read_file", "write_file", "execute_command", "echo"}
		for _, name := range tools {
			tool := createIntegrationTool(t, name, "Tool: "+name)
			err := session.RegisterTool(tool)
			require.NoError(t, err)
		}
		assert.Len(t, session.Tools(), 4)

		// Unregister a tool
		err := session.UnregisterTool("echo")
		require.NoError(t, err)
		assert.Len(t, session.Tools(), 3)

		// Verify tool was removed
		assert.Nil(t, session.GetTool("echo"))
		assert.NotNil(t, session.GetTool("read_file"))
	})
}

func TestConversationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("complete conversation lifecycle", func(t *testing.T) {
		session := createReadySession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		// Add to session
		err := session.AddConversation(conv)
		require.NoError(t, err)

		// Set system prompt
		err = conv.SetSystemPrompt("You are a helpful assistant for integration testing.")
		require.NoError(t, err)

		// Add user message
		userContent := []entities.Content{entities.NewTextContent("Hello, this is an integration test.")}
		err = conv.AddMessage(vo.RoleUser, userContent)
		require.NoError(t, err)

		// Add assistant response
		assistantContent := []entities.Content{entities.NewTextContent("Hello! I'm ready to help with testing.")}
		err = conv.AddMessage(vo.RoleAssistant, assistantContent)
		require.NoError(t, err)

		// Verify messages
		assert.Equal(t, 2, conv.MessageCount())
		messages := conv.Messages()
		assert.Equal(t, vo.RoleUser, messages[0].Role())
		assert.Equal(t, vo.RoleAssistant, messages[1].Role())

		// Close conversation
		err = conv.Close()
		require.NoError(t, err)
	})

	t.Run("conversation with tool use", func(t *testing.T) {
		session := createReadySession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		// Register tool in conversation
		tool := createIntegrationTool(t, "get_weather", "Get weather for a location")
		err := conv.RegisterTool(tool)
		require.NoError(t, err)

		// Add user message requesting tool use
		userContent := []entities.Content{entities.NewTextContent("What's the weather in San Francisco?")}
		err = conv.AddMessage(vo.RoleUser, userContent)
		require.NoError(t, err)

		// Simulate assistant response with tool use
		// In real integration, this would come from Claude API

		// Close conversation
		err = conv.Close()
		require.NoError(t, err)
	})

	t.Run("conversation settings persistence", func(t *testing.T) {
		session := createReadySession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		// Set various settings
		err := conv.SetMaxTokens(2048)
		require.NoError(t, err)
		err = conv.SetTemperature(0.7)
		require.NoError(t, err)
		err = conv.SetTopP(0.95)
		require.NoError(t, err)
		err = conv.SetTopK(50)
		require.NoError(t, err)

		// Verify settings
		assert.Equal(t, 2048, conv.MaxTokens())
		assert.InDelta(t, 0.7, conv.Temperature(), 0.01)
		assert.InDelta(t, 0.95, conv.TopP(), 0.01)
		assert.Equal(t, 50, conv.TopK())
	})
}

func TestConcurrentSessionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("concurrent tool registration", func(t *testing.T) {
		session := createReadySession(t)

		// Register tools concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				name := "concurrent_tool_" + string(rune('0'+idx))
				tool := createIntegrationTool(t, name, "Concurrent tool")
				_ = session.RegisterTool(tool)
				done <- true
			}(i)
		}

		// Wait for all registrations
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify some tools were registered (may have duplicates blocked)
		assert.NotEmpty(t, session.Tools())
	})

	t.Run("concurrent message addition", func(t *testing.T) {
		session := createReadySession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		// Add messages concurrently
		done := make(chan bool, 20)
		for i := 0; i < 20; i++ {
			go func(idx int) {
				content := []entities.Content{entities.NewTextContent("Message " + string(rune('0'+idx%10)))}
				role := vo.RoleUser
				if idx%2 == 1 {
					role = vo.RoleAssistant
				}
				_ = conv.AddMessage(role, content)
				done <- true
			}(i)
		}

		// Wait for all additions
		for i := 0; i < 20; i++ {
			<-done
		}

		// Verify messages were added
		assert.NotEmpty(t, conv.Messages())
	})
}

func TestSessionTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("session operations with context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		session := createReadySession(t)

		// Operations should complete before timeout
		tool := createIntegrationTool(t, "timeout_tool", "Timeout test tool")
		err := session.RegisterTool(tool)
		require.NoError(t, err)

		// Simulate some work
		select {
		case <-ctx.Done():
			t.Log("Context timed out as expected for long operations")
		case <-time.After(50 * time.Millisecond):
			t.Log("Operations completed within timeout")
		}
	})
}

func TestSessionCapabilitiesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("verify all capabilities are set", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()

		// Tools capability
		assert.True(t, caps.Tools.ListChanged)

		// Resources capability
		assert.True(t, caps.Resources.Subscribe)
		assert.True(t, caps.Resources.ListChanged)

		// Prompts capability
		assert.True(t, caps.Prompts.ListChanged)

		// Logging capability should exist
		assert.NotNil(t, caps.Logging)
	})
}

// Helper functions

func createReadySession(t *testing.T) *aggregates.Session {
	t.Helper()
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{
		Name:    "IntegrationTest",
		Version: "1.0.0",
	}
	err := session.Initialize(clientInfo, "2024-11-05")
	require.NoError(t, err)
	session.MarkReady()
	return session
}

func createIntegrationTool(t *testing.T, name, description string) *entities.Tool {
	t.Helper()
	toolName, err := vo.NewToolName(name)
	require.NoError(t, err)
	toolDesc, err := vo.NewToolDescription(description)
	require.NoError(t, err)
	tool, err := entities.NewTool(toolName, toolDesc, nil)
	require.NoError(t, err)

	// Set a simple handler
	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		return entities.NewTextToolResult("Integration test result"), nil
	})

	return tool
}
