package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config holds the configuration for the application
type Config struct {
	ClickhouseDSN string `json:"clickhouseDSN"`
	DbName        string `json:"dbName"`
	BucketKeyPath string `json:"bucketKeyPath"`
	BucketName    string `json:"bucketName"`
	ObjectName    string `json:"objectName"`
	CoinGeckoAPI  string `json:"coinGeckoAPI"`
}

// LoadConfig reads the config.json file and unmarshals it into a Config struct
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
