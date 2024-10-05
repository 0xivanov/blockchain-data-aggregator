CREATE TABLE marketplace_data (
    date Date,
    project_id String,
    num_transactions Int32,
    total_volume_usd Float32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, project_id);
