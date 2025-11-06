# 🏅 Gold Price Scraper & Automation

[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Automated scraper untuk mengambil harga emas dari [Galeri24.co.id](https://galeri24.co.id/harga-emas) dan update ke Supabase. Support Galeri24, Antam, UBS (0.5-100 gram).

## ✨ Features

- 🤖 Scraping otomatis dengan headless browser
- 📊 Multi-brand & multi-weight support
- 💾 Auto update ke Supabase PostgreSQL
- ⏰ Cron scheduler (fleksibel via `.env`)
- 📝 Comprehensive logging
- 🔄 Auto cleanup log files

## 🚀 Quick Start

### 1. Prerequisites

- Go 1.16+
- Chrome/Chromium
- Supabase account

### 2. Installation

```bash
# Clone repository
git clone https://github.com/yourusername/gold-scrapper.git
cd gold-scrapper/scheduler

# Install dependencies
go mod download
```

### 3. Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit .env dengan Supabase credentials Anda
nano .env
```

**Minimal configuration:**
```env
SUPABASE_HOST=your-project.supabase.co
SUPABASE_PORT=6543
SUPABASE_USER=postgres.xxxxx
SUPABASE_PASSWORD=your-password
SUPABASE_DB=postgres
CRON_SCHEDULE="10 8 * * *"
```

### 4. Database Setup

```sql
CREATE TABLE public.gold_prices_v3 (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    brand VARCHAR(50) NOT NULL,
    denom DECIMAL(10,2) NOT NULL,
    price_sell DECIMAL(15,2),
    price_buyback DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(date, brand, denom)
);
```

## 🎯 Usage

### Test Manual

```bash
# Test scraper
cd scrapper
go run scrapper.go

# Test full automation
cd scheduler
./run_scraper.sh
```

### Setup Cron (Auto-run)

```bash
cd scheduler
chmod +x manage_cron.sh setup.sh

# Option 1: Quick setup
./setup.sh

# Option 2: Manual
./manage_cron.sh install
./manage_cron.sh status
```

### Cron Management

```bash
./manage_cron.sh install   # Install/update cron job
./manage_cron.sh status    # Check status
./manage_cron.sh list      # List all cron jobs
./manage_cron.sh remove    # Remove cron job
./manage_cron.sh test      # Test run
```

### Ubah Jadwal

Edit `.env`:
```env
CRON_SCHEDULE="10 8 * * *"      # Daily 8:10 AM
CRON_SCHEDULE="0 */4 * * *"     # Every 4 hours
CRON_SCHEDULE="30 9,15 * * *"   # 9:30 & 15:30
CRON_SCHEDULE="0 8 * * 1-5"     # Mon-Fri 8 AM
```

Update: `./manage_cron.sh install`

### Monitoring

```bash
tail -f scheduler/logs/scraper_*.log
```

## 📁 Project Structure

```
gold-scrapper/
├── scheduler/          # Production (automated)
│   ├── execute_sql.go
│   ├── run_scraper.sh
│   ├── manage_cron.sh
│   ├── setup.sh
│   └── logs/
├── scrapper/          # Development (manual)
│   └── scrapper.go
├── sql/               # Output folder
│   ├── harga_emas.json
│   └── update_gold_prices.sql
└── README.md
```

## 🔄 How It Works

1. Cron trigger → `run_scraper.sh`
2. Headless browser scrape data
3. Parse HTML → Filter data
4. Generate JSON + SQL
5. Execute to Supabase
6. Save logs + cleanup

## 🐛 Troubleshooting

**Cron tidak jalan?**
```bash
crontab -l                    # Verify cron exists
./manage_cron.sh status       # Check status
./run_scraper.sh              # Test manual
```

**Database error?**
```bash
# Test connection
cd scheduler
export $(grep -v '^#' .env | grep -v '^$' | xargs)
go run execute_sql.go
```

**Scraper gagal?**
- Check website: https://galeri24.co.id/harga-emas
- Check Chrome/Chromium installed
- Review logs: `tail scheduler/logs/scraper_*.log`

## 📈 Performance

- Scraping: ~10-15s
- SQL execution: ~1-2s
- Total: ~15s per run

## 📚 Docs

- [QUICKSTART.md](scheduler/QUICKSTART.md) - 5 minute setup
- [README_SETUP.md](scheduler/README_SETUP.md) - Detailed guide

## 📄 License

MIT License - see [LICENSE](LICENSE)

## 👨‍💻 Author

**Your Name**
- GitHub: [@igmerwina](https://github.com/igmerwina)

---

⭐ Star this repo if helpful!
