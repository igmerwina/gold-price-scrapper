#!/bin/bash

# Script untuk menjalankan gold price scraper dan update ke Supabase
# Dijalankan oleh cron scheduler setiap hari jam 8:10 pagi

# Detect environment
if [ -f /.dockerenv ] || [ "$IS_DOCKER" = "true" ]; then
    # Running in Docker
    BASE_DIR="/app"
    IS_DOCKER_ENV=true
else
    # Running locally
    BASE_DIR="/Users/macbook/Documents/code/random/gold-scrapper"
    IS_DOCKER_ENV=false
fi

# Set working directory
cd "$BASE_DIR/scheduler" || cd "$BASE_DIR"

# Load environment variables properly (handles quoted values with spaces)
if [ -f .env ]; then
    set -a
    while IFS= read -r line; do
        # Skip comments and empty lines
        case "$line" in
            \#*|"") continue ;;
            *) export "$line" ;;
        esac
    done < .env
    set +a
elif [ -f "$BASE_DIR/.env" ]; then
    set -a
    while IFS= read -r line; do
        case "$line" in
            \#*|"") continue ;;
            *) export "$line" ;;
        esac
    done < "$BASE_DIR/.env"
    set +a
fi

# Log file
LOG_FILE="logs/scraper_$(date +%Y%m%d_%H%M%S).log"
mkdir -p logs

echo "========================================" | tee -a $LOG_FILE
echo "Gold Price Scraper - $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE

# 1. Jalankan scraper untuk ambil data terbaru
echo "" | tee -a $LOG_FILE
echo "Step 1: Scraping data harga emas..." | tee -a $LOG_FILE
echo "Environment: $([ "$IS_DOCKER_ENV" = true ] && echo "Docker" || echo "Local")" | tee -a $LOG_FILE

if [ "$IS_DOCKER_ENV" = true ]; then
    # Docker: gunakan binary yang sudah dikompilasi
    cd "$BASE_DIR"
    if [ -f "./scraper" ]; then
        ./scraper 2>&1 | tee -a $LOG_FILE
    else
        echo "ERROR: Scraper binary tidak ditemukan!" | tee -a $LOG_FILE
        exit 1
    fi
else
    # Local: gunakan go run
    cd "$BASE_DIR/scrapper"
    go run scrapper.go 2>&1 | tee -a "../scheduler/$LOG_FILE"
fi

if [ $? -ne 0 ]; then
    echo "ERROR: Scraping gagal!" | tee -a $LOG_FILE
    exit 1
fi

# File sudah otomatis tersimpan di folder sql/

# 2. Eksekusi SQL ke Supabase
echo "" | tee -a ../$LOG_FILE
echo "Step 2: Update database Supabase..." | tee -a ../$LOG_FILE
cd ../scheduler
go run execute_sql.go 2>&1 | tee -a $LOG_FILE

if [ $? -ne 0 ]; then
    echo "ERROR: Update database gagal!" | tee -a $LOG_FILE
    exit 1
fi

echo "" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE
echo "✅ Proses selesai - $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE

# Kirim notifikasi (optional)
# curl -X POST https://your-webhook-url -d "Gold price updated successfully"

# Cleanup old logs (hapus log lebih dari 30 hari)
find logs/ -name "scraper_*.log" -mtime +30 -delete
