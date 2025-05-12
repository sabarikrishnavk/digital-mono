package service

import (
	"context"
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