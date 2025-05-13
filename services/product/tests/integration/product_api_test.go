package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	commonLogger "github.com/omni-compos/digital-mono/libs/logger"
	commonMetrics "github.com/omni-compos/digital-mono/libs/metrics"

	// Assuming product service has similar structures
	"github.com/omni-compos/digital-mono/services/product/internal/domain"                   // Hypothetical
	productREST "github.com/omni-compos/digital-mono/services/product/internal/handler/rest" // Hypothetical
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductService for integration tests focusing on handlers
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) { // Hypothetical
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

// CreateProduct is needed to satisfy the service.ProductService interface
func (m *MockProductService) CreateProduct(ctx context.Context, arg1 string, arg2 string, arg3 string) (*domain.Product, error) { // Matched to want: CreateProduct(context.Context, string, string, string)
	args := m.Called(ctx, arg1, arg2, arg3)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}
// Add other mock methods as needed for your product service

func setupProductTestRouter(service *MockProductService, authenticator *commonAuth.JWTAuthenticator) *mux.Router {
	testLogger := commonLogger.NewStdLogger()
	promMetrics := commonMetrics.NewPrometheusMetrics("test_product", "api")

	// Ensure productREST.NewProductRESTHandler matches its actual signature
	restHandler := productREST.NewProductRESTHandler(service, testLogger, promMetrics) // Adjust if signature differs

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(authenticator.Middleware) // Apply JWT auth middleware
	restHandler.RegisterRoutes(apiRouter)   // Assuming RegisterRoutes exists

	return router
}

func TestProductAPI_GetProduct_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests; set INTEGRATION_TESTS to run.")
	}

	mockService := new(MockProductService)
	// Use a consistent test secret for generating and validating tokens
	jwtAuthenticator := commonAuth.NewJWTAuthenticator("test-jwt-secret-for-product-integration")
	router := setupProductTestRouter(mockService, jwtAuthenticator)

	productID := "prod-123"
	expectedProduct := &domain.Product{ID: productID, Name: "Test Product"} // Hypothetical, removed Price

	// --- Test Case 1: Valid Token ---
	mockService.On("GetProduct", mock.Anything, productID).Return(expectedProduct, nil).Once()

	validToken, err := jwtAuthenticator.GenerateToken("test-user-id", []string{"user", "product_viewer"}, time.Hour)
	assert.NoError(t, err)

	reqValid, _ := http.NewRequest(http.MethodGet, "/api/v1/products/"+productID, nil)
	reqValid.Header.Set("Authorization", "Bearer "+validToken)
	rrValid := httptest.NewRecorder()
	router.ServeHTTP(rrValid, reqValid)

	assert.Equal(t, http.StatusOK, rrValid.Code)
	var fetchedProduct domain.Product
	json.Unmarshal(rrValid.Body.Bytes(), &fetchedProduct)
	assert.Equal(t, expectedProduct.Name, fetchedProduct.Name)
	mockService.AssertCalled(t, "GetProduct", mock.Anything, productID)

	// --- Test Case 2: No Token ---
	reqNoToken, _ := http.NewRequest(http.MethodGet, "/api/v1/products/"+productID, nil)
	rrNoToken := httptest.NewRecorder()
	router.ServeHTTP(rrNoToken, reqNoToken)

	assert.Equal(t, http.StatusUnauthorized, rrNoToken.Code)
	assert.True(t, strings.Contains(rrNoToken.Body.String(), "Authorization header required"))

	// --- Test Case 3: Invalid Token ---
	reqInvalidToken, _ := http.NewRequest(http.MethodGet, "/api/v1/products/"+productID, nil)
	reqInvalidToken.Header.Set("Authorization", "Bearer aninvalidtokenstring")
	rrInvalidToken := httptest.NewRecorder()
	router.ServeHTTP(rrInvalidToken, reqInvalidToken)

	assert.Equal(t, http.StatusUnauthorized, rrInvalidToken.Code)
	assert.True(t, strings.Contains(rrInvalidToken.Body.String(), "Invalid token"))

	// --- Test Case 4: Expired Token ---
	expiredToken, err := jwtAuthenticator.GenerateToken("test-user-id", []string{"user"}, -time.Hour) // Token expired an hour ago
	assert.NoError(t, err)

	reqExpiredToken, _ := http.NewRequest(http.MethodGet, "/api/v1/products/"+productID, nil)
	reqExpiredToken.Header.Set("Authorization", "Bearer "+expiredToken)
	rrExpiredToken := httptest.NewRecorder()
	router.ServeHTTP(rrExpiredToken, reqExpiredToken)

	assert.Equal(t, http.StatusUnauthorized, rrExpiredToken.Code)
	assert.True(t, strings.Contains(rrExpiredToken.Body.String(), "Token has expired"))

	// Ensure GetProduct was only called once (for the valid token case)
	mockService.AssertNumberOfCalls(t, "GetProduct", 1)
	mockService.AssertExpectations(t)
}

// TODO: Add more integration tests for other product endpoints (create, update, delete)
// ensuring they also handle authentication correctly.
// Consider testing role-based access if your claims and handlers use roles.