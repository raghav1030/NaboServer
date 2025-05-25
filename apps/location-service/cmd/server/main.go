package main

import (
	"log"
	"net"

	"github.com/raghav1030/NaboServer/apps/location-service/internal/handler"
	"github.com/raghav1030/NaboServer/apps/location-service/internal/service"
	locationv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/location/v1"
	"google.golang.org/grpc"
)

func main() {
	locationService := service.NewLocationService()

	locationServer := handler.NewLocationServer(locationService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	locationv1.RegisterLocationServiceServer(grpcServer, locationServer)

	log.Println("LocationService gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
