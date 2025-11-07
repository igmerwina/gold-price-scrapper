#!/bin/sh

# Gold Scraper - Docker Entrypoint with Cron

echo "🚀 Starting Gold Scraper in Docker..."
echo "Timezone: $TZ"
echo "Cron Schedule: ${CRON_SCHEDULE:-10 8 * * *}"

# Create log directory
mkdir -p /app/logs /app/sql

# Export environment variables to .env file for cron to access
echo "📝 Creating .env file for cron..."
cat > /app/.env << EOF
SUPABASE_HOST=${SUPABASE_HOST}
SUPABASE_PORT=${SUPABASE_PORT}
SUPABASE_USER=${SUPABASE_USER}
SUPABASE_PASSWORD=${SUPABASE_PASSWORD}
SUPABASE_DB=${SUPABASE_DB}
SUPABASE_SSL_MODE=${SUPABASE_SSL_MODE}
CRON_SCHEDULE=${CRON_SCHEDULE:-"10 8 * * *"}
CRON_DESCRIPTION=${CRON_DESCRIPTION:-"Daily"}
TZ=${TZ}
EOF

# Setup cron job
CRON_SCHEDULE=${CRON_SCHEDULE:-"10 8 * * *"}

# Write crontab with environment variables inline
cat > /etc/crontabs/root << EOF
# Environment variables for cron
SUPABASE_HOST=${SUPABASE_HOST}
SUPABASE_PORT=${SUPABASE_PORT}
SUPABASE_USER=${SUPABASE_USER}
SUPABASE_PASSWORD=${SUPABASE_PASSWORD}
SUPABASE_DB=${SUPABASE_DB}
SUPABASE_SSL_MODE=${SUPABASE_SSL_MODE}
TZ=${TZ}

# Cron job
$CRON_SCHEDULE cd /app && ./run_scraper.sh >> /app/logs/cron.log 2>&1
EOF

# Start cron in background
crond -f -l 2 &

echo "✅ Cron job scheduled: $CRON_SCHEDULE"
echo "📋 Checking cron jobs:"
crontab -l

# Run once on startup (optional)
if [ "${RUN_ON_STARTUP:-true}" = "true" ]; then
    echo "🔄 Running scraper on startup..."
    cd /app && ./run_scraper.sh
fi

# Keep container running
echo "✅ Gold Scraper is running. Logs will appear in /app/logs/"
echo "Press Ctrl+C to stop."

# Tail logs to keep container alive
tail -f /app/logs/*.log 2>/dev/null || tail -f /dev/null
