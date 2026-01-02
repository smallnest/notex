# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make build-base

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Use CGO_ENABLED=0 since modernc.org/sqlite is used
RUN CGO_ENABLED=0 GOOS=linux go build -o notex main.go

# Runtime stage
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies
# Add python3 and pip for markitdown
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    python3 \
    py3-pip \
    libmagic

# Install markitdown
RUN pip install --break-system-packages markitdown

# Create data directory
RUN mkdir -p /data

# Copy binary from builder
COPY --from=builder /app/notex .

# Expose port
EXPOSE 8080

# Set environment variables
ENV STORE_PATH=/data/checkpoints.db
ENV SQLITE_PATH=/data/vector.db
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./notex", "-server"]
