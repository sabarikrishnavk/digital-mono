package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	// Import necessary packages from your service, similar to main.go
	// This is a simplified example; a real test might spin up the actual service
	// or use a test database.

	"github.com/gorilla/mux"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	commonLogger "github.com/omni-compos/digital-mono/libs/logger"
	commonMetrics "github.com/omni-compos/digital-mono/libs/metrics"
	"github.com/omni-compos/digital-mono/services/user/internal/domain"
	userREST "github.com/omni-compos/digital-mono/services/user/internal/handler/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock" // For mocking the service layer in this example
)

// MockUserService for integration tests focusing on handlers
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	args := m.Called(ctx, name, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// LoginHandler in the current implementation doesn't call a service method for authentication.
// If it did, e.g., service.AuthenticateUser(email, password), we'd mock that here.
// func (m *MockUserService) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
// 	args := m.Called(ctx, email, password)
// 	return args.Get(0).(*domain.User), args.Error(1)
// }
func TestUserAPI_CreateUser_Integration(t *testing.T) {
	// For a true integration test, you'd initialize a real DB and service.
	// Here, we'll mock the service layer to test the handler and routing.
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests; set INTEGRATION_TESTS to run.")
	}

	mockService := new(MockUserService)
	testLogger := commonLogger.NewStdLogger()
	promMetrics := commonMetrics.NewPrometheusMetrics("test", "api") // Dummy metrics
	// Use a test secret, can be different from the actual service secret for testing purposes
	jwtAuthenticator := commonAuth.NewJWTAuthenticator("test-jwt-secret-for-user-integration")

	restHandler := userREST.NewUserRESTHandler(mockService, testLogger, promMetrics, jwtAuthenticator)
	router := mux.NewRouter()
	restHandler.RegisterRoutes(router.PathPrefix("/api/v1").Subrouter())

	newUser := userREST.CreateUserRequest{Name: "Integration User", Email: "int@example.com"}
	expectedUser := &domain.User{ID: "some-uuid", Name: newUser.Name, Email: newUser.Email}
	mockService.On("CreateUser", mock.Anything, newUser.Name, newUser.Email).Return(expectedUser, nil)

	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdUser domain.User
	json.Unmarshal(rr.Body.Bytes(), &createdUser)
	assert.Equal(t, expectedUser.Name, createdUser.Name)
	assert.Equal(t, expectedUser.Email, createdUser.Email)

	mockService.AssertExpectations(t)
}

func TestUserAPI_Login_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests; set INTEGRATION_TESTS to run.")
	}

	mockService := new(MockUserService) // Login handler doesn't use service yet for auth
	testLogger := commonLogger.NewStdLogger()
	promMetrics := commonMetrics.NewPrometheusMetrics("test", "api")
	jwtAuthenticator := commonAuth.NewJWTAuthenticator("test-jwt-secret-for-login-integration")

	restHandler := userREST.NewUserRESTHandler(mockService, testLogger, promMetrics, jwtAuthenticator)
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	restHandler.RegisterRoutes(apiRouter) // This registers /login

	// Test successful login
	loginCreds := userREST.LoginRequest{Email: "test@example.com", Password: "password123"}
	body, _ := json.Marshal(loginCreds)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"], "Token should not be empty on successful login")

	// Test login with missing credentials (e.g., empty password)
	loginCredsMissing := userREST.LoginRequest{Email: "test@example.com", Password: ""}
	bodyMissing, _ := json.Marshal(loginCredsMissing)
	reqMissing, _ := http.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(bodyMissing))
	rrMissing := httptest.NewRecorder()

	router.ServeHTTP(rrMissing, reqMissing)
	assert.Equal(t, http.StatusUnauthorized, rrMissing.Code)

	// Note: The current LoginHandler in user_handler.go has mock authentication.
	// It doesn't actually validate credentials against a user service or database.
	// If it did, you would set up mockService expectations like:
	// mockService.On("AuthenticateUser", mock.Anything, loginCreds.Email, loginCreds.Password).Return(&domain.User{ID: "user-123"}, nil)
	// mockService.On("AuthenticateUser", mock.Anything, "wrong@example.com", "badpass").Return(nil, errors.New("invalid credentials"))

	mockService.AssertExpectations(t) // No calls expected to service yet for login
}

// TODO: Add tests for GetUser, GraphQL endpoint, error cases, auth middleware.