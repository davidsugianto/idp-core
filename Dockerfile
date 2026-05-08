# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /idp-core ./cmd/http

# =============================================================================
# Development stage (for docker-compose)
# =============================================================================
FROM golang:1.25-alpine AS development

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Install air for hot-reload (air-verse fork supports Go 1.25)
RUN go install github.com/air-verse/air@latest

# Copy go mod files first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

EXPOSE 8989

CMD ["air", "-c", ".air.toml"]

# =============================================================================
# Production stage
# =============================================================================
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user first
RUN adduser -D -g '' -u 1000 appuser

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /idp-core .

# Switch to non-root user
USER appuser

EXPOSE 8989

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8989/health || exit 1

ENTRYPOINT ["./idp-core"]
