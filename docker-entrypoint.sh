#!/bin/sh

set -e

# Trap signals for graceful shutdown
trap 'echo "Shutting down..."; kill $(jobs -p) 2>/dev/null; exit 0' SIGTERM SIGINT

echo "=========================================="
echo "🚀 Gold Scraper - Starting"
echo "=========================================="
echo "📅 Date: $(date)"
echo "🌍 Timezone: ${TZ:-UTC}"
echo "⏰ Cron Schedule: ${CRON_SCHEDULE:-10 8 * * *}"
echo "🏷️  Table Name: ${TABLE_NAME:-gold_prices_v2}"
echo "=========================================="

mkdir -p /app/logs /app/sql

# Write environment variables to .env file
cat > /app/.env << ENVEOF
SUPABASE_HOST=${SUPABASE_HOST}
SUPABASE_PORT=${SUPABASE_PORT}
SUPABASE_USER=${SUPABASE_USER}
SUPABASE_PASSWORD=${SUPABASE_PASSWORD}
SUPABASE_DB=${SUPABASE_DB}
SUPABASE_SSL_MODE=${SUPABASE_SSL_MODE}
TABLE_NAME=${TABLE_NAME:-gold_prices_v2}
IS_DOCKER=true
TZ=${TZ}
ENVEOF

# Get and clean CRON_SCHEDULE
CRON_SCHEDULE=${CRON_SCHEDULE:-"10 8 * * *"}
# Strip leading/trailing quotes and whitespace
CRON_SCHEDULE=$(echo "$CRON_SCHEDULE" | sed 's/^["'\'']\(.*\)["'\'']$/\1/' | xargs)

echo ""
echo "🔍 Validating cron schedule..."
echo "   Raw value: '${CRON_SCHEDULE}'"

# Validate cron schedule format (must have exactly 5 fields)
if [ -z "$CRON_SCHEDULE" ]; then
    echo "❌ ERROR: CRON_SCHEDULE is empty!"
    exit 1
fi

# Count fields (should be 5)
FIELD_COUNT=$(echo "$CRON_SCHEDULE" | awk '{print NF}')
if [ "$FIELD_COUNT" -ne 5 ]; then
    echo "❌ ERROR: Invalid CRON_SCHEDULE format (expected 5 fields, got $FIELD_COUNT)"
    echo "   Schedule: '$CRON_SCHEDULE'"
    echo ""
    echo "Valid examples:"
    echo "  0 8 * * *        - Daily at 8:00 AM"
    echo "  0 8,13 * * *     - Daily at 8:00 AM and 1:00 PM"
    echo "  */30 * * * *     - Every 30 minutes"
    echo "  0 */2 * * *      - Every 2 hours"
    exit 1
fi

# Additional validation: check if format matches cron pattern
if ! echo "$CRON_SCHEDULE" | grep -qE '^([0-9*,/-]+\s+){4}[0-9*,/-]+$'; then
    echo "❌ ERROR: Invalid CRON_SCHEDULE format"
    echo "   Schedule: '$CRON_SCHEDULE'"
    echo ""
    echo "Valid examples:"
    echo "  0 8 * * *        - Daily at 8:00 AM"
    echo "  0 8,13 * * *     - Daily at 8:00 AM and 1:00 PM"
    echo "  */30 * * * *     - Every 30 minutes"
    exit 1
fi

echo "✅ Valid cron format: $CRON_SCHEDULE"

# Create crontab with environment variables
cat > /etc/crontabs/root << CRONEOF
SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
SUPABASE_HOST=${SUPABASE_HOST}
SUPABASE_PORT=${SUPABASE_PORT}
SUPABASE_USER=${SUPABASE_USER}
SUPABASE_PASSWORD=${SUPABASE_PASSWORD}
SUPABASE_DB=${SUPABASE_DB}
SUPABASE_SSL_MODE=${SUPABASE_SSL_MODE}
TABLE_NAME=${TABLE_NAME:-gold_prices_v2}
TZ=${TZ}

$CRON_SCHEDULE cd /app && ./run_scraper.sh >> /app/logs/cron.log 2>&1

CRONEOF

echo ""
echo "📋 Cron Configuration:"
cat /etc/crontabs/root
echo ""
echo "✅ Cron configured successfully"

if [ "${RUN_ON_STARTUP}" = "true" ]; then
    echo ""
    echo "🔄 Running scraper on startup..."
    cd /app && ./run_scraper.sh
    echo ""
fi

echo "=========================================="

# Dummy HTTP listener so Render's port scan passes (this is a cron worker, not a web app)
PORT=${PORT:-10000}
mkdir -p /app/www
echo "OK" > /app/www/index.html
httpd -f -p "0.0.0.0:${PORT}" -h /app/www &
echo "🌐 Dummy HTTP listener on 0.0.0.0:${PORT}"

# Check if should use loop scheduler (for setpgid error fix)
if [ "${USE_LOOP_SCHEDULER}" = "true" ]; then
    echo "🔄 Using loop-based scheduler (no cron)"
    echo "📝 Logs will be written to: /app/logs/cron.log"
    echo "=========================================="
    echo ""
    exec /app/loop-scheduler.sh
else
    echo "🚀 Starting cron daemon..."
    echo "📝 Logs will be written to: /app/logs/cron.log"
    echo "=========================================="
    echo ""
    
    # Load crontab
    crontab /etc/crontabs/root
    
    # Start cron daemon in background
    crond -l 2
    
    # Keep container running by tailing the log
    touch /app/logs/cron.log
    tail -f /app/logs/cron.log
fi
