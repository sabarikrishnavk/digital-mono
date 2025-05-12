# digital-mono

This is the Go monorepo for "digital-mono" by omni-compos.

## Overview

This monorepo hosts:

- Domain Microservices: user, product, price, promotion
- BFF Microservices: cart, checkout, order, fulfillment

It includes common libraries for:

- JWT Authentication
- Standardized Logging
- Prometheus Metrics

Deployment configurations are provided for:

- Docker Compose (including Prometheus & Grafana)
- Kubernetes
- Linux Package (for a potential desktop/CLI application)

## Getting Started

1.  **Scripts for Service/BFF Creation**:
    - Use `scripts/create_service.sh <service_name>` to scaffold a new domain microservice.
    - Use `scripts/create_bff.sh <bff_name>` to scaffold a new BFF microservice.

## Directory Structure

Refer to the established directory layout for services, BFFs, libraries, and deployment configurations.
