package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/services/seller/internal/domain"
	"github.com/omni-compos/digital-mono/services/seller/internal/service"
)

// SellerGraphQLHandler holds the GraphQL schema and dependencies.
type SellerGraphQLHandler struct {
	Schema graphql.Schema
	logger logger.Logger
}

// NewProductGraphQLHandler creates a new SellerGraphQLHandler.
func NewSellerGraphQLHandler(service service.SellerService, logger logger.Logger) (*SellerGraphQLHandler, error) {
	// Define the Seller object type
	sellerType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Seller",
			Fields: graphql.Fields{
				"id":             &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"brandId":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"status":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"address":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"city":           &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"state":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"country":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"postcode":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"email":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"phoneNumber":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"latitude":       &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
				"longitude":      &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
				"lastUpdatedBy":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				"lastUpdateTime": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
			},
		},
	)

	// Define the root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"seller": &graphql.Field{
				Type: sellerType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("invalid seller ID")
					}
					// Authentication check (optional here if middleware handles it, but good practice in resolvers too)
					// _, authOK := p.Context.Value(commonAuth.UserIDContextKey).(string)
					// if !authOK {
					// 	return nil, fmt.Errorf("unauthorized")
					// }
					return service.GetSellerByID(p.Context, id)
				},
			},
			"sellers": &graphql.Field{
				Type: graphql.NewList(sellerType),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 10,
					},
					"offset": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 0,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					limit, _ := p.Args["limit"].(int)
					offset, _ := p.Args["offset"].(int)
					// Authentication check
					// _, authOK := p.Context.Value(commonAuth.UserIDContextKey).(string)
					// if !authOK {
					// 	return nil, fmt.Errorf("unauthorized")
					// }
					return service.ListSellers(p.Context, limit, offset)
				},
			},
		},
	})

	// Define the root mutation
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"createSeller": &graphql.Field{
				Type: sellerType, // Return the created seller
				Args: graphql.FieldConfigArgument{
					"brandId":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"status":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"address":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"city":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"state":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"country":     &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "AUS"},
					"postcode":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"email":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"phoneNumber": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					// lat/lng, lastUpdatedBy, lastUpdateTime are set by the service
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// Get UserID from JWT claims in context
					// Get UserID from JWT claims in context 
					claims, ok := commonAuth.GetClaimsFromContext(p.Context)
					if !ok   {
						return nil, fmt.Errorf("unauthorized: user ID not found in context")
					}

					newSeller := domain.NewSeller() // Use the constructor for defaults
					newSeller.BrandID, _ = p.Args["brandId"].(string)
					newSeller.Status, _ = p.Args["status"].(string)
					newSeller.Address, _ = p.Args["address"].(string)
					newSeller.City, _ = p.Args["city"].(string)
					newSeller.State, _ = p.Args["state"].(string)
					// Check if country is provided, otherwise use default from NewSeller()
					if country, ok := p.Args["country"].(string); ok {
						newSeller.Country = country
					}
					newSeller.Postcode, _ = p.Args["postcode"].(string)
					newSeller.Email, _ = p.Args["email"].(string)
					newSeller.PhoneNumber, _ = p.Args["phoneNumber"].(string)

					return service.CreateSeller(p.Context, newSeller, claims.UserID)
				},
			},
			"updateSeller": &graphql.Field{
				Type: sellerType, // Return the updated seller
				Args: graphql.FieldConfigArgument{
					"id":          &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"brandId":     &graphql.ArgumentConfig{Type: graphql.String},
					"status":      &graphql.ArgumentConfig{Type: graphql.String},
					"address":     &graphql.ArgumentConfig{Type: graphql.String},
					"city":        &graphql.ArgumentConfig{Type: graphql.String},
					"state":       &graphql.ArgumentConfig{Type: graphql.String},
					"country":     &graphql.ArgumentConfig{Type: graphql.String},
					"postcode":    &graphql.ArgumentConfig{Type: graphql.String},
					"email":       &graphql.ArgumentConfig{Type: graphql.String},
					"phoneNumber": &graphql.ArgumentConfig{Type: graphql.String},
					// lat/lng, lastUpdatedBy, lastUpdateTime are set by the service
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("invalid seller ID")
					}

					// Get UserID from JWT claims in context 
					claims, ok := commonAuth.GetClaimsFromContext(p.Context)
					if !ok   {
						return nil, fmt.Errorf("unauthorized: user ID not found in context")
					}

					updates := &domain.Seller{} // Create a temporary struct for updates
					if brandID, ok := p.Args["brandId"].(string); ok {
						updates.BrandID = brandID
					}
					if status, ok := p.Args["status"].(string); ok {
						updates.Status = status
					}
					if address, ok := p.Args["address"].(string); ok {
						updates.Address = address
					}
					if city, ok := p.Args["city"].(string); ok {
						updates.City = city
					}
					if state, ok := p.Args["state"].(string); ok {
						updates.State = state
					}
					if country, ok := p.Args["country"].(string); ok {
						updates.Country = country
					}
					if postcode, ok := p.Args["postcode"].(string); ok {
						updates.Postcode = postcode
					}
					if email, ok := p.Args["email"].(string); ok {
						updates.Email = email
					}
					if phoneNumber, ok := p.Args["phoneNumber"].(string); ok {
						updates.PhoneNumber = phoneNumber
					}

					return service.UpdateSeller(p.Context, id, updates, claims.UserID)
				},
			},
			"deleteSeller": &graphql.Field{
				Type: graphql.Boolean, // Or a custom success type
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return false, fmt.Errorf("invalid seller ID")
					}
					// Authentication check
					// _, authOK := p.Context.Value(commonAuth.UserIDContextKey).(string)
					// if !authOK {
					// 	return false, fmt.Errorf("unauthorized")
					// }
					err := service.DeleteSeller(p.Context, id)
					if err != nil {
						return false, err // Return false and the error
					}
					return true, nil // Return true on success
				},
			},
		},
	})

	// Create the schema
	schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    rootQuery,
			Mutation: rootMutation,
		},
	)
	if err != nil {
		logger.Error(err, "Failed to create GraphQL schema")
		return nil, fmt.Errorf("failed to create GraphQL schema: %w", err)
	}

	return &SellerGraphQLHandler{
		Schema: schema,
		logger: logger,
	}, nil
}