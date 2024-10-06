package coingecko

import (
	"encoding/csv"
	"os"
)

// Read the token API list from the CSV file
func getCoinGeckoTokenIds(filePathtokenApiListPath string) (map[string]string, error) {
	f, err := os.Open(filePathtokenApiListPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	// skip the header
	_, err = csvReader.Read()
	if err != nil {
		return nil, err
	}
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	symbolToIdMap := make(map[string]string)
	for _, record := range records {
		// record[0] is the token ID, record[1] is the currency symbol
		symbolToIdMap[record[1]] = record[0]
	}

	return symbolToIdMap, nil
}
