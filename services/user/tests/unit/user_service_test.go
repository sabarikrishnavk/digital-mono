package unit

import (
	"context"
	"testing"

	"github.com/omni-compos/digital-mono/libs/logger" // Mock or use a test logger
	"github.com/omni-compos/digital-mono/services/user/internal/domain"
	"github.com/omni-compos/digital-mono/services/user/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock type for the UserRepository type
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	// For logger, you can use a simple mock or a no-op logger for tests
	testLogger := logger.NewStdLogger() // Replace with a test-specific logger if needed

	userService := service.NewUserService(mockRepo, testLogger)

	name := "Test User"
	email := "test@example.com"

	// Setup expectation
	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.Name == name && user.Email == email
	})).Return(nil)

	createdUser, err := userService.CreateUser(context.Background(), name, email)

	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, name, createdUser.Name)
	assert.Equal(t, email, createdUser.Email)
	mockRepo.AssertExpectations(t)
}

// TODO: Add more tests for GetUser, error cases, etc.