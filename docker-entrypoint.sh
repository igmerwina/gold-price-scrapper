#!/bin/sh

set -e

echo "🚀 Starting Gold Scraper"
echo "Timezone: $TZ"
echo "Cron: ${CRON_SCHEDULE:-10 8 * * *}"

mkdir -p /app/logs /app/sql

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

CRON_SCHEDULE=${CRON_SCHEDULE:-"10 8 * * *"}

cat > /etc/crontabs/root << CRONEOF
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

echo "✅ Cron configured: $CRON_SCHEDULE"

if [ "${RUN_ON_STARTUP}" = "true" ]; then
    echo "🔄 Running scraper on startup..."
    cd /app && ./run_scraper.sh
fi

echo "🚀 Starting cron daemon..."
crond -f -l 2
