package main

import (
	"encoding/json"
	"flag"
	"strconv"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp091.Dial(rabbitmqURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	_, err = ch.QueueDeclare(
		"events",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare queue")

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		var req ingestpb.EventRequest

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("❌ Error reading body: %v", err)
			http.Error(w, "bad request", 400)
			return
		}
		log.Printf("📥 Raw body: %s", string(body))

		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("❌ Invalid JSON: %v", err)
			http.Error(w, "invalid json", 400)
			return
		}

		if req.Id == "" {
			req.Id = time.Now().Format("20060102150405")
		}
		req.Ts = time.Now().Format(time.RFC3339)
		log.Printf("✅ Parsed Request: %+v", req)

		data, err := proto.Marshal(&req)
		if err != nil {
			log.Printf("❌ Failed to marshal proto: %v", err)
			http.Error(w, "internal error", 500)
			return
		}

		err = ch.Publish(
			"",
			"events",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent,
				ContentType:  "application/octet-stream",
				Body:         data,
			},
		)
		if err != nil {
			log.Printf("❌ Failed to publish message: %v", err)
			http.Error(w, "failed to enqueue", 500)
			return
		}

		log.Printf("✅ Event enqueued")
		w.WriteHeader(http.StatusAccepted)
	})

	log.Printf("🚀 HTTP server started on port %d", strconv.Itoa(*port))
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}

