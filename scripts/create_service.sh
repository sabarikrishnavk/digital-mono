#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <service_name>"
  echo "Example: $0 user"
  exit 1
fi

SERVICE_NAME=$1
BASE_PATH="services/$SERVICE_NAME"
MODULE_PREFIX="github.com/omni-compos/digital-mono"
GO_VERSION="1.21" # Specify your desired Go version

echo "Creating service structure for '$SERVICE_NAME' in '$BASE_PATH'..."

# Create directories
mkdir -p "$BASE_PATH/cmd/api"
mkdir -p "$BASE_PATH/internal/app"
mkdir -p "$BASE_PATH/internal/domain"
mkdir -p "$BASE_PATH/internal/service"
mkdir -p "$BASE_PATH/internal/repository" # For PostgreSQL interactions
mkdir -p "$BASE_PATH/internal/handler/rest"
mkdir -p "$BASE_PATH/internal/handler/graphql"
mkdir -p "$BASE_PATH/api/rest" # For OpenAPI specs
mkdir -p "$BASE_PATH/api/graphql" # For GraphQL schemas
mkdir -p "$BASE_PATH/tests/unit"
mkdir -p "$BASE_PATH/tests/integration"

# Create placeholder Go files
echo "package main\n\nfunc main() {\n\tprintln(\"Hello from $SERVICE_NAME service\")\n}" > "$BASE_PATH/cmd/api/main.go"
touch "$BASE_PATH/internal/app/${SERVICE_NAME}_app.go"
touch "$BASE_PATH/internal/domain/${SERVICE_NAME}.go" # Core domain models
touch "$BASE_PATH/internal/service/${SERVICE_NAME}_service.go"
touch "$BASE_PATH/internal/repository/${SERVICE_NAME}_pg_repository.go" # PostgreSQL specific
touch "$BASE_PATH/internal/handler/rest/${SERVICE_NAME}_handler.go"
touch "$BASE_PATH/internal/handler/graphql/${SERVICE_NAME}_resolver.go"
touch "$BASE_PATH/api/rest/${SERVICE_NAME}.v1.yaml"
touch "$BASE_PATH/api/graphql/${SERVICE_NAME}.graphql"
touch "$BASE_PATH/tests/unit/${SERVICE_NAME}_service_test.go"
touch "$BASE_PATH/tests/integration/${SERVICE_NAME}_api_test.go"

# Create Dockerfile
echo "FROM golang:${GO_VERSION}-alpine AS builder\n\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api/main.go\n\nFROM alpine:latest\nWORKDIR /app\nCOPY --from=builder /app/server .\n\n# Add prometheus metrics port if needed, e.g., EXPOSE 9000\nEXPOSE 8080\nCMD [\"/app/server\"]" > "$BASE_PATH/Dockerfile"

# Create go.mod for the service
cat <<EOL > "$BASE_PATH/go.mod"
module ${MODULE_PREFIX}/${BASE_PATH}

go ${GO_VERSION}

require (
	// Example: ${MODULE_PREFIX}/libs/auth v0.0.0-unpublished
	// Example: ${MODULE_PREFIX}/libs/logger v0.0.0-unpublished
	// Example: ${MODULE_PREFIX}/libs/metrics v0.0.0-unpublished
	// Example: ${MODULE_PREFIX}/libs/database v0.0.0-unpublished
)

replace (
	// Example: ${MODULE_PREFIX}/libs/auth => ../../libs/auth
	// Example: ${MODULE_PREFIX}/libs/logger => ../../libs/logger
	// Example: ${MODULE_PREFIX}/libs/metrics => ../../libs/metrics
	// Example: ${MODULE_PREFIX}/libs/database => ../../libs/database
)
EOL

echo "Service '$SERVICE_NAME' structure created successfully in '$BASE_PATH'."
echo "Next steps:"
echo "1. cd $BASE_PATH"
echo "2. go mod tidy"
echo "3. Fill in the placeholder files with your service logic."
chmod +x "$0" # Make script executable if it's the first run from a new file