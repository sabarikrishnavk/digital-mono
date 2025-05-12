package auth

import "fmt"

// GenerateJWT generates a JWT token (placeholder)
func GenerateJWT(userID string) (string, error) {
	return fmt.Sprintf("dummy-jwt-for-%s", userID), nil
}

// ValidateJWT validates a JWT token (placeholder)
func ValidateJWT(token string) (string, error) {
	return "user-id-from-token", nil
}