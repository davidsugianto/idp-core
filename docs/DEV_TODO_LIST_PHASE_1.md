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
| 6. Testing & Validation | ~ Complete | 80% |
| 7. Deployment & Operations | ✅ Complete | 100% |

**Remaining Items:**
- Rate limiting middleware (API Layer)
- E2E & Contract Tests (Testing)
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
  - [x] Implement ArgoCD Application creation logic.
  - [x] Implement ArgoCD sync trigger.
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
  - [x] Implement `X-API-Key` validation middleware using `APIKeyRepository`.
  - [ ] Implement basic rate limiting (100 req/min per API key).
- [x] **API Handlers Implementation**
  - [x] `GET /health` and `GET /ready` (Liveness/Readiness probes).
  - [x] `POST /environments`
  - [x] `GET /environments`
  - [x] `GET /environments/:id`
  - [x] `DELETE /environments/:id`
  - [x] `POST /environments/:id/sync`
  - [x] `GET /environments/:id/gitops/status`
  - [x] `GET /environments/:id/workloads`
  - [x] `GET /environments/:id/workloads/:name`
- [x] **API Documentation**
  - [x] Generate OpenAPI spec (e.g., using `swaggo/swag`).
  - [x] Serve Swagger UI on `/swagger/*any`.

## 🧪 6. Testing & Validation (Milestone 4)

- [~] **Unit Tests (>80% coverage)**
  - [x] Write tests for Handlers using `testify` and `gomock`.
  - [x] Write tests for Use Cases using `testify` and `gomock`.
  - [x] Write tests for Utils and Validators.
  - [~] Achieve >80% coverage (currently 58.5% total, core packages 73-100%).
- [x] **Integration Tests**
  - [x] Setup `testcontainers-go` for PostgreSQL integration testing.
  - [x] Test DB repository layer against real PostgreSQL.
  - [x] Test K8s client against minikube/kind.
  - [x] Test ArgoCD client against cluster with ArgoCD installed.
- [ ] **E2E & Contract Tests**
  - [ ] Implement full flow E2E tests (API → K8s → ArgoCD → status).
  - [ ] Set up OpenAPI spec validation tests.

## 📦 7. Deployment & Operations (Milestone 5) ✅

- [x] **Containerization**
  - [x] Write optimized `Dockerfile` (multi-stage build, distroless or alpine base).
- [x] **Kubernetes Manifests**
  - [x] Create Deployment, Service, ConfigMap, and ServiceAccount manifests in `deployments/kubernetes/`.
  - [x] Define proper resource requests/limits, liveness, and readiness probes.
- [x] **CI/CD Setup**
  - [x] Create GitHub Actions workflows for tests, linting, and building/pushing the Docker image.
- [x] **Documentation & Runbooks**
  - [x] Write `README.md` with setup instructions and API examples.
  - [x] Document required Kubernetes RBAC permissions for the service account.
  - [x] Create initial on-call runbook for common alerts.
