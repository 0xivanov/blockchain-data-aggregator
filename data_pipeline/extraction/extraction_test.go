package extraction

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/0xivanov/blockchain-data-aggregator/models"
	"github.com/stretchr/testify/assert"
)

var sampleCSVData = `
ts,project_id,props,nums
2024-04-01 00:00:00,project_1,"{""currencySymbol"":""SFL"",""currencyValueDecimal"":""100.50""}","{""currencyValueDecimal"":""100.50""}"
2024-04-02 00:00:00,project_2,"{""currencySymbol"":""MATIC"",""currencyValueDecimal"":""200.75""}","{""currencyValueDecimal"":""200.75""}"
`

func TestExtractTransactions_ValidCSV(t *testing.T) {
	csvReader := csv.NewReader(strings.NewReader(sampleCSVData))

	expected := []models.Transaction{
		{
			Date:                 time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_1",
			CurrencySymbol:       "SFL",
			CurrencyValueDecimal: 100.50,
		},
		{
			Date:                 time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
			ProjectID:            "project_2",
			CurrencySymbol:       "MATIC",
			CurrencyValueDecimal: 200.75,
		},
	}

	result, err := extractTransactions(csvReader)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractTransactions_InvalidCSV(t *testing.T) {
	csvContent := `invalid_header,unknown_field,props,nums
2024-04-01T00:00:00,project_1,"{\"currencySymbol\":\"ETH\"}","{\"currencyValueDecimal\":\"2.0\"}"
`
	csvReader := csv.NewReader(strings.NewReader(csvContent))

	_, err := extractTransactions(csvReader)
	assert.Error(t, err)
}

func TestExtractTransactions_MissingFields(t *testing.T) {
	csvContent := `ts,project_id,props,nums
2024-04-01T00:00:00,project_1,"{\"currencySymbol\":\"ETH\"}",""
`
	csvReader := csv.NewReader(strings.NewReader(csvContent))

	_, err := extractTransactions(csvReader)
	assert.Error(t, err)
}

func TestExtractCurrencySymbol_Valid(t *testing.T) {
	props := `{"currencySymbol":"BTC"}`
	expected := "BTC"

	result, err := extractCurrencySymbol(props)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractCurrencySymbol_Missing(t *testing.T) {
	props := `{"someOtherField":"ETH"}`

	_, err := extractCurrencySymbol(props)
	assert.Error(t, err)
}

func TestExtractCurrencyValueDecimal_Valid(t *testing.T) {
	nums := `{"currencyValueDecimal":"30000.5"}`
	expected := 30000.5

	result, err := extractCurrencyValueDecimal(nums)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractCurrencyValueDecimal_Invalid(t *testing.T) {
	nums := `{"currencyValueDecimal":"invalid_value"}`

	_, err := extractCurrencyValueDecimal(nums)
	assert.Error(t, err)
}

func TestExtractCurrencyValueDecimal_Missing(t *testing.T) {
	nums := `{"someOtherField":"100.0"}`

	_, err := extractCurrencyValueDecimal(nums)
	assert.Error(t, err)
}
