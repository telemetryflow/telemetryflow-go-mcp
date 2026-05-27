package services_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsvc "github.com/telemetryflow/telemetryflow-go-mcp/internal/application/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewPromptBuilder(t *testing.T) {
	pb := appsvc.NewPromptBuilder()
	require.NotNil(t, pb)
}

func TestPromptBuilder_BuildSystemPrompt_KnownType(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	prompt := pb.BuildSystemPrompt(vo.ContextMetrics, "")
	assert.Contains(t, prompt, "expert observability analyst")
	assert.Contains(t, prompt, "metrics analysis")
	assert.Contains(t, prompt, "IMPORTANT INSTRUCTIONS")
	assert.Contains(t, prompt, "LIVE DATA")
}

func TestPromptBuilder_BuildSystemPrompt_WithCustomPrompt(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	prompt := pb.BuildSystemPrompt(vo.ContextLogs, "Focus on ERROR severity only")
	assert.Contains(t, prompt, "expert log analyst")
	assert.Contains(t, prompt, "Additional instructions: Focus on ERROR severity only")
}

func TestPromptBuilder_BuildSystemPrompt_UnknownType_FallsBackToDashboard(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	prompt := pb.BuildSystemPrompt(vo.ContextType("nonexistent"), "")
	assert.Contains(t, prompt, "expert dashboard analyst")
}

func TestPromptBuilder_BuildContextPrompt(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	ctx := &vo.TelemetryContext{
		Type:    vo.ContextMetrics,
		Summary: "10 metric series found.",
		Data:    map[string]interface{}{"count": 10},
	}
	ctx.TimeRange = vo.DefaultTimeRange()

	prompt := pb.BuildContextPrompt(ctx)
	assert.Contains(t, prompt, "## Current Context")
	assert.Contains(t, prompt, "**Type:** metrics")
	assert.Contains(t, prompt, "10 metric series found.")
	assert.Contains(t, prompt, "```json")
}

func TestPromptBuilder_BuildContextPrompt_Truncation(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	bigData := make(map[string]interface{})
	bigData["field"] = strings.Repeat("x", 15000)

	ctx := &vo.TelemetryContext{
		Type:    vo.ContextMetrics,
		Summary: "test",
		Data:    bigData,
	}
	ctx.TimeRange = vo.DefaultTimeRange()

	prompt := pb.BuildContextPrompt(ctx)
	assert.Contains(t, prompt, "...")
	assert.LessOrEqual(t, len(prompt), 15000)
}

func TestPromptBuilder_BuildInsightPrompt(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	ctx := &vo.TelemetryContext{
		Type:    vo.ContextMetrics,
		Summary: "High CPU usage detected",
		Data:    map[string]interface{}{"cpu": 95.5},
	}
	ctx.TimeRange = vo.DefaultTimeRange()

	prompt := pb.BuildInsightPrompt(appsvc.InsightRootCause, ctx)
	assert.Contains(t, prompt, "## Current Context")
	assert.Contains(t, prompt, "root cause analysis")
	assert.Contains(t, prompt, "## Analysis Task")
	assert.Contains(t, prompt, "## Analysis Structure")
}

func TestPromptBuilder_BuildInsightPrompt_AllTypes(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	ctx := &vo.TelemetryContext{
		Type:    vo.ContextMetrics,
		Summary: "test",
		Data:    map[string]interface{}{},
	}
	ctx.TimeRange = vo.DefaultTimeRange()

	types := []appsvc.InsightType{
		appsvc.InsightChronology,
		appsvc.InsightPrediction,
		appsvc.InsightRecommendation,
		appsvc.InsightRootCause,
		appsvc.InsightPattern,
	}

	for _, it := range types {
		prompt := pb.BuildInsightPrompt(it, ctx)
		assert.Contains(t, prompt, "## Analysis Task", "missing analysis task for %s", it)
	}
}

func TestPromptBuilder_GetAvailableContextTypes(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	types := pb.GetAvailableContextTypes()
	assert.GreaterOrEqual(t, len(types), 58)

	seen := make(map[string]bool)
	for _, ct := range types {
		assert.False(t, seen[ct.String()], "duplicate type: %s", ct)
		seen[ct.String()] = true
	}
}

func TestPromptBuilder_AllContextTypesHavePrompts(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	for _, ct := range vo.AllContextTypes() {
		prompt := pb.BuildSystemPrompt(ct, "")
		assert.NotEmpty(t, prompt, "empty prompt for context type: %s", ct)
		assert.Contains(t, prompt, "IMPORTANT INSTRUCTIONS", "missing instructions for: %s", ct)
	}
}

func TestPromptBuilder_SpecializedPrompts(t *testing.T) {
	pb := appsvc.NewPromptBuilder()

	tests := []struct {
		contextType vo.ContextType
		contains    string
	}{
		{vo.ContextMetrics, "observability analyst"},
		{vo.ContextLogs, "log analyst"},
		{vo.ContextTraces, "distributed tracing analyst"},
		{vo.ContextAlerts, "incident analyst"},
		{vo.ContextKubernetesOverview, "Kubernetes administrator"},
		{vo.ContextDBMonitoringMySQL, "MySQL"},
		{vo.ContextDBMonitoringClickHouse, "ClickHouse"},
		{vo.ContextAnomalyDetection, "anomaly detection"},
		{vo.ContextCostOptimization, "cost optimization"},
		{vo.ContextServiceMap, "service dependency"},
	}

	for _, tt := range tests {
		prompt := pb.BuildSystemPrompt(tt.contextType, "")
		assert.Contains(t, prompt, tt.contains, "prompt for %s missing '%s'", tt.contextType, tt.contains)
	}
}
