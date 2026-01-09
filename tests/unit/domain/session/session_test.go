// Package session_test provides unit tests for the session aggregate.
//
// TelemetryFlow MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package session_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	"github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/devopscorner/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

func TestNewSession(t *testing.T) {
	t.Run("should create session with unique ID", func(t *testing.T) {
		session := aggregates.NewSession()
		require.NotNil(t, session)
		assert.NotEmpty(t, session.ID().String())
	})

	t.Run("should create session in created state", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Equal(t, vo.SessionStateCreated, session.State())
	})

	t.Run("should set default protocol version", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Equal(t, vo.ProtocolVersion202411, session.ProtocolVersion())
	})

	t.Run("should generate unique IDs for different sessions", func(t *testing.T) {
		session1 := aggregates.NewSession()
		session2 := aggregates.NewSession()
		assert.NotEqual(t, session1.ID().String(), session2.ID().String())
	})

	t.Run("should have empty tool list initially", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Empty(t, session.Tools())
	})

	t.Run("should have empty resource list initially", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Empty(t, session.Resources())
	})

	t.Run("should have empty prompt list initially", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Empty(t, session.Prompts())
	})
}

func TestSessionInitialize(t *testing.T) {
	t.Run("should initialize session with client info", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)
		assert.Equal(t, vo.SessionStateInitializing, session.State())
	})

	t.Run("should store client info", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "2.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		storedClientInfo := session.ClientInfo()
		assert.Equal(t, "TestClient", storedClientInfo.Name)
		assert.Equal(t, "2.0.0", storedClientInfo.Version)
	})

	t.Run("should not initialize already initialized session", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		err = session.Initialize(clientInfo, "2024-11-05")
		assert.Error(t, err)
	})

	t.Run("should fail with nil client info", func(t *testing.T) {
		session := aggregates.NewSession()
		err := session.Initialize(nil, "2024-11-05")
		assert.Error(t, err)
	})

	t.Run("should fail with empty client name", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		assert.Error(t, err)
	})
}

func TestSessionMarkReady(t *testing.T) {
	t.Run("should mark session as ready after initialization", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		session.MarkReady()
		assert.Equal(t, vo.SessionStateReady, session.State())
	})

	t.Run("should not mark ready from created state", func(t *testing.T) {
		session := aggregates.NewSession()
		session.MarkReady()
		// Should remain in created state
		assert.Equal(t, vo.SessionStateCreated, session.State())
	})
}

func TestSessionClose(t *testing.T) {
	t.Run("should close session from ready state", func(t *testing.T) {
		session := createReadySession(t)

		err := session.Close()
		require.NoError(t, err)
		assert.Equal(t, vo.SessionStateClosed, session.State())
	})

	t.Run("should set closed time", func(t *testing.T) {
		session := createReadySession(t)
		beforeClose := time.Now()

		err := session.Close()
		require.NoError(t, err)

		closedAt := session.ClosedAt()
		assert.True(t, closedAt.After(beforeClose) || closedAt.Equal(beforeClose))
	})

	t.Run("should not close already closed session", func(t *testing.T) {
		session := createReadySession(t)

		err := session.Close()
		require.NoError(t, err)

		err = session.Close()
		assert.Error(t, err)
	})
}

func TestSessionToolManagement(t *testing.T) {
	t.Run("should register tool", func(t *testing.T) {
		session := createReadySession(t)
		tool := createTestTool(t, "test_tool", "Test tool description")

		err := session.RegisterTool(tool)
		require.NoError(t, err)
		assert.Len(t, session.Tools(), 1)
	})

	t.Run("should get registered tool by name", func(t *testing.T) {
		session := createReadySession(t)
		tool := createTestTool(t, "my_tool", "My tool description")

		err := session.RegisterTool(tool)
		require.NoError(t, err)

		retrievedTool := session.GetTool("my_tool")
		assert.NotNil(t, retrievedTool)
		assert.Equal(t, "my_tool", retrievedTool.Name().String())
	})

	t.Run("should return nil for non-existent tool", func(t *testing.T) {
		session := createReadySession(t)
		tool := session.GetTool("non_existent")
		assert.Nil(t, tool)
	})

	t.Run("should not register duplicate tool", func(t *testing.T) {
		session := createReadySession(t)
		tool1 := createTestTool(t, "duplicate_tool", "First description")
		tool2 := createTestTool(t, "duplicate_tool", "Second description")

		err := session.RegisterTool(tool1)
		require.NoError(t, err)

		err = session.RegisterTool(tool2)
		assert.Error(t, err)
	})

	t.Run("should register multiple tools", func(t *testing.T) {
		session := createReadySession(t)

		for i := 0; i < 5; i++ {
			tool := createTestTool(t, "tool_"+string(rune('a'+i)), "Description "+string(rune('a'+i)))
			err := session.RegisterTool(tool)
			require.NoError(t, err)
		}

		assert.Len(t, session.Tools(), 5)
	})

	t.Run("should unregister tool", func(t *testing.T) {
		session := createReadySession(t)
		tool := createTestTool(t, "removable_tool", "To be removed")

		err := session.RegisterTool(tool)
		require.NoError(t, err)
		assert.Len(t, session.Tools(), 1)

		err = session.UnregisterTool("removable_tool")
		require.NoError(t, err)
		assert.Empty(t, session.Tools())
	})
}

func TestSessionResourceManagement(t *testing.T) {
	t.Run("should register resource", func(t *testing.T) {
		session := createReadySession(t)
		uri, _ := vo.NewResourceURI("file:///test/path")
		resource, _ := entities.NewResource(uri, "Test Resource")

		err := session.RegisterResource(resource)
		require.NoError(t, err)
		assert.Len(t, session.Resources(), 1)
	})

	t.Run("should get registered resource by URI", func(t *testing.T) {
		session := createReadySession(t)
		uri, _ := vo.NewResourceURI("file:///my/resource")
		resource, _ := entities.NewResource(uri, "My Resource")

		err := session.RegisterResource(resource)
		require.NoError(t, err)

		retrievedResource := session.GetResource("file:///my/resource")
		assert.NotNil(t, retrievedResource)
	})

	t.Run("should return nil for non-existent resource", func(t *testing.T) {
		session := createReadySession(t)
		resource := session.GetResource("file:///non/existent")
		assert.Nil(t, resource)
	})
}

func TestSessionPromptManagement(t *testing.T) {
	t.Run("should register prompt", func(t *testing.T) {
		session := createReadySession(t)
		promptName, _ := vo.NewToolName("test_prompt")
		prompt := entities.NewPrompt(promptName, "Test prompt description")

		err := session.RegisterPrompt(prompt)
		require.NoError(t, err)
		assert.Len(t, session.Prompts(), 1)
	})

	t.Run("should get registered prompt by name", func(t *testing.T) {
		session := createReadySession(t)
		promptName, _ := vo.NewToolName("my_prompt")
		prompt := entities.NewPrompt(promptName, "My prompt description")

		err := session.RegisterPrompt(prompt)
		require.NoError(t, err)

		retrievedPrompt := session.GetPrompt("my_prompt")
		assert.NotNil(t, retrievedPrompt)
	})

	t.Run("should return nil for non-existent prompt", func(t *testing.T) {
		session := createReadySession(t)
		prompt := session.GetPrompt("non_existent")
		assert.Nil(t, prompt)
	})
}

func TestSessionConversations(t *testing.T) {
	t.Run("should add conversation", func(t *testing.T) {
		session := createReadySession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		err := session.AddConversation(conv)
		require.NoError(t, err)
		assert.Len(t, session.Conversations(), 1)
	})

	t.Run("should get conversation by ID", func(t *testing.T) {
		session := createReadySession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		err := session.AddConversation(conv)
		require.NoError(t, err)

		retrievedConv := session.GetConversation(conv.ID())
		assert.NotNil(t, retrievedConv)
		assert.Equal(t, conv.ID(), retrievedConv.ID())
	})

	t.Run("should return nil for non-existent conversation", func(t *testing.T) {
		session := createReadySession(t)
		nonExistentID, _ := vo.NewConversationID()
		conv := session.GetConversation(nonExistentID)
		assert.Nil(t, conv)
	})
}

func TestSessionCapabilities(t *testing.T) {
	t.Run("should have default capabilities", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.NotNil(t, caps)
	})

	t.Run("should enable tools capability", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.True(t, caps.Tools.ListChanged)
	})

	t.Run("should enable resources capability", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.True(t, caps.Resources.Subscribe)
		assert.True(t, caps.Resources.ListChanged)
	})

	t.Run("should enable prompts capability", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.True(t, caps.Prompts.ListChanged)
	})
}

func TestSessionServerInfo(t *testing.T) {
	t.Run("should return server info", func(t *testing.T) {
		session := aggregates.NewSession()
		info := session.ServerInfo()
		assert.NotEmpty(t, info.Name)
		assert.NotEmpty(t, info.Version)
	})
}

func TestSessionCreatedAt(t *testing.T) {
	t.Run("should set created time", func(t *testing.T) {
		beforeCreate := time.Now()
		session := aggregates.NewSession()
		afterCreate := time.Now()

		createdAt := session.CreatedAt()
		assert.True(t, createdAt.After(beforeCreate) || createdAt.Equal(beforeCreate))
		assert.True(t, createdAt.Before(afterCreate) || createdAt.Equal(afterCreate))
	})
}

// Helper functions

func createReadySession(t *testing.T) *aggregates.Session {
	t.Helper()
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{
		Name:    "TestClient",
		Version: "1.0.0",
	}
	err := session.Initialize(clientInfo, "2024-11-05")
	require.NoError(t, err)
	session.MarkReady()
	return session
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

func BenchmarkNewSession(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = aggregates.NewSession()
	}
}

func BenchmarkSessionInitialize(b *testing.B) {
	clientInfo := &aggregates.ClientInfo{
		Name:    "BenchClient",
		Version: "1.0.0",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := aggregates.NewSession()
		_ = session.Initialize(clientInfo, "2024-11-05")
	}
}

func BenchmarkSessionRegisterTool(b *testing.B) {
	toolName, _ := vo.NewToolName("bench_tool")
	toolDesc, _ := vo.NewToolDescription("Benchmark tool")
	tool, _ := entities.NewTool(toolName, toolDesc, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{Name: "Bench", Version: "1.0.0"}
		_ = session.Initialize(clientInfo, "2024-11-05")
		session.MarkReady()
		_ = session.RegisterTool(tool)
	}
}

func BenchmarkSessionGetTool(b *testing.B) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Bench", Version: "1.0.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	// Register 100 tools
	for i := 0; i < 100; i++ {
		name := "bench_tool_" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10))
		toolName, _ := vo.NewToolName(name)
		toolDesc, _ := vo.NewToolDescription("Benchmark tool")
		tool, _ := entities.NewTool(toolName, toolDesc, nil)
		_ = session.RegisterTool(tool)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = session.GetTool("bench_tool_50")
	}
}
