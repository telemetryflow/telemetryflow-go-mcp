package session_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewSession_Internal(t *testing.T) {
	session := aggregates.NewSession()

	if session == nil {
		t.Fatal("NewSession() returned nil")
	}

	if session.ID().IsEmpty() {
		t.Error("Session ID should not be empty")
	}

	if session.State() != aggregates.SessionStateCreated {
		t.Errorf("Expected state %s, got %s", aggregates.SessionStateCreated, session.State())
	}

	if session.ServerInfo() == nil {
		t.Error("ServerInfo should not be nil")
	}

	if session.ServerInfo().Name != "TelemetryFlow-MCP" {
		t.Errorf("Expected server name TelemetryFlow-MCP, got %s", session.ServerInfo().Name)
	}

	if session.Capabilities() == nil {
		t.Error("Capabilities should not be nil")
	}

	if session.Capabilities().Tools == nil {
		t.Error("Tools capability should not be nil")
	}

	if !session.Capabilities().Tools.ListChanged {
		t.Error("Tools.ListChanged should be true")
	}

	if session.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestSession_Initialize_Internal(t *testing.T) {
	session := aggregates.NewSession()

	clientInfo := &aggregates.ClientInfo{
		Name:    "TestClient",
		Version: "1.0.0",
	}

	err := session.Initialize(clientInfo, "2024-11-05")
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if session.State() != aggregates.SessionStateInitializing {
		t.Errorf("Expected state %s, got %s", aggregates.SessionStateInitializing, session.State())
	}

	if session.ClientInfo() == nil {
		t.Error("ClientInfo should not be nil after initialization")
	}

	if session.ClientInfo().Name != "TestClient" {
		t.Errorf("Expected client name TestClient, got %s", session.ClientInfo().Name)
	}

	err = session.Initialize(clientInfo, "2024-11-05")
	if err == nil {
		t.Error("Expected error when initializing already initialized session")
	}
}

func TestSession_MarkReady_Internal(t *testing.T) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Test", Version: "1.0"}

	session.MarkReady()
	if session.State() != aggregates.SessionStateCreated {
		t.Error("State should not change from Created to Ready directly")
	}

	_ = session.Initialize(clientInfo, "2024-11-05")

	session.MarkReady()
	if session.State() != aggregates.SessionStateReady {
		t.Errorf("Expected state %s, got %s", aggregates.SessionStateReady, session.State())
	}

	if !session.IsReady() {
		t.Error("IsReady() should return true")
	}
}

func TestSession_Close_Internal(t *testing.T) {
	session := aggregates.NewSession()

	session.Close()

	if session.State() != aggregates.SessionStateClosed {
		t.Errorf("Expected state %s, got %s", aggregates.SessionStateClosed, session.State())
	}

	if !session.IsClosed() {
		t.Error("IsClosed() should return true")
	}

	if session.ClosedAt() == nil {
		t.Error("ClosedAt should not be nil after closing")
	}

	session.Close()
	if session.State() != aggregates.SessionStateClosed {
		t.Error("Session should remain closed")
	}
}

func TestSession_Tools_Internal(t *testing.T) {
	session := aggregates.NewSession()

	toolName, _ := vo.NewToolName("test_tool")
	toolDesc, _ := vo.NewToolDescription("A test tool")
	tool, _ := entities.NewTool(toolName, toolDesc, nil)

	session.RegisterTool(tool)

	retrieved, ok := session.GetTool("test_tool")
	if !ok {
		t.Error("Should find registered tool")
	}
	if retrieved.Name().String() != "test_tool" {
		t.Errorf("Expected tool name test_tool, got %s", retrieved.Name().String())
	}

	tools := session.ListTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	session.UnregisterTool("test_tool")
	_, ok = session.GetTool("test_tool")
	if ok {
		t.Error("Should not find unregistered tool")
	}
}

func TestSession_Resources_Internal(t *testing.T) {
	session := aggregates.NewSession()

	resourceURI, _ := vo.NewResourceURI("file:///test/resource")
	resource, _ := entities.NewResource(resourceURI, "Test Resource")

	session.RegisterResource(resource)

	retrieved, ok := session.GetResource("file:///test/resource")
	if !ok {
		t.Error("Should find registered resource")
	}
	if retrieved.URI().String() != "file:///test/resource" {
		t.Errorf("Expected URI file:///test/resource, got %s", retrieved.URI().String())
	}

	resources := session.ListResources()
	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	session.UnregisterResource("file:///test/resource")
	_, ok = session.GetResource("file:///test/resource")
	if ok {
		t.Error("Should not find unregistered resource")
	}
}

func TestSession_Subscriptions_Internal(t *testing.T) {
	session := aggregates.NewSession()

	uri := "file:///test/resource"

	err := session.SubscribeResource(uri)
	if err != nil {
		t.Fatalf("SubscribeResource() failed: %v", err)
	}

	if !session.IsSubscribed(uri) {
		t.Error("Should be subscribed to resource")
	}

	session.UnsubscribeResource(uri)
	if session.IsSubscribed(uri) {
		t.Error("Should not be subscribed after unsubscribe")
	}
}

func TestSession_Prompts_Internal(t *testing.T) {
	session := aggregates.NewSession()

	promptName, _ := vo.NewToolName("test_prompt")
	prompt, _ := entities.NewPrompt(promptName, "Test prompt description")

	session.RegisterPrompt(prompt)

	retrieved, ok := session.GetPrompt("test_prompt")
	if !ok {
		t.Error("Should find registered prompt")
	}
	if retrieved.Name().String() != "test_prompt" {
		t.Errorf("Expected prompt name test_prompt, got %s", retrieved.Name().String())
	}

	prompts := session.ListPrompts()
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}

	session.UnregisterPrompt("test_prompt")
	_, ok = session.GetPrompt("test_prompt")
	if ok {
		t.Error("Should not find unregistered prompt")
	}
}

func TestSession_Conversations_Internal(t *testing.T) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	model := vo.ModelClaudeSonnet4

	conv, err := session.CreateConversation(model)
	if err != nil {
		t.Fatalf("CreateConversation() failed: %v", err)
	}

	if conv == nil {
		t.Fatal("Conversation should not be nil")
	}

	retrieved, ok := session.GetConversation(conv.ID())
	if !ok {
		t.Error("Should find created conversation")
	}
	if !retrieved.ID().Equals(conv.ID()) {
		t.Error("Retrieved conversation ID should match")
	}

	convs := session.ListConversations()
	if len(convs) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(convs))
	}

	err = session.CloseConversation(conv.ID())
	if err != nil {
		t.Fatalf("CloseConversation() failed: %v", err)
	}
}

func TestSession_CreateConversation_WhenClosed_Internal(t *testing.T) {
	session := aggregates.NewSession()
	session.Close()

	model := vo.ModelClaudeSonnet4
	_, err := session.CreateConversation(model)
	if err != aggregates.ErrSessionClosed {
		t.Errorf("Expected ErrSessionClosed, got %v", err)
	}
}

func TestSession_LogLevel_Internal(t *testing.T) {
	session := aggregates.NewSession()

	if session.LogLevel() != vo.LogLevelInfo {
		t.Errorf("Expected default log level %s, got %s", vo.LogLevelInfo, session.LogLevel())
	}

	err := session.SetLogLevel(vo.LogLevelDebug)
	if err != nil {
		t.Fatalf("SetLogLevel() failed: %v", err)
	}

	if session.LogLevel() != vo.LogLevelDebug {
		t.Errorf("Expected log level %s, got %s", vo.LogLevelDebug, session.LogLevel())
	}
}

func TestSession_Metadata_Internal(t *testing.T) {
	session := aggregates.NewSession()

	session.SetMetadata("key1", "value1")
	session.SetMetadata("key2", 42)

	val, ok := session.GetMetadata("key1")
	if !ok {
		t.Error("Should find metadata key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	val, ok = session.GetMetadata("key2")
	if !ok {
		t.Error("Should find metadata key2")
	}
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	metadata := session.Metadata()
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}
}

func TestSession_Events_Internal(t *testing.T) {
	session := aggregates.NewSession()

	events := session.Events()
	if len(events) == 0 {
		t.Error("Should have at least one event after creation")
	}

	events = session.Events()
	if len(events) != 0 {
		t.Error("Events should be cleared after retrieval")
	}
}

func TestSession_ToInitializeResult_Internal(t *testing.T) {
	session := aggregates.NewSession()

	result := session.ToInitializeResult()

	if result["serverInfo"] == nil {
		t.Error("Result should contain serverInfo")
	}

	if result["capabilities"] == nil {
		t.Error("Result should contain capabilities")
	}

	if result["protocolVersion"] == nil {
		t.Error("Result should contain protocolVersion")
	}
}

func TestSession_ConcurrentAccess_Internal(t *testing.T) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			toolName, _ := vo.NewToolName(fmt.Sprintf("tool_%d", idx))
			toolDesc, _ := vo.NewToolDescription("Description")
			tool, _ := entities.NewTool(toolName, toolDesc, nil)
			session.RegisterTool(tool)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = session.ListTools()
			_ = session.State()
			_ = session.IsReady()
		}()
	}

	wg.Wait()

	tools := session.ListTools()
	if len(tools) != numGoroutines {
		t.Errorf("Expected %d tools, got %d", numGoroutines, len(tools))
	}
}

func BenchmarkSession_RegisterTool_Internal(b *testing.B) {
	session := aggregates.NewSession()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toolName, _ := vo.NewToolName(fmt.Sprintf("tool_%d", i))
		toolDesc, _ := vo.NewToolDescription("Description")
		tool, _ := entities.NewTool(toolName, toolDesc, nil)
		session.RegisterTool(tool)
	}
}

func BenchmarkSession_GetTool_Internal(b *testing.B) {
	session := aggregates.NewSession()

	for i := 0; i < 100; i++ {
		toolName, _ := vo.NewToolName(fmt.Sprintf("tool_%d", i))
		toolDesc, _ := vo.NewToolDescription("Description")
		tool, _ := entities.NewTool(toolName, toolDesc, nil)
		session.RegisterTool(tool)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.GetTool("tool_50")
	}
}

func TestSession_Close_WithConversations(t *testing.T) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	conv, err := session.CreateConversation(vo.ModelClaudeSonnet4)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	session.Close()

	if conv.Status() != aggregates.ConversationStatusClosed {
		t.Errorf("expected conversation closed, got %s", conv.Status())
	}
}

func TestSession_CloseConversation_NotFound(t *testing.T) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	missingID := vo.GenerateConversationID()
	err := session.CloseConversation(missingID)
	if err != aggregates.ErrConversationNotFound {
		t.Errorf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestSession_RegisterResource_Template(t *testing.T) {
	session := aggregates.NewSession()
	tmpl, err := entities.NewResourceTemplate("file:///test/{id}", "Template", "desc")
	if err != nil {
		t.Fatalf("NewResourceTemplate: %v", err)
	}

	session.RegisterResource(tmpl)

	_, ok := session.GetResource("file:///test/{id}")
	if !ok {
		t.Error("expected to find template resource by uriTemplate key")
	}
}

func TestSession_SetLogLevel_Invalid(t *testing.T) {
	session := aggregates.NewSession()
	err := session.SetLogLevel(vo.MCPLogLevel("bogus"))
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestSession_ListTools_DisabledExcluded(t *testing.T) {
	session := aggregates.NewSession()
	tn, _ := vo.NewToolName("enabled_tool")
	td, _ := vo.NewToolDescription("desc")
	tool, _ := entities.NewTool(tn, td, nil)
	session.RegisterTool(tool)

	tn2, _ := vo.NewToolName("disabled_tool")
	td2, _ := vo.NewToolDescription("desc")
	tool2, _ := entities.NewTool(tn2, td2, nil)
	tool2.Disable()
	session.RegisterTool(tool2)

	tools := session.ListTools()
	for _, t_ := range tools {
		if !t_.IsEnabled() {
			t.Error("ListTools should only return enabled tools")
		}
	}
}

func TestSession_SubscribeResource_NoCapability(t *testing.T) {
	session := aggregates.NewSession()
	_ = session.SetLogLevel(vo.LogLevelInfo)

	caps := session.Capabilities()
	caps.Resources.Subscribe = false

	err := session.SubscribeResource("file:///test")
	if err != aggregates.ErrCapabilityNotSupported {
		t.Errorf("expected ErrCapabilityNotSupported, got %v", err)
	}
}

func TestSession_CreateConversation_MaxExceeded(t *testing.T) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	for i := 0; i < 100; i++ {
		_, err := session.CreateConversation(vo.ModelClaudeSonnet4)
		if err != nil {
			t.Fatalf("CreateConversation %d: %v", i, err)
		}
	}

	conv, err := session.CreateConversation(vo.ModelClaudeSonnet4)
	if err != nil {
		t.Fatalf("domain allows unlimited conversations: %v", err)
	}
	if conv == nil {
		t.Error("expected conversation to be created")
	}
	list := session.ListConversations()
	if len(list) != 101 {
		t.Errorf("expected 101 conversations, got %d", len(list))
	}
}

func TestSession_GetMetadata_Missing(t *testing.T) {
	session := aggregates.NewSession()
	_, ok := session.GetMetadata("nonexistent")
	if ok {
		t.Error("expected false for missing metadata key")
	}
}
