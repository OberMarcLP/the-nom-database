FROM golang:1.27-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o /server ./cmd/server

# Final stage - minimal image
FROM alpine:3.24

WORKDIR /app

# Install ca-certificates for HTTPS requests and create non-root user
RUN apk --no-cache add ca-certificates wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app/uploads/menu_photos && \
    chown -R appuser:appuser /app

# Copy binary from builder
COPY --from=builder /server .

# Copy swagger documentation
COPY docs/ ./docs/

# Copy database migrations
COPY --from=builder /app/db/migrations_new ./db/migrations_new

# Fix permissions for all copied files
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

EXPOSE 8080

# Add healthcheck (use -O- to output to stdout and avoid 404 on HEAD)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 -O- http://localhost:8080/api/health > /dev/null 2>&1 || exit 1

CMD ["./server"]
