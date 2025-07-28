package main

import (
	"database/sql"
	"log"
	"time"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	ingestpb "bigdata-perf/proto"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("❌ %s: %s", msg, err)
	}
}

func main() {
	log.Println("🔌 Connecting to ClickHouse...")
	db, err := sql.Open("clickhouse", "tcp://127.0.0.1:9000?debug=false")
	failOnError(err, "ClickHouse open error")
	defer db.Close()

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp091.Dial(rabbitmqURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open RabbitMQ channel")
	defer ch.Close()

	msgs, err := ch.Consume(
		"events",
		"go-consumer",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register consumer")

	log.Println("🟢 RabbitMQ consumer listening on queue 'events'")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for d := range msgs {
			var req ingestpb.EventRequest

			log.Printf("🐇 Received message: %d bytes", len(d.Body))
                        err := proto.Unmarshal(d.Body, &req)
                        if err != nil {
	                    log.Printf("❌ Protobuf decode error: %v", err)
	                    continue
                        }
                        log.Printf("✅ Parsed event: ID=%s | User=%s | Type=%s", req.Id, req.UserId, req.EventType)

                        tx, err := db.Begin()
                        if err != nil {
	                    log.Fatalf("❌ Failed to begin transaction: %v", err)
                        }			
			stmt, err := tx.Prepare("INSERT INTO analytics.page_events (id, user_id, event_type, url, referrer, ts, meta) VALUES (?, ?, ?, ?, ?, ?, ?)")
                        if err != nil {
	                    log.Fatalf("❌ Failed to prepare statement: %v", err)
                        }
                        defer stmt.Close()

                        tsTime, err := time.Parse(time.RFC3339, req.Ts)
                        if err != nil {
	                    log.Printf("❌ Failed to parse timestamp: %v", err)
	                    tsTime = time.Now() // fallback to now
                        }

			metaJSON, _ := json.Marshal(req.Meta)
                        _, err = stmt.Exec(
	                    req.Id,
	                    req.UserId,
	                    req.EventType,
	                    req.Url,
	                    req.Referrer,
	                    tsTime,
	                    string(metaJSON),
                        )
                        if err != nil {
	                    log.Printf("❌ Insert failed: %v", err)
                        } else {
	                    log.Printf("✅ Inserted event: %s", req.Id)
                        }

			if err := tx.Commit(); err != nil {
	                    log.Fatalf("❌ Commit failed: %v", err)
                        }
		}
	}()

	<-sig
	log.Println("👋 Graceful shutdown.")
}

