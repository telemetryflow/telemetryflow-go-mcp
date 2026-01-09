// Package telemetry provides OpenTelemetry integration utilities
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds telemetry configuration
type Config struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	SampleRate     float64
	ExportTimeout  time.Duration
}

// DefaultConfig returns default telemetry configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:        false,
		ServiceName:    "tfo-mcp",
		ServiceVersion: "1.1.2",
		Environment:    "development",
		Endpoint:       "localhost:4317",
		SampleRate:     1.0,
		ExportTimeout:  30 * time.Second,
	}
}

// Provider manages telemetry resources
type Provider struct {
	config         *Config
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
}

// NewProvider creates a new telemetry provider
func NewProvider(ctx context.Context, config *Config) (*Provider, error) {
	if !config.Enabled {
		return &Provider{
			config: config,
			tracer: otel.Tracer(config.ServiceName),
		}, nil
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(config.ExportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	if config.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SampleRate)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		config:         config,
		tracerProvider: tp,
		tracer:         tp.Tracer(config.ServiceName),
	}, nil
}

// Tracer returns the tracer
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Shutdown shuts down the telemetry provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tracerProvider != nil {
		return p.tracerProvider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span
func (p *Provider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return p.tracer.Start(ctx, name, opts...)
}

// Common attribute keys
const (
	AttrSessionID      = "session.id"
	AttrConversationID = "conversation.id"
	AttrToolName       = "tool.name"
	AttrResourceURI    = "resource.uri"
	AttrPromptName     = "prompt.name"
	AttrMCPMethod      = "mcp.method"
	AttrClaudeModel    = "claude.model"
	AttrTokensInput    = "claude.tokens.input"
	AttrTokensOutput   = "claude.tokens.output"
	AttrErrorCode      = "error.code"
	AttrErrorMessage   = "error.message"
)

// Span helper functions

// SetSessionAttributes sets session-related attributes on a span
func SetSessionAttributes(span trace.Span, sessionID string) {
	span.SetAttributes(attribute.String(AttrSessionID, sessionID))
}

// SetConversationAttributes sets conversation-related attributes on a span
func SetConversationAttributes(span trace.Span, conversationID string) {
	span.SetAttributes(attribute.String(AttrConversationID, conversationID))
}

// SetToolAttributes sets tool-related attributes on a span
func SetToolAttributes(span trace.Span, toolName string) {
	span.SetAttributes(attribute.String(AttrToolName, toolName))
}

// SetResourceAttributes sets resource-related attributes on a span
func SetResourceAttributes(span trace.Span, uri string) {
	span.SetAttributes(attribute.String(AttrResourceURI, uri))
}

// SetPromptAttributes sets prompt-related attributes on a span
func SetPromptAttributes(span trace.Span, promptName string) {
	span.SetAttributes(attribute.String(AttrPromptName, promptName))
}

// SetMCPMethodAttribute sets MCP method attribute on a span
func SetMCPMethodAttribute(span trace.Span, method string) {
	span.SetAttributes(attribute.String(AttrMCPMethod, method))
}

// SetClaudeAttributes sets Claude API related attributes on a span
func SetClaudeAttributes(span trace.Span, model string, inputTokens, outputTokens int) {
	span.SetAttributes(
		attribute.String(AttrClaudeModel, model),
		attribute.Int(AttrTokensInput, inputTokens),
		attribute.Int(AttrTokensOutput, outputTokens),
	)
}

// RecordError records an error on a span
func RecordError(span trace.Span, err error, code int) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	span.SetAttributes(
		attribute.Int(AttrErrorCode, code),
		attribute.String(AttrErrorMessage, err.Error()),
	)
}

// EndSpanWithError ends a span and records an error if present
func EndSpanWithError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// SpanFromContext extracts a span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan returns a context with the given span
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}
