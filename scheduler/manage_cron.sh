#!/bin/bash

# Script untuk manage cron job gold scraper

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables (safely, handle spaces in values)
if [ -f .env ]; then
    set -a
    source <(grep -v '^#' .env | grep -v '^$' | sed 's/^/export /')
    set +a
else
    echo "❌ File .env tidak ditemukan!"
    exit 1
fi

# Default values
CRON_SCHEDULE=${CRON_SCHEDULE:-"10 8 * * *"}
CRON_DESCRIPTION=${CRON_DESCRIPTION:-"Daily at 8:10 AM"}

show_help() {
    echo "🔧 Gold Scraper - Cron Manager"
    echo "======================================"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  install     - Install cron job"
    echo "  remove      - Remove cron job"
    echo "  list        - List current cron jobs"
    echo "  status      - Show cron job status"
    echo "  test        - Test cron command"
    echo "  edit        - Edit crontab"
    echo "  help        - Show this help"
    echo ""
    echo "Current Configuration:"
    echo "  Schedule: $CRON_SCHEDULE"
    echo "  Description: $CRON_DESCRIPTION"
    echo "  Script: $SCRIPT_DIR/run_scraper.sh"
    echo ""
}

install_cron() {
    echo "⏰ Installing Cron Job"
    echo "======================================"
    echo ""
    echo "Schedule: $CRON_DESCRIPTION"
    echo "Cron Expression: $CRON_SCHEDULE"
    echo "Script: $SCRIPT_DIR/run_scraper.sh"
    echo ""
    
    # Check if already exists
    if crontab -l 2>/dev/null | grep -q "run_scraper.sh"; then
        echo "⚠️  Cron job sudah ada!"
        echo ""
        read -p "Replace existing cron job? (y/n): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Cancelled"
            exit 0
        fi
    fi
    
    # Backup crontab
    BACKUP_FILE="/tmp/crontab_backup_$(date +%Y%m%d_%H%M%S).txt"
    crontab -l > "$BACKUP_FILE" 2>/dev/null
    echo "💾 Backup saved to: $BACKUP_FILE"
    
    # Add/Replace cron job
    (crontab -l 2>/dev/null | grep -v "run_scraper.sh"; echo "$CRON_SCHEDULE $SCRIPT_DIR/run_scraper.sh") | crontab -
    
    echo ""
    echo "✅ Cron job installed successfully!"
    echo ""
    echo "Current crontab:"
    crontab -l
}

remove_cron() {
    echo "🗑️  Removing Cron Job"
    echo "======================================"
    echo ""
    
    if ! crontab -l 2>/dev/null | grep -q "run_scraper.sh"; then
        echo "ℹ️  No cron job found"
        exit 0
    fi
    
    # Backup crontab
    BACKUP_FILE="/tmp/crontab_backup_$(date +%Y%m%d_%H%M%S).txt"
    crontab -l > "$BACKUP_FILE" 2>/dev/null
    echo "💾 Backup saved to: $BACKUP_FILE"
    
    # Remove cron job
    crontab -l 2>/dev/null | grep -v "run_scraper.sh" | crontab -
    
    echo ""
    echo "✅ Cron job removed successfully!"
}

list_cron() {
    echo "📋 Current Cron Jobs"
    echo "======================================"
    echo ""
    
    if crontab -l 2>/dev/null | grep -q "run_scraper.sh"; then
        echo "✅ Gold Scraper cron job found:"
        echo ""
        crontab -l | grep "run_scraper.sh"
        echo ""
        echo "All cron jobs:"
        crontab -l
    else
        echo "❌ No gold scraper cron job found"
        echo ""
        echo "All cron jobs:"
        crontab -l 2>/dev/null || echo "No cron jobs configured"
    fi
}

status_cron() {
    echo "📊 Cron Job Status"
    echo "======================================"
    echo ""
    
    echo "Configuration from .env:"
    echo "  Schedule: $CRON_SCHEDULE ($CRON_DESCRIPTION)"
    echo ""
    
    if crontab -l 2>/dev/null | grep -q "run_scraper.sh"; then
        echo "Status: ✅ Installed"
        echo ""
        echo "Installed cron job:"
        crontab -l | grep "run_scraper.sh"
        echo ""
        
        # Parse cron schedule
        INSTALLED_SCHEDULE=$(crontab -l | grep "run_scraper.sh" | sed 's|/.*||')
        
        if [ "$INSTALLED_SCHEDULE" = "$CRON_SCHEDULE" ]; then
            echo "✅ Cron schedule matches .env configuration"
        else
            echo "⚠️  Cron schedule differs from .env:"
            echo "   Installed: $INSTALLED_SCHEDULE"
            echo "   .env: $CRON_SCHEDULE"
            echo ""
            echo "   Run '$0 install' to update"
        fi
    else
        echo "Status: ❌ Not installed"
        echo ""
        echo "Run '$0 install' to install"
    fi
    
    echo ""
    echo "Recent logs:"
    if [ -d logs ]; then
        ls -lt logs/scraper_*.log 2>/dev/null | head -5
    else
        echo "  No logs found"
    fi
}

test_cron() {
    echo "🧪 Testing Cron Command"
    echo "======================================"
    echo ""
    echo "Running: $SCRIPT_DIR/run_scraper.sh"
    echo ""
    
    bash "$SCRIPT_DIR/run_scraper.sh"
}

edit_cron() {
    echo "✏️  Opening crontab editor..."
    crontab -e
}

# Main
case "${1:-help}" in
    install)
        install_cron
        ;;
    remove)
        remove_cron
        ;;
    list)
        list_cron
        ;;
    status)
        status_cron
        ;;
    test)
        test_cron
        ;;
    edit)
        edit_cron
        ;;
    help|*)
        show_help
        ;;
esac
