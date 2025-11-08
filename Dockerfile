# Multi-stage build untuk Go application
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git chromium

WORKDIR /app

COPY scrapper/go.mod scrapper/go.sum ./scrapper/
COPY scheduler/go.mod scheduler/go.sum ./scheduler/

WORKDIR /app/scrapper
RUN go mod download

WORKDIR /app/scheduler
RUN go mod download

WORKDIR /app
COPY scrapper/ ./scrapper/
COPY scheduler/ ./scheduler/

WORKDIR /app/scrapper
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper scrapper.go

WORKDIR /app/scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o execute_sql execute_sql.go

FROM alpine:latest

RUN apk add --no-cache \
    bash \
    ca-certificates \
    chromium \
    chromium-chromedriver \
    dcron \
    tzdata

ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/"$TZ" /etc/localtime && echo "$TZ" > /etc/timezone

ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CRON_SCHEDULE="10 8 * * *" \
    RUN_ON_STARTUP=true

WORKDIR /app

COPY --from=builder /app/scrapper/scraper ./scraper
COPY --from=builder /app/scheduler/execute_sql ./execute_sql
COPY run_scraper_docker.sh ./run_scraper.sh
COPY test_cron.sh ./test_cron.sh
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN mkdir -p logs sql && \
    chmod +x run_scraper.sh test_cron.sh /docker-entrypoint.sh scraper execute_sql

ENTRYPOINT ["/docker-entrypoint.sh"]
