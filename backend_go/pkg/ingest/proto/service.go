package proto

import "google.golang.org/grpc"

type EventRequest struct{}
type EventResponse struct{ Success bool }

type UnimplementedEventServiceServer struct{}
type EventServiceServer interface{}

func RegisterEventServiceServer(s *grpc.Server, srv EventServiceServer) {}
