# 🏅 Gold Price Scraper & Automation

Automated scraper untuk mengambil harga emas dari [Galeri24.co.id](https://galeri24.co.id/harga-emas) dan update ke Supabase PostgreSQL.

## ✨ Features

- 🤖 Auto scraping dengan headless browser (ChromeDP)
- 📊 Support: Galeri24, Antam, UBS (0.5-100 gram)
- 💾 Auto update ke Supabase PostgreSQL
- ⏰ Configurable cron schedule via environment variables
- 🐳 Docker & Dokploy ready
- 📝 Auto logging & cleanup

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Chromium/Chrome browser
- Supabase account

### Local Setup

```bash
git clone https://github.com/yourusername/gold-scrapper.git
cd gold-scrapper

# Setup environment
cp scheduler/.env.example scheduler/.env
nano scheduler/.env  # Edit with your Supabase credentials
```

**Database Setup:**
```sql
CREATE TABLE public.gold_prices_v2 (
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

**Environment Variables:**
```env
SUPABASE_HOST=aws-1-ap-southeast-1.pooler.supabase.com
SUPABASE_PORT=6543
SUPABASE_USER=postgres.your-project-ref
SUPABASE_PASSWORD=your-password
SUPABASE_DB=postgres
TABLE_NAME=gold_prices_v2
CRON_SCHEDULE="10 8 * * *"
```

### Run Manually

```bash
# Test scraper only
cd scrapper
go run scrapper.go

# Run full automation (scrape + update DB)
cd scheduler
bash run_scraper.sh
```

## 🐳 Docker Deployment

### Using Docker Compose

```bash
# Create .env file
cp scheduler/.env.example .env
nano .env  # Edit with your credentials

# Run
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Using Dokploy

See [DEPLOY_DOKPLOY.md](DEPLOY_DOKPLOY.md) for complete deployment guide.

**Quick Deploy:**
1. Go to Dokploy → Create Application
2. Select "Docker" type
3. Set environment variables (SUPABASE_HOST, SUPABASE_PASSWORD, etc.)
4. Deploy from GitHub/Docker Hub

## 📋 Configuration

### Change Cron Schedule

Edit `.env`:
```env
CRON_SCHEDULE="10 8 * * *"      # Daily at 8:10 AM
CRON_SCHEDULE="0 */4 * * *"     # Every 4 hours
CRON_SCHEDULE="30 9,15 * * *"   # At 9:30 and 15:30
CRON_SCHEDULE="0 8 * * 1-5"     # Mon-Fri at 8 AM
```

No restart needed in Docker - just update env var and container will reload.

### Change Table Name

Edit `.env`:
```env
TABLE_NAME=gold_prices_v2       # Production
TABLE_NAME=gold_prices_staging  # Staging
TABLE_NAME=gold_prices_test     # Testing
```

## 📁 Project Structure

```
gold-scrapper/
├── scrapper/
│   └── scrapper.go          # Scraper logic
├── scheduler/
│   ├── execute_sql.go       # Database updater
│   └── run_scraper.sh       # Main runner
├── sql/                     # Output files
│   ├── harga_emas.json
│   └── update_gold_prices.sql
├── Dockerfile               # Docker build
├── docker-compose.yml       # Local Docker setup
└── docker-entrypoint.sh     # Container startup
```

## 🔄 How It Works

1. **Scraper** (`scrapper.go`):
   - Launch headless Chrome
   - Scrape Galeri24.co.id
   - Parse HTML → Extract prices
   - Generate JSON & SQL files

2. **Executor** (`execute_sql.go`):
   - Read generated SQL
   - Connect to Supabase
   - Execute UPDATE queries

3. **Automation**:
   - Cron triggers `run_scraper.sh`
   - Runs scraper + executor
   - Logs to `logs/scraper_*.log`
   - Auto cleanup old logs (30+ days)

## 🐛 Troubleshooting

**Container not running?**
```bash
docker-compose logs -f
docker exec -it gold-scraper sh
```

**No SQL file generated?**
```bash
# Check if scraper ran
docker exec -it gold-scraper ls -la /app/sql/

# Run manually
docker exec -it gold-scraper ./scraper
```

**Database connection failed?**
```bash
# Check env vars
docker exec -it gold-scraper env | grep SUPABASE

# Test connection
docker exec -it gold-scraper ./execute_sql
```

**Cron not running?**
```bash
# Check cron config
docker exec -it gold-scraper crontab -l

# Check cron logs
docker exec -it gold-scraper tail -f /app/logs/cron.log
```

## 📊 Output Files

**JSON** (`sql/harga_emas.json`):
```json
[
  {
    "brand": "GALERI 24",
    "data": [
      {
        "berat": "0.5",
        "harga_jual": "892500",
        "harga_buyback": "842500"
      }
    ]
  }
]
```

**SQL** (`sql/update_gold_prices.sql`):
```sql
UPDATE public.gold_prices_v2
SET price_buyback=842500.0, price_sell=892500
WHERE "date"='2025-11-07' AND brand='Galeri24' AND denom=0.5;
```

## 📈 Performance

- Scraping: ~10-12s
- SQL execution: ~1-2s
- Total runtime: ~15s

## 🔐 Security

- Never commit `.env` file
- Use connection pooler (port 6543)
- Enable SSL mode: `SUPABASE_SSL_MODE=require`

## 📄 License

MIT License

## 👨‍💻 Author

[@igmerwina](https://github.com/igmerwina)

---

⭐ Star if helpful!
