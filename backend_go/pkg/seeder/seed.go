package seeder

import (
    "context"
    "fmt"
    "os"
    "time"

    clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)

// Seed populates ClickHouse with totalRows events in batches of batchSize.
func Seed(totalRows, batchSize int) error {
    // Build ClickHouse address
    host := os.Getenv("CLICKHOUSE_HOST")
    if host == "" {
        host = "127.0.0.1"
    }
    port := os.Getenv("CLICKHOUSE_PORT")
    if port == "" {
        port = "9000"
    }
    addr := fmt.Sprintf("%s:%s", host, port)

    // Connect
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{addr},
        Auth: clickhouse.Auth{
            Database: "analytics",
            Username: "default",
            Password: "",
        },
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        return fmt.Errorf("clickhouse connect: %w", err)
    }
    defer conn.Close()

    ctx := context.Background()
    inserted := 0

    for batch := range GenerateEvents(batchSize) {
        if err := LoadBatch(ctx, conn, batch); err != nil {
            return fmt.Errorf("load batch: %w", err)
        }
        inserted += len(batch)
        if inserted >= totalRows {
            break
        }
    }
    return nil
}

