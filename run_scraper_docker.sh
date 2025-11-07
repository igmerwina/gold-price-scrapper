#!/bin/sh

# Docker-compatible wrapper for gold scraper
# This script is designed to run inside Docker container

# Detect if running in Docker
if [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
    SCRIPT_DIR="/app"
else
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
fi

cd "$SCRIPT_DIR"

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
else
    echo "⚠️  Warning: .env file not found, using environment variables"
fi

# Log file
LOG_FILE="logs/scraper_$(date +%Y%m%d_%H%M%S).log"
mkdir -p logs sql

echo "========================================" | tee -a $LOG_FILE
echo "Gold Price Scraper - $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE
echo "Environment: ${CONTAINER_ENV:-local}" | tee -a $LOG_FILE
echo "Working Dir: $(pwd)" | tee -a $LOG_FILE
echo "" | tee -a $LOG_FILE

# Check if scraper binary exists (Docker) or need to compile (local)
if [ -f ./scraper ]; then
    # Running in Docker with pre-compiled binary
    echo "Step 1: Running scraper binary..." | tee -a $LOG_FILE
    ./scraper 2>&1 | tee -a $LOG_FILE
    
    if [ $? -ne 0 ]; then
        echo "ERROR: Scraping gagal!" | tee -a $LOG_FILE
        exit 1
    fi
    
    # Execute SQL
    echo "" | tee -a $LOG_FILE
    echo "Step 2: Executing SQL to Supabase..." | tee -a $LOG_FILE
    
    # Build execute_sql if needed
    if [ ! -f ./execute_sql ]; then
        echo "Building execute_sql..." | tee -a $LOG_FILE
        go build -o execute_sql execute_sql.go 2>&1 | tee -a $LOG_FILE
    fi
    
    ./execute_sql 2>&1 | tee -a $LOG_FILE
    
else
    # Running locally with go run
    echo "Step 1: Scraping data harga emas..." | tee -a $LOG_FILE
    
    if [ -d ../scrapper ]; then
        cd ../scrapper
        go run scrapper.go 2>&1 | tee -a "../$LOG_FILE"
        
        if [ $? -ne 0 ]; then
            echo "ERROR: Scraping gagal!" | tee -a "../$LOG_FILE"
            exit 1
        fi
        cd "$SCRIPT_DIR"
    else
        echo "ERROR: scrapper directory not found!" | tee -a $LOG_FILE
        exit 1
    fi
    
    # Execute SQL
    echo "" | tee -a $LOG_FILE
    echo "Step 2: Update database Supabase..." | tee -a $LOG_FILE
    go run execute_sql.go 2>&1 | tee -a $LOG_FILE
fi

if [ $? -ne 0 ]; then
    echo "ERROR: Update database gagal!" | tee -a $LOG_FILE
    exit 1
fi

echo "" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE
echo "✅ Proses selesai - $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE

# Cleanup old logs (older than 30 days)
find logs/ -name "scraper_*.log" -mtime +30 -delete 2>/dev/null

exit 0
