package persistence_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
)

func TestAnalyticsRepository_New(t *testing.T) {
	ch := &persistence.ClickHouse{}
	repo := persistence.NewAnalyticsRepository(ch)
	if repo == nil {
		t.Fatal("expected non-nil AnalyticsRepository")
	}
}

func TestAnalyticsRepository_GetTokenUsageByModel(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return &mockClickHouseRows{
					maxRows: 1,
					scanValues: [][]interface{}{
						{"model-a", uint64(100), uint64(50), uint64(150), uint64(10), 10.0, 5.0},
					},
				}, nil
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		stats, err := repo.GetTokenUsageByModel(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(stats) != 1 {
			t.Fatalf("expected 1 stat, got %d", len(stats))
		}
		if stats[0].Model != "model-a" {
			t.Errorf("expected model-a, got %s", stats[0].Model)
		}
		if stats[0].TotalTokens != 150 {
			t.Errorf("expected 150 tokens, got %d", stats[0].TotalTokens)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return nil, errors.New("query failed")
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		_, err := repo.GetTokenUsageByModel(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty result", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return &mockClickHouseRows{maxRows: 0}, nil
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		stats, err := repo.GetTokenUsageByModel(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(stats) != 0 {
			t.Fatalf("expected 0 stats, got %d", len(stats))
		}
	})
}

func TestAnalyticsRepository_GetToolUsageStats(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return &mockClickHouseRows{
					maxRows: 1,
					scanValues: [][]interface{}{
						{"tool-a", uint64(100), uint64(5), 0.95, 50.0, 45.0, 80.0, 120.0},
					},
				}, nil
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		stats, err := repo.GetToolUsageStats(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(stats) != 1 {
			t.Fatalf("expected 1 stat, got %d", len(stats))
		}
		if stats[0].ToolName != "tool-a" {
			t.Errorf("expected tool-a, got %s", stats[0].ToolName)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return nil, errors.New("query failed")
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		_, err := repo.GetToolUsageStats(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAnalyticsRepository_GetSessionStats(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return &mockClickHouseRows{
					maxRows: 1,
					scanValues: [][]interface{}{
						{uint64(100), uint64(50), 5000.0, 10.5, 3.2, 5000.0},
					},
				}, nil
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		stats, err := repo.GetSessionStats(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalSessions != 100 {
			t.Errorf("expected 100, got %d", stats.TotalSessions)
		}
	})

	t.Run("no rows", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return &mockClickHouseRows{maxRows: 0}, nil
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		stats, err := repo.GetSessionStats(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalSessions != 0 {
			t.Errorf("expected 0, got %d", stats.TotalSessions)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return nil, errors.New("query failed")
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		_, err := repo.GetSessionStats(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAnalyticsRepository_GetRequestsTimeSeries(t *testing.T) {
	mock := &mockClickHouseConn{
		queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
			return &mockClickHouseRows{
				maxRows: 2,
				scanValues: [][]interface{}{
					{time.Now(), 100.0},
					{time.Now().Add(time.Hour), 200.0},
				},
			}, nil
		},
	}
	repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
	points, err := repo.GetRequestsTimeSeries(context.Background(), time.Now().Add(-time.Hour), time.Now(), "1 HOUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
}

func TestAnalyticsRepository_GetTokensTimeSeries(t *testing.T) {
	mock := &mockClickHouseConn{
		queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
			return &mockClickHouseRows{
				maxRows: 1,
				scanValues: [][]interface{}{
					{time.Now(), 5000.0},
				},
			}, nil
		},
	}
	repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
	points, err := repo.GetTokensTimeSeries(context.Background(), time.Now().Add(-time.Hour), time.Now(), "1 HOUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
}

func TestAnalyticsRepository_GetLatencyTimeSeries(t *testing.T) {
	mock := &mockClickHouseConn{
		queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
			return &mockClickHouseRows{
				maxRows: 1,
				scanValues: [][]interface{}{
					{time.Now(), 250.0},
				},
			}, nil
		},
	}
	repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
	points, err := repo.GetLatencyTimeSeries(context.Background(), time.Now().Add(-time.Hour), time.Now(), "1 HOUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
}

func TestAnalyticsRepository_GetErrorRate(t *testing.T) {
	mock := &mockClickHouseConn{
		queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
			return &mockClickHouseRows{
				maxRows: 1,
				scanValues: [][]interface{}{
					{time.Now(), 2.5},
				},
			}, nil
		},
	}
	repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
	points, err := repo.GetErrorRate(context.Background(), time.Now().Add(-time.Hour), time.Now(), "1 HOUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
}

func TestAnalyticsRepository_GetTopTools(t *testing.T) {
	mock := &mockClickHouseConn{
		queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
			return &mockClickHouseRows{
				maxRows: 1,
				scanValues: [][]interface{}{
					{"top-tool", uint64(500), uint64(10), 0.98, 30.0, 25.0, 50.0, 100.0},
				},
			}, nil
		},
	}
	repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
	stats, err := repo.GetTopTools(context.Background(), time.Now().Add(-time.Hour), time.Now(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].ToolName != "top-tool" {
		t.Errorf("expected top-tool, got %s", stats[0].ToolName)
	}
}

func TestAnalyticsRepository_GetDashboardSummary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		callCount := 0
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				callCount++
				switch callCount {
				case 1:
					return &mockClickHouseRows{
						maxRows: 1,
						scanValues: [][]interface{}{
							{uint64(1000), uint64(500000), 250.0, 2.5},
						},
					}, nil
				case 2:
					return &mockClickHouseRows{
						maxRows: 1,
						scanValues: [][]interface{}{
							{uint64(500)},
						},
					}, nil
				case 3:
					return &mockClickHouseRows{
						maxRows: 1,
						scanValues: [][]interface{}{
							{uint64(100)},
						},
					}, nil
				default:
					return &mockClickHouseRows{maxRows: 0}, nil
				}
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		summary, err := repo.GetDashboardSummary(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.TotalRequests != 1000 {
			t.Errorf("expected 1000, got %d", summary.TotalRequests)
		}
		if summary.TotalToolCalls != 500 {
			t.Errorf("expected 500, got %d", summary.TotalToolCalls)
		}
		if summary.TotalSessions != 100 {
			t.Errorf("expected 100, got %d", summary.TotalSessions)
		}
		if summary.RequestsPerMinute <= 0 {
			t.Errorf("expected positive requests per minute, got %f", summary.RequestsPerMinute)
		}
	})

	t.Run("first query error", func(t *testing.T) {
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				return nil, errors.New("query failed")
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		_, err := repo.GetDashboardSummary(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("second query error", func(t *testing.T) {
		callCount := 0
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				callCount++
				if callCount == 1 {
					return &mockClickHouseRows{
						maxRows: 1,
						scanValues: [][]interface{}{
							{uint64(1000), uint64(500000), 250.0, 2.5},
						},
					}, nil
				}
				return nil, errors.New("query failed")
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		_, err := repo.GetDashboardSummary(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("third query error", func(t *testing.T) {
		callCount := 0
		mock := &mockClickHouseConn{
			queryFn: func(_ context.Context, _ string, _ ...interface{}) (driver.Rows, error) {
				callCount++
				switch callCount {
				case 1:
					return &mockClickHouseRows{
						maxRows: 1,
						scanValues: [][]interface{}{
							{uint64(1000), uint64(500000), 250.0, 2.5},
						},
					}, nil
				case 2:
					return &mockClickHouseRows{
						maxRows: 1,
						scanValues: [][]interface{}{
							{uint64(500)},
						},
					}, nil
				default:
					return nil, errors.New("query failed")
				}
			},
		}
		repo := persistence.NewAnalyticsRepository(newClickHouseWithMockConn(t, mock))
		_, err := repo.GetDashboardSummary(context.Background(), time.Now().Add(-time.Hour), time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAnalyticsStatsTypes(t *testing.T) {
	t.Run("TokenUsageStats", func(t *testing.T) {
		stats := persistence.TokenUsageStats{
			Model: "claude-opus-4-7", InputTokens: 1000, OutputTokens: 500,
			TotalTokens: 1500, RequestCount: 10, AvgInputSize: 100.0, AvgOutputSize: 50.0,
		}
		if stats.TotalTokens != 1500 {
			t.Errorf("expected 1500, got %d", stats.TotalTokens)
		}
	})

	t.Run("ToolUsageStats", func(t *testing.T) {
		stats := persistence.ToolUsageStats{
			ToolName: "echo", CallCount: 100, ErrorCount: 5,
			SuccessRate: 0.95, AvgDurationMs: 50.0, P50DurationMs: 45.0,
			P95DurationMs: 80.0, P99DurationMs: 120.0,
		}
		if stats.SuccessRate != 0.95 {
			t.Errorf("expected 0.95, got %f", stats.SuccessRate)
		}
	})

	t.Run("SessionStats", func(t *testing.T) {
		stats := persistence.SessionStats{
			TotalSessions: 100, ActiveSessions: 50, AvgDurationMs: 5000.0,
			AvgMessages: 10.5, AvgToolCalls: 3.2, AvgTokens: 5000.0,
		}
		if stats.TotalSessions != 100 {
			t.Errorf("expected 100, got %d", stats.TotalSessions)
		}
	})

	t.Run("TimeSeriesPoint", func(t *testing.T) {
		p := persistence.TimeSeriesPoint{Timestamp: time.Now(), Value: 42.0}
		if p.Value != 42.0 {
			t.Errorf("expected 42.0, got %f", p.Value)
		}
	})

	t.Run("DashboardSummary", func(t *testing.T) {
		summary := persistence.DashboardSummary{
			TotalRequests: 1000, TotalTokens: 500000, TotalToolCalls: 500,
			TotalSessions: 100, AvgLatencyMs: 250.0, ErrorRate: 2.5, RequestsPerMinute: 16.67,
		}
		if summary.TotalRequests != 1000 {
			t.Errorf("expected 1000, got %d", summary.TotalRequests)
		}
	})
}
