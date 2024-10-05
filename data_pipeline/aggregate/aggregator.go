package aggregate

import (
	"fmt"

	"github.com/0xivanov/blockchain-data-aggregator/models"
)

// AggregateTransactions aggregates the given transactions by day and project ID
func AggregateTransactions(transactions []models.Transaction, priceMap map[string]float64) ([]models.MarketplaceData, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions to aggregate")
	}
	// hash map to group transactions by day and project ID
	aggregated := make(map[string]models.MarketplaceData)

	for _, txn := range transactions {
		day := txn.Date.Format("2006-01-02")
		key := day + "-" + txn.ProjectID

		price := priceMap[txn.CurrencySymbol]
		if price == 0 {
			return nil, fmt.Errorf("no price found for %s", txn.CurrencySymbol)
		}

		transactionVolume := price * txn.CurrencyValueDecimal

		agg := aggregated[key]
		agg.Date = day
		agg.ProjectID = txn.ProjectID
		agg.NumTransactions++
		agg.TotalVolumeUSD += transactionVolume

		aggregated[key] = agg
	}

	// convert map to slice
	var result []models.MarketplaceData
	for _, v := range aggregated {
		result = append(result, v)
	}

	return result, nil
}
