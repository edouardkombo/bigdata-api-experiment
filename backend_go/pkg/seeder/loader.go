package seeder

import (
    "context"

    clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)

// LoadBatch writes a batch of Events to ClickHouse.
func LoadBatch(ctx context.Context, conn clickhouse.Conn, events []Event) error {
    batch, err := conn.PrepareBatch(ctx,
        "INSERT INTO page_events (id, user_id, event_type, url, referrer, ts, meta)")
    if err != nil {
        return err
    }
    for _, e := range events {
        if err := batch.Append(
            e.ID,
            e.UserID,
            e.EventType,
            e.URL,
            e.Referrer,
            e.Timestamp,
            e.Meta,
        ); err != nil {
            return err
        }
    }
    return batch.Send()
}

