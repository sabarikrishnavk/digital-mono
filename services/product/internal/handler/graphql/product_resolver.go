package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/services/product/internal/service"
)

var productType *graphql.Object

func init() {
	productType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Product",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"description": &graphql.Field{Type: graphql.String},
			"sku":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"createdAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
}

// ProductGraphQLHandler holds the GraphQL schema and dependencies.
type ProductGraphQLHandler struct {
	Schema  graphql.Schema
	service service.ProductService
	logger  logger.Logger
}

// NewProductGraphQLHandler creates a new GraphQL handler for products.
func NewProductGraphQLHandler(productService service.ProductService, log logger.Logger) (*ProductGraphQLHandler, error) {
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"product": &graphql.Field{
				Type: productType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, nil
					}
					return productService.GetProduct(p.Context, id)
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"createProduct": &graphql.Field{
				Type: productType,
				Args: graphql.FieldConfigArgument{
					"name":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"description": &graphql.ArgumentConfig{Type: graphql.String}, // Optional
					"sku":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name := p.Args["name"].(string)
					sku := p.Args["sku"].(string)
					description := ""
					if desc, ok := p.Args["description"].(string); ok {
						description = desc
					}
					product, err := productService.CreateProduct(p.Context, name, description, sku)
					if err != nil {
						log.Error(err, "GraphQL: Failed to create product")
						return nil, err
					}
					return product, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		return nil, err
	}
	return &ProductGraphQLHandler{Schema: schema, service: productService, logger: log}, nil
}