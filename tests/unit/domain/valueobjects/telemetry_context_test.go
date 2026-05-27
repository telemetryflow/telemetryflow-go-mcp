package valueobjects_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestContextType_IsValid(t *testing.T) {
	validTypes := []vo.ContextType{
		vo.ContextMetrics,
		vo.ContextLogs,
		vo.ContextTraces,
		vo.ContextAlerts,
		vo.ContextKubernetesOverview,
		vo.ContextAgents,
		vo.ContextUptime,
		vo.ContextServiceMap,
		vo.ContextIAM,
		vo.ContextTenancy,
		vo.ContextAudit,
		vo.ContextAnomalyDetection,
		vo.ContextDBMonitoringMySQL,
		vo.ContextDBMonitoringQAN,
	}
	for _, ct := range validTypes {
		assert.True(t, ct.IsValid(), "expected %s to be valid", ct)
	}

	invalid := vo.ContextType("nonexistent-type")
	assert.False(t, invalid.IsValid())
}

func TestContextType_String(t *testing.T) {
	assert.Equal(t, "metrics", vo.ContextMetrics.String())
	assert.Equal(t, "db-monitoring-qan", vo.ContextDBMonitoringQAN.String())
}

func TestAllContextTypes(t *testing.T) {
	types := vo.AllContextTypes()
	assert.GreaterOrEqual(t, len(types), 58)

	seen := make(map[string]bool)
	for _, ct := range types {
		s := ct.String()
		assert.False(t, seen[s], "duplicate context type: %s", s)
		assert.True(t, ct.IsValid(), "AllContextTypes contains invalid type: %s", s)
		seen[s] = true
	}
}

func TestDefaultTimeRange(t *testing.T) {
	tr := vo.DefaultTimeRange()
	now := time.Now().UTC()

	assert.WithinDuration(t, now.Add(-time.Hour), tr.From, 5*time.Second)
	assert.WithinDuration(t, now, tr.To, 5*time.Second)
	assert.True(t, tr.From.Before(tr.To))
}

func TestTelemetryContext(t *testing.T) {
	tc := &vo.TelemetryContext{
		Type: vo.ContextMetrics,
		TimeRange: vo.TimeRange{
			From: time.Now().Add(-time.Hour),
			To:   time.Now(),
		},
		Summary: "10 metric series found.",
		Data:    map[string]interface{}{"count": 10},
	}

	assert.Equal(t, vo.ContextMetrics, tc.Type)
	assert.Equal(t, "10 metric series found.", tc.Summary)
	assert.NotNil(t, tc.Data)
}

func TestCollectContextOptions(t *testing.T) {
	opts := vo.CollectContextOptions{
		OrganizationID: "org-123",
		UserID:         "user-456",
		ContextType:    vo.ContextLogs,
		ContextID:      "ctx-789",
		TimeRange:      &vo.TimeRange{From: time.Now().Add(-2 * time.Hour), To: time.Now()},
		MaxItems:       50,
	}

	assert.Equal(t, "org-123", opts.OrganizationID)
	assert.Equal(t, "user-456", opts.UserID)
	assert.Equal(t, vo.ContextLogs, opts.ContextType)
	assert.Equal(t, 50, opts.MaxItems)
	assert.NotNil(t, opts.TimeRange)
}

func TestCollectContextOptions_DefaultTimeRange(t *testing.T) {
	opts := vo.CollectContextOptions{
		OrganizationID: "org-123",
		ContextType:    vo.ContextMetrics,
	}

	assert.Nil(t, opts.TimeRange)
	assert.Equal(t, 0, opts.MaxItems)
}

func TestContextType_AllCategories(t *testing.T) {
	coreTypes := []vo.ContextType{
		vo.ContextMetrics, vo.ContextLogs, vo.ContextTraces,
		vo.ContextExemplars, vo.ContextCorrelations,
	}
	for _, ct := range coreTypes {
		assert.True(t, ct.IsValid())
	}

	k8sTypes := []vo.ContextType{
		vo.ContextKubernetesOverview, vo.ContextKubernetesClusters,
		vo.ContextKubernetesNodes, vo.ContextKubernetesPods,
	}
	for _, ct := range k8sTypes {
		assert.True(t, ct.IsValid())
	}

	dbTypes := []vo.ContextType{
		vo.ContextDBMonitoringInventory, vo.ContextDBMonitoringMySQL,
		vo.ContextDBMonitoringPostgreSQL, vo.ContextDBMonitoringQAN,
	}
	for _, ct := range dbTypes {
		assert.True(t, ct.IsValid())
	}
}

func TestTimeRange_Struct(t *testing.T) {
	now := time.Now()
	tr := vo.TimeRange{From: now.Add(-time.Hour), To: now}

	require.False(t, tr.From.IsZero())
	require.False(t, tr.To.IsZero())
	require.True(t, tr.From.Before(tr.To))
}
