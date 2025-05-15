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
	commonLoc "github.com/omni-compos/digital-mono/libs/localization"
	commonLogger "github.com/omni-compos/digital-mono/libs/logger"
	commonMetrics "github.com/omni-compos/digital-mono/libs/metrics"

	sellerGraphQL "github.com/omni-compos/digital-mono/services/seller/internal/handler/graphql"
	sellerREST "github.com/omni-compos/digital-mono/services/seller/internal/handler/rest"
	sellerRepo "github.com/omni-compos/digital-mono/services/seller/internal/repository"
	sellerService "github.com/omni-compos/digital-mono/services/seller/internal/service"
)

func main() {
	// Configuration
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		// Using a default DSN for local development if not set
		dbDSN = "host=localhost port=5432 user=omni_user password=strong_password dbname=digital_mono_db sslmode=disable"
		log.Println("Warning: DB_DSN not set, using default for seller service.")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Using a default JWT secret for local development if not set
		jwtSecret = "your-super-secret-key-for-user-service"
		log.Println("Warning: JWT_SECRET not set, using default for seller service.")
	}

	// Initialize common libraries
	appLogger := commonLogger.NewStdLogger()
	appLogger.Info("Starting seller service...")

	db, err := commonDB.NewPostgresDB(dbDSN)
	if err != nil {
		appLogger.Error(err, "Failed to connect to database")
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	appLogger.Info("Successfully connected to database")

	promMetrics := commonMetrics.NewPrometheusMetrics("seller_service", "api")
	authenticator := commonAuth.NewJWTAuthenticator(jwtSecret)

	// Initialize service-specific components
	repo := sellerRepo.NewPGSellerRepository(db)
	locService := commonLoc.NewDummyLocationalisationService() // Use the dummy locationalisation service
	service := sellerService.NewSellerService(repo, locService, appLogger)

	restHandler := sellerREST.NewSellerRESTHandler(service, appLogger, promMetrics)
	gqlHandler, err := sellerGraphQL.NewSellerGraphQLHandler(service, appLogger)
	if err != nil {
		appLogger.Error(err, "Failed to create GraphQL handler")
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	// Router
	r := mux.NewRouter()

	// REST API routes with JWT authentication middleware
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(authenticator.Middleware) // Apply JWT middleware to API routes
	restHandler.RegisterRoutes(apiRouter)

	// GraphQL endpoint
	// Note: GraphQL handler needs access to context for JWT claims if not handled by middleware
	// The graphql-go/handler can wrap middleware, or you can access context in resolvers
	// We'll rely on accessing context in resolvers as shown in seller_handler.go
	graphqlHTTPHandler := handler.New(&handler.Config{
		Schema:   &gqlHandler.Schema,
		Pretty:   true,
		GraphiQL: true, // Enable GraphiQL for easy testing
	})
	// Apply JWT middleware to GraphQL endpoint as well
	r.Handle("/graphql", authenticator.Middleware(graphqlHTTPHandler))

	// Metrics endpoint (usually doesn't require auth)
	r.Handle("/metrics", promMetrics.Handler())

	// Health check endpoint (optional, but good practice)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083" // Default port for seller service
	}
	appLogger.Info(fmt.Sprintf("Seller service listening on port %s", port))
	if err := http.ListenAndServe(":"+port, r); err != nil {
		appLogger.Error(err, "Failed to start server")
		log.Fatalf("Failed to start server: %v", err)
	}
}