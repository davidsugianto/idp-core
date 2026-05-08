# idp-core

A foundational Internal Developer Platform (IDP) API built with Go. Provides self-service Kubernetes environment provisioning with GitOps integration.

## Features

- **Environment Management**: Create, list, get, and delete isolated Kubernetes environments
- **GitOps Integration**: Automatic ArgoCD Application creation and management
- **Live Status**: Real-time workload status via Kubernetes informers
- **Team Isolation**: Multi-tenant with team-scoped access control
- **Policy Enforcement**: Admission webhook for resource validation
- **JWT Authentication**: Token-based authentication with team context

## Architecture

idp-core follows clean architecture with strict layer separation:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              External Systems                                │
│    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                   │
│    │ PostgreSQL  │    │ Kubernetes  │    │   ArgoCD    │                   │
│    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                   │
└───────────┼──────────────────┼──────────────────┼───────────────────────────┘
            │                  │                  │
            ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Repository Layer                                     │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐          │
│  │   Environment    │  │   Provisioner    │  │     GitOps       │          │
│  │   Repository     │  │   Repository     │  │   Repository     │          │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘          │
└───────────┼────────────────────┼────────────────────┼──────────────────────┘
            │                    │                    │
            ▼                    ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Use Case Layer                                     │
│                    ┌─────────────────────────────┐                          │
│                    │    Environment UseCase      │                          │
│                    │  - CreateEnvironment        │                          │
│                    │  - GetEnvironment           │                          │
│                    │  - DeleteEnvironment        │                          │
│                    │  - TriggerSync              │                          │
│                    │  - GetWorkloads             │                          │
│                    └──────────────┬──────────────┘                          │
└───────────────────────────────────┼─────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Handler Layer                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │    Auth      │  │ Environment  │  │   Webhook    │  │    Health    │    │
│  │   Handler    │  │   Handler    │  │   Handler    │  │   Handler    │    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│         │                 │                 │                 │             │
│  ┌──────▼─────────────────▼─────────────────▼─────────────────▼───────┐    │
│  │                    Middleware (JWT, Logging, RequestID)             │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Entry Point                                        │
│                    ┌─────────────────────────────┐                          │
│                    │        cmd/http/main.go     │                          │
│                    │    (Server Setup & Wire)    │                          │
│                    └─────────────────────────────┘                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layer Rules

| Layer | Depends On | Responsibility |
|-------|------------|----------------|
| Handler | UseCase | HTTP request/response handling, validation |
| UseCase | Repository | Business logic, orchestration |
| Repository | External APIs | Data access, external integrations |

### Dependency Injection

Each layer uses a `Dependencies` struct for constructor injection:

```go
type Dependencies struct {
    EnvironmentRepo environmentRepo.Repository
    ProvisionerRepo provisionerRepo.Repository
    GitOpsRepo      gitopsRepo.Repository
}

func New(deps Dependencies) Usecase { ... }
```

### Data Flow

```
Request → Middleware → Handler → UseCase → Repository → External System
                                                      ↓
Response ← Middleware ← Handler ← UseCase ← Repository ←
```

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for PostgreSQL and/or app)
- kubectl (for Kubernetes integration tests)
- kind (installed automatically via make)

### Local Development (PostgreSQL only)

```bash
# Start PostgreSQL in Docker
make dev-db-up

# Run the application with hot-reload (requires air)
make dev-run

# Or run app in Docker
make dev-app-up

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

# Access ArgoCD UI (https://localhost:8090)
make dev-k8s-argocd-ui

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
| GET | `/swagger/*any` | Swagger UI |
| POST | `/auth/login` | Login (get JWT token) |
| POST | `/v1/environments` | Create environment |
| GET | `/v1/environments` | List environments |
| GET | `/v1/environments/:id` | Get environment |
| DELETE | `/v1/environments/:id` | Delete environment |
| POST | `/v1/environments/:id/sync` | Trigger ArgoCD sync |
| GET | `/v1/environments/:id/status` | Get environment status |
| GET | `/v1/environments/:id/gitops/status` | Get ArgoCD status |
| GET | `/v1/environments/:id/workloads` | List workloads |
| GET | `/v1/environments/:id/workloads/:name` | Get workload details |
| POST | `/admission/validate` | Admission webhook |

## Configuration

Configuration is loaded from `configs/config.yaml` with environment variable overrides.

```yaml
server:
  port: 8989

database:
  host: postgres      # Docker service name (use localhost for local dev)
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
  token: ""
```

### Environment Variables

All configuration can be overridden via environment variables:

```bash
# Server
SERVER_PORT=8989

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=idp_core

# Auth
JWT_SECRET=your-secret-key

# Kubernetes
K8S_IN_CLUSTER=false
KUBECONFIG_PATH=~/.kube/config

# ArgoCD
ARGOCD_BASE_URL=http://argocd-server.argocd.svc.cluster.local:80
ARGOCD_TOKEN=your-token
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
│   ├── prd/                     # Product requirements
│   │   ├── PRD.md               # PRD overview
│   │   └── PRD_PHASE_1.md       # Phase 1 requirements
│   ├── DEV_GUIDELINE.md         # Development guidelines
│   ├── TEST.md                  # Test documentation
│   ├── RBAC.md                  # RBAC permissions
│   └── RUNBOOK.md               # On-call runbook
├── internal/
│   ├── handler/http/            # HTTP handlers
│   │   └── middleware/          # HTTP middleware (JWT, logging, etc.)
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
├── tests/
│   ├── e2e/                     # End-to-end tests
│   └── contract/                # OpenAPI contract tests
├── Makefile                     # Build automation
├── Dockerfile                   # Multi-stage build
└── docker-compose.yml           # Local services
```

## Development

### Common Commands

```bash
# Development
make dev-db-up          # Start PostgreSQL
make dev-run            # Run with hot-reload (requires air)
make dev-app-up         # Run app in Docker
make dev-db-down        # Stop PostgreSQL

# Kubernetes
make dev-k8s-setup      # Setup Kind + ArgoCD
make dev-k8s-setup-quick # Minimal ArgoCD setup
make dev-k8s-status     # Check K8s status
make dev-k8s-argocd-ui  # Port-forward ArgoCD UI
make dev-k8s-teardown   # Delete cluster

# Testing
make test-unit          # Unit tests (fast)
make test-db            # PostgreSQL tests
make test-k8s           # Kubernetes tests
make test-argocd        # ArgoCD tests
make test-e2e           # E2E tests
make test-contract      # OpenAPI contract tests
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

# E2E tests (requires: make dev-k8s-setup)
make test-e2e

# Contract tests (requires: make swagger-gen)
make swagger-gen
make test-contract

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
docker run -p 8989:8989 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=secret \
  -e JWT_SECRET=your-secret \
  idp-core:latest
```

## Documentation

- [Development Guidelines](docs/DEV_GUIDELINE.md) - Detailed development guide
- [Test Documentation](docs/TEST.md) - Testing strategy and E2E tests
- [RBAC Permissions](docs/RBAC.md) - Kubernetes RBAC requirements
- [On-Call Runbook](docs/RUNBOOK.md) - Operational runbook
- [PRD Overview](docs/prd/PRD.md) - Product requirements
- [Phase 1 MVP PRD](docs/prd/PRD_PHASE_1.md) - Detailed Phase 1 requirements
- [Development TODO List](docs/DEV_TODO_LIST_PHASE_1.md) - Phase 1 progress

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| Web Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Kubernetes Client | client-go |
| ArgoCD Client | REST API (v2.11+) |
| Configuration | Viper |
| Logging | zerolog |
| OpenAPI | swaggo |
| Testing | testify, gomock, testcontainers-go |

## Ports

| Service | Port |
|---------|------|
| API Server | 8989 |
| ArgoCD UI | 8090 |
| PostgreSQL | 5432 |

## License

MIT
