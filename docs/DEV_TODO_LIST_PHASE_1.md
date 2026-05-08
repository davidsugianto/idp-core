# 📋 idp-core — Phase 1 MVP Development To-Do List

This development to-do list is derived directly from the [Phase 1 MVP PRD](./prd/PRD_PHASE_1.md) to track implementation progress.

## Progress Summary

| Milestone | Status | Completion |
|-----------|--------|------------|
| 1. Project Setup & Foundation | ✅ Complete | 100% |
| 2. Data Models & Repository Layer | ✅ Complete | 100% |
| 3. External Integrations | ✅ Complete | 100% |
| 4. Core Business Logic | ✅ Complete | 100% |
| 5. API Layer & Routing | ~ Complete | 95% |
| 6. Testing & Validation | ✅ Complete | 100% |
| 7. Deployment & Operations | ✅ Complete | 100% |

**Remaining Items:**
- Rate limiting middleware (API Layer) - Optional for Phase 1 MVP
- Increase unit test coverage to >80%

---

## 🏗️ 1. Project Setup & Foundation (Milestone 1) ✅

- [x] **Repository Initialization**
  - [x] Initialize Go module (`go mod init`).
  - [x] Set up standard Go project layout (e.g., `cmd/`, `internal/`, `pkg/`, `api/`).
  - [x] Configure `golangci-lint` and pre-commit hooks.
  - [x] Set up Makefile with targets (`bootstrap`, `dev-up`, `dev-run`, `test`, `build`, `docker-build`).
- [x] **Configuration Management**
  - [x] Implement configuration loading (e.g., using Viper) for reading from environment variables and `config.yaml`.
  - [x] Define configuration structures for Server, Database, Kubernetes, and ArgoCD.
  - [x] Support environment variable overrides for all configuration.
- [x] **Logging & Observability Basics**
  - [x] Set up structured JSON logging (e.g., using Logrus or Zap).
  - [x] Implement Prometheus metrics registry and `/metrics` endpoint.
- [x] **Database Setup**
  - [x] Set up PostgreSQL connection using GORM.
  - [x] Implement auto-migration or schema versioning for Phase 1 data models.

## 🗃️ 2. Data Models & Repository Layer ✅

- [x] **Define GORM Entities**
  - [x] `Environment` struct with fields: ID, Name, Team, Template, Namespace, GitRepoURL, Status, ArgoCDApp, Config, timestamps.
  - [x] `WorkloadStatus` struct (cached view).
  - [x] `APIKey` struct (for simple service auth).
- [x] **Implement Repositories**
  - [x] `EnvironmentRepository` interface and PostgreSQL implementation (Create, Get, List, Delete, UpdateStatus).
  - [x] `APIKeyRepository` interface and PostgreSQL implementation.

## 🔌 3. External Integrations (Milestone 2) ✅

- [x] **Kubernetes Client (`client-go`)**
  - [x] Initialize dynamic and typed clients (in-cluster config + kubeconfig fallback).
  - [x] Implement namespace provisioning (create Namespace, ResourceQuota, NetworkPolicy).
  - [x] Implement workload status querying (Deployments, StatefulSets, Pods).
  - [x] Implement informer-based status caching.
- [x] **ArgoCD Client**
  - [x] Implement REST API client (v2.11+) or Git commit webhook fallback.
  - [x] Implement ArgoCD Application creation logic with project support.
  - [x] Implement ArgoCD sync trigger with proper resource versioning.
  - [x] Implement ArgoCD status polling.

## 🧠 4. Core Business Logic (Use Cases) ✅

- [x] **Environment Management Use Cases**
  - [x] `CreateEnvironment`: Validate input, create DB record, provision K8s namespace, setup ArgoCD app, update DB status.
  - [x] `GetEnvironment` & `ListEnvironments`: Fetch from DB.
  - [x] `DeleteEnvironment`: Clean up K8s resources, ArgoCD app, mark DB record as deleted.
- [x] **GitOps & Workload Use Cases**
  - [x] `TriggerSync`: Call ArgoCD client to sync application.
  - [x] `GetGitOpsStatus`: Fetch sync and health status from ArgoCD.
  - [x] `GetWorkloads`: Fetch live workload data from Kubernetes for a specific environment.
  - [x] `GetWorkloadDetails`: Fetch specific workload and its pods' status.

## 🌐 5. API Layer & Routing (Milestone 3)

- [x] **Gin Router Setup**
  - [x] Initialize Gin router with standard middlewares (recovery, logging).
  - [x] Implement custom logging middleware to redact sensitive headers and log request duration/trace_id.
- [~] **Authentication & Security Middleware**
  - [x] Implement JWT authentication middleware.
  - [x] Implement `X-API-Key` validation middleware using `APIKeyRepository`.
  - [x] Add `GenerateToken` and `ValidateToken` helper functions.
  - [ ] Implement basic rate limiting (100 req/min per API key) - Optional for MVP.
- [x] **API Handlers Implementation**
  - [x] `GET /health` and `GET /ready` (Liveness/Readiness probes).
  - [x] `POST /auth/login` (JWT token generation).
  - [x] `POST /environments`
  - [x] `GET /environments`
  - [x] `GET /environments/:id`
  - [x] `DELETE /environments/:id`
  - [x] `POST /environments/:id/sync`
  - [x] `GET /environments/:id/gitops/status`
  - [x] `GET /environments/:id/workloads`
  - [x] `GET /environments/:id/workloads/:name`
- [x] **API Documentation**
  - [x] Generate OpenAPI spec (using `swaggo/swag` v1.16.6).
  - [x] Add API title, version, description to swagger annotations.
  - [x] Add security definitions to OpenAPI spec.
  - [x] Serve Swagger UI on `/swagger/*any`.

## 🧪 6. Testing & Validation (Milestone 4) ✅

- [~] **Unit Tests (>80% coverage)**
  - [x] Write tests for Handlers using `testify` and `gomock`.
  - [x] Write tests for Use Cases using `testify` and `gomock`.
  - [x] Write tests for Utils and Validators.
  - [x] Write tests for Middleware (JWT, Logger, RequestID).
  - [~] Achieve >80% coverage (core packages 73-100%, overall ~60%).
- [x] **Integration Tests**
  - [x] Setup `testcontainers-go` for PostgreSQL integration testing.
  - [x] Add wait strategies for container readiness.
  - [x] Test DB repository layer against real PostgreSQL.
  - [x] Test K8s client against kind cluster.
  - [x] Test ArgoCD client against cluster with ArgoCD installed.
- [x] **E2E & Contract Tests**
  - [x] Implement E2E test suite in `tests/e2e/`.
  - [x] Implement OpenAPI contract tests in `tests/contract/`.
  - [x] Add Makefile targets: `test-e2e`, `test-contract`.
  - [x] Create comprehensive test documentation (`docs/TEST.md`).

## 📦 7. Deployment & Operations (Milestone 5) ✅

- [x] **Containerization**
  - [x] Write optimized `Dockerfile` (multi-stage build with alpine base).
  - [x] Use Go 1.25 with air-verse/air for hot-reload.
  - [x] Add health check in production stage.
  - [x] Run as non-root user in production.
- [x] **Kubernetes Manifests**
  - [x] Create Namespace, Deployment, Service, ConfigMap, Secret, ServiceAccount manifests.
  - [x] Create RBAC manifests (ClusterRole, ClusterRoleBinding).
  - [x] Define proper resource requests/limits, liveness, and readiness probes.
  - [x] Use named ports for flexibility.
- [x] **CI/CD Setup**
  - [x] Create GitHub Actions workflow for CI (`ci.yaml`).
  - [x] Create GitHub Actions workflow for build and release (`build.yaml`).
  - [x] Configure golangci-lint v1.65.0 with Go 1.25 support.
  - [x] Add dependency review with continue-on-error for flexibility.
- [x] **Documentation & Runbooks**
  - [x] Write `README.md` with setup instructions and API examples.
  - [x] Write `docs/DEV_GUIDELINE.md` for development guidelines.
  - [x] Write `docs/TEST.md` for test documentation.
  - [x] Write `docs/RBAC.md` for RBAC permissions.
  - [x] Write `docs/RUNBOOK.md` for on-call runbook.
  - [x] Write `docs/prd/PRD.md` for PRD overview.
  - [x] Write `docs/prd/PRD_PHASE_1.md` for Phase 1 requirements.

---

## 📁 Project Structure

```
idp-core/
├── cmd/
│   └── http/                    # Application entry point
├── internal/
│   ├── handler/http/            # HTTP handlers
│   │   └── middleware/          # HTTP middleware
│   ├── usecase/                 # Business logic layer
│   ├── repository/              # Data access layer
│   ├── model/                   # Domain models
│   ├── pkg/                     # Internal packages
│   │   ├── config/              # Configuration
│   │   ├── kubernetes/          # Kubernetes client
│   │   ├── argocd/              # ArgoCD client
│   │   └── ...
│   └── mocks/                   # Generated mocks
├── tests/
│   ├── e2e/                     # End-to-end tests
│   └── contract/                # Contract tests
├── docs/
│   ├── prd/                     # Product requirements
│   │   ├── PRD.md               # PRD overview
│   │   └── PRD_PHASE_1.md       # Phase 1 requirements
│   ├── swagger/                 # Generated OpenAPI docs
│   ├── DEV_GUIDELINE.md         # Development guidelines
│   ├── DEV_TODO_LIST_PHASE_1.md # This file
│   ├── TEST.md                  # Test documentation
│   ├── RBAC.md                  # RBAC documentation
│   └── RUNBOOK.md               # On-call runbook
├── deployments/
│   └── kubernetes/              # Kubernetes manifests
├── dev/                         # Development scripts
├── configs/                     # Configuration files
├── Makefile                     # Build and dev commands
├── Dockerfile                   # Multi-stage Dockerfile
├── docker-compose.yml           # Local development
└── .github/workflows/           # CI/CD workflows
```

---

## 🔧 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8989` | API server port |
| `DB_HOST` | `postgres` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_NAME` | `idp_core` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `JWT_SECRET` | - | JWT signing secret |
| `K8S_IN_CLUSTER` | `false` | Use in-cluster config |
| `KUBECONFIG_PATH` | `~/.kube/config` | Kubeconfig path |
| `ARGOCD_BASE_URL` | - | ArgoCD server URL |
| `ARGOCD_TOKEN` | - | ArgoCD API token |

---

## 📊 Test Coverage

| Package | Coverage |
|---------|----------|
| Handler | 84.7% |
| Middleware | 92.9% |
| Usecase | 73.6% |
| Repository | 0-0.9% (integration tests) |
| Errors | 100% |
| Validator | 100% |
| Webhook | 100% |

---

## 🚀 Next Steps (Post Phase 1)

1. **Phase 2 - Enhancement**
   - RBAC & OIDC authentication
   - FinOps & cost analysis
   - Resource rightsizing
   - Service catalog

2. **Phase 3 - Platform**
   - Developer portal UI
   - Template management
   - Multi-cluster support

3. **Phase 4 - Advanced**
   - AI/ML workflows
   - Advanced analytics
   - Policy engine
