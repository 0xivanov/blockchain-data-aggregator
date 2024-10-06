package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0xivanov/blockchain-data-aggregator/models"
)

var geckoDateFormat = "02-01-2006"

// Structs to unmarshal the response from the CoinGecko API into
type CoinGeckoResponse struct {
	MarketData MarketData `json:"market_data"`
}
type MarketData struct {
	CurrentPrice map[string]float64 `json:"current_price"`
}

// CoinGeckoClient handles the communication with the CoinGecko API
type CoinGeckoClient struct {
	apiKey  string
	baseUrl string
	// this is the path to the file containing the official list of token IDs for the CoinGecko API
	tokenApiListPath string
	// injected function for testing purposes
	getTokenIdsFunc func(filePathtokenApiListPath string) (map[string]string, error)
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
func (geckoClient *CoinGeckoClient) GetPriceMap(ctx context.Context, transactions []models.Transaction) (map[string]float64, error) {
	// get the token IDs for the given currency symbols
	symbolToIdMap, err := geckoClient.getTokenIdsFunc(geckoClient.tokenApiListPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get token IDs: %v", err)
	}

	// prices holds symbol -> price mappings
	prices := make(map[string]float64)

	// chan to collect the results from the goroutines fetching the prices
	resultsChan := make(chan struct {
		symbol string
		price  float64
		err    error
	}, len(transactions))

	var wg sync.WaitGroup

	for _, txn := range transactions {
		// skip if price is already fetched
		if prices[txn.CurrencySymbol] != 0 {
			continue
		}

		// get the token ID and lowercase the symbol
		symbol := strings.ToLower(txn.CurrencySymbol)
		tokenId := symbolToIdMap[symbol]

		wg.Add(1)
		// spawn a goroutine to concurrently call the CoinGecko API
		go func(symbol string, tokenId string, txnDate time.Time) {
			defer wg.Done()

			// fetch price from CoinGecko in the background
			price, err := geckoClient.getPriceInUsd(ctx, tokenId, txnDate)

			// send the result back through the channel
			resultsChan <- struct {
				symbol string
				price  float64
				err    error
			}{
				symbol: symbol,
				price:  price,
				err:    err,
			}
		}(symbol, tokenId, txn.Date)
	}

	// wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// collect results from the channel
	for result := range resultsChan {
		if result.err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %v", result.symbol, result.err)
		}
		prices[result.symbol] = result.price
	}

	return prices, nil
}

func (geckoClient *CoinGeckoClient) getPriceInUsd(ctx context.Context, symbol string, date time.Time) (float64, error) {
	parsedDate := date.Format(geckoDateFormat)

	url := fmt.Sprintf("%s/coins/%s/history?date=%s?localization=false", geckoClient.baseUrl, symbol, parsedDate)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) // Create a new request with context
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
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
