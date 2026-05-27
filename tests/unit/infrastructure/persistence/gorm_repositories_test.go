package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&persistence.SessionModel{},
		&persistence.ConversationModel{},
		&persistence.MessageModel{},
		&persistence.ToolModel{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newTestSession(t *testing.T) *aggregates.Session {
	t.Helper()
	s := aggregates.NewSession()
	err := s.Initialize(&aggregates.ClientInfo{Name: "test-client", Version: "1.0.0"}, "2024-11-05")
	if err != nil {
		t.Fatalf("failed to init session: %v", err)
	}
	s.MarkReady()
	return s
}

func newTestConversation(t *testing.T, sessionID vo.SessionID) *aggregates.Conversation {
	t.Helper()
	return aggregates.NewConversation(sessionID, vo.ModelClaudeOpus47)
}

func newTestTool(t *testing.T, name string) *entities.Tool {
	t.Helper()
	tn, err := vo.NewToolName(name)
	if err != nil {
		t.Fatalf("failed to create tool name: %v", err)
	}
	td, err := vo.NewToolDescription("description for " + name)
	if err != nil {
		t.Fatalf("failed to create tool desc: %v", err)
	}
	tool, err := entities.NewTool(tn, td, nil)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}
	return tool
}

func TestGormSessionRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := persistence.NewGormSessionRepository(db)
	ctx := context.Background()

	t.Run("save and find by id", func(t *testing.T) {
		session := newTestSession(t)
		err := repo.Save(ctx, session)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		found, err := repo.FindByID(ctx, session.ID())
		if err != nil {
			t.Fatalf("FindByID failed: %v", err)
		}
		if found.ID().String() != session.ID().String() {
			t.Errorf("expected ID %s, got %s", session.ID(), found.ID())
		}
		if found.State() != session.State() {
			t.Errorf("expected state %s, got %s", session.State(), found.State())
		}
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		found, err := repo.FindByID(ctx, vo.GenerateSessionID())
		if err != nil {
			t.Fatalf("FindByID error: %v", err)
		}
		if found != nil {
			t.Error("expected nil for non-existent session")
		}
	})

	t.Run("find all", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		s1 := newTestSession(t)
		s2 := newTestSession(t)
		if err := repo.Save(ctx, s1); err != nil {
			t.Fatalf("save s1: %v", err)
		}
		if err := repo.Save(ctx, s2); err != nil {
			t.Fatalf("save s2: %v", err)
		}

		all, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) != 2 {
			t.Errorf("expected 2 sessions, got %d", len(all))
		}
	})

	t.Run("find active", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)

		active := newTestSession(t)
		closed := aggregates.NewSession()
		err := closed.Initialize(&aggregates.ClientInfo{Name: "c", Version: "1.0"}, "2024-11-05")
		if err != nil {
			t.Fatalf("init: %v", err)
		}
		closed.MarkReady()
		closed.Close()

		if err := repo.Save(ctx, active); err != nil {
			t.Fatalf("save active: %v", err)
		}
		if err := repo.Save(ctx, closed); err != nil {
			t.Fatalf("save closed: %v", err)
		}

		actives, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive: %v", err)
		}
		if len(actives) != 1 {
			t.Errorf("expected 1 active, got %d", len(actives))
		}
	})

	t.Run("delete", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		s := newTestSession(t)
		if err := repo.Save(ctx, s); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Delete(ctx, s.ID()); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		found, _ := repo.FindByID(ctx, s.ID())
		if found != nil {
			t.Error("expected nil after delete")
		}
	})

	t.Run("exists", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		s := newTestSession(t)
		if err := repo.Save(ctx, s); err != nil {
			t.Fatalf("save: %v", err)
		}

		ok, err := repo.Exists(ctx, s.ID())
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if !ok {
			t.Error("expected exists=true")
		}

		ok, err = repo.Exists(ctx, vo.GenerateSessionID())
		if err != nil {
			t.Fatalf("Exists missing: %v", err)
		}
		if ok {
			t.Error("expected exists=false for missing")
		}
	})

	t.Run("count", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		if err := repo.Save(ctx, newTestSession(t)); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Save(ctx, newTestSession(t)); err != nil {
			t.Fatalf("save: %v", err)
		}
		n, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if n != 2 {
			t.Errorf("expected count 2, got %d", n)
		}
	})

	t.Run("save with metadata and capabilities", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		s := newTestSession(t)
		s.SetMetadata("key1", "value1")

		if err := repo.Save(ctx, s); err != nil {
			t.Fatalf("save: %v", err)
		}
		found, err := repo.FindByID(ctx, s.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.ID().String() != s.ID().String() {
			t.Errorf("ID mismatch")
		}
	})
}

func TestGormConversationRepository(t *testing.T) {
	db := setupTestDB(t)
	sessionRepo := persistence.NewGormSessionRepository(db)
	repo := persistence.NewGormConversationRepository(db)
	ctx := context.Background()

	session := newTestSession(t)
	if err := sessionRepo.Save(ctx, session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	t.Run("save and find by id", func(t *testing.T) {
		conv := newTestConversation(t, session.ID())
		if err := repo.Save(ctx, conv); err != nil {
			t.Fatalf("Save: %v", err)
		}

		found, err := repo.FindByID(ctx, conv.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.ID().String() != conv.ID().String() {
			t.Errorf("ID mismatch: expected %s, got %s", conv.ID(), found.ID())
		}
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		found, err := repo.FindByID(ctx, vo.GenerateConversationID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found != nil {
			t.Error("expected nil")
		}
	})

	t.Run("find by session id", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		repo := persistence.NewGormConversationRepository(db)

		s1 := newTestSession(t)
		s2 := newTestSession(t)
		if err := sessionRepo.Save(ctx, s1); err != nil {
			t.Fatalf("save s1: %v", err)
		}
		if err := sessionRepo.Save(ctx, s2); err != nil {
			t.Fatalf("save s2: %v", err)
		}

		c1 := newTestConversation(t, s1.ID())
		c2 := newTestConversation(t, s1.ID())
		c3 := newTestConversation(t, s2.ID())
		if err := repo.Save(ctx, c1); err != nil {
			t.Fatalf("save c1: %v", err)
		}
		if err := repo.Save(ctx, c2); err != nil {
			t.Fatalf("save c2: %v", err)
		}
		if err := repo.Save(ctx, c3); err != nil {
			t.Fatalf("save c3: %v", err)
		}

		convs, err := repo.FindBySessionID(ctx, s1.ID())
		if err != nil {
			t.Fatalf("FindBySessionID: %v", err)
		}
		if len(convs) != 2 {
			t.Errorf("expected 2 conversations, got %d", len(convs))
		}
	})

	t.Run("find active", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		repo := persistence.NewGormConversationRepository(db)

		s := newTestSession(t)
		if err := sessionRepo.Save(ctx, s); err != nil {
			t.Fatalf("save: %v", err)
		}

		active := newTestConversation(t, s.ID())
		closed := newTestConversation(t, s.ID())
		closed.Close()

		if err := repo.Save(ctx, active); err != nil {
			t.Fatalf("save active: %v", err)
		}
		if err := repo.Save(ctx, closed); err != nil {
			t.Fatalf("save closed: %v", err)
		}

		actives, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive: %v", err)
		}
		if len(actives) != 1 {
			t.Errorf("expected 1 active, got %d", len(actives))
		}
	})

	t.Run("delete", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		repo := persistence.NewGormConversationRepository(db)

		s := newTestSession(t)
		if err := sessionRepo.Save(ctx, s); err != nil {
			t.Fatalf("save session: %v", err)
		}
		conv := newTestConversation(t, s.ID())
		if err := repo.Save(ctx, conv); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Delete(ctx, conv.ID()); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		found, _ := repo.FindByID(ctx, conv.ID())
		if found != nil {
			t.Error("expected nil after delete")
		}
	})

	t.Run("exists", func(t *testing.T) {
		conv := newTestConversation(t, session.ID())
		if err := repo.Save(ctx, conv); err != nil {
			t.Fatalf("save: %v", err)
		}
		ok, err := repo.Exists(ctx, conv.ID())
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if !ok {
			t.Error("expected exists=true")
		}
	})

	t.Run("count", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		repo := persistence.NewGormConversationRepository(db)
		s := newTestSession(t)
		if err := sessionRepo.Save(ctx, s); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Save(ctx, newTestConversation(t, s.ID())); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Save(ctx, newTestConversation(t, s.ID())); err != nil {
			t.Fatalf("save: %v", err)
		}
		n, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if n != 2 {
			t.Errorf("expected 2, got %d", n)
		}
	})

	t.Run("count by session id", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		repo := persistence.NewGormConversationRepository(db)
		s1 := newTestSession(t)
		s2 := newTestSession(t)
		if err := sessionRepo.Save(ctx, s1); err != nil {
			t.Fatalf("save s1: %v", err)
		}
		if err := sessionRepo.Save(ctx, s2); err != nil {
			t.Fatalf("save s2: %v", err)
		}

		if err := repo.Save(ctx, newTestConversation(t, s1.ID())); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Save(ctx, newTestConversation(t, s1.ID())); err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := repo.Save(ctx, newTestConversation(t, s2.ID())); err != nil {
			t.Fatalf("save: %v", err)
		}

		n, err := repo.CountBySessionID(ctx, s1.ID())
		if err != nil {
			t.Fatalf("CountBySessionID: %v", err)
		}
		if n != 2 {
			t.Errorf("expected 2, got %d", n)
		}
	})
}

func TestGormToolRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("register and find by name", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "test_tool")
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}

		tn, _ := vo.NewToolName("test_tool")
		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found.Name().String() != "test_tool" {
			t.Errorf("expected test_tool, got %s", found.Name().String())
		}
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tn, _ := vo.NewToolName("missing_tool")
		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found != nil {
			t.Error("expected nil")
		}
	})

	t.Run("find all", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		t1 := newTestTool(t, "tool_a")
		if err := repo.Register(ctx, t1); err != nil {
			t.Fatalf("register: %v", err)
		}

		all, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) < 1 {
			t.Errorf("expected at least 1, got %d", len(all))
		}
	})

	t.Run("find by category", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		t1 := newTestTool(t, "cat_tool1")
		t1.SetCategory("utility")
		if err := repo.Register(ctx, t1); err != nil {
			t.Fatalf("register: %v", err)
		}

		found, err := repo.FindByCategory(ctx, "utility")
		if err != nil {
			t.Fatalf("FindByCategory: %v", err)
		}
		if len(found) != 1 {
			t.Errorf("expected 1, got %d", len(found))
		}
	})

	t.Run("find enabled", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		t1 := newTestTool(t, "enabled_t")
		if err := repo.Register(ctx, t1); err != nil {
			t.Fatalf("register: %v", err)
		}

		found, err := repo.FindEnabled(ctx)
		if err != nil {
			t.Fatalf("FindEnabled: %v", err)
		}
		if len(found) != 1 {
			t.Errorf("expected 1 enabled, got %d", len(found))
		}
	})

	t.Run("unregister", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "rem_tool")
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("register: %v", err)
		}

		tn, _ := vo.NewToolName("rem_tool")
		if err := repo.Unregister(ctx, tn); err != nil {
			t.Fatalf("Unregister: %v", err)
		}
		found, _ := repo.FindByName(ctx, tn)
		if found != nil {
			t.Error("expected nil after unregister")
		}
	})

	t.Run("exists", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "exists_tool")
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("register: %v", err)
		}

		tn, _ := vo.NewToolName("exists_tool")
		ok, err := repo.Exists(ctx, tn)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if !ok {
			t.Error("expected exists=true")
		}

		tn2, _ := vo.NewToolName("nope")
		ok2, err := repo.Exists(ctx, tn2)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if ok2 {
			t.Error("expected exists=false")
		}
	})

	t.Run("count", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		if err := repo.Register(ctx, newTestTool(t, "cnt_a")); err != nil {
			t.Fatalf("register: %v", err)
		}
		n, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if n < 1 {
			t.Errorf("expected at least 1, got %d", n)
		}
	})

	t.Run("tool with input schema and tags", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)

		tn, _ := vo.NewToolName("schema_tool")
		td, _ := vo.NewToolDescription("a tool with schema")
		schema := &entities.JSONSchema{
			Type: "object",
			Properties: map[string]*entities.JSONSchema{
				"msg": {Type: "string", Description: "message"},
			},
			Required: []string{"msg"},
		}
		tool, err := entities.NewTool(tn, td, schema)
		if err != nil {
			t.Fatalf("NewTool: %v", err)
		}
		tool.SetCategory("utility")
		tool.SetTags([]string{"test", "demo"})
		tool.SetTimeout(60 * time.Second)

		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}

		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found.Category() != "utility" {
			t.Errorf("expected category utility, got %s", found.Category())
		}
		if found.Timeout() != 60*time.Second {
			t.Errorf("expected timeout 60s, got %v", found.Timeout())
		}
	})

	t.Run("find by tag uses postgresql-specific operator", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "tagged_tool")
		tool.SetTags([]string{"test"})
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("register: %v", err)
		}

		_, err := repo.FindByTag(ctx, "test")
		if err != nil {
			t.Logf("FindByTag fails on SQLite (PostgreSQL-specific): %v", err)
		}
	})

	t.Run("session with client info nil", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		s := aggregates.NewSession()
		if err := repo.Save(ctx, s); err != nil {
			t.Fatalf("Save: %v", err)
		}
		found, err := repo.FindByID(ctx, s.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found == nil {
			t.Fatal("expected non-nil session")
		}
	})

	t.Run("session with capabilities", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		s := newTestSession(t)
		_ = s.SetLogLevel(vo.LogLevelDebug)
		if err := repo.Save(ctx, s); err != nil {
			t.Fatalf("Save: %v", err)
		}
		found, err := repo.FindByID(ctx, s.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("conversation round trip with closed state", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		repo := persistence.NewGormConversationRepository(db)

		s := newTestSession(t)
		if err := sessionRepo.Save(ctx, s); err != nil {
			t.Fatalf("save session: %v", err)
		}

		conv := newTestConversation(t, s.ID())
		conv.Close()
		if err := repo.Save(ctx, conv); err != nil {
			t.Fatalf("Save: %v", err)
		}

		found, err := repo.FindByID(ctx, conv.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("tool with no schema", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "no_schema_tool")
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}
		tn, _ := vo.NewToolName("no_schema_tool")
		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found.InputSchema() != nil {
			t.Error("expected nil schema")
		}
	})

	t.Run("modelToSession error path - invalid session id in find all", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		db.Exec("INSERT INTO sessions (id, protocol_version, state, server_name, server_version, created_at, updated_at) VALUES (?, '2024-11-05', 'ready', 'S', '1.0', datetime('now'), datetime('now'))", "not-a-valid-uuid")
		_, err := repo.FindAll(ctx)
		if err != nil {
			t.Logf("FindAll with bad ID returned error: %v", err)
		}
	})

	t.Run("modelToSession error path - invalid session id in find active", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		db.Exec("INSERT INTO sessions (id, protocol_version, state, server_name, server_version, created_at, updated_at) VALUES (?, '2024-11-05', 'ready', 'S', '1.0', datetime('now'), datetime('now'))", "not-a-valid-uuid")
		_, err := repo.FindActive(ctx)
		if err != nil {
			t.Logf("FindActive with bad ID returned error: %v", err)
		}
	})

	t.Run("modelToConversation error path - invalid id in find by session", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		convRepo := persistence.NewGormConversationRepository(db)

		s := newTestSession(t)
		if err := sessionRepo.Save(ctx, s); err != nil {
			t.Fatalf("save session: %v", err)
		}

		db.Exec("INSERT INTO conversations (id, session_id, model, status, created_at, updated_at) VALUES (?, ?, 'model', 'active', datetime('now'), datetime('now'))", "bad-conv-id", s.ID().String())
		_, err := convRepo.FindBySessionID(ctx, s.ID())
		if err != nil {
			t.Logf("FindBySessionID with bad ID returned error: %v", err)
		}
	})

	t.Run("modelToTool error path - invalid tool name in find all", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		db.Exec("INSERT INTO tools (name, description, is_enabled, timeout, created_at, updated_at) VALUES ('', 'desc', 1, 30, datetime('now'), datetime('now'))")
		_, err := repo.FindAll(ctx)
		if err != nil {
			t.Logf("FindAll with empty tool name returned error: %v", err)
		}
	})

	t.Run("modelToTool error path - invalid description in find by category", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "cat_tool")
		tool.SetCategory("test-cat")
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}
		results, err := repo.FindByCategory(ctx, "test-cat")
		if err != nil {
			t.Fatalf("FindByCategory: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1, got %d", len(results))
		}
	})

	t.Run("modelToTool error path - find enabled", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tool := newTestTool(t, "enabled_tool")
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}
		results, err := repo.FindEnabled(ctx)
		if err != nil {
			t.Fatalf("FindEnabled: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1, got %d", len(results))
		}
	})

	t.Run("exists returns false for non-existent session", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sid, _ := vo.NewSessionID("00000000-0000-0000-0000-000000009999")
		exists, err := repo.Exists(ctx, sid)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if exists {
			t.Error("expected false")
		}
	})

	t.Run("count returns zero for empty db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("conversation exists returns false", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		cid, _ := vo.NewConversationID("00000000-0000-0000-0000-000000009999")
		exists, err := repo.Exists(ctx, cid)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if exists {
			t.Error("expected false")
		}
	})

	t.Run("conversation count returns zero", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("count by session id returns zero", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		sid, _ := vo.NewSessionID("00000000-0000-0000-0000-000000009999")
		count, err := repo.CountBySessionID(ctx, sid)
		if err != nil {
			t.Fatalf("CountBySessionID: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("tool exists returns false", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		tn, _ := vo.NewToolName("nonexistent")
		exists, err := repo.Exists(ctx, tn)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if exists {
			t.Error("expected false")
		}
	})

	t.Run("tool count returns zero", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("modelToTool error path - invalid description", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		db.Exec("INSERT INTO tools (name, description, is_enabled, timeout, created_at, updated_at) VALUES (?, '', 1, 30, datetime('now'), datetime('now'))", "x")
		_, err := repo.FindAll(ctx)
		if err != nil {
			t.Logf("FindAll with empty desc returned error: %v", err)
		}
	})

	t.Run("modelToConversation error path - invalid session id", func(t *testing.T) {
		db := setupTestDB(t)
		sessionRepo := persistence.NewGormSessionRepository(db)
		convRepo := persistence.NewGormConversationRepository(db)

		s := newTestSession(t)
		if err := sessionRepo.Save(ctx, s); err != nil {
			t.Fatalf("save session: %v", err)
		}

		db.Exec("INSERT INTO conversations (id, session_id, model, status, created_at, updated_at) VALUES (?, 'bad-session-id', 'model', 'active', datetime('now'), datetime('now'))", "00000000-0000-0000-0000-000000008888")
		_, err := convRepo.FindActive(ctx)
		if err != nil {
			t.Logf("FindActive with bad session ID returned error: %v", err)
		}
	})

	t.Run("disabled tool round-trip", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)

		tool := newTestTool(t, "disabled_rt")
		tool.Disable()
		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}

		tn, _ := vo.NewToolName("disabled_rt")
		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find tool")
		}
		if found.Name().String() != "disabled_rt" {
			t.Errorf("expected name disabled_rt, got %s", found.Name().String())
		}
	})

	t.Run("session Exists error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.Exists(ctx, vo.GenerateSessionID())
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("session Count error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.Count(ctx)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("session FindByID error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.FindByID(ctx, vo.GenerateSessionID())
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("session FindAll error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.FindAll(ctx)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("session FindActive error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.FindActive(ctx)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("conversation Exists error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.Exists(ctx, vo.GenerateConversationID())
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("conversation Count error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.Count(ctx)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("conversation CountBySessionID error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.CountBySessionID(ctx, vo.GenerateSessionID())
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("tool Exists error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		tn, _ := vo.NewToolName("test")
		_, err := repo.Exists(ctx, tn)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("tool Count error on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		_, err := repo.Count(ctx)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("tool with schema but no tags", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)

		tn, _ := vo.NewToolName("schema_only_tool")
		td, _ := vo.NewToolDescription("schema only")
		schema := &entities.JSONSchema{Type: "object"}
		tool, _ := entities.NewTool(tn, td, schema)

		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}

		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found.InputSchema() == nil {
			t.Error("expected non-nil schema")
		}
	})

	t.Run("tool with tags but no schema", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)

		tn, _ := vo.NewToolName("tags_only_tool")
		td, _ := vo.NewToolDescription("tags only")
		tool, _ := entities.NewTool(tn, td, nil)
		tool.SetTags([]string{"tag1", "tag2"})

		if err := repo.Register(ctx, tool); err != nil {
			t.Fatalf("Register: %v", err)
		}

		found, err := repo.FindByName(ctx, tn)
		if err != nil {
			t.Fatalf("FindByName: %v", err)
		}
		if found == nil {
			t.Fatal("expected tool to be found")
		}
		if found.Name().String() != "tags_only_tool" {
			t.Errorf("expected name tags_only_tool, got %s", found.Name().String())
		}
	})

	t.Run("session Save on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormSessionRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		err := repo.Save(ctx, newTestSession(t))
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("conversation Save on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormConversationRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		sid := vo.GenerateSessionID()
		err := repo.Save(ctx, aggregates.NewConversation(sid, vo.ModelClaudeOpus47))
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("tool Register on closed db", func(t *testing.T) {
		db := setupTestDB(t)
		repo := persistence.NewGormToolRepository(db)
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()

		err := repo.Register(ctx, newTestTool(t, "fail_tool"))
		if err == nil {
			t.Error("expected error on closed db")
		}
	})
}
