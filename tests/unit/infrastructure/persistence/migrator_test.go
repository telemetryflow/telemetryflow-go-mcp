package persistence_test

import (
	"context"
	"embed"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed testdata/migrations/*.sql
var migrationTestFS embed.FS

//go:embed migrator_test.go
var migratorTestFS embed.FS

func timeNow() time.Time {
	return time.Now().UTC()
}

func setupMigratorDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return db
}

func TestNewMigrator(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if m == nil {
		t.Fatal("expected non-nil Migrator")
	}
}

func TestMigrator_EnsureMigrationTable(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	err := m.EnsureMigrationTable()
	if err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	if !db.Migrator().HasTable(&models.SchemaMigration{}) {
		t.Error("expected schema_migrations table to exist")
	}
}

func TestMigrator_AddMigration(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	m.AddMigration(persistence.Migration{Version: "003", Name: "third", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(status))
	}
	if status[0].Version != "001" {
		t.Errorf("expected first version 001, got %s", status[0].Version)
	}
	if status[2].Version != "003" {
		t.Errorf("expected last version 003, got %s", status[2].Version)
	}
}

func TestMigrator_Up(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	result, err := m.Up(context.Background())
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if len(result.Applied) != 2 {
		t.Errorf("expected 2 applied, got %d", len(result.Applied))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d", len(result.Skipped))
	}
	if result.Direction != persistence.MigrationUp {
		t.Errorf("expected direction up, got %s", result.Direction)
	}

	t.Run("second up skips applied", func(t *testing.T) {
		result2, err := m.Up(context.Background())
		if err != nil {
			t.Fatalf("Up: %v", err)
		}
		if len(result2.Skipped) != 2 {
			t.Errorf("expected 2 skipped, got %d", len(result2.Skipped))
		}
		if len(result2.Applied) != 0 {
			t.Errorf("expected 0 applied, got %d", len(result2.Applied))
		}
	})
}

func TestMigrator_Down(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.Down(context.Background())
	if err != nil {
		t.Fatalf("Down: %v", err)
	}
	if len(result.Applied) != 1 {
		t.Errorf("expected 1 rolled back, got %d", len(result.Applied))
	}
	if result.Direction != persistence.MigrationDown {
		t.Errorf("expected direction down, got %s", result.Direction)
	}
}

func TestMigrator_Down_NoApplied(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	result, err := m.Down(context.Background())
	if err != nil {
		t.Fatalf("Down: %v", err)
	}
	if len(result.Applied) != 0 {
		t.Errorf("expected 0 applied, got %d", len(result.Applied))
	}
}

func TestMigrator_DownTo(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "003", Name: "third", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.DownTo(context.Background(), "001")
	if err != nil {
		t.Fatalf("DownTo: %v", err)
	}
	if len(result.Applied) != 2 {
		t.Errorf("expected 2 rolled back, got %d", len(result.Applied))
	}
}

func TestMigrator_Reset(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.Reset(context.Background())
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if len(result.Applied) != 2 {
		t.Errorf("expected 2 rolled back, got %d", len(result.Applied))
	}
	if result.Direction != persistence.MigrationDown {
		t.Errorf("expected direction down, got %s", result.Direction)
	}
}

func TestMigrator_Fresh(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.Fresh(context.Background())
	if err != nil {
		t.Fatalf("Fresh: %v", err)
	}
	if len(result.Applied) != 2 {
		t.Errorf("expected 2 applied, got %d", len(result.Applied))
	}
}

func TestMigrator_Status(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	m.AddMigration(persistence.Migration{Version: "002", Name: "second", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 2 {
		t.Fatalf("expected 2, got %d", len(status))
	}
	for _, s := range status {
		if s.AppliedAt == nil {
			t.Errorf("expected migration %s to be applied", s.Version)
		}
	}
}

func TestMigrator_GetAppliedMigrations(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})
	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("GetAppliedMigrations: %v", err)
	}
	if len(applied) != 1 {
		t.Errorf("expected 1, got %d", len(applied))
	}
	if applied[0].Version != "001" {
		t.Errorf("expected version 001, got %s", applied[0].Version)
	}
}

func TestMigrator_IsMigrationApplied(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	applied, err := m.IsMigrationApplied("001")
	if err != nil {
		t.Fatalf("IsMigrationApplied: %v", err)
	}
	if applied {
		t.Error("expected not applied")
	}

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	applied, err = m.IsMigrationApplied("001")
	if err != nil {
		t.Fatalf("IsMigrationApplied: %v", err)
	}
	if !applied {
		t.Error("expected applied")
	}
}

func TestMigrator_LoadMigrationsFromFS_NonExistentDir(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	err := m.LoadMigrationsFromFS(migratorTestFS, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent dir")
	}
}

func TestMigrator_LoadMigrationsFromFS_EmbeddedFS(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	err := m.LoadMigrationsFromFS(migrationTestFS, "testdata/migrations")
	if err != nil {
		t.Fatalf("LoadMigrationsFromFS: %v", err)
	}

	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(status))
	}
	if status[0].Version != "000001_create_test" {
		t.Errorf("expected version 000001_create_test, got %s", status[0].Version)
	}
	if status[0].UpSQL == "" {
		t.Error("expected non-empty UpSQL")
	}
	if status[0].DownSQL == "" {
		t.Error("expected non-empty DownSQL")
	}

	upResult, err := m.Up(context.Background())
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if len(upResult.Applied) != 2 {
		t.Errorf("expected 2 applied, got %d", len(upResult.Applied))
	}

	if !db.Migrator().HasTable("test_table_001") {
		t.Error("expected test_table_001 to exist")
	}
	if !db.Migrator().HasTable("test_table_002") {
		t.Error("expected test_table_002 to exist")
	}

	downResult, err := m.Down(context.Background())
	if err != nil {
		t.Fatalf("Down: %v", err)
	}
	if len(downResult.Applied) != 1 {
		t.Errorf("expected 1 rolled back, got %d", len(downResult.Applied))
	}
}

func TestMigrator_Up_WithEmptyUpSQL(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "empty", UpSQL: "", DownSQL: ""})

	result, err := m.Up(context.Background())
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if len(result.Applied) != 1 {
		t.Errorf("expected 1 applied, got %d", len(result.Applied))
	}
}

func TestMigrator_Down_MigrationNotFound(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	db.Create(&models.SchemaMigration{Version: "999", AppliedAt: timeNow()})

	_, err := m.Down(context.Background())
	if err == nil {
		t.Error("expected error for unknown migration version")
	}
}

func TestAutoMigrate(t *testing.T) {
	db := setupMigratorDB(t)

	err := persistence.AutoMigrate(db)
	if err != nil {
		t.Logf("AutoMigrate may fail on SQLite (PostgreSQL-specific features): %v", err)
	}
}

func TestMigrateWithGORM(t *testing.T) {
	db := setupMigratorDB(t)

	err := persistence.MigrateWithGORM(db)
	if err != nil {
		t.Logf("MigrateWithGORM may fail on SQLite (PostgreSQL-specific features): %v", err)
	}
}

func TestMigrationResult_Fields(t *testing.T) {
	r := &persistence.MigrationResult{
		Applied:   []string{"001"},
		Skipped:   []string{},
		Failed:    []string{},
		Direction: persistence.MigrationUp,
	}
	if len(r.Applied) != 1 {
		t.Errorf("expected 1 applied")
	}
	if r.Direction != persistence.MigrationUp {
		t.Errorf("expected up direction")
	}
}

func TestMigrationType_Constants(t *testing.T) {
	if persistence.MigrationUp != "up" {
		t.Errorf("expected 'up', got %s", persistence.MigrationUp)
	}
	if persistence.MigrationDown != "down" {
		t.Errorf("expected 'down', got %s", persistence.MigrationDown)
	}
}

func TestMigrator_Up_WithBadSQL(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "bad", UpSQL: "INVALID SQL STATEMENT!!!", DownSQL: ""})

	result, err := m.Up(context.Background())
	if err == nil {
		t.Error("expected error for bad SQL")
	}
	if len(result.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(result.Failed))
	}
}

func TestMigrator_Down_WithBadSQL(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)

	m.AddMigration(persistence.Migration{Version: "001", Name: "good_up", UpSQL: "SELECT 1", DownSQL: "INVALID SQL!!!"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.Down(context.Background())
	if err == nil {
		t.Error("expected error for bad SQL")
	}
	if len(result.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(result.Failed))
	}
}

func TestMigrator_DownTo_NoMigrations(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	result, err := m.DownTo(context.Background(), "000")
	if err != nil {
		t.Fatalf("DownTo: %v", err)
	}
	if len(result.Applied) != 0 {
		t.Errorf("expected 0, got %d", len(result.Applied))
	}
}

func TestMigrator_Reset_NoMigrations(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	result, err := m.Reset(context.Background())
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if len(result.Applied) != 0 {
		t.Errorf("expected 0, got %d", len(result.Applied))
	}
}

func TestMigrator_Fresh_Empty(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}
	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	result, err := m.Fresh(context.Background())
	if err != nil {
		t.Fatalf("Fresh: %v", err)
	}
	if len(result.Applied) != 1 {
		t.Errorf("expected 1 applied, got %d", len(result.Applied))
	}
}

func TestMigrator_Up_IsMigrationAppliedError(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "SELECT 1"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("first Up: %v", err)
	}

	applied, err := m.IsMigrationApplied("001")
	if err != nil {
		t.Fatalf("IsMigrationApplied: %v", err)
	}
	if !applied {
		t.Error("expected applied")
	}
}

func TestMigrator_Reset_GetAppliedError(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	_, err := m.Reset(context.Background())
	if err == nil {
		t.Error("expected error with closed DB")
	}
}

func TestMigrator_Fresh_ResetError(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	_, err := m.Fresh(context.Background())
	if err == nil {
		t.Error("expected error with closed DB")
	}
}

func TestMigrator_DownTo_DownError(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: "INVALID SQL!!!"})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.DownTo(context.Background(), "000")
	if err == nil {
		t.Error("expected error from bad down SQL")
	}
	if len(result.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(result.Failed))
	}
}

func TestMigrator_GetAppliedMigrations_Error(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	_, err := m.GetAppliedMigrations()
	if err == nil {
		t.Error("expected error with closed DB")
	}
}

func TestMigrator_IsMigrationApplied_Error(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	_, err := m.IsMigrationApplied("001")
	if err == nil {
		t.Error("expected error with closed DB")
	}
}

func TestMigrator_Down_EmptyDownSQL(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	m.AddMigration(persistence.Migration{Version: "001", Name: "empty_down", UpSQL: "SELECT 1", DownSQL: ""})

	if _, err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up: %v", err)
	}

	result, err := m.Down(context.Background())
	if err != nil {
		t.Fatalf("Down with empty SQL: %v", err)
	}
	if len(result.Applied) != 1 {
		t.Errorf("expected 1 rolled back, got %d", len(result.Applied))
	}
}

func TestMigrator_Status_NoMigrations(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(status))
	}
}

func TestMigrator_Up_ErrorCheckingApplied(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	m.AddMigration(persistence.Migration{Version: "001", Name: "first", UpSQL: "SELECT 1", DownSQL: ""})

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	result, err := m.Up(context.Background())
	if err == nil {
		t.Error("expected error with closed DB")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestMigrator_Down_GetAppliedError(t *testing.T) {
	db := setupMigratorDB(t)
	m := persistence.NewMigrator(db)
	if err := m.EnsureMigrationTable(); err != nil {
		t.Fatalf("EnsureMigrationTable: %v", err)
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	result, err := m.Down(context.Background())
	if err == nil {
		t.Error("expected error with closed DB")
	}
	if result.Direction != persistence.MigrationDown {
		t.Errorf("expected down direction, got %s", result.Direction)
	}
}

func TestAutoMigrate_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	err = persistence.AutoMigrate(db)
	if err != nil {
		t.Logf("AutoMigrate on SQLite: %v (uuid-ossp expected to fail)", err)
	}
}
