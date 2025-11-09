#!/bin/sh
# Alternative scheduler tanpa cron (untuk fix setpgid error)

echo "=========================================="
echo "🔄 Loop-based Scheduler Started"
echo "=========================================="
echo "Schedule: 08:00 and 13:00 WIB daily"
echo "Timezone: ${TZ:-UTC}"
echo "=========================================="
echo ""

# Parse CRON_SCHEDULE jika ada (format: minute hour * * *)
if [ -n "$CRON_SCHEDULE" ]; then
    HOUR1=$(echo "$CRON_SCHEDULE" | awk '{print $2}' | cut -d',' -f1)
    HOUR2=$(echo "$CRON_SCHEDULE" | awk '{print $2}' | cut -d',' -f2)
    MINUTE=$(echo "$CRON_SCHEDULE" | awk '{print $1}')
    
    echo "📅 Parsed schedule:"
    echo "   - Hour 1: $HOUR1:$MINUTE"
    if [ -n "$HOUR2" ] && [ "$HOUR2" != "$HOUR1" ]; then
        echo "   - Hour 2: $HOUR2:$MINUTE"
    fi
    echo ""
else
    # Default
    HOUR1="08"
    HOUR2="13"
    MINUTE="00"
fi

# Tambah leading zero jika perlu
HOUR1=$(printf "%02d" $HOUR1)
HOUR2=$(printf "%02d" $HOUR2)
MINUTE=$(printf "%02d" $MINUTE)

LAST_RUN=""

while true; do
    CURRENT_TIME=$(date +%H:%M)
    CURRENT_HOUR=$(date +%H)
    CURRENT_MIN=$(date +%M)
    CURRENT_DATE=$(date +%Y-%m-%d)
    
    # Format untuk comparison
    RUN_KEY="${CURRENT_DATE}_${CURRENT_HOUR}${CURRENT_MIN}"
    
    SHOULD_RUN=false
    
    # Check if should run
    if [ "$CURRENT_HOUR" = "$HOUR1" ] && [ "$CURRENT_MIN" = "$MINUTE" ]; then
        SHOULD_RUN=true
    elif [ -n "$HOUR2" ] && [ "$CURRENT_HOUR" = "$HOUR2" ] && [ "$CURRENT_MIN" = "$MINUTE" ]; then
        SHOULD_RUN=true
    fi
    
    # Run if scheduled and not already run this time
    if [ "$SHOULD_RUN" = "true" ] && [ "$LAST_RUN" != "$RUN_KEY" ]; then
        echo ""
        echo "=========================================="
        echo "⏰ Scheduled time reached: $CURRENT_TIME"
        echo "🚀 Starting scraper..."
        echo "=========================================="
        
        cd /app && ./run_scraper.sh >> /app/logs/cron.log 2>&1
        EXIT_CODE=$?
        
        LAST_RUN="$RUN_KEY"
        
        echo "=========================================="
        if [ $EXIT_CODE -eq 0 ]; then
            echo "✅ Scraper completed successfully"
        else
            echo "❌ Scraper failed with exit code: $EXIT_CODE"
        fi
        echo "⏰ Next run: $HOUR1:$MINUTE"
        if [ -n "$HOUR2" ] && [ "$HOUR2" != "$HOUR1" ]; then
            echo "            $HOUR2:$MINUTE"
        fi
        echo "=========================================="
        echo ""
        
        # Sleep 2 menit untuk avoid double-run
        sleep 120
    else
        # Print heartbeat setiap 10 menit
        if [ "$CURRENT_MIN" = "00" ] || [ "$CURRENT_MIN" = "30" ]; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] 💓 Scheduler running... Next: $HOUR1:$MINUTE"
            if [ -n "$HOUR2" ] && [ "$HOUR2" != "$HOUR1" ]; then
                echo "                                           and $HOUR2:$MINUTE"
            fi
        fi
    fi
    
    # Check setiap 60 detik
    sleep 60
done
