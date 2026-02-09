# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY src/ ./src/
COPY config/ ./config/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o fusion ./src

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install Chrome for headless browser automation
RUN apk add --no-cache chromium chromium-chromedriver

# Copy binary from builder
COPY --from=builder /app/fusion /app/fusion
COPY --from=builder /app/config /app/config

# Create necessary directories
RUN mkdir -p /app/output /app/logs

# Environment variables
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/lib/chromium/

# Expose API port
EXPOSE 5039

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5039/api/dashboard || exit 1

# Run the application
CMD ["/app/fusion"]
