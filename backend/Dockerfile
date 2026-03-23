# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o fusion ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install Chrome for headless browser automation, certs and timezone
RUN apk add --no-cache chromium chromium-chromedriver tzdata ca-certificates nss freetype harfbuzz font-noto-cjk wget

# Copy binary from builder
COPY --from=builder /app/fusion /app/fusion

# Create necessary directories
RUN mkdir -p /app/output /app/logs /app/config

# Environment variables
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/lib/chromium/

# Timezone
ENV TZ=Asia/Ho_Chi_Minh

# Expose both API ports
EXPOSE 5039 5040

# Health check on legacy port
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5040/healthz || exit 1

# Run the application
CMD ["/app/fusion"]
