#!/bin/bash

# Quick Setup Script untuk Gold Price Scraper

echo "🚀 Gold Price Scraper - Quick Setup"
echo "===================================="
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Membuat file .env dari template..."
    cp .env.example .env
    echo "⚠️  PENTING: Edit file .env dan isi dengan credentials Supabase Anda!"
    echo ""
    echo "Cara mendapatkan credentials:"
    echo "1. Buka Supabase Dashboard"
    echo "2. Settings > Database"
    echo "3. Copy Connection Info"
    echo ""
    read -p "Tekan ENTER setelah selesai edit .env..."
else
    echo "✅ File .env sudah ada"
fi

# Test .env file
echo ""
echo "🔍 Validasi .env file..."
# Export variables safely (skip comments and empty lines)
export $(grep -v '^#' .env | grep -v '^$' | xargs)

if [ -z "$SUPABASE_HOST" ] || [ -z "$SUPABASE_PASSWORD" ]; then
    echo "❌ Error: SUPABASE_HOST atau SUPABASE_PASSWORD belum diset!"
    echo "   Edit file .env terlebih dahulu"
    exit 1
fi

echo "✅ Environment variables OK"
echo "   Host: $SUPABASE_HOST"
echo "   User: $SUPABASE_USER"
echo ""

# Setup cron
if [ -z "$CRON_SCHEDULE" ]; then
    CRON_SCHEDULE="10 8 * * *"
    CRON_DESCRIPTION="Setiap hari jam 8:10 pagi"
fi

echo "⏰ Setup Cron Job"
echo ""
echo "Jadwal: $CRON_DESCRIPTION"
echo "Cron line yang akan ditambahkan:"
echo "$CRON_SCHEDULE $(pwd)/run_scraper.sh"
echo ""
echo "💡 Tip: Ubah jadwal di file .env (CRON_SCHEDULE)"
echo ""
read -p "Tambahkan ke crontab? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Backup crontab
    crontab -l > /tmp/crontab_backup_$(date +%Y%m%d_%H%M%S).txt 2>/dev/null
    
    # Add new cron job (check if not exists)
    (crontab -l 2>/dev/null | grep -v "run_scraper.sh"; echo "$CRON_SCHEDULE $(pwd)/run_scraper.sh") | crontab -
    
    echo "✅ Cron job ditambahkan!"
    echo ""
    echo "Crontab saat ini:"
    crontab -l
else
    echo "⏭️  Skip cron setup"
    echo ""
    echo "Untuk setup manual, jalankan:"
    echo "  crontab -e"
    echo ""
    echo "Lalu tambahkan baris ini:"
    echo "  $CRON_SCHEDULE $(pwd)/run_scraper.sh"
fi

echo ""
echo "===================================="
echo "✅ Setup Selesai!"
echo "===================================="
echo ""
echo "📋 Next Steps:"
echo "1. Test scraper:"
echo "   go run scrapper.go"
echo ""
echo "2. Test SQL executor:"
echo "   export \$(cat .env | xargs) && go run execute_sql.go"
echo ""
echo "3. Test full automation:"
echo "   ./run_scraper.sh"
echo ""
echo "4. Check logs:"
echo "   tail -f logs/scraper_*.log"
echo ""
echo "5. List cron jobs:"
echo "   crontab -l"
echo ""
