package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/omni-compos/digital-mono/libs/logger"

	//"github.com/omni-compos/digital-mono/services/user/internal/domain"
	"github.com/omni-compos/digital-mono/services/user/internal/service"
)

var userType *graphql.Object

func init() {
	userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"email":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.String)}, // Simplification
			"updatedAt": &graphql.Field{Type: graphql.NewNonNull(graphql.String)}, // Simplification
		},
	})
}

// UserGraphQLHandler holds the GraphQL schema and dependencies.
type UserGraphQLHandler struct {
	Schema  graphql.Schema
	service service.UserService
	logger  logger.Logger
}

// NewUserGraphQLHandler creates a new GraphQL handler for users.
func NewUserGraphQLHandler(userService service.UserService, log logger.Logger) (*UserGraphQLHandler, error) {
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, nil // Or an error
					}
					return userService.GetUser(p.Context, id)
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"createUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"email": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name := p.Args["name"].(string)
					email := p.Args["email"].(string)
					user, err := userService.CreateUser(p.Context, name, email)
					if err != nil {
						log.Error(err, "GraphQL: Failed to create user")
						return nil, err
					}
					// Map domain.User to a struct that graphql-go can serialize easily if needed,
					// or ensure domain.User fields match graphql.Fields (which they do here).
					return user, nil
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
	return &UserGraphQLHandler{Schema: schema, service: userService, logger: log}, nil
}