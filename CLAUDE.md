# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Install dependencies
go mod download

# Run the application
go run ./cmd/http

# Build binary
go build -o bin/idp-api ./cmd/http

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a single test
go test -run TestFunctionName ./path/to/package
```

## Architecture

This is an Internal Developer Platform (IDP) API following clean architecture with strict layer separation:

```
cmd/http/                    # Entry point & server setup
internal/
├── handler/http/            # HTTP handlers (controllers)
├── usecase/                 # Business logic layer
├── repository/              # Data access layer (DB, external APIs)
└── model/                   # Domain models and types
```

**Layer Rules:**
- `handler` → calls `usecase` only
- `usecase` → calls `repository` only
- `repository` → accesses DB/external APIs
- Dependencies flow inward; inner layers never import outer layers

**Dependency Injection Pattern:**
Each layer uses a `Dependencies` struct for constructor injection:
```go
type Dependencies struct {
    EnvironmentRepo environmentRepo.Repository
}

func New(deps Dependencies) Usecase { ... }
```

When adding a new feature (e.g., "project"):
1. Create `internal/model/project/type.go` for domain types
2. Create `internal/repository/project/` with `init.go` (interface + struct) and `project.go` (methods)
3. Create `internal/usecase/project/` with same pattern
4. Create `internal/handler/http/project.go` for HTTP endpoints
5. Wire dependencies in `cmd/http/server.go`

## Tech Stack

- Go 1.25+, Gin (web framework), GORM (ORM), PostgreSQL
- client-go (Kubernetes), controller-runtime (CRD controllers)
- Viper (config), Logrus (logging), swaggo (OpenAPI)

## Configuration

Config loaded from `configs/config.yaml` with environment variable overrides. See README.md for full schema.
