# Development Environment

This directory contains configuration and scripts for setting up a local development environment for testing idp-core.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (for PostgreSQL)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (for K8s integration tests)
- [kind](https://kind.sigs.k8s.io/) (installed automatically via make)
- Go 1.25+

## Two Development Modes

### Local Development (PostgreSQL only)

For running the application locally with just a database:

```bash
# Start PostgreSQL in Docker
make dev-db-up

# Run the application
make dev-run

# Stop PostgreSQL when done
make dev-db-down
```

### Kubernetes Integration Testing

For testing Kubernetes and ArgoCD integration:

```bash
# Setup Kind cluster with ArgoCD
make dev-k8s-setup

# Or quick setup with minimal ArgoCD
make dev-k8s-setup-quick

# Check environment status
make dev-k8s-status

# Teardown when done
make dev-k8s-teardown
```

## FinOps Integration Testing

For testing cost tracking with OpenCost and Prometheus:

```bash
# After K8s setup, install Prometheus + OpenCost
make dev-finops-setup

# Check FinOps component status
make dev-finops-status

# Sync costs manually from the app
curl http://localhost:8989/v1/costs

# View cost data
curl http://localhost:8989/v1/costs?namespace=default
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `make dev-db-up` | Start PostgreSQL in Docker |
| `make dev-db-down` | Stop PostgreSQL |
| `make dev-k8s-setup` | Full K8s setup (Kind + ArgoCD) |
| `make dev-k8s-setup-quick` | Minimal K8s setup |
| `make dev-k8s-teardown` | Delete Kind cluster |
| `make dev-prometheus-setup` | Install Prometheus in Kind |
| `make dev-opencost-setup` | Install OpenCost in Kind |
| `make dev-finops-setup` | Full FinOps setup (Prometheus + OpenCost) |
| `make dev-finops-status` | Check FinOps component status |

## Files

```
dev/
├── kind-config.yaml           # Kind cluster configuration
├── setup-kind.sh              # Full K8s setup script
├── setup-argocd-minimal.sh    # Minimal ArgoCD setup (faster)
├── setup-prometheus.sh        # Prometheus setup for OpenCost
├── setup-opencost.sh          # OpenCost setup for cost tracking
├── teardown-kind.sh           # K8s teardown script
└── README.md                  # This file
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CLUSTER_NAME` | `idp-test` | Name of the kind cluster |
| `ARGOCD_VERSION` | `v2.11.0` | ArgoCD version to install |
| `PROMETHEUS_NAMESPACE` | `monitoring` | Namespace for Prometheus |
| `PROMETHEUS_CHART_VERSION` | `25.30.0` | Prometheus Helm chart version |
| `OPENCOST_NAMESPACE` | `opencost` | Namespace for OpenCost |
| `OPENCOST_CHART_VERSION` | `1.43.0` | OpenCost Helm chart version |
| `TIMEOUT` | `600` | Setup timeout in seconds |

## Accessing ArgoCD UI (Optional)

```bash
# Port-forward ArgoCD UI
make dev-k8s-argocd-ui

# Or manually:
kubectl port-forward svc/argocd-server -n argocd 8090:443

# Get initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d

# Open https://localhost:8090
# Username: admin
```

## Troubleshooting

### Cluster already exists

```bash
# Delete and recreate
make dev-k8s-teardown
make dev-k8s-setup
```

### Tests fail to connect to K8s

```bash
# Verify kubectl context
kubectl config current-context
# Should show: kind-idp-test

# If not, switch context
kubectl config use-context kind-idp-test
```

### ArgoCD not ready

```bash
# Check ArgoCD pods
kubectl get pods -n argocd

# Check ArgoCD logs
kubectl logs -n argocd deployment/argocd-server
```

### OpenCost verification

```bash
# Port-forward OpenCost API
kubectl port-forward svc/opencost -n opencost 9003:9003 &

# Check allocation data
curl -s "http://localhost:9003/allocation?window=1h" | jq .

# Check OpenCost logs
kubectl logs -n opencost deployment/opencost
```

### Prometheus verification

```bash
# Port-forward Prometheus UI
kubectl port-forward svc/prometheus-server -n monitoring 9090:80 &

# Verify targets are healthy
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health=="up") | .labels.job'
```
