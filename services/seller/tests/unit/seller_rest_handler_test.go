package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	"github.com/omni-compos/digital-mono/libs/metrics"
	"github.com/omni-compos/digital-mono/services/seller/internal/handler/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations (re-using service mocks)
type MockSellerService struct {
	mock.Mock
}

func (m *MockSellerService) CreateSeller(ctx context.Context, seller *model.Seller, userID string) (*model.Seller, error) {
	args := m.Called(ctx, seller, userID)
	return args.Get(0).(*model.Seller), args.Error(1)
}

func (m *MockSellerService) GetSellerByID(ctx context.Context, id string) (*model.Seller, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Seller), args.Error(1)
}

func (m *MockSellerService) UpdateSeller(ctx context.Context, id string, updates *model.Seller, userID string) (*model.Seller, error) {
	args := m.Called(ctx, id, updates, userID)
	return args.Get(0).(*model.Seller), args.Error(1)
}

func (m *MockSellerService) DeleteSeller(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSellerService) ListSellers(ctx context.Context, limit, offset int) ([]*model.Seller, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Seller), args.Error(1)
}

type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) IncRequestsTotal(operation, handlerType string) {
	m.Called(operation, handlerType)
}
func (m *MockMetrics) IncResponsesTotal(operation, handlerType, code string) {
	m.Called(operation, handlerType, code)
}
func (m *MockMetrics) NewRequestDurationTimer(operation, handlerType string) metrics.RequestDurationTimer {
	args := m.Called(operation, handlerType)
	return args.Get(0).(metrics.RequestDurationTimer)
}
func (m *MockMetrics) Handler() http.Handler {
	args := m.Called()
	return args.Get(0).(http.Handler)
}

type MockRequestDurationTimer struct {
	mock.Mock
}

func (m *MockRequestDurationTimer) ObserveDuration() {
	m.Called()
}

// MockLogger (re-using service mock)
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *MockLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	m.Called(err, msg, keysAndValues)
}
func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

// Helper to create a handler with mocks
func newMockHandler(t *testing.T) (*MockSellerService, *MockLogger, *MockMetrics, *rest.SellerRESTHandler) {
	mockService := new(MockSellerService)
	mockLogger := new(MockLogger)
	mockMetrics := new(MockMetrics)
	handler := rest.NewSellerRESTHandler(mockService, mockLogger, mockMetrics)
	return mockService, mockLogger, mockMetrics, handler
}

// Helper to create a request with context
func newRequestWithContext(method, url string, body interface{}, userID string) *http.Request {
	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}

	req := httptest.NewRequest(method, url, &reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Add UserID to context, simulating JWT middleware
	ctx := context.WithValue(req.Context(), commonAuth.UserIDContextKey, userID)
	return req.WithContext(ctx)
}

func TestSellerRESTHandler_CreateSeller(t *testing.T) {
	mockService, mockLogger, mockMetrics, handler := newMockHandler(t)

	inputSeller := map[string]interface{}{
		"brandId":     model.BrandIDBrandA,
		"status":      model.StatusActive,
		"address":     "1 Test St",
		"city":        "Testville",
		"state":       "TS",
		"country":     "AUS",
		"postcode":    "1000",
		"email":       "test@example.com",
		"phoneNumber": "1234567890",
	}
	userID := "test-user-123"
	expectedSeller := &model.Seller{ID: "new-seller-id", BrandID: model.BrandIDBrandA, Status: model.StatusActive, LastUpdatedBy: userID} // Simplified

	req := newRequestWithContext("POST", "/api/v1/sellers", inputSeller, userID)
	rr := httptest.NewRecorder()

	mockTimer := new(MockRequestDurationTimer)
	mockMetrics.On("IncRequestsTotal", "create_seller", "rest").Once()
	mockMetrics.On("NewRequestDurationTimer", "create_seller", "rest").Return(mockTimer).Once()
	mockTimer.On("ObserveDuration").Once()
	mockMetrics.On("IncResponsesTotal", "create_seller", "rest", "201").Once()

	// Expect the service call
	mockService.On("CreateSeller", mock.Anything, mock.AnythingOfType("*model.Seller"), userID).
		Return(expectedSeller, nil).Once()

	// Use a router to handle the path matching
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	handler.RegisterRoutes(apiRouter)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var responseSeller model.Seller
	err := json.NewDecoder(rr.Body).Decode(&responseSeller)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeller.ID, responseSeller.ID)
	assert.Equal(t, expectedSeller.BrandID, responseSeller.BrandID)
	assert.Equal(t, expectedSeller.Status, responseSeller.Status)
	// Add more assertions for other fields

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t) // Should not be called on success
	mockMetrics.AssertExpectations(t)
	mockTimer.AssertExpectations(t)
}

func TestSellerRESTHandler_CreateSeller_InvalidPayload(t *testing.T) {
	mockService, mockLogger, mockMetrics, handler := newMockHandler(t)

	invalidInput := `{"brandId": 123}` // brandId should be string
	userID := "test-user-123"

	req := newRequestWithContext("POST", "/api/v1/sellers", bytes.NewBufferString(invalidInput), userID)
	rr := httptest.NewRecorder()

	mockTimer := new(MockRequestDurationTimer)
	mockMetrics.On("IncRequestsTotal", "create_seller", "rest").Once()
	mockMetrics.On("NewRequestDurationTimer", "create_seller", "rest").Return(mockTimer).Once()
	mockTimer.On("ObserveDuration").Once()
	mockMetrics.On("IncResponsesTotal", "create_seller", "rest", "400").Once()

	mockLogger.On("Error", mock.Anything, "Failed to decode request body for CreateSeller").Once() // Expect logger call

	// No service expectation as decoding should fail first

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	handler.RegisterRoutes(apiRouter)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request payload")

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTimer.AssertExpectations(t)
}

func TestSellerRESTHandler_CreateSeller_ServiceError(t *testing.T) {
	mockService, mockLogger, mockMetrics, handler := newMockHandler(t)

	inputSeller := map[string]interface{}{
		"brandId":     model.BrandIDBrandA,
		"status":      model.StatusActive,
		"address":     "1 Test St",
		"city":        "Testville",
		"state":       "TS",
		"country":     "AUS",
		"postcode":    "1000",
		"email":       "test@example.com",
		"phoneNumber": "1234567890",
	}
	userID := "test-user-123"
	serviceErr := errors.New("database error")

	req := newRequestWithContext("POST", "/api/v1/sellers", inputSeller, userID)
	rr := httptest.NewRecorder()

	mockTimer := new(MockRequestDurationTimer)
	mockMetrics.On("IncRequestsTotal", "create_seller", "rest").Once()
	mockMetrics.On("NewRequestDurationTimer", "create_seller", "rest").Return(mockTimer).Once()
	mockTimer.On("ObserveDuration").Once()
	mockMetrics.On("IncResponsesTotal", "create_seller", "rest", "500").Once()

	// Expect the service call to return an error
	mockService.On("CreateSeller", mock.Anything, mock.AnythingOfType("*model.Seller"), userID).
		Return((*model.Seller)(nil), serviceErr).Once()

	mockLogger.On("Error", serviceErr, "Failed to create seller via service").Once() // Expect logger call

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	handler.RegisterRoutes(apiRouter)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to create seller")

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTimer.AssertExpectations(t)
}

// Add tests for GetSellerByID, UpdateSeller, DeleteSeller, ListSellers
// Remember to test cases like "not found" (404), invalid ID (if applicable), service errors (500), and successful responses (200, 204).
// Ensure JWT context is handled (e.g., test missing UserID in context).

func TestSellerRESTHandler_GetSellerByID(t *testing.T) {
	mockService, mockLogger, mockMetrics, handler := newMockHandler(t)

	sellerID := "existing-seller-id"
	expectedSeller := &model.Seller{ID: sellerID, BrandID: model.BrandIDBrandA} // Simplified
	userID := "test-user-123" // UserID needed for middleware, but not used by GetByID service method in this example

	req := newRequestWithContext("GET", "/api/v1/sellers/"+sellerID, nil, userID)
	rr := httptest.NewRecorder()

	mockTimer := new(MockRequestDurationTimer)
	mockMetrics.On("IncRequestsTotal", "get_seller_by_id", "rest").Once()
	mockMetrics.On("NewRequestDurationTimer", "get_seller_by_id", "rest").Return(mockTimer).Once()
	mockTimer.On("ObserveDuration").Once()
	mockMetrics.On("IncResponsesTotal", "get_seller_by_id", "rest", "200").Once()

	mockService.On("GetSellerByID", mock.Anything, sellerID).Return(expectedSeller, nil).Once()

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	handler.RegisterRoutes(apiRouter)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var responseSeller model.Seller
	err := json.NewDecoder(rr.Body).Decode(&responseSeller)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeller.ID, responseSeller.ID)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTimer.AssertExpectations(t)
}

func TestSellerRESTHandler_GetSellerByID_NotFound(t *testing.T) {
	mockService, mockLogger, mockMetrics, handler := newMockHandler(t)

	sellerID := "non-existent-id"
	userID := "test-user-123"

	req := newRequestWithContext("GET", "/api/v1/sellers/"+sellerID, nil, userID)
	rr := httptest.NewRecorder()

	mockTimer := new(MockRequestDurationTimer)
	mockMetrics.On("IncRequestsTotal", "get_seller_by_id", "rest").Once()
	mockMetrics.On("NewRequestDurationTimer", "get_seller_by_id", "rest").Return(mockTimer).Once()
	mockTimer.On("ObserveDuration").Once()
	mockMetrics.On("IncResponsesTotal", "get_seller_by_id", "rest", "404").Once()

	mockService.On("GetSellerByID", mock.Anything, sellerID).Return((*model.Seller)(nil), nil).Once() // Service returns nil, nil for not found

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	handler.RegisterRoutes(apiRouter)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Seller not found")

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTimer.AssertExpectations(t)
}

// Add tests for UpdateSeller, DeleteSeller, ListSellers