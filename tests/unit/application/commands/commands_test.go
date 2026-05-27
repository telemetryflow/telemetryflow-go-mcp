package commands_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/commands"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestCommandInterface(t *testing.T) {
	t.Run("all commands implement Command interface", func(t *testing.T) {
		var cmds = []commands.Command{
			&commands.InitializeSessionCommand{},
			&commands.CloseSessionCommand{},
			&commands.SetLogLevelCommand{},
			&commands.CreateConversationCommand{},
			&commands.SendMessageCommand{},
			&commands.AddToolResultCommand{},
			&commands.CloseConversationCommand{},
			&commands.RegisterToolCommand{},
			&commands.UnregisterToolCommand{},
			&commands.ExecuteToolCommand{},
			&commands.RegisterResourceCommand{},
			&commands.UnregisterResourceCommand{},
			&commands.SubscribeResourceCommand{},
			&commands.UnsubscribeResourceCommand{},
			&commands.RegisterPromptCommand{},
			&commands.UnregisterPromptCommand{},
			&commands.ExecutePromptCommand{},
			&commands.PingCommand{},
			&commands.CancelRequestCommand{},
			&commands.SendNotificationCommand{},
		}
		for _, cmd := range cmds {
			assert.NotEmpty(t, cmd.CommandName(), "CommandName should not be empty for %T", cmd)
		}
	})
}

func TestInitializeSessionCommand(t *testing.T) {
	cmd := &commands.InitializeSessionCommand{
		ClientName:      "TestClient",
		ClientVersion:   "1.0.0",
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{"tools": true},
	}
	assert.Equal(t, "InitializeSession", cmd.CommandName())
	assert.Equal(t, "TestClient", cmd.ClientName)
	assert.Equal(t, "1.0.0", cmd.ClientVersion)
	assert.Equal(t, "2024-11-05", cmd.ProtocolVersion)
}

func TestCloseSessionCommand(t *testing.T) {
	sid := vo.GenerateSessionID()
	cmd := &commands.CloseSessionCommand{SessionID: sid}
	assert.Equal(t, "CloseSession", cmd.CommandName())
	assert.Equal(t, sid, cmd.SessionID)
}

func TestSetLogLevelCommand(t *testing.T) {
	sid := vo.GenerateSessionID()
	cmd := &commands.SetLogLevelCommand{SessionID: sid, Level: vo.LogLevelDebug}
	assert.Equal(t, "SetLogLevel", cmd.CommandName())
	assert.Equal(t, vo.LogLevelDebug, cmd.Level)
}

func TestCreateConversationCommand(t *testing.T) {
	sid := vo.GenerateSessionID()
	cmd := &commands.CreateConversationCommand{
		SessionID:    sid,
		Model:        vo.ModelClaudeOpus47,
		SystemPrompt: "You are helpful",
		MaxTokens:    4096,
		Temperature:  0.7,
	}
	assert.Equal(t, "CreateConversation", cmd.CommandName())
	assert.Equal(t, vo.ModelClaudeOpus47, cmd.Model)
	assert.Equal(t, 4096, cmd.MaxTokens)
}

func TestSendMessageCommand(t *testing.T) {
	cid := vo.GenerateConversationID()
	cmd := &commands.SendMessageCommand{
		ConversationID: cid,
		Content:        "Hello",
		Stream:         false,
	}
	assert.Equal(t, "SendMessage", cmd.CommandName())
	assert.Equal(t, "Hello", cmd.Content)
}

func TestAddToolResultCommand(t *testing.T) {
	cid := vo.GenerateConversationID()
	cmd := &commands.AddToolResultCommand{
		ConversationID: cid,
		ToolUseID:      "tool_123",
		Content:        "result",
		IsError:        false,
	}
	assert.Equal(t, "AddToolResult", cmd.CommandName())
	assert.Equal(t, "tool_123", cmd.ToolUseID)
	assert.False(t, cmd.IsError)
}

func TestCloseConversationCommand(t *testing.T) {
	cid := vo.GenerateConversationID()
	cmd := &commands.CloseConversationCommand{ConversationID: cid}
	assert.Equal(t, "CloseConversation", cmd.CommandName())
}

func TestRegisterToolCommand(t *testing.T) {
	sid := vo.GenerateSessionID()
	cmd := &commands.RegisterToolCommand{
		SessionID:   sid,
		Name:        "my_tool",
		Description: "A test tool",
		Category:    "test",
		Tags:        []string{"test", "unit"},
	}
	assert.Equal(t, "RegisterTool", cmd.CommandName())
	assert.Equal(t, "my_tool", cmd.Name)
}

func TestUnregisterToolCommand(t *testing.T) {
	sid := vo.GenerateSessionID()
	cmd := &commands.UnregisterToolCommand{SessionID: sid, Name: "my_tool"}
	assert.Equal(t, "UnregisterTool", cmd.CommandName())
}

func TestExecuteToolCommand(t *testing.T) {
	sid := vo.GenerateSessionID()
	args := map[string]interface{}{"key": "value"}
	cmd := &commands.ExecuteToolCommand{
		SessionID: sid,
		Name:      "my_tool",
		Arguments: args,
	}
	assert.Equal(t, "ExecuteTool", cmd.CommandName())
	assert.Equal(t, args, cmd.Arguments)
}

func TestResourceCommands(t *testing.T) {
	sid := vo.GenerateSessionID()

	registerCmd := &commands.RegisterResourceCommand{
		SessionID: sid, URI: "file:///test", Name: "test", Description: "desc", MimeType: "text/plain",
	}
	assert.Equal(t, "RegisterResource", registerCmd.CommandName())

	unregisterCmd := &commands.UnregisterResourceCommand{SessionID: sid, URI: "file:///test"}
	assert.Equal(t, "UnregisterResource", unregisterCmd.CommandName())

	subCmd := &commands.SubscribeResourceCommand{SessionID: sid, URI: "file:///test"}
	assert.Equal(t, "SubscribeResource", subCmd.CommandName())

	unsubCmd := &commands.UnsubscribeResourceCommand{SessionID: sid, URI: "file:///test"}
	assert.Equal(t, "UnsubscribeResource", unsubCmd.CommandName())
}

func TestPromptCommands(t *testing.T) {
	sid := vo.GenerateSessionID()

	registerCmd := &commands.RegisterPromptCommand{
		SessionID: sid, Name: "test_prompt", Description: "desc",
	}
	assert.Equal(t, "RegisterPrompt", registerCmd.CommandName())

	unregisterCmd := &commands.UnregisterPromptCommand{SessionID: sid, Name: "test_prompt"}
	assert.Equal(t, "UnregisterPrompt", unregisterCmd.CommandName())

	execCmd := &commands.ExecutePromptCommand{
		SessionID: sid, Name: "test_prompt", Arguments: map[string]string{"key": "val"},
	}
	assert.Equal(t, "ExecutePrompt", execCmd.CommandName())
}

func TestMCPProtocolCommands(t *testing.T) {
	sid := vo.GenerateSessionID()

	pingCmd := &commands.PingCommand{SessionID: sid}
	assert.Equal(t, "Ping", pingCmd.CommandName())

	cancelCmd := &commands.CancelRequestCommand{SessionID: sid, RequestID: "req_123", Reason: "timeout"}
	assert.Equal(t, "CancelRequest", cancelCmd.CommandName())

	notifyCmd := &commands.SendNotificationCommand{
		SessionID: sid, Method: vo.MethodNotificationsMessage, Params: map[string]interface{}{},
	}
	assert.Equal(t, "SendNotification", notifyCmd.CommandName())
}
