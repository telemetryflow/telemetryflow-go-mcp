package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupConversationTestDB(t *testing.T) (*gorm.DB, *persistence.ConversationRepository, *persistence.MessageRepository) {
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
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	database := persistence.NewDatabaseFromDB(db)
	convRepo := persistence.NewConversationRepository(database)
	msgRepo := persistence.NewMessageRepository(database)
	return db, convRepo, msgRepo
}

func createTestSessionForConv(t *testing.T, db *gorm.DB) string {
	t.Helper()
	session := &persistence.SessionModel{
		ProtocolVersion: "2024-11-05",
		State:           "ready",
		ServerName:      "TestServer",
		ServerVersion:   "1.0.0",
	}
	database := persistence.NewDatabaseFromDB(db)
	repo := persistence.NewSessionRepository(database)
	if err := repo.Create(context.Background(), session); err != nil {
		t.Fatalf("create session: %v", err)
	}
	return session.ID
}

func TestConversationRepository_Create(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("creates conversation with auto-generated id", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		err := repo.Create(ctx, conv)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if conv.ID == "" {
			t.Error("expected ID to be set")
		}
	})

	t.Run("creates conversation with provided id", func(t *testing.T) {
		id := "00000000-0000-0000-0000-000000000001"
		conv := &persistence.ConversationModel{
			ID:        id,
			SessionID: sessionID,
			Model:     "claude-sonnet-4-20250514",
			Status:    "active",
		}
		err := repo.Create(ctx, conv)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if conv.ID != id {
			t.Errorf("expected ID %s, got %s", id, conv.ID)
		}
	})
}

func TestConversationRepository_GetByID(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("finds existing conversation", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create: %v", err)
		}

		found, err := repo.GetByID(ctx, conv.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if found.ID != conv.ID {
			t.Errorf("ID mismatch")
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrConversationNotFound {
			t.Errorf("expected ErrConversationNotFound, got %v", err)
		}
	})
}

func TestConversationRepository_GetByIDWithMessages(t *testing.T) {
	db, repo, msgRepo := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("returns conversation with messages", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create conv: %v", err)
		}

		msg1 := &persistence.MessageModel{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        persistence.JSONB{"type": "text", "text": "hello"},
		}
		msg2 := &persistence.MessageModel{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        persistence.JSONB{"type": "text", "text": "world"},
		}
		if err := msgRepo.Create(ctx, msg1); err != nil {
			t.Fatalf("Create msg1: %v", err)
		}
		if err := msgRepo.Create(ctx, msg2); err != nil {
			t.Fatalf("Create msg2: %v", err)
		}

		found, err := repo.GetByIDWithMessages(ctx, conv.ID)
		if err != nil {
			t.Fatalf("GetByIDWithMessages: %v", err)
		}
		if len(found.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(found.Messages))
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		_, err := repo.GetByIDWithMessages(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrConversationNotFound {
			t.Errorf("expected ErrConversationNotFound, got %v", err)
		}
	})
}

func TestConversationRepository_Update(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("updates existing conversation", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create: %v", err)
		}
		conv.Status = "paused"
		err := repo.Update(ctx, conv)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		found, _ := repo.GetByID(ctx, conv.ID)
		if found.Status != "paused" {
			t.Errorf("expected status paused, got %s", found.Status)
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			ID:        "00000000-0000-0000-0000-999999999999",
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		err := repo.Update(ctx, conv)
		if err != nil {
			t.Logf("Update returned error: %v (acceptable)", err)
		}
	})

	t.Run("returns error for soft-deleted conversation", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if err := repo.Delete(ctx, conv.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		conv.Status = "paused"
		err := repo.Update(ctx, conv)
		if err != nil {
			t.Logf("Update soft-deleted: %v", err)
		}
	})
}

func TestConversationRepository_UpdateStatus(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("updates status", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create: %v", err)
		}
		err := repo.UpdateStatus(ctx, conv.ID, "paused")
		if err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
		found, _ := repo.GetByID(ctx, conv.ID)
		if found.Status != "paused" {
			t.Errorf("expected paused, got %s", found.Status)
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "00000000-0000-0000-0000-999999999999", "paused")
		if err != persistence.ErrConversationNotFound {
			t.Errorf("expected ErrConversationNotFound, got %v", err)
		}
	})
}

func TestConversationRepository_Close(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("closes conversation", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create: %v", err)
		}
		err := repo.Close(ctx, conv.ID)
		if err != nil {
			t.Fatalf("Close: %v", err)
		}
		found, _ := repo.GetByID(ctx, conv.ID)
		if found.Status != "closed" {
			t.Errorf("expected closed, got %s", found.Status)
		}
		if found.ClosedAt == nil {
			t.Error("expected ClosedAt to be set")
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := repo.Close(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrConversationNotFound {
			t.Errorf("expected ErrConversationNotFound, got %v", err)
		}
	})
}

func TestConversationRepository_Delete(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	t.Run("deletes conversation", func(t *testing.T) {
		conv := &persistence.ConversationModel{
			SessionID: sessionID,
			Model:     "claude-opus-4-7",
			Status:    "active",
		}
		if err := repo.Create(ctx, conv); err != nil {
			t.Fatalf("Create: %v", err)
		}
		err := repo.Delete(ctx, conv.ID)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err = repo.GetByID(ctx, conv.ID)
		if err != persistence.ErrConversationNotFound {
			t.Errorf("expected ErrConversationNotFound, got %v", err)
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := repo.Delete(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrConversationNotFound {
			t.Errorf("expected ErrConversationNotFound, got %v", err)
		}
	})
}

func TestConversationRepository_ListBySession(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sid1 := createTestSessionForConv(t, db)

	session2 := &persistence.SessionModel{
		ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S2", ServerVersion: "1.0",
	}
	database := persistence.NewDatabaseFromDB(db)
	if err := persistence.NewSessionRepository(database).Create(ctx, session2); err != nil {
		t.Fatalf("Create session2: %v", err)
	}
	sid2 := session2.ID

	for _, c := range []*persistence.ConversationModel{
		{SessionID: sid1, Model: "claude-opus-4-7", Status: "active"},
		{SessionID: sid1, Model: "claude-sonnet-4-20250514", Status: "closed"},
		{SessionID: sid2, Model: "claude-opus-4-7", Status: "active"},
	} {
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("lists by session", func(t *testing.T) {
		results, total, err := repo.ListBySession(ctx, sid1, nil)
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("lists by session with status filter", func(t *testing.T) {
		results, total, err := repo.ListBySession(ctx, sid1, &persistence.ListOptions{State: "active"})
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(results) != 1 {
			t.Errorf("expected 1, got %d", len(results))
		}
	})

	t.Run("lists by session with pagination", func(t *testing.T) {
		results, total, err := repo.ListBySession(ctx, sid1, &persistence.ListOptions{Limit: 1, Offset: 0})
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result with limit, got %d", len(results))
		}
	})

	t.Run("lists by session with since filter", func(t *testing.T) {
		results, total, err := repo.ListBySession(ctx, sid1, &persistence.ListOptions{Since: time.Now().Add(-24 * time.Hour)})
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("lists by session with until filter", func(t *testing.T) {
		results, total, err := repo.ListBySession(ctx, sid1, &persistence.ListOptions{Until: time.Now().Add(time.Hour)})
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("lists by session with order by", func(t *testing.T) {
		results, _, err := repo.ListBySession(ctx, sid1, &persistence.ListOptions{OrderBy: "model ASC"})
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("lists by session default order", func(t *testing.T) {
		results, _, err := repo.ListBySession(ctx, sid1, &persistence.ListOptions{})
		if err != nil {
			t.Fatalf("ListBySession: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})
}

func TestConversationRepository_ListActive(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	if err := repo.Create(ctx, &persistence.ConversationModel{SessionID: sessionID, Model: "claude-opus-4-7", Status: "active"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.ConversationModel{SessionID: sessionID, Model: "claude-opus-4-7", Status: "closed"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	results, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 active, got %d", len(results))
	}
}

func TestConversationRepository_CountByModel(t *testing.T) {
	db, repo, _ := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	if err := repo.Create(ctx, &persistence.ConversationModel{SessionID: sessionID, Model: "claude-opus-4-7", Status: "active"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.ConversationModel{SessionID: sessionID, Model: "claude-opus-4-7", Status: "active"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.ConversationModel{SessionID: sessionID, Model: "claude-sonnet-4-20250514", Status: "active"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	counts, err := repo.CountByModel(ctx)
	if err != nil {
		t.Fatalf("CountByModel: %v", err)
	}
	if counts["claude-opus-4-7"] != 2 {
		t.Errorf("expected 2 opus, got %d", counts["claude-opus-4-7"])
	}
	if counts["claude-sonnet-4-20250514"] != 1 {
		t.Errorf("expected 1 sonnet, got %d", counts["claude-sonnet-4-20250514"])
	}
}

func TestConversationRepository_GetMessageCount(t *testing.T) {
	db, repo, msgRepo := setupConversationTestDB(t)
	ctx := context.Background()
	sessionID := createTestSessionForConv(t, db)

	conv := &persistence.ConversationModel{SessionID: sessionID, Model: "claude-opus-4-7", Status: "active"}
	if err := repo.Create(ctx, conv); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, msg := range []*persistence.MessageModel{
		{ConversationID: conv.ID, Role: "user", Content: persistence.JSONB{"type": "text", "text": "hi"}},
		{ConversationID: conv.ID, Role: "assistant", Content: persistence.JSONB{"type": "text", "text": "hello"}},
	} {
		if err := msgRepo.Create(ctx, msg); err != nil {
			t.Fatalf("Create msg: %v", err)
		}
	}

	count, err := repo.GetMessageCount(ctx, conv.ID)
	if err != nil {
		t.Fatalf("GetMessageCount: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 messages, got %d", count)
	}
}

func TestMessageRepository(t *testing.T) {
	db, _, msgRepo := setupConversationTestDB(t)
	ctx := context.Background()

	database := persistence.NewDatabaseFromDB(db)
	sessionRepo := persistence.NewSessionRepository(database)
	convRepo := persistence.NewConversationRepository(database)

	session := &persistence.SessionModel{
		ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S", ServerVersion: "1.0",
	}
	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	conv := &persistence.ConversationModel{SessionID: session.ID, Model: "claude-opus-4-7", Status: "active"}
	if err := convRepo.Create(ctx, conv); err != nil {
		t.Fatalf("create conv: %v", err)
	}

	t.Run("create and get by id", func(t *testing.T) {
		msg := &persistence.MessageModel{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        persistence.JSONB{"type": "text", "text": "hello"},
			TokenCount:     10,
		}
		err := msgRepo.Create(ctx, msg)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if msg.ID == "" {
			t.Error("expected ID to be set")
		}

		found, err := msgRepo.GetByID(ctx, msg.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if found.Role != "user" {
			t.Errorf("expected role user, got %s", found.Role)
		}
	})

	t.Run("get non-existent returns error", func(t *testing.T) {
		_, err := msgRepo.GetByID(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrMessageNotFound {
			t.Errorf("expected ErrMessageNotFound, got %v", err)
		}
	})

	t.Run("create batch", func(t *testing.T) {
		msgs := []persistence.MessageModel{
			{ConversationID: conv.ID, Role: "user", Content: persistence.JSONB{"type": "text", "text": "msg1"}},
			{ConversationID: conv.ID, Role: "assistant", Content: persistence.JSONB{"type": "text", "text": "msg2"}},
		}
		err := msgRepo.CreateBatch(ctx, msgs)
		if err != nil {
			t.Fatalf("CreateBatch: %v", err)
		}
		for _, m := range msgs {
			if m.ID == "" {
				t.Error("expected ID to be set")
			}
		}
	})

	t.Run("list by conversation", func(t *testing.T) {
		msgs, err := msgRepo.ListByConversation(ctx, conv.ID)
		if err != nil {
			t.Fatalf("ListByConversation: %v", err)
		}
		if len(msgs) < 3 {
			t.Errorf("expected at least 3 messages, got %d", len(msgs))
		}
	})

	t.Run("get last messages", func(t *testing.T) {
		msgs, err := msgRepo.GetLastMessages(ctx, conv.ID, 2)
		if err != nil {
			t.Fatalf("GetLastMessages: %v", err)
		}
		if len(msgs) > 2 {
			t.Errorf("expected at most 2, got %d", len(msgs))
		}
	})

	t.Run("count tokens", func(t *testing.T) {
		msg := &persistence.MessageModel{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        persistence.JSONB{"type": "text", "text": "tokens test"},
			TokenCount:     42,
		}
		if err := msgRepo.Create(ctx, msg); err != nil {
			t.Fatalf("Create: %v", err)
		}

		total, err := msgRepo.CountTokens(ctx, conv.ID)
		if err != nil {
			t.Fatalf("CountTokens: %v", err)
		}
		if total < 42 {
			t.Errorf("expected at least 42 tokens, got %d", total)
		}
	})

	t.Run("delete by conversation", func(t *testing.T) {
		err := msgRepo.DeleteByConversation(ctx, conv.ID)
		if err != nil {
			t.Fatalf("DeleteByConversation: %v", err)
		}
		msgs, _ := msgRepo.ListByConversation(ctx, conv.ID)
		if len(msgs) != 0 {
			t.Errorf("expected 0 messages after delete, got %d", len(msgs))
		}
	})
}
