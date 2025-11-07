#!/bin/bash

# Script untuk menjalankan gold price scraper dan update ke Supabase
# Dijalankan oleh cron scheduler setiap hari jam 8:10 pagi

# Set working directory
cd /Users/macbook/Documents/code/random/gold-scrapper/scheduler

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
cd ../scrapper
go run scrapper.go 2>&1 | tee -a ../$LOG_FILE

if [ $? -ne 0 ]; then
    echo "ERROR: Scraping gagal!" | tee -a ../$LOG_FILE
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
