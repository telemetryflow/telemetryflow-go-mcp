package persistence_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
)

func TestJSONB_Value_Nil(t *testing.T) {
	var j persistence.JSONB
	v, err := j.Value()
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestJSONB_Value_Data(t *testing.T) {
	j := persistence.JSONB{"key": "val"}
	v, err := j.Value()
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestJSONB_Scan_Nil(t *testing.T) {
	var j persistence.JSONB
	err := j.Scan(nil)
	require.NoError(t, err)
	assert.Nil(t, j)
}

func TestJSONB_Scan_Bytes(t *testing.T) {
	var j persistence.JSONB
	err := j.Scan([]byte(`{"hello":"world"}`))
	require.NoError(t, err)
	assert.Equal(t, "world", j["hello"])
}

func TestJSONB_Scan_String(t *testing.T) {
	var j persistence.JSONB
	err := j.Scan(`{"a":1}`)
	require.NoError(t, err)
	assert.Equal(t, float64(1), j["a"])
}

func TestJSONB_Scan_InvalidType(t *testing.T) {
	var j persistence.JSONB
	err := j.Scan(12345)
	require.Error(t, err)
}

func TestJSONBArray_Value_Nil(t *testing.T) {
	var j persistence.JSONBArray
	v, err := j.Value()
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestJSONBArray_Value_Data(t *testing.T) {
	j := persistence.JSONBArray{"a", "b"}
	v, err := j.Value()
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestJSONBArray_Scan_Nil(t *testing.T) {
	var j persistence.JSONBArray
	err := j.Scan(nil)
	require.NoError(t, err)
	assert.Nil(t, j)
}

func TestJSONBArray_Scan_Bytes(t *testing.T) {
	var j persistence.JSONBArray
	err := j.Scan([]byte(`[1,2,3]`))
	require.NoError(t, err)
	assert.Len(t, j, 3)
}

func TestJSONBArray_Scan_String(t *testing.T) {
	var j persistence.JSONBArray
	err := j.Scan(`["x","y"]`)
	require.NoError(t, err)
	assert.Len(t, j, 2)
}

func TestJSONBArray_Scan_InvalidType(t *testing.T) {
	var j persistence.JSONBArray
	err := j.Scan(42)
	require.Error(t, err)
}

func TestAllModels(t *testing.T) {
	models := persistence.AllModels()
	assert.Len(t, models, 9)
	for i, m := range models {
		assert.NotNil(t, m, "model %d is nil", i)
	}
}

func TestSessionModel_TableName(t *testing.T) {
	assert.Equal(t, "sessions", persistence.SessionModel{}.TableName())
}

func TestConversationModel_TableName(t *testing.T) {
	assert.Equal(t, "conversations", persistence.ConversationModel{}.TableName())
}

func TestMessageModel_TableName(t *testing.T) {
	assert.Equal(t, "messages", persistence.MessageModel{}.TableName())
}

func TestToolModel_TableName(t *testing.T) {
	assert.Equal(t, "tools", persistence.ToolModel{}.TableName())
}

func TestResourceModel_TableName(t *testing.T) {
	assert.Equal(t, "resources", persistence.ResourceModel{}.TableName())
}

func TestPromptModel_TableName(t *testing.T) {
	assert.Equal(t, "prompts", persistence.PromptModel{}.TableName())
}

func TestToolCallModel_TableName(t *testing.T) {
	assert.Equal(t, "tool_calls", persistence.ToolCallModel{}.TableName())
}

func TestAPIRequestModel_TableName(t *testing.T) {
	assert.Equal(t, "api_requests", persistence.APIRequestModel{}.TableName())
}

func TestAuditLogModel_TableName(t *testing.T) {
	assert.Equal(t, "audit_logs", persistence.AuditLogModel{}.TableName())
}

func TestSessionModel_JSON(t *testing.T) {
	s := persistence.SessionModel{ID: "test", State: "ready", ProtocolVersion: "2024-11-05"}
	data, err := json.Marshal(s)
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Equal(t, "ready", m["State"])
}

func TestConversationModel_JSON(t *testing.T) {
	c := persistence.ConversationModel{ID: "c1", SessionID: "s1", Model: "claude-opus-4-7", Status: "active"}
	data, err := json.Marshal(c)
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Equal(t, "active", m["Status"])
}
