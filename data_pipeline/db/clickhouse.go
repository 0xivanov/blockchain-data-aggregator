package db

import (
	"database/sql"
	"fmt"

	"github.com/0xivanov/blockchain-data-aggregator/models"
	"github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseDB handles the communication with the ClickHouse database
type ClickHouseDB struct {
	dsn  string
	conn *sql.DB
}

func NewClickHouseDB(dsn, dbName string) (*ClickHouseDB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: dbName,
			Username: "default",
			Password: "",
		},
		Protocol: clickhouse.HTTP,
	})

	// Ensure the connection is working - fail fast
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %v", err)
	}

	return &ClickHouseDB{dsn: dsn, conn: conn}, nil
}

// SaveMarketplaceData saves the given marketplace data to the ClickHouse database
func (clickHouse *ClickHouseDB) SaveMarketplaceData(data []models.MarketplaceData) error {

	// build the insert query
	var values string
	for _, d := range data {
		values += fmt.Sprintf(`('%s', '%s', %d, %f) `, d.Date, d.ProjectID, d.NumTransactions, d.TotalVolumeUSD)
	}

	query := "INSERT INTO marketplace_data (date, project_id, num_transactions, total_volume_usd) VALUES " + values
	_, err := clickHouse.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute insert statement: %v", err)
	}

	return nil
}
