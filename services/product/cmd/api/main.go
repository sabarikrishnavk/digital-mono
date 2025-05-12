package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq" // PostgreSQL driver

	commonDB "github.com/omni-compos/digital-mono/libs/database"
	commonLogger "github.com/omni-compos/digital-mono/libs/logger"
	commonMetrics "github.com/omni-compos/digital-mono/libs/metrics"

	productGraphQL "github.com/omni-compos/digital-mono/services/product/internal/handler/graphql"
	productREST "github.com/omni-compos/digital-mono/services/product/internal/handler/rest"
	productRepo "github.com/omni-compos/digital-mono/services/product/internal/repository"
	productService "github.com/omni-compos/digital-mono/services/product/internal/service"
)

func main() {
	// Configuration
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "host=localhost port=5432 user=omni_user password=strong_password dbname=digital_mono_db sslmode=disable"
		log.Println("Warning: DB_DSN not set, using default for product service.")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-key-for-product-service"
		log.Println("Warning: JWT_SECRET not set, using default for product service.")
	}

	// Initialize common libraries
	appLogger := commonLogger.NewStdLogger()
	appLogger.Info("Starting product service...")

	db, err := commonDB.NewPostgresDB(dbDSN)
	if err != nil {
		appLogger.Error(err, "Failed to connect to database")
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	appLogger.Info("Successfully connected to database")

	promMetrics := commonMetrics.NewPrometheusMetrics("product_service", "api")
	// authenticator := commonAuth.NewJWTAuthenticator(jwtSecret)

	// Dependency Injection
	repo := productRepo.NewPGProductRepository(db)
	service := productService.NewProductService(repo, appLogger)

	restHandler := productREST.NewProductRESTHandler(service, appLogger, promMetrics)
	gqlHandler, err := productGraphQL.NewProductGraphQLHandler(service, appLogger)
	if err != nil {
		appLogger.Error(err, "Failed to create GraphQL handler")
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	// Router
	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	// apiRouter.Use(authenticator.Middleware)
	restHandler.RegisterRoutes(apiRouter)

	graphqlHTTPHandler := handler.New(&handler.Config{
		Schema:   &gqlHandler.Schema,
		Pretty:   true,
		GraphiQL: true,
	})
	r.Handle("/graphql", graphqlHTTPHandler)
	// r.Handle("/metrics", promMetrics.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Default port for product service
	}
	appLogger.Info(fmt.Sprintf("Product service listening on port %s", port))
	if err := http.ListenAndServe(":"+port, r); err != nil {
		appLogger.Error(err, "Failed to start server")
		log.Fatalf("Failed to start server: %v", err)
	}
}