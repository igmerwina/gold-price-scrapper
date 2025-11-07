# Multi-stage build untuk Go application
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git chromium

# Set working directory
WORKDIR /app

# Copy go mod files
COPY scheduler/go.mod scheduler/go.sum ./
RUN go mod download

# Copy source code
COPY scheduler/ ./
COPY scrapper/ ../scrapper/
COPY sql/ ../sql/

# Build all binaries
# Build scraper from scrapper directory
WORKDIR /scrapper
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper scrapper.go

# Build execute_sql
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o execute_sql execute_sql.go

# Final stage
FROM alpine:latest

# Install Chrome, cron, and dependencies
RUN apk --no-cache add \
    chromium \
    chromium-chromedriver \
    ca-certificates \
    tzdata \
    bash \
    dcron

# Set timezone
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Set Chrome path
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CRON_SCHEDULE="10 8 * * *" \
    RUN_ON_STARTUP=true

WORKDIR /app

# Copy built binaries from builder
COPY --from=builder /scrapper/scraper ./scraper
COPY --from=builder /app/execute_sql ./execute_sql
COPY --from=builder /app/execute_sql.go ./execute_sql.go

# Copy scripts
COPY run_scraper_docker.sh ./run_scraper.sh
COPY scheduler/.env.example ./.env.example

# Copy entrypoint
COPY docker-entrypoint.sh /docker-entrypoint.sh

# Create necessary directories
RUN mkdir -p logs sql

# Make scripts executable
RUN chmod +x run_scraper.sh /docker-entrypoint.sh scraper execute_sql

# Expose port (optional, for health check)
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
