# 🚀 Quick Start - Gold Price Scraper dengan Supabase

## Setup Cepat (5 Menit)

### 1. **Install Dependencies**
```bash
cd /Users/macbook/Documents/code/random/gold-scrapper
go get github.com/lib/pq
go mod tidy
```

### 2. **Konfigurasi Database**
```bash
# Copy template
cp .env.example .env

# Edit dengan credentials Supabase Anda
nano .env  # atau gunakan text editor favorit
```

Isi `.env`:
```env
SUPABASE_HOST=db.xxxxx.supabase.co
SUPABASE_PASSWORD=your-password-here
```

**📍 Cara dapat credentials:**
- Login ke [Supabase Dashboard](https://app.supabase.com)
- Pilih project → **Settings** → **Database**
- Lihat **Connection string** → Copy Host dan Password

### 3. **Setup Otomatis**
```bash
./setup.sh
```

Atau manual:

### 4. **Setup Cron (Manual)**
```bash
# Buka crontab
crontab -e

# Tambahkan baris ini (jalankan setiap hari jam 8:10 pagi)
10 8 * * * /Users/macbook/Documents/code/random/gold-scrapper/run_scraper.sh

# Simpan dan keluar (ESC lalu :wq di vi)
```

### 5. **Test**
```bash
# Test scraper
go run scrapper.go

# Test SQL executor  
export $(cat .env | xargs)
go run execute_sql.go

# Test full automation
./run_scraper.sh
```

## 📊 Yang Terjadi Setiap Hari Jam 8:10 Pagi

```mermaid
graph LR
    A[Cron Trigger] --> B[Scrape Website]
    B --> C[Generate JSON]
    C --> D[Generate SQL]
    D --> E[Execute to Supabase]
    E --> F[Log Results]
```

1. ✅ Scrape harga emas dari website
2. ✅ Generate `harga_emas.json`
3. ✅ Generate `update_gold_prices.sql`
4. ✅ Execute SQL ke Supabase
5. ✅ Save logs ke `logs/scraper_YYYYMMDD_HHMMSS.log`

## 📝 File Penting

| File | Fungsi |
|------|--------|
| `scrapper.go` | Web scraper utama |
| `execute_sql.go` | Eksekusi SQL ke Supabase |
| `run_scraper.sh` | Script automation |
| `.env` | Database credentials (**JANGAN COMMIT!**) |
| `logs/` | Folder untuk log files |

## 🔍 Monitoring

### Lihat log terbaru:
```bash
tail -f logs/scraper_*.log
```

### List semua cron jobs:
```bash
crontab -l
```

### Check last run:
```bash
ls -lt logs/ | head -5
```

## 🛠️ Troubleshooting

### Cron tidak jalan?
```bash
# Check cron service
sudo launchctl list | grep cron

# Check system logs
log show --predicate 'process == "cron"' --last 1h --info

# Test manual
./run_scraper.sh
```

### Database connection error?
```bash
# Test koneksi dengan psql
psql "postgresql://postgres:PASSWORD@HOST:5432/postgres"

# Atau test di code
export $(cat .env | xargs)
go run execute_sql.go
```

### Path issues di cron?
Edit `run_scraper.sh` dan tambahkan full path:
```bash
export PATH=/usr/local/go/bin:/usr/bin:/bin
```

## 🎯 Cron Schedule Examples

```bash
# Setiap hari jam 8:10 pagi
10 8 * * * /path/to/run_scraper.sh

# Senin-Jumat jam 8:10 pagi
10 8 * * 1-5 /path/to/run_scraper.sh

# Setiap 2 jam
0 */2 * * * /path/to/run_scraper.sh

# Jam 8:10 pagi dan 8:10 malam
10 8,20 * * * /path/to/run_scraper.sh
```

## 🔐 Security Checklist

- ✅ File `.env` sudah di `.gitignore`
- ✅ Jangan hardcode password di code
- ✅ Set proper file permissions (`chmod 600 .env`)
- ✅ Rotate password berkala
- ✅ Monitor logs untuk anomali

## 📞 Quick Commands

```bash
# Edit crontab
crontab -e

# List crontab
crontab -l

# Remove crontab
crontab -r

# Edit .env
nano .env

# View logs
tail -f logs/scraper_*.log

# Test scraper only
go run scrapper.go

# Test SQL executor only  
export $(cat .env | xargs) && go run execute_sql.go
```

## ✨ Features

- 🤖 **Otomatis scraping** setiap hari jam 8:10 pagi
- 💾 **Auto save** ke Supabase database
- 📊 **Multi-brand support**: Galeri24, Antam, UBS
- ⚖️ **Multi-weight**: 0.5, 1, 2, 5, 10, 25, 50, 100 gram
- 📝 **Logging** lengkap dengan timestamp
- 🔄 **Auto cleanup** log files > 30 hari
- ⚡ **Fast execution** dengan connection pooling

## 🎉 Done!

Sekarang sistem akan otomatis update harga emas setiap hari jam 8:10 pagi! 🚀

---

**Need help?** Check `README_SETUP.md` untuk dokumentasi lengkap.
