package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/commands"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/handlers"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/queries"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type mockConversationRepo struct {
	mock.Mock
}

func (m *mockConversationRepo) Save(ctx context.Context, c *aggregates.Conversation) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *mockConversationRepo) FindByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.Conversation), args.Error(1)
}
func (m *mockConversationRepo) FindBySessionID(ctx context.Context, sid vo.SessionID) ([]*aggregates.Conversation, error) {
	args := m.Called(ctx, sid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Conversation), args.Error(1)
}
func (m *mockConversationRepo) FindActive(ctx context.Context) ([]*aggregates.Conversation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Conversation), args.Error(1)
}
func (m *mockConversationRepo) Delete(ctx context.Context, id vo.ConversationID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockConversationRepo) Exists(ctx context.Context, id vo.ConversationID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockConversationRepo) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockConversationRepo) CountBySessionID(ctx context.Context, sid vo.SessionID) (int, error) {
	args := m.Called(ctx, sid)
	return args.Int(0), args.Error(1)
}

type mockClaudeSvc struct {
	mock.Mock
}

func (m *mockClaudeSvc) CreateMessage(ctx context.Context, req *services.ClaudeRequest) (*services.ClaudeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ClaudeResponse), args.Error(1)
}
func (m *mockClaudeSvc) CreateMessageStream(ctx context.Context, req *services.ClaudeRequest) (<-chan *services.ClaudeStreamEvent, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *services.ClaudeStreamEvent), args.Error(1)
}
func (m *mockClaudeSvc) CountTokens(ctx context.Context, req *services.ClaudeRequest) (int, error) {
	args := m.Called(ctx, req)
	return args.Int(0), args.Error(1)
}
func (m *mockClaudeSvc) ValidateRequest(req *services.ClaudeRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestNewConversationHandler(t *testing.T) {
	h := handlers.NewConversationHandler(
		new(mockSessionRepo), new(mockConversationRepo), new(mockClaudeSvc), new(mockEventPublisher),
	)
	assert.NotNil(t, h)
}

func TestHandleCreateConversation(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()

	t.Run("success", func(t *testing.T) {
		sr := new(mockSessionRepo)
		cr := new(mockConversationRepo)
		pub := new(mockEventPublisher)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(sr, cr, new(mockClaudeSvc), pub)

		conv, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID:    session.ID(),
			Model:        vo.ModelClaudeOpus47,
			SystemPrompt: "You are helpful",
			MaxTokens:    4096,
			Temperature:  0.7,
		})
		require.NoError(t, err)
		assert.NotNil(t, conv)
		assert.True(t, conv.IsActive())
	})

	t.Run("session not found", func(t *testing.T) {
		sr := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		sr.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewConversationHandler(sr, new(mockConversationRepo), new(mockClaudeSvc), new(mockEventPublisher))

		_, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: sid, Model: vo.ModelClaudeOpus47,
		})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("closed session", func(t *testing.T) {
		sr := new(mockSessionRepo)
		closedSession := createInitializedSession()
		closedSession.Close()
		sr.On("FindByID", ctx, closedSession.ID()).Return(closedSession, nil)
		h := handlers.NewConversationHandler(sr, new(mockConversationRepo), new(mockClaudeSvc), new(mockEventPublisher))

		_, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: closedSession.ID(), Model: vo.ModelClaudeOpus47,
		})
		assert.Error(t, err)
	})

	t.Run("defaults model when invalid", func(t *testing.T) {
		sr := new(mockSessionRepo)
		cr := new(mockConversationRepo)
		pub := new(mockEventPublisher)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(sr, cr, new(mockClaudeSvc), pub)

		conv, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: session.ID(), Model: vo.Model("invalid"),
		})
		require.NoError(t, err)
		assert.Equal(t, vo.DefaultModel, conv.Model())
	})
}

func TestHandleSendMessage(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("success", func(t *testing.T) {
		cr := new(mockConversationRepo)
		pub := new(mockEventPublisher)
		cs := new(mockClaudeSvc)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cs.On("CreateMessage", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return(&services.ClaudeResponse{
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "response"},
			},
		}, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, pub)

		resp, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv.ID(), Content: "Hello", Stream: false,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("empty message", func(t *testing.T) {
		h := handlers.NewConversationHandler(nil, new(mockConversationRepo), new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv.ID(), Content: "",
		})
		assert.Equal(t, handlers.ErrMessageEmpty, err)
	})

	t.Run("conversation not found", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: cid, Content: "Hello",
		})
		assert.Equal(t, handlers.ErrConversationNotFound, err)
	})
}

func TestHandleCloseConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		session := createInitializedSession()
		conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr := new(mockConversationRepo)
		pub := new(mockEventPublisher)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), pub)

		err := h.HandleCloseConversation(ctx, &commands.CloseConversationCommand{ConversationID: conv.ID()})
		require.NoError(t, err)
		assert.False(t, conv.IsActive())
	})

	t.Run("not found", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		err := h.HandleCloseConversation(ctx, &commands.CloseConversationCommand{ConversationID: cid})
		assert.Equal(t, handlers.ErrConversationNotFound, err)
	})
}

func TestHandleGetConversation(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("found", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		result, err := h.HandleGetConversation(ctx, &queries.GetConversationQuery{ConversationID: conv.ID()})
		require.NoError(t, err)
		assert.Equal(t, conv.ID(), result.ID())
	})

	t.Run("not found", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		_, err := h.HandleGetConversation(ctx, &queries.GetConversationQuery{ConversationID: cid})
		assert.Equal(t, handlers.ErrConversationNotFound, err)
	})
}

func TestHandleListConversations(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("by session id", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cr.On("FindBySessionID", ctx, session.ID()).Return([]*aggregates.Conversation{conv}, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		result, err := h.HandleListConversations(ctx, &queries.ListConversationsQuery{SessionID: session.ID()})
		require.NoError(t, err)
		assert.Len(t, result.Conversations, 1)
	})

	t.Run("active only", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cr.On("FindActive", ctx).Return([]*aggregates.Conversation{conv}, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		result, err := h.HandleListConversations(ctx, &queries.ListConversationsQuery{ActiveOnly: true})
		require.NoError(t, err)
		assert.Len(t, result.Conversations, 1)
	})

	t.Run("with pagination limit", func(t *testing.T) {
		cr := new(mockConversationRepo)
		conv2, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr.On("FindBySessionID", ctx, session.ID()).Return([]*aggregates.Conversation{conv, conv2}, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		result, err := h.HandleListConversations(ctx, &queries.ListConversationsQuery{
			SessionID: session.ID(), Limit: 1,
		})
		require.NoError(t, err)
		assert.Len(t, result.Conversations, 1)
	})
}

func TestHandleGetConversationMessages(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)
	_, _ = conv.AddUserMessage("Hello")

	t.Run("success", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		msgs, err := h.HandleGetConversationMessages(ctx, &queries.GetConversationMessagesQuery{
			ConversationID: conv.ID(),
		})
		require.NoError(t, err)
		assert.Len(t, msgs, 1)
	})

	t.Run("not found", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		_, err := h.HandleGetConversationMessages(ctx, &queries.GetConversationMessagesQuery{ConversationID: cid})
		assert.Equal(t, handlers.ErrConversationNotFound, err)
	})

	t.Run("with offset and limit", func(t *testing.T) {
		conv2, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		_, _ = conv2.AddUserMessage("msg1")
		_, _ = conv2.AddAssistantMessage(nil)
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, conv2.ID()).Return(conv2, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		msgs, err := h.HandleGetConversationMessages(ctx, &queries.GetConversationMessagesQuery{
			ConversationID: conv2.ID(), Offset: 1, Limit: 1,
		})
		require.NoError(t, err)
		assert.Len(t, msgs, 1)
	})
}

func TestHandleSendMessage_Streaming(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("streaming success", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cs := new(mockClaudeSvc)
		pub := new(mockEventPublisher)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)

		eventCh := make(chan *services.ClaudeStreamEvent, 3)
		eventCh <- &services.ClaudeStreamEvent{Message: &services.ClaudeResponse{ID: "resp1"}}
		eventCh <- &services.ClaudeStreamEvent{ContentBlock: &entities.ContentBlock{Type: vo.ContentTypeText, Text: "hello"}}
		close(eventCh)
		cs.On("CreateMessageStream", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return((<-chan *services.ClaudeStreamEvent)(eventCh), nil)

		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, pub)
		result, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv.ID(), Content: "hi", Stream: true,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "resp1", result.Response.ID)
	})

	t.Run("streaming error", func(t *testing.T) {
		conv2, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr := new(mockConversationRepo)
		cs := new(mockClaudeSvc)
		cr.On("FindByID", ctx, conv2.ID()).Return(conv2, nil)

		eventCh := make(chan *services.ClaudeStreamEvent, 1)
		eventCh <- &services.ClaudeStreamEvent{Error: errors.New("stream error")}
		close(eventCh)
		cs.On("CreateMessageStream", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return((<-chan *services.ClaudeStreamEvent)(eventCh), nil)

		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv2.ID(), Content: "hi", Stream: true,
		})
		assert.Error(t, err)
	})

	t.Run("streaming create error", func(t *testing.T) {
		conv3, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr := new(mockConversationRepo)
		cs := new(mockClaudeSvc)
		cr.On("FindByID", ctx, conv3.ID()).Return(conv3, nil)
		cs.On("CreateMessageStream", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return((<-chan *services.ClaudeStreamEvent)(nil), errors.New("conn refused"))

		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv3.ID(), Content: "hi", Stream: true,
		})
		assert.Error(t, err)
	})
}

func TestHandleAddToolResult(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("success", func(t *testing.T) {
		cr := new(mockConversationRepo)
		pub := new(mockEventPublisher)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), pub)

		err := h.HandleAddToolResult(ctx, &commands.AddToolResultCommand{
			ConversationID: conv.ID(), ToolUseID: "tool_123", Content: "result", IsError: false,
		})
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		err := h.HandleAddToolResult(ctx, &commands.AddToolResultCommand{
			ConversationID: cid, ToolUseID: "tool_123", Content: "result",
		})
		assert.Equal(t, handlers.ErrConversationNotFound, err)
	})

	t.Run("repo error", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(errors.New("db"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))

		err := h.HandleAddToolResult(ctx, &commands.AddToolResultCommand{
			ConversationID: conv.ID(), ToolUseID: "tool_123", Content: "result",
		})
		assert.Error(t, err)
	})
}

func TestHandleCreateConversation_AdditionalPaths(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()

	t.Run("session repo find error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		sr.On("FindByID", ctx, session.ID()).Return(nil, errors.New("db"))
		h := handlers.NewConversationHandler(sr, new(mockConversationRepo), new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: session.ID(), Model: vo.ModelClaudeOpus47,
		})
		assert.Error(t, err)
	})

	t.Run("session repo save error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		cr := new(mockConversationRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(errors.New("db"))
		h := handlers.NewConversationHandler(sr, cr, new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: session.ID(), Model: vo.ModelClaudeOpus47,
		})
		assert.Error(t, err)
	})

	t.Run("conversation repo save error", func(t *testing.T) {
		sr := new(mockSessionRepo)
		cr := new(mockConversationRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(errors.New("db"))
		h := handlers.NewConversationHandler(sr, cr, new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: session.ID(), Model: vo.ModelClaudeOpus47,
		})
		assert.Error(t, err)
	})

	t.Run("negative temperature", func(t *testing.T) {
		sr := new(mockSessionRepo)
		cr := new(mockConversationRepo)
		pub := new(mockEventPublisher)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		sr.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(sr, cr, new(mockClaudeSvc), pub)
		conv, err := h.HandleCreateConversation(ctx, &commands.CreateConversationCommand{
			SessionID: session.ID(), Model: vo.ModelClaudeOpus47, Temperature: -1.0,
		})
		require.NoError(t, err)
		assert.NotNil(t, conv)
	})
}

func TestHandleSendMessage_AdditionalPaths(t *testing.T) {
	ctx := context.Background()
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("repo find error", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(nil, errors.New("db"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv.ID(), Content: "hi",
		})
		assert.Error(t, err)
	})

	t.Run("claude error", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cs := new(mockClaudeSvc)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cs.On("CreateMessage", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return(nil, errors.New("api"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv.ID(), Content: "hi",
		})
		assert.Error(t, err)
	})

	t.Run("closed conversation", func(t *testing.T) {
		closedConv, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		closedConv.Close()
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, closedConv.ID()).Return(closedConv, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: closedConv.ID(), Content: "hi",
		})
		assert.Error(t, err)
	})

	t.Run("save error after message", func(t *testing.T) {
		conv2, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr := new(mockConversationRepo)
		cs := new(mockClaudeSvc)
		cr.On("FindByID", ctx, conv2.ID()).Return(conv2, nil)
		cs.On("CreateMessage", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return(&services.ClaudeResponse{
			Content: []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "ok"}},
		}, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(errors.New("db"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, new(mockEventPublisher))
		_, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv2.ID(), Content: "hi",
		})
		assert.Error(t, err)
	})

	t.Run("tool use response", func(t *testing.T) {
		conv3, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr := new(mockConversationRepo)
		cs := new(mockClaudeSvc)
		pub := new(mockEventPublisher)
		cr.On("FindByID", ctx, conv3.ID()).Return(conv3, nil)
		cs.On("CreateMessage", ctx, mock.AnythingOfType("*services.ClaudeRequest")).Return(&services.ClaudeResponse{
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "result"},
				{Type: vo.ContentTypeToolUse, Text: "tool call"},
			},
		}, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, cs, pub)
		resp, err := h.HandleSendMessage(ctx, &commands.SendMessageCommand{
			ConversationID: conv3.ID(), Content: "use tool",
		})
		require.NoError(t, err)
		assert.True(t, resp.HasToolUse)
		assert.Len(t, resp.ToolUses, 1)
	})
}

func TestHandleCloseConversation_AdditionalPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("repo find error", func(t *testing.T) {
		cr := new(mockConversationRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, errors.New("db"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
		err := h.HandleCloseConversation(ctx, &commands.CloseConversationCommand{ConversationID: cid})
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		session := createInitializedSession()
		conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)
		cr := new(mockConversationRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(errors.New("db"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
		err := h.HandleCloseConversation(ctx, &commands.CloseConversationCommand{ConversationID: conv.ID()})
		assert.Error(t, err)
	})
}

func TestHandleGetConversation_RepoError(t *testing.T) {
	ctx := context.Background()
	cr := new(mockConversationRepo)
	cid := vo.GenerateConversationID()
	cr.On("FindByID", ctx, cid).Return(nil, errors.New("db"))
	h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
	_, err := h.HandleGetConversation(ctx, &queries.GetConversationQuery{ConversationID: cid})
	assert.Error(t, err)
}

func TestHandleListConversations_AdditionalPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("all conversations no session no active", func(t *testing.T) {
		cr := new(mockConversationRepo)
		conv, _ := createInitializedSession().CreateConversation(vo.ModelClaudeOpus47)
		cr.On("FindActive", ctx).Return([]*aggregates.Conversation{conv}, nil)
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
		result, err := h.HandleListConversations(ctx, &queries.ListConversationsQuery{})
		require.NoError(t, err)
		assert.Len(t, result.Conversations, 1)
	})

	t.Run("repo error", func(t *testing.T) {
		cr := new(mockConversationRepo)
		sid := vo.GenerateSessionID()
		cr.On("FindBySessionID", ctx, sid).Return(nil, errors.New("db"))
		h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
		_, err := h.HandleListConversations(ctx, &queries.ListConversationsQuery{SessionID: sid})
		assert.Error(t, err)
	})
}

func TestHandleGetConversationMessages_RepoError(t *testing.T) {
	ctx := context.Background()
	cr := new(mockConversationRepo)
	cid := vo.GenerateConversationID()
	cr.On("FindByID", ctx, cid).Return(nil, errors.New("db"))
	h := handlers.NewConversationHandler(new(mockSessionRepo), cr, new(mockClaudeSvc), new(mockEventPublisher))
	_, err := h.HandleGetConversationMessages(ctx, &queries.GetConversationMessagesQuery{ConversationID: cid})
	assert.Error(t, err)
}
