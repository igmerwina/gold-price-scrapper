# 🏅 Gold Price Scraper & Automation

[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-316192?style=flat&logo=postgresql)](https://supabase.com/)

Automated gold price scraper untuk mengambil harga emas dari [Galeri24.co.id](https://galeri24.co.id/harga-emas) dan otomatis update ke database Supabase. Support multiple brands (Galeri24, Antam, UBS) dan berbagai gramasi.

## ✨ Features

- 🤖 **Automated Scraping** - Scrape harga emas secara otomatis dengan headless browser
- 📊 **Multi-Brand Support** - Galeri24, Antam, UBS
- ⚖️ **Multi-Weight** - 0.5, 1, 2, 5, 10, 25, 50, 100 gram
- 💾 **Auto Database Update** - Langsung update ke Supabase PostgreSQL
- ⏰ **Scheduled Execution** - Cron job support (default: setiap hari jam 8:10 pagi)
- 📝 **Comprehensive Logging** - Log semua aktivitas dengan timestamp
- ⚡ **Real-time Loading Animation** - Spinner dengan elapsed time
- 🔄 **Auto Cleanup** - Hapus log files > 30 hari otomatis

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [Scheduling](#scheduling)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## 🚀 Prerequisites

- **Go** 1.16 atau lebih baru ([Download](https://golang.org/dl/))
- **Chrome/Chromium** browser (untuk headless scraping)
- **Supabase Account** ([Sign up](https://supabase.com/))
- **macOS/Linux** dengan cron support

## 📦 Installation

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/gold-scrapper.git
cd gold-scrapper
```

### 2. Install Dependencies

#### Untuk Scheduler (Production)
```bash
cd scheduler
go mod download
go get github.com/lib/pq
go get github.com/chromedp/chromedp
go get github.com/antchfx/htmlquery
```

#### Untuk Scrapper (Development)
```bash
cd scrapper
go mod download
```

### 3. Setup Database

Buat tabel di Supabase:

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

## ⚙️ Configuration

### 1. Environment Variables

Copy file template:
```bash
cd scheduler
cp .env.example .env
```

Edit `.env` dengan credentials Supabase Anda:
```env
SUPABASE_HOST=db.xxxxx.supabase.co
SUPABASE_PORT=5432
SUPABASE_USER=postgres
SUPABASE_PASSWORD=your-database-password
SUPABASE_DB=postgres
SUPABASE_SSL_MODE=require
```

### 2. Mendapatkan Credentials Supabase

1. Login ke [Supabase Dashboard](https://app.supabase.com)
2. Pilih project Anda
3. Klik **Settings** → **Database**
4. Scroll ke **Connection string** → **URI**
5. Copy informasi:
   - **Host**: bagian setelah `@` dan sebelum `:`
   - **Password**: dari Database password section

## 🎯 Usage

### Development Mode (Manual Testing)

```bash
# Test scraper only
cd scrapper
go run scrapper.go

# Test dengan SQL execution
cd scheduler
export $(cat .env | xargs)
go run scrapper.go
go run execute_sql.go
```

### Production Mode (Automated)

#### Quick Setup
```bash
cd scheduler
chmod +x setup.sh
./setup.sh
```

#### Manual Setup
```bash
# Make script executable
chmod +x run_scraper.sh

# Run manually
./run_scraper.sh

# Setup cron job (jam 8:10 pagi setiap hari)
crontab -e
# Add this line:
10 8 * * * /full/path/to/gold-scrapper/scheduler/run_scraper.sh
```

### View Logs

```bash
# Real-time log monitoring
tail -f scheduler/logs/scraper_*.log

# View latest log
ls -t scheduler/logs/scraper_*.log | head -1 | xargs cat
```

## 📁 Project Structure

```
gold-scrapper/
├── scheduler/              # Production automation scripts
│   ├── scrapper.go        # Main scraper dengan SQL generation
│   ├── execute_sql.go     # SQL executor ke Supabase
│   ├── generate_sql.go    # Standalone SQL generator
│   ├── run_scraper.sh     # Automation runner script
│   ├── setup.sh           # Quick setup script
│   ├── .env.example       # Environment variables template
│   ├── go.mod             # Go dependencies
│   ├── go.sum             # Go dependencies lock
│   ├── logs/              # Log files directory
│   ├── QUICKSTART.md      # Quick start guide
│   └── README_SETUP.md    # Detailed setup guide
│
├── scrapper/              # Development & testing
│   ├── scrapper.go        # Standalone scraper
│   ├── harga_emas.json    # Output JSON
│   ├── update_gold_prices.sql  # Generated SQL
│   ├── go.mod
│   └── go.sum
│
├── .gitignore
└── README.md              # This file
```

## 🔄 How It Works

```mermaid
graph TD
    A[Cron Trigger] --> B[Launch run_scraper.sh]
    B --> C[Run scrapper.go]
    C --> D[Headless Browser Load Page]
    D --> E[Parse HTML with XPath]
    E --> F[Filter by Weight & Brand]
    F --> G[Generate JSON]
    G --> H[Generate SQL]
    H --> I[Run execute_sql.go]
    I --> J[Connect to Supabase]
    J --> K[Execute UPDATE queries]
    K --> L[Save Logs]
    L --> M[Cleanup Old Logs]
```

### Process Flow

1. **Scraping**: Menggunakan `chromedp` untuk render JavaScript dan scrape data
2. **Parsing**: Extract data dengan `htmlquery` XPath
3. **Filtering**: Hanya ambil gramasi: 0.5, 1, 2, 5, 10, 25, 50, 100
4. **JSON Export**: Simpan ke `harga_emas.json`
5. **SQL Generation**: Generate UPDATE queries dengan timestamp
6. **Database Update**: Execute ke Supabase PostgreSQL
7. **Logging**: Save semua aktivitas ke log files

## 📊 Data Structure

### JSON Output Example
```json
[
  {
    "brand": "GALERI 24",
    "data": [
      {
        "berat": "1",
        "harga_jual": "2368000",
        "harga_buyback": "2212000"
      }
    ]
  }
]
```

### SQL Output Example
```sql
-- SQL UPDATE Queries untuk Gold Prices
-- Generated on: 2025-11-06 13:00:53

UPDATE public.gold_prices_v2
SET price_buyback=2212000.0, price_sell=2368000
WHERE "date"='2025-11-06' AND brand='Galeri24' AND denom=1.0;
```

## ⏰ Scheduling

### Cron Schedule Examples

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

### Cron Format
```
* * * * * command
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, Sunday = 0 or 7)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

## 🔧 API Reference

### Main Functions

#### `scrapper.go`
```go
func fetchRenderedHTML(url string) (string, error)
// Mengambil HTML dengan chromedp headless browser

func generateSQL(allBrandsData []BrandData) error
// Generate SQL UPDATE queries dari data

func showLoadingAnimation(done chan bool, startTime time.Time)
// Tampilkan loading spinner dengan elapsed time
```

#### `execute_sql.go`
```go
func connectSupabase(config SupabaseConfig) (*sql.DB, error)
// Koneksi ke Supabase PostgreSQL

func executeSQLFile(db *sql.DB, filename string) (int, error)
// Eksekusi SQL file ke database
```

## 🐛 Troubleshooting

### Cron tidak jalan?

```bash
# Check cron service
sudo launchctl list | grep cron

# Check system logs
log show --predicate 'process == "cron"' --last 1h

# Verify cron is installed
crontab -l
```

### Database connection error?

```bash
# Test connection dengan psql
psql "postgresql://postgres:PASSWORD@HOST:5432/postgres"

# Test in Go
cd scheduler
export $(cat .env | xargs)
go run execute_sql.go
```

### Scraper tidak mengambil data?

1. Check website masih online: https://galeri24.co.id/harga-emas
2. Verifikasi XPath selector masih valid
3. Check Chrome/Chromium terinstall
4. Lihat logs untuk error details

### Path issues di cron?

Edit `run_scraper.sh` dan tambahkan full path:
```bash
export PATH=/usr/local/go/bin:/usr/bin:/bin
```

## 📈 Performance

- **Scraping time**: ~10-20 detik (tergantung koneksi)
- **SQL execution**: ~1-2 detik untuk 24 queries
- **Total runtime**: ~15-25 detik
- **Memory usage**: ~50-100 MB
- **Log file size**: ~5-10 KB per run

## 🔐 Security Best Practices

- ✅ Jangan commit file `.env` 
- ✅ Gunakan environment variables untuk credentials
- ✅ Set proper file permissions (`chmod 600 .env`)
- ✅ Rotate password secara berkala
- ✅ Limit database permissions hanya untuk operasi yang diperlukan
- ✅ Monitor logs untuk aktivitas mencurigakan

## 📝 Logging

Logs disimpan di `scheduler/logs/` dengan format:
```
scraper_YYYYMMDD_HHMMSS.log
```

Auto cleanup: Log files > 30 hari otomatis dihapus.

## 🧪 Testing

```bash
# Test scraper
cd scrapper
go run scrapper.go

# Test SQL generation
cd scheduler
go run generate_sql.go

# Test SQL execution
cd scheduler
export $(cat .env | xargs)
go run execute_sql.go

# Test full automation
cd scheduler
./run_scraper.sh
```

## 📚 Documentation

- [Quick Start Guide](scheduler/QUICKSTART.md) - Setup cepat 5 menit
- [Detailed Setup Guide](scheduler/README_SETUP.md) - Dokumentasi lengkap

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Your Name**
- GitHub: [@yourusername](https://github.com/yourusername)
- Email: your.email@example.com

## 🙏 Acknowledgments

- [chromedp](https://github.com/chromedp/chromedp) - Headless browser automation
- [htmlquery](https://github.com/antchfx/htmlquery) - XPath query for HTML
- [Supabase](https://supabase.com/) - Backend as a Service
- [Galeri24](https://galeri24.co.id/) - Gold price data source

## 📞 Support

Jika ada pertanyaan atau masalah:
1. Check [Troubleshooting](#troubleshooting) section
2. Review logs di `scheduler/logs/`
3. Open an [Issue](https://github.com/yourusername/gold-scrapper/issues)

## 🗺️ Roadmap

- [ ] Add support untuk brand emas lainnya
- [ ] Implement webhook notifications
- [ ] Add GraphQL API
- [ ] Create web dashboard
- [ ] Add price prediction with ML
- [ ] Support export ke CSV/Excel
- [ ] Add email alerts untuk price changes

---

⭐ **Star this repo** if you find it useful!

Made with ❤️ by [Your Name](https://github.com/yourusername)
