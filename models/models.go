package models

import "time"

// The aggregated data for a single day and project
type MarketplaceData struct {
	Date            string
	ProjectID       string
	NumTransactions uint64
	TotalVolumeUSD  float64
}

// A single transaction record
type Transaction struct {
	Date                 time.Time
	ProjectID            string
	CurrencySymbol       string
	CurrencyValueDecimal float64
}
