// Package services provides functionality for the TelemetryFlow GO MCP Server.
//
// TelemetryFlow GO MCP Server - Community Enterprise Observability Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
// Open Source Software built by Telemetri Data Indonesia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type DBProvider interface {
	GormDB() *gorm.DB
	ClickHouseConn() driver.Conn
	ClickHouseDB() string
	HasClickHouse() bool
	HasPostgres() bool
}

type ContextCollector struct {
	db     DBProvider
	logger zerolog.Logger
}

func NewContextCollector(db DBProvider, logger zerolog.Logger) *ContextCollector {
	return &ContextCollector{
		db:     db,
		logger: logger,
	}
}

func (c *ContextCollector) CollectContext(ctx context.Context, opts vo.CollectContextOptions) (*vo.TelemetryContext, error) {
	timeRange := opts.TimeRange
	if timeRange == nil {
		tr := vo.DefaultTimeRange()
		timeRange = &tr
	}

	collectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resultCh := make(chan *vo.TelemetryContext, 1)
	errCh := make(chan error, 1)

	go func() {
		tc, err := c.dispatch(collectCtx, opts.OrganizationID, opts.UserID, opts.ContextType, *timeRange, opts.MaxItems)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- tc
	}()

	select {
	case <-collectCtx.Done():
		c.logger.Warn().Str("type", string(opts.ContextType)).Msg("Context collection timed out")
		return c.unavailable(opts.ContextType, *timeRange, "collection timed out"), nil
	case err := <-errCh:
		c.logger.Error().Err(err).Str("type", string(opts.ContextType)).Msg("Context collection failed")
		return c.unavailable(opts.ContextType, *timeRange, err.Error()), nil
	case tc := <-resultCh:
		return tc, nil
	}
}

func (c *ContextCollector) dispatch(ctx context.Context, orgID, userID string, contextType vo.ContextType, timeRange vo.TimeRange, maxItems int) (*vo.TelemetryContext, error) {
	if maxItems <= 0 {
		maxItems = 30
	}

	switch contextType {
	case vo.ContextMetrics:
		return c.metricsContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextLogs:
		return c.logsContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextTraces:
		return c.tracesContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextExemplars:
		return c.exemplarsContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextCorrelations, vo.ContextDashboard:
		return c.correlationsContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextUptime:
		return c.uptimeContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextStatusPage:
		return c.statusPageContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextAudit:
		return c.auditContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextInfraOverview:
		return c.infraOverviewContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextInfraCPU, vo.ContextInfraMemory, vo.ContextInfraStorage, vo.ContextInfraNetwork:
		metric := infraTypeToMetric(contextType)
		return c.infraContext(ctx, orgID, timeRange, metric, maxItems)
	case vo.ContextAgents:
		return c.agentsContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextServiceMap:
		return c.serviceMapContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextNetworkMap:
		return c.networkMapContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextAlerts, vo.ContextAlertRules:
		return c.alertsContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextIAM, vo.ContextIAMUsers, vo.ContextIAMRoles, vo.ContextIAMPermissions, vo.ContextIAMMatrix, vo.ContextIAMAssignments:
		return c.iamContext(ctx, orgID, userID, contextType, maxItems)
	case vo.ContextTenancy, vo.ContextTenancyRegions, vo.ContextTenancyOrganizations, vo.ContextTenancyWorkspaces, vo.ContextTenancyTenants:
		return c.tenancyContext(ctx, orgID, contextType, maxItems)
	case vo.ContextRetention:
		return c.retentionContext(ctx, orgID, maxItems)
	case vo.ContextSubscription:
		return c.subscriptionContext(ctx, orgID, maxItems)
	case vo.ContextAPIKeys:
		return c.apiKeysContext(ctx, orgID, maxItems)
	case vo.ContextNotifications, vo.ContextSystemChannels, vo.ContextAccountNotifications:
		return c.notificationsContext(ctx, orgID, maxItems)
	case vo.ContextReports:
		return c.reportsContext(ctx, orgID, maxItems)
	case vo.ContextDataMasking:
		return c.dataMaskingContext(ctx, orgID, maxItems)
	case vo.ContextAIAssistant:
		return c.aiAssistantContext(ctx, orgID, maxItems)
	case vo.ContextSystemSetup:
		return c.systemSetupContext(ctx, orgID, maxItems)
	case vo.ContextAccountProfile, vo.ContextAccountSecurity, vo.ContextAccountSessions,
		vo.ContextAccountPreferences, vo.ContextAccountOrganization:
		return c.accountContext(ctx, userID, orgID, contextType, maxItems)
	case vo.ContextKubernetesOverview, vo.ContextKubernetesClusters, vo.ContextKubernetesNamespaces,
		vo.ContextKubernetesNodes, vo.ContextKubernetesPods, vo.ContextKubernetesDeployments,
		vo.ContextKubernetesPV, vo.ContextKubernetesAPIServer, vo.ContextKubernetesCoreDNS:
		return c.kubernetesContext(ctx, orgID, timeRange, contextType, maxItems)
	case vo.ContextAnomalyDetection:
		return c.anomalyDetectionContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextCorrectiveMaintenance:
		return c.correctiveMaintenanceContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextPredictiveMaintenance:
		return c.predictiveMaintenanceContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextCostOptimization:
		return c.costOptimizationContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringInventory:
		return c.dbInventoryContext(ctx, orgID, maxItems)
	case vo.ContextDBMonitoringMariaDB:
		return c.dbMariaDBContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringMySQL:
		return c.dbMySQLContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringPercona:
		return c.dbPerconaContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringSQLite3:
		return c.dbSQLite3Context(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringTimescaleDB:
		return c.dbTimescaleDBContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringAurora:
		return c.dbAuroraContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringMSSQL:
		return c.dbMSSQLContext(ctx, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringPostgreSQL, vo.ContextDBMonitoringMongoDBCommunity,
		vo.ContextDBMonitoringMongoDBAtlas, vo.ContextDBMonitoringAWSRDSMySQL,
		vo.ContextDBMonitoringAWSRDSAurora, vo.ContextDBMonitoringAWSDynamoDB,
		vo.ContextDBMonitoringCockroachDB:
		return c.dbEngineMetricsContext(ctx, contextType, orgID, timeRange, maxItems)
	case vo.ContextDBMonitoringQAN:
		return c.dbMonitoringQANContext(ctx, orgID, timeRange, maxItems)
	default:
		return c.empty(contextType, timeRange), nil
	}
}

func (c *ContextCollector) unavailable(contextType vo.ContextType, timeRange vo.TimeRange, reason string) *vo.TelemetryContext {
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: timeRange,
		Summary:   fmt.Sprintf("[SYSTEM] Data source unavailable: %s", reason),
		Data:      map[string]interface{}{"unavailable": true, "reason": reason},
	}
}

func (c *ContextCollector) empty(contextType vo.ContextType, timeRange vo.TimeRange) *vo.TelemetryContext {
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: timeRange,
		Summary:   "No data available for the requested context type.",
		Data:      map[string]interface{}{},
	}
}

func fmtCH(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func infraTypeToMetric(ct vo.ContextType) string {
	switch ct {
	case vo.ContextInfraCPU:
		return "cpu"
	case vo.ContextInfraMemory:
		return "memory"
	case vo.ContextInfraStorage:
		return "disk"
	case vo.ContextInfraNetwork:
		return "network"
	default:
		return "cpu"
	}
}

func (c *ContextCollector) chQuery(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	if !c.db.HasClickHouse() {
		return nil, fmt.Errorf("ClickHouse not connected")
	}
	conn := c.db.ClickHouseConn()
	db := c.db.ClickHouseDB()
	query = strings.ReplaceAll(query, "{db}", db)

	rows, err := conn.Query(ctx, query, params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []map[string]interface{}
	colTypes := rows.ColumnTypes()
	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(colTypes))
		valuePtrs := make([]interface{}, len(colTypes))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		for i, ct := range colTypes {
			row[ct.Name()] = values[i]
		}
		results = append(results, row)
	}
	return results, nil
}

func (c *ContextCollector) pgQuery(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if !c.db.HasPostgres() {
		return nil, fmt.Errorf("PostgreSQL not connected")
	}
	rows, err := c.db.GormDB().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []map[string]interface{}
	cols, _ := rows.Columns()
	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	return results, nil
}

func (c *ContextCollector) metricsContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT service_name, metric_name, metric_type,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val,
		       round(minMerge(min_value), 4) AS min_val,
		       countMerge(count) AS sample_count
		FROM {db}.metrics_5m
		WHERE five_minutes >= @from AND five_minutes <= @to
		  AND organization_id = @orgId
		GROUP BY service_name, metric_name, metric_type
		ORDER BY sample_count DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextMetrics, timeRange, err.Error()), nil
	}

	highlights := buildHighlights(rows, "metric", timeRange)
	return &vo.TelemetryContext{
		Type:      vo.ContextMetrics,
		TimeRange: timeRange,
		Summary:   strings.Join(highlights, " "),
		Data:      map[string]interface{}{"metrics": rows, "highlights": highlights},
	}, nil
}

func (c *ContextCollector) logsContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	sevRows, err := c.chQuery(ctx, `
		SELECT service_name, severity_text, sum(count) AS total
		FROM {db}.logs_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		GROUP BY service_name, severity_text
		ORDER BY total DESC
		LIMIT 50`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextLogs, timeRange, err.Error()), nil
	}

	errRows, _ := c.chQuery(ctx, `
		SELECT severity_text, service_name,
		       substring(body, 1, 300) AS body, timestamp
		FROM {db}.logs
		WHERE timestamp >= @from AND organization_id = @orgId
		  AND severity_text IN ('ERROR', 'FATAL', 'CRITICAL')
		ORDER BY timestamp DESC
		LIMIT 20`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"orgId": orgID,
		},
	)

	bySeverity := make(map[string]float64)
	totalLogs := 0.0
	for _, row := range sevRows {
		if t, ok := row["total"]; ok {
			switch v := t.(type) {
			case float64:
				bySeverity[fmt.Sprintf("%v", row["severity_text"])] += v
				totalLogs += v
			case uint64:
				bySeverity[fmt.Sprintf("%v", row["severity_text"])] += float64(v)
				totalLogs += float64(v)
			}
		}
	}

	var highlights []string
	if totalLogs == 0 {
		highlights = append(highlights, "No log data available for the selected time range.")
	} else {
		highlights = append(highlights, fmt.Sprintf("%.0f total log entries.", totalLogs))
		if e, ok := bySeverity["ERROR"]; ok && e > 0 {
			highlights = append(highlights, fmt.Sprintf("%.0f ERROR logs.", e))
		}
		if f, ok := bySeverity["FATAL"]; ok && f > 0 {
			highlights = append(highlights, fmt.Sprintf("%.0f FATAL logs.", f))
		}
	}

	return &vo.TelemetryContext{
		Type:      vo.ContextLogs,
		TimeRange: timeRange,
		Summary:   strings.Join(highlights, " "),
		Data: map[string]interface{}{
			"severityDistribution": bySeverity,
			"totalLogs":            totalLogs,
			"recentErrors":         errRows,
			"highlights":           highlights,
		},
	}, nil
}

func (c *ContextCollector) tracesContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	latencyRows, err := c.chQuery(ctx, `
		SELECT service_name,
		       countMerge(request_count) AS requests,
		       round(avgMerge(avg_duration_ns) / 1e6, 2) AS avg_ms,
		       round(quantileMerge(0.95)(p95_duration_ns) / 1e6, 2) AS p95_ms,
		       round(quantileMerge(0.99)(p99_duration_ns) / 1e6, 2) AS p99_ms
		FROM {db}.service_latency_percentiles_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		GROUP BY service_name
		ORDER BY p95_ms DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextTraces, timeRange, err.Error()), nil
	}

	errRateRows, _ := c.chQuery(ctx, `
		SELECT service_name,
		       sum(total_requests) AS total_reqs,
		       sum(error_count) AS errors,
		       round(sum(error_count) * 100.0 / sum(total_requests), 2) AS error_rate_pct
		FROM {db}.service_error_rates_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		GROUP BY service_name
		HAVING total_reqs > 0
		ORDER BY error_rate_pct DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)

	var highlights []string
	if len(latencyRows) == 0 && len(errRateRows) == 0 {
		highlights = append(highlights, "No trace data available for the selected time range.")
	} else {
		if len(latencyRows) > 0 {
			highlights = append(highlights, fmt.Sprintf("Latency data for %d service(s).", len(latencyRows)))
		}
		if len(errRateRows) > 0 {
			highlights = append(highlights, fmt.Sprintf("Error rate data for %d service(s).", len(errRateRows)))
		}
	}

	return &vo.TelemetryContext{
		Type:      vo.ContextTraces,
		TimeRange: timeRange,
		Summary:   strings.Join(highlights, " "),
		Data: map[string]interface{}{
			"serviceLatencies":  latencyRows,
			"serviceErrorRates": errRateRows,
			"highlights":        highlights,
		},
	}, nil
}

func (c *ContextCollector) exemplarsContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT timestamp, metric_name, service_name, trace_id, span_id, value
		FROM {db}.exemplars
		WHERE timestamp >= @from AND timestamp <= @to
		  AND organization_id = @orgId
		ORDER BY timestamp DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextExemplars, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Found %d exemplars.", len(rows))
	if len(rows) == 0 {
		summary = "No exemplar data available for the selected time range."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextExemplars,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"exemplars": rows},
	}, nil
}

func (c *ContextCollector) correlationsContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT service_name, correlation_type, sum(count) AS total
		FROM {db}.signal_correlations_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		GROUP BY service_name, correlation_type
		ORDER BY total DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextCorrelations, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Found %d correlations.", len(rows))
	if len(rows) == 0 {
		summary = "No correlation data available for the selected time range."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextCorrelations,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"correlations": rows},
	}, nil
}

func (c *ContextCollector) uptimeContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT monitor_id, monitor_name, region, status,
		       count() AS total_checks,
		       sum(CASE WHEN status = 'up' THEN 1 ELSE 0 END) AS up_count,
		       avg(response_time) AS avg_response_time
		FROM {db}.uptime_checks
		WHERE checked_at >= @from AND checked_at <= @to
		  AND organization_id = @orgId
		GROUP BY monitor_id, monitor_name, region, status
		ORDER BY total_checks DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextUptime, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Uptime data for %d monitor(s).", len(rows))
	if len(rows) == 0 {
		summary = "No uptime data available for the selected time range."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextUptime,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"checks": rows},
	}, nil
}

func (c *ContextCollector) statusPageContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	return c.uptimeContext(ctx, orgID, timeRange, limit)
}

func (c *ContextCollector) auditContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT event_type, result, sum(count) AS total
		FROM {db}.audit_logs_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		GROUP BY event_type, result
		ORDER BY total DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextAudit, timeRange, err.Error()), nil
	}

	recentEvents, _ := c.chQuery(ctx, `
		SELECT timestamp, user_email, event_type, action, resource, result
		FROM {db}.audit_logs
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC
		LIMIT 20`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"orgId": orgID,
		},
	)

	summary := fmt.Sprintf("Audit stats: %d event types. Recent: %d events.", len(rows), len(recentEvents))
	return &vo.TelemetryContext{
		Type:      vo.ContextAudit,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"stats": rows, "recent": recentEvents},
	}, nil
}

func (c *ContextCollector) infraOverviewContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT vm_id, metric_name,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val
		FROM {db}.vm_metrics_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		GROUP BY vm_id, metric_name
		ORDER BY max_val DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextInfraOverview, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Infrastructure overview: %d metric series.", len(rows))
	if len(rows) == 0 {
		summary = "No infrastructure metrics available for the selected time range."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextInfraOverview,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": rows},
	}, nil
}

func (c *ContextCollector) infraContext(ctx context.Context, orgID string, timeRange vo.TimeRange, metric string, limit int) (*vo.TelemetryContext, error) {
	pattern := fmt.Sprintf("%%%s%%", metric)
	rows, err := c.chQuery(ctx, `
		SELECT vm_id, metric_name,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val
		FROM {db}.vm_metrics_1h
		WHERE hour >= @from AND hour <= @to
		  AND organization_id = @orgId
		  AND metric_name LIKE @pattern
		GROUP BY vm_id, metric_name
		ORDER BY max_val DESC
		LIMIT @limit`,
		map[string]interface{}{
			"from":    fmtCH(timeRange.From),
			"to":      fmtCH(timeRange.To),
			"orgId":   orgID,
			"pattern": pattern,
			"limit":   limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextType("infra-"+metric), timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Infrastructure %s metrics: %d series.", metric, len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextType("infra-" + metric),
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": rows},
	}, nil
}

func (c *ContextCollector) agentsContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	pgAgents, pgErr := c.pgQuery(ctx, `
		SELECT id, name, type, host, status, version, last_heartbeat
		FROM agents WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	chMetrics, _ := c.chQuery(ctx, `
		SELECT vm_id, metric_name,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val
		FROM {db}.vm_metrics_1h
		WHERE hour >= @from AND hour <= @to AND organization_id = @orgId
		GROUP BY vm_id, metric_name
		ORDER BY max_val DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)

	if pgErr != nil && len(chMetrics) == 0 {
		return c.unavailable(vo.ContextAgents, timeRange, pgErr.Error()), nil
	}

	agentCount := 0
	if pgAgents != nil {
		agentCount = len(pgAgents)
	}
	summary := fmt.Sprintf("Agents: %d registered. VM metrics: %d series.", agentCount, len(chMetrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextAgents,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"agents": pgAgents, "metrics": chMetrics},
	}, nil
}

func (c *ContextCollector) serviceMapContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	services, _ := c.pgQuery(ctx, `
		SELECT id, name, type, status, environment
		FROM service_map_services
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	deps, _ := c.pgQuery(ctx, `
		SELECT id, source_service_id, target_service_id, type, status, protocol, latency_ms, error_rate
		FROM service_map_dependencies
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit*2)

	metrics, _ := c.chQuery(ctx, `
		SELECT service_id, service_name,
		       avg(avg_health_score) AS avg_health_score,
		       avg(avg_latency) AS avg_latency,
		       avg(avg_error_rate) AS avg_error_rate
		FROM {db}.service_map_metrics_1h
		WHERE hour >= @from AND hour <= @to AND organization_id = @orgId
		GROUP BY service_id, service_name
		ORDER BY avg_error_rate DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)

	svcCount := 0
	if services != nil {
		svcCount = len(services)
	}
	depCount := 0
	if deps != nil {
		depCount = len(deps)
	}
	summary := fmt.Sprintf("Service map: %d services, %d dependencies.", svcCount, depCount)
	return &vo.TelemetryContext{
		Type:      vo.ContextServiceMap,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"services": services, "dependencies": deps, "metrics": metrics},
	}, nil
}

func (c *ContextCollector) networkMapContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	nodes, _ := c.pgQuery(ctx, `
		SELECT id, name, type, status, ip_address, region
		FROM network_map_nodes
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	conns, _ := c.pgQuery(ctx, `
		SELECT id, source_node_id, target_node_id, type, status, protocol, latency
		FROM network_map_connections
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit*2)

	traffic, _ := c.chQuery(ctx, `
		SELECT node_id, node_name,
		       avg(avg_cpu_usage) AS avg_cpu_usage,
		       avg(avg_memory_usage) AS avg_memory_usage,
		       avg(avg_network_in) AS avg_network_in,
		       avg(avg_network_out) AS avg_network_out
		FROM {db}.network_map_traffic_1h
		WHERE hour >= @from AND hour <= @to AND organization_id = @orgId
		GROUP BY node_id, node_name
		ORDER BY avg_cpu_usage DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)

	nodeCount := 0
	if nodes != nil {
		nodeCount = len(nodes)
	}
	connCount := 0
	if conns != nil {
		connCount = len(conns)
	}
	summary := fmt.Sprintf("Network map: %d nodes, %d connections.", nodeCount, connCount)
	return &vo.TelemetryContext{
		Type:      vo.ContextNetworkMap,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"nodes": nodes, "connections": conns, "traffic": traffic},
	}, nil
}

func (c *ContextCollector) alertsContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rules, _ := c.pgQuery(ctx, `
		SELECT id, name, severity, state, enabled, last_triggered_at
		FROM alert_rules
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	instances, _ := c.pgQuery(ctx, `
		SELECT title, severity, status, starts_at, labels->>'service' AS service
		FROM alert_instances
		WHERE organization_id = $1 LIMIT $2`, orgID, limit)

	ruleCount := 0
	if rules != nil {
		ruleCount = len(rules)
	}
	instCount := 0
	if instances != nil {
		instCount = len(instances)
	}
	summary := fmt.Sprintf("Alerts: %d rules, %d active instances.", ruleCount, instCount)
	return &vo.TelemetryContext{
		Type:      vo.ContextAlerts,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"rules": rules, "instances": instances},
	}, nil
}

func (c *ContextCollector) iamContext(ctx context.Context, orgID, userID string, contextType vo.ContextType, limit int) (*vo.TelemetryContext, error) {
	data := make(map[string]interface{})

	users, _ := c.pgQuery(ctx, `
		SELECT id, email, first_name, last_name, is_active, is_mfa_enabled, last_login_at, status
		FROM users WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)
	data["users"] = users

	roles, _ := c.pgQuery(ctx, `
		SELECT id, name, description
		FROM roles WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)
	data["roles"] = roles

	summary := fmt.Sprintf("IAM data: %d users, %d roles.", len(users), len(roles))
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: vo.DefaultTimeRange(),
		Summary:   summary,
		Data:      data,
	}, nil
}

func (c *ContextCollector) tenancyContext(ctx context.Context, orgID string, contextType vo.ContextType, limit int) (*vo.TelemetryContext, error) {
	data := make(map[string]interface{})

	orgs, _ := c.pgQuery(ctx, `
		SELECT id, name, plan, region, status
		FROM organizations WHERE deleted_at IS NULL LIMIT $1`, limit)
	data["organizations"] = orgs

	workspaces, _ := c.pgQuery(ctx, `
		SELECT id, name, status, created_at
		FROM workspaces WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)
	data["workspaces"] = workspaces

	summary := fmt.Sprintf("Tenancy data: %d organizations, %d workspaces.", len(orgs), len(workspaces))
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: vo.DefaultTimeRange(),
		Summary:   summary,
		Data:      data,
	}, nil
}

func (c *ContextCollector) retentionContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	rows, _ := c.pgQuery(ctx, `
		SELECT id, name, data_type, retention_days, archive_enabled, is_default, is_active
		FROM retention_policies
		WHERE organization_id = $1 LIMIT $2`, orgID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("Retention policies: %d configured.", len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextRetention,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"policies": rows},
	}, nil
}

func (c *ContextCollector) subscriptionContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	rows, _ := c.pgQuery(ctx, `
		SELECT id, name, type, is_active, trial_days
		FROM plans LIMIT $1`, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("Subscription plans: %d available.", len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextSubscription,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"plans": rows},
	}, nil
}

func (c *ContextCollector) apiKeysContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	rows, _ := c.pgQuery(ctx, `
		SELECT id, name, key_hint, is_active, expires_at, last_used_at, usage_count
		FROM api_keys
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("API keys: %d registered.", len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextAPIKeys,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"keys": rows},
	}, nil
}

func (c *ContextCollector) notificationsContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	rows, _ := c.pgQuery(ctx, `
		SELECT id, name, type, enabled, verified, last_tested_at
		FROM notification_channels
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("Notification channels: %d configured.", len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextNotifications,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"channels": rows},
	}, nil
}

func (c *ContextCollector) reportsContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	defs, _ := c.pgQuery(ctx, `
		SELECT id, name, type, schedule, enabled, created_at
		FROM report_definitions
		WHERE organization_id = $1 LIMIT $2`, orgID, limit)

	execs, _ := c.pgQuery(ctx, `
		SELECT report_definition_id, status, started_at, completed_at
		FROM report_executions
		WHERE organization_id = $1
		ORDER BY started_at DESC LIMIT $2`, orgID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("Reports: %d definitions, %d executions.", len(defs), len(execs))
	return &vo.TelemetryContext{
		Type:      vo.ContextReports,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"definitions": defs, "executions": execs},
	}, nil
}

func (c *ContextCollector) dataMaskingContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	rows, _ := c.pgQuery(ctx, `
		SELECT id, name, is_enabled, rules, created_at, updated_at
		FROM data_masking_policies
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("Data masking policies: %d configured.", len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextDataMasking,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"policies": rows},
	}, nil
}

func (c *ContextCollector) aiAssistantContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	rows, _ := c.pgQuery(ctx, `
		SELECT id, name, provider_type, model_id, api_key_hint, is_default, is_active, usage_count, last_used_at
		FROM llm_providers
		WHERE organization_id = $1 LIMIT $2`, orgID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("AI providers: %d configured.", len(rows))
	return &vo.TelemetryContext{
		Type:      vo.ContextAIAssistant,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"providers": rows},
	}, nil
}

func (c *ContextCollector) systemSetupContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	apiKeys, _ := c.pgQuery(ctx, `
		SELECT count(*) AS total, count(CASE WHEN is_active THEN 1 END) AS active
		FROM api_keys WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)

	channels, _ := c.pgQuery(ctx, `
		SELECT count(*) AS total, count(CASE WHEN enabled THEN 1 END) AS enabled
		FROM notification_channels WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)

	sso, _ := c.pgQuery(ctx, `
		SELECT provider_type, is_enabled
		FROM sso_providers WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)

	tr := vo.DefaultTimeRange()
	summary := "System setup data retrieved."
	return &vo.TelemetryContext{
		Type:      vo.ContextSystemSetup,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"apiKeys": apiKeys, "channels": channels, "ssoProviders": sso},
	}, nil
}

func (c *ContextCollector) accountContext(ctx context.Context, userID, orgID string, contextType vo.ContextType, limit int) (*vo.TelemetryContext, error) {
	if userID == "" {
		return c.unavailable(contextType, vo.DefaultTimeRange(), "user ID required"), nil
	}

	user, _ := c.pgQuery(ctx, `
		SELECT id, email, first_name, last_name, is_active, is_mfa_enabled, last_login_at
		FROM users WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`, userID, orgID)

	sessions, _ := c.pgQuery(ctx, `
		SELECT id, ip_address, user_agent, created_at, last_active_at, expires_at
		FROM user_sessions WHERE user_id = $1 ORDER BY last_active_at DESC LIMIT $2`, userID, limit)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("Account data for user %s.", userID)
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"user": user, "sessions": sessions},
	}, nil
}

func (c *ContextCollector) kubernetesContext(ctx context.Context, orgID string, timeRange vo.TimeRange, contextType vo.ContextType, limit int) (*vo.TelemetryContext, error) {
	data := make(map[string]interface{})

	clusters, _ := c.pgQuery(ctx, `
		SELECT id, name, provider, region, status, node_count, pod_count, namespace_count, version
		FROM kubernetes_clusters
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)
	data["clusters"] = clusters

	chMetrics, _ := c.chQuery(ctx, `
		SELECT resource_type, resource_name, namespace, metric_name,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val
		FROM {db}.kubernetes_metrics_1h
		WHERE hour >= @from AND hour <= @to AND organization_id = @orgId
		GROUP BY resource_type, resource_name, namespace, metric_name
		ORDER BY max_val DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	data["metrics"] = chMetrics

	switch contextType {
	case vo.ContextKubernetesNodes:
		nodes, _ := c.pgQuery(ctx, `
			SELECT kn.name, kn.status, kn.roles, kn.cpu_capacity, kn.memory_capacity, kn.conditions
			FROM kubernetes_nodes kn
			JOIN kubernetes_clusters kc ON kn.cluster_id = kc.id
			WHERE kc.organization_id = $1 AND kc.deleted_at IS NULL`, orgID)
		data["nodes"] = nodes
	case vo.ContextKubernetesPods:
		pods, _ := c.pgQuery(ctx, `
			SELECT kp.name, kp.phase, kp.restart_count
			FROM kubernetes_pods kp
			JOIN kubernetes_clusters kc ON kp.cluster_id = kc.id
			WHERE kc.organization_id = $1 AND kc.deleted_at IS NULL`, orgID)
		data["pods"] = pods
	case vo.ContextKubernetesDeployments:
		deploys, _ := c.pgQuery(ctx, `
			SELECT kd.name, kd.replicas, kd.ready_replicas, kd.unavailable_replicas
			FROM kubernetes_deployments kd
			JOIN kubernetes_clusters kc ON kd.cluster_id = kc.id
			WHERE kc.organization_id = $1 AND kc.deleted_at IS NULL`, orgID)
		data["deployments"] = deploys
	}

	clusterCount := 0
	if clusters != nil {
		clusterCount = len(clusters)
	}
	summary := fmt.Sprintf("Kubernetes %s: %d clusters.", contextType, clusterCount)
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      data,
	}, nil
}

func (c *ContextCollector) anomalyDetectionContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT timestamp, detection_rule_id, metric_name, signal_type,
		       severity, anomaly_score, z_score, sigma_level,
		       observed_value, expected_value
		FROM {db}.anomaly_events
		WHERE timestamp >= @from AND timestamp <= @to
		  AND organization_id = @orgId
		ORDER BY anomaly_score DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextAnomalyDetection, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Anomaly detection: %d events.", len(rows))
	if len(rows) == 0 {
		summary = "No anomaly events in the selected time range."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextAnomalyDetection,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"anomalies": rows},
	}, nil
}

func (c *ContextCollector) correctiveMaintenanceContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	anomalies, _ := c.chQuery(ctx, `
		SELECT timestamp, metric_name, severity, anomaly_score, z_score
		FROM {db}.anomaly_events
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY anomaly_score DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"orgId": orgID,
			"limit": limit,
		},
	)

	plans, _ := c.pgQuery(ctx, `
		SELECT id, trigger_type, title, status, risk_level, created_at
		FROM remediation_plans
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $2`, orgID, limit)

	summary := fmt.Sprintf("Corrective maintenance: %d anomalies, %d remediation plans.", len(anomalies), len(plans))
	return &vo.TelemetryContext{
		Type:      vo.ContextCorrectiveMaintenance,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"anomalies": anomalies, "plans": plans},
	}, nil
}

func (c *ContextCollector) predictiveMaintenanceContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT timestamp, resource_type, resource_identifier, horizon,
		       failure_probability, health_score, health_status
		FROM {db}.predictions
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY failure_probability DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextPredictiveMaintenance, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Predictive maintenance: %d predictions.", len(rows))
	if len(rows) == 0 {
		summary = "No predictions available."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextPredictiveMaintenance,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"predictions": rows},
	}, nil
}

func (c *ContextCollector) costOptimizationContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	rows, err := c.chQuery(ctx, `
		SELECT day, service_name, provider, total_cost_usd
		FROM {db}.cost_data_daily_mv
		WHERE day >= @from AND day <= @to AND organization_id = @orgId
		ORDER BY total_cost_usd DESC LIMIT @limit`,
		map[string]interface{}{
			"from":  fmtCH(timeRange.From),
			"to":    fmtCH(timeRange.To),
			"orgId": orgID,
			"limit": limit,
		},
	)
	if err != nil {
		return c.unavailable(vo.ContextCostOptimization, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("Cost optimization: %d cost entries.", len(rows))
	if len(rows) == 0 {
		summary = "No cost data available."
	}
	return &vo.TelemetryContext{
		Type:      vo.ContextCostOptimization,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"costs": rows},
	}, nil
}

func (c *ContextCollector) dbInventoryContext(ctx context.Context, orgID string, limit int) (*vo.TelemetryContext, error) {
	instances, _ := c.pgQuery(ctx, `
		SELECT id, name, type, host, port, status, provider, last_error, last_seen_at
		FROM database_instances
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	rules, _ := c.pgQuery(ctx, `
		SELECT count(*) AS total, count(CASE WHEN enabled THEN 1 END) AS enabled
		FROM database_monitoring_rules
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)

	tr := vo.DefaultTimeRange()
	summary := fmt.Sprintf("DB inventory: %d instances.", len(instances))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringInventory,
		TimeRange: tr,
		Summary:   summary,
		Data:      map[string]interface{}{"instances": instances, "rules": rules},
	}, nil
}

func (c *ContextCollector) dbEngineMetricsContext(ctx context.Context, contextType vo.ContextType, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	pattern := fmt.Sprintf("db.%s%%", dbEngineFromType(contextType))
	rows, err := c.chQuery(ctx, `
		SELECT metric_name,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val
		FROM {db}.metrics_5m
		WHERE five_minutes >= @from AND five_minutes <= @to
		  AND organization_id = @orgId AND metric_name LIKE @pattern
		GROUP BY metric_name
		ORDER BY max_val DESC LIMIT @limit`,
		map[string]interface{}{
			"from":    fmtCH(timeRange.From),
			"to":      fmtCH(timeRange.To),
			"orgId":   orgID,
			"pattern": pattern,
			"limit":   limit,
		},
	)
	if err != nil {
		return c.unavailable(contextType, timeRange, err.Error()), nil
	}

	summary := fmt.Sprintf("%s metrics: %d series.", contextType, len(rows))
	return &vo.TelemetryContext{
		Type:      contextType,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": rows},
	}, nil
}

func (c *ContextCollector) dbMySQLContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	metrics, _ := c.chQuery(ctx, `
		SELECT metric_name, metric_value, labels, timestamp
		FROM {db}.db_mysql_metrics
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	queries, _ := c.chQuery(ctx, `
		SELECT digest_text, schema_name, calls, avg_time_us, timestamp
		FROM {db}.db_mysql_queries
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY avg_time_us DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("MySQL metrics: %d, queries: %d.", len(metrics), len(queries))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringMySQL,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": metrics, "queries": queries},
	}, nil
}

func (c *ContextCollector) dbMariaDBContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	metrics, _ := c.chQuery(ctx, `
		SELECT metric_name, metric_value, labels, timestamp
		FROM {db}.db_mysql_metrics
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("MariaDB metrics: %d series.", len(metrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringMariaDB,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": metrics},
	}, nil
}

func (c *ContextCollector) dbPerconaContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	metrics, _ := c.chQuery(ctx, `
		SELECT metric_name, metric_value, labels, timestamp
		FROM {db}.db_mysql_metrics
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("Percona metrics: %d series.", len(metrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringPercona,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": metrics},
	}, nil
}

func (c *ContextCollector) dbSQLite3Context(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	metrics, _ := c.chQuery(ctx, `
		SELECT metric_name, value_float, timestamp
		FROM {db}.db_sqlite3_metrics
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("SQLite3 metrics: %d series.", len(metrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringSQLite3,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": metrics},
	}, nil
}

func (c *ContextCollector) dbTimescaleDBContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	metrics, _ := c.chQuery(ctx, `
		SELECT metric_name, metric_value, timestamp
		FROM {db}.db_timescaledb_metrics
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("TimescaleDB metrics: %d series.", len(metrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringTimescaleDB,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"metrics": metrics},
	}, nil
}

func (c *ContextCollector) dbAuroraContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	pgTopo, _ := c.pgQuery(ctx, `
		SELECT id, cluster_identifier, engine_type, engine_version, cluster_status,
		       writer_instance_id, reader_instance_ids
		FROM aurora_cluster_topology
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	chMetrics, _ := c.chQuery(ctx, `
		SELECT cluster_identifier, metric_name,
		       round(avgMerge(avg_value), 4) AS avg_val,
		       round(maxMerge(max_value), 4) AS max_val
		FROM {db}.db_aurora_metrics_1m
		WHERE timestamp_1m >= @from AND organization_id = @orgId
		GROUP BY cluster_identifier, metric_name
		ORDER BY max_val DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("Aurora: %d clusters, %d metric series.", len(pgTopo), len(chMetrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringAurora,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"topology": pgTopo, "metrics": chMetrics},
	}, nil
}

func (c *ContextCollector) dbMSSQLContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	pgInstances, _ := c.pgQuery(ctx, `
		SELECT id, name, host, port, edition, product_version, engine_edition, status, last_seen_at
		FROM mssql_instances
		WHERE organization_id = $1 AND deleted_at IS NULL LIMIT $2`, orgID, limit)

	chMetrics, _ := c.chQuery(ctx, `
		SELECT metric_name, value, timestamp
		FROM {db}.db_mssql_metrics
		WHERE timestamp >= @from AND organization_id = @orgId
		ORDER BY timestamp DESC LIMIT @limit`,
		map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit},
	)

	summary := fmt.Sprintf("MSSQL: %d instances, %d metrics.", len(pgInstances), len(chMetrics))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringMSSQL,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"instances": pgInstances, "metrics": chMetrics},
	}, nil
}

func (c *ContextCollector) dbMonitoringQANContext(ctx context.Context, orgID string, timeRange vo.TimeRange, limit int) (*vo.TelemetryContext, error) {
	engines := []string{"mysql", "postgresql", "clickhouse", "mongodb"}
	var allQueries []map[string]interface{}

	for _, engine := range engines {
		rows, _ := c.chQuery(ctx, fmt.Sprintf(`
			SELECT fingerprint, avg_duration_ms, calls, timestamp
			FROM {db}.db_%s_queries
			WHERE timestamp >= @from AND organization_id = @orgId
			ORDER BY avg_duration_ms DESC LIMIT @limit`, engine),
			map[string]interface{}{"from": fmtCH(timeRange.From), "orgId": orgID, "limit": limit / len(engines)},
		)
		for _, r := range rows {
			r["engine"] = engine
			allQueries = append(allQueries, r)
		}
	}

	summary := fmt.Sprintf("QAN: %d slow queries across %d engines.", len(allQueries), len(engines))
	return &vo.TelemetryContext{
		Type:      vo.ContextDBMonitoringQAN,
		TimeRange: timeRange,
		Summary:   summary,
		Data:      map[string]interface{}{"queries": allQueries},
	}, nil
}

func buildHighlights(rows []map[string]interface{}, dataType string, timeRange vo.TimeRange) []string {
	var highlights []string
	if len(rows) == 0 {
		highlights = append(highlights, fmt.Sprintf("No %s data available for the selected time range.", dataType))
	} else {
		highlights = append(highlights, fmt.Sprintf("%d %s series found.", len(rows), dataType))
	}
	return highlights
}

func dbEngineFromType(ct vo.ContextType) string {
	m := map[vo.ContextType]string{
		vo.ContextDBMonitoringPostgreSQL:       "postgresql",
		vo.ContextDBMonitoringMongoDBCommunity: "mongodb",
		vo.ContextDBMonitoringMongoDBAtlas:     "mongodb",
		vo.ContextDBMonitoringAWSRDSMySQL:      "aws_rds_mysql",
		vo.ContextDBMonitoringAWSRDSAurora:     "aws_rds_aurora",
		vo.ContextDBMonitoringAWSDynamoDB:      "aws_dynamodb",
		vo.ContextDBMonitoringCockroachDB:      "cockroachdb",
	}
	if e, ok := m[ct]; ok {
		return e
	}
	return strings.TrimPrefix(string(ct), "db-monitoring-")
}
