package services_test

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	appsvc "github.com/telemetryflow/telemetryflow-go-mcp/internal/application/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type mockDBProvider struct {
	gormDB *gorm.DB
	chConn driver.Conn
	chDB   string
	hasPG  bool
	hasCH  bool
}

func (m *mockDBProvider) GormDB() *gorm.DB            { return m.gormDB }
func (m *mockDBProvider) ClickHouseConn() driver.Conn { return m.chConn }
func (m *mockDBProvider) ClickHouseDB() string        { return m.chDB }
func (m *mockDBProvider) HasClickHouse() bool         { return m.hasCH }
func (m *mockDBProvider) HasPostgres() bool           { return m.hasPG }

type mockCHColType struct{ name string }

func (t mockCHColType) Name() string             { return t.name }
func (t mockCHColType) Nullable() bool           { return false }
func (t mockCHColType) ScanType() reflect.Type   { return nil }
func (t mockCHColType) DatabaseTypeName() string { return "String" }

type mockCHRows struct {
	colTypes []driver.ColumnType
	data     [][]interface{}
	idx      int
}

func newCHRows(cols []string, data [][]interface{}) *mockCHRows {
	ct := make([]driver.ColumnType, len(cols))
	for i, c := range cols {
		ct[i] = mockCHColType{name: c}
	}
	return &mockCHRows{colTypes: ct, data: data, idx: -1}
}

func (r *mockCHRows) Next() bool {
	if r.data == nil {
		return false
	}
	r.idx++
	return r.idx < len(r.data)
}

func (r *mockCHRows) Scan(dest ...any) error {
	if r.idx < 0 || r.idx >= len(r.data) {
		return fmt.Errorf("out of bounds")
	}
	row := r.data[r.idx]
	for i := range dest {
		if i >= len(row) {
			break
		}
		if ptr, ok := dest[i].(*interface{}); ok {
			*ptr = row[i]
		}
	}
	return nil
}

func (r *mockCHRows) ColumnTypes() []driver.ColumnType { return r.colTypes }
func (r *mockCHRows) Columns() []string                { return nil }
func (r *mockCHRows) Close() error                     { return nil }
func (r *mockCHRows) Err() error                       { return nil }
func (r *mockCHRows) Totals(dest ...any) error         { return nil }
func (r *mockCHRows) ScanStruct(dest any) error        { return nil }

type mockCHConn struct {
	sets     []*mockCHRows
	callIdx  int
	queryErr error
}

func (c *mockCHConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	if c.queryErr != nil {
		return nil, c.queryErr
	}
	if c.callIdx >= len(c.sets) {
		return newCHRows(nil, nil), nil
	}
	rs := c.sets[c.callIdx]
	c.callIdx++
	return rs, nil
}
func (c *mockCHConn) Contributors() []string                        { return nil }
func (c *mockCHConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (c *mockCHConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (c *mockCHConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row { return nil }
func (c *mockCHConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}
func (c *mockCHConn) Exec(ctx context.Context, query string, args ...any) error { return nil }
func (c *mockCHConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (c *mockCHConn) Ping(context.Context) error { return nil }
func (c *mockCHConn) Stats() driver.Stats        { return driver.Stats{} }
func (c *mockCHConn) Close() error               { return nil }

type mockPgRows struct {
	cols []string
	data [][]interface{}
	idx  int
}

func (r *mockPgRows) Columns() []string { return r.cols }
func (r *mockPgRows) Close() error      { return nil }
func (r *mockPgRows) Next(dest []sqldriver.Value) error {
	if r.data == nil || r.idx >= len(r.data) {
		return io.EOF
	}
	for i, v := range r.data[r.idx] {
		if i < len(dest) {
			dest[i] = v
		}
	}
	r.idx++
	return nil
}

type mockPgConn struct {
	sets     []*mockPgRows
	callIdx  int
	queryErr error
}

func (c *mockPgConn) Prepare(query string) (sqldriver.Stmt, error) { return nil, sqldriver.ErrSkip }
func (c *mockPgConn) Close() error                                 { return nil }
func (c *mockPgConn) Begin() (sqldriver.Tx, error)                 { return nil, sqldriver.ErrSkip }
func (c *mockPgConn) ResetSession(ctx context.Context) error       { return nil }
func (c *mockPgConn) ExecContext(ctx context.Context, query string, args []sqldriver.NamedValue) (sqldriver.Result, error) {
	return nil, nil
}
func (c *mockPgConn) CheckNamedValue(nv *sqldriver.NamedValue) error { return nil }
func (c *mockPgConn) QueryContext(ctx context.Context, query string, args []sqldriver.NamedValue) (sqldriver.Rows, error) {
	if c.queryErr != nil {
		return nil, c.queryErr
	}
	if c.callIdx >= len(c.sets) {
		return &mockPgRows{cols: []string{}}, nil
	}
	rs := c.sets[c.callIdx]
	c.callIdx++
	return rs, nil
}

type mockPgConnector struct {
	conn *mockPgConn
}

func (c *mockPgConnector) Connect(ctx context.Context) (sqldriver.Conn, error) { return c.conn, nil }
func (c *mockPgConnector) Driver() sqldriver.Driver                            { return nil }

func newMockGormDB(sets ...*mockPgRows) *gorm.DB {
	conn := &mockPgConn{sets: sets}
	sqlDB := sql.OpenDB(&mockPgConnector{conn: conn})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to create mock gorm db: %v", err))
	}
	return db
}

func newTestProvider(chSets []*mockCHRows, pgSets []*mockPgRows) *mockDBProvider {
	chConn := &mockCHConn{sets: chSets}
	var gormDB *gorm.DB
	hasPG := len(pgSets) > 0
	if hasPG {
		gormDB = newMockGormDB(pgSets...)
	}
	return &mockDBProvider{
		gormDB: gormDB,
		chConn: chConn,
		chDB:   "testdb",
		hasPG:  hasPG,
		hasCH:  true,
	}
}

func TestNewContextCollector(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()

	collector := appsvc.NewContextCollector(provider, logger)
	require.NotNil(t, collector)
}

func TestContextCollector_CollectContext_NoDB(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	opts := vo.CollectContextOptions{
		OrganizationID: "org-123",
		ContextType:    vo.ContextMetrics,
	}

	tc, err := collector.CollectContext(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Contains(t, tc.Summary, "unavailable")
	assert.Equal(t, vo.ContextMetrics, tc.Type)
}

func TestContextCollector_CollectContext_DefaultTimeRange(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	opts := vo.CollectContextOptions{
		OrganizationID: "org-123",
		ContextType:    vo.ContextLogs,
	}

	tc, err := collector.CollectContext(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, tc)

	now := time.Now().UTC()
	assert.WithinDuration(t, now, tc.TimeRange.To, 5*time.Second)
	assert.WithinDuration(t, now.Add(-time.Hour), tc.TimeRange.From, 5*time.Second)
}

func TestContextCollector_CollectContext_CustomTimeRange(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	from := time.Now().UTC().Add(-2 * time.Hour)
	to := time.Now().UTC()
	tr := vo.TimeRange{From: from, To: to}
	opts := vo.CollectContextOptions{
		OrganizationID: "org-123",
		ContextType:    vo.ContextTraces,
		TimeRange:      &tr,
	}

	tc, err := collector.CollectContext(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.WithinDuration(t, from, tc.TimeRange.From, time.Second)
	assert.WithinDuration(t, to, tc.TimeRange.To, time.Second)
}

func TestContextCollector_Dispatch_AllContextTypes_ReturnsResult(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	for _, ct := range vo.AllContextTypes() {
		opts := vo.CollectContextOptions{
			OrganizationID: "org-123",
			ContextType:    ct,
		}
		tc, err := collector.CollectContext(context.Background(), opts)
		require.NoError(t, err, "failed for context type: %s", ct)
		require.NotNil(t, tc, "nil result for context type: %s", ct)
		assert.NotEmpty(t, tc.Summary, "empty summary for: %s", ct)
	}
}

func TestContextCollector_AccountContext_RequiresUserID(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	accountTypes := []vo.ContextType{
		vo.ContextAccountProfile,
		vo.ContextAccountSecurity,
		vo.ContextAccountSessions,
	}

	for _, ct := range accountTypes {
		opts := vo.CollectContextOptions{
			OrganizationID: "org-123",
			ContextType:    ct,
		}
		tc, err := collector.CollectContext(context.Background(), opts)
		require.NoError(t, err)
		assert.Contains(t, tc.Summary, "user ID required", "expected user ID required for %s", ct)
	}
}

func TestContextCollector_InfraContext_MetricMapping(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	infraTypes := []vo.ContextType{
		vo.ContextInfraCPU,
		vo.ContextInfraMemory,
		vo.ContextInfraStorage,
		vo.ContextInfraNetwork,
	}

	for _, ct := range infraTypes {
		opts := vo.CollectContextOptions{
			OrganizationID: "org-123",
			ContextType:    ct,
		}
		tc, err := collector.CollectContext(context.Background(), opts)
		require.NoError(t, err, "failed for infra type: %s", ct)
		require.NotNil(t, tc)
		assert.Contains(t, tc.Summary, "unavailable")
	}
}

func TestDefaultDBProvider_NoConnections(t *testing.T) {
	provider := appsvc.NewDefaultDBProvider(nil, nil, "")
	require.NotNil(t, provider)
	assert.False(t, provider.HasPostgres())
	assert.False(t, provider.HasClickHouse())
	assert.Nil(t, provider.GormDB())
	assert.Nil(t, provider.ClickHouseConn())
	assert.Equal(t, "", provider.ClickHouseDB())
}

func TestDefaultDBProvider_WithPG(t *testing.T) {
	db := &gorm.DB{}
	provider := appsvc.NewDefaultDBProvider(db, nil, "testdb")
	assert.True(t, provider.HasPostgres())
	assert.False(t, provider.HasClickHouse())
	assert.NotNil(t, provider.GormDB())
	assert.Equal(t, "testdb", provider.ClickHouseDB())
}

func TestContextCollector_CanceledContext(t *testing.T) {
	provider := &mockDBProvider{}
	logger := zerolog.Nop()
	collector := appsvc.NewContextCollector(provider, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := vo.CollectContextOptions{
		OrganizationID: "org-123",
		ContextType:    vo.ContextMetrics,
	}

	tc, err := collector.CollectContext(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Contains(t, tc.Summary, "timed out")
}

func TestContextCollector_CHQuerySuccess(t *testing.T) {
	cols := []string{"service_name", "metric_name", "metric_type", "avg_val", "max_val", "min_val", "sample_count"}
	data := [][]interface{}{
		{"api-server", "http_requests", "counter", 1520.5, 3200.0, 50.0, 1440.0},
		{"worker", "cpu_usage", "gauge", 45.0, 95.0, 5.0, 2880.0},
	}
	provider := newTestProvider([]*mockCHRows{newCHRows(cols, data)}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextMetrics,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.Type != vo.ContextMetrics {
		t.Errorf("expected metrics type, got %s", tc.Type)
	}
	d := tc.Data.(map[string]interface{})
	if d["metrics"] == nil {
		t.Error("expected metrics in data")
	}
	if d["highlights"] == nil {
		t.Error("expected highlights in data")
	}
}

func TestContextCollector_CHQueryError(t *testing.T) {
	provider := newTestProvider([]*mockCHRows{}, nil)
	provider.chConn = &mockCHConn{queryErr: fmt.Errorf("connection refused")}
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextMetrics,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(tc.Summary, "unavailable") {
		t.Errorf("expected unavailable summary, got %s", tc.Summary)
	}
}

func TestContextCollector_PGQuerySuccess(t *testing.T) {
	cols := []string{"id", "name", "data_type", "retention_days", "archive_enabled", "is_default", "is_active"}
	data := [][]interface{}{
		{"p1", "Default Metrics", "metrics", 30.0, true, true, true},
		{"p2", "Logs Hot", "logs", 7.0, false, false, true},
	}
	provider := newTestProvider(nil, []*mockPgRows{{cols: cols, data: data}})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextRetention,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.Type != vo.ContextRetention {
		t.Errorf("expected retention type, got %s", tc.Type)
	}
}

func TestContextCollector_LogsContext_WithData(t *testing.T) {
	sevCols := []string{"service_name", "severity_text", "total"}
	sevData := [][]interface{}{
		{"api-gw", "INFO", float64(5000)},
		{"api-gw", "ERROR", float64(120)},
		{"worker", "FATAL", float64(3)},
	}
	errCols := []string{"severity_text", "service_name", "body", "timestamp"}
	errData := [][]interface{}{
		{"ERROR", "api-gw", "connection refused", "2024-01-01T00:00:00Z"},
		{"FATAL", "worker", "out of memory", "2024-01-01T00:05:00Z"},
	}
	provider := newTestProvider([]*mockCHRows{
		newCHRows(sevCols, sevData),
		newCHRows(errCols, errData),
	}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextLogs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := tc.Data.(map[string]interface{})
	if d["severityDistribution"] == nil {
		t.Error("expected severityDistribution")
	}
	if d["totalLogs"] == nil {
		t.Error("expected totalLogs")
	}
	if d["recentErrors"] == nil {
		t.Error("expected recentErrors")
	}
}

func TestContextCollector_LogsContext_NoData(t *testing.T) {
	provider := newTestProvider([]*mockCHRows{
		newCHRows([]string{"service_name", "severity_text", "total"}, nil),
		newCHRows(nil, nil),
	}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextLogs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(tc.Summary, "No log data") {
		t.Errorf("expected 'No log data' in summary, got %s", tc.Summary)
	}
}

func TestContextCollector_TracesContext_WithData(t *testing.T) {
	latCols := []string{"service_name", "requests", "avg_ms", "p95_ms", "p99_ms"}
	latData := [][]interface{}{
		{"api-server", 1000.0, 45.2, 120.5, 250.0},
	}
	errCols := []string{"service_name", "total_reqs", "errors", "error_rate_pct"}
	errData := [][]interface{}{
		{"api-server", 1000.0, 50.0, 5.0},
	}
	provider := newTestProvider([]*mockCHRows{
		newCHRows(latCols, latData),
		newCHRows(errCols, errData),
	}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextTraces,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := tc.Data.(map[string]interface{})
	if d["serviceLatencies"] == nil {
		t.Error("expected serviceLatencies")
	}
	if d["serviceErrorRates"] == nil {
		t.Error("expected serviceErrorRates")
	}
}

func TestContextCollector_TracesContext_NoData(t *testing.T) {
	provider := newTestProvider([]*mockCHRows{
		newCHRows(nil, nil),
		newCHRows(nil, nil),
	}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextTraces,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(tc.Summary, "No trace data") {
		t.Errorf("expected 'No trace data', got %s", tc.Summary)
	}
}

func TestContextCollector_AccountContext_WithData(t *testing.T) {
	userCols := []string{"id", "email", "first_name", "last_name", "is_active", "is_mfa_enabled", "last_login_at"}
	userData := [][]interface{}{
		{"user-1", "test@example.com", "Test", "User", true, true, "2024-01-01"},
	}
	sessCols := []string{"id", "ip_address", "user_agent", "created_at", "last_active_at", "expires_at"}
	sessData := [][]interface{}{
		{"sess-1", "127.0.0.1", "Mozilla/5.0", "2024-01-01", "2024-01-01", "2024-01-02"},
	}
	provider := newTestProvider(nil, []*mockPgRows{
		{cols: userCols, data: userData},
		{cols: sessCols, data: sessData},
	})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		UserID:         "user-1",
		ContextType:    vo.ContextAccountProfile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := tc.Data.(map[string]interface{})
	if d["user"] == nil {
		t.Error("expected user data")
	}
	if d["sessions"] == nil {
		t.Error("expected sessions data")
	}
}

func TestContextCollector_ContextTypes_CH_WithData(t *testing.T) {
	singleRow := func(cols []string) []*mockCHRows {
		data := make([][]interface{}, len(cols))
		for i := range cols {
			data[i] = []interface{}{fmt.Sprintf("val%d", i)}
		}
		return []*mockCHRows{newCHRows(cols, data)}
	}

	tests := []struct {
		name  string
		ct    vo.ContextType
		cols  []string
		extra []*mockCHRows
	}{
		{"exemplars", vo.ContextExemplars, []string{"timestamp", "metric_name", "service_name", "trace_id", "span_id", "value"}, nil},
		{"correlations", vo.ContextCorrelations, []string{"service_name", "correlation_type", "total"}, nil},
		{"uptime", vo.ContextUptime, []string{"monitor_id", "monitor_name", "region", "status", "total_checks", "up_count", "avg_response_time"}, nil},
		{"audit", vo.ContextAudit, []string{"event_type", "result", "total"}, []*mockCHRows{newCHRows([]string{"timestamp", "user_email", "event_type", "action", "resource", "result"}, [][]interface{}{{"t", "u@e.com", "login", "auth", "system", "ok"}})}},
		{"infra overview", vo.ContextInfraOverview, []string{"vm_id", "metric_name", "avg_val", "max_val"}, nil},
		{"anomaly detection", vo.ContextAnomalyDetection, []string{"timestamp", "detection_rule_id", "metric_name", "signal_type", "severity", "anomaly_score", "z_score", "sigma_level", "observed_value", "expected_value"}, nil},
		{"predictive maintenance", vo.ContextPredictiveMaintenance, []string{"timestamp", "resource_type", "resource_identifier", "horizon", "failure_probability", "health_score", "health_status"}, nil},
		{"cost optimization", vo.ContextCostOptimization, []string{"day", "service_name", "provider", "total_cost_usd"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sets := singleRow(tt.cols)
			if tt.extra != nil {
				sets = append(sets, tt.extra...)
			}
			provider := newTestProvider(sets, nil)
			collector := appsvc.NewContextCollector(provider, zerolog.Nop())
			tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
				OrganizationID: "org-1",
				ContextType:    tt.ct,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc == nil {
				t.Fatal("expected non-nil result")
			}
			if contains(tc.Summary, "unavailable") {
				t.Errorf("expected data, got unavailable: %s", tc.Summary)
			}
		})
	}
}

func TestContextCollector_InfraCPU_WithData(t *testing.T) {
	cols := []string{"vm_id", "metric_name", "avg_val", "max_val"}
	data := [][]interface{}{{"vm-1", "cpu_utilization", 45.5, 92.0}}
	provider := newTestProvider([]*mockCHRows{newCHRows(cols, data)}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextInfraCPU,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_AgentsContext_WithData(t *testing.T) {
	agentCols := []string{"id", "name", "type", "host", "status", "version", "last_heartbeat"}
	agentData := [][]interface{}{{"a1", "agent-1", "vm", "10.0.0.1", "active", "1.0", "2024-01-01"}}
	metricCols := []string{"vm_id", "metric_name", "avg_val", "max_val"}
	metricData := [][]interface{}{{"vm-1", "cpu", 50.0, 80.0}}
	provider := newTestProvider(
		[]*mockCHRows{newCHRows(metricCols, metricData)},
		[]*mockPgRows{{cols: agentCols, data: agentData}},
	)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextAgents,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_ServiceMap_WithData(t *testing.T) {
	svcCols := []string{"id", "name", "type", "status", "environment"}
	svcData := [][]interface{}{{"s1", "api", "http", "healthy", "prod"}}
	depCols := []string{"id", "source_service_id", "target_service_id", "type", "status", "protocol", "latency_ms", "error_rate"}
	depData := [][]interface{}{{"d1", "s1", "s2", "sync", "healthy", "http", 45.0, 0.1}}
	metricCols := []string{"service_id", "service_name", "avg_health_score", "avg_latency", "avg_error_rate"}
	metricData := [][]interface{}{{"s1", "api", 95.0, 50.0, 0.5}}
	provider := newTestProvider(
		[]*mockCHRows{newCHRows(metricCols, metricData)},
		[]*mockPgRows{{cols: svcCols, data: svcData}, {cols: depCols, data: depData}},
	)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextServiceMap,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_Alerts_WithData(t *testing.T) {
	ruleCols := []string{"id", "name", "severity", "state", "enabled", "last_triggered_at"}
	ruleData := [][]interface{}{{"r1", "High CPU", "critical", "firing", true, "2024-01-01"}}
	instCols := []string{"title", "severity", "status", "starts_at", "service"}
	instData := [][]interface{}{{"CPU Alert", "critical", "firing", "2024-01-01", "api-server"}}
	provider := newTestProvider(nil, []*mockPgRows{
		{cols: ruleCols, data: ruleData},
		{cols: instCols, data: instData},
	})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextAlerts,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_IAM_WithData(t *testing.T) {
	userCols := []string{"id", "email", "first_name", "last_name", "is_active", "is_mfa_enabled", "last_login_at", "status"}
	userData := [][]interface{}{{"u1", "admin@test.com", "Admin", "User", true, true, "2024-01-01", "active"}}
	roleCols := []string{"id", "name", "description"}
	roleData := [][]interface{}{{"role-1", "admin", "Administrator"}}
	provider := newTestProvider(nil, []*mockPgRows{
		{cols: userCols, data: userData},
		{cols: roleCols, data: roleData},
	})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextIAM,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_Kubernetes_WithData(t *testing.T) {
	clusterCols := []string{"id", "name", "provider", "region", "status", "node_count", "pod_count", "namespace_count", "version"}
	clusterData := [][]interface{}{{"c1", "prod-1", "aws", "us-east-1", "healthy", 10.0, 50.0, 5.0, "1.28"}}
	metricCols := []string{"resource_type", "resource_name", "namespace", "metric_name", "avg_val", "max_val"}
	metricData := [][]interface{}{{"pod", "api-pod", "default", "cpu", 50.0, 90.0}}
	nodeCols := []string{"name", "status", "roles", "cpu_capacity", "memory_capacity", "conditions"}
	nodeData := [][]interface{}{{"node-1", "ready", "master", "8", "32Gi", "Healthy"}}
	provider := newTestProvider(
		[]*mockCHRows{newCHRows(metricCols, metricData)},
		[]*mockPgRows{{cols: clusterCols, data: clusterData}, {cols: nodeCols, data: nodeData}},
	)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextKubernetesNodes,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_DBEngineMetrics_WithData(t *testing.T) {
	cols := []string{"metric_name", "avg_val", "max_val"}
	data := [][]interface{}{{"db.postgresql.connections", 25.0, 50.0}}
	provider := newTestProvider([]*mockCHRows{newCHRows(cols, data)}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextDBMonitoringPostgreSQL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_DBMySQL_WithData(t *testing.T) {
	metricCols := []string{"metric_name", "metric_value", "labels", "timestamp"}
	metricData := [][]interface{}{{"connections", 25.0, "pool=main", "2024-01-01"}}
	queryCols := []string{"digest_text", "schema_name", "calls", "avg_time_us", "timestamp"}
	queryData := [][]interface{}{{"SELECT * FROM users", "appdb", 1000.0, 150.0, "2024-01-01"}}
	provider := newTestProvider([]*mockCHRows{
		newCHRows(metricCols, metricData),
		newCHRows(queryCols, queryData),
	}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextDBMonitoringMySQL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_CorrectiveMaintenance_WithData(t *testing.T) {
	anomalyCols := []string{"timestamp", "metric_name", "severity", "anomaly_score", "z_score"}
	anomalyData := [][]interface{}{{"2024-01-01", "cpu", "high", 8.5, 3.2}}
	planCols := []string{"id", "trigger_type", "title", "status", "risk_level", "created_at"}
	planData := [][]interface{}{{"p1", "anomaly", "Fix CPU", "pending", "medium", "2024-01-01"}}
	provider := newTestProvider(
		[]*mockCHRows{newCHRows(anomalyCols, anomalyData)},
		[]*mockPgRows{{cols: planCols, data: planData}},
	)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextCorrectiveMaintenance,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestContextCollector_DBMonitoringQAN_WithData(t *testing.T) {
	cols := []string{"fingerprint", "avg_duration_ms", "calls", "timestamp"}
	data := [][]interface{}{{"SELECT ?", 50.0, 100.0, "2024-01-01"}}
	sets := make([]*mockCHRows, 4)
	for i := range sets {
		sets[i] = newCHRows(cols, data)
	}
	provider := newTestProvider(sets, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextDBMonitoringQAN,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_SystemSetup_WithData(t *testing.T) {
	apiKeyCols := []string{"total", "active"}
	apiKeyData := [][]interface{}{{5.0, 3.0}}
	channelCols := []string{"total", "enabled"}
	channelData := [][]interface{}{{3.0, 2.0}}
	ssoCols := []string{"provider_type", "is_enabled"}
	ssoData := [][]interface{}{{"saml", true}}
	provider := newTestProvider(nil, []*mockPgRows{
		{cols: apiKeyCols, data: apiKeyData},
		{cols: channelCols, data: channelData},
		{cols: ssoCols, data: ssoData},
	})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextSystemSetup,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestContextCollector_NetworkMap_WithData(t *testing.T) {
	nodeCols := []string{"id", "name", "type", "status", "ip_address", "region"}
	nodeData := [][]interface{}{{"n1", "router-1", "router", "up", "10.0.0.1", "us-east"}}
	connCols := []string{"id", "source_node_id", "target_node_id", "type", "status", "protocol", "latency"}
	connData := [][]interface{}{{"c1", "n1", "n2", "direct", "up", "tcp", 5.0}}
	trafficCols := []string{"node_id", "node_name", "avg_cpu_usage", "avg_memory_usage", "avg_network_in", "avg_network_out"}
	trafficData := [][]interface{}{{"n1", "router-1", 30.0, 40.0, 100.0, 80.0}}
	provider := newTestProvider(
		[]*mockCHRows{newCHRows(trafficCols, trafficData)},
		[]*mockPgRows{{cols: nodeCols, data: nodeData}, {cols: connCols, data: connData}},
	)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextNetworkMap,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_Tenancy_WithData(t *testing.T) {
	orgCols := []string{"id", "name", "plan", "region", "status"}
	orgData := [][]interface{}{{"o1", "Acme Corp", "enterprise", "us-east", "active"}}
	wsCols := []string{"id", "name", "status", "created_at"}
	wsData := [][]interface{}{{"w1", "Production", "active", "2024-01-01"}}
	provider := newTestProvider(nil, []*mockPgRows{
		{cols: orgCols, data: orgData},
		{cols: wsCols, data: wsData},
	})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextTenancy,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(tc.Summary, "unavailable") {
		t.Errorf("expected data, got unavailable: %s", tc.Summary)
	}
}

func TestContextCollector_Reports_WithData(t *testing.T) {
	defCols := []string{"id", "name", "type", "schedule", "enabled", "created_at"}
	defData := [][]interface{}{{"r1", "Weekly Report", "summary", "weekly", true, "2024-01-01"}}
	execCols := []string{"report_definition_id", "status", "started_at", "completed_at"}
	execData := [][]interface{}{{"r1", "completed", "2024-01-01", "2024-01-01"}}
	provider := newTestProvider(nil, []*mockPgRows{
		{cols: defCols, data: defData},
		{cols: execCols, data: execData},
	})
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextReports,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestContextCollector_DBAurora_WithData(t *testing.T) {
	topoCols := []string{"id", "cluster_identifier", "engine_type", "engine_version", "cluster_status", "writer_instance_id", "reader_instance_ids"}
	topoData := [][]interface{}{{"a1", "prod-cluster", "aurora-mysql", "8.0", "available", "i1", "[]"}}
	metricCols := []string{"cluster_identifier", "metric_name", "avg_val", "max_val"}
	metricData := [][]interface{}{{"prod-cluster", "connections", 50.0, 100.0}}
	provider := newTestProvider(
		[]*mockCHRows{newCHRows(metricCols, metricData)},
		[]*mockPgRows{{cols: topoCols, data: topoData}},
	)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextDBMonitoringAurora,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestContextCollector_MaxItems_Default(t *testing.T) {
	cols := []string{"service_name", "metric_name", "metric_type", "avg_val", "max_val", "min_val", "sample_count"}
	data := [][]interface{}{{"svc", "cpu", "gauge", 50.0, 90.0, 10.0, 100.0}}
	provider := newTestProvider([]*mockCHRows{newCHRows(cols, data)}, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextMetrics,
		MaxItems:       0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil result")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestContextCollector_CollectContext_Timeout(t *testing.T) {
	provider := newTestProvider(nil, nil)
	provider.hasCH = false
	provider.hasPG = false
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc, err := collector.CollectContext(ctx, vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextMetrics,
	})
	if err != nil {
		t.Fatalf("expected nil error on timeout, got %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil telemetry context on timeout")
	}
}

func TestContextCollector_Dispatch_DefaultBranch(t *testing.T) {
	provider := newTestProvider(nil, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextType("unknown-type"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil telemetry context for unknown type")
	}
	if tc.Summary == "" {
		t.Error("expected non-empty summary for default branch")
	}
}

func TestContextCollector_dbEngineFromType_Default(t *testing.T) {
	provider := newTestProvider(nil, nil)
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextDBMonitoringPostgreSQL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil")
	}
}

func TestContextCollector_buildHighlights_Empty(t *testing.T) {
	provider := newTestProvider(nil, nil)
	provider.hasCH = false
	provider.hasPG = false
	collector := appsvc.NewContextCollector(provider, zerolog.Nop())

	tc, err := collector.CollectContext(context.Background(), vo.CollectContextOptions{
		OrganizationID: "org-1",
		ContextType:    vo.ContextMetrics,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc == nil {
		t.Fatal("expected non-nil")
	}
}
