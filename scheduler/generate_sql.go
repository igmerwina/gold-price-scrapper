package main

import (
"encoding/json"
"fmt"
"io/ioutil"
"log"
"strconv"
"time"
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
	jsonFile, err := ioutil.ReadFile("harga_emas.json")
	if err != nil {
		log.Fatalf("Gagal membaca file JSON: %v", err)
	}

	var allBrandsData []BrandData
	err = json.Unmarshal(jsonFile, &allBrandsData)
	if err != nil {
		log.Fatalf("Gagal parse JSON: %v", err)
	}

	today := time.Now().Format("2006-01-02")

	fmt.Println("-- SQL UPDATE Queries untuk Gold Prices")
	fmt.Printf("-- Generated on: %s\n\n", today)

	for _, brandData := range allBrandsData {
		brand := brandData.Brand
		
		brandSQL := brand
		switch brand {
		case "GALERI 24":
			brandSQL = "Galeri24"
		case "BABY GALERI 24":
			brandSQL = "BabyGaleri24"
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

			fmt.Printf("UPDATE public.gold_prices_v2\n")
			fmt.Printf("SET price_buyback=%.1f, price_sell=%.0f\n", priceBuyback, priceSell)
			fmt.Printf("WHERE \"date\"='%s' AND brand='%s' AND denom=%.1f;\n\n", today, brandSQL, denom)
		}
	}

	fmt.Println("-- Total queries generated successfully")
}
