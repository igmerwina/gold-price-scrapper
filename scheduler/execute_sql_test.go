package main

import (
	"strings"
	"testing"
)

const generatedSQL = `-- SQL Upsert Queries untuk Gold Prices
-- Generated on: 2026-09-19 08:10:00

INSERT INTO public.gold_prices_v2 ("date", brand, denom, price_buyback, price_sell)
VALUES ('2026-09-19', 'Galeri24', 0.5, 842500.0, 892500)
ON CONFLICT ("date", brand, denom) DO UPDATE
SET price_buyback=EXCLUDED.price_buyback, price_sell=EXCLUDED.price_sell;

INSERT INTO public.gold_prices_v2 ("date", brand, denom, price_buyback, price_sell)
VALUES ('2026-09-19', 'Antam', 1.0, 1700000.0, 1800000)
ON CONFLICT ("date", brand, denom) DO UPDATE
SET price_buyback=EXCLUDED.price_buyback, price_sell=EXCLUDED.price_sell;

-- Total 2 upsert queries generated successfully
`

func TestParseStatements(t *testing.T) {
	statements := parseStatements(generatedSQL)

	if len(statements) != 2 {
		t.Fatalf("dapat %d statement, mau 2", len(statements))
	}

	// Query pertama menempel pada komentar header; pernah terbuang karena itu.
	if !strings.Contains(statements[0], "Galeri24") {
		t.Errorf("statement pertama hilang, dapat: %q", statements[0])
	}

	for i, stmt := range statements {
		if strings.Contains(stmt, "--") {
			t.Errorf("statement %d masih mengandung komentar: %q", i, stmt)
		}
	}
}
