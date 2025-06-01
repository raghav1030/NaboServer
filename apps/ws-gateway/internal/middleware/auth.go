package middleware

import (
	"errors"
	"net/http"
	"strings"
)

// Dummy claims struct for testing
type Claims struct {
	UserID string
}

// Dummy JWT parser: for testing, treat the token as the user ID
func parseJWT(token string) (*Claims, error) {
	// In production, you would parse and validate the JWT with a library.
	// For dummy, just return the token as user ID if it's not empty.
	if token == "" {
		return nil, errors.New("invalid token")
	}
	return &Claims{UserID: token}, nil
}

// AuthenticateUser extracts the userID from a JWT token in the Authorization header.
func AuthenticateUser(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("Unauthorized")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := parseJWT(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
