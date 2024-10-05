package coingecko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xivanov/blockchain-data-aggregator/models"
	"github.com/stretchr/testify/assert"
)

// Mock for getCoinGeckoTokenIds
func mockGetCoinGeckoTokenIds(tokenApiListPath string) (map[string]string, error) {
	// returns a mock mapping of currency symbols to their CoinGecko token IDs
	return map[string]string{
		"eth": "ethereum",
		"btc": "bitcoin",
	}, nil
}

// mock HTTP server for testing
func setupMockServer(response string, statusCode int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	})
	return httptest.NewServer(handler)
}

func TestCoinGeckoClient_GetPriceMap(t *testing.T) {
	mockResponse := `{
		"market_data": {
			"current_price": {
				"usd": 2000.5
			}
		}
	}`
	mockServer := setupMockServer(mockResponse, http.StatusOK)
	defer mockServer.Close()

	geckoClient := &CoinGeckoClient{
		baseUrl:          mockServer.URL,
		tokenApiListPath: "mock/path",
		getTokenIdsFunc:  mockGetCoinGeckoTokenIds,
	}

	transactions := []models.Transaction{
		{
			CurrencySymbol: "ETH",
			Date:           time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			CurrencySymbol: "BTC",
			Date:           time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	priceMap, err := geckoClient.GetPriceMap(context.TODO(), transactions)
	assert.NoError(t, err)

	assert.Equal(t, 2000.5, priceMap["ETH"])
	assert.Equal(t, 2000.5, priceMap["BTC"])
}

func TestCoinGeckoClient_GetPriceMap_ReusePrice(t *testing.T) {
	mockResponse := `{
		"market_data": {
			"current_price": {
				"usd": 2000.5
			}
		}
	}`
	mockServer := setupMockServer(mockResponse, http.StatusOK)
	defer mockServer.Close()

	geckoClient := &CoinGeckoClient{
		baseUrl:          mockServer.URL,
		tokenApiListPath: "mock/path",
		getTokenIdsFunc:  mockGetCoinGeckoTokenIds,
	}

	transactions := []models.Transaction{
		{
			CurrencySymbol: "ETH",
			Date:           time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			CurrencySymbol: "ETH",
			Date:           time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	priceMap, err := geckoClient.GetPriceMap(context.TODO(), transactions)
	assert.NoError(t, err)
	assert.Equal(t, 2000.5, priceMap["ETH"])
}

func TestCoinGeckoClient_GetPriceInUsd(t *testing.T) {
	mockResponse := `{
		"market_data": {
			"current_price": {
				"usd": 1500.75
			}
		}
	}`
	mockServer := setupMockServer(mockResponse, http.StatusOK)
	defer mockServer.Close()

	geckoClient := &CoinGeckoClient{
		baseUrl: mockServer.URL,
	}

	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	price, err := geckoClient.getPriceInUsd(context.TODO(), "ethereum", date)
	assert.NoError(t, err)
	assert.Equal(t, 1500.75, price)
}

func TestCoinGeckoClient_GetPriceInUsd_MalformedResponse(t *testing.T) {
	mockResponse := `{
		"market_data": {
			"current_price": {
				"usd": "invalid_price"
			}
		}
	}`
	mockServer := setupMockServer(mockResponse, http.StatusOK)
	defer mockServer.Close()

	geckoClient := &CoinGeckoClient{
		baseUrl: mockServer.URL,
	}

	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := geckoClient.getPriceInUsd(context.TODO(), "ethereum", date)
	assert.Error(t, err)
}

func TestCoinGeckoClient_GetPriceInUsd_EmptyResp(t *testing.T) {
	mockServer := setupMockServer("{}", http.StatusOK)
	defer mockServer.Close()

	geckoClient := &CoinGeckoClient{
		baseUrl: mockServer.URL,
	}

	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := geckoClient.getPriceInUsd(context.TODO(), "ethereum", date)
	assert.Error(t, err)
}

func TestCoinGeckoClient_GetPriceInUsd_ApiError(t *testing.T) {
	mockServer := setupMockServer("{}", http.StatusInternalServerError)
	defer mockServer.Close()

	geckoClient := &CoinGeckoClient{
		baseUrl: mockServer.URL,
	}

	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := geckoClient.getPriceInUsd(context.TODO(), "ethereum", date)
	assert.ErrorContains(t, err, "request failed with status")
}
