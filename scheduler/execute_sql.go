package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Constants
const (
	dockerSQLPath = "/app/sql/update_gold_prices.sql"
	localSQLPath  = "../sql/update_gold_prices.sql"
	rootSQLPath   = "sql/update_gold_prices.sql"
)

type SupabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func getSupabaseConfig() SupabaseConfig {
	return SupabaseConfig{
		Host:     getEnv("SUPABASE_HOST", ""),
		Port:     getEnv("SUPABASE_PORT", "5432"),
		User:     getEnv("SUPABASE_USER", "postgres"),
		Password: getEnv("SUPABASE_PASSWORD", ""),
		DBName:   getEnv("SUPABASE_DB", "postgres"),
		SSLMode:  getEnv("SUPABASE_SSL_MODE", "require"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getSQLFilePath() string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return dockerSQLPath
	}
	
	if os.Getenv("IS_DOCKER") == "true" {
		return dockerSQLPath
	}
	
	possiblePaths := []string{localSQLPath, rootSQLPath, dockerSQLPath}
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return localSQLPath
}

func connectSupabase(config SupabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func executeSQLFile(db *sql.DB, filename string) (int, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("gagal membaca file %s: %v", filename, err)
	}

	statements := strings.Split(string(content), ";")
	executedCount := 0
	failedCount := 0

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		if !strings.Contains(stmt, "UPDATE") && !strings.Contains(stmt, "INSERT") {
			continue
		}

		stmt = strings.ReplaceAll(stmt, "gold_prices_v2", "gold_prices_v2")

		if _, err := db.Exec(stmt); err != nil {
			log.Printf("⚠️  Error executing statement %d: %v", i+1, err)
			log.Printf("Statement: %s", stmt[:min(len(stmt), 100)])
			failedCount++
			continue
		}

		executedCount++

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

	config := getSupabaseConfig()

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

	db, err := connectSupabase(config)
	if err != nil {
		log.Fatalf("❌ Gagal koneksi ke Supabase: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Koneksi berhasil!")

	sqlFile := getSQLFilePath()
	
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
