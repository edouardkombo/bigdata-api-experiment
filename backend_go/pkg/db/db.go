// backend_go/pkg/db/db.go
package db

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"

    clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)

type Overview struct {
    TotalEvents    uint64 `json:"total_events"`
    UniqueUsers    uint64 `json:"unique_users"`
    EventsLastHour uint64 `json:"events_last_hour"`
}

type Event struct {
    ID        string                 `json:"id"`
    UserID    string                 `json:"user_id"`
    EventType string                 `json:"event_type"`
    URL       string                 `json:"url"`
    Referrer  string                 `json:"referrer"`
    Timestamp time.Time              `json:"ts"`
    Meta      map[string]interface{} `json:"meta"`
}

type TimeSeriesPoint struct {
    Bucket time.Time `json:"bucket"`
    Count  uint64    `json:"count"`
}

// openConn creates a ClickHouse connection using env vars
func openConn(ctx context.Context) (clickhouse.Conn, error) {
    host := os.Getenv("CLICKHOUSE_HOST")
    if host == "" {
        host = "127.0.0.1"
    }
    port := os.Getenv("CLICKHOUSE_TCP_PORT")
    if port == "" {
        port = "9000"
    }
    addr := fmt.Sprintf("%s:%s", host, port)

    return clickhouse.Open(&clickhouse.Options{
        Addr: []string{addr},
        Auth: clickhouse.Auth{
            Database: "analytics",
            Username: "default",
            Password: "",
        },
        DialTimeout: 5 * time.Second,
    })
}

// GetOverview runs aggregate queries to build the dashboard summary
func GetOverview() (*Overview, error) {
    ctx := context.Background()
    conn, err := openConn(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    const q = `
        SELECT
          (SELECT count() FROM page_events) AS total,
          (SELECT count(DISTINCT user_id) FROM page_events) AS unique_users,
          (SELECT count() FROM page_events WHERE ts >= now() - INTERVAL 1 HOUR) AS last_hour
    `
    row := conn.QueryRow(ctx, q)
    var ov Overview
    if err := row.Scan(&ov.TotalEvents, &ov.UniqueUsers, &ov.EventsLastHour); err != nil {
        return nil, err
    }
    return &ov, nil
}

// StreamEvents returns up to `limit` events after the optional cursor timestamp
func StreamEvents(cursor string, limit int) ([]Event, error) {
    ctx := context.Background()
    conn, err := openConn(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    var since time.Time
    if t, err := time.Parse(time.RFC3339, cursor); err == nil {
        since = t
    }

    const q = `
        SELECT id, user_id, event_type, url, referrer, ts, meta
        FROM page_events
        WHERE ts > ?
        ORDER BY ts
        LIMIT ?
    `
    rows, err := conn.Query(ctx, q, since, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []Event
    for rows.Next() {
        var e Event
        var rawMeta string
        if err := rows.Scan(
            &e.ID,
            &e.UserID,
            &e.EventType,
            &e.URL,
            &e.Referrer,
            &e.Timestamp,
            &rawMeta,
        ); err != nil {
            return nil, err
        }
        if err := json.Unmarshal([]byte(rawMeta), &e.Meta); err != nil {
            e.Meta = map[string]interface{}{}
        }
        events = append(events, e)
    }
    return events, nil
}

// GetTimeSeries returns counts per interval between two timestamps.
func GetTimeSeries(fromISO, toISO, interval string) ([]TimeSeriesPoint, error) {
    ctx := context.Background()
    conn, err := openConn(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    // Default to last 24h if none provided
    now := time.Now().UTC()
    if toISO == "" {
        toISO = now.Format(time.RFC3339)
    }
    if fromISO == "" {
        fromISO = now.Add(-24 * time.Hour).Format(time.RFC3339)
    }
    if interval == "" {
        interval = "1 minute"
    }

    const q = `
        SELECT
            toStartOfInterval(ts, INTERVAL ? ) AS bucket,
            count() AS cnt
        FROM page_events
        WHERE ts BETWEEN parseDateTimeBestEffort(?) AND parseDateTimeBestEffort(?)
        GROUP BY bucket
        ORDER BY bucket
    `
    rows, err := conn.Query(ctx, q, interval, fromISO, toISO)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var series []TimeSeriesPoint
    for rows.Next() {
        var p TimeSeriesPoint
        if err := rows.Scan(&p.Bucket, &p.Count); err != nil {
            return nil, err
        }
        series = append(series, p)
    }
    return series, nil
}

// GetTypeBreakdown returns a map of event_type to count.
func GetTypeBreakdown() (map[string]uint64, error) {
    ctx := context.Background()
    conn, err := openConn(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    const q = `
        SELECT event_type, count() AS cnt
        FROM page_events
        GROUP BY event_type
    `
    rows, err := conn.Query(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    breakdown := make(map[string]uint64)
    for rows.Next() {
        var t string
        var cnt uint64
        if err := rows.Scan(&t, &cnt); err != nil {
            return nil, err
        }
        breakdown[t] = cnt
    }
    return breakdown, nil
}

