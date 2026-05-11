.PHONY: all bootstrap dev-up dev-run test build docker-build docker-push lint fmt clean help

# Variables
APP_NAME := idp-core
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Docker
DOCKER_REGISTRY ?= ghcr.io
DOCKER_IMAGE := $(DOCKER_REGISTRY)/davidsugianto/$(APP_NAME)

# Go
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Directories
BIN_DIR := bin
CMD_DIR := cmd/http

# =============================================================================
# Development Setup
# =============================================================================

all: clean build

## bootstrap: Install development dependencies and tools
bootstrap:
	@echo "==> Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang/mock/mockgen@latest
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install; \
		echo "==> Pre-commit hooks installed"; \
	else \
		echo "==> Installing pre-commit..."; \
		pip install pre-commit; \
		pre-commit install; \
	fi
	@echo "==> Running go mod download..."
	$(GOMOD) download
	@echo "✅ Bootstrap complete!"

## dev-run: Run the application with hot-reload using Air (requires: go install github.com/air-verse/air@latest)
dev-run:
	@echo "==> Starting development server with hot-reload..."
	@if ! command -v air >/dev/null 2>&1; then \
		echo "❌ air not found. Install with: go install github.com/air-verse/air@latest"; \
		echo "   Or use 'make dev-app-up' to run in Docker instead."; \
		exit 1; \
	fi
	air -c .air.toml

## dev-app-up: Run the application in Docker (no air required)
dev-app-up:
	@echo "==> Starting application in Docker..."
	docker-compose up -d app
	@echo "==> Waiting for application to be ready..."
	sleep 3
	@echo "✅ Application running at http://localhost:8080"
	@echo ""
	@echo "View logs: make dev-app-logs"
	@echo "Stop app:  make dev-app-down"

## dev-app-down: Stop the application container
dev-app-down:
	@echo "==> Stopping application container..."
	docker-compose stop app
	@echo "✅ Application stopped!"

## dev-app-logs: View application logs
dev-app-logs:
	docker-compose logs -f app

## dev-all-up: Start PostgreSQL and application in Docker
dev-all-up: dev-db-up
	@echo "==> Starting application..."
	docker-compose up -d app
	sleep 3
	@echo "✅ All services running!"
	@echo ""
	@echo "PostgreSQL: localhost:5432"
	@echo "Application: http://localhost:8080"

## dev-all-down: Stop all services
dev-all-down:
	@echo "==> Stopping all services..."
	docker-compose down
	@echo "✅ All services stopped!"

# =============================================================================
# Building
# =============================================================================

## build: Build the binary
build:
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

## build-linux: Build for Linux (for Docker)
build-linux:
	@echo "==> Building $(APP_NAME) for Linux..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "✅ Linux build complete: $(BIN_DIR)/$(APP_NAME)"

# =============================================================================
# Testing
# =============================================================================

## test: Run all tests
test:
	@echo "==> Running tests..."
	$(GOTEST) -v -race ./...
	@echo "✅ Tests complete!"

## test-coverage: Run tests and generate coverage report
test-coverage:
	@echo "==> Running tests with coverage..."
	$(GOTEST) -v -short -race -coverprofile=coverage.out ./internal/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## test-short: Run short tests (skip integration tests)
test-short:
	@echo "==> Running short tests..."
	$(GOTEST) -v -short ./...
	@echo "✅ Short tests complete!"

# =============================================================================
# Development Environment (Local - PostgreSQL only)
# =============================================================================

## dev-db-up: Start PostgreSQL in Docker for local development
dev-db-up:
	@echo "==> Starting PostgreSQL in Docker..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose up -d postgres; \
		echo "==> Waiting for PostgreSQL to be ready..."; \
		sleep 3; \
		echo "✅ PostgreSQL ready at localhost:5432"; \
		echo ""; \
		echo "Connection details:"; \
		echo "  Host: localhost"; \
		echo "  Port: 5432"; \
		echo "  Database: idp_core"; \
		echo "  User: postgres"; \
		echo "  Password: postgres"; \
	else \
		echo "❌ docker-compose not found. Please install Docker Compose."; \
		exit 1; \
	fi

## dev-db-down: Stop PostgreSQL container
dev-db-down:
	@echo "==> Stopping PostgreSQL..."
	docker-compose down
	@echo "✅ PostgreSQL stopped!"

## dev-db-reset: Reset PostgreSQL database
dev-db-reset:
	@echo "==> Resetting PostgreSQL database..."
	docker-compose down -v
	docker-compose up -d postgres
	@echo "✅ Database reset complete!"

## dev-db-logs: Show PostgreSQL logs
dev-db-logs:
	docker-compose logs -f postgres

## dev-db-ps: Show PostgreSQL container status
dev-db-ps:
	@echo "==> PostgreSQL container status..."
	docker-compose ps postgres

# =============================================================================
# Development Environment (Kubernetes - Kind + ArgoCD)
# =============================================================================

## dev-k8s-setup: Setup Kind cluster with ArgoCD for integration tests
dev-k8s-setup:
	@echo "==> Setting up Kubernetes development environment..."
	@if ! command -v kind >/dev/null 2>&1; then \
		echo "Installing kind..."; \
		go install sigs.k8s.io/kind@latest; \
	fi
	./dev/setup-kind.sh
	@echo "✅ Kubernetes environment ready!"

## dev-k8s-setup-quick: Quick setup (kind + minimal ArgoCD)
dev-k8s-setup-quick:
	@echo "==> Quick setup (kind + minimal ArgoCD)..."
	@if ! command -v kind >/dev/null 2>&1; then \
		echo "Installing kind..."; \
		go install sigs.k8s.io/kind@latest; \
	fi
	@if ! kind get clusters 2>/dev/null | grep -q "idp-test"; then \
		kind create cluster --name idp-test --config dev/kind-config.yaml; \
	fi
	kubectl config use-context kind-idp-test
	./dev/setup-argocd-minimal.sh
	@echo "✅ Quick Kubernetes environment ready!"

## dev-k8s-status: Check Kubernetes environment status
dev-k8s-status:
	@echo "=== Kubernetes Environment Status ==="
	@echo ""
	@echo "Kind clusters:"
	@kind get clusters 2>/dev/null || echo "  No clusters found"
	@echo ""
	@echo "Current kubectl context:"
	@kubectl config current-context 2>/dev/null || echo "  No context set"
	@echo ""
	@echo "ArgoCD pods:"
	@kubectl get pods -n argocd 2>/dev/null || echo "  ArgoCD not installed"
	@echo ""

## dev-k8s-teardown: Tear down Kubernetes environment
dev-k8s-teardown:
	@echo "==> Tearing down Kubernetes environment..."
	./dev/teardown-kind.sh
	@echo "✅ Kubernetes environment removed!"

## dev-k8s-argocd-ui: Port-forward ArgoCD UI
dev-k8s-argocd-ui:
	@echo "==> Port-forwarding ArgoCD UI to https://localhost:8090"
	@echo "==> Getting initial admin password..."
	@kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' 2>/dev/null | base64 -d && echo ""
	@echo ""
	kubectl port-forward svc/argocd-server -n argocd 8090:443

# =============================================================================
# Development Environment (Full Setup)
# =============================================================================

## dev-setup: Full setup (PostgreSQL + Kind + ArgoCD)
dev-setup: dev-db-up dev-k8s-setup
	@echo ""
	@echo "✅ Full development environment ready!"
	@echo ""
	@echo "PostgreSQL: localhost:5432"
	@echo "Kind cluster: idp-test"
	@echo ""

## dev-setup-quick: Quick full setup (PostgreSQL + minimal ArgoCD)
dev-setup-quick: dev-db-up dev-k8s-setup-quick
	@echo ""
	@echo "✅ Quick development environment ready!"

## dev-teardown: Tear down all development environments
dev-teardown: dev-db-down dev-k8s-teardown
	@echo "✅ All development environments removed!"

## dev-status: Check all development environment status
dev-status: dev-db-ps
	@echo ""
	@echo "=== Full Development Environment Status ==="
	@echo ""
	@echo "--- PostgreSQL (Docker) ---"
	@docker-compose ps postgres 2>/dev/null || echo "  PostgreSQL not running"
	@echo ""
	@echo "--- Kubernetes (Kind) ---"
	@kind get clusters 2>/dev/null || echo "  No clusters found"
	@kubectl config current-context 2>/dev/null || echo "  No context set"
	@kubectl get pods -n argocd 2>/dev/null || echo "  ArgoCD not installed"

## test-unit: Run unit tests only (fast)
test-unit:
	@echo "==> Running unit tests..."
	$(GOTEST) -v -short -race ./...
	@echo "✅ Unit tests complete!"

## test-db: Run PostgreSQL integration tests (requires: make dev-db-up)
test-db:
	@echo "==> Running PostgreSQL integration tests..."
	@echo "Note: Run 'make dev-db-up' first if tests fail."
	$(GOTEST) -v ./internal/repository/environment/...
	@echo "✅ PostgreSQL integration tests complete!"

## test-k8s: Run Kubernetes integration tests (requires: make dev-k8s-setup)
test-k8s:
	@echo "==> Running Kubernetes integration tests..."
	@if ! command -v kubectl >/dev/null 2>&1; then \
		echo "Error: kubectl not found"; exit 1; \
	fi
	@if ! kubectl config current-context 2>/dev/null | grep -q "kind"; then \
		echo "Warning: Not connected to a kind cluster. Run 'make dev-k8s-setup' first."; \
	fi
	$(GOTEST) -v ./internal/repository/provisioner/...
	@echo "✅ Kubernetes integration tests complete!"

## test-argocd: Run ArgoCD integration tests (requires: make dev-k8s-setup)
test-argocd:
	@echo "==> Running ArgoCD integration tests..."
	@if ! kubectl get ns argocd >/dev/null 2>&1; then \
		echo "Error: ArgoCD not found. Run 'make dev-k8s-setup' first."; \
		exit 1; \
	fi
	$(GOTEST) -v ./internal/repository/gitops/...
	@echo "✅ ArgoCD integration tests complete!"

## test-integration: Run all PostgreSQL integration tests
test-integration: test-db

## test-integration-k8s: Run Kubernetes integration tests
test-integration-k8s: test-k8s

## test-integration-argocd: Run ArgoCD integration tests
test-integration-argocd: test-argocd

## test-all-integration: Run all integration tests (requires full setup)
test-all-integration:
	@echo "==> Running all integration tests..."
	@echo "Note: Requires PostgreSQL (make dev-db-up) and K8s (make dev-k8s-setup)"
	$(GOTEST) -v ./internal/repository/...
	@echo "✅ All integration tests complete!"

## test-e2e: Run E2E tests (requires: make dev-k8s-setup)
test-e2e:
	@echo "==> Running E2E tests..."
	@if ! command -v kubectl >/dev/null 2>&1; then \
		echo "Error: kubectl not found"; exit 1; \
	fi
	$(GOTEST) -v ./tests/e2e/...
	@echo "✅ E2E tests complete!"

## test-contract: Run OpenAPI contract tests
test-contract:
	@echo "==> Running contract tests..."
	$(GOTEST) -v ./tests/contract/...
	@echo "✅ Contract tests complete!"

## test-full: Run all tests with full environment setup
test-full: dev-setup
	@echo "==> Running all tests..."
	$(GOTEST) -v ./...
	@echo "✅ All tests complete!"

# =============================================================================
# Code Quality
# =============================================================================

## lint: Run golangci-lint
lint:
	@echo "==> Running golangci-lint..."
	golangci-lint run ./...
	@echo "✅ Linting complete!"

## fmt: Format code with gofmt and goimports
fmt:
	@echo "==> Formatting code..."
	gofmt -l -w .
	goimports -l -w -local github.com/davidsugianto/idp-core .
	$(GOMOD) tidy
	@echo "✅ Code formatted!"

## vet: Run go vet
vet:
	@echo "==> Running go vet..."
	$(GOCMD) vet ./...
	@echo "✅ Vet complete!"

# =============================================================================
# Database
# =============================================================================

## db-migrate: Run database migrations
db-migrate:
	@echo "==> Running database migrations..."
	$(GOCMD) run ./cmd/migrate -direction up
	@echo "✅ Migrations complete!"

## db-rollback: Rollback last database migration
db-rollback:
	@echo "==> Rolling back last migration..."
	$(GOCMD) run ./cmd/migrate -direction down
	@echo "✅ Rollback complete!"

# =============================================================================
# Docker
# =============================================================================

## docker-build: Build Docker image
docker-build: build-linux
	@echo "==> Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(VERSION)"

## docker-push: Push Docker image to registry
docker-push: docker-build
	@echo "==> Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest
	@echo "✅ Docker image pushed!"

## docker-run: Run Docker container locally
docker-run:
	@echo "==> Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

# =============================================================================
# Swagger / API Docs
# =============================================================================

## swagger-gen: Generate Swagger documentation
swagger-gen:
	@echo "==> Generating Swagger docs..."
	swag init -g cmd/http/main.go -o docs/swagger --parseDependency --parseInternal
	@echo "✅ Swagger docs generated at docs/swagger/"

# =============================================================================
# Mocks
# =============================================================================

## mock-gen: Generate mocks for testing
mock-gen:
	@echo "==> Generating mocks..."
	@find ./internal -name "mocks" -type d -exec rm -rf {} + 2>/dev/null || true
	mockgen -source=internal/repository/environment/init.go -destination=internal/mocks/repository/environment_mock.go -package=mocks
	mockgen -source=internal/usecase/environment/init.go -destination=internal/mocks/usecase/environment_mock.go -package=mocks
	@echo "✅ Mocks generated!"

# =============================================================================
# Utilities
# =============================================================================

## clean: Clean build artifacts
clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Clean complete!"

## deps: Update dependencies
deps:
	@echo "==> Updating dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated!"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
