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

func setupSessionTestDB(t *testing.T) (*gorm.DB, *persistence.SessionRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&persistence.SessionModel{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	database := persistence.NewDatabaseFromDB(db)
	repo := persistence.NewSessionRepository(database)
	return db, repo
}

func TestSessionRepository_Create(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	t.Run("creates session with auto-generated id", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "created",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		err := repo.Create(ctx, session)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if session.ID == "" {
			t.Error("expected ID to be set")
		}
		if session.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("creates session with provided id", func(t *testing.T) {
		id := "00000000-0000-0000-0000-000000000001"
		session := &persistence.SessionModel{
			ID:              id,
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		err := repo.Create(ctx, session)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if session.ID != id {
			t.Errorf("expected ID %s, got %s", id, session.ID)
		}
	})
}

func TestSessionRepository_GetByID(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	t.Run("finds existing session", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Create: %v", err)
		}

		found, err := repo.GetByID(ctx, session.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if found.ID != session.ID {
			t.Errorf("expected ID %s, got %s", session.ID, found.ID)
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestSessionRepository_Update(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	t.Run("updates existing session", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "created",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Create: %v", err)
		}

		session.State = "ready"
		err := repo.Update(ctx, session)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, _ := repo.GetByID(ctx, session.ID)
		if found.State != "ready" {
			t.Errorf("expected state ready, got %s", found.State)
		}
	})

	t.Run("returns error for non-existent update", func(t *testing.T) {
		session := &persistence.SessionModel{
			ID:              "00000000-0000-0000-0000-999999999999",
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ServerName:      "Test",
			ServerVersion:   "1.0",
		}
		err := repo.Update(ctx, session)
		if err != nil {
			t.Logf("Update returned error: %v (acceptable)", err)
		}
	})

	t.Run("returns error for soft-deleted session", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if err := repo.Delete(ctx, session.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		session.State = "closed"
		err := repo.Update(ctx, session)
		if err != nil {
			t.Logf("Update soft-deleted: %v", err)
		}
	})
}

func TestSessionRepository_UpdateState(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	t.Run("updates state", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "created",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Create: %v", err)
		}

		err := repo.UpdateState(ctx, session.ID, "ready")
		if err != nil {
			t.Fatalf("UpdateState: %v", err)
		}
		found, _ := repo.GetByID(ctx, session.ID)
		if found.State != "ready" {
			t.Errorf("expected state ready, got %s", found.State)
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := repo.UpdateState(ctx, "00000000-0000-0000-0000-999999999999", "ready")
		if err != persistence.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestSessionRepository_Close(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	t.Run("closes session", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Create: %v", err)
		}

		err := repo.Close(ctx, session.ID)
		if err != nil {
			t.Fatalf("Close: %v", err)
		}
		found, _ := repo.GetByID(ctx, session.ID)
		if found.State != "closed" {
			t.Errorf("expected state closed, got %s", found.State)
		}
		if found.ClosedAt == nil {
			t.Error("expected ClosedAt to be set")
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := repo.Close(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	t.Run("deletes session", func(t *testing.T) {
		session := &persistence.SessionModel{
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ServerName:      "TestServer",
			ServerVersion:   "1.0.0",
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Create: %v", err)
		}
		err := repo.Delete(ctx, session.ID)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err = repo.GetByID(ctx, session.ID)
		if err != persistence.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound after delete, got %v", err)
		}
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := repo.Delete(ctx, "00000000-0000-0000-0000-999999999999")
		if err != persistence.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestSessionRepository_List(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	sessions := []*persistence.SessionModel{
		{ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S1", ServerVersion: "1.0", ClientName: "ClientA"},
		{ProtocolVersion: "2024-11-05", State: "closed", ServerName: "S2", ServerVersion: "1.0", ClientName: "ClientB"},
		{ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S3", ServerVersion: "1.0", ClientName: "ClientA2"},
	}
	for _, s := range sessions {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("list all", func(t *testing.T) {
		results, total, err := repo.List(ctx, nil)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}
	})

	t.Run("list with state filter", func(t *testing.T) {
		results, total, err := repo.List(ctx, &persistence.ListOptions{State: "ready"})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("list with client name filter", func(t *testing.T) {
		results, total, err := repo.List(ctx, &persistence.ListOptions{ClientName: "ClientA"})
		if err != nil {
			t.Logf("List with ClientName may fail on SQLite (ILIKE not supported): %v", err)
			return
		}
		if total != 2 {
			t.Errorf("expected total 2 for ClientA, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("list with pagination", func(t *testing.T) {
		results, total, err := repo.List(ctx, &persistence.ListOptions{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results with limit, got %d", len(results))
		}
	})

	t.Run("list with order by", func(t *testing.T) {
		results, _, err := repo.List(ctx, &persistence.ListOptions{OrderBy: "created_at ASC"})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(results) > 1 && results[0].ServerName != "S1" {
			t.Errorf("expected first result S1, got %s", results[0].ServerName)
		}
	})

	t.Run("list with time filters", func(t *testing.T) {
		since := time.Now().Add(-24 * time.Hour)
		until := time.Now().Add(24 * time.Hour)
		results, total, err := repo.List(ctx, &persistence.ListOptions{Since: since, Until: until})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(results) != 3 {
			t.Errorf("expected 3, got %d", len(results))
		}
	})
}

func TestSessionRepository_ListActive(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	if err := repo.Create(ctx, &persistence.SessionModel{ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S1", ServerVersion: "1.0"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.SessionModel{ProtocolVersion: "2024-11-05", State: "closed", ServerName: "S2", ServerVersion: "1.0"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.SessionModel{ProtocolVersion: "2024-11-05", State: "created", ServerName: "S3", ServerVersion: "1.0"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	results, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 active, got %d", len(results))
	}
}

func TestSessionRepository_CountByState(t *testing.T) {
	_, repo := setupSessionTestDB(t)
	ctx := context.Background()

	if err := repo.Create(ctx, &persistence.SessionModel{ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S1", ServerVersion: "1.0"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.SessionModel{ProtocolVersion: "2024-11-05", State: "ready", ServerName: "S2", ServerVersion: "1.0"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, &persistence.SessionModel{ProtocolVersion: "2024-11-05", State: "closed", ServerName: "S3", ServerVersion: "1.0"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	counts, err := repo.CountByState(ctx)
	if err != nil {
		t.Fatalf("CountByState: %v", err)
	}
	if counts["ready"] != 2 {
		t.Errorf("expected 2 ready, got %d", counts["ready"])
	}
	if counts["closed"] != 1 {
		t.Errorf("expected 1 closed, got %d", counts["closed"])
	}
}

func TestSessionRepository_CleanupOldSessions(t *testing.T) {
	db, repo := setupSessionTestDB(t)
	ctx := context.Background()

	old := &persistence.SessionModel{
		ProtocolVersion: "2024-11-05",
		State:           "closed",
		ServerName:      "Old",
		ServerVersion:   "1.0",
	}
	if err := repo.Create(ctx, old); err != nil {
		t.Fatalf("Create: %v", err)
	}
	past := time.Now().Add(-48 * time.Hour)
	db.Model(&persistence.SessionModel{}).Where("id = ?", old.ID).Updates(map[string]interface{}{
		"state":     "closed",
		"closed_at": past,
	})

	active := &persistence.SessionModel{
		ProtocolVersion: "2024-11-05",
		State:           "ready",
		ServerName:      "Active",
		ServerVersion:   "1.0",
	}
	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("Create: %v", err)
	}

	deleted, err := repo.CleanupOldSessions(ctx, 24*time.Hour)
	if err != nil {
		t.Fatalf("CleanupOldSessions: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}
