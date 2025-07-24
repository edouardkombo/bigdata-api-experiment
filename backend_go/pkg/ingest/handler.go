package ingest

import (
    "context"

    "project/pkg/kafka"
    proto "project/pkg/ingest/proto"
    "google.golang.org/grpc"
)

// server implements the EventServiceServer interface
type server struct {
    proto.UnimplementedEventServiceServer
}

// RegisterGRPCServer hooks up your service implementation to the gRPC server
func RegisterGRPCServer(s *grpc.Server) {
    proto.RegisterEventServiceServer(s, &server{})
}

// SendEvent receives an EventRequest and forwards it to Kafka
func (s *server) SendEvent(ctx context.Context, req *proto.EventRequest) (*proto.EventResponse, error) {
    if err := kafka.Publish(req); err != nil {
        return nil, err
    }
    return &proto.EventResponse{Success: true}, nil
}

