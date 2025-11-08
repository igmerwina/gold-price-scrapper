#!/bin/sh

set -e

echo "=========================================="
echo "🚀 Gold Scraper - Starting"
echo "=========================================="
echo "📅 Date: $(date)"
echo "🌍 Timezone: ${TZ:-UTC}"
echo "⏰ Cron Schedule: ${CRON_SCHEDULE:-10 8 * * *}"
echo "🏷️  Table Name: ${TABLE_NAME:-gold_prices_v3}"
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
TABLE_NAME=${TABLE_NAME:-gold_prices_v3}
IS_DOCKER=true
TZ=${TZ}
ENVEOF

CRON_SCHEDULE=${CRON_SCHEDULE:-"10 8 * * *"}

# Validate cron schedule format
if ! echo "$CRON_SCHEDULE" | grep -qE '^([0-9*,/-]+\s+){4}[0-9*,/-]+$'; then
    echo "❌ ERROR: Invalid CRON_SCHEDULE format: $CRON_SCHEDULE"
    echo "Examples:"
    echo "  0 8 * * *        - Daily at 8:00 AM"
    echo "  0 8,13 * * *     - Daily at 8:00 AM and 1:00 PM"
    echo "  */30 * * * *     - Every 30 minutes"
    exit 1
fi

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
TABLE_NAME=${TABLE_NAME:-gold_prices_v3}
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
echo "🚀 Starting cron daemon in foreground..."
echo "📝 Logs will be written to: /app/logs/cron.log"
echo "=========================================="
echo ""

# Start cron in foreground with verbose logging
exec crond -f -l 0 -L /dev/stdout
