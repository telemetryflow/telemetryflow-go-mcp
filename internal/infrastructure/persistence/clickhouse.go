// Package persistence provides ClickHouse database connectivity for analytics
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog/log"
)

// ClickHouseConfig holds ClickHouse configuration
type ClickHouseConfig struct {
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	Debug           bool
	DialTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	Compression     string
	Secure          bool
}

// DefaultClickHouseConfig returns default ClickHouse configuration
func DefaultClickHouseConfig() *ClickHouseConfig {
	return &ClickHouseConfig{
		Host:            "localhost",
		Port:            9000,
		Database:        "telemetryflow_analytics",
		Username:        "default",
		Password:        "",
		Debug:           false,
		DialTimeout:     10 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		Compression:     "lz4",
		Secure:          false,
	}
}

// ClickHouse wraps the ClickHouse database connection
type ClickHouse struct {
	conn   driver.Conn
	config *ClickHouseConfig
}

// NewClickHouse creates a new ClickHouse connection
func NewClickHouse(config *ClickHouseConfig) (*ClickHouse, error) {
	if config == nil {
		config = DefaultClickHouseConfig()
	}

	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Debug: config.Debug,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     config.DialTimeout,
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}

	if config.Secure {
		options.TLS = nil // Will use default TLS config
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Database).
		Msg("Connected to ClickHouse database")

	return &ClickHouse{
		conn:   conn,
		config: config,
	}, nil
}

// Conn returns the underlying ClickHouse connection
func (c *ClickHouse) Conn() driver.Conn {
	return c.conn
}

// Ping checks the database connection
func (c *ClickHouse) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// Close closes the database connection
func (c *ClickHouse) Close() error {
	return c.conn.Close()
}

// CreateTables creates the analytics tables
func (c *ClickHouse) CreateTables(ctx context.Context) error {
	tables := []string{
		// Tool call analytics table
		`CREATE TABLE IF NOT EXISTS tool_call_analytics (
			timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
			session_id UUID,
			conversation_id UUID,
			tool_name LowCardinality(String),
			duration_ms UInt64,
			is_error UInt8,
			input_size UInt32,
			output_size UInt32
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (timestamp, tool_name, session_id)
		TTL timestamp + INTERVAL 90 DAY`,

		// API request analytics table
		`CREATE TABLE IF NOT EXISTS api_request_analytics (
			timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
			session_id UUID,
			conversation_id UUID,
			model LowCardinality(String),
			input_tokens UInt32,
			output_tokens UInt32,
			total_tokens UInt32,
			duration_ms UInt64,
			status_code UInt16,
			is_error UInt8
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (timestamp, model, session_id)
		TTL timestamp + INTERVAL 90 DAY`,

		// Session analytics table
		`CREATE TABLE IF NOT EXISTS session_analytics (
			timestamp DateTime64(3) CODEC(Delta, ZSTD(1)),
			session_id UUID,
			event_type LowCardinality(String),
			client_name String,
			client_version String,
			duration_ms UInt64,
			message_count UInt32,
			tool_call_count UInt32,
			total_tokens UInt64
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (timestamp, event_type, session_id)
		TTL timestamp + INTERVAL 180 DAY`,

		// Token usage aggregates (materialized view)
		`CREATE TABLE IF NOT EXISTS token_usage_hourly (
			hour DateTime CODEC(Delta, ZSTD(1)),
			model LowCardinality(String),
			input_tokens UInt64,
			output_tokens UInt64,
			total_tokens UInt64,
			request_count UInt64
		) ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(hour)
		ORDER BY (hour, model)`,

		// Tool usage aggregates
		`CREATE TABLE IF NOT EXISTS tool_usage_hourly (
			hour DateTime CODEC(Delta, ZSTD(1)),
			tool_name LowCardinality(String),
			call_count UInt64,
			error_count UInt64,
			total_duration_ms UInt64,
			avg_duration_ms Float64
		) ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(hour)
		ORDER BY (hour, tool_name)`,
	}

	for _, table := range tables {
		if err := c.conn.Exec(ctx, table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	log.Info().Msg("ClickHouse analytics tables created")
	return nil
}

// ToolCallEvent represents a tool call analytics event
type ToolCallEvent struct {
	Timestamp      time.Time
	SessionID      string
	ConversationID string
	ToolName       string
	DurationMs     uint64
	IsError        bool
	InputSize      uint32
	OutputSize     uint32
}

// APIRequestEvent represents an API request analytics event
type APIRequestEvent struct {
	Timestamp      time.Time
	SessionID      string
	ConversationID string
	Model          string
	InputTokens    uint32
	OutputTokens   uint32
	TotalTokens    uint32
	DurationMs     uint64
	StatusCode     uint16
	IsError        bool
}

// SessionEvent represents a session analytics event
type SessionEvent struct {
	Timestamp     time.Time
	SessionID     string
	EventType     string
	ClientName    string
	ClientVersion string
	DurationMs    uint64
	MessageCount  uint32
	ToolCallCount uint32
	TotalTokens   uint64
}

// InsertToolCallEvent inserts a tool call analytics event
func (c *ClickHouse) InsertToolCallEvent(ctx context.Context, event *ToolCallEvent) error {
	isError := uint8(0)
	if event.IsError {
		isError = 1
	}

	return c.conn.Exec(ctx, `
		INSERT INTO tool_call_analytics
		(timestamp, session_id, conversation_id, tool_name, duration_ms, is_error, input_size, output_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.Timestamp,
		event.SessionID,
		event.ConversationID,
		event.ToolName,
		event.DurationMs,
		isError,
		event.InputSize,
		event.OutputSize,
	)
}

// InsertAPIRequestEvent inserts an API request analytics event
func (c *ClickHouse) InsertAPIRequestEvent(ctx context.Context, event *APIRequestEvent) error {
	isError := uint8(0)
	if event.IsError {
		isError = 1
	}

	return c.conn.Exec(ctx, `
		INSERT INTO api_request_analytics
		(timestamp, session_id, conversation_id, model, input_tokens, output_tokens, total_tokens, duration_ms, status_code, is_error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.Timestamp,
		event.SessionID,
		event.ConversationID,
		event.Model,
		event.InputTokens,
		event.OutputTokens,
		event.TotalTokens,
		event.DurationMs,
		event.StatusCode,
		isError,
	)
}

// InsertSessionEvent inserts a session analytics event
func (c *ClickHouse) InsertSessionEvent(ctx context.Context, event *SessionEvent) error {
	return c.conn.Exec(ctx, `
		INSERT INTO session_analytics
		(timestamp, session_id, event_type, client_name, client_version, duration_ms, message_count, tool_call_count, total_tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.Timestamp,
		event.SessionID,
		event.EventType,
		event.ClientName,
		event.ClientVersion,
		event.DurationMs,
		event.MessageCount,
		event.ToolCallCount,
		event.TotalTokens,
	)
}

// BatchInsert provides batch insert functionality
type BatchInsert struct {
	ch        *ClickHouse
	batchSize int
	events    []interface{}
	tableName string
}

// NewBatchInsert creates a new batch insert
func (c *ClickHouse) NewBatchInsert(tableName string, batchSize int) *BatchInsert {
	return &BatchInsert{
		ch:        c,
		batchSize: batchSize,
		events:    make([]interface{}, 0, batchSize),
		tableName: tableName,
	}
}

// Add adds an event to the batch
func (b *BatchInsert) Add(event interface{}) error {
	b.events = append(b.events, event)
	if len(b.events) >= b.batchSize {
		return b.Flush(context.Background())
	}
	return nil
}

// Flush flushes the batch to the database
func (b *BatchInsert) Flush(ctx context.Context) error {
	if len(b.events) == 0 {
		return nil
	}

	batch, err := b.ch.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", b.tableName))
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, event := range b.events {
		switch e := event.(type) {
		case *ToolCallEvent:
			isError := uint8(0)
			if e.IsError {
				isError = 1
			}
			if err := batch.Append(
				e.Timestamp,
				e.SessionID,
				e.ConversationID,
				e.ToolName,
				e.DurationMs,
				isError,
				e.InputSize,
				e.OutputSize,
			); err != nil {
				return err
			}
		case *APIRequestEvent:
			isError := uint8(0)
			if e.IsError {
				isError = 1
			}
			if err := batch.Append(
				e.Timestamp,
				e.SessionID,
				e.ConversationID,
				e.Model,
				e.InputTokens,
				e.OutputTokens,
				e.TotalTokens,
				e.DurationMs,
				e.StatusCode,
				isError,
			); err != nil {
				return err
			}
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	b.events = b.events[:0]
	return nil
}

// HealthCheck performs a health check on ClickHouse
func (c *ClickHouse) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("ClickHouse health check failed: %w", err)
	}
	return nil
}
