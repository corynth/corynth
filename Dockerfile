# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make build-base

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o corynth ./cmd/corynth

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates git

# Create non-root user
RUN addgroup -g 1000 corynth && \
    adduser -D -u 1000 -G corynth corynth

# Create working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/corynth /usr/local/bin/corynth

# Create config directory
RUN mkdir -p /home/corynth/.corynth && \
    chown -R corynth:corynth /home/corynth

# Switch to non-root user
USER corynth

# Set home directory
ENV HOME=/home/corynth

# Expose port for API (future feature)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD corynth version || exit 1

# Default command
ENTRYPOINT ["corynth"]
CMD ["--help"]