# Test Documentation

This document provides comprehensive information about testing in idp-core.

## Test Categories

| Category | Location | Purpose | Dependencies |
|----------|----------|---------|--------------|
| Unit Tests | `internal/**/` | Test individual functions and methods | None |
| Integration Tests | `internal/**/` | Test components against real dependencies | Docker, Kind |
| E2E Tests | `tests/e2e/` | Test complete workflows | Kind, ArgoCD |
| Contract Tests | `tests/contract/` | Validate API specifications | None |

## Running Tests

### Unit Tests

Fast tests with no external dependencies:

```bash
make test-unit
```

### Integration Tests

#### PostgreSQL Tests

```bash
# Start PostgreSQL
make dev-db-up

# Run tests
make test-db

# Stop PostgreSQL when done
make dev-db-down
```

#### Kubernetes Tests

```bash
# Setup Kind cluster
make dev-k8s-setup-quick

# Run Kubernetes tests
make test-k8s

# Run ArgoCD tests
make test-argocd

# Teardown when done
make dev-k8s-teardown
```

### E2E Tests

End-to-end tests validate complete workflows:

```bash
# Setup environment
make dev-k8s-setup

# Run E2E tests
make test-e2e
```

### Contract Tests

Validate OpenAPI specification:

```bash
# Generate swagger docs first
make swagger-gen

# Run contract tests
make test-contract
```

### All Tests

```bash
# Full setup and run all tests
make dev-setup
make test-all-integration
```

---

## E2E Tests

### Overview

E2E tests validate complete workflows from API to Kubernetes to ArgoCD. They ensure all components work together correctly.

### Test Files

| File | Description |
|------|-------------|
| `tests/e2e/e2e_test.go` | Main E2E test suite |

### Test Cases

#### TestE2E_HealthCheck

Validates the health endpoint returns correct status.

```go
// Tests: GET /health
// Expected: {"status": "ok"}
```

#### TestE2E_FullEnvironmentFlow

Tests the complete environment lifecycle:

1. **Create Namespace** - Simulates environment creation
2. **Verify Namespace** - Confirms namespace exists in K8s
3. **Create Pod** - Deploys a test workload
4. **Verify Pod** - Confirms pod is running
5. **Delete Namespace** - Cleanup

```go
// Requires: Kubernetes cluster with ArgoCD
// Timeout: ~30 seconds
```

#### TestE2E_AuthFlow

Tests JWT authentication:

- Token generation
- Token validation
- Claims extraction

#### TestE2E_APIAuthentication

Tests API authentication middleware:

| Scenario | Expected Status |
|----------|-----------------|
| Missing token | 401 Unauthorized |
| Invalid token | 401 Unauthorized |
| Valid token | 200 OK |

#### TestE2E_NamespaceNaming

Validates namespace naming conventions:

- DNS-1123 compliance
- Max 63 characters
- Lowercase alphanumeric and hyphens only
- Auto-truncation for long names

### Prerequisites

1. **Kubernetes Cluster**

```bash
# Create Kind cluster
make dev-k8s-setup-quick
```

2. **ArgoCD (for full flow tests)**

```bash
# Setup with ArgoCD
make dev-k8s-setup
```

### Running E2E Tests

```bash
# Quick E2E tests (no ArgoCD required)
make test-e2e

# Full E2E tests (requires ArgoCD)
make dev-k8s-setup
make test-e2e
```

### E2E Test Configuration

Tests automatically skip if dependencies are not available:

```go
func skipIfNoK8s(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    // Check Kubernetes availability...
}
```

To skip E2E tests:

```bash
go test -short ./tests/e2e/...
```

### E2E Test Best Practices

1. **Unique Resource Names**

```go
namespace := fmt.Sprintf("idp-e2e-test-%d", time.Now().UnixNano())
```

2. **Cleanup After Tests**

```go
defer func() {
    _ = k8sClient.Clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
}()
```

3. **Wait for Async Operations**

```go
for i := 0; i < 30; i++ {
    _, err = client.Get(ctx, namespace)
    if err != nil {
        break
    }
    time.Sleep(500 * time.Millisecond)
}
```

---

## Contract Tests

### Overview

Contract tests validate that the OpenAPI specification is correct and complete.

### Test Files

| File | Description |
|------|-------------|
| `tests/contract/openapi_test.go` | OpenAPI spec validation |

### Test Cases

| Test | Description |
|------|-------------|
| `TestOpenAPISpecExists` | Validates spec file exists |
| `TestOpenAPISpecValidJSON` | Validates JSON format |
| `TestOpenAPIInfoSection` | Validates info section |
| `TestOpenAPIPaths` | Validates API paths exist |
| `TestOpenAPIPathMethods` | Validates HTTP methods |
| `TestOpenAPISecurityDefinitions` | Validates security definitions |
| `TestOpenAPIDefinitions` | Validates model definitions |
| `TestOpenAPIResponseCodes` | Validates HTTP status codes |

### Running Contract Tests

```bash
# Generate swagger docs first
make swagger-gen

# Run contract tests
make test-contract
```

### Contract Test Output

```
=== RUN   TestOpenAPISpecExists
    Found OpenAPI spec at: docs/swagger/swagger.json
--- PASS: TestOpenAPISpecExists

=== RUN   TestOpenAPIPaths
    Available API paths:
      /v1/environments
      /v1/environments/{id}
      /v1/environments/{id}/sync
      ...
--- PASS: TestOpenAPIPaths
```

---

## Integration Tests

### PostgreSQL Integration Tests

Located in `internal/repository/environment/environment_integration_test.go`

Uses `testcontainers-go` for isolated PostgreSQL instances.

```bash
make dev-db-up
make test-db
```

### Kubernetes Integration Tests

Located in `internal/repository/provisioner/kubernetes_integration_test.go`

Tests namespace provisioning, resource quotas, and network policies.

```bash
make dev-k8s-setup-quick
make test-k8s
```

### ArgoCD Integration Tests

Located in `internal/repository/gitops/argocd_integration_test.go`

Tests ArgoCD Application CRUD operations.

```bash
make dev-k8s-setup
make test-argocd
```

---

## CI/CD Integration

### GitHub Actions

Tests run automatically in CI:

```yaml
# Unit tests on every PR
- name: Run unit tests
  run: make test-unit

# Integration tests on main branch
- name: Run PostgreSQL tests
  run: make test-db

- name: Run Kubernetes tests
  run: make test-k8s
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
open coverage.html
```

---

## Troubleshooting

### Tests Skip Unexpectedly

Tests skip when dependencies are not available:

```bash
# Check Kubernetes
kubectl cluster-info

# Check ArgoCD
kubectl get ns argocd
kubectl get pods -n argocd
```

### Database Connection Errors

```bash
# Ensure PostgreSQL is running
docker ps | grep postgres

# Restart PostgreSQL
make dev-db-down
make dev-db-up
```

### Kubernetes Connection Errors

```bash
# Check context
kubectl config current-context

# Switch to Kind cluster
kubectl config use-context kind-idp-test
```

### Timeout Errors

Increase timeout for slow environments:

```bash
# For ArgoCD setup
TIMEOUT=900 make dev-k8s-setup
```

---

## Test Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Unit | `Test<Function>_<Scenario>` | `TestCreate_Success` |
| Integration | `TestRepository_<Method>_<Scenario>` | `TestRepository_Create_Success` |
| E2E | `TestE2E_<Feature>_<Scenario>` | `TestE2E_AuthFlow_ValidToken` |
| Contract | `TestOpenAPI<Component>_<Check>` | `TestOpenAPIPaths_ValidMethods` |

---

## Adding New Tests

### Adding E2E Tests

1. Create test in `tests/e2e/`
2. Use skip helpers for dependencies
3. Follow naming convention
4. Add cleanup with `defer`

```go
func TestE2E_NewFeature(t *testing.T) {
    skipIfNoK8s(t)

    // Setup
    ctx := context.Background()

    // Cleanup
    defer cleanup()

    // Test logic
    // ...
}
```

### Adding Contract Tests

1. Create test in `tests/contract/`
2. Skip if swagger not generated
3. Validate OpenAPI spec aspects

```go
func TestOpenAPI<Aspect>(t *testing.T) {
    specPath := "docs/swagger/swagger.json"
    if _, err := os.Stat(specPath); os.IsNotExist(err) {
        t.Skip("Run 'make swagger-gen' first")
    }

    // Validation logic
    // ...
}
```
