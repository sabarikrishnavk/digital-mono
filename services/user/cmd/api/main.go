package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq" // PostgreSQL driver

	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	commonDB "github.com/omni-compos/digital-mono/libs/database"
	commonLogger "github.com/omni-compos/digital-mono/libs/logger"
	commonMetrics "github.com/omni-compos/digital-mono/libs/metrics"

	userGraphQL "github.com/omni-compos/digital-mono/services/user/internal/handler/graphql"
	userREST "github.com/omni-compos/digital-mono/services/user/internal/handler/rest"
	userRepo "github.com/omni-compos/digital-mono/services/user/internal/repository"
	userService "github.com/omni-compos/digital-mono/services/user/internal/service"
)

func main() {
	println("Hello from $SERVICE_NAME service")
	// Configuration (ideally from env vars or config file)
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "host=localhost port=5432 user=omni_user password=strong_password dbname=digital_mono_db sslmode=disable"
		log.Println("Warning: DB_DSN not set, using default.")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-key-for-user-service" // Use a strong, unique secret
		log.Println("Warning: JWT_SECRET not set, using default.")
	}

	// Initialize common libraries
	appLogger := commonLogger.NewStdLogger() // Replace with your actual logger implementation
	appLogger.Info("Starting user service...")

	db, err := commonDB.NewPostgresDB(dbDSN)
	if err != nil {
		appLogger.Error(err, "Failed to connect to database")
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	appLogger.Info("Successfully connected to database")

	// Initialize Prometheus metrics (placeholder, replace with actual implementation)
	promMetrics := commonMetrics.NewPrometheusMetrics("user_service", "api")

	// Initialize Auth
	authenticator := commonAuth.NewJWTAuthenticator(jwtSecret)

	// Dependency Injection
	repo := userRepo.NewPGUserRepository(db)
	service := userService.NewUserService(repo, appLogger)

	restHandler := userREST.NewUserRESTHandler(service, jwtSecret, appLogger, promMetrics)
	gqlHandler, err := userGraphQL.NewUserGraphQLHandler(service, appLogger)
	if err != nil {
		appLogger.Error(err, "Failed to create GraphQL handler")
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	// Router
	r := mux.NewRouter()

	// // REST API routes
	// apiRouter := r.PathPrefix("/api/v1").Subrouter()
	// // apiRouter.Use(authenticator.Middleware) // Apply JWT auth middleware
	// restHandler.RegisterRoutes(apiRouter)

	// Register public routes (like login) BEFORE applying middleware
	restHandler.RegisterRoutes(r) // Register all routes, including /login

	// REST API routes
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(authenticator.Middleware) 
	// GraphQL endpoint
	// You might want to protect this with authenticator.Middleware as well
	graphqlHTTPHandler := handler.New(&handler.Config{
		Schema:   &gqlHandler.Schema,
		Pretty:   true,
		GraphiQL: true, // Enable GraphiQL UI at /graphql
	})
	r.Handle("/graphql", graphqlHTTPHandler)

	// Prometheus metrics endpoint
	r.Handle("/metrics", promMetrics.Handler()) // Assuming your metrics lib provides an http.Handler

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port for user service
	}
	appLogger.Info(fmt.Sprintf("User service listening on port %s", port))
	if err := http.ListenAndServe(":"+port, r); err != nil {
		appLogger.Error(err, "Failed to start server")
		log.Fatalf("Failed to start server: %v", err)
	}
}