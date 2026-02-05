package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	host := os.Getenv("SUPABASE_HOST")
	port := os.Getenv("SUPABASE_PORT")
	user := os.Getenv("SUPABASE_USER")
	password := os.Getenv("SUPABASE_PASSWORD")
	dbname := os.Getenv("SUPABASE_DB")
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		tableName = "gold_prices_v2"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var maxID int64
	query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM public.%s", tableName)
	err = db.QueryRow(query).Scan(&maxID)
	if err != nil {
		log.Fatal(err)
	}

	// Print hanya angka untuk mudah di-parse
	fmt.Print(maxID)
}
