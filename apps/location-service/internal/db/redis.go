package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisManager struct {
	client *redis.Client
	geoKey string
}

func NewRedisManager(redisAddr string) *RedisManager {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisManager{
		client: rdb,
		geoKey: "user_locations",
	}
}

// Adds or updates a user's location in Redis
func (r *RedisManager) UpdateUserLocation(ctx context.Context, userID string, latitude, longitude float64) error {
	return r.client.GeoAdd(ctx, r.geoKey, &redis.GeoLocation{
		Name:      userID,
		Latitude:  latitude,
		Longitude: longitude,
	}).Err()
}

// Returns user IDs within the radius (km) of the given location
func (r *RedisManager) GetNearbyUsers(ctx context.Context, latitude, longitude, radiusKm float64) ([]string, error) {
	locations, err := r.client.GeoRadius(ctx, r.geoKey, longitude, latitude, &redis.GeoRadiusQuery{
		Radius:    radiusKm,
		Unit:      "km",
		WithCoord: true,
		WithDist:  true,
		Count:     100, // limit to 100 results
		Sort:      "ASC",
	}).Result()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(locations))
	for _, loc := range locations {
		userIDs = append(userIDs, loc.Name)
	}
	return userIDs, nil
}
