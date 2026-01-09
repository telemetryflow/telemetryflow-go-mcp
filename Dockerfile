# ==============================================================================
# TelemetryFlow MCP Server Dockerfile
# Multi-stage build for minimal production image
# ==============================================================================

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=1.1.2
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o tfo-mcp \
    ./cmd/mcp

# ==============================================================================
# Production stage
# ==============================================================================
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S telemetryflow && adduser -S telemetryflow -G telemetryflow

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/tfo-mcp /app/tfo-mcp

# Copy default config
COPY configs/config.yaml /app/configs/config.yaml

# Set ownership
RUN chown -R telemetryflow:telemetryflow /app

# Switch to non-root user
USER telemetryflow

# Environment variables
ENV ANTHROPIC_API_KEY=""
ENV TFO_MCP_SERVER_TRANSPORT="stdio"
ENV TFO_MCP_LOG_LEVEL="info"
ENV TFO_MCP_LOG_FORMAT="json"

# Health check (for SSE/WebSocket modes)
# HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Expose port (for SSE/WebSocket modes)
EXPOSE 8080

# Entry point
ENTRYPOINT ["/app/tfo-mcp"]

# Default command
CMD ["--config", "/app/configs/config.yaml"]
