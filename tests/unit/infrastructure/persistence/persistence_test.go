package persistence_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
)

func TestInMemorySessionRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemorySessionRepository()

	t.Run("save and find by id", func(t *testing.T) {
		session := aggregates.NewSession()
		_ = session.Initialize(&aggregates.ClientInfo{Name: "test", Version: "1.0"}, "2024-11-05")
		session.MarkReady()

		err := repo.Save(ctx, session)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, session.ID())
		require.NoError(t, err)
		assert.Equal(t, session.ID(), found.ID())
		assert.Equal(t, aggregates.SessionStateReady, found.State())
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		sid := vo.GenerateSessionID()
		found, err := repo.FindByID(ctx, sid)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("find all", func(t *testing.T) {
		repo := persistence.NewInMemorySessionRepository()
		s1 := aggregates.NewSession()
		s2 := aggregates.NewSession()
		_ = repo.Save(ctx, s1)
		_ = repo.Save(ctx, s2)

		all, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, all, 2)
	})

	t.Run("find active", func(t *testing.T) {
		repo := persistence.NewInMemorySessionRepository()
		active := aggregates.NewSession()
		_ = active.Initialize(&aggregates.ClientInfo{Name: "a", Version: "1.0"}, "2024-11-05")
		active.MarkReady()

		closed := aggregates.NewSession()
		_ = closed.Initialize(&aggregates.ClientInfo{Name: "b", Version: "1.0"}, "2024-11-05")
		closed.MarkReady()
		closed.Close()

		_ = repo.Save(ctx, active)
		_ = repo.Save(ctx, closed)

		actives, err := repo.FindActive(ctx)
		require.NoError(t, err)
		assert.Len(t, actives, 1)
		assert.Equal(t, active.ID(), actives[0].ID())
	})

	t.Run("delete", func(t *testing.T) {
		repo := persistence.NewInMemorySessionRepository()
		s := aggregates.NewSession()
		_ = repo.Save(ctx, s)
		err := repo.Delete(ctx, s.ID())
		require.NoError(t, err)
		found, _ := repo.FindByID(ctx, s.ID())
		assert.Nil(t, found)
	})

	t.Run("exists", func(t *testing.T) {
		repo := persistence.NewInMemorySessionRepository()
		s := aggregates.NewSession()
		_ = repo.Save(ctx, s)
		ok, err := repo.Exists(ctx, s.ID())
		require.NoError(t, err)
		assert.True(t, ok)

		ok, err = repo.Exists(ctx, vo.GenerateSessionID())
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("count", func(t *testing.T) {
		repo := persistence.NewInMemorySessionRepository()
		_ = repo.Save(ctx, aggregates.NewSession())
		_ = repo.Save(ctx, aggregates.NewSession())
		n, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
	})
}

func TestInMemoryConversationRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryConversationRepository()
	sessionID := vo.GenerateSessionID()

	t.Run("save and find by id", func(t *testing.T) {
		conv := aggregates.NewConversation(sessionID, vo.ModelClaudeOpus47)
		err := repo.Save(ctx, conv)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, conv.ID())
		require.NoError(t, err)
		assert.Equal(t, conv.ID(), found.ID())
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		found, err := repo.FindByID(ctx, vo.GenerateConversationID())
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("find by session id", func(t *testing.T) {
		repo := persistence.NewInMemoryConversationRepository()
		sid1 := vo.GenerateSessionID()
		sid2 := vo.GenerateSessionID()
		c1 := aggregates.NewConversation(sid1, vo.ModelClaudeOpus47)
		c2 := aggregates.NewConversation(sid1, vo.ModelClaudeSonnet4)
		c3 := aggregates.NewConversation(sid2, vo.ModelClaudeOpus47)
		_ = repo.Save(ctx, c1)
		_ = repo.Save(ctx, c2)
		_ = repo.Save(ctx, c3)

		convs, err := repo.FindBySessionID(ctx, sid1)
		require.NoError(t, err)
		assert.Len(t, convs, 2)
	})

	t.Run("find active", func(t *testing.T) {
		repo := persistence.NewInMemoryConversationRepository()
		active := aggregates.NewConversation(sessionID, vo.ModelClaudeOpus47)
		closed := aggregates.NewConversation(sessionID, vo.ModelClaudeSonnet4)
		closed.Close()
		_ = repo.Save(ctx, active)
		_ = repo.Save(ctx, closed)

		actives, err := repo.FindActive(ctx)
		require.NoError(t, err)
		assert.Len(t, actives, 1)
	})

	t.Run("delete", func(t *testing.T) {
		repo := persistence.NewInMemoryConversationRepository()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaudeOpus47)
		_ = repo.Save(ctx, conv)
		err := repo.Delete(ctx, conv.ID())
		require.NoError(t, err)
		found, _ := repo.FindByID(ctx, conv.ID())
		assert.Nil(t, found)
	})

	t.Run("exists", func(t *testing.T) {
		repo := persistence.NewInMemoryConversationRepository()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaudeOpus47)
		_ = repo.Save(ctx, conv)
		ok, err := repo.Exists(ctx, conv.ID())
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("count", func(t *testing.T) {
		repo := persistence.NewInMemoryConversationRepository()
		_ = repo.Save(ctx, aggregates.NewConversation(sessionID, vo.ModelClaudeOpus47))
		_ = repo.Save(ctx, aggregates.NewConversation(sessionID, vo.ModelClaudeSonnet4))
		n, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
	})

	t.Run("count by session id", func(t *testing.T) {
		repo := persistence.NewInMemoryConversationRepository()
		sid1 := vo.GenerateSessionID()
		sid2 := vo.GenerateSessionID()
		_ = repo.Save(ctx, aggregates.NewConversation(sid1, vo.ModelClaudeOpus47))
		_ = repo.Save(ctx, aggregates.NewConversation(sid1, vo.ModelClaudeSonnet4))
		_ = repo.Save(ctx, aggregates.NewConversation(sid2, vo.ModelClaudeOpus47))
		n, err := repo.CountBySessionID(ctx, sid1)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
	})
}

func TestInMemoryToolRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryToolRepository()

	createTool := func(t *testing.T, name string) *entities.Tool {
		t.Helper()
		tn, err := vo.NewToolName(name)
		require.NoError(t, err)
		td, err := vo.NewToolDescription("desc for " + name)
		require.NoError(t, err)
		tool, err := entities.NewTool(tn, td, nil)
		require.NoError(t, err)
		return tool
	}

	t.Run("register and find by name", func(t *testing.T) {
		tool := createTool(t, "test_tool")
		err := repo.Register(ctx, tool)
		require.NoError(t, err)

		tn, _ := vo.NewToolName("test_tool")
		found, err := repo.FindByName(ctx, tn)
		require.NoError(t, err)
		assert.Equal(t, "test_tool", found.Name().String())
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		tn, _ := vo.NewToolName("missing")
		found, err := repo.FindByName(ctx, tn)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("find all", func(t *testing.T) {
		repo := persistence.NewInMemoryToolRepository()
		_ = repo.Register(ctx, createTool(t, "a"))
		_ = repo.Register(ctx, createTool(t, "b"))
		all, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, all, 2)
	})

	t.Run("find by category", func(t *testing.T) {
		repo := persistence.NewInMemoryToolRepository()
		t1 := createTool(t, "cat1_tool")
		t1.SetCategory("utility")
		t2 := createTool(t, "cat2_tool")
		t2.SetCategory("system")
		_ = repo.Register(ctx, t1)
		_ = repo.Register(ctx, t2)

		found, err := repo.FindByCategory(ctx, "utility")
		require.NoError(t, err)
		assert.Len(t, found, 1)
	})

	t.Run("find by tag", func(t *testing.T) {
		repo := persistence.NewInMemoryToolRepository()
		t1 := createTool(t, "tagged_tool")
		t1.SetTags([]string{"test", "unit"})
		_ = repo.Register(ctx, t1)

		found, err := repo.FindByTag(ctx, "test")
		require.NoError(t, err)
		assert.Len(t, found, 1)

		found2, err := repo.FindByTag(ctx, "missing")
		require.NoError(t, err)
		assert.Len(t, found2, 0)
	})

	t.Run("find enabled", func(t *testing.T) {
		repo := persistence.NewInMemoryToolRepository()
		t1 := createTool(t, "enabled_t")
		t2 := createTool(t, "disabled_t")
		t2.Disable()
		_ = repo.Register(ctx, t1)
		_ = repo.Register(ctx, t2)

		found, err := repo.FindEnabled(ctx)
		require.NoError(t, err)
		assert.Len(t, found, 1)
	})

	t.Run("unregister", func(t *testing.T) {
		repo := persistence.NewInMemoryToolRepository()
		_ = repo.Register(ctx, createTool(t, "rem"))
		tn, _ := vo.NewToolName("rem")
		err := repo.Unregister(ctx, tn)
		require.NoError(t, err)
		found, _ := repo.FindByName(ctx, tn)
		assert.Nil(t, found)
	})

	t.Run("exists", func(t *testing.T) {
		_ = repo.Register(ctx, createTool(t, "exists_check"))
		tn, _ := vo.NewToolName("exists_check")
		ok, _ := repo.Exists(ctx, tn)
		assert.True(t, ok)
		tn2, _ := vo.NewToolName("nope")
		ok2, _ := repo.Exists(ctx, tn2)
		assert.False(t, ok2)
	})

	t.Run("count", func(t *testing.T) {
		repo := persistence.NewInMemoryToolRepository()
		_ = repo.Register(ctx, createTool(t, "c1"))
		_ = repo.Register(ctx, createTool(t, "c2"))
		n, _ := repo.Count(ctx)
		assert.Equal(t, 2, n)
	})
}

func TestInMemoryResourceRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryResourceRepository()

	t.Run("register and find by uri", func(t *testing.T) {
		uri, _ := vo.NewResourceURI("file:///test")
		res, err := entities.NewResource(uri, "Test")
		require.NoError(t, err)
		err = repo.Register(ctx, res)
		require.NoError(t, err)

		found, err := repo.FindByURI(ctx, uri)
		require.NoError(t, err)
		assert.Equal(t, "Test", found.Name())
	})

	t.Run("register template", func(t *testing.T) {
		repo := persistence.NewInMemoryResourceRepository()
		tmpl, err := entities.NewResourceTemplate("file:///{path}", "File", "desc")
		require.NoError(t, err)
		err = repo.Register(ctx, tmpl)
		require.NoError(t, err)
		found, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, found, 1)
	})

	t.Run("find all", func(t *testing.T) {
		repo := persistence.NewInMemoryResourceRepository()
		uri1, _ := vo.NewResourceURI("file:///a")
		uri2, _ := vo.NewResourceURI("file:///b")
		r1, _ := entities.NewResource(uri1, "A")
		r2, _ := entities.NewResource(uri2, "B")
		_ = repo.Register(ctx, r1)
		_ = repo.Register(ctx, r2)
		all, _ := repo.FindAll(ctx)
		assert.Len(t, all, 2)
	})

	t.Run("find templates", func(t *testing.T) {
		repo := persistence.NewInMemoryResourceRepository()
		uri, _ := vo.NewResourceURI("file:///regular")
		regular, _ := entities.NewResource(uri, "R")
		tmpl, _ := entities.NewResourceTemplate("file:///{path}", "T", "")
		_ = repo.Register(ctx, regular)
		_ = repo.Register(ctx, tmpl)
		templates, _ := repo.FindTemplates(ctx)
		assert.Len(t, templates, 1)
	})

	t.Run("unregister", func(t *testing.T) {
		repo := persistence.NewInMemoryResourceRepository()
		uri, _ := vo.NewResourceURI("file:///del")
		r, _ := entities.NewResource(uri, "Del")
		_ = repo.Register(ctx, r)
		err := repo.Unregister(ctx, uri)
		require.NoError(t, err)
		found, _ := repo.FindByURI(ctx, uri)
		assert.Nil(t, found)
	})

	t.Run("exists", func(t *testing.T) {
		uri, _ := vo.NewResourceURI("file:///ex")
		r, _ := entities.NewResource(uri, "Ex")
		_ = repo.Register(ctx, r)
		ok, _ := repo.Exists(ctx, uri)
		assert.True(t, ok)
	})

	t.Run("count", func(t *testing.T) {
		repo := persistence.NewInMemoryResourceRepository()
		uri1, _ := vo.NewResourceURI("file:///c1")
		uri2, _ := vo.NewResourceURI("file:///c2")
		_ = repo.Register(ctx, mustNewResource(uri1, "C1"))
		_ = repo.Register(ctx, mustNewResource(uri2, "C2"))
		n, _ := repo.Count(ctx)
		assert.Equal(t, 2, n)
	})
}

func TestInMemoryPromptRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryPromptRepository()

	t.Run("register and find by name", func(t *testing.T) {
		pn, _ := vo.NewToolName("test_prompt")
		p, err := entities.NewPrompt(pn, "desc")
		require.NoError(t, err)
		err = repo.Register(ctx, p)
		require.NoError(t, err)

		found, err := repo.FindByName(ctx, pn)
		require.NoError(t, err)
		assert.Equal(t, "test_prompt", found.Name().String())
	})

	t.Run("find all", func(t *testing.T) {
		repo := persistence.NewInMemoryPromptRepository()
		pn1, _ := vo.NewToolName("p1")
		pn2, _ := vo.NewToolName("p2")
		p1, _ := entities.NewPrompt(pn1, "d1")
		p2, _ := entities.NewPrompt(pn2, "d2")
		_ = repo.Register(ctx, p1)
		_ = repo.Register(ctx, p2)
		all, _ := repo.FindAll(ctx)
		assert.Len(t, all, 2)
	})

	t.Run("unregister", func(t *testing.T) {
		repo := persistence.NewInMemoryPromptRepository()
		pn, _ := vo.NewToolName("del_prompt")
		p, _ := entities.NewPrompt(pn, "d")
		_ = repo.Register(ctx, p)
		err := repo.Unregister(ctx, pn)
		require.NoError(t, err)
		found, _ := repo.FindByName(ctx, pn)
		assert.Nil(t, found)
	})

	t.Run("exists", func(t *testing.T) {
		pn, _ := vo.NewToolName("ex_prompt")
		p, _ := entities.NewPrompt(pn, "d")
		_ = repo.Register(ctx, p)
		ok, _ := repo.Exists(ctx, pn)
		assert.True(t, ok)
	})

	t.Run("count", func(t *testing.T) {
		repo := persistence.NewInMemoryPromptRepository()
		pn1, _ := vo.NewToolName("cnt1")
		pn2, _ := vo.NewToolName("cnt2")
		_ = repo.Register(ctx, mustNewPrompt(pn1, "d"))
		_ = repo.Register(ctx, mustNewPrompt(pn2, "d"))
		n, _ := repo.Count(ctx)
		assert.Equal(t, 2, n)
	})
}

func TestDatabaseConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := persistence.DefaultDatabaseConfig()
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, "telemetryflow_mcp", cfg.Database)
	})

	t.Run("DSN", func(t *testing.T) {
		cfg := persistence.DefaultDatabaseConfig()
		dsn := cfg.DSN()
		assert.Contains(t, dsn, "host=localhost")
		assert.Contains(t, dsn, "port=5432")
	})
}

func TestClickHouseConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := persistence.DefaultClickHouseConfig()
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 9000, cfg.Port)
		assert.Equal(t, "telemetryflow_analytics", cfg.Database)
	})
}

func mustNewResource(uri vo.ResourceURI, name string) *entities.Resource {
	r, _ := entities.NewResource(uri, name)
	return r
}

func mustNewPrompt(name vo.ToolName, desc string) *entities.Prompt {
	p, _ := entities.NewPrompt(name, desc)
	return p
}
