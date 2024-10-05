package extraction

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/0xivanov/blockchain-data-aggregator/models"
)

var (
	currencySymbolRegex       = regexp.MustCompile(`"currencySymbol":"([^"]+)"`)
	currencyValueDecimalRegex = regexp.MustCompile(`"currencyValueDecimal":"([^"]+)"`)
)

type GCPExtractor struct {
	client *storage.Client
}

// NewGCPExtractor creates a new GCPExtractor.
func NewGCPExtractor(client *storage.Client) *GCPExtractor {
	return &GCPExtractor{client}
}

func (gcpExtractor *GCPExtractor) ExtractTransactionsFromGCS(bucketName, objectName string, ctx context.Context) ([]models.Transaction, error) {
	reader, err := gcpExtractor.client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ','

	return extractTransactions(csvReader)
}

func extractTransactions(csvReader *csv.Reader) ([]models.Transaction, error) {
	var transactions []models.Transaction

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	var dateIndex, projectIDIndex, propsIndex, numsIndex int
	for i, header := range headers {
		if header == "ts" {
			dateIndex = i
		} else if header == "project_id" {
			projectIDIndex = i
		} else if header == "props" {
			propsIndex = i
		} else if header == "nums" {
			numsIndex = i
		} else {
			return nil, fmt.Errorf("unknown header: %s", header)
		}
	}

	for {
		record, err := csvReader.Read()
		// read until EOF
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error occured during reading csv file: %v", err)
		}

		dateStr := record[dateIndex]        // "ts"
		projectID := record[projectIDIndex] // "project_id"
		props := record[propsIndex]         // "props"
		nums := record[numsIndex]           // "nums"

		// Extract currencySymbol and currencyValueDecimal from props and nums fields
		currencySymbol, err := extractCurrencySymbol(props)
		if err != nil {
			return nil, fmt.Errorf("failed to get currency symbol: %v", err)
		}
		currencyValueDecimal, err := extractCurrencyValueDecimal(nums)
		if err != nil {
			return nil, fmt.Errorf("failed to get currency value symbol: %v", err)
		}

		// Parse the timestamp into a time.Time object
		parsedTime, err := time.Parse(time.DateTime, dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %v", err)
		}

		transactions = append(transactions, models.Transaction{
			Date:                 parsedTime,
			ProjectID:            projectID,
			CurrencySymbol:       currencySymbol,
			CurrencyValueDecimal: currencyValueDecimal,
		})
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in CSV")
	}
	return transactions, nil
}

func extractCurrencySymbol(propsString string) (string, error) {
	matches := currencySymbolRegex.FindStringSubmatch(propsString)

	// Check if a match was found
	if len(matches) > 1 {
		return matches[1], nil
	} else {
		return "", fmt.Errorf("currencySymbol not found in props")
	}
}

func extractCurrencyValueDecimal(numsString string) (float64, error) {
	matches := currencyValueDecimalRegex.FindStringSubmatch(numsString)

	// Check if a match was found
	if len(matches) > 1 {
		// Convert the value to a float64
		currencyValueDecimal, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse currency value: %v", err)
		}
		return currencyValueDecimal, nil
	} else {
		return 0, fmt.Errorf("currencyValueDecimal not found in nums")
	}
}
