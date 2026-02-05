package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
)

// GoldData merepresentasikan struktur data emas per berat
type GoldData struct {
	Berat        string `json:"berat"`
	HargaJual    string `json:"harga_jual"`
	HargaBuyback string `json:"harga_buyback"`
}

// BrandData merepresentasikan struktur data untuk satu merk emas
type BrandData struct {
	Brand string     `json:"brand"`
	Data  []GoldData `json:"data"`
}

const (
	url            = "https://galeri24.co.id/harga-emas"
	dockerJSONPath = "/app/sql/harga_emas.json"
	dockerSQLPath  = "/app/sql/update_gold_prices.sql"
	localJSONPath  = "../sql/harga_emas.json"
	localSQLPath   = "../sql/update_gold_prices.sql"
	dateFormat     = "2006-01-02"
	timeFormat     = "2006-01-02 15:04:05"
	defaultTable   = "gold_prices_v2"
)

// isDockerEnvironment mengecek apakah sedang berjalan di Docker
func isDockerEnvironment() bool {
	// Cek file /.dockerenv
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Cek environment variable
	if os.Getenv("IS_DOCKER") == "true" {
		return true
	}

	return false
}

// getJSONPath mengembalikan path JSON yang sesuai dengan environment
func getJSONPath() string {
	if isDockerEnvironment() {
		return dockerJSONPath
	}
	return localJSONPath
}

// getSQLPath mengembalikan path SQL yang sesuai dengan environment
func getSQLPath() string {
	if isDockerEnvironment() {
		return dockerSQLPath
	}
	return localSQLPath
}

// cleanPrice menghilangkan karakter non-digit kecuali koma dan mengganti koma dengan titik (jika ada)
func cleanPrice(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "Rp", "")
	s = strings.ReplaceAll(s, ".", "") // Hapus titik sebagai pemisah ribuan
	s = strings.TrimSpace(s)
	return s
}

// showLoadingAnimation menampilkan animasi loading dengan spinner dan elapsed time
func showLoadingAnimation(done chan bool, startTime time.Time) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r\033[K") // Clear the line
			return
		default:
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\r   %s Memuat halaman... (%.1fs)", spinner[i%len(spinner)], elapsed)
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// fetchRenderedHTML menggunakan chromedp untuk mengambil HTML yang sudah di-render oleh JavaScript
func fetchRenderedHTML(url string) (string, error) {
	// Buat context dengan timeout yang lebih lama
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Buat chromedp context dengan opsi headless
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var htmlContent string

	// Start loading animation
	loadingStart := time.Now()
	done := make(chan bool)
	go showLoadingAnimation(done, loadingStart)

	// Jalankan chromedp tasks
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Tunggu lebih lama untuk load JavaScript
		chromedp.Sleep(8*time.Second),
		// Ambil HTML yang sudah di-render
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)

	// Stop loading animation
	done <- true
	time.Sleep(200 * time.Millisecond) // Wait a bit to ensure spinner is cleared

	if err != nil {
		return "", fmt.Errorf("chromedp error: %v", err)
	}

	return htmlContent, nil
}

// generateSQL membuat SQL INSERT/UPDATE queries dari data JSON
func generateSQL(allBrandsData []BrandData) error {
	today := time.Now().Format(dateFormat)
	generatedTime := time.Now().Format(timeFormat)

	tableName := getTableName()
	fmt.Println("\n🔄 Membuat SQL INSERT/UPDATE queries...")
	fmt.Printf("📊 Target table: %s\n", tableName)

	// Query last ID dari database via helper script
	fmt.Println("🔍 Mengambil last ID dari database...")
	lastID, err := getLastID(tableName)
	if err != nil {
		log.Printf("⚠️  Gagal mendapatkan last ID: %v. Menggunakan nilai default 0", err)
		lastID = 0
	}
	fmt.Printf("✅ Last ID dari database: %d\n", lastID)

	sqlContent := fmt.Sprintf("-- SQL INSERT/UPDATE Queries untuk Gold Prices\n-- Generated on: %s\n\n", generatedTime)
	sqlContent += fmt.Sprintf("-- Last ID: %d, Next INSERT akan dimulai dari ID %d\n\n", lastID, lastID+1)

	insertCount, updateCount := buildSQLQueriesWithInsert(&sqlContent, allBrandsData, tableName, today, lastID)

	sqlContent += fmt.Sprintf("\n-- Total %d INSERT queries, %d UPDATE queries generated successfully\n", insertCount, updateCount)

	if err := os.WriteFile(getSQLPath(), []byte(sqlContent), 0644); err != nil {
		return fmt.Errorf("gagal menulis SQL file: %v", err)
	}

	fmt.Printf("✅ SQL queries berhasil dibuat dan disimpan ke %s\n", getSQLPath())
	fmt.Printf("   📊 INSERT: %d | UPDATE: %d queries\n", insertCount, updateCount)
	fmt.Printf("   🔢 INSERT ID range: %d - %d\n", lastID+1, lastID+int64(insertCount))
	return nil
}

func getTableName() string {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		tableName = defaultTable
	}
	return tableName
}

func buildSQLQueriesWithInsert(sqlContent *string, allBrandsData []BrandData, tableName, today string, lastID int64) (int, int) {
	insertCount := 0
	updateCount := 0
	currentID := lastID + 1

	for _, brandData := range allBrandsData {
		brandSQL := normalizeBrandName(brandData.Brand)
		for _, gold := range brandData.Data {
			insertQuery, updateQuery := buildInsertAndUpdateQueries(gold, tableName, today, brandSQL, currentID)

			if insertQuery != "" {
				*sqlContent += insertQuery
				insertCount++
				currentID++ // Increment ID untuk INSERT berikutnya
			}
			if updateQuery != "" {
				*sqlContent += updateQuery
				updateCount++
			}
		}
	}
	return insertCount, updateCount
}

func normalizeBrandName(brand string) string {
	switch brand {
	case "GALERI 24":
		return "Galeri24"
	case "ANTAM":
		return "Antam"
	case "UBS":
		return "UBS"
	default:
		return brand
	}
}

func buildInsertAndUpdateQueries(gold GoldData, tableName, today, brandSQL string, insertID int64) (string, string) {
	denom, err := strconv.ParseFloat(gold.Berat, 64)
	if err != nil {
		log.Printf("Warning: Gagal parse berat '%s': %v", gold.Berat, err)
		return "", ""
	}

	priceSell, err := strconv.ParseFloat(gold.HargaJual, 64)
	if err != nil {
		log.Printf("Warning: Gagal parse harga jual '%s': %v", gold.HargaJual, err)
		return "", ""
	}

	priceBuyback, err := strconv.ParseFloat(gold.HargaBuyback, 64)
	if err != nil {
		log.Printf("Warning: Gagal parse harga buyback '%s': %v", gold.HargaBuyback, err)
		return "", ""
	}

	// INSERT query dengan ID manual (tidak pakai auto-increment)
	insertQuery := fmt.Sprintf(
		"INSERT INTO public.%s (id, \"date\", brand, denom, price_buyback, price_sell)\n"+
			"SELECT %d, '%s', '%s', %.1f, %.1f, %.0f\n"+
			"WHERE NOT EXISTS (\n"+
			"  SELECT 1 FROM public.%s \n"+
			"  WHERE \"date\"='%s' AND brand='%s' AND denom=%.1f\n"+
			");\n\n",
		tableName, insertID, today, brandSQL, denom, priceBuyback, priceSell,
		tableName, today, brandSQL, denom)

	// UPDATE query
	updateQuery := fmt.Sprintf(
		"UPDATE public.%s\n"+
			"SET price_buyback=%.1f, price_sell=%.0f\n"+
			"WHERE \"date\"='%s' AND brand='%s' AND denom=%.1f;\n\n",
		tableName, priceBuyback, priceSell, today, brandSQL, denom)

	return insertQuery, updateQuery
}

// loadEnv memuat environment variables dari file .env
func loadEnv() {
	envPath := "../scheduler/.env"
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️  Tidak dapat memuat file .env: %v", err)
		log.Printf("   Pastikan file %s sudah dibuat dan diisi", envPath)
	}
}

// getLastID mengambil ID terakhir dari table gold_prices_v2 via helper script
func getLastID(tableName string) (int64, error) {
	// Check if get_last_id binary exists
	// Try Docker path first (/app/get_last_id), then local path (../scheduler/get_last_id)
	var cmd *exec.Cmd
	dockerPath := "./get_last_id"
	localPath := "../scheduler/get_last_id"
	
	if _, err := os.Stat(dockerPath); err == nil {
		// Docker environment - binary in same directory
		cmd = exec.Command(dockerPath)
	} else if _, err := os.Stat(localPath); err == nil {
		// Local environment - compiled binary exists
		cmd = exec.Command(localPath)
	} else {
		// Local development - use go run
		cmd = exec.Command("go", "run", "get_last_id.go")
		cmd.Dir = "../scheduler"
	}
	
	cmd.Env = append(os.Environ(), fmt.Sprintf("TABLE_NAME=%s", tableName))

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return 0, fmt.Errorf("gagal run get_last_id script: %v, stderr: %s", err, string(exitErr.Stderr))
		}
		return 0, fmt.Errorf("gagal run get_last_id script: %v", err)
	}

	var lastID int64
	_, err = fmt.Sscanf(string(output), "%d", &lastID)
	if err != nil {
		return 0, fmt.Errorf("gagal parse last ID '%s': %v", string(output), err)
	}

	return lastID, nil
}

func main() {
	startTime := time.Now()
	fmt.Println("🚀 Memulai proses scraping...")
	fmt.Printf("⏰ Waktu mulai: %s\n\n", startTime.Format(timeFormat))

	loadEnv()

	htmlContent := fetchHTML()
	allBrandsData := parseHTML(htmlContent)
	saveJSON(allBrandsData)
	generateSQL(allBrandsData)

	printSummary(startTime)
}

func fetchHTML() string {
	fmt.Println("🔄 Memuat halaman dengan headless browser...")
	stepStart := time.Now()

	htmlContent, err := fetchRenderedHTML(url)
	if err != nil {
		log.Fatalf("Gagal memuat URL dengan chromedp: %v", err)
	}

	fmt.Printf("✅ Halaman berhasil dimuat (%.2f detik)\n", time.Since(stepStart).Seconds())
	return htmlContent
}

func parseHTML(htmlContent string) []BrandData {
	fmt.Println("\n🔄 Parsing HTML dan ekstraksi data...")
	stepStart := time.Now()

	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatalf("Gagal parse HTML: %v", err)
	}

	allBrandsData := extractBrandsData(doc)

	fmt.Printf("✅ Parsing dan ekstraksi selesai (%.2f detik)\n", time.Since(stepStart).Seconds())
	return allBrandsData
}

func extractBrandsData(doc *html.Node) []BrandData {
	xpathSelectors := map[string]string{
		"GALERI 24": "//div[@id='GALERI 24']//div[contains(@class, 'grid-cols-5') and contains(@class, 'divide-x') and contains(@class, 'lg:hover:bg-neutral-50')]",
		"ANTAM":     "//div[@id='ANTAM']//div[contains(@class, 'grid-cols-5') and contains(@class, 'divide-x') and contains(@class, 'lg:hover:bg-neutral-50')]",
		"UBS":       "//div[@id='UBS']//div[contains(@class, 'grid-cols-5') and contains(@class, 'divide-x') and contains(@class, 'lg:hover:bg-neutral-50')]",
	}

	var allBrandsData []BrandData
	for brand, xpath := range xpathSelectors {
		data := extractBrandData(htmlquery.Find(doc, xpath))
		if len(data) > 0 {
			allBrandsData = append(allBrandsData, BrandData{Brand: brand, Data: data})
		} else {
			fmt.Printf("⚠️  Tidak ada data ditemukan untuk brand %s\n", brand)
		}
	}
	return allBrandsData
}

func extractBrandData(nodes []*html.Node) []GoldData {
	var data []GoldData
	allowedWeights := []string{"0.5", "1", "2", "5", "10", "25", "50", "100"}

	for _, node := range nodes {
		gold := extractGoldData(node)
		if isValidGoldData(gold, allowedWeights) {
			data = append(data, gold)
		}
	}
	return data
}

func extractGoldData(node *html.Node) GoldData {
	beratNode := htmlquery.FindOne(node, ".//div[contains(@class, 'col-span-1')]")
	berat := strings.TrimSpace(htmlquery.InnerText(beratNode))

	hargaNodes := htmlquery.Find(node, ".//div[contains(@class, 'col-span-2')]")

	hargaJual := ""
	if len(hargaNodes) >= 1 {
		hargaJual = cleanPrice(htmlquery.InnerText(hargaNodes[0]))
	}

	hargaBuyback := ""
	if len(hargaNodes) >= 2 {
		hargaBuyback = cleanPrice(htmlquery.InnerText(hargaNodes[1]))
	}

	return GoldData{
		Berat:        berat,
		HargaJual:    hargaJual,
		HargaBuyback: hargaBuyback,
	}
}

func isValidGoldData(gold GoldData, allowedWeights []string) bool {
	if gold.Berat == "" || gold.HargaJual == "" || strings.Contains(strings.ToLower(gold.Berat), "berat") {
		return false
	}

	beratClean := strings.TrimSpace(strings.ToLower(gold.Berat))
	for _, weight := range allowedWeights {
		if strings.Contains(beratClean, weight+" gr") ||
			strings.Contains(beratClean, weight+" gram") ||
			beratClean == weight+"gr" ||
			beratClean == weight+" gr" ||
			beratClean == weight {
			return true
		}
	}
	return false
}

func saveJSON(allBrandsData []BrandData) {
	fmt.Println("\n🔄 Menyimpan data ke JSON...")
	stepStart := time.Now()

	jsonData, err := json.MarshalIndent(allBrandsData, "", "  ")
	if err != nil {
		log.Fatalf("Gagal meng-encode ke JSON: %v", err)
	}

	sqlDir := getSQLDir()
	os.MkdirAll(sqlDir, 0755)

	if err := os.WriteFile(getJSONPath(), jsonData, 0644); err != nil {
		log.Fatalf("Gagal menulis ke file: %v", err)
	}

	fmt.Printf("✅ Data harga emas berhasil disimpan ke %s (%.2f detik)\n", getJSONPath(), time.Since(stepStart).Seconds())
}

func getSQLDir() string {
	if isDockerEnvironment() {
		return "/app/sql"
	}
	return "../sql"
}

func printSummary(startTime time.Time) {
	totalDuration := time.Since(startTime)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ PROSES SELESAI!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📁 File yang dibuat:")
	fmt.Println("   - sql/harga_emas.json")
	fmt.Println("   - sql/update_gold_prices.sql")
	fmt.Println()
	fmt.Printf("⏱️  Total waktu eksekusi: %.2f detik (%.2f menit)\n", totalDuration.Seconds(), totalDuration.Minutes())
	fmt.Printf("⏰ Waktu selesai: %s\n", time.Now().Format(timeFormat))
	fmt.Println(strings.Repeat("=", 60))
}
