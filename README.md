# Blockchain Data Aggregator

## Features

- **Transaction Extraction from GCS**: Extracts and parses CSV transaction data from Google Cloud Storage.
- **Currency Price Fetching**: Integrates with the CoinGecko API to fetch historical prices for cryptocurrencies.
- **Transaction Aggregation**: Aggregates transaction data by day and project, computes total transaction volume, and converts it into USD.
- **Data loading to Clickhouse**: Loads the aggregated data into clickhouse db schema
- **Error Handling**: Implements comprehensive error handling during data extraction, transformation, and API calls.

---

## Requirements

- **Go**: v1.18+
- **Google Cloud SDK**: For accessing Google Cloud Storage
- **CoinGecko API Key**: For fetching historical cryptocurrency prices
- **Docker**: For running the Clickhouse db server
- **Task**: For utility commands

---

## Installation

1. **Clone the Repository**:

    ```bash
    git clone https://github.com/0xivanov/blockchain-data-aggregator.git
    cd blockchain-data-aggregator
    ```

2. **Install Dependencies**:

    This project uses Go modules. Ensure Go is installed, then run:

    ```bash
    go mod tidy
    ```

3. **Set Up Config Variables**:

    - **Google Cloud Storage**:
        - Create new GCP bucket and link Service Account with a key to it. More info [Here](https://medium.com/@manjunath.kmph/access-to-specific-gcs-bucket-using-service-account-and-key-f1f7c16445ae)
        - Upload the data and download the key associated with the Service Account
    - **CoinGecko API**:
        - Create CoinGecko Developer account and create new API key

    - **Create config.json**:
        - Create config.json file in the root of the project and fill your details from the above steps
        - See config.example.json for reference

4. **Start Clickhouse Server**:

    Run the following taskfile command

    ```bash
    task init-db
    ```
---

## Usage

### 1. Run the Aggregator

To run the main aggregator script:

```bash
go run main.go
```

Make sure to provide the necessary fields in `config.json` file.

### 2. Viewing the aggregated data

You can use 3rd party UI tool to view the aggregated data in Clickhouse.
See the available options [Here](https://clickhouse.com/docs/en/interfaces/third-party/gui)

### 3. Testing

The project uses `go test` for unit tests. Each component has its own tests.

Run all tests with:

```bash
go test ./...
```