package aggregates_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func createInitializedSession() *aggregates.Session {
	s := aggregates.NewSession()
	_ = s.Initialize(&aggregates.ClientInfo{Name: "test", Version: "1.0"}, "2024-11-05")
	s.MarkReady()
	return s
}

func TestSession_UpdatedAt(t *testing.T) {
	s := createInitializedSession()
	ua := s.UpdatedAt()
	assert.False(t, ua.IsZero(), "UpdatedAt should not be zero after initialization")
}

func TestSession_ClearEvents(t *testing.T) {
	s := createInitializedSession()
	s.ClearEvents()
	events := s.Events()
	assert.Empty(t, events)
}

func TestRestoreSession(t *testing.T) {
	id := vo.GenerateSessionID()
	pv := vo.NewMCPProtocolVersion("")
	now := time.Now().UTC()

	t.Run("without client info", func(t *testing.T) {
		s := aggregates.RestoreSession(id, pv, aggregates.SessionStateReady, "srv", "1.0", "", "", "debug", now, now, nil)
		assert.Equal(t, id, s.ID())
		assert.Equal(t, aggregates.SessionStateReady, s.State())
	})

	t.Run("with client info", func(t *testing.T) {
		s := aggregates.RestoreSession(id, pv, aggregates.SessionStateReady, "srv", "1.0", "cli", "2.0", "info", now, now, nil)
		assert.Equal(t, id, s.ID())
	})
}

func TestConversation_UpdatedAt(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	require.NoError(t, err)
	ua := conv.UpdatedAt()
	assert.False(t, ua.IsZero())
}

func TestRestoreConversation(t *testing.T) {
	cid := vo.GenerateConversationID()
	sid := vo.GenerateSessionID()
	now := time.Now().UTC()

	conv := aggregates.RestoreConversation(cid, sid, string(vo.ModelClaudeOpus47), "active", now, now, nil)
	assert.Equal(t, cid, conv.ID())
	assert.Equal(t, sid, conv.SessionID())
	assert.True(t, conv.IsActive())
}

func TestConversation_ClearEvents(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	require.NoError(t, err)
	conv.ClearEvents()
	events := conv.Events()
	assert.Empty(t, events)
}

func TestConversation_MetadataMethods(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	require.NoError(t, err)

	conv.SetMetadata("key1", "val1")
	val, ok := conv.GetMetadata("key1")
	assert.True(t, ok)
	assert.Equal(t, "val1", val)

	_, ok = conv.GetMetadata("missing")
	assert.False(t, ok)

	md := conv.Metadata()
	assert.Equal(t, "val1", md["key1"])
}

func TestConversation_GetMessagesForAPI(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	require.NoError(t, err)
	_, err = conv.AddUserMessage("hello")
	require.NoError(t, err)

	apiMsgs := conv.GetMessagesForAPI()
	assert.Len(t, apiMsgs, 1)
	assert.Equal(t, "user", apiMsgs[0]["role"])
}

func TestConversation_AddRemoveTool(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	require.NoError(t, err)

	tn, _ := vo.NewToolName("my_tool")
	td, _ := vo.NewToolDescription("desc")
	tool, err := entities.NewTool(tn, td, nil)
	require.NoError(t, err)

	conv.AddTool(tool)

	tn2, _ := vo.NewToolName("my_tool")
	found := conv.GetTool(tn2)
	assert.NotNil(t, found)

	conv.RemoveTool(tn2)
	found = conv.GetTool(tn2)
	assert.Nil(t, found)
}

func TestConversation_GetMessagesForAPI_AllContentTypes(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	_, err = conv.AddUserMessage("hello from user")
	if err != nil {
		t.Fatalf("AddUserMessage: %v", err)
	}

	assistantContent := []entities.ContentBlock{
		{Type: vo.ContentTypeText, Text: "response text"},
		{Type: vo.ContentTypeToolUse, ID: "tool-call-1", Name: "read_file", Input: map[string]interface{}{"path": "/tmp/test"}},
	}
	_, err = conv.AddAssistantMessage(assistantContent)
	if err != nil {
		t.Fatalf("AddAssistantMessage: %v", err)
	}

	toolResultContent := []entities.ContentBlock{
		{Type: vo.ContentTypeToolResult, ToolUseID: "tool-call-1", Content: "file contents here"},
	}
	toolResultMsg, err := entities.NewMessage(vo.RoleUser, toolResultContent)
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if err := conv.AddMessage(toolResultMsg); err != nil {
		t.Fatalf("AddMessage: %v", err)
	}

	errContent := []entities.ContentBlock{
		{Type: vo.ContentTypeText, Text: "let me try again"},
	}
	_, err = conv.AddAssistantMessage(errContent)
	if err != nil {
		t.Fatalf("AddAssistantMessage after tool result: %v", err)
	}

	errToolResultContent := []entities.ContentBlock{
		{Type: vo.ContentTypeToolResult, ToolUseID: "tool-call-1", Content: "error occurred", IsError: true},
	}
	errToolResultMsg, err := entities.NewMessage(vo.RoleUser, errToolResultContent)
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if err := conv.AddMessage(errToolResultMsg); err != nil {
		t.Fatalf("AddMessage error result: %v", err)
	}

	apiMsgs := conv.GetMessagesForAPI()
	if len(apiMsgs) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(apiMsgs))
	}

	userMsg := apiMsgs[0]
	if userMsg["role"] != "user" {
		t.Errorf("expected user role, got %v", userMsg["role"])
	}
	userContent := userMsg["content"].([]map[string]interface{})
	if userContent[0]["type"] != "text" {
		t.Errorf("expected text type, got %v", userContent[0]["type"])
	}
	if userContent[0]["text"] != "hello from user" {
		t.Errorf("expected 'hello from user', got %v", userContent[0]["text"])
	}

	assistantMsg := apiMsgs[1]
	assistantContentArr := assistantMsg["content"].([]map[string]interface{})
	if assistantContentArr[0]["type"] != "text" {
		t.Errorf("expected text type, got %v", assistantContentArr[0]["type"])
	}
	if assistantContentArr[0]["text"] != "response text" {
		t.Errorf("expected 'response text', got %v", assistantContentArr[0]["text"])
	}
	if assistantContentArr[1]["type"] != "tool_use" {
		t.Errorf("expected tool_use type, got %v", assistantContentArr[1]["type"])
	}
	if assistantContentArr[1]["id"] != "tool-call-1" {
		t.Errorf("expected tool-call-1, got %v", assistantContentArr[1]["id"])
	}
	if assistantContentArr[1]["name"] != "read_file" {
		t.Errorf("expected read_file, got %v", assistantContentArr[1]["name"])
	}
	if assistantContentArr[1]["input"].(map[string]interface{})["path"] != "/tmp/test" {
		t.Errorf("expected input path /tmp/test, got %v", assistantContentArr[1]["input"])
	}

	toolResultAPI := apiMsgs[2]
	trContent := toolResultAPI["content"].([]map[string]interface{})
	if trContent[0]["type"] != "tool_result" {
		t.Errorf("expected tool_result type, got %v", trContent[0]["type"])
	}
	if trContent[0]["tool_use_id"] != "tool-call-1" {
		t.Errorf("expected tool-call-1, got %v", trContent[0]["tool_use_id"])
	}
	if trContent[0]["content"] != "file contents here" {
		t.Errorf("expected file contents, got %v", trContent[0]["content"])
	}
	if _, hasIsError := trContent[0]["is_error"]; hasIsError {
		t.Error("expected no is_error field for non-error tool result")
	}

	errToolResultAPI := apiMsgs[4]
	treContent := errToolResultAPI["content"].([]map[string]interface{})
	if treContent[0]["type"] != "tool_result" {
		t.Errorf("expected tool_result type, got %v", treContent[0]["type"])
	}
	if treContent[0]["is_error"] != true {
		t.Errorf("expected is_error=true, got %v", treContent[0]["is_error"])
	}
}

func TestConversation_GetMessagesForAPI_Empty(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	apiMsgs := conv.GetMessagesForAPI()
	if len(apiMsgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(apiMsgs))
	}
}

func TestConversation_Lifecycle(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	if !conv.IsActive() {
		t.Error("expected active")
	}

	conv.Pause()
	if conv.Status() != aggregates.ConversationStatusPaused {
		t.Errorf("expected paused, got %s", conv.Status())
	}

	conv.Resume()
	if !conv.IsActive() {
		t.Error("expected active after resume")
	}

	conv.Close()
	if conv.Status() != aggregates.ConversationStatusClosed {
		t.Errorf("expected closed, got %s", conv.Status())
	}
	if conv.ClosedAt() == nil {
		t.Error("expected ClosedAt to be set")
	}

	conv.Archive()
	if conv.Status() != aggregates.ConversationStatusArchived {
		t.Errorf("expected archived, got %s", conv.Status())
	}
}

func TestConversation_SetParameters(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	conv.SetMaxTokens(8192)
	if conv.MaxTokens() != 8192 {
		t.Errorf("expected 8192, got %d", conv.MaxTokens())
	}

	conv.SetTemperature(0.5)
	if conv.Temperature() != 0.5 {
		t.Errorf("expected 0.5, got %f", conv.Temperature())
	}

	conv.SetTemperature(-1.0)
	if conv.Temperature() != 0.0 {
		t.Errorf("expected clamped to 0.0, got %f", conv.Temperature())
	}

	conv.SetTemperature(5.0)
	if conv.Temperature() != 2.0 {
		t.Errorf("expected clamped to 2.0, got %f", conv.Temperature())
	}

	conv.SetTopP(0.9)
	if conv.TopP() != 0.9 {
		t.Errorf("expected 0.9, got %f", conv.TopP())
	}

	conv.SetTopP(-0.1)
	if conv.TopP() != 0.0 {
		t.Errorf("expected clamped to 0.0, got %f", conv.TopP())
	}

	conv.SetTopP(1.5)
	if conv.TopP() != 1.0 {
		t.Errorf("expected clamped to 1.0, got %f", conv.TopP())
	}

	conv.SetTopK(50)
	if conv.TopK() != 50 {
		t.Errorf("expected 50, got %d", conv.TopK())
	}

	conv.SetTopK(-5)
	if conv.TopK() != 0 {
		t.Errorf("expected clamped to 0, got %d", conv.TopK())
	}

	conv.SetStopSequences([]string{"STOP", "END"})
	if len(conv.StopSequences()) != 2 {
		t.Errorf("expected 2 stop sequences, got %d", len(conv.StopSequences()))
	}
}

func TestConversation_AddMessageErrors(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	conv.Close()

	_, err = conv.AddUserMessage("should fail")
	if err != aggregates.ErrConversationClosed {
		t.Errorf("expected ErrConversationClosed, got %v", err)
	}
}

func TestConversation_SystemPromptImmutable(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	sp, err := vo.NewSystemPrompt("initial prompt")
	if err != nil {
		t.Fatalf("NewSystemPrompt: %v", err)
	}
	if err := conv.SetSystemPrompt(sp); err != nil {
		t.Fatalf("SetSystemPrompt before messages: %v", err)
	}

	_, err = conv.AddUserMessage("hello")
	if err != nil {
		t.Fatalf("AddUserMessage: %v", err)
	}

	sp2, err := vo.NewSystemPrompt("changed prompt")
	if err != nil {
		t.Fatalf("NewSystemPrompt: %v", err)
	}
	err = conv.SetSystemPrompt(sp2)
	if err != aggregates.ErrSystemPromptImmutable {
		t.Errorf("expected ErrSystemPromptImmutable, got %v", err)
	}
}

func TestConversation_SetModel(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	err = conv.SetModel(vo.ModelClaudeSonnet4)
	if err != nil {
		t.Fatalf("SetModel: %v", err)
	}
	if conv.Model() != vo.ModelClaudeSonnet4 {
		t.Errorf("expected %s, got %s", vo.ModelClaudeSonnet4, conv.Model())
	}
}

func TestConversation_InvalidMessageOrder(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	_, err = conv.AddUserMessage("first")
	if err != nil {
		t.Fatalf("AddUserMessage: %v", err)
	}

	_, err = conv.AddUserMessage("second consecutive user")
	if err != aggregates.ErrInvalidMessageOrder {
		t.Errorf("expected ErrInvalidMessageOrder, got %v", err)
	}
}

func TestConversation_AddAssistantMessage_Closed(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	_, err = conv.AddUserMessage("hello")
	if err != nil {
		t.Fatalf("AddUserMessage: %v", err)
	}

	conv.Close()

	_, err = conv.AddAssistantMessage([]entities.ContentBlock{
		{Type: vo.ContentTypeText, Text: "should fail"},
	})
	if err != aggregates.ErrConversationClosed {
		t.Errorf("expected ErrConversationClosed, got %v", err)
	}
}

func TestConversation_AddUserMessage_EmptyText(t *testing.T) {
	s := createInitializedSession()
	conv, err := s.CreateConversation(vo.ModelClaudeOpus47)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	msg, err := conv.AddUserMessage("")
	if err != nil {
		t.Fatalf("AddUserMessage with empty text should not error: %v", err)
	}
	if msg == nil {
		t.Error("expected message to be created")
	}
}
