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
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type Transaction struct {
	Date                 time.Time
	ProjectID            string
	CurrencySymbol       string
	CurrencyValueDecimal float64
}

func ExtractTransactionsFromGCS(credentials []byte, bucketName, objectName string) ([]Transaction, error) {
	ctx := context.Background()

	config, err := google.CredentialsFromJSON(ctx, credentials, storage.ScopeReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials from JSON: %v", err)
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(config))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	reader, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ','

	return extractTransactions(csvReader)
}

func extractTransactions(csvReader *csv.Reader) ([]Transaction, error) {
	var transactions []Transaction

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

		transactions = append(transactions, Transaction{
			Date:                 parsedTime,
			ProjectID:            projectID,
			CurrencySymbol:       currencySymbol,
			CurrencyValueDecimal: currencyValueDecimal,
		})
	}

	return transactions, nil
}

var currencySymbolRegex = regexp.MustCompile(`"currencySymbol":"([^"]+)"`)

func extractCurrencySymbol(propsString string) (string, error) {
	matches := currencySymbolRegex.FindStringSubmatch(propsString)

	// Check if a match was found
	if len(matches) > 1 {
		return matches[1], nil
	} else {
		return "", fmt.Errorf("currencySymbol not found in props")
	}
}

var currencyValueDecimalRegex = regexp.MustCompile(`"currencyValueDecimal":"([^"]+)"`)

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
