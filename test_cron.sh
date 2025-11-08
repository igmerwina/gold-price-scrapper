#!/bin/sh
# Test script untuk debug cron di Dokploy

echo "=========================================="
echo "🔍 Cron Debugging Tool"
echo "=========================================="
echo ""

echo "1️⃣ Current Date/Time:"
date
echo ""

echo "2️⃣ Timezone:"
echo "TZ=$TZ"
echo ""

echo "3️⃣ Environment Variables:"
env | grep -E "SUPABASE|TABLE|CRON|TZ" | sort
echo ""

echo "4️⃣ Crontab Configuration:"
crontab -l 2>/dev/null || echo "No crontab configured"
echo ""

echo "5️⃣ Cron Process Status:"
ps aux | grep crond | grep -v grep || echo "Cron not running"
echo ""

echo "6️⃣ Scripts Permissions:"
ls -la /app/*.sh /app/scraper /app/execute_sql 2>/dev/null
echo ""

echo "7️⃣ Log Files:"
ls -lah /app/logs/ 2>/dev/null || echo "No logs directory"
echo ""

echo "8️⃣ SQL Output Files:"
ls -lah /app/sql/ 2>/dev/null || echo "No sql directory"
echo ""

echo "9️⃣ Recent Cron Log (last 20 lines):"
tail -20 /app/logs/cron.log 2>/dev/null || echo "No cron log yet"
echo ""

echo "=========================================="
echo "✅ Debug complete!"
echo "=========================================="
