# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o fambot cmd/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata

# Create app user
RUN addgroup -g 1001 -S fambot && \
    adduser -u 1001 -S fambot -G fambot

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/fambot .

# Create data directory for database
RUN mkdir -p /app/data && chown -R fambot:fambot /app

# Switch to non-root user
USER fambot

# Set environment variables
ENV DATABASE_PATH=/app/data/fambot.db
ENV PEOPLE_CHANNEL=people
ENV DEBUG=false

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep fambot || exit 1

# Run the application
ENTRYPOINT ["./fambot"]
