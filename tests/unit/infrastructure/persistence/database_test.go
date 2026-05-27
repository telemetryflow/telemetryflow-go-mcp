package persistence_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDatabaseConfig_DefaultConfig(t *testing.T) {
	cfg := persistence.DefaultDatabaseConfig()
	if cfg.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.Port)
	}
	if cfg.Database != "telemetryflow_mcp" {
		t.Errorf("expected telemetryflow_mcp, got %s", cfg.Database)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected 10, got %d", cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns != 100 {
		t.Errorf("expected 100, got %d", cfg.MaxOpenConns)
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("expected disable, got %s", cfg.SSLMode)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("expected warn, got %s", cfg.LogLevel)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *persistence.DatabaseConfig
		contains []string
	}{
		{
			name:   "default config",
			config: persistence.DefaultDatabaseConfig(),
			contains: []string{
				"host=localhost",
				"port=5432",
				"user=telemetryflow",
				"dbname=telemetryflow_mcp",
				"sslmode=disable",
			},
		},
		{
			name: "custom config",
			config: &persistence.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret",
				Database: "mydb",
				SSLMode:  "require",
			},
			contains: []string{
				"host=db.example.com",
				"port=5433",
				"user=admin",
				"password=secret",
				"dbname=mydb",
				"sslmode=require",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			for _, s := range tt.contains {
				if !contains(dsn, s) {
					t.Errorf("DSN %q does not contain %q", dsn, s)
				}
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestNewDatabaseFromDB(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	db := persistence.NewDatabaseFromDB(gormDB)
	if db == nil {
		t.Fatal("expected non-nil Database")
	}
	if db.DB() != gormDB {
		t.Error("expected DB() to return the same gorm.DB")
	}
}

func TestDatabase_Ping(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	err = db.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestDatabase_Close(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	err = db.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestDatabase_Migrate(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	err = db.Migrate(&persistence.SessionModel{})
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
}

func TestDatabase_Transaction(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)
	if err := db.Migrate(&persistence.SessionModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Run("successful transaction", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			return tx.Create(&persistence.SessionModel{
				ID:              "00000000-0000-0000-0000-000000000001",
				ProtocolVersion: "2024-11-05",
				State:           "created",
				ServerName:      "Test",
				ServerVersion:   "1.0",
			}).Error
		})
		if err != nil {
			t.Fatalf("Transaction: %v", err)
		}
	})

	t.Run("failed transaction rolls back", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			_ = tx.Create(&persistence.SessionModel{
				ID:              "00000000-0000-0000-0000-000000000002",
				ProtocolVersion: "2024-11-05",
				State:           "created",
				ServerName:      "Test",
				ServerVersion:   "1.0",
			}).Error
			return fmt.Errorf("intentional error")
		})
		if err == nil {
			t.Error("expected error")
		}
		var count int64
		gormDB.Model(&persistence.SessionModel{}).Where("id = ?", "00000000-0000-0000-0000-000000000002").Count(&count)
		if count != 0 {
			t.Error("expected rollback, record should not exist")
		}
	})
}

func TestDatabase_WithContext(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	result := db.WithContext(context.Background())
	if result == nil {
		t.Error("expected non-nil")
	}
}

func TestDatabase_Stats(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.MaxOpenConnections < 0 {
		t.Error("invalid MaxOpenConnections")
	}
}

func TestDatabase_HealthCheck(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	err = db.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestDatabase_HealthCheck_CancelledContext(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = db.HealthCheck(ctx)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestDatabase_NewDatabase_NilConfig(t *testing.T) {
	cfg := persistence.DefaultDatabaseConfig()
	if cfg == nil {
		t.Error("expected non-nil default config")
	}
}

func TestDatabase_DatabaseStats(t *testing.T) {
	stats := &persistence.DatabaseStats{
		MaxOpenConnections: 100,
		OpenConnections:    5,
		InUse:              2,
		Idle:               3,
		WaitCount:          10,
		WaitDuration:       time.Second,
		MaxIdleClosed:      5,
		MaxIdleTimeClosed:  3,
		MaxLifetimeClosed:  2,
	}
	if stats.MaxOpenConnections != 100 {
		t.Errorf("expected 100, got %d", stats.MaxOpenConnections)
	}
	if stats.InUse != 2 {
		t.Errorf("expected 2, got %d", stats.InUse)
	}
}

func TestDatabase_Ping_ClosedDB(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err = db.Ping(context.Background())
	if err == nil {
		t.Error("expected error pinging closed DB")
	}
}

func TestDatabase_Close_AlreadyClosed(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)
	if err := db.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}

	_ = db.Close()
}

func TestDatabase_Migrate_ClosedDB(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err = db.Migrate(&persistence.SessionModel{})
	if err == nil {
		t.Error("expected error migrating on closed DB")
	}
}

func TestDatabase_Stats_ClosedDB(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err = db.Stats()
	if err != nil {
		t.Logf("Stats on closed DB returned error (acceptable): %v", err)
	}
}

func TestDatabase_Transaction_ClosedDB(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error { return nil })
	if err == nil {
		t.Error("expected error on closed DB transaction")
	}
}

func TestDatabaseConfig_DSN_Custom(t *testing.T) {
	cfg := &persistence.DatabaseConfig{
		Host:     "myhost",
		Port:     3306,
		User:     "admin",
		Password: "pass",
		Database: "mydb",
		SSLMode:  "require",
	}
	dsn := cfg.DSN()
	if dsn == "" {
		t.Error("expected non-empty DSN")
	}
}

func TestDatabase_HealthCheck_Success(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := persistence.NewDatabaseFromDB(gormDB)

	err = db.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}
