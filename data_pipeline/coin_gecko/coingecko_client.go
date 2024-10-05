package coingecko

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xivanov/blockchain-data-aggregator/models"
)

// Structs to hold the response from the CoinGecko API
type CoinGeckoResponse struct {
	MarketData MarketData `json:"market_data"`
}
type MarketData struct {
	CurrentPrice map[string]float64 `json:"current_price"`
}

// CoinGeckoClient handles the communication with the CoinGecko API
type CoinGeckoClient struct {
	apiKey           string
	baseUrl          string
	tokenApiListPath string
	getTokenIdsFunc  func(filePathtokenApiListPath string) (map[string]string, error) // Injected function for testing
}

func NewCoinGeckoClient(apiKey, tokenApiListPath string) *CoinGeckoClient {
	return &CoinGeckoClient{
		apiKey:           apiKey,
		tokenApiListPath: tokenApiListPath,
		// use https://pro-api.coingecko.com/api/v3 for pro api keys
		baseUrl:         "https://api.coingecko.com/api/v3",
		getTokenIdsFunc: getCoinGeckoTokenIds,
	}
}

// GetPriceMap returns a map of currency symbols to their respective prices in USD at the given date
func (geckoClient *CoinGeckoClient) GetPriceMap(transactions []models.Transaction) (map[string]float64, error) {
	// Get the token IDs for the given currency symbols
	symbolToIdMap, err := geckoClient.getTokenIdsFunc(geckoClient.tokenApiListPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get token IDs: %v", err)
	}

	// prices holds symbol -> price mappings
	prices := make(map[string]float64)
	for _, txn := range transactions {

		// Skip if the price is already fetched
		if prices[txn.CurrencySymbol] != 0 {
			continue
		}

		// fetch the historical prices via the CoinGecko API
		price, err := geckoClient.getPriceInUsd(symbolToIdMap[strings.ToLower(txn.CurrencySymbol)], txn.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %v", txn.CurrencySymbol, err)
		}
		prices[txn.CurrencySymbol] = price
	}

	return prices, nil
}

var geckoDateFormat = "02-01-2006"

func (geckoClient *CoinGeckoClient) getPriceInUsd(symbol string, date time.Time) (float64, error) {
	parsedDate := date.Format(geckoDateFormat)

	url := fmt.Sprintf("%s/coins/%s/history?date=%s?localization=false", geckoClient.baseUrl, symbol, parsedDate)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("request failed with status: %v", resp.Status)
	}
	defer resp.Body.Close()

	var result CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	price := result.MarketData.CurrentPrice["usd"]
	if price == 0 {
		return 0, fmt.Errorf("price not found in response")
	}
	return price, nil
}
