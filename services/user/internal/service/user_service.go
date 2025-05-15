package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/services/user/internal/domain"
	"github.com/omni-compos/digital-mono/services/user/internal/repository"
)

// UserService defines the interface for user business logic.
type UserService interface {
	CreateUser(ctx context.Context, name, email string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
		// AuthenticateUser is a placeholder for user authentication logic.
	// In a real application, this would involve checking password hashes, etc.
	AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error)

}

type userService struct {
	repo   repository.UserRepository
	logger logger.Logger // Using the common logger interface
}

// NewUserService creates a new user service.
func NewUserService(repo repository.UserRepository, log logger.Logger) UserService {
	return &userService{repo: repo, logger: log}
}

func (s *userService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	s.logger.Info("Creating user", "name", name, "email", email)
	now := time.Now()
	user := &domain.User{
		ID:        uuid.NewString(),
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return user, s.repo.CreateUser(ctx, user)
}

func (s *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	s.logger.Info("Getting user", "id", id)
	return s.repo.GetUserByID(ctx, id)
}



// AuthenticateUser is a placeholder implementation.
// Replace with actual authentication logic (e.g., password hashing comparison).
func (s *userService) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	s.logger.Info("Attempting to authenticate user", "email", email)

	// --- Placeholder Authentication Logic ---
	// In a real app:
	// 1. Retrieve user by email from the repository.
	// 2. Compare the provided password with the stored hashed password.
	// 3. Return the user object if authentication is successful.
	// 4. Return an error (e.g., ErrInvalidCredentials) if authentication fails.

	// Dummy logic: Assume authentication is successful if email and password are not empty
	if email != "" && password != "" {
		// In a real scenario, you'd fetch the actual user from the DB here
		return &domain.User{ID: "dummy-user-id-for-auth", Email: email, Roles: []string{"user"}}, nil // Return a dummy user
	}
	return nil, fmt.Errorf("invalid credentials") // Or a specific authentication error type
}