package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

type GoldData struct {
	Berat        string `json:"berat"`
	HargaJual    string `json:"harga_jual"`
	HargaBuyback string `json:"harga_buyback"`
}

type BrandData struct {
	Brand string     `json:"brand"`
	Data  []GoldData `json:"data"`
}

func main() {
	// Baca file JSON
	jsonFile, err := ioutil.ReadFile("harga_emas.json")
	if err != nil {
		log.Fatalf("Gagal membaca file: %v", err)
	}

	var brands []BrandData
	err = json.Unmarshal(jsonFile, &brands)
	if err != nil {
		log.Fatalf("Gagal parse JSON: %v", err)
	}

	// Buat map untuk menyimpan berat per brand
	brandWeights := make(map[string]map[string]bool)
	
	for _, brand := range brands {
		brandWeights[brand.Brand] = make(map[string]bool)
		for _, item := range brand.Data {
			brandWeights[brand.Brand][item.Berat] = true
		}
	}

	// Tampilkan berat per brand
	fmt.Println("📊 BERAT PER BRAND:")
	fmt.Println(strings.Repeat("=", 60))
	for _, brand := range brands {
		fmt.Printf("\n%s (%d berat):\n", brand.Brand, len(brand.Data))
		weights := []string{}
		for _, item := range brand.Data {
			weights = append(weights, item.Berat)
		}
		fmt.Printf("  %s\n", strings.Join(weights, ", "))
	}

	// Cari berat yang ada di semua brand (4 brand)
	fmt.Println("\n\n🔍 BERAT YANG ADA DI SEMUA 4 BRAND:")
	fmt.Println(strings.Repeat("=", 60))
	commonInAll := []string{}
	
	// Ambil semua berat dari brand pertama sebagai referensi
	for weight := range brandWeights[brands[0].Brand] {
		foundInAll := true
		for _, brand := range brands {
			if !brandWeights[brand.Brand][weight] {
				foundInAll = false
				break
			}
		}
		if foundInAll {
			commonInAll = append(commonInAll, weight)
		}
	}
	
	if len(commonInAll) > 0 {
		fmt.Printf("Berat yang sama: %s\n", strings.Join(commonInAll, ", "))
	} else {
		fmt.Println("Tidak ada berat yang sama di semua brand")
	}

	// Cari berat yang ada di 3 brand
	fmt.Println("\n\n🔍 BERAT YANG ADA DI MINIMAL 3 BRAND:")
	fmt.Println(strings.Repeat("=", 60))
	
	weightCount := make(map[string][]string)
	for _, brand := range brands {
		for weight := range brandWeights[brand.Brand] {
			weightCount[weight] = append(weightCount[weight], brand.Brand)
		}
	}
	
	commonIn3 := []string{}
	for weight, brandList := range weightCount {
		if len(brandList) >= 3 {
			commonIn3 = append(commonIn3, weight)
			fmt.Printf("Berat %s gram → Ada di %d brand: %s\n", 
				weight, len(brandList), strings.Join(brandList, ", "))
		}
	}
	
	if len(commonIn3) == 0 {
		fmt.Println("Tidak ada berat yang ada di minimal 3 brand")
	}

	// Cari berat yang unik per brand
	fmt.Println("\n\n✨ BERAT UNIK PER BRAND:")
	fmt.Println(strings.Repeat("=", 60))
	
	for _, brand := range brands {
		uniqueWeights := []string{}
		for weight := range brandWeights[brand.Brand] {
			if len(weightCount[weight]) == 1 {
				uniqueWeights = append(uniqueWeights, weight)
			}
		}
		if len(uniqueWeights) > 0 {
			fmt.Printf("\n%s:\n", brand.Brand)
			fmt.Printf("  Berat unik: %s\n", strings.Join(uniqueWeights, ", "))
		}
	}
}
