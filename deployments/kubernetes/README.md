# Kubernetes Configuration

This directory will contain Kubernetes manifests for deploying the microservices, BFFs, PostgreSQL, Prometheus, and Grafana.

## Structure

- `services/`: Contains deployments, services, configmaps, etc., for each domain microservice (e.g., `user-deployment.yaml`, `user-service.yaml`).
- `bffs/`: Contains deployments, services, etc., for each BFF (e.g., `cart-deployment.yaml`).
- `database/`: Contains statefulsets, services for PostgreSQL (e.g., `postgres-statefulset.yaml`).
- `monitoring/`: Contains configurations for Prometheus and Grafana (e.g., `prometheus-deployment.yaml`, `grafana-deployment.yaml`).
- `rbac/`: Contains roles, rolebindings, serviceaccounts if needed.
