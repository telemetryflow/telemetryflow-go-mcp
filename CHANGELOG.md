# Changelog

All notable changes to TelemetryFlow MCP Server (TFO-MCP) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.2] - 2025-01-09

### Added
- PostgreSQL database support with GORM ORM
  - Session repository for persistent session storage
  - Conversation repository for message history
  - Database models with JSONB support
  - Connection pooling and health checks
- ClickHouse analytics database support
  - Tool call analytics table
  - API request analytics table
  - Session analytics table
  - Time series aggregation queries
  - Batch insert support for high throughput
- Analytics repository with dashboard queries
  - Token usage by model
  - Tool usage statistics with percentiles
  - Request/token/latency time series
  - Error rate tracking
- Redis caching infrastructure
  - Cache service with TTL support
  - Session and conversation caching
  - Cache invalidation strategies
- NATS JetStream queue infrastructure
  - Durable message queuing
  - Publisher/subscriber pattern
  - Task acknowledgment with retry support
- Comprehensive test suite
  - Session aggregate tests
  - Conversation aggregate tests
  - Tool entity tests
  - MCP protocol tests
  - Benchmarks for performance testing
- Additional test coverage
- Performance benchmarks
- GitHub Actions CI/CD workflows
  - Lint, test, build pipeline
  - Multi-platform build support
  - Security scanning with Gosec and govulncheck
  - Coverage reporting

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

### Fixed
- Minor bug fixes
- Fixed anthropic SDK API compatibility in client.go
- Fixed observability.go StartSpan return type

---

## [1.1.2] - 2026-01-09

### Added

#### Core Features
- Complete MCP protocol implementation (version 2024-11-05)
- Claude AI integration with Anthropic SDK v0.2.0-beta.3
- Domain-Driven Design (DDD) architecture
- CQRS pattern for command/query separation

#### Domain Layer
- Session aggregate for MCP session lifecycle management
- Conversation aggregate for multi-turn conversations
- Value objects: SessionID, ConversationID, MessageID, ToolID, ResourceID, PromptID
- Domain events: SessionCreated, MessageAdded, ToolExecuted, etc.
- Repository interfaces for persistence abstraction

#### Application Layer
- Session commands: InitializeSession, CloseSession
- Conversation commands: CreateConversation, SendMessage
- Tool commands: ExecuteTool, RegisterTool
- Query handlers for tools, resources, and prompts

#### Infrastructure Layer
- Claude API client with retry logic and streaming support
- In-memory repositories for session, conversation, tool management
- Configuration management with Viper
- Structured logging with Zerolog

#### Presentation Layer
- MCP server with JSON-RPC 2.0 transport
- Built-in tools: claude_conversation, read_file, write_file, list_directory
- Additional tools: execute_command, search_files, system_info, echo
- Resource and prompt endpoints

#### Observability
- OpenTelemetry SDK v1.39.0 integration
- Distributed tracing support
- Metrics collection
- Structured logging

#### Developer Experience
- Comprehensive Makefile with 30+ targets
- Docker support with multi-stage builds
- Cross-platform builds (Linux, macOS, Windows)
- Development scripts (build, test, release, validate)

#### Documentation
- README with Mermaid diagrams
- Architecture documentation
- Configuration guide
- Commands reference
- Development guide
- Installation guide
- Troubleshooting guide

### Technical Details

- **Go Version**: 1.24+
- **MCP Protocol**: 2024-11-05
- **Claude API**: anthropic-sdk-go v0.2.0-beta.3
- **OpenTelemetry**: v1.39.0
- **Logging**: Zerolog v1.33.0
- **CLI**: Cobra v1.8.1
- **Config**: Viper v1.19.0

---

## [1.1.1] - 2026-01-05

### Added
- Initial project structure
- Basic MCP protocol support
- Claude API integration prototype

### Changed
- Refined DDD architecture

---

## [1.1.0] - 2026-01-01

### Added
- Project inception
- Architecture design
- Technology stack selection

---

## Version History Summary

| Version | Date | Highlights |
|---------|------|------------|
| 1.1.2 | 2026-01-09 | Full MCP implementation, comprehensive documentation |
| 1.1.1 | 2026-01-05 | Initial structure, basic protocol support |
| 1.1.0 | 2026-01-01 | Project inception |

---

## Migration Guide

### Upgrading to 1.1.2

No breaking changes. Update your binary and restart the server.

```bash
# Download new version
curl -LO https://github.com/devopscorner/telemetryflow/releases/download/v1.1.2/tfo-mcp_$(uname -s)_$(uname -m).tar.gz

# Extract and install
tar -xzf tfo-mcp_*.tar.gz
sudo mv tfo-mcp /usr/local/bin/

# Verify
tfo-mcp version
```

---

## Links

- [GitHub Repository](https://github.com/devopscorner/telemetryflow)
- [Documentation](docs/README.md)
- [Issue Tracker](https://github.com/devopscorner/telemetryflow/issues)

[Unreleased]: https://github.com/devopscorner/telemetryflow/compare/v1.1.2...HEAD
[1.1.2]: https://github.com/devopscorner/telemetryflow/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/devopscorner/telemetryflow/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/devopscorner/telemetryflow/releases/tag/v1.1.0
