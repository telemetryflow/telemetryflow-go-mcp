// Package telemetry provides OpenTelemetry integration utilities
package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds application metrics
type Metrics struct {
	meter metric.Meter

	// Request metrics
	RequestsTotal    metric.Int64Counter
	RequestDuration  metric.Float64Histogram
	RequestsInFlight metric.Int64UpDownCounter

	// Tool metrics
	ToolCallsTotal   metric.Int64Counter
	ToolCallDuration metric.Float64Histogram
	ToolErrors       metric.Int64Counter

	// Claude API metrics
	ClaudeRequestsTotal metric.Int64Counter
	ClaudeTokensInput   metric.Int64Counter
	ClaudeTokensOutput  metric.Int64Counter
	ClaudeLatency       metric.Float64Histogram
	ClaudeErrors        metric.Int64Counter

	// Session metrics
	ActiveSessions  metric.Int64UpDownCounter
	SessionDuration metric.Float64Histogram

	// Resource metrics
	ResourceReadsTotal  metric.Int64Counter
	ResourceCacheHits   metric.Int64Counter
	ResourceCacheMisses metric.Int64Counter
}

// NewMetrics creates a new Metrics instance
func NewMetrics(serviceName string) (*Metrics, error) {
	meter := otel.Meter(serviceName)
	m := &Metrics{meter: meter}

	var err error

	// Request metrics
	m.RequestsTotal, err = meter.Int64Counter(
		"mcp.requests.total",
		metric.WithDescription("Total number of MCP requests"),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		return nil, err
	}

	m.RequestDuration, err = meter.Float64Histogram(
		"mcp.request.duration",
		metric.WithDescription("Duration of MCP requests"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.RequestsInFlight, err = meter.Int64UpDownCounter(
		"mcp.requests.in_flight",
		metric.WithDescription("Number of requests currently in flight"),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		return nil, err
	}

	// Tool metrics
	m.ToolCallsTotal, err = meter.Int64Counter(
		"mcp.tool.calls.total",
		metric.WithDescription("Total number of tool calls"),
		metric.WithUnit("{calls}"),
	)
	if err != nil {
		return nil, err
	}

	m.ToolCallDuration, err = meter.Float64Histogram(
		"mcp.tool.call.duration",
		metric.WithDescription("Duration of tool calls"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.ToolErrors, err = meter.Int64Counter(
		"mcp.tool.errors.total",
		metric.WithDescription("Total number of tool errors"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		return nil, err
	}

	// Claude API metrics
	m.ClaudeRequestsTotal, err = meter.Int64Counter(
		"claude.requests.total",
		metric.WithDescription("Total number of Claude API requests"),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		return nil, err
	}

	m.ClaudeTokensInput, err = meter.Int64Counter(
		"claude.tokens.input",
		metric.WithDescription("Total input tokens consumed"),
		metric.WithUnit("{tokens}"),
	)
	if err != nil {
		return nil, err
	}

	m.ClaudeTokensOutput, err = meter.Int64Counter(
		"claude.tokens.output",
		metric.WithDescription("Total output tokens generated"),
		metric.WithUnit("{tokens}"),
	)
	if err != nil {
		return nil, err
	}

	m.ClaudeLatency, err = meter.Float64Histogram(
		"claude.latency",
		metric.WithDescription("Claude API latency"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.ClaudeErrors, err = meter.Int64Counter(
		"claude.errors.total",
		metric.WithDescription("Total number of Claude API errors"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		return nil, err
	}

	// Session metrics
	m.ActiveSessions, err = meter.Int64UpDownCounter(
		"mcp.sessions.active",
		metric.WithDescription("Number of active sessions"),
		metric.WithUnit("{sessions}"),
	)
	if err != nil {
		return nil, err
	}

	m.SessionDuration, err = meter.Float64Histogram(
		"mcp.session.duration",
		metric.WithDescription("Duration of sessions"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Resource metrics
	m.ResourceReadsTotal, err = meter.Int64Counter(
		"mcp.resource.reads.total",
		metric.WithDescription("Total number of resource reads"),
		metric.WithUnit("{reads}"),
	)
	if err != nil {
		return nil, err
	}

	m.ResourceCacheHits, err = meter.Int64Counter(
		"mcp.resource.cache.hits",
		metric.WithDescription("Number of resource cache hits"),
		metric.WithUnit("{hits}"),
	)
	if err != nil {
		return nil, err
	}

	m.ResourceCacheMisses, err = meter.Int64Counter(
		"mcp.resource.cache.misses",
		metric.WithDescription("Number of resource cache misses"),
		metric.WithUnit("{misses}"),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// RecordRequest records a request metric
func (m *Metrics) RecordRequest(ctx context.Context, method string, duration time.Duration, err error) {
	attrs := metric.WithAttributes()
	m.RequestsTotal.Add(ctx, 1, attrs)
	m.RequestDuration.Record(ctx, duration.Seconds(), attrs)
}

// RecordToolCall records a tool call metric
func (m *Metrics) RecordToolCall(ctx context.Context, toolName string, duration time.Duration, err error) {
	attrs := metric.WithAttributes()
	m.ToolCallsTotal.Add(ctx, 1, attrs)
	m.ToolCallDuration.Record(ctx, duration.Seconds(), attrs)
	if err != nil {
		m.ToolErrors.Add(ctx, 1, attrs)
	}
}

// RecordClaudeRequest records a Claude API request metric
func (m *Metrics) RecordClaudeRequest(ctx context.Context, model string, inputTokens, outputTokens int, duration time.Duration, err error) {
	attrs := metric.WithAttributes()
	m.ClaudeRequestsTotal.Add(ctx, 1, attrs)
	m.ClaudeTokensInput.Add(ctx, int64(inputTokens), attrs)
	m.ClaudeTokensOutput.Add(ctx, int64(outputTokens), attrs)
	m.ClaudeLatency.Record(ctx, duration.Seconds(), attrs)
	if err != nil {
		m.ClaudeErrors.Add(ctx, 1, attrs)
	}
}

// IncrementActiveSessions increments active sessions counter
func (m *Metrics) IncrementActiveSessions(ctx context.Context) {
	m.ActiveSessions.Add(ctx, 1)
}

// DecrementActiveSessions decrements active sessions counter
func (m *Metrics) DecrementActiveSessions(ctx context.Context) {
	m.ActiveSessions.Add(ctx, -1)
}

// RecordSessionDuration records session duration
func (m *Metrics) RecordSessionDuration(ctx context.Context, duration time.Duration) {
	m.SessionDuration.Record(ctx, duration.Seconds())
}

// RecordResourceRead records a resource read
func (m *Metrics) RecordResourceRead(ctx context.Context, uri string, cacheHit bool) {
	attrs := metric.WithAttributes()
	m.ResourceReadsTotal.Add(ctx, 1, attrs)
	if cacheHit {
		m.ResourceCacheHits.Add(ctx, 1, attrs)
	} else {
		m.ResourceCacheMisses.Add(ctx, 1, attrs)
	}
}

// IncrementRequestsInFlight increments in-flight requests
func (m *Metrics) IncrementRequestsInFlight(ctx context.Context) {
	m.RequestsInFlight.Add(ctx, 1)
}

// DecrementRequestsInFlight decrements in-flight requests
func (m *Metrics) DecrementRequestsInFlight(ctx context.Context) {
	m.RequestsInFlight.Add(ctx, -1)
}
