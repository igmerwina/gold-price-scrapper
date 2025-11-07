#!/bin/sh

set -e

cd /app

if [ -f .env ]; then
    set -a
    while IFS= read -r line; do
        case "$line" in
            \#*|"") continue ;;
            *) export "$line" ;;
        esac
    done < .env
    set +a
fi

LOG_FILE="logs/scraper_$(date +%Y%m%d_%H%M%S).log"
mkdir -p logs

echo "========================================" | tee -a $LOG_FILE
echo "Gold Price Scraper - $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE

echo "" | tee -a $LOG_FILE
echo "Step 1: Scraping..." | tee -a $LOG_FILE

./scraper 2>&1 | tee -a $LOG_FILE

if [ $? -ne 0 ]; then
    echo "ERROR: Scraping failed!" | tee -a $LOG_FILE
    exit 1
fi

echo "" | tee -a $LOG_FILE
echo "Step 2: Database update..." | tee -a $LOG_FILE

./execute_sql 2>&1 | tee -a $LOG_FILE

if [ $? -ne 0 ]; then
    echo "ERROR: Database update failed!" | tee -a $LOG_FILE
    exit 1
fi

echo "" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE
echo "✅ Completed - $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE
