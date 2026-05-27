<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-light.svg">
    <img src="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-light.svg" alt="TelemetryFlow Logo" width="80%">
  </picture>

  <h3>TelemetryFlow GO MCP Server (TFO-GO-MCP)</h3>

[![Version](https://img.shields.io/badge/Version-1.2.0-orange.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org/)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-purple?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTIgMkM2LjQ4IDIgMiA2LjQ4IDIgMTJzNC40OCAxMCAxMCAxMCAxMC00LjQ4IDEwLTEwUzE3LjUyIDIgMTIgMnoiIGZpbGw9IiNmZmYiLz48L3N2Zz4=)](https://modelcontextprotocol.io/)
[![LLM Providers](https://img.shields.io/badge/LLM-11_Providers%20%7C%20100%2B_Models-E1BEE7)](https://telemetryflow.id)
[![OTEL SDK](https://img.shields.io/badge/OpenTelemetry_SDK-1.43.0-blueviolet)](https://opentelemetry.io/)
[![MCP Go](https://img.shields.io/badge/mcp--go-v0.54.1-9cf)](https://github.com/mark3labs/mcp-go)
[![Architecture](https://img.shields.io/badge/Architecture-DDD%2FCQRS-success)](docs/ARCHITECTURE.md)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql)](https://www.postgresql.org/)
[![ClickHouse](https://img.shields.io/badge/ClickHouse-23+-FFCC00?logo=clickhouse)](https://clickhouse.com/)

</div>

---

# Changelog

All notable changes to TelemetryFlow GO MCP Server (TFO-GO-MCP) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-05-28

### Added

- **ContextCollector service** (`internal/application/services/context_collector.go`) — real-time telemetry context collection from ClickHouse + PostgreSQL for AI analysis, ported from TFO Platform
  - 70+ context types across 7 categories: Core Telemetry, Infrastructure, Kubernetes, AI Intelligence, DB Monitoring (15 engines), Platform Management, Account
  - ClickHouse queries against materialized views: `metrics_5m`, `logs_1h`, `service_latency_percentiles_1h`, `service_error_rates_1h`, `exemplars`, `signal_correlations_1h`, `vm_metrics_1h`, `kubernetes_metrics_1h`, `anomaly_events`, `predictions`, `cost_data_daily_mv`, `db_*_metrics`, `db_*_queries`, and more
  - PostgreSQL queries against platform tables: `agents`, `alert_rules`, `users`, `roles`, `organizations`, `workspaces`, `kubernetes_clusters`, `service_map_services`, `database_instances`, `retention_policies`, `api_keys`, and more
  - Hybrid PG+CH queries for kubernetes, agents, service-map, network-map, aurora, mssql
  - 5-second timeout with graceful fallback to unavailable context
- **PromptBuilder service** (`internal/application/services/prompt_builder.go`) — context-aware system prompt generation, ported from TFO Platform
  - 58+ specialized system prompts for all context types (metrics analyst, log analyst, tracing analyst, Kubernetes expert, DB-specific DBA prompts for ClickHouse/MariaDB/MySQL/Percona/TimescaleDB/SQLite3/Aurora/MSSQL/PostgreSQL/MongoDB/DynamoDB/CockroachDB, anomaly detection, cost optimization, etc.)
  - `BuildSystemPrompt()` with IMPORTANT INSTRUCTIONS block and optional custom instructions
  - `BuildContextPrompt()` with 10KB data truncation
  - `BuildInsightPrompt()` for chronology, prediction, recommendation, root-cause, and pattern analysis
- **3 new MCP tools** for telemetry context access:
  - `collect_telemetry_context` — collect live telemetry data from ClickHouse/PostgreSQL for AI analysis (requires DB connection)
  - `list_context_types` — list all 70+ available context types organized by category
  - `build_system_prompt` — generate context-aware system prompts for a given context type
- **Telemetry context value objects** (`internal/domain/valueobjects/telemetry_context.go`):
  - `ContextType` enum with 70+ types and `IsValid()` validation
  - `TelemetryContext`, `TimeRange`, `CollectContextOptions` structs
  - `AllContextTypes()` and `DefaultTimeRange()` helpers
- **DBProvider interface** (`internal/application/services/db_provider.go`) — abstraction layer for dual PostgreSQL + ClickHouse access
- GORM adapter repositories implementing domain interfaces: `GormSessionRepository`, `GormConversationRepository`, `GormToolRepository`
- `RestoreSession()` and `RestoreConversation()` factory functions in domain aggregates for DB rehydration
- Database & ClickHouse config fields in `config.Config` struct with env var bindings (`TELEMETRYFLOW_MCP_POSTGRES_*`, `TELEMETRYFLOW_MCP_CLICKHOUSE_*`)
- Conditional DB initialization in `main.go` — PostgreSQL repos activate when `database.enabled=true` or `TELEMETRYFLOW_MCP_POSTGRES_URL` is set
- Auto-migration on startup when `database.auto_migrate=true`
- LLM Model catalog aligned with TFO Platform (11 providers, 100+ models):
  - Anthropic Claude (Opus 4.7, Sonnet 4.6, Haiku 4.5, Mythos Preview, etc.)
  - Google Gemini (3.5 Flash, 2.5 Pro/Flash, etc.)
  - OpenAI (GPT-5.5 Pro, GPT-5.4, o3, etc.)
  - DeepSeek (V4 Pro/Flash, R1/R2 Reasoner, etc.)
  - Alibaba Qwen (3.6 Max/Plus/Flash, etc.)
  - Mistral AI (Medium 3.5, Large 3, Codestral, etc.)
  - xAI Grok (4.3, 4.20 Multi-Agent/Reasoning, etc.)
  - Kimi / Moonshot (K2.6, K2.5, Moonshot V1, etc.)
  - Zhipu GLM (5.1, 5 Turbo, 4.7, etc.)
  - Xiaomi MiMo (V2.5 Pro, V2 Pro, etc.)
- 31 new unit tests for ContextCollector, PromptBuilder, and telemetry context value objects
- **Docker Compose v2** aligned with TFO Go-SDK v1.2.0 reference:
  - 3 profiles: `dev` (MCP + infrastructure), `full` (+ tfo-collector, tfo-agent, prometheus), `platform` (+ tfo-backend, tfo-viz)
  - Network: `telemetryflow_mcp_net` on subnet `172.153.0.0/16` with static IPs
  - PostgreSQL 16-alpine (`tfo_admin`/`telemetryflow_mcp`)
  - ClickHouse (`tfo_admin`/`telemetryflow_analytics`)
  - Redis 7-alpine with LRU eviction and AOF persistence
  - NATS 2-alpine with JetStream enabled
   - TFO-Collector v1.2.1 (OCB-native with tfoexporter, OTEL v0.152.1)
   - TFO-Agent v1.2.0 (OTEL SDK v1.43.0, system metrics + heartbeat)
   - Prometheus v2.54.1
   - TFO-Platform v1.4.0 (backend + frontend, TFO-Core NestJS IAM)
  - All services use named volumes with `tfo-mcp-` prefix
- **Config files** for infrastructure services:
  - `configs/tfo-collector.yaml` — TFO Collector with tfootlp receiver, tfo exporter, span_metrics/service_graph connectors
  - `configs/tfo-agent.yaml` — Lightweight agent for MCP dev environment
  - `configs/prometheus.yaml` — Scrapes tfo-collector, tfo-mcp, and self
- **Updated init scripts**:
  - `scripts/init-db.sql` — Fixed version to 1.2.0, default model to `claude-opus-4-7`
  - `scripts/init-clickhouse.sql` — Database renamed to `telemetryflow_analytics`, updated sample model names
- **Updated `.env.example`** — 20 sections aligned with docker-compose profiles, aligned credentials, port mappings, static IPs

### Changed

- Default LLM model changed from `claude-sonnet-4-20250514` to `claude-opus-4-7` (TFO Platform default)
- `Model` value object now validates against full TFO Platform model catalog (100+ models across 11 providers) instead of 9 Anthropic-only models
- `claude_conversation` tool schema enum expanded to cover representative models from all providers
- Built-in tool count increased from 8 to 11 (added `collect_telemetry_context`, `list_context_types`, `build_system_prompt`)
- Bumped Go toolchain to 1.26.3 (`go.mod` version, `toolchain` directive)
- Adopted `github.com/mark3labs/mcp-go v0.54.1` as the official MCP Go SDK (replacing custom `pkg/mcp/` protocol implementation)
- Bumped OpenTelemetry SDK to `v1.43.0` — patches GO-2026-4985 and GO-2026-4394
- Bumped TelemetryFlow Go-SDK to `v1.2.0`
- Updated CI/CD workflows: `actions/checkout@v6`, `actions/setup-go@v6`, `actions/upload-artifact@v8`, `golangci-lint-action@v9`, `GO_VERSION: '1.26'`
- Updated `Dockerfile` — hardened to match TFO Go-SDK reference: Alpine 3.23, UID/GID 10001, apk cache cleanup, build verification step
- Bumped default version string to `1.2.0`
- Updated ecosystem version references: TFO-Core v1.4.0, TFO-Collector v1.2.1 (OTEL v0.152.1), TFO-Agent v1.2.0 (SDK v1.43.0), TFO-Python-SDK v1.2.0 (SDK v1.42.1), TFO-Python-MCP v1.2.0 (SDK v1.42.1)

### Fixed

- Patched OpenTelemetry vulnerability GO-2026-4985
- Patched OpenTelemetry vulnerability GO-2026-4394
- Fixed `gopkg.in/natefinch/lumberjack.v2` version typo in `go.mod`
- Fixed stale model references in e2e and integration tests (`ModelClaude4Sonnet` → `ModelClaudeSonnet4`, `ModelClaude4Opus` → `ModelClaudeOpus47`, `ModelClaude35Sonnet` → `ModelClaudeSonnet46`)
- `govulncheck ./...` now reports **0 vulnerabilities**

---

## [1.1.2] - 2025-01-09

### Added

#### Database Infrastructure

- PostgreSQL database support with GORM ORM
  - Session repository for persistent session storage
  - Conversation repository for message history
  - Database models with JSONB support for flexible data storage
  - Connection pooling and health checks
  - Custom GORM types: `JSONB`, `JSONBArray`, `StringArray`
- ClickHouse analytics database support
  - Tool call analytics table with MergeTree engine
  - API request analytics table for Claude API usage
  - Session analytics table for lifecycle events
  - Error analytics table for debugging
  - Time series aggregation queries with percentiles (p50/p95/p99)
  - Hourly aggregation tables (SummingMergeTree, ReplacingMergeTree)
  - Materialized views for real-time aggregations
  - Batch insert support for high throughput

#### Analytics & Dashboard

- Analytics repository with comprehensive dashboard queries
  - Token usage statistics by model
  - Tool usage statistics with p50/p95/p99 percentiles
  - Request/token/latency time series
  - Error rate tracking over time
  - Dashboard summary with key metrics

#### Migration & Seeding Infrastructure

- Database migration infrastructure
  - Versioned migrations for PostgreSQL and ClickHouse
  - Migration runner with up/down/reset/fresh operations
  - Migration status tracking via `schema_migrations` table
  - GORM AutoMigrate integration
- Database seeding infrastructure
  - Idempotent seeders using FirstOrCreate pattern
  - Default seeders: tools, resources, prompts, api_keys, demo_session
  - Production-safe seeder subset (excludes demo data)
  - SeederResult tracking for executed/skipped/failed seeders

#### Caching & Queue Infrastructure

- Redis caching infrastructure
  - Cache service with TTL support
  - Session and conversation caching
  - Cache invalidation strategies
- NATS JetStream queue infrastructure
  - Durable message queuing
  - Publisher/subscriber pattern
  - Task acknowledgment with retry support

#### Testing & CI/CD

- Comprehensive test suite
  - Session aggregate tests
  - Conversation aggregate tests
  - Tool entity tests
  - MCP protocol tests
  - Migration and seeder tests
  - GORM model tests
  - Benchmarks for performance testing
- CI-specific Makefile targets for GitHub Actions
  - `test-unit-ci`: Unit tests with coverage output
  - `test-integration-ci`: Integration tests with coverage
  - `test-e2e-ci`: End-to-end tests
  - `ci-build`: Cross-platform CI builds with GOOS/GOARCH
  - `ci-test`: Combined format, vet, lint, test pipeline
  - `test-all`: Run all test types
  - `deps-verify`: Dependency verification
  - `staticcheck`: Static analysis
  - `govulncheck`: Vulnerability scanning
  - `coverage-report`: Merged coverage report generation
- GitHub Actions CI/CD workflows
  - Lint, test, build pipeline
  - Multi-platform build support (Linux, macOS, Windows)
  - Security scanning with Gosec and govulncheck
  - Coverage reporting with merged reports

### Changed

- **BREAKING**: Refactored all environment variable keys from `TFO_*` to `TELEMETRYFLOW_*`
  - `TFO_CLAUDE_API_KEY` → `TELEMETRYFLOW_MCP_CLAUDE_API_KEY`
  - `TFO_LOG_LEVEL` → `TELEMETRYFLOW_MCP_LOG_LEVEL`
  - `TFO_SERVER_HOST` → `TELEMETRYFLOW_MCP_SERVER_HOST`
  - `TFO_SERVER_PORT` → `TELEMETRYFLOW_MCP_SERVER_PORT`
  - All other `TFO_*` variables follow the same pattern
- Updated go.mod with GORM, ClickHouse, Redis, and NATS dependencies
- Enhanced configuration with database, cache, and queue settings
- Updated Viper SetEnvPrefix from `TFO_MCP` to `TELEMETRYFLOW_MCP`
- Improved error handling in analytics repository with proper resource cleanup
- Enhanced event publishing with explicit error ignoring for best-effort delivery

### Fixed

- Fixed all golangci-lint errors (0 issues)
  - Added proper error handling for `rows.Close()` in analytics repository
  - Added explicit error ignoring for event publisher calls (best-effort delivery)
  - Fixed empty branches in test assertions
  - Removed unused variables in server tests
- Fixed anthropic SDK API compatibility in client.go
- Fixed observability.go StartSpan return type

---

## [1.1.1] - 2025-01-05

### Added

- Initial project structure
- Basic MCP protocol support
- Claude API integration prototype

### Changed

- Refined DDD architecture

---

## [1.1.0] - 2025-01-01

### Added

- Project inception
- Architecture design
- Technology stack selection

---

## Version History Summary

| Version | Date       | Highlights                                                                |
| ------- | ---------- | ------------------------------------------------------------------------- |
| 1.2.0   | 2026-05-28 | ContextCollector, PromptBuilder, 70+ context types, 11 MCP tools, Go 1.26 |
| 1.1.2   | 2025-01-09 | Database infrastructure, CI/CD, analytics, comprehensive tests            |
| 1.1.1   | 2025-01-05 | Initial structure, basic protocol support                                 |
| 1.1.0   | 2025-01-01 | Project inception                                                         |

---

## Migration Guide

### Upgrading to 1.1.2

No breaking changes. Update your binary and restart the server.

```bash
# Download new version
curl -LO https://github.com/telemetryflow/telemetryflow-go-mcp/releases/download/v1.1.2/tfo-mcp_$(uname -s)_$(uname -m).tar.gz

# Extract and install
tar -xzf tfo-mcp_*.tar.gz
sudo mv tfo-mcp /usr/local/bin/

# Verify
tfo-mcp version
```

---

## Links

- [GitHub Repository](https://github.com/telemetryflow/telemetryflow-go-mcp)
- [Documentation](docs/README.md)
- [Issue Tracker](https://github.com/telemetryflow/telemetryflow-go-mcp/issues)

[Unreleased]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.1.2...v1.2.0
[1.1.2]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/telemetryflow/telemetryflow-go-mcp/releases/tag/v1.1.0
