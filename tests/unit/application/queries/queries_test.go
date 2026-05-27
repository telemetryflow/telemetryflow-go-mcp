package queries_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/queries"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestQueryInterface(t *testing.T) {
	t.Run("all queries implement Query interface", func(t *testing.T) {
		var qs = []queries.Query{
			&queries.GetSessionQuery{},
			&queries.ListSessionsQuery{},
			&queries.GetSessionStatsQuery{},
			&queries.GetConversationQuery{},
			&queries.ListConversationsQuery{},
			&queries.GetConversationMessagesQuery{},
			&queries.GetToolQuery{},
			&queries.ListToolsQuery{},
			&queries.GetResourceQuery{},
			&queries.ReadResourceQuery{},
			&queries.ListResourcesQuery{},
			&queries.GetPromptQuery{},
			&queries.ListPromptsQuery{},
			&queries.CompleteQuery{},
			&queries.HealthCheckQuery{},
			&queries.GetMetricsQuery{},
		}
		for _, q := range qs {
			assert.NotEmpty(t, q.QueryName(), "QueryName should not be empty for %T", q)
		}
	})
}

func TestGetSessionQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.GetSessionQuery{SessionID: sid}
	assert.Equal(t, "GetSession", q.QueryName())
	assert.Equal(t, sid, q.SessionID)
}

func TestListSessionsQuery(t *testing.T) {
	q := &queries.ListSessionsQuery{ActiveOnly: true, Cursor: "abc", Limit: 10}
	assert.Equal(t, "ListSessions", q.QueryName())
	assert.True(t, q.ActiveOnly)
	assert.Equal(t, 10, q.Limit)
}

func TestGetSessionStatsQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.GetSessionStatsQuery{SessionID: sid}
	assert.Equal(t, "GetSessionStats", q.QueryName())
}

func TestGetConversationQuery(t *testing.T) {
	cid := vo.GenerateConversationID()
	q := &queries.GetConversationQuery{ConversationID: cid}
	assert.Equal(t, "GetConversation", q.QueryName())
}

func TestListConversationsQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.ListConversationsQuery{SessionID: sid, ActiveOnly: true, Limit: 5}
	assert.Equal(t, "ListConversations", q.QueryName())
}

func TestGetConversationMessagesQuery(t *testing.T) {
	cid := vo.GenerateConversationID()
	q := &queries.GetConversationMessagesQuery{ConversationID: cid, Offset: 10, Limit: 20}
	assert.Equal(t, "GetConversationMessages", q.QueryName())
	assert.Equal(t, 10, q.Offset)
	assert.Equal(t, 20, q.Limit)
}

func TestGetToolQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.GetToolQuery{SessionID: sid, Name: "my_tool"}
	assert.Equal(t, "GetTool", q.QueryName())
}

func TestListToolsQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.ListToolsQuery{SessionID: sid, Category: "utility", EnabledOnly: true}
	assert.Equal(t, "ListTools", q.QueryName())
}

func TestResourceQueries(t *testing.T) {
	sid := vo.GenerateSessionID()

	getQ := &queries.GetResourceQuery{SessionID: sid, URI: "file:///test"}
	assert.Equal(t, "GetResource", getQ.QueryName())

	readQ := &queries.ReadResourceQuery{SessionID: sid, URI: "file:///test"}
	assert.Equal(t, "ReadResource", readQ.QueryName())

	listQ := &queries.ListResourcesQuery{SessionID: sid, TemplatesOnly: true}
	assert.Equal(t, "ListResources", listQ.QueryName())
}

func TestPromptQueries(t *testing.T) {
	sid := vo.GenerateSessionID()

	getQ := &queries.GetPromptQuery{SessionID: sid, Name: "test", Arguments: map[string]string{"k": "v"}}
	assert.Equal(t, "GetPrompt", getQ.QueryName())

	listQ := &queries.ListPromptsQuery{SessionID: sid}
	assert.Equal(t, "ListPrompts", listQ.QueryName())
}

func TestCompleteQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.CompleteQuery{
		SessionID: sid,
		Ref:       queries.CompletionRef{Type: "ref/prompt", Name: "test"},
		Argument:  queries.CompletionArgument{Name: "arg1", Value: "val"},
	}
	assert.Equal(t, "Complete", q.QueryName())
	assert.Equal(t, "ref/prompt", q.Ref.Type)
	assert.Equal(t, "arg1", q.Argument.Name)
}

func TestHealthCheckQuery(t *testing.T) {
	q := &queries.HealthCheckQuery{}
	assert.Equal(t, "HealthCheck", q.QueryName())
}

func TestGetMetricsQuery(t *testing.T) {
	sid := vo.GenerateSessionID()
	q := &queries.GetMetricsQuery{SessionID: sid}
	assert.Equal(t, "GetMetrics", q.QueryName())
}
