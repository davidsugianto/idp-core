# idp-core

**Internal Developer Platform (IDP) API** — Self-service Kubernetes environment provisioning with GitOps integration.

## Overview

`idp-core` enables engineering teams to self-provision Kubernetes environments on-demand, trigger GitOps deployments via ArgoCD, and monitor live workload status through a clean REST API.

**Vision**: Reduce environment provisioning time from days to minutes while abstracting Kubernetes complexity from developers.

## Features

| Feature | Description |
|---------|-------------|
| 🚀 **Environment Management** | Create, list, get, delete isolated K8s namespaces |
| 🎯 **GitOps Integration** | Automatic ArgoCD Application creation and sync |
| 📊 **Live Status** | Real-time workload status via Kubernetes informers |
| 🔐 **Authentication** | JWT-based auth with team context |
| 👥 **User & Team Management** | Multi-tenant with team-scoped resources |
| 🔑 **RBAC** | Role-based access control with permissions |
| 🛡️ **Policy Enforcement** | Admission webhook for resource validation |

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| Web Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL 15+ |
| Kubernetes | client-go |
| GitOps | ArgoCD |
| API Docs | Swagger/OpenAPI |

## Quick Start

```bash
# Install dependencies
make bootstrap

# Start PostgreSQL
make dev-db-up

# Run with hot-reload
make dev-run

# Run tests
make test
```

## API Endpoints

### Core

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/ping` | Health check |
| POST | `/auth/login` | Authenticate user |
| GET | `/swagger/*any` | API documentation |

### Environments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/environments` | List environments |
| POST | `/v1/environments` | Create environment |
| GET | `/v1/environments/:id` | Get environment |
| DELETE | `/v1/environments/:id` | Delete environment |
| POST | `/v1/environments/:id/sync` | Trigger GitOps sync |
| GET | `/v1/environments/:id/status` | Get environment status |

### Users & Teams

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/users` | List users |
| POST | `/v1/users` | Create user |
| GET | `/v1/users/:id` | Get user |
| GET | `/v1/teams` | List teams |
| POST | `/v1/teams` | Create team |
| GET | `/v1/teams/:id/members` | List team members |

### RBAC

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/roles` | List roles |
| POST | `/v1/roles` | Create role |
| POST | `/v1/roles/assign` | Assign role to user |
| POST | `/v1/roles/revoke` | Revoke role from user |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Handler Layer                     │
│         (HTTP handlers, validation, auth)           │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                    UseCase Layer                     │
│            (Business logic, orchestration)           │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                  Repository Layer                    │
│       (Data access, K8s, ArgoCD, PostgreSQL)         │
└─────────────────────────────────────────────────────┘
```

**Layer Rules**: Handler → UseCase → Repository. Dependencies flow inward.

## Development

```bash
# Development
make dev-db-up        # Start PostgreSQL
make dev-run          # Run with hot-reload
make dev-db-down      # Stop PostgreSQL

# Testing
make test             # Run all tests
make test-coverage    # Coverage report
make lint             # Run linter

# K8s Integration
make dev-k8s-setup    # Setup Kind + ArgoCD
make test-k8s         # Run K8s tests

# Code Generation
make swagger-gen      # Generate API docs
```

## Configuration

Configuration via `configs/config.yaml` or environment variables:

```bash
SERVER_PORT=8989
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=idp_core
JWT_SECRET=your-secret
```

## Project Structure

```
idp-core/
├── cmd/http/           # Entry point
├── internal/
│   ├── handler/http/   # HTTP handlers
│   ├── usecase/        # Business logic
│   ├── repository/     # Data access
│   ├── model/          # Domain models
│   └── pkg/            # Internal packages
├── docs/               # Documentation
├── configs/            # Configuration
└── migrations/         # Database migrations
```

## Roadmap

| Phase | Timeline | Status | Focus |
|-------|----------|--------|-------|
| Phase 1 - MVP | Q2 2026 | ✅ Complete | Core API, K8s/ArgoCD |
| Phase 2 - Enhancement | Q3 2026 | 🔄 In Progress | RBAC, FinOps, Rightsizing |
| Phase 3 - Platform | Q4 2026 | 📋 Planned | UI, Templates |
| Phase 4 - Advanced | Q1 2027+ | 🔮 Roadmap | AI/ML, Analytics |

## Documentation

- [PRD Overview](docs/prd/PRD.md) — Product requirements
- [PRD Phase 2](docs/prd/PRD_PHASE_2.md) — Enhancement requirements
- [Dev TODO List](docs/DEV_TODO_LIST_PHASE_2.md) — Current progress
- [Development Guidelines](docs/DEV_GUIDELINE.md) — Coding standards

## License

MIT
