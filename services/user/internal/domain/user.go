package domain

import "time"

// User represents a user in the system.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Roles	  []string  `json:"roles"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response body for the login endpoint.
type LoginResponse struct {
	Token string `json:"token"`
}