# ==============================================================================
# TelemetryFlow GO MCP Server Dockerfile
# Multi-stage build for minimal production image
# ==============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM golang:1.26-alpine AS builder

# Build arguments
ARG VERSION=1.2.0
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Set working directory
WORKDIR /build

# Set GOPRIVATE to bypass checksum database for telemetryflow SDK
ENV GOPRIVATE=github.com/telemetryflow/*

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Download dependencies (leverages Docker layer caching)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
    -X 'main.version=${VERSION}' \
    -X 'main.commit=${COMMIT}' \
    -X 'main.buildDate=${BUILD_DATE}'" \
    -o /tfo-mcp ./cmd/mcp

# Verify binary
RUN /tfo-mcp version

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:3.23

# =============================================================================
# TelemetryFlow Metadata Labels (OCI Image Spec)
# =============================================================================
LABEL org.opencontainers.image.title="TelemetryFlow GO MCP Server" \
    org.opencontainers.image.description="Enterprise MCP (Model Context Protocol) server for AI-powered observability - Community Enterprise Observability Platform (CEOP)" \
    org.opencontainers.image.version="1.2.0" \
    org.opencontainers.image.vendor="TelemetryFlow" \
    org.opencontainers.image.authors="Telemetri Data Indonesia <support@devopscorner.id>" \
    org.opencontainers.image.url="https://telemetryflow.id" \
    org.opencontainers.image.documentation="https://docs.telemetryflow.id" \
    org.opencontainers.image.source="https://github.com/telemetryflow/telemetryflow-go-mcp" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.base.name="alpine:3.23" \
    io.telemetryflow.product="TelemetryFlow GO MCP Server" \
    io.telemetryflow.component="tfo-mcp" \
    io.telemetryflow.platform="CEOP" \
    io.telemetryflow.maintainer="Telemetri Data Indonesia"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user and group
RUN addgroup -g 10001 -S telemetryflow && \
    adduser -u 10001 -S telemetryflow -G telemetryflow -h /home/telemetryflow

# Create app directory
RUN mkdir -p /app/configs && chown -R telemetryflow:telemetryflow /app

# Copy binary from builder
COPY --from=builder /tfo-mcp /usr/local/bin/tfo-mcp
RUN chmod +x /usr/local/bin/tfo-mcp

# Copy default config
COPY --chown=telemetryflow:telemetryflow configs/tfo-mcp.yaml /app/configs/tfo-mcp.yaml

# Switch to non-root user
USER telemetryflow

# Set working directory
WORKDIR /app

# =============================================================================
# Environment Variables (synchronized with .env.example)
# =============================================================================

# Claude API
ENV ANTHROPIC_API_KEY=""

# TelemetryFlow Observability (TelemetryFlow SDK)
ENV TELEMETRYFLOW_API_KEY=""
ENV TELEMETRYFLOW_ENDPOINT="https://api.telemetryflow.io"
ENV TELEMETRYFLOW_SERVICE_NAME="telemetryflow-go-mcp"
ENV TELEMETRYFLOW_SERVICE_VERSION="1.2.0"
ENV TELEMETRYFLOW_ENVIRONMENT="production"

# Server Configuration
ENV TELEMETRYFLOW_MCP_SERVER_HOST="0.0.0.0"
ENV TELEMETRYFLOW_MCP_SERVER_PORT="8080"
ENV TELEMETRYFLOW_MCP_SERVER_TRANSPORT="stdio"
ENV TELEMETRYFLOW_MCP_DEBUG="false"

# Logging
ENV TELEMETRYFLOW_MCP_LOG_LEVEL="info"
ENV TELEMETRYFLOW_MCP_LOG_FORMAT="json"

# Redis (Caching & Queue)
ENV TELEMETRYFLOW_MCP_REDIS_URL="redis://localhost:6379"
ENV TELEMETRYFLOW_MCP_CACHE_ENABLED="true"
ENV TELEMETRYFLOW_MCP_CACHE_TTL="300"
ENV TELEMETRYFLOW_MCP_QUEUE_ENABLED="true"
ENV TELEMETRYFLOW_MCP_QUEUE_CONCURRENCY="5"

# Database (PostgreSQL)
ENV TELEMETRYFLOW_MCP_POSTGRES_URL=""
ENV TELEMETRYFLOW_MCP_POSTGRES_MAX_CONNS="25"
ENV TELEMETRYFLOW_MCP_POSTGRES_MIN_CONNS="5"

# Analytics Database (ClickHouse)
ENV TELEMETRYFLOW_MCP_CLICKHOUSE_URL=""

# OpenTelemetry (Fallback)
ENV TELEMETRYFLOW_MCP_TELEMETRY_ENABLED="true"
ENV TELEMETRYFLOW_MCP_OTLP_ENDPOINT="localhost:4317"
ENV TELEMETRYFLOW_MCP_SERVICE_NAME="telemetryflow-go-mcp"

# Health check (for SSE/WebSocket modes)
# HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Expose port (for SSE/WebSocket modes)
EXPOSE 8080

# Entry point
ENTRYPOINT ["/usr/local/bin/tfo-mcp"]

# Default command
CMD ["--config", "/app/configs/tfo-mcp.yaml"]
