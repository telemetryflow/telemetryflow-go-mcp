package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestJSONRPCVersion(t *testing.T) {
	assert.True(t, vo.JSONRPCVersion("2.0").IsValid())
	assert.False(t, vo.JSONRPCVersion("1.0").IsValid())
	assert.Equal(t, "2.0", vo.JSONRPCVersion("2.0").String())
}

func TestMCPMethod_IsValid(t *testing.T) {
	methods := []vo.MCPMethod{
		vo.MethodInitialize, vo.MethodInitialized, vo.MethodPing, vo.MethodShutdown,
		vo.MethodToolsList, vo.MethodToolsCall,
		vo.MethodResourcesList, vo.MethodResourcesRead, vo.MethodResourcesSubscribe, vo.MethodResourcesUnsubscribe,
		vo.MethodPromptsList, vo.MethodPromptsGet,
		vo.MethodCompletionComplete, vo.MethodLoggingSetLevel,
		vo.MethodNotificationsCancelled, vo.MethodNotificationsProgress, vo.MethodNotificationsMessage,
		vo.MethodNotificationsResourcesUpdated, vo.MethodNotificationsResourcesListChanged,
		vo.MethodNotificationsToolsListChanged, vo.MethodNotificationsPromptsListChanged,
	}
	for _, m := range methods {
		if !m.IsValid() {
			t.Errorf("expected %s to be valid", m)
		}
	}
	assert.False(t, vo.MCPMethod("invalid").IsValid())
	assert.False(t, vo.MCPMethod("").IsValid())
}

func TestMCPMethod_String(t *testing.T) {
	assert.Equal(t, "initialize", vo.MethodInitialize.String())
}

func TestMCPMethod_IsNotification(t *testing.T) {
	assert.True(t, vo.MethodNotificationsMessage.IsNotification())
	assert.False(t, vo.MethodInitialize.IsNotification())
}

func TestMCPCapability_IsValid(t *testing.T) {
	caps := []vo.MCPCapability{
		vo.CapabilityTools, vo.CapabilityResources, vo.CapabilityPrompts,
		vo.CapabilityLogging, vo.CapabilitySampling, vo.CapabilityRoots, vo.CapabilityExperimental,
	}
	for _, c := range caps {
		if !c.IsValid() {
			t.Errorf("expected %s to be valid", c)
		}
	}
	assert.False(t, vo.MCPCapability("invalid").IsValid())
}

func TestMCPCapability_String(t *testing.T) {
	assert.Equal(t, "tools", vo.CapabilityTools.String())
}

func TestMCPLogLevel_IsValid(t *testing.T) {
	levels := []vo.MCPLogLevel{
		vo.LogLevelDebug, vo.LogLevelInfo, vo.LogLevelNotice, vo.LogLevelWarning,
		vo.LogLevelError, vo.LogLevelCritical, vo.LogLevelAlert, vo.LogLevelEmergency,
	}
	for _, l := range levels {
		if !l.IsValid() {
			t.Errorf("expected %s to be valid", l)
		}
	}
	assert.False(t, vo.MCPLogLevel("invalid").IsValid())
}

func TestMCPLogLevel_String(t *testing.T) {
	assert.Equal(t, "info", vo.LogLevelInfo.String())
}

func TestMCPLogLevel_Severity(t *testing.T) {
	assert.Equal(t, 0, vo.LogLevelDebug.Severity())
	assert.Equal(t, 1, vo.LogLevelInfo.Severity())
	assert.Equal(t, 2, vo.LogLevelNotice.Severity())
	assert.Equal(t, 3, vo.LogLevelWarning.Severity())
	assert.Equal(t, 4, vo.LogLevelError.Severity())
	assert.Equal(t, 5, vo.LogLevelCritical.Severity())
	assert.Equal(t, 6, vo.LogLevelAlert.Severity())
	assert.Equal(t, 7, vo.LogLevelEmergency.Severity())
	assert.Equal(t, 0, vo.MCPLogLevel("unknown").Severity())
}

func TestMCPProtocolVersion(t *testing.T) {
	v := vo.NewMCPProtocolVersion("2024-11-05")
	assert.Equal(t, "2024-11-05", v.String())
	assert.True(t, v.IsLatest())

	v2 := vo.NewMCPProtocolVersion("")
	assert.Equal(t, vo.CurrentMCPProtocolVersion, v2.String())

	v3 := vo.NewMCPProtocolVersion("2023-01-01")
	assert.False(t, v3.IsLatest())
}

func TestMCPErrorCode(t *testing.T) {
	codes := []struct {
		code vo.MCPErrorCode
		msg  string
		std  bool
		mcp  bool
	}{
		{vo.ErrorCodeParseError, "Parse error", true, false},
		{vo.ErrorCodeInvalidRequest, "Invalid Request", true, false},
		{vo.ErrorCodeMethodNotFound, "Method not found", true, false},
		{vo.ErrorCodeInvalidParams, "Invalid params", true, false},
		{vo.ErrorCodeInternalError, "Internal error", true, false},
		{vo.ErrorCodeToolNotFound, "Tool not found", false, true},
		{vo.ErrorCodeResourceNotFound, "Resource not found", false, true},
		{vo.ErrorCodePromptNotFound, "Prompt not found", false, true},
		{vo.ErrorCodeToolExecutionError, "Tool execution error", false, true},
		{vo.ErrorCodeResourceReadError, "Resource read error", false, true},
		{vo.ErrorCodeUnauthorized, "Unauthorized", false, true},
		{vo.ErrorCodeRateLimited, "Rate limited", false, true},
		{vo.ErrorCodeTimeout, "Request timeout", false, true},
		{vo.ErrorCodeCancelled, "Request cancelled", false, true},
	}
	for _, c := range codes {
		assert.Equal(t, c.msg, c.code.Message(), "Message for code %d", c.code)
		assert.Equal(t, c.std, c.code.IsStandardError(), "IsStandardError for code %d", c.code)
		assert.Equal(t, c.mcp, c.code.IsMCPError(), "IsMCPError for code %d", c.code)
	}
	assert.Equal(t, "Unknown error", vo.MCPErrorCode(0).Message())
}
