// Package persistence provides database connectivity and repositories
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogLevel        string
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "telemetryflow",
		Password:        "",
		Database:        "telemetryflow_mcp",
		SSLMode:         "disable",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        "warn",
	}
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// Database wraps the GORM database connection
type Database struct {
	db     *gorm.DB
	config *DatabaseConfig
}

// NewDatabase creates a new database connection
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	// Configure GORM logger
	var gormLogLevel logger.LogLevel
	switch config.LogLevel {
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	default:
		gormLogLevel = logger.Warn
	}

	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(gormLogLevel),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}

	db, err := gorm.Open(postgres.Open(config.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	log.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Database).
		Msg("Connected to PostgreSQL database")

	return &Database{
		db:     db,
		config: config,
	}, nil
}

// DB returns the underlying GORM database
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Ping checks the database connection
func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate runs database migrations
func (d *Database) Migrate(models ...interface{}) error {
	if err := d.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Info().Msg("Database migrations completed")
	return nil
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}

// WithContext returns a database connection with context
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// Stats returns database connection pool statistics
func (d *Database) Stats() (*DatabaseStats, error) {
	sqlDB, err := d.db.DB()
	if err != nil {
		return nil, err
	}
	stats := sqlDB.Stats()
	return &DatabaseStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, nil
}

// DatabaseStats holds database connection pool statistics
type DatabaseStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxIdleTimeClosed  int64
	MaxLifetimeClosed  int64
}

// HealthCheck performs a health check on the database
func (d *Database) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
