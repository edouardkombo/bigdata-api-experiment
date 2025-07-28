package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	ingestpb "bigdata-perf/proto"
)

type server struct {
	ingestpb.UnimplementedEventServiceServer
	channel *amqp091.Channel
}

func (s *server) PublishEvent(ctx context.Context, req *ingestpb.EventRequest) (*ingestpb.EventResponse, error) {
	if req.Id == "" {
		req.Id = time.Now().Format("20060102150405")
	}
	req.Ts = time.Now().Format(time.RFC3339)

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = s.channel.Publish(
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
		return nil, err
	}

	return &ingestpb.EventResponse{Status: "queued", Id: req.Id}, nil
}

func main() {
	port := flag.Int("port", 50051, "Port to run the gRPC server on")
	flag.Parse()

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp091.Dial(rabbitmqURL)
	if err != nil {
		log.Fatalf("❌ RabbitMQ connection failed: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare("events", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Queue declaration failed: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	ingestpb.RegisterEventServiceServer(grpcServer, &server{channel: ch})

	log.Printf("🚀 gRPC server listening on port %d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ gRPC server failed: %v", err)
	}
}

