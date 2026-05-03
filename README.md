# idp-core

A foundational Internal Developer Platform (IDP) API built with Go. Provides self-service Kubernetes environment provisioning with GitOps integration.

## Features

- **Environment Management**: Create, list, get, and delete isolated Kubernetes environments
- **GitOps Integration**: Automatic ArgoCD Application creation and management
- **Live Status**: Real-time workload status via Kubernetes informers
- **Team Isolation**: Multi-tenant with team-scoped access control
- **Policy Enforcement**: Admission webhook for resource validation

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        cmd/http                                  │
│                    (Entry Point & Server)                        │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                    internal/handler/http                         │
│                    (HTTP Handlers / Controllers)                 │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                    internal/usecase                              │
│                       (Business Logic)                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                   internal/repository                            │
│                     (Data Access Layer)                          │
├─────────────────────┬─────────────────────┬─────────────────────┤
│      environment    │        k8s          │       argocd        │
│  (Environment DB)   │  (Kubernetes API)   │   (ArgoCD API)      │
└─────────────────────┴─────────────────────┴─────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for PostgreSQL)
- kubectl (for Kubernetes integration tests)
- kind (installed automatically via make)

### Local Development (PostgreSQL only)

```bash
# Start PostgreSQL in Docker
make dev-db-up

# Run the application with hot-reload
make dev-run

# Run tests
make test-unit
make test-db

# Stop PostgreSQL
make dev-db-down
```

### Kubernetes Integration Testing

```bash
# Setup Kind cluster with ArgoCD
make dev-k8s-setup

# Run Kubernetes integration tests
make test-k8s
make test-argocd

# Teardown
make dev-k8s-teardown
```

### Full Setup

```bash
# Setup everything (PostgreSQL + Kind + ArgoCD)
make dev-setup

# Run all tests
make test-all-integration

# Teardown everything
make dev-teardown
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/metrics` | Prometheus metrics |
| POST | `/v1/environments` | Create environment |
| GET | `/v1/environments` | List environments |
| GET | `/v1/environments/:id` | Get environment |
| DELETE | `/v1/environments/:id` | Delete environment |
| POST | `/v1/environments/:id/sync` | Trigger ArgoCD sync |
| GET | `/v1/environments/:id/status` | Get environment status |
| GET | `/v1/environments/:id/gitops/status` | Get ArgoCD status |
| GET | `/v1/environments/:id/workloads` | List workloads |
| GET | `/v1/environments/:id/workloads/:name` | Get workload details |
| POST | `/auth/login` | Login (get JWT token) |
| POST | `/admission/validate` | Admission webhook |

## Configuration

Configuration is loaded from `configs/config.yaml` with environment variable overrides.

```yaml
server:
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: idp_core
  sslmode: disable

auth:
  jwt_secret: "your-secret-key"

kubernetes:
  in_cluster: false
  kubeconfig_path: ""  # Defaults to ~/.kube/config

argocd:
  base_url: "http://argocd-server.argocd.svc.cluster.local:80"
  namespace: argocd
```

### Environment Variables

All configuration can be overridden via environment variables:

```bash
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=idp_core
K8S_IN_CLUSTER=false
ARGOCD_NAMESPACE=argocd
JWT_SECRET=your-secret-key
```

## Project Structure

```
idp-core/
├── cmd/
│   └── http/                    # Entry point
├── configs/
│   └── config.yaml              # Configuration
├── deployments/
│   └── kubernetes/              # K8s manifests
│       ├── base/                # Base manifests
│       └── overlays/            # Environment overlays
├── dev/                         # Development scripts
├── docs/                        # Documentation
├── internal/
│   ├── handler/http/            # HTTP handlers
│   ├── usecase/                 # Business logic
│   ├── repository/              # Data access
│   ├── model/                   # Domain models
│   ├── mocks/                   # Test mocks
│   └── pkg/                     # Internal packages
│       ├── config/              # Configuration
│       ├── kubernetes/          # K8s client
│       ├── argocd/              # ArgoCD client
│       ├── validator/           # Request validation
│       └── errors/              # Error types
├── Makefile                     # Build automation
├── Dockerfile                   # Container build
└── docker-compose.yml           # Local services
```

## Development

### Common Commands

```bash
# Development
make dev-db-up          # Start PostgreSQL
make dev-run            # Run with hot-reload
make dev-db-down        # Stop PostgreSQL

# Kubernetes
make dev-k8s-setup      # Setup Kind + ArgoCD
make dev-k8s-status     # Check K8s status
make dev-k8s-teardown   # Delete cluster

# Testing
make test-unit          # Unit tests (fast)
make test-db            # PostgreSQL tests
make test-k8s           # Kubernetes tests
make test-coverage      # Coverage report

# Code Quality
make lint               # Run golangci-lint
make fmt                # Format code

# Build
make build              # Build binary
make docker-build       # Build Docker image

# Swagger
make swagger-gen        # Generate OpenAPI docs
```

### Running Tests

```bash
# Unit tests only (no external dependencies)
make test-unit

# PostgreSQL integration tests (requires: make dev-db-up)
make test-db

# Kubernetes integration tests (requires: make dev-k8s-setup)
make test-k8s
make test-argocd

# All integration tests
make test-all-integration
```

## Deployment

### Kubernetes

```bash
# Apply base manifests
kubectl apply -k deployments/kubernetes/base/

# Or production overlay
kubectl apply -k deployments/kubernetes/overlays/production/
```

See [deployments/kubernetes/README.md](deployments/kubernetes/README.md) for detailed deployment instructions.

### Docker

```bash
# Build
docker build -t idp-core:latest .

# Run
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=secret \
  -e JWT_SECRET=your-secret \
  idp-core:latest
```

## Documentation

- [Development Guidelines](docs/DEV_GUIDELINE.md)
- [RBAC Permissions](docs/RBAC.md)
- [On-Call Runbook](docs/RUNBOOK.md)
- [Development TODO List](docs/DEV_TODO_LIST_PHASE_1.md)

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| Web Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Kubernetes Client | client-go |
| Configuration | Viper |
| Logging | zerolog |
| OpenAPI | swaggo |
| Testing | testify, gomock |

## License

MIT
