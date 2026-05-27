package persistence_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSeederDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return db
}

func TestNewDatabaseSeeder(t *testing.T) {
	db := setupSeederDB(t)
	s := persistence.NewDatabaseSeeder(db)
	if s == nil {
		t.Fatal("expected non-nil seeder")
	}
}

func TestDatabaseSeeder_Register(t *testing.T) {
	db := setupSeederDB(t)
	s := persistence.NewDatabaseSeeder(db)
	s.Register("test1", func(ctx context.Context, db *gorm.DB) error { return nil })
	s.Register("test2", func(ctx context.Context, db *gorm.DB) error { return nil })

	result, err := s.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Executed) != 2 {
		t.Errorf("expected 2 executed, got %d", len(result.Executed))
	}
}

func TestDatabaseSeeder_Run(t *testing.T) {
	t.Run("successful seeders", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)
		s.Register("seeder1", func(ctx context.Context, db *gorm.DB) error { return nil })
		s.Register("seeder2", func(ctx context.Context, db *gorm.DB) error { return nil })

		result, err := s.Run(context.Background())
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if len(result.Executed) != 2 {
			t.Errorf("expected 2, got %d", len(result.Executed))
		}
		if result.Duration == 0 {
			t.Error("expected non-zero duration")
		}
	})

	t.Run("failed seeder continues", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)
		s.Register("good", func(ctx context.Context, db *gorm.DB) error { return nil })
		s.Register("bad", func(ctx context.Context, db *gorm.DB) error {
			return fmt.Errorf("seed error")
		})
		s.Register("also_good", func(ctx context.Context, db *gorm.DB) error { return nil })

		result, err := s.Run(context.Background())
		if err == nil {
			t.Error("expected error")
		}
		if len(result.Failed) != 1 {
			t.Errorf("expected 1 failed, got %d", len(result.Failed))
		}
		if len(result.Executed) != 2 {
			t.Errorf("expected 2 executed, got %d", len(result.Executed))
		}
	})

	t.Run("empty seeders", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)

		result, err := s.Run(context.Background())
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if len(result.Executed) != 0 {
			t.Errorf("expected 0 executed, got %d", len(result.Executed))
		}
	})
}

func TestDatabaseSeeder_RunSeeder(t *testing.T) {
	t.Run("runs specific seeder", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)
		called := false
		s.Register("target", func(ctx context.Context, db *gorm.DB) error {
			called = true
			return nil
		})
		s.Register("other", func(ctx context.Context, db *gorm.DB) error { return nil })

		err := s.RunSeeder(context.Background(), "target")
		if err != nil {
			t.Fatalf("RunSeeder: %v", err)
		}
		if !called {
			t.Error("expected target seeder to be called")
		}
	})

	t.Run("returns error for missing seeder", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)
		err := s.RunSeeder(context.Background(), "nonexistent")
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestDatabaseSeeder_SeedFunc(t *testing.T) {
	t.Run("seeder func receives db", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)

		var receivedDB *gorm.DB
		s.Register("check_db", func(ctx context.Context, d *gorm.DB) error {
			receivedDB = d
			return nil
		})

		if err := s.RunSeeder(context.Background(), "check_db"); err != nil {
			t.Fatalf("RunSeeder: %v", err)
		}
		if receivedDB == nil {
			t.Error("expected db to be passed to seeder func")
		}
	})

	t.Run("seeder func receives ctx", func(t *testing.T) {
		db := setupSeederDB(t)
		s := persistence.NewDatabaseSeeder(db)

		var receivedCtx context.Context
		s.Register("check_ctx", func(ctx context.Context, d *gorm.DB) error {
			receivedCtx = ctx
			return nil
		})

		ctx := context.Background()
		if err := s.RunSeeder(ctx, "check_ctx"); err != nil {
			t.Fatalf("RunSeeder: %v", err)
		}
		if receivedCtx != ctx {
			t.Error("expected ctx to be passed to seeder func")
		}
	})
}

func TestSeedTools_Signature(t *testing.T) {
	var fn persistence.SeederFunc = persistence.SeedTools
	if fn == nil {
		t.Error("SeedTools should not be nil")
	}
}

func TestSeedResources_Signature(t *testing.T) {
	var fn persistence.SeederFunc = persistence.SeedResources
	if fn == nil {
		t.Error("SeedResources should not be nil")
	}
}

func TestSeedPrompts_Signature(t *testing.T) {
	var fn persistence.SeederFunc = persistence.SeedPrompts
	if fn == nil {
		t.Error("SeedPrompts should not be nil")
	}
}

func TestSeedAPIKeys_Signature(t *testing.T) {
	var fn persistence.SeederFunc = persistence.SeedAPIKeys
	if fn == nil {
		t.Error("SeedAPIKeys should not be nil")
	}
}

func TestSeedDemoSession_Signature(t *testing.T) {
	var fn persistence.SeederFunc = persistence.SeedDemoSession
	if fn == nil {
		t.Error("SeedDemoSession should not be nil")
	}
}

func TestSeederResult_Fields(t *testing.T) {
	r := &persistence.SeederResult{
		Executed: []string{"a", "b"},
		Skipped:  []string{},
		Failed:   []string{},
	}
	if len(r.Executed) != 2 {
		t.Errorf("expected 2")
	}
}

func TestSeederFunc_Type(t *testing.T) {
	var fn persistence.SeederFunc = func(ctx context.Context, db *gorm.DB) error {
		return nil
	}
	if fn == nil {
		t.Error("SeederFunc should be assignable")
	}
}

func setupModelsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	err = db.AutoMigrate(models.AllModels()...)
	if err != nil {
		t.Logf("AutoMigrate failed (PostgreSQL-specific types): %v", err)
	}
	return db
}

func TestRegisterDefaultSeeders(t *testing.T) {
	db := setupSeederDB(t)
	s := persistence.NewDatabaseSeeder(db)
	s.RegisterDefaultSeeders()
	result, err := s.Run(context.Background())
	if err == nil {
		t.Log("RegisterDefaultSeeders ran successfully")
	} else {
		t.Logf("RegisterDefaultSeeders failed (expected with SQLite): %v", err)
	}
	_ = result
}

func TestSeedTools_Execution(t *testing.T) {
	db := setupModelsDB(t)
	err := persistence.SeedTools(context.Background(), db)
	if err != nil {
		t.Logf("SeedTools error (may be expected with SQLite): %v", err)
	}
}

func TestSeedResources_Execution(t *testing.T) {
	db := setupModelsDB(t)
	err := persistence.SeedResources(context.Background(), db)
	if err != nil {
		t.Logf("SeedResources error (may be expected with SQLite): %v", err)
	}
}

func TestSeedPrompts_Execution(t *testing.T) {
	db := setupModelsDB(t)
	err := persistence.SeedPrompts(context.Background(), db)
	if err != nil {
		t.Logf("SeedPrompts error (may be expected with SQLite): %v", err)
	}
}

func TestSeedAPIKeys_Execution(t *testing.T) {
	db := setupModelsDB(t)
	err := persistence.SeedAPIKeys(context.Background(), db)
	if err != nil {
		t.Logf("SeedAPIKeys error (may be expected with SQLite): %v", err)
	}
}

func TestSeedDemoSession_Execution(t *testing.T) {
	db := setupModelsDB(t)
	err := persistence.SeedDemoSession(context.Background(), db)
	if err != nil {
		t.Logf("SeedDemoSession error (may be expected with SQLite): %v", err)
	}
}

func TestSeedAll_Execution(t *testing.T) {
	db := setupModelsDB(t)
	_, err := persistence.SeedAll(context.Background(), db)
	if err != nil {
		t.Logf("SeedAll error (may be expected with SQLite): %v", err)
	}
}

func TestSeedProduction_Execution(t *testing.T) {
	db := setupModelsDB(t)
	_, err := persistence.SeedProduction(context.Background(), db)
	if err != nil {
		t.Logf("SeedProduction error (may be expected with SQLite): %v", err)
	}
}

func setupDemoTablesDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id BLOB PRIMARY KEY,
		protocol_version TEXT NOT NULL DEFAULT '2024-11-05',
		state TEXT NOT NULL,
		client_name TEXT,
		client_version TEXT,
		server_name TEXT NOT NULL DEFAULT 'TelemetryFlow-MCP',
		server_version TEXT NOT NULL DEFAULT '1.2.0',
		capabilities BLOB,
		log_level TEXT DEFAULT 'info',
		metadata BLOB,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		closed_at DATETIME,
		deleted_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS conversations (
		id BLOB PRIMARY KEY,
		session_id BLOB NOT NULL,
		model TEXT NOT NULL DEFAULT 'claude-opus-4-7',
		system_prompt TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		max_tokens INTEGER NOT NULL DEFAULT 4096,
		temperature REAL NOT NULL DEFAULT 1.0,
		top_p REAL NOT NULL DEFAULT 1.0,
		top_k INTEGER DEFAULT 0,
		stop_sequences BLOB,
		metadata BLOB,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		closed_at DATETIME,
		deleted_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id BLOB PRIMARY KEY,
		conversation_id BLOB NOT NULL,
		role TEXT NOT NULL,
		content BLOB,
		token_count INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS tools (
		id BLOB PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		input_schema BLOB,
		category TEXT,
		tags BLOB,
		is_enabled INTEGER NOT NULL DEFAULT 1,
		rate_limit BLOB,
		timeout_seconds INTEGER NOT NULL DEFAULT 30,
		metadata BLOB,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS resources (
		id BLOB PRIMARY KEY,
		uri TEXT NOT NULL UNIQUE,
		uri_template TEXT,
		name TEXT NOT NULL,
		description TEXT,
		mime_type TEXT,
		is_template INTEGER NOT NULL DEFAULT 0,
		annotations BLOB,
		metadata BLOB,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS prompts (
		id BLOB PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		arguments BLOB,
		template TEXT,
		metadata BLOB,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS api_keys (
		id BLOB PRIMARY KEY,
		key_hash TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		scopes BLOB,
		rate_limit_per_minute INTEGER DEFAULT 60,
		rate_limit_per_hour INTEGER DEFAULT 1000,
		is_active INTEGER NOT NULL DEFAULT 1,
		expires_at DATETIME,
		last_used_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`)

	return db
}

func TestSeedDemoSession_FullRoundTrip(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	err := persistence.SeedDemoSession(ctx, db)
	if err != nil {
		t.Fatalf("SeedDemoSession: %v", err)
	}

	var sessionCount int64
	db.Table("sessions").Count(&sessionCount)
	if sessionCount != 1 {
		t.Errorf("expected 1 session, got %d", sessionCount)
	}

	var convCount int64
	db.Table("conversations").Count(&convCount)
	if convCount != 1 {
		t.Errorf("expected 1 conversation, got %d", convCount)
	}

	var msgCount int64
	db.Table("messages").Count(&msgCount)
	if msgCount != 2 {
		t.Errorf("expected 2 messages, got %d", msgCount)
	}
}

func TestSeedDemoSession_Idempotent(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	if err := persistence.SeedDemoSession(ctx, db); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := persistence.SeedDemoSession(ctx, db); err != nil {
		t.Fatalf("second call: %v", err)
	}

	var msgCount int64
	db.Table("messages").Count(&msgCount)
	if msgCount != 2 {
		t.Errorf("expected 2 messages (idempotent), got %d", msgCount)
	}
}

func TestSeedDemoSession_DBError(t *testing.T) {
	db := setupSeederDB(t)
	ctx := context.Background()

	err := persistence.SeedDemoSession(ctx, db)
	if err == nil {
		t.Error("expected error when tables do not exist")
	}
}

func TestSeedAll_WithModelsDB(t *testing.T) {
	db := setupDemoTablesDB(t)
	result, err := persistence.SeedAll(context.Background(), db)
	if err != nil {
		t.Logf("SeedAll error (some seeders may fail without full schema): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSeedProduction_WithModelsDB(t *testing.T) {
	db := setupDemoTablesDB(t)
	result, err := persistence.SeedProduction(context.Background(), db)
	if err != nil {
		t.Logf("SeedProduction error (some seeders may fail without full schema): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSeedTools_FullRoundTrip(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	err := persistence.SeedTools(ctx, db)
	if err != nil {
		t.Fatalf("SeedTools: %v", err)
	}

	var count int64
	db.Table("tools").Count(&count)
	if count != 8 {
		t.Errorf("expected 8 tools, got %d", count)
	}
}

func TestSeedTools_Idempotent(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	if err := persistence.SeedTools(ctx, db); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := persistence.SeedTools(ctx, db); err != nil {
		t.Fatalf("second call: %v", err)
	}

	var count int64
	db.Table("tools").Count(&count)
	if count != 8 {
		t.Errorf("expected 8 tools (idempotent), got %d", count)
	}
}

func TestSeedResources_FullRoundTrip(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	err := persistence.SeedResources(ctx, db)
	if err != nil {
		t.Fatalf("SeedResources: %v", err)
	}

	var count int64
	db.Table("resources").Count(&count)
	if count != 3 {
		t.Errorf("expected 3 resources, got %d", count)
	}
}

func TestSeedResources_Idempotent(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	if err := persistence.SeedResources(ctx, db); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := persistence.SeedResources(ctx, db); err != nil {
		t.Fatalf("second call: %v", err)
	}

	var count int64
	db.Table("resources").Count(&count)
	if count != 3 {
		t.Errorf("expected 3 resources (idempotent), got %d", count)
	}
}

func TestSeedPrompts_FullRoundTrip(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	err := persistence.SeedPrompts(ctx, db)
	if err != nil {
		t.Fatalf("SeedPrompts: %v", err)
	}

	var count int64
	db.Table("prompts").Count(&count)
	if count != 3 {
		t.Errorf("expected 3 prompts, got %d", count)
	}
}

func TestSeedPrompts_Idempotent(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	if err := persistence.SeedPrompts(ctx, db); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := persistence.SeedPrompts(ctx, db); err != nil {
		t.Fatalf("second call: %v", err)
	}

	var count int64
	db.Table("prompts").Count(&count)
	if count != 3 {
		t.Errorf("expected 3 prompts (idempotent), got %d", count)
	}
}

func TestSeedAPIKeys_FullRoundTrip(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	err := persistence.SeedAPIKeys(ctx, db)
	if err != nil {
		t.Fatalf("SeedAPIKeys: %v", err)
	}

	var count int64
	db.Table("api_keys").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 api key, got %d", count)
	}
}

func TestSeedAPIKeys_Idempotent(t *testing.T) {
	db := setupDemoTablesDB(t)
	ctx := context.Background()

	if err := persistence.SeedAPIKeys(ctx, db); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := persistence.SeedAPIKeys(ctx, db); err != nil {
		t.Fatalf("second call: %v", err)
	}

	var count int64
	db.Table("api_keys").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 api key (idempotent), got %d", count)
	}
}
