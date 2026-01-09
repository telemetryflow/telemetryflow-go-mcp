// Package persistence provides analytics repository for ClickHouse
package persistence

import (
	"context"
	"fmt"
	"time"
)

// AnalyticsRepository handles analytics queries on ClickHouse
type AnalyticsRepository struct {
	ch *ClickHouse
}

// NewAnalyticsRepository creates a new AnalyticsRepository
func NewAnalyticsRepository(ch *ClickHouse) *AnalyticsRepository {
	return &AnalyticsRepository{ch: ch}
}

// TokenUsageStats represents token usage statistics
type TokenUsageStats struct {
	Model         string
	InputTokens   uint64
	OutputTokens  uint64
	TotalTokens   uint64
	RequestCount  uint64
	AvgInputSize  float64
	AvgOutputSize float64
}

// ToolUsageStats represents tool usage statistics
type ToolUsageStats struct {
	ToolName      string
	CallCount     uint64
	ErrorCount    uint64
	SuccessRate   float64
	AvgDurationMs float64
	P50DurationMs float64
	P95DurationMs float64
	P99DurationMs float64
}

// SessionStats represents session statistics
type SessionStats struct {
	TotalSessions  uint64
	ActiveSessions uint64
	AvgDurationMs  float64
	AvgMessages    float64
	AvgToolCalls   float64
	AvgTokens      float64
}

// TimeSeriesPoint represents a point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// GetTokenUsageByModel returns token usage statistics by model
func (r *AnalyticsRepository) GetTokenUsageByModel(ctx context.Context, since, until time.Time) ([]TokenUsageStats, error) {
	query := `
		SELECT
			model,
			sum(input_tokens) as input_tokens,
			sum(output_tokens) as output_tokens,
			sum(total_tokens) as total_tokens,
			count() as request_count,
			avg(input_tokens) as avg_input_size,
			avg(output_tokens) as avg_output_size
		FROM api_request_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY model
		ORDER BY total_tokens DESC
	`

	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query token usage: %w", err)
	}
	defer rows.Close()

	var stats []TokenUsageStats
	for rows.Next() {
		var s TokenUsageStats
		if err := rows.Scan(
			&s.Model,
			&s.InputTokens,
			&s.OutputTokens,
			&s.TotalTokens,
			&s.RequestCount,
			&s.AvgInputSize,
			&s.AvgOutputSize,
		); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// GetToolUsageStats returns tool usage statistics
func (r *AnalyticsRepository) GetToolUsageStats(ctx context.Context, since, until time.Time) ([]ToolUsageStats, error) {
	query := `
		SELECT
			tool_name,
			count() as call_count,
			countIf(is_error = 1) as error_count,
			1 - (countIf(is_error = 1) / count()) as success_rate,
			avg(duration_ms) as avg_duration_ms,
			quantile(0.5)(duration_ms) as p50_duration_ms,
			quantile(0.95)(duration_ms) as p95_duration_ms,
			quantile(0.99)(duration_ms) as p99_duration_ms
		FROM tool_call_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY tool_name
		ORDER BY call_count DESC
	`

	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool usage: %w", err)
	}
	defer rows.Close()

	var stats []ToolUsageStats
	for rows.Next() {
		var s ToolUsageStats
		if err := rows.Scan(
			&s.ToolName,
			&s.CallCount,
			&s.ErrorCount,
			&s.SuccessRate,
			&s.AvgDurationMs,
			&s.P50DurationMs,
			&s.P95DurationMs,
			&s.P99DurationMs,
		); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// GetSessionStats returns session statistics
func (r *AnalyticsRepository) GetSessionStats(ctx context.Context, since, until time.Time) (*SessionStats, error) {
	query := `
		SELECT
			count() as total_sessions,
			countIf(event_type = 'session_active') as active_sessions,
			avg(duration_ms) as avg_duration_ms,
			avg(message_count) as avg_messages,
			avg(tool_call_count) as avg_tool_calls,
			avg(total_tokens) as avg_tokens
		FROM session_analytics
		WHERE timestamp >= ? AND timestamp <= ?
	`

	var stats SessionStats
	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query session stats: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(
			&stats.TotalSessions,
			&stats.ActiveSessions,
			&stats.AvgDurationMs,
			&stats.AvgMessages,
			&stats.AvgToolCalls,
			&stats.AvgTokens,
		); err != nil {
			return nil, err
		}
	}

	return &stats, rows.Err()
}

// GetRequestsTimeSeries returns requests count over time
func (r *AnalyticsRepository) GetRequestsTimeSeries(ctx context.Context, since, until time.Time, interval string) ([]TimeSeriesPoint, error) {
	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(timestamp, INTERVAL %s) as bucket,
			count() as value
		FROM api_request_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY bucket
		ORDER BY bucket
	`, interval)

	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests time series: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// GetTokensTimeSeries returns tokens usage over time
func (r *AnalyticsRepository) GetTokensTimeSeries(ctx context.Context, since, until time.Time, interval string) ([]TimeSeriesPoint, error) {
	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(timestamp, INTERVAL %s) as bucket,
			sum(total_tokens) as value
		FROM api_request_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY bucket
		ORDER BY bucket
	`, interval)

	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens time series: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// GetLatencyTimeSeries returns average latency over time
func (r *AnalyticsRepository) GetLatencyTimeSeries(ctx context.Context, since, until time.Time, interval string) ([]TimeSeriesPoint, error) {
	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(timestamp, INTERVAL %s) as bucket,
			avg(duration_ms) as value
		FROM api_request_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY bucket
		ORDER BY bucket
	`, interval)

	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query latency time series: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// GetErrorRate returns the error rate over time
func (r *AnalyticsRepository) GetErrorRate(ctx context.Context, since, until time.Time, interval string) ([]TimeSeriesPoint, error) {
	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(timestamp, INTERVAL %s) as bucket,
			countIf(is_error = 1) * 100.0 / count() as value
		FROM api_request_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY bucket
		ORDER BY bucket
	`, interval)

	rows, err := r.ch.conn.Query(ctx, query, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query error rate: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// GetTopTools returns the most used tools
func (r *AnalyticsRepository) GetTopTools(ctx context.Context, since, until time.Time, limit int) ([]ToolUsageStats, error) {
	query := `
		SELECT
			tool_name,
			count() as call_count,
			countIf(is_error = 1) as error_count,
			1 - (countIf(is_error = 1) / count()) as success_rate,
			avg(duration_ms) as avg_duration_ms,
			quantile(0.5)(duration_ms) as p50_duration_ms,
			quantile(0.95)(duration_ms) as p95_duration_ms,
			quantile(0.99)(duration_ms) as p99_duration_ms
		FROM tool_call_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY tool_name
		ORDER BY call_count DESC
		LIMIT ?
	`

	rows, err := r.ch.conn.Query(ctx, query, since, until, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top tools: %w", err)
	}
	defer rows.Close()

	var stats []ToolUsageStats
	for rows.Next() {
		var s ToolUsageStats
		if err := rows.Scan(
			&s.ToolName,
			&s.CallCount,
			&s.ErrorCount,
			&s.SuccessRate,
			&s.AvgDurationMs,
			&s.P50DurationMs,
			&s.P95DurationMs,
			&s.P99DurationMs,
		); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// GetDashboardSummary returns a summary for the dashboard
type DashboardSummary struct {
	TotalRequests     uint64
	TotalTokens       uint64
	TotalToolCalls    uint64
	TotalSessions     uint64
	AvgLatencyMs      float64
	ErrorRate         float64
	RequestsPerMinute float64
}

// GetDashboardSummary returns dashboard summary statistics
func (r *AnalyticsRepository) GetDashboardSummary(ctx context.Context, since, until time.Time) (*DashboardSummary, error) {
	// Get API request stats
	apiQuery := `
		SELECT
			count() as total_requests,
			sum(total_tokens) as total_tokens,
			avg(duration_ms) as avg_latency_ms,
			countIf(is_error = 1) * 100.0 / count() as error_rate
		FROM api_request_analytics
		WHERE timestamp >= ? AND timestamp <= ?
	`

	var summary DashboardSummary
	rows, err := r.ch.conn.Query(ctx, apiQuery, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query API stats: %w", err)
	}

	if rows.Next() {
		if err := rows.Scan(
			&summary.TotalRequests,
			&summary.TotalTokens,
			&summary.AvgLatencyMs,
			&summary.ErrorRate,
		); err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()

	// Get tool call count
	toolQuery := `
		SELECT count() as total_tool_calls
		FROM tool_call_analytics
		WHERE timestamp >= ? AND timestamp <= ?
	`

	rows, err = r.ch.conn.Query(ctx, toolQuery, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool stats: %w", err)
	}

	if rows.Next() {
		if err := rows.Scan(&summary.TotalToolCalls); err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()

	// Get session count
	sessionQuery := `
		SELECT count() as total_sessions
		FROM session_analytics
		WHERE timestamp >= ? AND timestamp <= ?
	`

	rows, err = r.ch.conn.Query(ctx, sessionQuery, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query session stats: %w", err)
	}

	if rows.Next() {
		if err := rows.Scan(&summary.TotalSessions); err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()

	// Calculate requests per minute
	duration := until.Sub(since)
	if duration.Minutes() > 0 {
		summary.RequestsPerMinute = float64(summary.TotalRequests) / duration.Minutes()
	}

	return &summary, nil
}
