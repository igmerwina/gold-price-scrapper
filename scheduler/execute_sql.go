package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Config untuk koneksi Supabase
type SupabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// getSupabaseConfig membaca konfigurasi dari environment variables atau .env
func getSupabaseConfig() SupabaseConfig {
	// Default values
	config := SupabaseConfig{
		Host:     getEnv("SUPABASE_HOST", ""),
		Port:     getEnv("SUPABASE_PORT", "5432"),
		User:     getEnv("SUPABASE_USER", "postgres"),
		Password: getEnv("SUPABASE_PASSWORD", ""),
		DBName:   getEnv("SUPABASE_DB", "postgres"),
		SSLMode:  getEnv("SUPABASE_SSL_MODE", "require"),
	}

	return config
}

// getEnv mendapatkan environment variable dengan default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// connectSupabase membuat koneksi ke Supabase PostgreSQL
func connectSupabase(config SupabaseConfig) (*sql.DB, error) {
	// Format connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	// Buka koneksi
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi: %v", err)
	}

	// Test koneksi
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("gagal ping database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// executeSQLFile membaca dan mengeksekusi file SQL
func executeSQLFile(db *sql.DB, filename string) (int, error) {
	// Baca file SQL
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("gagal membaca file %s: %v", filename, err)
	}

	sqlContent := string(content)

	// Split SQL statements (sederhana, asumsi setiap statement diakhiri dengan ;)
	statements := strings.Split(sqlContent, ";")
	
	executedCount := 0
	failedCount := 0

	// Eksekusi setiap statement
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		
		// Skip empty statements dan comments
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// Skip pure comment lines
		if !strings.Contains(stmt, "UPDATE") && !strings.Contains(stmt, "INSERT") {
			continue
		}

		// Replace gold_prices_v2 with gold_prices_v3 if found
		stmt = strings.ReplaceAll(stmt, "gold_prices_v2", "gold_prices_v3")

		// Eksekusi statement
		_, err := db.Exec(stmt)
		if err != nil {
			log.Printf("⚠️  Error executing statement %d: %v", i+1, err)
			log.Printf("Statement: %s", stmt[:min(len(stmt), 100)])
			failedCount++
			continue
		}

		executedCount++

		// Log progress setiap 10 queries
		if executedCount%10 == 0 {
			fmt.Printf("   Progress: %d queries executed...\n", executedCount)
		}
	}

	if failedCount > 0 {
		return executedCount, fmt.Errorf("%d queries failed", failedCount)
	}

	return executedCount, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	startTime := time.Now()
	fmt.Println("🚀 Gold Price SQL Executor - Supabase")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏰ Waktu: %s\n\n", startTime.Format("2006-01-02 15:04:05"))

	// Load konfigurasi
	config := getSupabaseConfig()

	// Validasi konfigurasi
	if config.Host == "" {
		log.Fatal("❌ SUPABASE_HOST tidak ditemukan. Set environment variable terlebih dahulu.")
	}
	if config.Password == "" {
		log.Fatal("❌ SUPABASE_PASSWORD tidak ditemukan. Set environment variable terlebih dahulu.")
	}

	fmt.Println("📡 Menghubungkan ke Supabase...")
	fmt.Printf("   Host: %s\n", config.Host)
	fmt.Printf("   Database: %s\n", config.DBName)
	fmt.Printf("   User: %s\n\n", config.User)

	// Koneksi ke database
	db, err := connectSupabase(config)
	if err != nil {
		log.Fatalf("❌ Gagal koneksi ke Supabase: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Koneksi berhasil!")

	// Eksekusi SQL file
	sqlFile := "../sql/update_gold_prices.sql"
	
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		log.Fatalf("❌ File %s tidak ditemukan", sqlFile)
	}

	fmt.Printf("\n📝 Mengeksekusi file: %s\n", sqlFile)
	
	executedCount, err := executeSQLFile(db, sqlFile)
	
	duration := time.Since(startTime)

	fmt.Println("\n" + strings.Repeat("=", 60))
	if err != nil {
		fmt.Printf("⚠️  Selesai dengan error: %v\n", err)
		fmt.Printf("✅ Berhasil: %d queries\n", executedCount)
	} else {
		fmt.Println("✅ EKSEKUSI BERHASIL!")
		fmt.Printf("📊 Total queries dieksekusi: %d\n", executedCount)
	}
	fmt.Printf("⏱️  Waktu eksekusi: %.2f detik\n", duration.Seconds())
	fmt.Printf("⏰ Selesai: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))
}
