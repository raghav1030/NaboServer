package handler

import (
	"context"
	"log"

	locationv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/location/v1"
	"github.com/raghav1030/NaboServer/apps/location-service/internal/service"
)

type LocationServer struct {
	locationv1.UnimplementedLocationServiceServer
	locationService *service.LocationService // business logic
}

func NewLocationServer(locationService *service.LocationService) *LocationServer {
	return &LocationServer{locationService: locationService}
}

// Handles UpdateLocation gRPC requests
func (s *LocationServer) UpdateLocation(ctx context.Context, req *locationv1.UpdateLocationRequest) (*locationv1.UpdateLocationResponse, error) {
	log.Printf("Received UpdateLocation: lat=%f, lon=%f", req.Latitude, req.Longitude)
	success, err := s.locationService.UpdateUserLocation(ctx, req)
	return &locationv1.UpdateLocationResponse{Success: success}, err
}

// Handles GetNearbyUsers gRPC requests
func (s *LocationServer) GetNearbyUsers(ctx context.Context, req *locationv1.GetNearbyUsersRequest) (*locationv1.GetNearbyUsersResponse, error) {
	log.Printf("Received GetNearbyUsers: lat=%f, lon=%f, radius=%f", req.Latitude, req.Longitude, req.RadiusKm)
	users, err := s.locationService.FindNearbyUsers(ctx, req)
	return &locationv1.GetNearbyUsersResponse{Users: users}, err
}
