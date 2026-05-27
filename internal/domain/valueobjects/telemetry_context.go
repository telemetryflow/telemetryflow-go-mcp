package valueobjects

import "time"

type ContextType string

const (
	ContextMetrics                      ContextType = "metrics"
	ContextLogs                         ContextType = "logs"
	ContextTraces                       ContextType = "traces"
	ContextExemplars                    ContextType = "exemplars"
	ContextCorrelations                 ContextType = "correlations"
	ContextAlerts                       ContextType = "alerts"
	ContextAlertRules                   ContextType = "alert-rules"
	ContextKubernetesOverview           ContextType = "kubernetes-overview"
	ContextKubernetesClusters           ContextType = "kubernetes-clusters"
	ContextKubernetesNamespaces         ContextType = "kubernetes-namespaces"
	ContextKubernetesNodes              ContextType = "kubernetes-nodes"
	ContextKubernetesPods               ContextType = "kubernetes-pods"
	ContextKubernetesDeployments        ContextType = "kubernetes-deployments"
	ContextKubernetesPV                 ContextType = "kubernetes-pv"
	ContextKubernetesAPIServer          ContextType = "kubernetes-api-server"
	ContextKubernetesCoreDNS            ContextType = "kubernetes-coredns"
	ContextAgents                       ContextType = "agents"
	ContextUptime                       ContextType = "uptime"
	ContextStatusPage                   ContextType = "status-page"
	ContextInfraOverview                ContextType = "infra-overview"
	ContextInfraCPU                     ContextType = "infra-cpu"
	ContextInfraMemory                  ContextType = "infra-memory"
	ContextInfraStorage                 ContextType = "infra-storage"
	ContextInfraNetwork                 ContextType = "infra-network"
	ContextServiceMap                   ContextType = "service-map"
	ContextNetworkMap                   ContextType = "network-map"
	ContextDashboard                    ContextType = "dashboard"
	ContextReports                      ContextType = "reports"
	ContextIAM                          ContextType = "iam"
	ContextIAMUsers                     ContextType = "iam-users"
	ContextIAMRoles                     ContextType = "iam-roles"
	ContextIAMPermissions               ContextType = "iam-permissions"
	ContextIAMMatrix                    ContextType = "iam-matrix"
	ContextIAMAssignments               ContextType = "iam-assignments"
	ContextTenancy                      ContextType = "tenancy"
	ContextTenancyRegions               ContextType = "tenancy-regions"
	ContextTenancyOrganizations         ContextType = "tenancy-organizations"
	ContextTenancyWorkspaces            ContextType = "tenancy-workspaces"
	ContextTenancyTenants               ContextType = "tenancy-tenants"
	ContextAudit                        ContextType = "audit"
	ContextRetention                    ContextType = "retention"
	ContextSubscription                 ContextType = "subscription"
	ContextAPIKeys                      ContextType = "api-keys"
	ContextNotifications                ContextType = "notifications"
	ContextDataMasking                  ContextType = "data-masking"
	ContextSystemSetup                  ContextType = "system-setup"
	ContextSystemChannels               ContextType = "system-channels"
	ContextAIAssistant                  ContextType = "ai-assistant"
	ContextAccountProfile               ContextType = "account-profile"
	ContextAccountSecurity              ContextType = "account-security"
	ContextAccountSessions              ContextType = "account-sessions"
	ContextAccountNotifications         ContextType = "account-notifications"
	ContextAccountPreferences           ContextType = "account-preferences"
	ContextAccountOrganization          ContextType = "account-organization"
	ContextAnomalyDetection             ContextType = "anomaly-detection"
	ContextCorrectiveMaintenance        ContextType = "corrective-maintenance"
	ContextPredictiveMaintenance        ContextType = "predictive-maintenance"
	ContextCostOptimization             ContextType = "cost-optimization"
	ContextDBMonitoringInventory        ContextType = "db-monitoring-inventory"
	ContextDBMonitoringClickHouse       ContextType = "db-monitoring-clickhouse"
	ContextDBMonitoringMariaDB          ContextType = "db-monitoring-mariadb"
	ContextDBMonitoringMySQL            ContextType = "db-monitoring-mysql"
	ContextDBMonitoringPercona          ContextType = "db-monitoring-percona"
	ContextDBMonitoringSQLite3          ContextType = "db-monitoring-sqlite3"
	ContextDBMonitoringTimescaleDB      ContextType = "db-monitoring-timescaledb"
	ContextDBMonitoringAurora           ContextType = "db-monitoring-aurora"
	ContextDBMonitoringMSSQL            ContextType = "db-monitoring-mssql"
	ContextDBMonitoringPostgreSQL       ContextType = "db-monitoring-postgresql"
	ContextDBMonitoringMongoDBCommunity ContextType = "db-monitoring-mongodb-community"
	ContextDBMonitoringMongoDBAtlas     ContextType = "db-monitoring-mongodb-atlas"
	ContextDBMonitoringAWSRDSMySQL      ContextType = "db-monitoring-aws-rds-mysql"
	ContextDBMonitoringAWSRDSAurora     ContextType = "db-monitoring-aws-rds-aurora"
	ContextDBMonitoringAWSDynamoDB      ContextType = "db-monitoring-aws-dynamodb"
	ContextDBMonitoringCockroachDB      ContextType = "db-monitoring-cockroachdb"
	ContextDBMonitoringQAN              ContextType = "db-monitoring-qan"
)

func (c ContextType) String() string {
	return string(c)
}

func (c ContextType) IsValid() bool {
	switch c {
	case ContextMetrics, ContextLogs, ContextTraces, ContextExemplars, ContextCorrelations,
		ContextAlerts, ContextAlertRules,
		ContextKubernetesOverview, ContextKubernetesClusters, ContextKubernetesNamespaces,
		ContextKubernetesNodes, ContextKubernetesPods, ContextKubernetesDeployments,
		ContextKubernetesPV, ContextKubernetesAPIServer, ContextKubernetesCoreDNS,
		ContextAgents, ContextUptime, ContextStatusPage,
		ContextInfraOverview, ContextInfraCPU, ContextInfraMemory, ContextInfraStorage, ContextInfraNetwork,
		ContextServiceMap, ContextNetworkMap,
		ContextDashboard, ContextReports,
		ContextIAM, ContextIAMUsers, ContextIAMRoles, ContextIAMPermissions, ContextIAMMatrix, ContextIAMAssignments,
		ContextTenancy, ContextTenancyRegions, ContextTenancyOrganizations, ContextTenancyWorkspaces, ContextTenancyTenants,
		ContextAudit, ContextRetention, ContextSubscription, ContextAPIKeys,
		ContextNotifications, ContextDataMasking,
		ContextSystemSetup, ContextSystemChannels, ContextAIAssistant,
		ContextAccountProfile, ContextAccountSecurity, ContextAccountSessions,
		ContextAccountNotifications, ContextAccountPreferences, ContextAccountOrganization,
		ContextAnomalyDetection, ContextCorrectiveMaintenance, ContextPredictiveMaintenance, ContextCostOptimization,
		ContextDBMonitoringInventory, ContextDBMonitoringClickHouse, ContextDBMonitoringMariaDB,
		ContextDBMonitoringMySQL, ContextDBMonitoringPercona, ContextDBMonitoringSQLite3,
		ContextDBMonitoringTimescaleDB, ContextDBMonitoringAurora, ContextDBMonitoringMSSQL,
		ContextDBMonitoringPostgreSQL, ContextDBMonitoringMongoDBCommunity, ContextDBMonitoringMongoDBAtlas,
		ContextDBMonitoringAWSRDSMySQL, ContextDBMonitoringAWSRDSAurora, ContextDBMonitoringAWSDynamoDB,
		ContextDBMonitoringCockroachDB, ContextDBMonitoringQAN:
		return true
	}
	return false
}

type TimeRange struct {
	From time.Time
	To   time.Time
}

type TelemetryContext struct {
	Type      ContextType `json:"type"`
	TimeRange TimeRange   `json:"timeRange"`
	Summary   string      `json:"summary"`
	Data      interface{} `json:"data"`
}

type CollectContextOptions struct {
	OrganizationID string
	UserID         string
	ContextType    ContextType
	ContextID      string
	TimeRange      *TimeRange
	MaxItems       int
}

func DefaultTimeRange() TimeRange {
	return TimeRange{
		From: time.Now().UTC().Add(-time.Hour),
		To:   time.Now().UTC(),
	}
}

func AllContextTypes() []ContextType {
	return []ContextType{
		ContextMetrics, ContextLogs, ContextTraces, ContextExemplars, ContextCorrelations,
		ContextAlerts, ContextAlertRules,
		ContextKubernetesOverview, ContextKubernetesClusters, ContextKubernetesNamespaces,
		ContextKubernetesNodes, ContextKubernetesPods, ContextKubernetesDeployments,
		ContextKubernetesPV, ContextKubernetesAPIServer, ContextKubernetesCoreDNS,
		ContextAgents, ContextUptime, ContextStatusPage,
		ContextInfraOverview, ContextInfraCPU, ContextInfraMemory, ContextInfraStorage, ContextInfraNetwork,
		ContextServiceMap, ContextNetworkMap,
		ContextDashboard, ContextReports,
		ContextIAM, ContextIAMUsers, ContextIAMRoles, ContextIAMPermissions, ContextIAMMatrix, ContextIAMAssignments,
		ContextTenancy, ContextTenancyRegions, ContextTenancyOrganizations, ContextTenancyWorkspaces, ContextTenancyTenants,
		ContextAudit, ContextRetention, ContextSubscription, ContextAPIKeys,
		ContextNotifications, ContextDataMasking,
		ContextSystemSetup, ContextSystemChannels, ContextAIAssistant,
		ContextAccountProfile, ContextAccountSecurity, ContextAccountSessions,
		ContextAccountNotifications, ContextAccountPreferences, ContextAccountOrganization,
		ContextAnomalyDetection, ContextCorrectiveMaintenance, ContextPredictiveMaintenance, ContextCostOptimization,
		ContextDBMonitoringInventory, ContextDBMonitoringClickHouse, ContextDBMonitoringMariaDB,
		ContextDBMonitoringMySQL, ContextDBMonitoringPercona, ContextDBMonitoringSQLite3,
		ContextDBMonitoringTimescaleDB, ContextDBMonitoringAurora, ContextDBMonitoringMSSQL,
		ContextDBMonitoringPostgreSQL, ContextDBMonitoringMongoDBCommunity, ContextDBMonitoringMongoDBAtlas,
		ContextDBMonitoringAWSRDSMySQL, ContextDBMonitoringAWSRDSAurora, ContextDBMonitoringAWSDynamoDB,
		ContextDBMonitoringCockroachDB, ContextDBMonitoringQAN,
	}
}
