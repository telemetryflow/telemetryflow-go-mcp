// Package telemetry provides OpenTelemetry integration utilities
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracing functionality for MCP operations
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new Tracer instance
func NewTracer(provider *Provider) *Tracer {
	return &Tracer{
		tracer: provider.Tracer(),
	}
}

// TraceOption represents options for creating spans
type TraceOption func(*traceOptions)

type traceOptions struct {
	sessionID      string
	conversationID string
	toolName       string
	resourceURI    string
	promptName     string
	mcpMethod      string
	model          string
	attributes     []attribute.KeyValue
}

// WithSessionID sets the session ID attribute
func WithSessionID(id string) TraceOption {
	return func(o *traceOptions) {
		o.sessionID = id
	}
}

// WithConversationID sets the conversation ID attribute
func WithConversationID(id string) TraceOption {
	return func(o *traceOptions) {
		o.conversationID = id
	}
}

// WithToolName sets the tool name attribute
func WithToolName(name string) TraceOption {
	return func(o *traceOptions) {
		o.toolName = name
	}
}

// WithResourceURI sets the resource URI attribute
func WithResourceURI(uri string) TraceOption {
	return func(o *traceOptions) {
		o.resourceURI = uri
	}
}

// WithPromptName sets the prompt name attribute
func WithPromptName(name string) TraceOption {
	return func(o *traceOptions) {
		o.promptName = name
	}
}

// WithMCPMethod sets the MCP method attribute
func WithMCPMethod(method string) TraceOption {
	return func(o *traceOptions) {
		o.mcpMethod = method
	}
}

// WithModel sets the Claude model attribute
func WithModel(model string) TraceOption {
	return func(o *traceOptions) {
		o.model = model
	}
}

// WithAttribute adds a custom attribute
func WithAttribute(key string, value interface{}) TraceOption {
	return func(o *traceOptions) {
		switch v := value.(type) {
		case string:
			o.attributes = append(o.attributes, attribute.String(key, v))
		case int:
			o.attributes = append(o.attributes, attribute.Int(key, v))
		case int64:
			o.attributes = append(o.attributes, attribute.Int64(key, v))
		case float64:
			o.attributes = append(o.attributes, attribute.Float64(key, v))
		case bool:
			o.attributes = append(o.attributes, attribute.Bool(key, v))
		case []string:
			o.attributes = append(o.attributes, attribute.StringSlice(key, v))
		default:
			o.attributes = append(o.attributes, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
}

// StartSpan starts a new span with the given name and options
func (t *Tracer) StartSpan(ctx context.Context, spanName string, opts ...TraceOption) (context.Context, trace.Span) {
	options := &traceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	ctx, span := t.tracer.Start(ctx, spanName)

	// Apply attributes
	attrs := make([]attribute.KeyValue, 0)

	if options.sessionID != "" {
		attrs = append(attrs, attribute.String(AttrSessionID, options.sessionID))
	}
	if options.conversationID != "" {
		attrs = append(attrs, attribute.String(AttrConversationID, options.conversationID))
	}
	if options.toolName != "" {
		attrs = append(attrs, attribute.String(AttrToolName, options.toolName))
	}
	if options.resourceURI != "" {
		attrs = append(attrs, attribute.String(AttrResourceURI, options.resourceURI))
	}
	if options.promptName != "" {
		attrs = append(attrs, attribute.String(AttrPromptName, options.promptName))
	}
	if options.mcpMethod != "" {
		attrs = append(attrs, attribute.String(AttrMCPMethod, options.mcpMethod))
	}
	if options.model != "" {
		attrs = append(attrs, attribute.String(AttrClaudeModel, options.model))
	}

	// Add custom attributes
	attrs = append(attrs, options.attributes...)

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	return ctx, span
}

// MCP Operation Spans

// StartMCPRequestSpan starts a span for an MCP request
func (t *Tracer) StartMCPRequestSpan(ctx context.Context, method string, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.request",
		WithMCPMethod(method),
		WithSessionID(sessionID),
	)
}

// StartInitializeSpan starts a span for session initialization
func (t *Tracer) StartInitializeSpan(ctx context.Context) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.initialize",
		WithMCPMethod("initialize"),
	)
}

// StartToolCallSpan starts a span for tool execution
func (t *Tracer) StartToolCallSpan(ctx context.Context, toolName, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.tool.call",
		WithMCPMethod("tools/call"),
		WithToolName(toolName),
		WithSessionID(sessionID),
	)
}

// StartToolListSpan starts a span for tool listing
func (t *Tracer) StartToolListSpan(ctx context.Context, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.tool.list",
		WithMCPMethod("tools/list"),
		WithSessionID(sessionID),
	)
}

// StartResourceReadSpan starts a span for resource reading
func (t *Tracer) StartResourceReadSpan(ctx context.Context, uri, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.resource.read",
		WithMCPMethod("resources/read"),
		WithResourceURI(uri),
		WithSessionID(sessionID),
	)
}

// StartResourceListSpan starts a span for resource listing
func (t *Tracer) StartResourceListSpan(ctx context.Context, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.resource.list",
		WithMCPMethod("resources/list"),
		WithSessionID(sessionID),
	)
}

// StartPromptGetSpan starts a span for prompt retrieval
func (t *Tracer) StartPromptGetSpan(ctx context.Context, promptName, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.prompt.get",
		WithMCPMethod("prompts/get"),
		WithPromptName(promptName),
		WithSessionID(sessionID),
	)
}

// StartPromptListSpan starts a span for prompt listing
func (t *Tracer) StartPromptListSpan(ctx context.Context, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "mcp.prompt.list",
		WithMCPMethod("prompts/list"),
		WithSessionID(sessionID),
	)
}

// Claude API Spans

// StartClaudeRequestSpan starts a span for Claude API requests
func (t *Tracer) StartClaudeRequestSpan(ctx context.Context, model, sessionID, conversationID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "claude.request",
		WithModel(model),
		WithSessionID(sessionID),
		WithConversationID(conversationID),
	)
}

// StartClaudeStreamSpan starts a span for Claude streaming requests
func (t *Tracer) StartClaudeStreamSpan(ctx context.Context, model, sessionID, conversationID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "claude.stream",
		WithModel(model),
		WithSessionID(sessionID),
		WithConversationID(conversationID),
	)
}

// StartTokenCountSpan starts a span for token counting
func (t *Tracer) StartTokenCountSpan(ctx context.Context, model string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "claude.token_count",
		WithModel(model),
	)
}

// Session Lifecycle Spans

// StartSessionCreateSpan starts a span for session creation
func (t *Tracer) StartSessionCreateSpan(ctx context.Context) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "session.create")
}

// StartSessionCloseSpan starts a span for session closure
func (t *Tracer) StartSessionCloseSpan(ctx context.Context, sessionID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "session.close",
		WithSessionID(sessionID),
	)
}

// Conversation Spans

// StartConversationCreateSpan starts a span for conversation creation
func (t *Tracer) StartConversationCreateSpan(ctx context.Context, sessionID, model string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "conversation.create",
		WithSessionID(sessionID),
		WithModel(model),
	)
}

// StartConversationCloseSpan starts a span for conversation closure
func (t *Tracer) StartConversationCloseSpan(ctx context.Context, sessionID, conversationID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "conversation.close",
		WithSessionID(sessionID),
		WithConversationID(conversationID),
	)
}

// StartMessageAddSpan starts a span for adding a message
func (t *Tracer) StartMessageAddSpan(ctx context.Context, sessionID, conversationID, role string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "conversation.message.add",
		WithSessionID(sessionID),
		WithConversationID(conversationID),
		WithAttribute("message.role", role),
	)
}

// Span completion helpers

// EndSpanOK ends a span with OK status
func EndSpanOK(span trace.Span) {
	span.SetStatus(codes.Ok, "")
	span.End()
}

// EndSpanError ends a span with error status
func EndSpanError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// EndSpanErrorWithCode ends a span with error status and code
func EndSpanErrorWithCode(span trace.Span, err error, code int) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Int(AttrErrorCode, code))
	}
	span.End()
}

// AddSpanEvent adds an event to the span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// AddSpanTokenInfo adds token information to a span
func AddSpanTokenInfo(span trace.Span, inputTokens, outputTokens int) {
	span.SetAttributes(
		attribute.Int(AttrTokensInput, inputTokens),
		attribute.Int(AttrTokensOutput, outputTokens),
		attribute.Int("claude.tokens.total", inputTokens+outputTokens),
	)
}

// Common span events
const (
	EventToolStart    = "tool.start"
	EventToolComplete = "tool.complete"
	EventToolError    = "tool.error"
	EventMessageSent  = "message.sent"
	EventMessageRecv  = "message.received"
	EventStreamStart  = "stream.start"
	EventStreamChunk  = "stream.chunk"
	EventStreamEnd    = "stream.end"
	EventCacheHit     = "cache.hit"
	EventCacheMiss    = "cache.miss"
	EventRetryAttempt = "retry.attempt"
	EventRateLimited  = "rate.limited"
)

// TracedOperation wraps an operation with tracing
func TracedOperation[T any](ctx context.Context, tracer *Tracer, spanName string, opts []TraceOption, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := tracer.StartSpan(ctx, spanName, opts...)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return result, err
}

// TracedVoidOperation wraps a void operation with tracing
func TracedVoidOperation(ctx context.Context, tracer *Tracer, spanName string, opts []TraceOption, fn func(context.Context) error) error {
	ctx, span := tracer.StartSpan(ctx, spanName, opts...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}
