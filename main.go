package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/0xivanov/blockchain-data-aggregator/config"
	"github.com/0xivanov/blockchain-data-aggregator/data_pipeline/aggregate"
	coingecko "github.com/0xivanov/blockchain-data-aggregator/data_pipeline/coin_gecko"
	"github.com/0xivanov/blockchain-data-aggregator/data_pipeline/db"
	"github.com/0xivanov/blockchain-data-aggregator/data_pipeline/extraction"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func main() {
	// Load the configuration from config.json
	config, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize the ClickHouse database
	db, err := db.NewClickHouseDB(config.ClickhouseDSN, config.DbName)
	if err != nil {
		log.Fatalf("Failed to initialize ClickHouse: %v", err)
	}

	// Initialize the CoinGecko client
	geckoClient := coingecko.NewCoinGeckoClient(config.CoinGeckoAPI, "coingecko_token_api_list.csv")

	// Load the service account credentials from the JSON key file
	credentials, err := os.ReadFile(config.BucketKeyPath)
	if err != nil {
		log.Fatalf("Failed to read service account key file: %v", err)
	}

	ctx := context.Background()

	// Initialize the GCP extractor
	gcpConf, err := google.CredentialsFromJSON(ctx, credentials, storage.ScopeReadOnly)
	if err != nil {
		log.Fatalf("Failed to create credentials from JSON: %v", err)
	}
	client, err := storage.NewClient(ctx, option.WithCredentials(gcpConf))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	gcpExtractor := extraction.NewGCPExtractor(client)

	// Download the transactions from GCS
	transactions, err := gcpExtractor.ExtractTransactionsFromGCS(config.BucketName, config.ObjectName, ctx)
	if err != nil {
		log.Fatalf("Failed to fetch data from GCS: %v", err)
	}
	log.Println("Data successfully extracted from GCS")

	// Get the prices from CoinGecko
	priceMap, err := geckoClient.GetPriceMap(transactions)
	if err != nil {
		log.Fatalf("Failed to get price map: %v", err)
	}
	log.Println("Prices successfully fetched from CoinGecko")

	// Aggregate the transactions
	marketplaceData, err := aggregate.AggregateTransactions(transactions, priceMap)
	if err != nil {
		log.Fatalf("Failed to aggregate transactions: %v", err)
	}

	// Save the aggregated data into ClickHouse
	if err := db.SaveMarketplaceData(marketplaceData); err != nil {
		log.Fatalf("Failed to save data into ClickHouse: %v", err)
	}
	log.Println("Data successfully inserted into ClickHouse")
}
