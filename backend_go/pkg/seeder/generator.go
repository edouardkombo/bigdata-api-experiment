package seeder

import (
    "time"

    "github.com/brianvoe/gofakeit/v6"
)

type Event struct {
    ID        string
    UserID    string
    EventType string
    URL       string
    Referrer  string
    Timestamp time.Time
    Meta      map[string]interface{}
}

// GenerateEvents yields batches of fake events indefinitely.
func GenerateEvents(batchSize int) <-chan []Event {
    ch := make(chan []Event)
    go func() {
        defer close(ch)
        for {
            batch := make([]Event, batchSize)
            for i := range batch {
                batch[i] = Event{
                    ID:        gofakeit.UUID(),
                    UserID:    gofakeit.UUID(),
                    EventType: gofakeit.RandomString([]string{"page_view", "click", "scroll"}),
                    URL:       gofakeit.URL(),
                    Referrer:  gofakeit.URL(),
                    Timestamp: gofakeit.Date(),
                    Meta: map[string]interface{}{
                        "x": gofakeit.Number(0, 1000),
                        "y": gofakeit.Number(0, 1000),
                    },
                }
            }
            ch <- batch
        }
    }()
    return ch
}

