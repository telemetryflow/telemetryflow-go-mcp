package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	appsvc "github.com/telemetryflow/telemetryflow-go-mcp/internal/application/services"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type mockSvcSessionRepo struct {
	mock.Mock
}

func (m *mockSvcSessionRepo) Save(ctx context.Context, s *aggregates.Session) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockSvcSessionRepo) FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.Session), args.Error(1)
}
func (m *mockSvcSessionRepo) FindAll(ctx context.Context) ([]*aggregates.Session, error) {
	return nil, nil
}
func (m *mockSvcSessionRepo) FindActive(ctx context.Context) ([]*aggregates.Session, error) {
	return nil, nil
}
func (m *mockSvcSessionRepo) Delete(ctx context.Context, id vo.SessionID) error { return nil }
func (m *mockSvcSessionRepo) Exists(ctx context.Context, id vo.SessionID) (bool, error) {
	return false, nil
}
func (m *mockSvcSessionRepo) Count(ctx context.Context) (int, error) { return 0, nil }

type mockSvcConvRepo struct {
	mock.Mock
}

func (m *mockSvcConvRepo) Save(ctx context.Context, c *aggregates.Conversation) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockSvcConvRepo) FindByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.Conversation), args.Error(1)
}
func (m *mockSvcConvRepo) FindBySessionID(ctx context.Context, sid vo.SessionID) ([]*aggregates.Conversation, error) {
	args := m.Called(ctx, sid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Conversation), args.Error(1)
}
func (m *mockSvcConvRepo) FindActive(ctx context.Context) ([]*aggregates.Conversation, error) {
	return nil, nil
}
func (m *mockSvcConvRepo) Delete(ctx context.Context, id vo.ConversationID) error { return nil }
func (m *mockSvcConvRepo) Exists(ctx context.Context, id vo.ConversationID) (bool, error) {
	return false, nil
}
func (m *mockSvcConvRepo) Count(ctx context.Context) (int, error) { return 0, nil }
func (m *mockSvcConvRepo) CountBySessionID(ctx context.Context, sid vo.SessionID) (int, error) {
	return 0, nil
}

func createSvcSession() *aggregates.Session {
	s := aggregates.NewSession()
	_ = s.Initialize(&aggregates.ClientInfo{Name: "test", Version: "1.0"}, "2024-11-05")
	s.MarkReady()
	return s
}

func TestNewConversationService(t *testing.T) {
	svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
	assert.NotNil(t, svc)
}

type mockSvcClaude struct{ mock.Mock }

func (m *mockSvcClaude) CreateMessage(ctx context.Context, req *services.ClaudeRequest) (*services.ClaudeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ClaudeResponse), args.Error(1)
}
func (m *mockSvcClaude) CreateMessageStream(ctx context.Context, req *services.ClaudeRequest) (<-chan *services.ClaudeStreamEvent, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *services.ClaudeStreamEvent), args.Error(1)
}
func (m *mockSvcClaude) CountTokens(ctx context.Context, req *services.ClaudeRequest) (int, error) {
	args := m.Called(ctx, req)
	return args.Int(0), args.Error(1)
}
func (m *mockSvcClaude) ValidateRequest(req *services.ClaudeRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestStartConversation(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()

	t.Run("success", func(t *testing.T) {
		sr := new(mockSvcSessionRepo)
		cr := new(mockSvcConvRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		svc := appsvc.NewConversationService(sr, cr, new(mockSvcClaude))

		dto, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{SessionID: session.ID().String()})
		require.NoError(t, err)
		assert.Equal(t, session.ID().String(), dto.SessionID)
	})

	t.Run("empty session id", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{SessionID: ""})
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})

	t.Run("session not found", func(t *testing.T) {
		sr := new(mockSvcSessionRepo)
		sid := vo.GenerateSessionID()
		sr.On("FindByID", ctx, sid).Return(nil, nil)
		svc := appsvc.NewConversationService(sr, new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{SessionID: sid.String()})
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})
}

func TestSendMessage(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("success", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))

		resp, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: conv.ID().String(), Content: "Hello"})
		require.NoError(t, err)
		assert.False(t, resp.HasToolUse)
	})

	t.Run("empty content", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: conv.ID().String(), Content: ""})
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})

	t.Run("empty conversation id", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: "", Content: "hi"})
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})

	t.Run("conversation not found", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		_, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: cid.String(), Content: "hi"})
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})
}

func TestGetConversation(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("success", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))

		dto, err := svc.GetConversation(ctx, conv.ID().String())
		require.NoError(t, err)
		assert.Equal(t, conv.ID().String(), dto.ID)
	})

	t.Run("not found", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		_, err := svc.GetConversation(ctx, cid.String())
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})
}

func TestListConversations(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("success", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cr.On("FindBySessionID", ctx, session.ID()).Return([]*aggregates.Conversation{conv}, nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))

		dtos, err := svc.ListConversations(ctx, session.ID().String())
		require.NoError(t, err)
		assert.Len(t, dtos, 1)
	})
}

func TestCloseConversation(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("success", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))

		err := svc.CloseConversation(ctx, conv.ID().String())
		require.NoError(t, err)
		assert.False(t, conv.IsActive())
	})

	t.Run("not found", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		err := svc.CloseConversation(ctx, cid.String())
		assert.Equal(t, appsvc.ErrInvalidInput, err)
	})
}

func TestStartConversation_AdditionalPaths(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()

	t.Run("repo find error", func(t *testing.T) {
		sr := new(mockSvcSessionRepo)
		sid := vo.GenerateSessionID()
		sr.On("FindByID", ctx, sid).Return(nil, fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(sr, new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{SessionID: sid.String()})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("repo save error", func(t *testing.T) {
		sr := new(mockSvcSessionRepo)
		cr := new(mockSvcConvRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(sr, cr, new(mockSvcClaude))
		_, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{SessionID: session.ID().String()})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("custom model", func(t *testing.T) {
		sr := new(mockSvcSessionRepo)
		cr := new(mockSvcConvRepo)
		sr.On("FindByID", ctx, session.ID()).Return(session, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(nil)
		svc := appsvc.NewConversationService(sr, cr, new(mockSvcClaude))
		dto, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{
			SessionID: session.ID().String(),
			Model:     "claude-opus-4-7",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dto.Model != "claude-opus-4-7" {
			t.Errorf("expected model claude-opus-4-7, got %s", dto.Model)
		}
	})

	t.Run("invalid session id format", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.StartConversation(ctx, &appsvc.StartConversationRequest{SessionID: "not-a-uuid"})
		if err == nil {
			t.Error("expected error for invalid session ID")
		}
	})
}

func TestSendMessage_AdditionalPaths(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("repo find error", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		_, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: cid.String(), Content: "hi"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("repo save error", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		_, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: conv.ID().String(), Content: "hi"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("invalid conversation id format", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.SendMessage(ctx, &appsvc.SendMessageRequest{ConversationID: "not-a-uuid", Content: "hi"})
		if err == nil {
			t.Error("expected error for invalid conversation ID")
		}
	})
}

func TestGetConversation_AdditionalPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("repo find error", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		_, err := svc.GetConversation(ctx, cid.String())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("invalid id format", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.GetConversation(ctx, "not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid ID")
		}
	})
}

func TestListConversations_AdditionalPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("repo error", func(t *testing.T) {
		sid := vo.GenerateSessionID()
		cr := new(mockSvcConvRepo)
		cr.On("FindBySessionID", ctx, sid).Return(nil, fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		_, err := svc.ListConversations(ctx, sid.String())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("invalid session id format", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		_, err := svc.ListConversations(ctx, "not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid session ID")
		}
	})

	t.Run("empty result", func(t *testing.T) {
		sid := vo.GenerateSessionID()
		cr := new(mockSvcConvRepo)
		cr.On("FindBySessionID", ctx, sid).Return([]*aggregates.Conversation{}, nil)
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		result, err := svc.ListConversations(ctx, sid.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d", len(result))
		}
	})
}

func TestCloseConversation_AdditionalPaths(t *testing.T) {
	ctx := context.Background()
	session := createSvcSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)

	t.Run("repo find error", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cid := vo.GenerateConversationID()
		cr.On("FindByID", ctx, cid).Return(nil, fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		err := svc.CloseConversation(ctx, cid.String())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("repo save error", func(t *testing.T) {
		cr := new(mockSvcConvRepo)
		cr.On("FindByID", ctx, conv.ID()).Return(conv, nil)
		cr.On("Save", ctx, mock.AnythingOfType("*aggregates.Conversation")).Return(fmt.Errorf("db error"))
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), cr, new(mockSvcClaude))
		err := svc.CloseConversation(ctx, conv.ID().String())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("invalid id format", func(t *testing.T) {
		svc := appsvc.NewConversationService(new(mockSvcSessionRepo), new(mockSvcConvRepo), new(mockSvcClaude))
		err := svc.CloseConversation(ctx, "not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid ID")
		}
	})
}
