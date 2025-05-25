package service

import (
	"context"

	"github.com/raghav1030/NaboServer/apps/location-service/internal/db"
	locationv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/location/v1"
)

type LocationService struct {
	redisManager *db.RedisManager
}

func NewLocationService(redisAddr string) *LocationService {
	return &LocationService{
		redisManager: db.NewRedisManager(redisAddr),
	}
}

func (s *LocationService) UpdateUserLocation(ctx context.Context, req *locationv1.UpdateLocationRequest) (bool, error) {
	err := s.redisManager.UpdateUserLocation(ctx, req.GetUserId(), req.Latitude, req.Longitude)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *LocationService) FindNearbyUsers(ctx context.Context, req *locationv1.GetNearbyUsersRequest) ([]*locationv1.User, error) {
	userIDs, err := s.redisManager.GetNearbyUsers(ctx, req.Latitude, req.Longitude, req.RadiusKm)
	if err != nil {
		return nil, err
	}

	// For simplicity, create User objects with user_id only
	users := make([]*locationv1.User, 0, len(userIDs))
	for _, id := range userIDs {
		users = append(users, &locationv1.User{UserId: id})
	}
	return users, nil
}
