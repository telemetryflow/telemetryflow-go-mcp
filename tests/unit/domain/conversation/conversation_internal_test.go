package conversation_test

import (
	"sync"
	"testing"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewConversation_Internal(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	model := vo.ModelClaudeSonnet4

	conv := aggregates.NewConversation(sessionID, model)

	if conv == nil {
		t.Fatal("NewConversation() returned nil")
	}

	if conv.ID().IsEmpty() {
		t.Error("Conversation ID should not be empty")
	}

	if !conv.SessionID().Equals(sessionID) {
		t.Error("SessionID should match")
	}

	if conv.Model() != model {
		t.Errorf("Expected model %s, got %s", model, conv.Model())
	}

	if conv.Status() != aggregates.ConversationStatusActive {
		t.Errorf("Expected status %s, got %s", aggregates.ConversationStatusActive, conv.Status())
	}

	if !conv.IsActive() {
		t.Error("IsActive() should return true")
	}

	if conv.MessageCount() != 0 {
		t.Errorf("Expected 0 messages, got %d", conv.MessageCount())
	}

	if conv.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestConversation_SetModel_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	err := conv.SetModel(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("SetModel() failed: %v", err)
	}

	if conv.Model() != vo.ModelClaudeOpus47 {
		t.Errorf("Expected model %s, got %s", vo.ModelClaudeOpus47, conv.Model())
	}

	err = conv.SetModel(vo.Model("invalid-model"))
	if err == nil {
		t.Error("Expected error for invalid model")
	}
}

func TestConversation_SystemPrompt_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	prompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")
	err := conv.SetSystemPrompt(prompt)
	if err != nil {
		t.Fatalf("SetSystemPrompt() failed: %v", err)
	}

	if conv.SystemPrompt().String() != "You are a helpful assistant." {
		t.Errorf("Expected system prompt 'You are a helpful assistant.', got '%s'", conv.SystemPrompt().String())
	}
}

func TestConversation_AddUserMessage_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	msg, err := conv.AddUserMessage("Hello, Claude!")
	if err != nil {
		t.Fatalf("AddUserMessage() failed: %v", err)
	}

	if msg == nil {
		t.Fatal("Message should not be nil")
	}

	if conv.MessageCount() != 1 {
		t.Errorf("Expected 1 message, got %d", conv.MessageCount())
	}

	if conv.LastMessage() != msg {
		t.Error("LastMessage should be the added message")
	}

	messages := conv.Messages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message in Messages(), got %d", len(messages))
	}
}

func TestConversation_SystemPromptImmutableAfterUserMessage_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	_, _ = conv.AddUserMessage("Hello!")

	prompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")
	err := conv.SetSystemPrompt(prompt)
	if err != aggregates.ErrSystemPromptImmutable {
		t.Errorf("Expected ErrSystemPromptImmutable, got %v", err)
	}
}

func TestConversation_AddMessageToClosedConversation_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)
	conv.Close()

	_, err := conv.AddUserMessage("Hello!")
	if err != aggregates.ErrConversationClosed {
		t.Errorf("Expected ErrConversationClosed, got %v", err)
	}
}

func TestConversation_PauseAndResume_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	if !conv.IsActive() {
		t.Error("Conversation should be active initially")
	}

	conv.Pause()
	if conv.Status() != aggregates.ConversationStatusPaused {
		t.Errorf("Expected status %s, got %s", aggregates.ConversationStatusPaused, conv.Status())
	}
	if conv.IsActive() {
		t.Error("IsActive() should return false when paused")
	}

	conv.Resume()
	if conv.Status() != aggregates.ConversationStatusActive {
		t.Errorf("Expected status %s, got %s", aggregates.ConversationStatusActive, conv.Status())
	}
	if !conv.IsActive() {
		t.Error("IsActive() should return true after resume")
	}
}

func TestConversation_Close_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	conv.Close()

	if conv.Status() != aggregates.ConversationStatusClosed {
		t.Errorf("Expected status %s, got %s", aggregates.ConversationStatusClosed, conv.Status())
	}

	if conv.ClosedAt() == nil {
		t.Error("ClosedAt should not be nil after closing")
	}

	conv.Close()
	if conv.Status() != aggregates.ConversationStatusClosed {
		t.Error("Conversation should remain closed")
	}
}

func TestConversation_Archive_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	conv.Archive()
	if conv.Status() == aggregates.ConversationStatusArchived {
		t.Error("Should not be able to archive active conversation")
	}

	conv.Close()
	conv.Archive()
	if conv.Status() != aggregates.ConversationStatusArchived {
		t.Errorf("Expected status %s, got %s", aggregates.ConversationStatusArchived, conv.Status())
	}
}

func TestConversation_MaxTokens_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	if conv.MaxTokens() != 4096 {
		t.Errorf("Expected default max tokens 4096, got %d", conv.MaxTokens())
	}

	conv.SetMaxTokens(8192)
	if conv.MaxTokens() != 8192 {
		t.Errorf("Expected max tokens 8192, got %d", conv.MaxTokens())
	}
}

func TestConversation_Temperature_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	if conv.Temperature() != 1.0 {
		t.Errorf("Expected default temperature 1.0, got %f", conv.Temperature())
	}

	conv.SetTemperature(0.5)
	if conv.Temperature() != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", conv.Temperature())
	}

	conv.SetTemperature(-1.0)
	if conv.Temperature() != 0 {
		t.Errorf("Expected temperature 0 (clamped), got %f", conv.Temperature())
	}

	conv.SetTemperature(3.0)
	if conv.Temperature() != 2.0 {
		t.Errorf("Expected temperature 2.0 (clamped), got %f", conv.Temperature())
	}
}

func TestConversation_TopP_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	if conv.TopP() != 1.0 {
		t.Errorf("Expected default top_p 1.0, got %f", conv.TopP())
	}

	conv.SetTopP(0.9)
	if conv.TopP() != 0.9 {
		t.Errorf("Expected top_p 0.9, got %f", conv.TopP())
	}

	conv.SetTopP(-0.5)
	if conv.TopP() != 0 {
		t.Errorf("Expected top_p 0 (clamped), got %f", conv.TopP())
	}

	conv.SetTopP(1.5)
	if conv.TopP() != 1.0 {
		t.Errorf("Expected top_p 1.0 (clamped), got %f", conv.TopP())
	}
}

func TestConversation_TopK_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	if conv.TopK() != 0 {
		t.Errorf("Expected default top_k 0, got %d", conv.TopK())
	}

	conv.SetTopK(50)
	if conv.TopK() != 50 {
		t.Errorf("Expected top_k 50, got %d", conv.TopK())
	}

	conv.SetTopK(-10)
	if conv.TopK() != 0 {
		t.Errorf("Expected top_k 0 (clamped), got %d", conv.TopK())
	}
}

func TestConversation_StopSequences_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	if conv.StopSequences() != nil && len(conv.StopSequences()) != 0 {
		t.Error("Expected empty stop sequences initially")
	}

	sequences := []string{"STOP", "END"}
	conv.SetStopSequences(sequences)

	result := conv.StopSequences()
	if len(result) != 2 {
		t.Errorf("Expected 2 stop sequences, got %d", len(result))
	}
}

func TestConversation_Tools_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	toolName, _ := vo.NewToolName("test_tool")
	toolDesc, _ := vo.NewToolDescription("A test tool")
	tool, _ := entities.NewTool(toolName, toolDesc, nil)

	conv.AddTool(tool)

	tools := conv.Tools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	retrieved := conv.GetTool(toolName)
	if retrieved == nil {
		t.Error("Should find added tool")
	}

	conv.RemoveTool(toolName)
	tools = conv.Tools()
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools after removal, got %d", len(tools))
	}
}

func TestConversation_Metadata_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	conv.SetMetadata("key1", "value1")
	conv.SetMetadata("key2", 42)

	val, ok := conv.GetMetadata("key1")
	if !ok {
		t.Error("Should find metadata key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	val, ok = conv.GetMetadata("key2")
	if !ok {
		t.Error("Should find metadata key2")
	}
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	metadata := conv.Metadata()
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}
}

func TestConversation_Events_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	events := conv.Events()
	if len(events) == 0 {
		t.Error("Should have at least one event after creation")
	}

	events = conv.Events()
	if len(events) != 0 {
		t.Error("Events should be cleared after retrieval")
	}

	_, _ = conv.AddUserMessage("Test")
	conv.ClearEvents()
	events = conv.Events()
	if len(events) != 0 {
		t.Error("Events should be empty after ClearEvents()")
	}
}

func TestConversation_GetMessagesForAPI_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	_, _ = conv.AddUserMessage("Hello, Claude!")

	messages := conv.GetMessagesForAPI()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg["role"] != "user" {
		t.Errorf("Expected role 'user', got '%v'", msg["role"])
	}
}

func TestConversation_ConcurrentAccess_Internal(t *testing.T) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conv.SetMetadata("key", idx)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = conv.Status()
			_ = conv.MessageCount()
			_ = conv.IsActive()
			_ = conv.Messages()
		}()
	}

	wg.Wait()
}

func BenchmarkConversation_AddUserMessage_Internal(b *testing.B) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if conv.MessageCount() >= 1000 {
			conv = aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)
		}
		_, _ = conv.AddUserMessage("Test message")
	}
}

func BenchmarkConversation_GetMessagesForAPI_Internal(b *testing.B) {
	conv := aggregates.NewConversation(vo.GenerateSessionID(), vo.ModelClaudeSonnet4)

	for i := 0; i < 100; i++ {
		_, _ = conv.AddUserMessage("Test message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.GetMessagesForAPI()
	}
}
