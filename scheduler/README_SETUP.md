# Gold Price Scraper - Setup Guide

## 📋 Prerequisites
- Go 1.16 atau lebih baru
- Akses ke Supabase database
- macOS dengan cron support

## 🔧 Setup

### 1. Install Dependencies
```bash
cd /Users/macbook/Documents/code/random/gold-scrapper
go get github.com/lib/pq
go mod tidy
```

### 2. Konfigurasi Database

Buat file `.env` dari template:
```bash
cp .env.example .env
```

Edit `.env` dan isi dengan credentials Supabase Anda:
```env
SUPABASE_HOST=db.xxxxx.supabase.co
SUPABASE_PORT=5432
SUPABASE_USER=postgres
SUPABASE_PASSWORD=your-password-here
SUPABASE_DB=postgres
SUPABASE_SSL_MODE=require
```

**Cara mendapatkan credentials Supabase:**
1. Buka Supabase Dashboard
2. Pilih project Anda
3. Klik **Settings** > **Database**
4. Scroll ke **Connection string** > **URI**
5. Copy informasi:
   - Host: dari URI (bagian setelah `@` dan sebelum `:`)
   - Password: dari Database password

### 3. Test Manual

#### Test Scraper:
```bash
go run scrapper.go
```

#### Test SQL Executor:
```bash
export $(cat .env | xargs)
go run execute_sql.go
```

### 4. Setup Cron Job (Scheduling)

#### Buat executable:
```bash
chmod +x run_scraper.sh
```

#### Edit crontab:
```bash
crontab -e
```

#### Tambahkan baris ini untuk jalankan setiap hari jam 8:10 pagi:
```cron
10 8 * * * /Users/macbook/Documents/code/random/gold-scrapper/run_scraper.sh
```

#### Simpan dan keluar (tekan `ESC` lalu ketik `:wq` jika menggunakan vi)

#### Verifikasi cron job terpasang:
```bash
crontab -l
```

### 5. Monitoring Logs

Logs disimpan di folder `logs/`:
```bash
# Lihat log terbaru
tail -f logs/scraper_*.log

# Lihat semua log hari ini
ls -lh logs/scraper_$(date +%Y%m%d)*.log
```

## 📝 Cron Schedule Format

```
* * * * * command
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, 0 dan 7 = Sunday)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

### Contoh Schedule Lainnya:

```bash
# Setiap hari jam 8:10 pagi
10 8 * * * /path/to/run_scraper.sh

# Setiap hari jam 8:10 pagi dan 8:10 malam
10 8,20 * * * /path/to/run_scraper.sh

# Setiap Senin-Jumat jam 8:10 pagi
10 8 * * 1-5 /path/to/run_scraper.sh

# Setiap 2 jam
0 */2 * * * /path/to/run_scraper.sh

# Setiap 30 menit
*/30 * * * * /path/to/run_scraper.sh
```

## 🧪 Testing

### Test koneksi Supabase:
```bash
export $(cat .env | xargs)
go run -e "package main; import (\"database/sql\"; \"fmt\"; \"log\"; _ \"github.com/lib/pq\"); func main() { db, err := sql.Open(\"postgres\", \"host=$SUPABASE_HOST port=$SUPABASE_PORT user=$SUPABASE_USER password=$SUPABASE_PASSWORD dbname=$SUPABASE_DB sslmode=$SUPABASE_SSL_MODE\"); if err != nil { log.Fatal(err) }; defer db.Close(); if err := db.Ping(); err != nil { log.Fatal(err) }; fmt.Println(\"✅ Koneksi berhasil!\") }"
```

### Atau gunakan psql:
```bash
psql "postgresql://postgres:your-password@db.xxxxx.supabase.co:5432/postgres"
```

## 🐛 Troubleshooting

### Cron tidak jalan?
```bash
# Check cron service
sudo launchctl list | grep cron

# Check system log
log show --predicate 'process == "cron"' --last 1h
```

### Path tidak ditemukan di cron?
Tambahkan full path di script:
```bash
export PATH=/usr/local/bin:/usr/bin:/bin
```

### Environment variables tidak load?
Pastikan `.env` file ada dan readable:
```bash
chmod 644 .env
cat .env
```

## 📊 Database Schema

Pastikan tabel `gold_prices_v2` sudah ada di Supabase:

```sql
CREATE TABLE IF NOT EXISTS public.gold_prices_v2 (
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

-- Index untuk performance
CREATE INDEX idx_gold_prices_date ON public.gold_prices_v2(date);
CREATE INDEX idx_gold_prices_brand ON public.gold_prices_v2(brand);
CREATE INDEX idx_gold_prices_denom ON public.gold_prices_v2(denom);
```

## 📁 Struktur File

```
gold-scrapper/
├── scrapper.go              # Web scraper
├── execute_sql.go           # SQL executor ke Supabase
├── generate_sql.go          # SQL generator
├── run_scraper.sh          # Automation script
├── .env                     # Environment variables (jangan commit!)
├── .env.example            # Template environment variables
├── harga_emas.json         # Output scraper
├── update_gold_prices.sql  # Generated SQL
├── go.mod                  # Go dependencies
├── go.sum                  # Go dependencies lock
├── logs/                   # Log files
│   └── scraper_*.log
└── README_SETUP.md         # This file
```

## 🔒 Security Notes

1. **Jangan commit file `.env`** - Tambahkan ke `.gitignore`
2. **Gunakan environment variables** untuk credentials
3. **Rotate password** secara berkala
4. **Limit database permissions** hanya untuk operasi yang diperlukan
5. **Monitor logs** untuk aktivitas mencurigakan

## 📞 Support

Jika ada masalah, check:
1. Logs di folder `logs/`
2. Cron logs: `grep CRON /var/log/system.log`
3. Database connection string
4. Network/firewall settings

## 🎯 Next Steps

1. ✅ Setup environment variables
2. ✅ Test scraper manual
3. ✅ Test SQL executor manual
4. ✅ Setup cron job
5. ✅ Monitor logs
6. 🔄 Setup alerting (optional)
7. 🔄 Setup backup (optional)
