package aggregate

import (
	"testing"
	"time"

	"github.com/0xivanov/blockchain-data-aggregator/models"
	"github.com/stretchr/testify/assert"
)

// Constants for the prices of BTC and ETH
const (
	BTCPrice = 60000.0
	ETHPrice = 1500.0
)

func TestAggregateTransactions_Basic(t *testing.T) {
	transactions := []models.Transaction{
		{
			Date:                 time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_1",
			CurrencySymbol:       "ETH",
			CurrencyValueDecimal: 2.0,
		},
		{
			Date:                 time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_1",
			CurrencySymbol:       "ETH",
			CurrencyValueDecimal: 3.0,
		},
		{
			Date:                 time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_1",
			CurrencySymbol:       "BTC",
			CurrencyValueDecimal: 1.0,
		},
		{
			Date:                 time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_2",
			CurrencySymbol:       "BTC",
			CurrencyValueDecimal: 0.5,
		},
	}

	priceMap := map[string]float64{
		"ETH": ETHPrice,
		"BTC": BTCPrice,
	}

	expected := []models.MarketplaceData{
		{
			Date:            "2024-04-01",
			ProjectID:       "project_1",
			NumTransactions: 2,
			TotalVolumeUSD:  5 * ETHPrice, // 2 ETH + 3 ETH = 5 ETH * 1500 = 7500
		},
		{
			Date:            "2024-04-02",
			ProjectID:       "project_1",
			NumTransactions: 1,
			TotalVolumeUSD:  BTCPrice, // 1 BTC * 30000 = 30000
		},
		{
			Date:            "2024-04-02",
			ProjectID:       "project_2",
			NumTransactions: 1,
			TotalVolumeUSD:  0.5 * BTCPrice, // 0.5 BTC * 30000 = 15000
		},
	}

	result, err := AggregateTransactions(transactions, priceMap)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, result)
}

func TestAggregateTransactions_Empty(t *testing.T) {
	transactions := []models.Transaction{}
	priceMap := map[string]float64{}

	_, err := AggregateTransactions(transactions, priceMap)
	assert.ErrorContains(t, err, "no transactions to aggregate")
}

func TestAggregateTransactions_NoPriceInMap(t *testing.T) {
	transactions := []models.Transaction{
		{
			Date:                 time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_1",
			CurrencySymbol:       "ETH",
			CurrencyValueDecimal: 2.0,
		},
	}

	priceMap := map[string]float64{
		// Empty price map, no price for "ETH"
	}

	_, err := AggregateTransactions(transactions, priceMap)
	assert.ErrorContains(t, err, "no price found for")
}

func TestAggregateTransactions_MultipleProjects(t *testing.T) {
	transactions := []models.Transaction{
		{
			Date:                 time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_1",
			CurrencySymbol:       "ETH",
			CurrencyValueDecimal: 2.0,
		},
		{
			Date:                 time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_2",
			CurrencySymbol:       "BTC",
			CurrencyValueDecimal: 1.0,
		},
	}

	priceMap := map[string]float64{
		"ETH": ETHPrice,
		"BTC": BTCPrice,
	}

	expected := []models.MarketplaceData{
		{
			Date:            "2024-04-01",
			ProjectID:       "project_1",
			NumTransactions: 1,
			TotalVolumeUSD:  2 * ETHPrice, // 2 ETH * 1500 = 3000
		},
		{
			Date:            "2024-04-01",
			ProjectID:       "project_2",
			NumTransactions: 1,
			TotalVolumeUSD:  1 * BTCPrice, // 1 BTC * 30000 = 30000
		},
	}

	result, err := AggregateTransactions(transactions, priceMap)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, result)
}
