package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chromedp/chromedp"
)

// GoldData merepresentasikan struktur data emas per berat
type GoldData struct {
	Berat       string `json:"berat"`
	HargaJual   string `json:"harga_jual"`
	HargaBuyback string `json:"harga_buyback"`
}

// BrandData merepresentasikan struktur data untuk satu merk emas
type BrandData struct {
	Brand string     `json:"brand"`
	Data  []GoldData `json:"data"`
}

const url = "https://galeri24.co.id/harga-emas"

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

// generateSQL membuat SQL UPDATE queries dari data JSON
func generateSQL(allBrandsData []BrandData) error {
	today := time.Now().Format("2006-01-02")
	now := time.Now()
	generatedTime := now.Format("2006-01-02 15:04:05")

	// Get table name from environment variable
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		tableName = "gold_prices_v3" // Default fallback
	}

	fmt.Println("\n🔄 Membuat SQL UPDATE queries...")
	fmt.Printf("📊 Target table: %s\n", tableName)
	
	queryCount := 0
	// Buka file untuk menulis SQL
	sqlFile, err := ioutil.TempFile("", "sql_*.txt")
	if err != nil {
		return fmt.Errorf("gagal membuat temp file: %v", err)
	}
	defer sqlFile.Close()

	sqlContent := fmt.Sprintf("-- SQL UPDATE Queries untuk Gold Prices\n-- Generated on: %s\n\n", generatedTime)

	for _, brandData := range allBrandsData {
		brand := brandData.Brand
		
		brandSQL := brand
		switch brand {
		case "GALERI 24":
			brandSQL = "Galeri24"
		case "ANTAM":
			brandSQL = "Antam"
		case "UBS":
			brandSQL = "UBS"
		}

		for _, gold := range brandData.Data {
			denom, err := strconv.ParseFloat(gold.Berat, 64)
			if err != nil {
				log.Printf("Warning: Gagal parse berat '%s': %v", gold.Berat, err)
				continue
			}

			priceSell, err := strconv.ParseFloat(gold.HargaJual, 64)
			if err != nil {
				log.Printf("Warning: Gagal parse harga jual '%s': %v", gold.HargaJual, err)
				continue
			}

			priceBuyback, err := strconv.ParseFloat(gold.HargaBuyback, 64)
			if err != nil {
				log.Printf("Warning: Gagal parse harga buyback '%s': %v", gold.HargaBuyback, err)
				continue
			}

			sqlContent += fmt.Sprintf("UPDATE public.%s\n", tableName)
			sqlContent += fmt.Sprintf("SET price_buyback=%.1f, price_sell=%.0f\n", priceBuyback, priceSell)
			sqlContent += fmt.Sprintf("WHERE \"date\"='%s' AND brand='%s' AND denom=%.1f;\n\n", today, brandSQL, denom)
			queryCount++
		}
	}

	sqlContent += fmt.Sprintf("-- Total %d queries generated successfully\n", queryCount)

	// Tulis ke file update_gold_prices.sql
	err = ioutil.WriteFile("../sql/update_gold_prices.sql", []byte(sqlContent), 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis SQL file: %v", err)
	}

	fmt.Printf("✅ %d SQL queries berhasil dibuat dan disimpan ke sql/update_gold_prices.sql\n", queryCount)
	// fmt.Println("\n--- Preview SQL Queries ---")
	
	// // Tampilkan beberapa baris pertama
	// lines := strings.Split(sqlContent, "\n")
	// previewLines := 20
	// if len(lines) < previewLines {
	// 	previewLines = len(lines)
	// }
	// for i := 0; i < previewLines; i++ {
	// 	fmt.Println(lines[i])
	// }
	// if len(lines) > previewLines {
	// 	fmt.Printf("... dan %d baris lainnya\n", len(lines)-previewLines)
	// }

	return nil
}

func main() {
	startTime := time.Now()
	fmt.Println("🚀 Memulai proses scraping...")
	fmt.Printf("⏰ Waktu mulai: %s\n\n", startTime.Format("2006-01-02 15:04:05"))

	// 1. Ambil konten HTML yang sudah di-render menggunakan chromedp
	fmt.Println("🔄 Memuat halaman dengan headless browser...")
	stepStart := time.Now()
	
	htmlContent, err := fetchRenderedHTML(url)
	if err != nil {
		log.Fatalf("Gagal memuat URL dengan chromedp: %v", err)
	}

	stepDuration := time.Since(stepStart)
	fmt.Printf("✅ Halaman berhasil dimuat (%.2f detik)\n", stepDuration.Seconds())

	// 2. Parse HTML dengan htmlquery
	fmt.Println("\n🔄 Parsing HTML dan ekstraksi data...")
	stepStart = time.Now()
	
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatalf("Gagal parse HTML: %v", err)
	}

	allBrandsData := []BrandData{}

	// Daftar XPath untuk menemukan div rows berdasarkan ID vendor
	// Struktur: <div id="GALERI 24"> berisi grid dengan div row yang berisi col-span-1, col-span-2, col-span-2
	xpathSelectors := map[string]string{
		"GALERI 24":      "//div[@id='GALERI 24']//div[contains(@class, 'grid-cols-5') and contains(@class, 'divide-x') and contains(@class, 'lg:hover:bg-neutral-50')]",
		"ANTAM":          "//div[@id='ANTAM']//div[contains(@class, 'grid-cols-5') and contains(@class, 'divide-x') and contains(@class, 'lg:hover:bg-neutral-50')]",
		"UBS":            "//div[@id='UBS']//div[contains(@class, 'grid-cols-5') and contains(@class, 'divide-x') and contains(@class, 'lg:hover:bg-neutral-50')]",
	}

	for brand, xpath := range xpathSelectors {
		nodes := htmlquery.Find(doc, xpath)
		data := []GoldData{}

		for _, node := range nodes {
			// Kolom 1: Berat (col-span-1)
			beratNode := htmlquery.FindOne(node, ".//div[contains(@class, 'col-span-1')]")
			berat := htmlquery.InnerText(beratNode)
			berat = strings.TrimSpace(berat)
			
			// Kolom 2: Harga Jual (col-span-2, posisi ke-2)
			hargaJualNodes := htmlquery.Find(node, ".//div[contains(@class, 'col-span-2')]")
			hargaJual := ""
			if len(hargaJualNodes) >= 1 {
				hargaJual = htmlquery.InnerText(hargaJualNodes[0])
				hargaJual = cleanPrice(hargaJual)
			}

			// Kolom 3: Harga Buyback (col-span-2, posisi ke-3)
			hargaBuyback := ""
			if len(hargaJualNodes) >= 2 {
				hargaBuyback = htmlquery.InnerText(hargaJualNodes[1])
				hargaBuyback = cleanPrice(hargaBuyback)
			}

			// Filter hanya gramasi tertentu
			allowedWeights := []string{"0.5", "1", "2", "5", "10", "25", "50", "100"}
			isAllowed := false
			beratClean := strings.TrimSpace(strings.ToLower(berat))
			
			// Cek apakah berat termasuk dalam daftar yang diizinkan
			for _, weight := range allowedWeights {
				if strings.Contains(beratClean, weight+" gr") || 
				   strings.Contains(beratClean, weight+" gram") ||
				   beratClean == weight+"gr" ||
				   beratClean == weight+" gr" ||
				   beratClean == weight {
					isAllowed = true
					break
				}
			}
			
			// Hanya proses jika Berat dan Harga terdeteksi dan termasuk gramasi yang diizinkan
			if berat != "" && hargaJual != "" && !strings.Contains(strings.ToLower(berat), "berat") && isAllowed {
				data = append(data, GoldData{
					Berat:        berat,
					HargaJual:    hargaJual,
					HargaBuyback: hargaBuyback,
				})
			}
		}

		// Tambahkan data brand ke slice utama jika ada data
		if len(data) > 0 {
			allBrandsData = append(allBrandsData, BrandData{
				Brand: brand,
				Data:  data,
			})
		} else {
			fmt.Printf("⚠️  Tidak ada data ditemukan untuk brand %s\n", brand)
		}
	}

	stepDuration = time.Since(stepStart)
	fmt.Printf("✅ Parsing dan ekstraksi selesai (%.2f detik)\n", stepDuration.Seconds())

	// 3. Export ke JSON
	fmt.Println("\n🔄 Menyimpan data ke JSON...")
	stepStart = time.Now()
	
	jsonData, err := json.MarshalIndent(allBrandsData, "", "  ")
	if err != nil {
		log.Fatalf("Gagal meng-encode ke JSON: %v", err)
	}

	// 4. Tulis output ke file
	err = ioutil.WriteFile("../sql/harga_emas.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Gagal menulis ke file: %v", err)
	}

	stepDuration = time.Since(stepStart)
	fmt.Printf("✅ Data harga emas berhasil disimpan ke sql/harga_emas.json (%.2f detik)\n", stepDuration.Seconds())
	// fmt.Println("\n--- Tampilan Hasil JSON ---")
	// fmt.Println(string(jsonData))

	// 5. Generate SQL queries
	stepStart = time.Now()
	err = generateSQL(allBrandsData)
	if err != nil {
		log.Fatalf("Gagal generate SQL: %v", err)
	}
	stepDuration = time.Since(stepStart)

	// Summary
	totalDuration := time.Since(startTime)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ PROSES SELESAI!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📁 File yang dibuat:")
	fmt.Println("   - sql/harga_emas.json")
	fmt.Println("   - sql/update_gold_prices.sql")
	fmt.Println()
	fmt.Printf("⏱️  Total waktu eksekusi: %.2f detik (%.2f menit)\n", totalDuration.Seconds(), totalDuration.Minutes())
	fmt.Printf("⏰ Waktu selesai: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))
}