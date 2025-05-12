#!/bin/bash

set -e # Exit immediately if a command exits with a non-zero status.

MONOREPO_ROOT=$(pwd) # Assuming the script is run from the monorepo root
OUTPUT_DIR="${MONOREPO_ROOT}/bin"
mkdir -p "$OUTPUT_DIR"

echo "Ensuring common libraries are tidy and can compile..."
for lib_dir in libs/*; do
  if [ -d "$lib_dir" ] && [ -f "$lib_dir/go.mod" ]; then
    echo "--- Processing library: $lib_dir ---"
    ( # Run in a subshell to avoid cd side effects
      cd "$lib_dir"
      go mod tidy
      go build ./... # Check compilation
    )
    echo "--- Done checking: $lib_dir ---"
  fi
done
echo "Common libraries check complete."
echo ""

echo "Building services..."
for service_dir in services/*; do
  if [ -d "$service_dir" ] && [ -f "$service_dir/go.mod" ]; then
    SERVICE_NAME=$(basename "$service_dir")
    echo "--- Building service: $SERVICE_NAME ---"
    (
      cd "$service_dir"
      go mod tidy
      go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${SERVICE_NAME}_service" ./cmd/api/main.go
    )
    echo "--- Done: $SERVICE_NAME (Output: ${OUTPUT_DIR}/${SERVICE_NAME}_service) ---"
  fi
done
echo "Services build complete."
echo ""

echo "Building BFFs..."
for bff_dir in bff/*; do
  if [ -d "$bff_dir" ] && [ -f "$bff_dir/go.mod" ]; then
    BFF_NAME=$(basename "$bff_dir")
    echo "--- Building BFF: $BFF_NAME ---"
    (
      cd "$bff_dir"
      go mod tidy
      go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BFF_NAME}_bff" ./cmd/api/main.go
    )
    echo "--- Done: $BFF_NAME (Output: ${OUTPUT_DIR}/${BFF_NAME}_bff) ---"
  fi
done
echo "BFFs build complete."
echo ""

echo "Building CLI application..."
echo "--- Building CLI: digital-mono-cli ---"
( # Run in a subshell from monorepo root
  cd "$MONOREPO_ROOT" # Ensure we are at the root for CLI build if it uses root go.mod
  go mod tidy # Tidy root module
  go build -ldflags="-s -w" -o "${OUTPUT_DIR}/digital-mono-cli" ./cmd/digital-mono-cli/main.go
)
echo "--- Done: digital-mono-cli (Output: ${OUTPUT_DIR}/digital-mono-cli) ---"
echo ""

echo "All builds finished. Executables are in ${OUTPUT_DIR}"
echo "Run 'chmod +x ${OUTPUT_DIR}/*' to make them executable if needed."