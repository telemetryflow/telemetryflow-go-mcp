package persistence_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
)

func TestClickHouseConfig_Defaults(t *testing.T) {
	cfg := persistence.DefaultClickHouseConfig()
	if cfg.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Errorf("expected 9000, got %d", cfg.Port)
	}
	if cfg.Database != "telemetryflow_analytics" {
		t.Errorf("expected telemetryflow_analytics, got %s", cfg.Database)
	}
	if cfg.Username != "default" {
		t.Errorf("expected default, got %s", cfg.Username)
	}
	if cfg.DialTimeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.DialTimeout)
	}
	if cfg.MaxOpenConns != 10 {
		t.Errorf("expected 10, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("expected 5, got %d", cfg.MaxIdleConns)
	}
	if cfg.Compression != "lz4" {
		t.Errorf("expected lz4, got %s", cfg.Compression)
	}
	if cfg.Secure {
		t.Error("expected secure false")
	}
	if cfg.Debug {
		t.Error("expected debug false")
	}
	if cfg.ConnMaxLifetime != time.Hour {
		t.Errorf("expected 1h, got %v", cfg.ConnMaxLifetime)
	}
}

func TestClickHouse_Ping(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockClickHouseConn{
			pingFn: func(_ context.Context) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.Ping(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			pingFn: func(_ context.Context) error { return errors.New("ping failed") },
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.Ping(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouse_Close(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockClickHouseConn{
			closeFn: func() error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			closeFn: func() error { return errors.New("close failed") },
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.Close(); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouse_Conn(t *testing.T) {
	mock := &mockClickHouseConn{}
	ch := newClickHouseWithMockConn(t, mock)
	if ch.Conn() == nil {
		t.Fatal("expected non-nil conn")
	}
}

func TestClickHouse_CreateTables(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		execCount := 0
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error {
				execCount++
				return nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.CreateTables(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if execCount != 5 {
			t.Errorf("expected 5 exec calls, got %d", execCount)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error {
				return errors.New("exec failed")
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.CreateTables(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouse_InsertToolCallEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertToolCallEvent(context.Background(), &persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test", DurationMs: 100, IsError: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("isError true", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertToolCallEvent(context.Background(), &persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test", DurationMs: 100, IsError: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error {
				return errors.New("exec failed")
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertToolCallEvent(context.Background(), &persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test", DurationMs: 100,
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouse_InsertAPIRequestEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertAPIRequestEvent(context.Background(), &persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "test", StatusCode: 200, IsError: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("isError true", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertAPIRequestEvent(context.Background(), &persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "test", StatusCode: 500, IsError: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error {
				return errors.New("exec failed")
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertAPIRequestEvent(context.Background(), &persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "test", StatusCode: 200,
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouse_InsertSessionEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertSessionEvent(context.Background(), &persistence.SessionEvent{
			Timestamp: time.Now(), SessionID: "s1", EventType: "active",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			execFn: func(_ context.Context, _ string, _ ...interface{}) error {
				return errors.New("exec failed")
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		err := ch.InsertSessionEvent(context.Background(), &persistence.SessionEvent{
			Timestamp: time.Now(), SessionID: "s1", EventType: "active",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouse_BatchInsert(t *testing.T) {
	t.Run("NewBatchInsert", func(t *testing.T) {
		ch := newClickHouseWithMockConn(t, &mockClickHouseConn{})
		batch := ch.NewBatchInsert("test_table", 10)
		if batch == nil {
			t.Fatal("expected non-nil batch")
		}
	})

	t.Run("Add below batch size", func(t *testing.T) {
		ch := newClickHouseWithMockConn(t, &mockClickHouseConn{})
		batch := ch.NewBatchInsert("test_table", 100)
		err := batch.Add(&persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test",
		})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Add API request event below batch size", func(t *testing.T) {
		ch := newClickHouseWithMockConn(t, &mockClickHouseConn{})
		batch := ch.NewBatchInsert("test_table", 100)
		err := batch.Add(&persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "test", StatusCode: 200,
		})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Flush empty batch", func(t *testing.T) {
		ch := newClickHouseWithMockConn(t, &mockClickHouseConn{})
		batch := ch.NewBatchInsert("test_table", 10)
		err := batch.Flush(context.Background())
		if err != nil {
			t.Fatalf("Flush empty: %v", err)
		}
	})

	t.Run("Add triggers flush at batch size with ToolCallEvent", func(t *testing.T) {
		var appended bool
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{}, nil
			},
		}
		_ = appended
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add(&persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test", DurationMs: 100, IsError: false,
		})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Add triggers flush at batch size with ToolCallEvent isError", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{}, nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add(&persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test", DurationMs: 100, IsError: true,
		})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Add triggers flush at batch size with APIRequestEvent", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{}, nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add(&persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "test", StatusCode: 200, IsError: false,
		})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Add triggers flush at batch size with APIRequestEvent isError", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{}, nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add(&persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "test", StatusCode: 500, IsError: true,
		})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Flush with unknown event type", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{}, nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add("unknown event type")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("Flush with prepare error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return nil, errors.New("prepare failed")
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add(&persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Flush with send error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{sendErr: errors.New("send failed")}, nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 1)
		err := batch.Add(&persistence.ToolCallEvent{
			Timestamp: time.Now(), ToolName: "test",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Manual flush with events", func(t *testing.T) {
		mock := &mockClickHouseConn{
			prepareFn: func(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
				return &mockClickHouseBatch{}, nil
			},
		}
		ch := newClickHouseWithMockConn(t, mock)
		batch := ch.NewBatchInsert("test_table", 100)
		_ = batch.Add(&persistence.ToolCallEvent{Timestamp: time.Now(), ToolName: "t1"})
		_ = batch.Add(&persistence.APIRequestEvent{Timestamp: time.Now(), Model: "m1"})
		err := batch.Flush(context.Background())
		if err != nil {
			t.Fatalf("Flush: %v", err)
		}
	})
}

func TestClickHouse_HealthCheck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockClickHouseConn{
			pingFn: func(_ context.Context) error { return nil },
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.HealthCheck(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			pingFn: func(_ context.Context) error { return errors.New("ping failed") },
		}
		ch := newClickHouseWithMockConn(t, mock)
		if err := ch.HealthCheck(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClickHouseEventTypes(t *testing.T) {
	t.Run("ToolCallEvent", func(t *testing.T) {
		event := &persistence.ToolCallEvent{
			Timestamp: time.Now(), SessionID: "session-1", ConversationID: "conv-1",
			ToolName: "echo", DurationMs: 100, IsError: false, InputSize: 50, OutputSize: 200,
		}
		if event.ToolName != "echo" {
			t.Errorf("expected echo, got %s", event.ToolName)
		}
	})

	t.Run("APIRequestEvent", func(t *testing.T) {
		event := &persistence.APIRequestEvent{
			Timestamp: time.Now(), Model: "claude-opus-4-7",
			InputTokens: 100, OutputTokens: 50, TotalTokens: 150,
			DurationMs: 200, StatusCode: 200, IsError: false,
		}
		if event.TotalTokens != 150 {
			t.Errorf("expected 150, got %d", event.TotalTokens)
		}
	})

	t.Run("SessionEvent", func(t *testing.T) {
		event := &persistence.SessionEvent{
			EventType: "session_active", MessageCount: 10, ToolCallCount: 5, TotalTokens: 1000,
		}
		if event.EventType != "session_active" {
			t.Errorf("expected session_active, got %s", event.EventType)
		}
	})
}
