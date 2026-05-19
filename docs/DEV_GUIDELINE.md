# Development Guidelines

This document provides guidelines and best practices for developing idp-core.

## Code Architecture

### Clean Architecture Layers

The project follows strict clean architecture with dependency inversion:

```
cmd/                        # Entry points
internal/
├── handler/http/           # HTTP handlers (controllers) - outer layer
├── usecase/                # Business logic - middle layer
├── repository/             # Data access - inner layer
└── model/                  # Domain models - core
```

**Key Rules:**

- Handlers → call Usecase only
- Usecase → calls Repository only
- Repository → accesses DB/external APIs
- Dependencies flow **inward**; inner layers never import outer layers

### Dependency Injection Pattern

Each layer uses a `Dependencies` struct for constructor injection:

```go
// Example: usecase layer
type Dependencies struct {
    EnvironmentRepo  environmentRepo.Repository
    ProvisionerRepo  provisionerRepo.Repository
    GitopsRepo       gitopsRepo.Repository
}

func New(deps Dependencies) Usecase {
    return &usecase{
        environmentRepo:  deps.EnvironmentRepo,
        provisionerRepo:  deps.ProvisionerRepo,
        gitopsRepo:       deps.GitopsRepo,
    }
}
```

### Adding a New Feature

When adding a new feature (e.g., "project"):

1. **Model**: Create `internal/model/project/type.go`
2. **Repository**: Create `internal/repository/project/` with `init.go` (interface) and `project.go` (implementation)
3. **Usecase**: Create `internal/usecase/project/` with same pattern
4. **Handler**: Create `internal/handler/http/project.go`
5. **Wire**: Connect dependencies in `cmd/http/server.go`

## Error Handling

### Sentinel Errors

Define sentinel errors in the usecase layer:

```go
// In usecase/environment/environment.go
var (
    ErrEnvironmentNotFound = errors.New("environment not found")
    ErrNoArgoApp           = errors.New("environment has no ArgoCD application")
    ErrGitOpsNotConfigured = errors.New("GitOps integration not configured")
    ErrK8sNotConfigured    = errors.New("Kubernetes integration not configured")
    ErrWorkloadNotFound    = errors.New("workload not found")
)
```

### Error Wrapping

Always wrap errors with context:

```go
if err != nil {
    return fmt.Errorf("failed to create environment: %w", err)
}
```

### HTTP Error Responses

In handlers, use `errors.Is()` to check sentinel errors:

```go
if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
    c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
    return
}
```

## Development Environment

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (for PostgreSQL and/or app)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (for K8s integration tests)
- [kind](https://kind.sigs.k8s.io/) (installed automatically via make)
- Go 1.25+

### Running the Application

#### Option 1: Run in Docker (Recommended)

No additional tools required - everything runs in Docker:

```bash
# Start PostgreSQL and application together
make dev-all-up

# Or start separately
make dev-db-up      # Start PostgreSQL
make dev-app-up     # Start application

# View logs
make dev-app-logs

# Stop everything
make dev-all-down
```

#### Option 2: Run Locally with Hot-Reload

Requires Air to be installed:

```bash
# Install Air (one-time) - use air-verse fork for Go 1.25 support
go install github.com/air-verse/air@latest

# Start PostgreSQL
make dev-db-up

# Run with hot-reload
make dev-run

# Stop PostgreSQL when done
make dev-db-down
```

### Development Modes

#### 1. Local Development (PostgreSQL only)

For running the application locally with just a database:

```bash
# Start PostgreSQL in Docker
make dev-db-up

# Run the application (choose one)
make dev-app-up     # In Docker
make dev-run        # Locally with Air

# Run PostgreSQL integration tests
make test-db

# Stop PostgreSQL when done
make dev-db-down
```

#### 2. Kubernetes Integration Testing (Kind + ArgoCD)

For testing Kubernetes and ArgoCD integration:

```bash
# Setup Kind cluster with ArgoCD
make dev-k8s-setup

# Or quick setup with minimal ArgoCD
make dev-k8s-setup-quick

# Run Kubernetes/ArgoCD integration tests
make test-k8s
make test-argocd

# Check environment status
make dev-k8s-status

# Teardown when done
make dev-k8s-teardown
```

#### 3. Cron Server Development (Redis + PostgreSQL)

For running and testing the cron job server:

```bash
# Start dependencies
make dev-db-up         # PostgreSQL
make dev-redis-up      # Redis (master + slave + sentinel)

# Run the cron server (choose one)
make dev-cron-run      # Locally with Air
make dev-cron-up       # In Docker

# View logs
make dev-cron-logs

# Stop when done
make dev-cron-down
make dev-redis-down
```

#### 4. Full Setup (Both)

```bash
# Setup everything (PostgreSQL + Kind + ArgoCD)
make dev-setup

# Run all tests
make test-all-integration

# Teardown everything
make dev-teardown
```

### Quick Reference

| Command                    | Description                    | Requirements    |
| -------------------------- | ------------------------------ | --------------- |
| `make dev-db-up`           | Start PostgreSQL in Docker     | Docker          |
| `make dev-db-down`         | Stop PostgreSQL                | -               |
| `make dev-app-up`          | Start app in Docker            | Docker          |
| `make dev-app-down`        | Stop app container             | -               |
| `make dev-app-logs`        | View app logs                  | -               |
| `make dev-all-up`          | Start PostgreSQL + App         | Docker          |
| `make dev-all-down`        | Stop all services              | -               |
| `make dev-run`             | Run API server with Air        | Air installed   |
| `make dev-cron-run`        | Run cron server with Air       | Air installed   |
| `make dev-redis-up`        | Start Redis in Docker          | Docker          |
| `make dev-redis-down`      | Stop Redis containers          | -               |
| `make dev-cron-up`         | Start cron server in Docker    | Docker          |
| `make dev-cron-down`       | Stop cron server container     | -               |
| `make dev-cron-logs`       | View cron server logs          | -               |
| `make dev-k8s-setup`       | Full K8s setup (Kind + ArgoCD) | kubectl, kind   |
| `make dev-k8s-setup-quick` | Minimal K8s setup              | kubectl, kind   |
| `make dev-k8s-teardown`    | Delete Kind cluster            | -               |
| `make dev-setup`           | Full setup (DB + K8s)          | Docker, kubectl |
| `make dev-teardown`        | Teardown everything            | -               |

### Dev Directory Structure

```
dev/
├── kind-config.yaml           # Kind cluster configuration
├── setup-kind.sh              # Full K8s setup script
├── setup-argocd-minimal.sh    # Minimal ArgoCD setup (faster)
├── setup-prometheus.sh        # Prometheus setup in Kind
├── setup-opencost.sh          # OpenCost setup in Kind
├── teardown-kind.sh           # K8s teardown script
└── README.md                  # Dev environment docs
```

### Makefile Variables for K8s Setup

```bash
# Customize cluster name
CLUSTER_NAME=my-cluster make dev-k8s-setup

# Customize ArgoCD version
ARGOCD_VERSION=v2.11.0 make dev-k8s-setup

# Increase timeout for slow machines
TIMEOUT=900 make dev-k8s-setup
```

## Testing

For detailed test documentation, see [TEST.md](./TEST.md).

### Test Structure

```
internal/
├── usecase/environment/
│   └── environment_test.go              # Unit tests
├── repository/environment/
│   └── environment_integration_test.go  # PostgreSQL tests
├── repository/provisioner/
│   └── kubernetes_integration_test.go   # Kubernetes tests
├── repository/gitops/
│   └── argocd_integration_test.go       # ArgoCD tests
└── mocks/
    ├── environment_repository.go
    ├── provisioner_repository.go
    └── gitops_repository.go
tests/
├── e2e/
│   └── e2e_test.go                      # End-to-end tests
└── contract/
    └── openapi_test.go                  # OpenAPI contract tests
```

### Test Categories

| Category        | Command                     | Requirements         |
| --------------- | --------------------------- | -------------------- |
| Unit Tests      | `make test-unit`            | None                 |
| PostgreSQL      | `make test-db`              | `make dev-db-up`     |
| Kubernetes      | `make test-k8s`             | `make dev-k8s-setup` |
| ArgoCD          | `make test-argocd`          | `make dev-k8s-setup` |
| E2E Tests       | `make test-e2e`             | `make dev-k8s-setup` |
| Contract Tests  | `make test-contract`        | `make swagger-gen`   |
| All Integration | `make test-all-integration` | `make dev-setup`     |

### Unit Tests

Use `testify` and `gomock`:

```go
func TestCreate(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
    mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)

    uc := New(Dependencies{
        EnvironmentRepo: mockEnvRepo,
        ProvisionerRepo: mockProvRepo,
    })

    // Setup expectations
    mockEnvRepo.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        Return(nil)

    // Test
    result, err := uc.Create(context.Background(), "team-123", req)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests

#### PostgreSQL (testcontainers-go)

```go
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
    ctx := context.Background()

    pgContainer, err := postgres.Run(ctx, "postgres:15-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    require.NoError(t, err)

    // ... setup and return cleanup function
}
```

#### Kubernetes (kind)

```go
func skipIfNoK8s(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client, err := kubernetes.NewClient(false, "")
    if err != nil {
        t.Skipf("Failed to connect to Kubernetes: %v", err)
    }

    ctx := context.Background()
    _, err = client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
    if err != nil {
        t.Skipf("Kubernetes cluster not accessible: %v", err)
    }
}
```

#### ArgoCD

```go
func skipIfNoArgoCD(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Check ArgoCD namespace exists
    client, err := argocd.NewClient(false, "")
    if err != nil {
        t.Skipf("Failed to create client: %v", err)
    }

    // Check ArgoCD CRD is installed
    _, err = client.DynamicClient.Resource(schema.GroupVersionResource{
        Group:    "argoproj.io",
        Version:  "v1alpha1",
        Resource: "applications",
    }).Namespace("argocd").List(ctx, metav1.ListOptions{Limit: 1})

    if err != nil {
        t.Skip("ArgoCD not installed. Run 'make dev-setup'")
    }
}
```

### Running Tests

```bash
# Unit tests only (fast, no external deps)
make test-unit

# PostgreSQL integration tests (requires Docker)
make test-integration

# Kubernetes integration tests (requires kind)
make test-integration-k8s

# ArgoCD integration tests (requires ArgoCD)
make test-integration-argocd

# All integration tests
make test-all-integration

# Full workflow
make dev-setup && make test-all-integration

# With coverage
make test-coverage
```

## Naming Conventions

### Files

| Layer             | File Pattern                 | Example                                   |
| ----------------- | ---------------------------- | ----------------------------------------- |
| Model             | `type.go`                    | `internal/model/environment/type.go`      |
| Repository        | `init.go`, `<name>.go`       | `internal/repository/environment/init.go` |
| Usecase           | `init.go`, `<name>.go`       | `internal/usecase/environment/create.go`  |
| Handler           | `<name>.go`                  | `internal/handler/http/environment.go`    |
| Tests             | `<name>_test.go`             | `environment_test.go`                     |
| Integration Tests | `<name>_integration_test.go` | `kubernetes_integration_test.go`          |

### Interfaces

Define interfaces in `init.go`:

```go
// init.go
type Repository interface {
    Create(ctx context.Context, env *environment.Environment) error
    GetByID(ctx context.Context, id string) (*environment.Environment, error)
    // ...
}
```

### Structs

- Repository struct: `type repository struct { ... }` (unexported)
- Usecase struct: `type usecase struct { ... }` (unexported)
- Handler struct: `type Handler struct { ... }` (exported)

## Kubernetes Integration

### Namespace Naming

Namespaces follow DNS-1123 compliant naming:

```
idp-{teamSlug}-{envSlug}
```

- Max 63 characters
- Lowercase alphanumeric and hyphens only
- Auto-truncated if too long

### Labels

All resources created by idp-core have:

```yaml
labels:
  idp-core/managed-by: idp-core
  idp-core/team-id: <team-id>
  idp-core/environment-id: <env-id>
```

### Resource Isolation

Each environment gets:

1. **Namespace** - Kubernetes namespace
2. **ResourceQuota** - CPU/Memory limits (optional)
3. **NetworkPolicy** - Network isolation

## GitOps / ArgoCD

### Application Naming

```
env-{environmentIdPrefix}
```

Example: `env-a1b2c3d4`

### Application Spec

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: env-a1b2c3d4
  namespace: argocd
  labels:
    idp-core/managed-by: idp-core
spec:
  project: default              # ArgoCD project (defaults to "default")
  destination:
    namespace: idp-team-dev
    server: https://kubernetes.default.svc
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: main
    path: manifests
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

## Configuration

### Environment Variables

All configuration can be overridden via environment variables:

```bash
# Server
SERVER_PORT=8989

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=idp_core
DB_USER=postgres
DB_PASSWORD=secret

# Auth
JWT_SECRET=your-secret-key

# Kubernetes
K8S_IN_CLUSTER=false
KUBECONFIG_PATH=~/.kube/config

# ArgoCD
ARGOCD_BASE_URL=https://argocd.example.com
ARGOCD_TOKEN=your-token
```

For local development outside Docker:

```bash
export DB_HOST=localhost
export DB_NAME=idp_core
make dev-run
```

### Config File

Default configuration in `configs/config.yaml`:

```yaml
server:
  port: 8989

database:
  host: postgres      # Docker service name (use localhost for local dev)
  port: 5432
  name: idp_core
  user: postgres
  password: postgres

kubernetes:
  in_cluster: false
  kubeconfig_path: "~/.kube/config"

argocd:
  base_url: "https://argocd.example.com"
  token: ""
```

**Note:** Config values can be overridden via environment variables. See [Environment Variables](#environment-variables) section.

## API Design

### RESTful Endpoints

```
POST   /v1/environments              # Create
GET    /v1/environments              # List
GET    /v1/environments/:id          # Get
DELETE /v1/environments/:id          # Delete
POST   /v1/environments/:id/sync     # Trigger sync
GET    /v1/environments/:id/status   # Get status
GET    /v1/environments/:id/gitops/status  # ArgoCD status
GET    /v1/environments/:id/workloads      # List workloads
GET    /v1/environments/:id/workloads/:name # Get workload details
```

### Response Format

Success:

```json
{
  "id": "env-123",
  "name": "dev",
  "status": "ready",
  ...
}
```

Error:

```json
{
  "error": "environment not found"
}
```

### Swagger Documentation

Add annotations to handlers:

```go
// CreateEnvironment godoc
// @Summary Create a new environment
// @Description Create a new Kubernetes environment with ArgoCD integration
// @Tags environments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body environment.CreateEnvironmentRequest true "Environment request"
// @Success 201 {object} environment.EnvironmentResponse
// @Failure 400 {object} map[string]string
// @Router /v1/environments [post]
func (h *Handler) CreateEnvironment(c *gin.Context) { ... }
```

Generate docs:

```bash
make swagger-gen
```

## Common Commands

```bash
# Run Application
make dev-app-up            # Run API server in Docker (recommended)
make dev-app-down          # Stop API server container
make dev-app-logs          # View API server logs
make dev-run               # Run API server locally with Air

# Cron Server
make dev-cron-up           # Run cron server in Docker
make dev-cron-down         # Stop cron server container
make dev-cron-logs         # View cron server logs
make dev-cron-run          # Run cron server locally with Air

# Local Development (PostgreSQL only)
make dev-db-up             # Start PostgreSQL in Docker
make dev-db-down           # Stop PostgreSQL
make dev-db-reset          # Reset PostgreSQL database
make dev-db-logs           # Show PostgreSQL logs

# Redis
make dev-redis-up          # Start Redis in Docker
make dev-redis-down        # Stop Redis containers
make dev-redis-logs        # View Redis logs

# Run Everything (PostgreSQL + App)
make dev-all-up            # Start PostgreSQL + App in Docker
make dev-all-down          # Stop all services

# Kubernetes Development (Kind + ArgoCD)
make dev-k8s-setup         # Setup Kind + ArgoCD (full)
make dev-k8s-setup-quick   # Setup Kind + minimal ArgoCD (fast)
make dev-k8s-status        # Check K8s environment status
make dev-k8s-teardown      # Delete Kind cluster
make dev-k8s-argocd-ui     # Port-forward ArgoCD UI

# Full Setup (Both)
make dev-setup             # Setup PostgreSQL + Kind + ArgoCD
make dev-setup-quick       # Quick full setup
make dev-teardown          # Teardown everything
make dev-status            # Check all environments

# Testing
make test                  # All tests
make test-unit             # Unit tests only (fast, no deps)
make test-db               # PostgreSQL tests (requires: dev-db-up)
make test-k8s              # Kubernetes tests (requires: dev-k8s-setup)
make test-argocd           # ArgoCD tests (requires: dev-k8s-setup)
make test-e2e              # E2E tests (requires: dev-k8s-setup)
make test-contract         # OpenAPI contract tests
make test-all-integration  # All integration tests
make test-coverage         # Generate coverage report

# Code Quality
make lint                  # Run golangci-lint
make fmt                   # Format code
make vet                   # Run go vet

# Build
make build                 # Build binary
make docker-build          # Build API server Docker image
make docker-build-cron     # Build cron server Docker image

# Database
make db-migrate            # Run migrations
make db-rollback           # Rollback migration

# Swagger
make swagger-gen           # Generate Swagger docs
```

## Troubleshooting

### ArgoCD Setup Timeout

If `make dev-setup` times out waiting for ArgoCD:

```bash
# Option 1: Use quick setup (minimal ArgoCD)
make dev-teardown
make dev-setup-quick

# Option 2: Increase timeout
TIMEOUT=900 make dev-setup

# Option 3: Manual minimal setup
kind create cluster --name idp-test --config dev/kind-config.yaml
kubectl config use-context kind-idp-test
./dev/setup-argocd-minimal.sh

# Check ArgoCD status
kubectl get pods -n argocd
kubectl logs -n argocd deployment/argocd-server
```

### Tests Fail to Connect to K8s

```bash
# Check current context
kubectl config current-context

# Should show: kind-idp-test
# If not:
kubectl config use-context kind-idp-test

# Verify cluster is accessible
kubectl get nodes
```

### ArgoCD Tests Skip

```bash
# Check ArgoCD is installed
kubectl get ns argocd
kubectl get pods -n argocd

# If missing, run setup
make dev-k8s-setup

# Or quick setup
make dev-k8s-setup-quick
```

### Cluster Already Exists

```bash
# Check existing clusters
kind get clusters

# Delete and recreate
make dev-k8s-teardown
make dev-k8s-setup

# Or just reinstall ArgoCD
./dev/setup-argocd-minimal.sh
```

### Database Migration Issues

```bash
# Check migration status
make db-migrate

# Rollback and re-run
make db-rollback
make db-migrate
```

### Port Conflicts

```bash
# Check what's using port 8989
lsof -i :8989

# Kill process
kill -9 <PID>

# Or use different port
SERVER_PORT=8990 make dev-run
```

### Docker Issues

```bash
# Ensure Docker is running
docker ps

# Restart Docker if needed
# (varies by OS)

# Check Docker resources
docker info | grep -i memory
```

### Test Failures

**PostgreSQL integration tests fail with "unexpected EOF":**

```bash
# Tests use testcontainers which need Docker
docker ps

# Pull the image manually if slow
docker pull postgres:15-alpine

# Re-run tests
make test-db
```

**Kubernetes tests fail with "Namespace should not exist":**

Namespace deletion is asynchronous. Tests wait up to 15 seconds for deletion.

**ArgoCD tests fail with "spec.project: Required value":**

The Application CRD requires a project. The code defaults to "default" project.

**ArgoCD tests fail with "resourceVersion: must be specified":**

This is handled by fetching the resource before update. Ensure you have the latest code.

### Check Everything

```bash
# Run status check
make dev-status

# Expected output:
# - PostgreSQL container: Running
# - Kind cluster: idp-test
# - kubectl context: kind-idp-test
# - ArgoCD pods: Running
```

## Accessing ArgoCD UI (Optional)

```bash
# Port-forward ArgoCD server (uses port 8090)
make dev-k8s-argocd-ui

# Or manually:
kubectl port-forward svc/argocd-server -n argocd 8090:443

# Get initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d

# Open https://localhost:8090
# Username: admin
```

## CI/CD Notes

Integration tests are designed to work in CI environments:

```bash
# In CI pipeline - Kubernetes tests
make dev-k8s-setup-quick      # Fast K8s setup
make test-all-integration     # Run tests
make dev-k8s-teardown         # Cleanup

# Or just PostgreSQL tests
make dev-db-up                # Start PostgreSQL
make test-db                  # Run PostgreSQL tests
make dev-db-down              # Cleanup
```

Tests automatically skip if dependencies are not available:

- Tests skip with `-short` flag
- Tests skip if kubeconfig not found
- Tests skip if ArgoCD not installed

## Phase 2 Patterns

### Service Catalog

The service catalog follows the standard clean architecture pattern with extended domain relationships:

```
Service → ServiceVersion → ServiceEndpoint
Service → ServiceDependency → Service (depends on)
ServiceVersion → ServiceEnvironment (deployments)
```

**Key Model Relationships:**

```go
// Service has many versions
type Service struct {
    ID          string
    Name        string
    TeamID      string
    Visibility  string  // public, team, private
    Status      string  // active, inactive
}

// Version belongs to service
type ServiceVersion struct {
    ID        string
    ServiceID string
    Version   string  // semver
    GitRef    string  // git commit/branch/tag
    Status    string
}

// Dependency between services
type ServiceDependency struct {
    ServiceID          string  // source service
    DependsOnServiceID string  // target service
    DependencyType     string  // runtime, build, data, api
}
```

### Dependency Management

**Circular Dependency Detection:**

Use BFS to detect cycles before adding dependencies:

```go
func (u *usecase) checkCircularDependency(ctx context.Context, serviceID, dependsOnServiceID string) error {
    visited := make(map[string]bool)
    queue := []string{dependsOnServiceID}

    for len(queue) > 0 {
        currentID := queue[0]
        queue = queue[1:]

        if currentID == serviceID {
            return ErrCircularDependency
        }

        if visited[currentID] {
            continue
        }
        visited[currentID] = true

        // Get dependencies of current service
        deps, _, _ := u.serviceRepo.ListDependenciesByService(ctx, currentID, nil)
        for _, dep := range deps {
            if !visited[dep.DependsOnServiceID] {
                queue = append(queue, dep.DependsOnServiceID)
            }
        }
    }
    return nil
}
```

**Dependency Graph Response:**

```go
type DependencyGraphResponse struct {
    ServiceID   string
    ServiceName string
    Nodes       []GraphNode  // All services in the graph
    Edges       []GraphEdge  // All dependency relationships
}

type GraphNode struct {
    ID   string
    Name string
    Type string  // root, dependency, dependent
}

type GraphEdge struct {
    From string
    To   string
    Type string  // runtime, build, data, api
}
```

### Rightsizing Recommendations

**Usage Analysis Pattern:**

Query Prometheus for resource metrics:

```go
func (u *usecase) getCPUUsage(ctx context.Context, namespace, container string, start, end time.Time) (avg, max float64, err error) {
    avgQuery := fmt.Sprintf(
        `avg_over_time(rate(container_cpu_usage_seconds_total{namespace="%s",container="%s"}[5m])[%dd])`,
        namespace, container, lookbackDays,
    )
    results, err := u.monitoringRepo.Query(ctx, avgQuery)
    // ...
}
```

**Recommendation Types:**

- `scale_down`: utilization < 50% of request
- `scale_up`: utilization > 90% of request
- `optimal`: no recommendation needed

**Apply/Rollback Pattern:**

Store previous state as JSONB for rollback:

```go
type PreviousResourceState struct {
    CPURequest    string `json:"cpu_request"`
    CPULimit      string `json:"cpu_limit"`
    MemoryRequest string `json:"memory_request"`
    MemoryLimit   string `json:"memory_limit"`
}

// Store before applying
previousState := &PreviousResourceState{
    CPURequest: rec.CurrentCPURequest,
    // ...
}
rec.SetPreviousState(previousState)

// Apply to Kubernetes
u.provisionerRepo.UpdateDeploymentResources(ctx, namespace, name, container, ...)
```

### Resource Quotas

**Usage Calculation:**

Calculate from Kubernetes pods via informer cache:

```go
func (u *usecase) calculateUsage(ctx context.Context, namespace string) (*UsageResponse, error) {
    pods, err := u.provisionerRepo.GetPods(namespace)
    // Sum resources from all containers
    for _, pod := range pods {
        for _, container := range pod.Spec.Containers {
            totalCPURequest += container.Resources.Requests.Cpu().Value()
            totalMemRequest += container.Resources.Requests.Memory().Value()
        }
    }
    return &UsageResponse{...}, nil
}
```

**Quota Enforcement:**

Via admission webhook:

```go
func (v *Validator) checkQuota(req *admissionv1.AdmissionRequest) admission.Response {
    quota, err := v.quotaRepo.GetActiveByNamespace(ctx, namespace)
    if err != nil {
        return admission.Allowed("no quota defined")
    }

    if !quota.Enforce {
        return admission.Allowed("quota not enforced")
    }

    // Check limits
    if exceedsLimit(quota, requested) {
        return admission.Denied("quota exceeded")
    }
    return admission.Allowed("")
}
```

### Budget Alerts

**Period Window Calculation:**

```go
func getPeriodWindow(period string, now time.Time) (start, end time.Time) {
    switch period {
    case "daily":
        start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    case "weekly":
        daysSinceMonday := int(now.Weekday() - time.Monday)
        start = time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, time.UTC)
    case "monthly":
        start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    }
    return start, now
}
```

**Alert Deduplication:**

Check for existing alert before sending:

```go
existing, _ := u.budgetRepo.GetLatestAlertForThreshold(ctx, budget.ID, threshold, periodStart)
if existing != nil {
    continue  // Already alerted for this threshold in this period
}
```

### Cron Job Pattern

**Distributed Locking:**

Use Redis Sentinel for exactly-once execution:

```go
func (h *Handler) RunWithLock(ctx context.Context, jobName string, fn func() error) error {
    lock, err := h.redisLock.Acquire(ctx, jobName, 120*time.Second)
    if err != nil {
        return nil  // Another instance is running
    }
    defer lock.Release(ctx)

    return fn()
}
```

**Job Registration:**

```go
// In cmd/cron/server.go
c.AddFunc("0 * * * *", func() {
    h.RunWithLock(ctx, "cost-sync", func() error {
        return h.costSync.Sync(ctx)
    })
})
```

### Testing Patterns for Phase 2

**Mocking Multi-Repo Dependencies:**

```go
func TestAddDependency(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockSvcRepo := mocks.NewMockServiceRepository(ctrl)
    uc := New(Dependencies{ServiceRepo: mockSvcRepo})

    // Expect service lookups
    mockSvcRepo.EXPECT().GetByID(gomock.Any(), "svc-1").Return(&service.Service{...}, nil)
    mockSvcRepo.EXPECT().GetByID(gomock.Any(), "svc-2").Return(&service.Service{...}, nil)

    // Expect dependency check
    mockSvcRepo.EXPECT().ExistsDependency(gomock.Any(), "svc-1", "svc-2").Return(false, nil)

    // Expect circular dependency check
    mockSvcRepo.EXPECT().ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
        Return([]depModel.ServiceDependency{}, int64(0), nil)

    // Expect create
    mockSvcRepo.EXPECT().CreateDependency(gomock.Any(), gomock.Any()).Return(nil)

    resp, err := uc.AddDependency(ctx, "svc-1", &depModel.CreateDependencyRequest{...})
    assert.NoError(t, err)
}
```

**Testing Kubernetes Resource Updates:**

```go
func TestApplyRecommendation(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockRightsizingRepository(ctrl)
    mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)

    // Expect Kubernetes update
    mockProvisioner.EXPECT().
        UpdateDeploymentResources(gomock.Any(), "default", "api-server", "main",
            "50m", "100m", "64Mi", "128Mi").
        Return(nil)

    err := uc.ApplyRecommendation(ctx, "rec-1", "user-1")
    assert.NoError(t, err)
}
```

