# idp-core

**Internal Developer Platform (IDP) API** — Enables engineering teams to self-provision isolated Kubernetes environments, manage GitOps deployments via ArgoCD, and monitor workload status — all through a RESTful API without direct cluster access.

## Overview

`idp-core` is a self-service platform API that abstracts Kubernetes complexity from developers. Teams can request isolated environments (namespaces), deploy applications via GitOps, and monitor resources in real-time — without needing `kubectl` access or Kubernetes expertise.

**Key Capabilities:**

- 🚀 **Self-Service Provisioning** — Request K8s environments in minutes, not days
- 🎯 **GitOps Native** — ArgoCD integration for declarative, version-controlled deployments
- 📊 **Real-Time Visibility** — Live workload status, health, and resource metrics
- 👥 **Multi-Tenant** — Team-scoped resources with RBAC for secure access control
- 🔐 **Secure by Default** — JWT + API Key authentication, team isolation, policy enforcement
- 📋 **Audit Logging** — Automatic request logging with filterable audit trails
- 💰 **Cost Tracking** — OpenCost integration for real-time Kubernetes cost visibility by team/namespace
- 💸 **Budget Management** — Set spend limits with Slack alerts at configurable thresholds
- ⚖️ **Rightsizing** — CPU/memory optimization recommendations based on Prometheus usage data
- 📦 **Resource Quotas** — Namespace-scoped resource limits with admission webhook enforcement

**Use Cases:**

| Who            | What                                                |
| -------------- | --------------------------------------------------- |
| Developers     | Spin up dev/test environments for feature testing   |
| SREs           | Automate environment lifecycle with audit trails    |
| Tech Leads     | Monitor team deployments without cluster access     |
| Platform Teams | Enforce policies and governance across environments |

## Features

| Feature                       | Description                                                       |
| ----------------------------- | ----------------------------------------------------------------- |
| 🚀 **Environment Management** | Create, list, get, delete isolated K8s namespaces                 |
| 🎯 **GitOps Integration**     | Automatic ArgoCD Application creation and sync                    |
| 📊 **Live Status**            | Real-time workload status via Kubernetes informers                |
| 🔐 **Authentication**         | JWT + API Key auth with team context                              |
| 👥 **User & Team Management** | Multi-tenant with team-scoped resources                           |
| 🔑 **RBAC**                   | Role-based access control with permissions                        |
| 🔑 **API Keys**               | Service-to-service auth with scoped permissions and rate limiting |
| 📋 **Audit Logging**          | Automatic request logging with filterable, paginated audit trails |
| 💰 **Cost Tracking**          | Real-time K8s cost visibility via OpenCost (open source)          |
| 💸 **Budget Management**      | Spend limits with Slack alerts at configurable thresholds         |
| ⚖️ **Rightsizing**            | CPU/memory optimization recommendations with apply/rollback       |
| 📦 **Resource Quotas**        | Namespace-scoped resource limits with admission webhook enforcement |
| 🛡️ **Policy Enforcement**    | Admission webhook for resource validation                         |

## Tech Stack

| Component     | Technology      |
| ------------- | --------------- |
| Language      | Go 1.25+        |
| Web Framework | Gin             |
| ORM           | GORM            |
| Database      | PostgreSQL 15+  |
| Kubernetes    | client-go       |
| GitOps        | ArgoCD          |
| Cost Analysis | OpenCost        |
| Monitoring    | Prometheus      |
| Notifications | Slack Webhooks  |
| Caching/Lock  | Redis Sentinel  |
| API Docs      | Swagger/OpenAPI |

## Quick Start

```bash
# Install dependencies
make bootstrap

# Start PostgreSQL
make dev-db-up

# Start Redis (for cron server)
make dev-redis-up

# Run API server with hot-reload
make dev-run

# Run cron server with hot-reload
make dev-cron-run

# Run tests
make test
```

## API Endpoints

### Core

| Method | Endpoint        | Description       |
| ------ | --------------- | ----------------- |
| GET    | `/ping`         | Health check      |
| POST   | `/auth/login`   | Authenticate user |
| GET    | `/swagger/*any` | API documentation |

### Environments

| Method | Endpoint                      | Description            |
| ------ | ----------------------------- | ---------------------- |
| GET    | `/v1/environments`            | List environments      |
| POST   | `/v1/environments`            | Create environment     |
| GET    | `/v1/environments/:id`        | Get environment        |
| DELETE | `/v1/environments/:id`        | Delete environment     |
| POST   | `/v1/environments/:id/sync`   | Trigger GitOps sync    |
| GET    | `/v1/environments/:id/status` | Get environment status |

### Users & Teams

| Method | Endpoint                | Description       |
| ------ | ----------------------- | ----------------- |
| GET    | `/v1/users`             | List users        |
| POST   | `/v1/users`             | Create user       |
| GET    | `/v1/users/:id`         | Get user          |
| GET    | `/v1/teams`             | List teams        |
| POST   | `/v1/teams`             | Create team       |
| GET    | `/v1/teams/:id/members` | List team members |

### API Keys

| Method | Endpoint           | Description         |
| ------ | ------------------ | ------------------- |
| GET    | `/v1/api-keys`     | List API keys       |
| POST   | `/v1/api-keys`     | Create API key      |
| GET    | `/v1/api-keys/:id` | Get API key details |
| PATCH  | `/v1/api-keys/:id` | Update API key      |
| DELETE | `/v1/api-keys/:id` | Revoke API key      |

### Audit Logs

| Method | Endpoint             | Description                  |
| ------ | -------------------- | ---------------------------- |
| GET    | `/v1/audit-logs`     | List audit logs (filterable) |
| GET    | `/v1/audit-logs/:id` | Get audit log entry          |

### RBAC

| Method | Endpoint           | Description           |
| ------ | ------------------ | --------------------- |
| GET    | `/v1/roles`        | List roles            |
| POST   | `/v1/roles`        | Create role           |
| POST   | `/v1/roles/assign` | Assign role to user   |
| POST   | `/v1/roles/revoke` | Revoke role from user |

### Cost Tracking

| Method | Endpoint         | Description                                             |
| ------ | ---------------- | ------------------------------------------------------- |
| GET    | `/v1/costs`      | List cost records (filterable by team, namespace, date) |
| GET    | `/v1/costs/team` | Get team cost records by time range                     |

### Budget Management

| Method | Endpoint                   | Description                         |
| ------ | -------------------------- | ----------------------------------- |
| GET    | `/v1/budgets`              | List budgets (team-scoped)          |
| POST   | `/v1/budgets`              | Create budget with alert thresholds |
| GET    | `/v1/budgets/:id`          | Get budget details                  |
| PATCH  | `/v1/budgets/:id`          | Update budget                       |
| DELETE | `/v1/budgets/:id`          | Delete budget                       |
| GET    | `/v1/budgets/:id/alerts`   | Get alert history for budget        |

### Rightsizing

| Method | Endpoint                                      | Description                        |
| ------ | --------------------------------------------- | ---------------------------------- |
| GET    | `/v1/rightsizing/recommendations`             | List recommendations (filterable)  |
| GET    | `/v1/rightsizing/recommendations/:id`         | Get recommendation details         |
| POST   | `/v1/rightsizing/recommendations/:id/apply`   | Apply recommendation to workload   |
| POST   | `/v1/rightsizing/recommendations/:id/rollback`| Rollback to previous resources     |
| POST   | `/v1/rightsizing/recommendations/:id/dismiss` | Dismiss recommendation             |

### Resource Quotas

| Method | Endpoint                                        | Description                         |
| ------ | ----------------------------------------------- | ----------------------------------- |
| GET    | `/v1/quotas`                                    | List resource quotas (filterable)   |
| POST   | `/v1/quotas`                                    | Create resource quota               |
| GET    | `/v1/quotas/:id`                                | Get quota details                   |
| PATCH  | `/v1/quotas/:id`                                | Update quota                        |
| DELETE | `/v1/quotas/:id`                                | Delete quota                        |
| GET    | `/v1/quotas/namespace/:namespace`               | Get quota by namespace              |
| GET    | `/v1/quotas/namespace/:namespace/usage`         | Get namespace resource usage        |
| POST   | `/v1/quotas/namespace/:namespace/usage/refresh` | Refresh cached usage                |
| POST   | `/v1/quotas/check`                              | Check if request exceeds quota      |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Handler Layer                    │
│         (HTTP handlers, validation, auth)           │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                    UseCase Layer                    │
│            (Business logic, orchestration)          │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                  Repository Layer                   │
│       (Data access, K8s, ArgoCD, PostgreSQL)        │
└─────────────────────────────────────────────────────┘
```

**Layer Rules**: Handler → UseCase → Repository. Dependencies flow inward.

## Development

```bash
# Development
make dev-db-up        # Start PostgreSQL
make dev-redis-up     # Start Redis (master + slave + sentinel)
make dev-run          # Run API server with hot-reload
make dev-cron-run     # Run cron server with hot-reload
make dev-db-down      # Stop PostgreSQL

# Docker (no local tools needed)
make dev-app-up       # Start API server in Docker
make dev-cron-up      # Start cron server in Docker
make dev-app-logs     # View API server logs
make dev-cron-logs    # View cron server logs

# Testing
make test             # Run all tests
make test-coverage    # Coverage report
make lint             # Run linter

# K8s Integration
make dev-k8s-setup    # Setup Kind + ArgoCD
make test-k8s         # Run K8s tests

# FinOps (Cost Tracking)
make dev-finops-setup # Setup Prometheus + OpenCost
make dev-finops-status # Check FinOps components

# Docker Builds
make docker-build      # Build API server image
make docker-build-cron # Build cron server image

# Code Generation
make swagger-gen      # Generate API docs
```

## Configuration

Configuration via `configs/config.<env>.yaml` (set `APP_ENV` to select) or environment variables:

```bash
APP_ENV=development
SERVER_PORT=8989
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=idp_core
JWT_SECRET=your-secret
FINOPS_ENABLED=true
FINOPS_OPENCOST_BASE_URL=http://opencost.opencost.svc.cluster.local:9003
FINOPS_OPENCOST_POLL_INTERVAL=1h
FINOPS_PROMETHEUS_URL=http://prometheus-server.monitoring.svc.cluster.local:80
CRON_PORT=8983
REDIS_MASTER_NAME=idp-core-redis_sentinel
REDIS_ADDRESS=localhost:26379
REDIS_PASSWORD=redispassword
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx
SLACK_CHANNEL=#budget-alerts
```

## Project Structure

```
idp-core/
├── cmd/
│   ├── http/           # API server entry point
│   └── cron/           # Cron job server entry point
├── internal/
│   ├── handler/
│   │   ├── http/       # HTTP handlers + middleware
│   │   └── cron/       # Cron job handlers
│   ├── usecase/        # Business logic
│   ├── repository/     # Data access
│   ├── model/          # Domain models
│   ├── mocks/          # Test mock objects
│   └── pkg/            # Internal packages (opencost, prometheus, redislock, oidc, etc.)
├── deployments/        # Kubernetes manifests
│   └── kubernetes/
├── docs/               # Documentation
├── configs/            # Configuration
└── migrations/         # Database migrations
```

## Roadmap

| Phase                 | Timeline | Status         | Focus                     |
| --------------------- | -------- | -------------- | ------------------------- |
| Phase 1 - MVP         | Q2 2026  | ✅ Complete     | Core API, K8s/ArgoCD      |
| Phase 2 - Enhancement | Q3 2026  | 🔄 In Progress | RBAC, FinOps, Rightsizing |
| Phase 3 - Platform    | Q4 2026  | 📋 Planned     | UI, Templates             |
| Phase 4 - Advanced    | Q1 2027+ | 🔮 Roadmap     | AI/ML, Analytics          |

## Documentation

- [PRD Overview](docs/prd/PRD.md) — Product requirements
- [PRD Phase 2](docs/prd/PRD_PHASE_2.md) — Enhancement requirements
- [Dev TODO List](docs/DEV_TODO_LIST_PHASE_2.md) — Current progress
- [Development Guidelines](docs/DEV_GUIDELINE.md) — Coding standards

## License

MIT
