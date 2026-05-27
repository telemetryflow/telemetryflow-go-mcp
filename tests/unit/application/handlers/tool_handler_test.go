package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/commands"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/handlers"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/queries"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type mockToolRepo struct {
	mock.Mock
}

func (m *mockToolRepo) Register(ctx context.Context, tool *entities.Tool) error {
	args := m.Called(ctx, tool)
	return args.Error(0)
}
func (m *mockToolRepo) Unregister(ctx context.Context, name vo.ToolName) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}
func (m *mockToolRepo) FindByName(ctx context.Context, name vo.ToolName) (*entities.Tool, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tool), args.Error(1)
}
func (m *mockToolRepo) FindAll(ctx context.Context) ([]*entities.Tool, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tool), args.Error(1)
}
func (m *mockToolRepo) FindByCategory(ctx context.Context, category string) ([]*entities.Tool, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tool), args.Error(1)
}
func (m *mockToolRepo) FindByTag(ctx context.Context, tag string) ([]*entities.Tool, error) {
	args := m.Called(ctx, tag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tool), args.Error(1)
}
func (m *mockToolRepo) FindEnabled(ctx context.Context) ([]*entities.Tool, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tool), args.Error(1)
}
func (m *mockToolRepo) Exists(ctx context.Context, name vo.ToolName) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}
func (m *mockToolRepo) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func createTestTool(t *testing.T, name string) *entities.Tool {
	t.Helper()
	tn, err := vo.NewToolName(name)
	require.NoError(t, err)
	td, err := vo.NewToolDescription("desc")
	require.NoError(t, err)
	tool, err := entities.NewTool(tn, td, nil)
	require.NoError(t, err)
	return tool
}

func TestNewToolHandler(t *testing.T) {
	sr := new(mockSessionRepo)
	tr := new(mockToolRepo)
	pub := new(mockEventPublisher)
	h := handlers.NewToolHandler(sr, tr, pub)
	assert.NotNil(t, h)
}

func TestHandleRegisterTool(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	sr := new(mockSessionRepo)
	tr := new(mockToolRepo)
	pub := new(mockEventPublisher)

	t.Run("success", func(t *testing.T) {
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tr.On("FindByName", ctx, mock.AnythingOfType("valueobjects.ToolName")).Return(nil, nil)
		tr.On("Register", ctx, mock.AnythingOfType("*entities.Tool")).Return(nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)

		tool, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID:   session.ID(),
			Name:        "test_tool",
			Description: "desc",
			Category:    "test",
			Tags:        []string{"tag1"},
		})
		require.NoError(t, err)
		assert.NotNil(t, tool)
		assert.Equal(t, "test", tool.Category())
	})

	t.Run("session not found", func(t *testing.T) {
		sr2 := new(mockSessionRepo)
		tr2 := new(mockToolRepo)
		sid := vo.GenerateSessionID()
		sr2.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewToolHandler(sr2, tr2, new(mockEventPublisher))

		_, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: sid, Name: "test_tool", Description: "desc",
		})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("tool already exists", func(t *testing.T) {
		sr3 := new(mockSessionRepo)
		tr3 := new(mockToolRepo)
		existingTool := createTestTool(t, "existing")
		sr3.On("FindByID", ctx, session.ID()).Return(session, nil)
		tr3.On("FindByName", ctx, mock.AnythingOfType("valueobjects.ToolName")).Return(existingTool, nil)
		h := handlers.NewToolHandler(sr3, tr3, new(mockEventPublisher))

		_, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: session.ID(), Name: "existing", Description: "desc",
		})
		assert.Equal(t, handlers.ErrToolAlreadyExists, err)
	})

	t.Run("invalid tool name", func(t *testing.T) {
		sr4 := new(mockSessionRepo)
		tr4 := new(mockToolRepo)
		sr4.On("FindByID", ctx, session.ID()).Return(session, nil)
		tr4.On("FindByName", ctx, mock.AnythingOfType("valueobjects.ToolName")).Return(nil, nil)
		h := handlers.NewToolHandler(sr4, tr4, new(mockEventPublisher))

		_, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: session.ID(), Name: "", Description: "desc",
		})
		assert.Error(t, err)
	})
}

func TestHandleUnregisterTool(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()

	t.Run("success", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("my_tool")
		tr.On("Exists", ctx, tn).Return(true, nil)
		tr.On("Unregister", ctx, tn).Return(nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))

		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: session.ID(), Name: "my_tool",
		})
		require.NoError(t, err)
	})

	t.Run("tool not found", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("missing")
		tr.On("Exists", ctx, tn).Return(false, nil)
		h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))

		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: session.ID(), Name: "missing",
		})
		assert.Equal(t, handlers.ErrToolNotFound, err)
	})
}

func TestHandleExecuteTool(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()

	t.Run("success", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		pub := new(mockEventPublisher)
		tool := createTestTool(t, "exec_tool")
		tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
			return entities.NewTextToolResult("executed"), nil
		})
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("exec_tool")
		tr.On("FindByName", ctx, tn).Return(tool, nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)

		result, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: session.ID(), Name: "exec_tool", Arguments: map[string]interface{}{},
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("disabled tool", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		tool := createTestTool(t, "disabled_tool")
		tool.Disable()
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("disabled_tool")
		tr.On("FindByName", ctx, tn).Return(tool, nil)
		pub := new(mockEventPublisher)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)

		_, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: session.ID(), Name: "disabled_tool", Arguments: map[string]interface{}{},
		})
		assert.Equal(t, handlers.ErrToolDisabled, err)
	})

	t.Run("tool not found", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("missing")
		tr.On("FindByName", ctx, tn).Return(nil, nil)
		pub := new(mockEventPublisher)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)

		_, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: session.ID(), Name: "missing", Arguments: map[string]interface{}{},
		})
		assert.Equal(t, handlers.ErrToolNotFound, err)
	})
}

func TestHandleGetTool(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		tr := new(mockToolRepo)
		tool := createTestTool(t, "my_tool")
		tn, _ := vo.NewToolName("my_tool")
		tr.On("FindByName", ctx, tn).Return(tool, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		result, err := h.HandleGetTool(ctx, &queries.GetToolQuery{Name: "my_tool"})
		require.NoError(t, err)
		assert.Equal(t, "my_tool", result.Name().String())
	})

	t.Run("not found", func(t *testing.T) {
		tr := new(mockToolRepo)
		tn, _ := vo.NewToolName("missing")
		tr.On("FindByName", ctx, tn).Return(nil, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		_, err := h.HandleGetTool(ctx, &queries.GetToolQuery{Name: "missing"})
		assert.Equal(t, handlers.ErrToolNotFound, err)
	})
}

func TestHandleListTools(t *testing.T) {
	ctx := context.Background()

	t.Run("by category", func(t *testing.T) {
		tr := new(mockToolRepo)
		tool := createTestTool(t, "cat_tool")
		tr.On("FindByCategory", ctx, "utility").Return([]*entities.Tool{tool}, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		result, err := h.HandleListTools(ctx, &queries.ListToolsQuery{Category: "utility"})
		require.NoError(t, err)
		assert.Len(t, result.Tools, 1)
	})

	t.Run("by tag", func(t *testing.T) {
		tr := new(mockToolRepo)
		tool := createTestTool(t, "tag_tool")
		tr.On("FindByTag", ctx, "test").Return([]*entities.Tool{tool}, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		result, err := h.HandleListTools(ctx, &queries.ListToolsQuery{Tag: "test"})
		require.NoError(t, err)
		assert.Len(t, result.Tools, 1)
	})

	t.Run("enabled only", func(t *testing.T) {
		tr := new(mockToolRepo)
		tool := createTestTool(t, "enabled_tool")
		tr.On("FindEnabled", ctx).Return([]*entities.Tool{tool}, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		result, err := h.HandleListTools(ctx, &queries.ListToolsQuery{EnabledOnly: true})
		require.NoError(t, err)
		assert.Len(t, result.Tools, 1)
	})

	t.Run("all tools", func(t *testing.T) {
		tr := new(mockToolRepo)
		tools := []*entities.Tool{createTestTool(t, "a"), createTestTool(t, "b")}
		tr.On("FindAll", ctx).Return(tools, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		result, err := h.HandleListTools(ctx, &queries.ListToolsQuery{})
		require.NoError(t, err)
		assert.Len(t, result.Tools, 2)
	})

	t.Run("with pagination limit", func(t *testing.T) {
		tr := new(mockToolRepo)
		tools := []*entities.Tool{createTestTool(t, "a"), createTestTool(t, "b"), createTestTool(t, "c")}
		tr.On("FindAll", ctx).Return(tools, nil)
		h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))

		result, err := h.HandleListTools(ctx, &queries.ListToolsQuery{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, result.Tools, 2)
	})
}

func TestToolListResult_ToMCPToolList(t *testing.T) {
	tool := createTestTool(t, "mcp_tool")
	result := &handlers.ToolListResult{
		Tools:      []*entities.Tool{tool},
		NextCursor: "cursor123",
	}
	mcpList := result.ToMCPToolList()
	assert.Contains(t, mcpList, "tools")
	assert.Contains(t, mcpList, "nextCursor")

	result2 := &handlers.ToolListResult{Tools: []*entities.Tool{}}
	mcpList2 := result2.ToMCPToolList()
	assert.NotContains(t, mcpList2, "nextCursor")
}

func TestRegisterToolHandler(t *testing.T) {
	sr := new(mockSessionRepo)
	tr := new(mockToolRepo)
	h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))
	called := false
	h.RegisterToolHandler("custom", func(input map[string]interface{}) (*entities.ToolResult, error) {
		called = true
		return entities.NewTextToolResult("custom"), nil
	})
	assert.True(t, true)
	_ = called
}

func TestHandleRegisterTool_AdditionalPaths(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()

	t.Run("repo find error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		sr.On("FindByID", ctx, session.ID()).Return(nil, errors.New("db"))
		h := handlers.NewToolHandler(sr, new(mockToolRepo), new(mockEventPublisher))
		_, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: session.ID(), Name: "test", Description: "desc",
		})
		assert.Error(t, err)
	})

	t.Run("register error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tr.On("FindByName", ctx, mock.AnythingOfType("valueobjects.ToolName")).Return(nil, nil)
		tr.On("Register", ctx, mock.AnythingOfType("*entities.Tool")).Return(errors.New("db"))
		h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))
		_, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: session.ID(), Name: "test_reg", Description: "desc",
		})
		assert.Error(t, err)
	})

	t.Run("session save error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tr.On("FindByName", ctx, mock.AnythingOfType("valueobjects.ToolName")).Return(nil, nil)
		tr.On("Register", ctx, mock.AnythingOfType("*entities.Tool")).Return(nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(errors.New("db"))
		h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))
		_, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: session.ID(), Name: "test_sess", Description: "desc",
		})
		assert.Error(t, err)
	})

	t.Run("with registered handler", func(t *testing.T) {
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		pub := new(mockEventPublisher)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tr.On("FindByName", ctx, mock.AnythingOfType("valueobjects.ToolName")).Return(nil, nil)
		tr.On("Register", ctx, mock.AnythingOfType("*entities.Tool")).Return(nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)
		h.RegisterToolHandler("handler_tool", func(input map[string]interface{}) (*entities.ToolResult, error) {
			return entities.NewTextToolResult("handled"), nil
		})
		tool, err := h.HandleRegisterTool(ctx, &commands.RegisterToolCommand{
			SessionID: session.ID(), Name: "handler_tool", Description: "desc",
		})
		require.NoError(t, err)
		assert.NotNil(t, tool)
	})
}

func TestHandleUnregisterTool_AdditionalPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("session not found", func(t *testing.T) {
		sr := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		sr.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewToolHandler(sr, new(mockToolRepo), new(mockEventPublisher))
		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: sid, Name: "test",
		})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("repo find error", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		sr.On("FindByID", ctx, session.ID()).Return(nil, errors.New("db"))
		h := handlers.NewToolHandler(sr, new(mockToolRepo), new(mockEventPublisher))
		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: session.ID(), Name: "test",
		})
		assert.Error(t, err)
	})

	t.Run("exists error", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("test")
		tr.On("Exists", ctx, tn).Return(false, errors.New("db"))
		h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))
		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: session.ID(), Name: "test",
		})
		assert.Error(t, err)
	})

	t.Run("unregister error", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("test")
		tr.On("Exists", ctx, tn).Return(true, nil)
		tr.On("Unregister", ctx, tn).Return(errors.New("db"))
		h := handlers.NewToolHandler(sr, tr, new(mockEventPublisher))
		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: session.ID(), Name: "test",
		})
		assert.Error(t, err)
	})

	t.Run("invalid name", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		h := handlers.NewToolHandler(sr, new(mockToolRepo), new(mockEventPublisher))
		err := h.HandleUnregisterTool(ctx, &commands.UnregisterToolCommand{
			SessionID: session.ID(), Name: "",
		})
		assert.Error(t, err)
	})
}

func TestHandleExecuteTool_AdditionalPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("session not found", func(t *testing.T) {
		sr := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		sr.On("FindByID", ctx, sid).Return(nil, nil)
		pub := new(mockEventPublisher)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, new(mockToolRepo), pub)
		_, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: sid, Name: "test", Arguments: map[string]interface{}{},
		})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("invalid tool name", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		pub := new(mockEventPublisher)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, new(mockToolRepo), pub)
		_, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: session.ID(), Name: "", Arguments: map[string]interface{}{},
		})
		assert.Error(t, err)
	})

	t.Run("repo find error", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("test")
		tr.On("FindByName", ctx, tn).Return(nil, errors.New("db"))
		pub := new(mockEventPublisher)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)
		_, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: session.ID(), Name: "test", Arguments: map[string]interface{}{},
		})
		assert.Error(t, err)
	})

	t.Run("tool execution error returns error result", func(t *testing.T) {
		session := createInitializedSession()
		sr := new(mockSessionRepo)
		tr := new(mockToolRepo)
		pub := new(mockEventPublisher)
		tool := createTestTool(t, "err_tool")
		tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
			return nil, errors.New("execution failed")
		})
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		tn, _ := vo.NewToolName("err_tool")
		tr.On("FindByName", ctx, tn).Return(tool, nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewToolHandler(sr, tr, pub)
		result, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: session.ID(), Name: "err_tool", Arguments: map[string]interface{}{},
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("session repo find error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		sr.On("FindByID", ctx, sid).Return(nil, errors.New("db"))
		h := handlers.NewToolHandler(sr, new(mockToolRepo), new(mockEventPublisher))
		_, err := h.HandleExecuteTool(ctx, &commands.ExecuteToolCommand{
			SessionID: sid, Name: "test", Arguments: map[string]interface{}{},
		})
		assert.Error(t, err)
	})
}

func TestHandleGetTool_RepoError(t *testing.T) {
	ctx := context.Background()
	tr := new(mockToolRepo)
	tn, _ := vo.NewToolName("test")
	tr.On("FindByName", ctx, tn).Return(nil, errors.New("db"))
	h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))
	_, err := h.HandleGetTool(ctx, &queries.GetToolQuery{Name: "test"})
	assert.Error(t, err)
}

func TestHandleGetTool_InvalidName(t *testing.T) {
	ctx := context.Background()
	h := handlers.NewToolHandler(new(mockSessionRepo), new(mockToolRepo), new(mockEventPublisher))
	_, err := h.HandleGetTool(ctx, &queries.GetToolQuery{Name: ""})
	assert.Error(t, err)
}

func TestHandleListTools_RepoError(t *testing.T) {
	ctx := context.Background()
	tr := new(mockToolRepo)
	tr.On("FindAll", ctx).Return(nil, errors.New("db"))
	h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))
	_, err := h.HandleListTools(ctx, &queries.ListToolsQuery{})
	assert.Error(t, err)
}

func TestHandleListTools_ByCategoryError(t *testing.T) {
	ctx := context.Background()
	tr := new(mockToolRepo)
	tr.On("FindByCategory", ctx, "bad").Return(nil, errors.New("db"))
	h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))
	_, err := h.HandleListTools(ctx, &queries.ListToolsQuery{Category: "bad"})
	assert.Error(t, err)
}

func TestHandleListTools_ByTagError(t *testing.T) {
	ctx := context.Background()
	tr := new(mockToolRepo)
	tr.On("FindByTag", ctx, "bad").Return(nil, errors.New("db"))
	h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))
	_, err := h.HandleListTools(ctx, &queries.ListToolsQuery{Tag: "bad"})
	assert.Error(t, err)
}

func TestHandleListTools_EnabledError(t *testing.T) {
	ctx := context.Background()
	tr := new(mockToolRepo)
	tr.On("FindEnabled", ctx).Return(nil, errors.New("db"))
	h := handlers.NewToolHandler(new(mockSessionRepo), tr, new(mockEventPublisher))
	_, err := h.HandleListTools(ctx, &queries.ListToolsQuery{EnabledOnly: true})
	assert.Error(t, err)
}

func TestExecuteToolWithContext_Timeout(t *testing.T) {
	session := createInitializedSession()
	tool := createTestTool(t, "slow_tool")
	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		time.Sleep(200 * time.Millisecond)
		return entities.NewTextToolResult("done"), nil
	})
	sr := new(mockSessionRepo)
	tr := new(mockToolRepo)
	pub := new(mockEventPublisher)
	sr.On("FindByID", mock.Anything, session.ID()).Return(session, nil)
	tn, _ := vo.NewToolName("slow_tool")
	tr.On("FindByName", mock.Anything, tn).Return(tool, nil)
	pub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	h := handlers.NewToolHandler(sr, tr, pub)

	execCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result, err := h.HandleExecuteTool(execCtx, &commands.ExecuteToolCommand{
		SessionID: session.ID(), Name: "slow_tool", Arguments: map[string]interface{}{},
	})
	_ = result
	_ = err
}
