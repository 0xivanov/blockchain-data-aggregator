package aggregate

import "github.com/0xivanov/blockchain-data-aggregator/data_pipeline/extraction"

// The aggregated data for a single day and project
type MarketplaceData struct {
	Date            string
	ProjectID       string
	NumTransactions uint64
	TotalVolumeUSD  float64
}

// AggregateTransactions aggregates the given transactions by day and project ID
func AggregateTransactions(transactions []extraction.Transaction, priceMap map[string]float64) []MarketplaceData {
	// Hash map to group transactions by day and project ID
	aggregated := make(map[string]MarketplaceData)

	for _, txn := range transactions {
		day := txn.Date.Format("2006-01-02")
		key := day + "-" + txn.ProjectID

		transactionVolume := priceMap[txn.CurrencySymbol] * txn.CurrencyValueDecimal

		agg := aggregated[key]
		agg.Date = day
		agg.ProjectID = txn.ProjectID
		agg.NumTransactions++
		agg.TotalVolumeUSD += transactionVolume

		aggregated[key] = agg
	}

	// Convert map to slice
	var result []MarketplaceData
	for _, v := range aggregated {
		result = append(result, v)
	}

	return result
}
