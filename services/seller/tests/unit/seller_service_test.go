package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omni-compos/digital-mono/services/seller/internal/domain"
	"github.com/omni-compos/digital-mono/services/seller/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockSellerRepository struct {
	mock.Mock
}

func (m *MockSellerRepository) CreateSeller(ctx context.Context, seller *model.Seller) error {
	args := m.Called(ctx, seller)
	return args.Error(0)
}

func (m *MockSellerRepository) GetSellerByID(ctx context.Context, id string) (*model.Seller, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Seller), args.Error(1)
}

func (m *MockSellerRepository) UpdateSeller(ctx context.Context, seller *model.Seller) error {
	args := m.Called(ctx, seller)
	return args.Error(0)
}

func (m *MockSellerRepository) DeleteSeller(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSellerRepository) ListSellers(ctx context.Context, limit, offset int) ([]*model.Seller, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Seller), args.Error(1)
}

type MockLocationalisationService struct {
	mock.Mock
}

func (m *MockLocationalisationService) GetLatLngFromAddress(ctx context.Context, address, city, state, country, postcode string) (latitude, longitude float64, err error) {
	args := m.Called(ctx, address, city, state, country, postcode)
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

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

func TestDefaultSellerService_CreateSeller(t *testing.T) {
	mockRepo := new(MockSellerRepository)
	mockLoc := new(MockLocationalisationService)
	mockLogger := new(MockLogger)
	sellerService := service.NewSellerService(mockRepo, mockLoc, mockLogger)
	ctx := context.Background()
	userID := "test-user-123"

	inputSeller := model.NewSeller()
	inputSeller.BrandID = model.BrandIDBrandA
	inputSeller.Status = model.StatusActive
	inputSeller.Address = "1 Test St"
	inputSeller.City = "Testville"
	inputSeller.State = "TS"
	inputSeller.Country = "AUS"
	inputSeller.Postcode = "1000"
	inputSeller.Email = "test@example.com"
	inputSeller.PhoneNumber = "1234567890"

	expectedLat := -34.0
	expectedLng := 151.0

	// Setup mock expectations
	mockLoc.On("GetLatLngFromAddress", ctx, inputSeller.Address, inputSeller.City, inputSeller.State, inputSeller.Country, inputSeller.Postcode).
		Return(expectedLat, expectedLng, nil).Once()

	// Expect CreateSeller to be called with a seller object that has ID, Lat/Lng, and audit fields set
	mockRepo.On("CreateSeller", ctx, mock.AnythingOfType("*model.Seller")).
		Return(nil).Once()

	mockLogger.On("Info", "Seller created successfully", mock.Anything, mock.Anything).Maybe() // Expect logger call

	// Call the service method
	createdSeller, err := sellerService.CreateSeller(ctx, inputSeller, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, createdSeller)
	assert.NotEmpty(t, createdSeller.ID) // Check if ID was generated
	assert.Equal(t, expectedLat, createdSeller.Latitude)
	assert.Equal(t, expectedLng, createdSeller.Longitude)
	assert.Equal(t, userID, createdSeller.LastUpdatedBy)
	assert.WithinDuration(t, time.Now(), createdSeller.LastUpdateTime, time.Second) // Check if time is recent
	assert.Equal(t, inputSeller.BrandID, createdSeller.BrandID)
	assert.Equal(t, inputSeller.Status, createdSeller.Status)
	// Add assertions for other fields copied from input

	// Verify mock expectations
	mockLoc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDefaultSellerService_CreateSeller_InvalidBrandID(t *testing.T) {
	mockRepo := new(MockSellerRepository)
	mockLoc := new(MockLocationalisationService)
	mockLogger := new(MockLogger)
	sellerService := service.NewSellerService(mockRepo, mockLoc, mockLogger)
	ctx := context.Background()
	userID := "test-user-123"

	inputSeller := model.NewSeller()
	inputSeller.BrandID = "INVALID_BRAND" // Invalid value
	inputSeller.Status = model.StatusActive
	// ... other required fields

	// No mock expectations for loc or repo as validation should fail first

	createdSeller, err := sellerService.CreateSeller(ctx, inputSeller, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid brand ID")
	assert.Nil(t, createdSeller)

	mockLoc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDefaultSellerService_CreateSeller_GeocodingError(t *testing.T) {
	mockRepo := new(MockSellerRepository)
	mockLoc := new(MockLocationalisationService)
	mockLogger := new(MockLogger)
	sellerService := service.NewSellerService(mockRepo, mockLoc, mockLogger)
	ctx := context.Background()
	userID := "test-user-123"

	inputSeller := model.NewSeller()
	inputSeller.BrandID = model.BrandIDBrandA
	inputSeller.Status = model.StatusActive
	inputSeller.Address = "Bad Address"
	// ... other required fields

	locError := errors.New("geocoding failed")
	mockLoc.On("GetLatLngFromAddress", ctx, inputSeller.Address, inputSeller.City, inputSeller.State, inputSeller.Country, inputSeller.Postcode).
		Return(0.0, 0.0, locError).Once()

	mockLogger.On("Error", locError, "Failed to get lat/lng for seller", mock.Anything, mock.Anything).Maybe() // Expect logger call

	// No repo expectation as it shouldn't be called

	createdSeller, err := sellerService.CreateSeller(ctx, inputSeller, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to geocode address")
	assert.Nil(t, createdSeller)

	mockLoc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Add tests for GetSellerByID, UpdateSeller, DeleteSeller, ListSellers
// Mock the repository calls and assert the service logic (e.g., handling not found, setting audit fields on update).

func TestDefaultSellerService_GetSellerByID(t *testing.T) {
	mockRepo := new(MockSellerRepository)
	mockLoc := new(MockLocationalisationService)
	mockLogger := new(MockLogger)
	sellerService := service.NewSellerService(mockRepo, mockLoc, mockLogger)
	ctx := context.Background()
	sellerID := "existing-id"

	expectedSeller := &model.Seller{ID: sellerID, BrandID: model.BrandIDBrandA, Status: model.StatusActive} // Simplified

	mockRepo.On("GetSellerByID", ctx, sellerID).Return(expectedSeller, nil).Once()

	seller, err := sellerService.GetSellerByID(ctx, sellerID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSeller, seller)

	mockRepo.AssertExpectations(t)
	mockLoc.AssertExpectations(t) // Should not be called
	mockLogger.AssertExpectations(t)
}

func TestDefaultSellerService_GetSellerByID_NotFound(t *testing.T) {
	mockRepo := new(MockSellerRepository)
	mockLoc := new(MockLocationalisationService)
	mockLogger := new(MockLogger)
	sellerService := service.NewSellerService(mockRepo, mockLoc, mockLogger)
	ctx := context.Background()
	sellerID := "non-existent-id"

	mockRepo.On("GetSellerByID", ctx, sellerID).Return((*model.Seller)(nil), nil).Once() // Simulate not found

	seller, err := sellerService.GetSellerByID(ctx, sellerID)

	assert.NoError(t, err) // Service returns nil, nil for not found
	assert.Nil(t, seller)

	mockRepo.AssertExpectations(t)
	mockLoc.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Add more tests covering update logic, delete logic, list logic, and error handling.