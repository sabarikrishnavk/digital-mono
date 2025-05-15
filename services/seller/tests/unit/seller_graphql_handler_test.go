package graphql_test

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	sellerGraphQL "github.com/omni-compos/digital-mono/services/seller/internal/handler/graphql"
	"github.com/omni-compos/digital-mono/services/seller/internal/domain"
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

// Helper to create a GraphQL schema with mocks
func newMockGraphQLSchema(t *testing.T) (*MockSellerService, *MockLogger, graphql.Schema) {
	mockService := new(MockSellerService)
	mockLogger := new(MockLogger)
	handler, err := sellerGraphQL.NewSellerGraphQLHandler(mockService, mockLogger)
	if err != nil {
		t.Fatalf("Failed to create GraphQL handler: %v", err)
	}
	return mockService, mockLogger, handler.Schema
}

// Helper to execute a GraphQL query/mutation
func executeGraphQLQuery(schema graphql.Schema, query string, variables map[string]interface{}, ctx context.Context) *graphql.Result {
	params := graphql.Params{
		Schema:         schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        ctx,
	}
	return graphql.Do(params)
}

func TestSellerGraphQLHandler_GetSellerQuery(t *testing.T) {
	mockService, mockLogger, schema := newMockGraphQLSchema(t)

	sellerID := "seller-123"
	expectedSeller := &model.Seller{ID: sellerID, BrandID: model.BrandIDBrandA, Status: model.StatusActive} // Simplified
	userID := "test-user-456"

	// Add UserID to context, simulating JWT middleware
	ctx := context.WithValue(context.Background(), commonAuth.UserIDContextKey, userID)

	// Expect the service call
	mockService.On("GetSellerByID", ctx, sellerID).Return(expectedSeller, nil).Once()

	query := `
		query GetSeller($id: String!) {
			seller(id: $id) {
				id
				brandId
				status
			}
		}
	`
	variables := map[string]interface{}{
		"id": sellerID,
	}

	result := executeGraphQLQuery(schema, query, variables, ctx)

	assert.Empty(t, result.Errors)
	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok)
	sellerData, ok := data["seller"].(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, expectedSeller.ID, sellerData["id"])
	assert.Equal(t, expectedSeller.BrandID, sellerData["brandId"])
	assert.Equal(t, expectedSeller.Status, sellerData["status"])

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSellerGraphQLHandler_CreateSellerMutation(t *testing.T) {
	mockService, mockLogger, schema := newMockGraphQLSchema(t)

	inputSellerArgs := map[string]interface{}{
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
	userID := "test-user-456"
	createdSeller := &model.Seller{ID: "new-seller-id", BrandID: model.BrandIDBrandA, Status: model.StatusActive, LastUpdatedBy: userID} // Simplified

	// Add UserID to context, simulating JWT middleware
	ctx := context.WithValue(context.Background(), commonAuth.UserIDContextKey, userID)

	// Expect the service call
	// Use mock.AnythingOfType to match the seller object passed to the service
	mockService.On("CreateSeller", ctx, mock.AnythingOfType("*model.Seller"), userID).
		Return(createdSeller, nil).Once()

	mutation := `
		mutation CreateSeller($input: CreateSellerInput!) {
			createSeller(input: $input) {
				id
				brandId
				status
				lastUpdatedBy
			}
		}
	`
	// GraphQL input types need to be defined or inferred.
	// For simplicity in the test, we'll pass the args directly if the schema allows it,
	// or define a simple input type structure if needed.
	// The current schema definition uses direct args, not an input object type.
	// Let's adjust the test query/variables to match the schema args.
	mutationArgs := `
		mutation CreateSeller(
			$brandId: String!, $status: String!, $address: String!, $city: String!,
			$state: String!, $country: String, $postcode: String!, $email: String!,
			$phoneNumber: String!
		) {
			createSeller(
				brandId: $brandId, status: $status, address: $address, city: $city,
				state: $state, country: $country, postcode: $postcode, email: $email,
				phoneNumber: $phoneNumber
			) {
				id
				brandId
				status
				lastUpdatedBy
			}
		}
	`

	result := executeGraphQLQuery(schema, mutationArgs, inputSellerArgs, ctx)

	assert.Empty(t, result.Errors)
	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok)
	createSellerData, ok := data["createSeller"].(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, createdSeller.ID, createSellerData["id"])
	assert.Equal(t, createdSeller.BrandID, createSellerData["brandId"])
	assert.Equal(t, createdSeller.Status, createSellerData["status"])
	assert.Equal(t, createdSeller.LastUpdatedBy, createSellerData["lastUpdatedBy"])

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSellerGraphQLHandler_CreateSellerMutation_Unauthorized(t *testing.T) {
	mockService, mockLogger, schema := newMockGraphQLSchema(t)

	inputSellerArgs := map[string]interface{}{
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
	// Context *without* UserID, simulating missing JWT claim
	ctx := context.Background()

	mutationArgs := `
		mutation CreateSeller(
			$brandId: String!, $status: String!, $address: String!, $city: String!,
			$state: String!, $country: String, $postcode: String!, $email: String!,
			$phoneNumber: String!
		) {
			createSeller(
				brandId: $brandId, status: $status, address: $address, city: $city,
				state: $state, country: $country, postcode: $postcode, email: $email,
				phoneNumber: $phoneNumber
			) {
				id
			}
		}
	`

	result := executeGraphQLQuery(schema, mutationArgs, inputSellerArgs, ctx)

	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Error(), "unauthorized")
	assert.Nil(t, result.Data) // Data should be nil on error

	mockService.AssertExpectations(t) // Service should not be called
	mockLogger.AssertExpectations(t)
}

// Add tests for updateSeller, deleteSeller, sellers (list) queries.
// Ensure error handling and authorization checks in resolvers are tested.