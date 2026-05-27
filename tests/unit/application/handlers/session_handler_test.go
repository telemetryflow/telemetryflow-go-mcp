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
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type mockEventPublisher struct {
	mock.Mock
}

func (m *mockEventPublisher) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type mockSessionRepo struct {
	mock.Mock
}

func (m *mockSessionRepo) Save(ctx context.Context, s *aggregates.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}
func (m *mockSessionRepo) FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.Session), args.Error(1)
}
func (m *mockSessionRepo) FindAll(ctx context.Context) ([]*aggregates.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Session), args.Error(1)
}
func (m *mockSessionRepo) FindActive(ctx context.Context) ([]*aggregates.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Session), args.Error(1)
}
func (m *mockSessionRepo) Delete(ctx context.Context, id vo.SessionID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockSessionRepo) Exists(ctx context.Context, id vo.SessionID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockSessionRepo) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func createInitializedSession() *aggregates.Session {
	s := aggregates.NewSession()
	_ = s.Initialize(&aggregates.ClientInfo{Name: "test", Version: "1.0"}, "2024-11-05")
	s.MarkReady()
	return s
}

func TestNewSessionHandler(t *testing.T) {
	repo := new(mockSessionRepo)
	pub := new(mockEventPublisher)
	h := handlers.NewSessionHandler(repo, pub)
	assert.NotNil(t, h)
}

func TestHandleInitializeSession(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		repo.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewSessionHandler(repo, pub)

		session, err := h.HandleInitializeSession(ctx, &commands.InitializeSessionCommand{
			ClientName:      "TestClient",
			ClientVersion:   "1.0.0",
			ProtocolVersion: "2024-11-05",
		})
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, aggregates.SessionStateReady, session.State())
		repo.AssertExpectations(t)
	})

	t.Run("save error", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		repo.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(errors.New("db error"))
		h := handlers.NewSessionHandler(repo, pub)

		_, err := h.HandleInitializeSession(ctx, &commands.InitializeSessionCommand{
			ClientName: "TestClient", ClientVersion: "1.0", ProtocolVersion: "2024-11-05",
		})
		assert.Error(t, err)
	})
}

func TestHandleCloseSession(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		session := createInitializedSession()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		repo.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		pub.On("Publish", ctx, mock.Anything).Return(nil)
		h := handlers.NewSessionHandler(repo, pub)

		err := h.HandleCloseSession(ctx, &commands.CloseSessionCommand{SessionID: session.ID()})
		require.NoError(t, err)
		assert.Equal(t, aggregates.SessionStateClosed, session.State())
	})

	t.Run("session not found", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		sid := vo.GenerateSessionID()
		repo.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewSessionHandler(repo, pub)

		err := h.HandleCloseSession(ctx, &commands.CloseSessionCommand{SessionID: sid})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("repo error on find", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		sid := vo.GenerateSessionID()
		repo.On("FindByID", ctx, sid).Return(nil, errors.New("db error"))
		h := handlers.NewSessionHandler(repo, pub)

		err := h.HandleCloseSession(ctx, &commands.CloseSessionCommand{SessionID: sid})
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		session := createInitializedSession()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		repo.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(errors.New("db error"))
		h := handlers.NewSessionHandler(repo, pub)

		err := h.HandleCloseSession(ctx, &commands.CloseSessionCommand{SessionID: session.ID()})
		assert.Error(t, err)
	})
}

func TestHandleSetLogLevel(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mockSessionRepo)
		pub := new(mockEventPublisher)
		session := createInitializedSession()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		repo.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(nil)
		h := handlers.NewSessionHandler(repo, pub)

		err := h.HandleSetLogLevel(ctx, &commands.SetLogLevelCommand{SessionID: session.ID(), Level: vo.LogLevelDebug})
		require.NoError(t, err)
		assert.Equal(t, vo.LogLevelDebug, session.LogLevel())
	})

	t.Run("session not found", func(t *testing.T) {
		repo := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		repo.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		err := h.HandleSetLogLevel(ctx, &commands.SetLogLevelCommand{SessionID: sid, Level: vo.LogLevelDebug})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("invalid log level", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		err := h.HandleSetLogLevel(ctx, &commands.SetLogLevelCommand{SessionID: session.ID(), Level: vo.MCPLogLevel("invalid")})
		assert.Error(t, err)
	})
}

func TestHandlePing(t *testing.T) {
	ctx := context.Background()

	t.Run("active session", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		err := h.HandlePing(ctx, &commands.PingCommand{SessionID: session.ID()})
		assert.NoError(t, err)
	})

	t.Run("session not found", func(t *testing.T) {
		repo := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		repo.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		err := h.HandlePing(ctx, &commands.PingCommand{SessionID: sid})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})

	t.Run("closed session", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		session.Close()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		err := h.HandlePing(ctx, &commands.PingCommand{SessionID: session.ID()})
		assert.Equal(t, aggregates.ErrSessionClosed, err)
	})
}

func TestHandleGetSession(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		result, err := h.HandleGetSession(ctx, &queries.GetSessionQuery{SessionID: session.ID()})
		require.NoError(t, err)
		assert.Equal(t, session.ID(), result.ID())
	})

	t.Run("not found", func(t *testing.T) {
		repo := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		repo.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		_, err := h.HandleGetSession(ctx, &queries.GetSessionQuery{SessionID: sid})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})
}

func TestHandleListSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("active only", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		repo.On("FindActive", ctx).Return([]*aggregates.Session{session}, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		result, err := h.HandleListSessions(ctx, &queries.ListSessionsQuery{ActiveOnly: true})
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("all sessions", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		repo.On("FindAll", ctx).Return([]*aggregates.Session{session}, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		result, err := h.HandleListSessions(ctx, &queries.ListSessionsQuery{ActiveOnly: false})
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestHandleGetSessionStats(t *testing.T) {
	ctx := context.Background()

	t.Run("success with conversations", func(t *testing.T) {
		repo := new(mockSessionRepo)
		session := createInitializedSession()
		_, _ = session.CreateConversation(vo.ModelClaudeOpus47)
		repo.On("FindByID", ctx, session.ID()).Return(session, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		stats, err := h.HandleGetSessionStats(ctx, &queries.GetSessionStatsQuery{SessionID: session.ID()})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.TotalConversations)
	})

	t.Run("session not found", func(t *testing.T) {
		repo := new(mockSessionRepo)
		sid := vo.GenerateSessionID()
		repo.On("FindByID", ctx, sid).Return(nil, nil)
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))

		_, err := h.HandleGetSessionStats(ctx, &queries.GetSessionStatsQuery{SessionID: sid})
		assert.Equal(t, handlers.ErrSessionNotFound, err)
	})
}

func TestHandleCloseSession_RepoFindError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockSessionRepo)
	sid := vo.GenerateSessionID()
	repo.On("FindByID", ctx, sid).Return(nil, errors.New("db"))
	h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
	err := h.HandleCloseSession(ctx, &commands.CloseSessionCommand{SessionID: sid})
	assert.Error(t, err)
}

func TestHandleSetLogLevel_SaveError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockSessionRepo)
	session := createInitializedSession()
	repo.On("FindByID", ctx, session.ID()).Return(session, nil)
	repo.On("Save", ctx, mock.AnythingOfType("*aggregates.Session")).Return(errors.New("db"))
	h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
	err := h.HandleSetLogLevel(ctx, &commands.SetLogLevelCommand{SessionID: session.ID(), Level: vo.LogLevelDebug})
	assert.Error(t, err)
}

func TestHandleGetSession_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockSessionRepo)
	sid := vo.GenerateSessionID()
	repo.On("FindByID", ctx, sid).Return(nil, errors.New("db"))
	h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
	_, err := h.HandleGetSession(ctx, &queries.GetSessionQuery{SessionID: sid})
	assert.Error(t, err)
}

func TestHandleListSessions_RepoErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("active error", func(t *testing.T) {
		repo := new(mockSessionRepo)
		repo.On("FindActive", ctx).Return(nil, errors.New("db"))
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
		_, err := h.HandleListSessions(ctx, &queries.ListSessionsQuery{ActiveOnly: true})
		assert.Error(t, err)
	})

	t.Run("all error", func(t *testing.T) {
		repo := new(mockSessionRepo)
		repo.On("FindAll", ctx).Return(nil, errors.New("db"))
		h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
		_, err := h.HandleListSessions(ctx, &queries.ListSessionsQuery{ActiveOnly: false})
		assert.Error(t, err)
	})
}

func TestHandleGetSessionStats_WithConversations(t *testing.T) {
	ctx := context.Background()
	repo := new(mockSessionRepo)
	session := createInitializedSession()
	conv, _ := session.CreateConversation(vo.ModelClaudeOpus47)
	_, _ = conv.AddUserMessage("hello")
	conv2, _ := session.CreateConversation(vo.ModelClaudeOpus47)
	conv2.Close()
	repo.On("FindByID", ctx, session.ID()).Return(session, nil)
	h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
	stats, err := h.HandleGetSessionStats(ctx, &queries.GetSessionStatsQuery{SessionID: session.ID()})
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalConversations)
	assert.Equal(t, 1, stats.ActiveConversations)
}

func TestHandleGetSessionStats_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockSessionRepo)
	sid := vo.GenerateSessionID()
	repo.On("FindByID", ctx, sid).Return(nil, errors.New("db"))
	h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
	_, err := h.HandleGetSessionStats(ctx, &queries.GetSessionStatsQuery{SessionID: sid})
	assert.Error(t, err)
}

func TestHandlePing_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockSessionRepo)
	sid := vo.GenerateSessionID()
	repo.On("FindByID", ctx, sid).Return(nil, errors.New("db"))
	h := handlers.NewSessionHandler(repo, new(mockEventPublisher))
	err := h.HandlePing(ctx, &commands.PingCommand{SessionID: sid})
	assert.Error(t, err)
}
