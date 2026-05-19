# 📋 idp-core — Development TODO List (Phase 2)

> **Phase**: 2 - Enhancement\
> **Timeline**: Q3 2026 (\~10 weeks)\
> **Status**: 🔄 In Development\
> **Last Updated**: May 2026

***

## 📊 Progress Overview

| Milestone            | Status         | Progress |
| -------------------- | -------------- | -------- |
| M1: Auth & RBAC      | 🔄 In Progress | 85%      |
| M2: FinOps           | ✅ Complete    | 100%     |
| M3: Rightsizing      | ✅ Complete    | 100%     |
| M4: Service Catalog  | ✅ Complete    | 100%     |
| M5: Testing & Polish | 🔄 In Progress | 80%      |

***

## 🗓️ M1: Auth & RBAC (Week 1-3)

> **Implementation**: ✅ Complete (all models, repositories, usecases, handlers, middleware, unit tests)
> **Remaining**: Integration tests (6 suites), platform admin seed script

### Week 1: User & Team Management ✅ COMPLETED

#### Database & Models

- [x] Create migration: `users` table
- [x] Create migration: `teams` table
- [x] Create migration: `team_members` table
- [x] Create model: `internal/model/user/type.go`
- [x] Create model: `internal/model/team/type.go`
- [x] Create model: `internal/model/team_member/type.go` (included in team/type.go)

#### Repository Layer

- [x] Create `internal/repository/user/init.go` (interface + struct)
- [x] Create `internal/repository/user/user.go` (CRUD methods)
- [x] Create `internal/repository/team/init.go`
- [x] Create `internal/repository/team/team.go`
- [x] Create `internal/repository/team/member.go`

#### Usecase Layer

- [x] Create `internal/usecase/user/init.go`
- [x] Create `internal/usecase/user/user.go`
- [x] Create `internal/usecase/team/init.go`
- [x] Create `internal/usecase/team/team.go`

#### Handler Layer

- [x] Create `internal/handler/http/user.go`
- [x] Create `internal/handler/http/team.go`
- [x] Add routes in `cmd/http/server.go`

#### Tests

- [x] Unit tests: user repository (mock created)
- [x] Unit tests: team repository (mock created)
- [x] Unit tests: user usecase
- [x] Unit tests: team usecase
- [ ] Integration tests: user API
- [ ] Integration tests: team API

***

### Week 2: OIDC Integration & RBAC ✅ COMPLETED

#### OIDC Integration

- [x] Add dependency: `github.com/coreos/go-oidc`
- [x] Create `internal/pkg/oidc/client.go`
- [x] Create `internal/pkg/oidc/verifier.go`
- [x] Update config: OIDC settings in `config.go`
- [x] Create auth middleware: `internal/handler/http/middleware/oidc.go`

#### Role & Permission Models

- [x] Create migration: `roles` table
- [x] Create migration: `permissions` table
- [x] Create migration: `role_permissions` table
- [x] Create migration: `user_roles` table
- [x] Create model: `internal/model/role/type.go`
- [x] Create model: `internal/model/permission/type.go`

#### RBAC Repository

- [x] Create `internal/repository/role/init.go`
- [x] Create `internal/repository/role/role.go`
- [x] Create `internal/repository/role/member.go`
- [x] Create `internal/repository/permission/init.go`
- [x] Create `internal/repository/permission/permission.go`

#### RBAC Usecase

- [x] Create `internal/usecase/role/init.go`
- [x] Create `internal/usecase/role/role.go`
- [x] Create RBAC engine: `internal/usecase/auth/rbac.go`

#### RBAC Handler

- [x] Create `internal/handler/http/role.go`
- [x] Update auth middleware with RBAC check
- [x] Add role routes in `cmd/http/server.go`

#### Seed Default Data

- [x] Create seed script for default roles
- [x] Create seed script for default permissions
- [ ] Create seed script for platform admin user

#### Tests

- [x] Unit tests: OIDC client
- [x] Unit tests: role repository (mock created)
- [x] Unit tests: RBAC engine
- [ ] Integration tests: OIDC flow
- [ ] Integration tests: RBAC enforcement

***

### Week 3: API Keys & Audit Logging ✅ COMPLETED

#### API Key Models

- [x] Create migration: `api_keys` table
- [x] Create model: `internal/model/apikey/type.go`

#### API Key Repository

- [x] Create `internal/repository/apikey/init.go`
- [x] Create `internal/repository/apikey/apikey.go`

#### API Key Usecase

- [x] Create `internal/usecase/apikey/init.go`
- [x] Create `internal/usecase/apikey/apikey.go`
- [x] Implement key generation & hashing

#### API Key Handler

- [x] Create `internal/handler/http/apikey.go`
- [x] Add API key authentication middleware
- [x] Add API key routes

#### Audit Logging

- [x] Create migration: `audit_logs` table
- [x] Create model: `internal/model/audit_log/type.go`
- [x] Create `internal/repository/auditlog/init.go`
- [x] Create `internal/repository/auditlog/auditlog.go`
- [x] Create `internal/usecase/auditlog/init.go`
- [x] Create `internal/usecase/auditlog/auditlog.go`
- [x] Create audit middleware for automatic logging
- [x] Create `internal/handler/http/auditlog.go`

#### Tests

- [x] Unit tests: API key generation
- [x] Unit tests: audit log repository
- [ ] Integration tests: API key auth
- [ ] Integration tests: audit log retrieval

***

## 🗓️ M2: FinOps (Week 4-5)

### Week 4: OpenCost Integration & Cost Tracking ✅ COMPLETED

#### Configuration

- [x] Add OpenCost config to `config.go`
- [x] Add Prometheus config to `config.go`

#### OpenCost Client

- [x] Create `internal/pkg/opencost/client.go`
- [x] Create `internal/pkg/opencost/types.go` (AllocationRequest, AllocationResponse, AllocationData)

#### Prometheus Client

- [x] Create `internal/pkg/prometheus/client.go` (stub for future queries)

#### Cost Models

- [x] Create migration: `cost_records` table
- [x] Create model: `internal/model/cost/type.go`

#### Cost Repository

- [x] Create `internal/repository/cost/init.go` (interface + implementation)
- [x] Implement Create, BatchCreate, List, GetByTeamAndPeriod

#### Cost Usecase

- [x] Create `internal/usecase/cost/init.go`
- [x] Create `internal/usecase/cost/cost.go`
- [x] Implement cost sync job via separate cron server (`cmd/cron/`) with Redis distributed locking

#### Cost Handler

- [x] Create `internal/handler/http/cost.go`
- [x] Add cost routes (`GET /v1/costs`, `GET /v1/costs/team`)

#### Tests

- [x] Unit tests: OpenCost client (5 tests with httptest.Server)
- [x] Unit tests: cost usecase (11 tests with mocked repo + opencost client)
- [ ] Integration tests: cost API

***

### Week 5: Budget Management & Alerts ✅ COMPLETED

#### Budget Models

- [x] Create migration: `budgets` table
- [x] Create migration: `budget_alerts` table
- [x] Create model: `internal/model/budget/type.go`
- [x] Create model: `internal/model/budget_alert/type.go` (included in budget/type.go)

#### Budget Repository

- [x] Create `internal/repository/budget/init.go` (interface + implementation)
- [x] Implement Create, GetByID, ListByTeam, ListActive, Update, Delete
- [x] Implement CreateAlert, GetAlertsByBudget, GetLatestAlertForThreshold

#### Budget Usecase

- [x] Create `internal/usecase/budget/init.go`
- [x] Create `internal/usecase/budget/budget.go`
- [x] Implement budget alert checker (CheckAlerts with cron)
- [x] Implement Slack notification sender

#### Slack Client

- [x] Create `internal/pkg/slack/client.go`
- [x] Add SlackConfig to `internal/pkg/config/config.go`

#### Budget Handler

- [x] Create `internal/handler/http/budget.go`
- [x] Add budget routes (`/v1/budgets`)
- [x] Create `internal/handler/cron/budget.go` (BudgetAlertCheck)

#### Tests

- [x] Unit tests: budget usecase (24 tests)
- [ ] Integration tests: budget API
- [ ] Integration tests: alert triggering

***

## 🗓️ M3: Rightsizing (Week 6-7)

### Week 6: Rightsizing Recommendations ✅ COMPLETED

#### Configuration

- [x] Add rightsizing config to `config.go`

#### Rightsizing Models

- [x] Create migration: `rightsizing_recommendations` table
- [x] Create model: `internal/model/rightsizing/type.go`

#### Rightsizing Repository

- [x] Create `internal/repository/rightsizing/init.go`
- [x] Create `internal/repository/rightsizing/rightsizing.go`

#### Rightsizing Usecase

- [x] Create `internal/usecase/rightsizing/init.go`
- [x] Create `internal/usecase/rightsizing/rightsizing.go`
- [x] Implement usage analyzer (query Prometheus)
- [x] Implement recommendation generator
- [x] Implement recommendation scheduler (cron)

#### Rightsizing Handler

- [x] Create `internal/handler/http/rightsizing.go`
- [x] Add rightsizing routes

#### K8s Integration

- [x] Implement apply recommendation (update workload)
- [x] Handle rollback on failure

#### Tests

- [x] Unit tests: usage analyzer
- [x] Unit tests: recommendation generator
- [ ] Integration tests: rightsizing API
- [ ] E2E test: apply recommendation

***

### Week 7: Resource Quotas ✅ COMPLETED

#### Resource Quota Models

- [x] Create migration: `resource_quotas` table
- [x] Create model: `internal/model/resourcequota/type.go`

#### Resource Quota Repository

- [x] Create `internal/repository/quota/init.go`
- [x] Create `internal/repository/quota/quota.go`

#### Resource Quota Usecase

- [x] Create `internal/usecase/quota/init.go`
- [x] Create `internal/usecase/quota/quota.go`
- [x] Implement usage calculator
- [x] Implement quota enforcement (admission webhook)

#### Resource Quota Handler

- [x] Create `internal/handler/http/quota.go`
- [x] Add quota routes

#### Admission Webhook Update

- [x] Update webhook to check quotas
- [x] Return quota exceeded error

#### Tests

- [x] Unit tests: quota repository
- [x] Unit tests: quota enforcement
- [ ] Integration tests: quota API
- [ ] E2E test: quota enforcement

***

## 🗓️ M4: Service Catalog (Week 8-9)

### Week 8: Service Registration & Discovery ✅ COMPLETED

#### Service Models

- [x] Create migration: `services` table
- [x] Create migration: `service_versions` table
- [x] Create migration: `service_endpoints` table
- [x] Create model: `internal/model/service/type.go`
- [x] Create model: `internal/model/service_version/type.go`
- [x] Create model: `internal/model/service_endpoint/type.go`

#### Service Repository

- [x] Create `internal/repository/service/init.go`
- [x] Create `internal/repository/service/service.go`
- [x] Create `internal/repository/service/version.go`

#### Service Usecase

- [x] Create `internal/usecase/service/init.go`
- [x] Create `internal/usecase/service/service.go`

#### Service Handler

- [x] Create `internal/handler/http/service.go`
- [x] Add service routes

#### Tests

- [x] Unit tests: service repository
- [x] Unit tests: service usecase
- [x] Integration tests: service API

***

### Week 9: Dependencies & Environments ✅ COMPLETED

#### Dependency Model

- [x] Create migration: `service_dependencies` table
- [x] Create model: `internal/model/service_dependency/type.go`

#### Service Environment Model

- [x] Create migration: `service_environments` table
- [x] Create model: `internal/model/service_environment/type.go`

#### Repository Extensions

- [x] Add dependency methods to service repository
- [x] Add environment methods to service repository

#### Usecase Extensions

- [x] Add dependency management methods
- [x] Add environment deployment tracking

#### Handler Extensions

- [x] Add dependency routes
- [x] Add service environment routes

#### Dependency Visualization

- [x] Create endpoint for dependency graph
- [x] Format for frontend consumption

#### Tests

- [x] Unit tests: dependency logic
- [ ] Integration tests: dependency API
- [ ] Integration tests: environment tracking

***

## 🗓️ M5: Testing & Polish (Week 10)

### Integration Testing

- [x] Create comprehensive integration test suite
- [ ] Test all Phase 2 API endpoints
- [ ] Test OIDC flow end-to-end
- [ ] Test RBAC enforcement across features
- [ ] Test cost sync and budget alerts
- [ ] Test rightsizing recommendations

### E2E Testing

- [ ] Create E2E test scenarios for Phase 2
- [ ] Test full user journey: login → create env → view costs
- [ ] Test admin journey: manage roles → set budgets → view audit logs

### Performance Testing

- [ ] Load test auth endpoints
- [ ] Load test cost queries
- [ ] Identify and fix bottlenecks

### Documentation

- [x] Update Swagger/OpenAPI specs
- [x] Update `README.md` with Phase 2 features
- [x] Update `DEV_GUIDELINE.md` with new patterns
- [ ] Create Phase 2 deployment guide
- [ ] Update `TEST.md` with new test scenarios

### Security Review

- [ ] Review OIDC implementation for vulnerabilities
- [ ] Review RBAC for privilege escalation risks
- [ ] Review API key handling
- [ ] Review audit log for sensitive data

### Final Checks

- [x] Run all tests: `go test ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Check test coverage: `go test -cover ./...`
- [ ] Verify Swagger docs render correctly
- [ ] Test Docker build: `docker build -t idp-core:v2 .`

***

## 📦 Dependencies to Add

```go
// Phase 2 dependencies
require (
    // OIDC
    github.com/coreos/go-oidc v1.2.1
    golang.org/x/oauth2 v0.20.0
    
    // Password hashing
    golang.org/x/crypto v0.23.0
    
    // API Key generation
    github.com/google/uuid v1.6.0
    
    // Prometheus client (for queries)
    github.com/prometheus/client_golang v1.19.0
    github.com/prometheus/common v0.53.0
    
    // CSV export
    encoding/csv (stdlib)
    
    // Slack webhook (optional)
    github.com/slack-go/slack v0.13.0
)
```

***

## 🔧 Configuration Updates

Add to `configs/config.yaml`:

```yaml
# Phase 2 configuration
auth:
  oidc:
    enabled: true
    provider: "${OIDC_PROVIDER}"
    issuer_url: "${OIDC_ISSUER_URL}"
    client_id: "${OIDC_CLIENT_ID}"
    client_secret: "${OIDC_CLIENT_SECRET}"
    redirect_url: "${OIDC_REDIRECT_URL}"
    scopes:
      - openid
      - profile
      - email
      - groups
    groups_claim: "groups"
    admin_group: "platform-admins"

finops:
  enabled: true
  opencost:
    base_url: "http://opencost.opencost.svc.cluster.local:9003"
    poll_interval: "1h"
  prometheus:
    url: "http://prometheus-server.monitoring.svc.cluster.local:80"

cron:
  grace_timeout: "900s"
  schedules:
    ping: "*/5 * * * *"
    cost-sync: "0 * * * *"
  port: 8983

redis:
  master_name: "idp-core-redis_sentinel"
  address: "redis-sentinel.idp-core.svc.cluster.local:26379"
  password: "${REDIS_PASSWORD}"

rightsizing:
  enabled: true
  analysis_interval: "1h"
  lookback_days: 7
  recommendation_ttl: "168h"

service_catalog:
  enabled: true
  default_visibility: "team"
```

***

## 📁 File Structure (Phase 2)

```
cmd/
├── http/main.go             # ✅ API server entry point
├── http/server.go           # ✅ API server setup
├── cron/main.go             # ✅ Cron server entry point
└── cron/server.go           # ✅ Cron server setup

internal/
├── handler/http/
│   ├── user.go              # ✅ CREATED
│   ├── team.go              # ✅ CREATED
│   ├── role.go              # ✅ CREATED
│   ├── apikey.go            # ✅ CREATED
│   ├── auditlog.go          # ✅ CREATED
│   ├── cost.go              # ✅ CREATED
│   ├── budget.go            # ✅ CREATED
│   ├── rightsizing.go       # ✅ CREATED
│   ├── quota.go             # ✅ CREATED
│   └── service.go           # ✅ CREATED
│
├── handler/cron/
│   ├── init.go              # ✅ CREATED
│   ├── cost.go              # ✅ CREATED
│   ├── budget.go            # ✅ CREATED
│   └── rightsizing.go       # ✅ CREATED
│
├── usecase/
│   ├── user/                # ✅ CREATED
│   ├── team/                # ✅ CREATED
│   ├── role/                # ✅ CREATED
│   ├── apikey/              # ✅ CREATED
│   ├── auditlog/            # ✅ CREATED
│   ├── auth/                # ✅ CREATED (RBAC engine)
│   ├── cost/                # ✅ CREATED
│   ├── budget/              # ✅ CREATED
│   ├── rightsizing/         # ✅ CREATED
│   ├── quota/               # ✅ CREATED
│   └── service/             # ✅ CREATED
│
├── repository/
│   ├── user/                # ✅ CREATED
│   ├── team/                # ✅ CREATED
│   ├── role/                # ✅ CREATED
│   ├── permission/          # ✅ CREATED
│   ├── apikey/              # ✅ CREATED
│   ├── auditlog/            # ✅ CREATED
│   ├── cost/                # ✅ CREATED
│   ├── budget/              # ✅ CREATED
│   ├── rightsizing/         # ✅ CREATED
│   ├── monitoring/          # ✅ CREATED (Prometheus wrapper)
│   ├── quota/               # ✅ CREATED
│   └── service/             # ✅ CREATED
│
├── model/
│   ├── user/                # ✅ CREATED
│   ├── team/                # ✅ CREATED
│   ├── role/                # ✅ CREATED
│   ├── permission/          # ✅ CREATED
│   ├── apikey/              # ✅ CREATED
│   ├── auditlog/            # ✅ CREATED
│   ├── cost/                # ✅ CREATED
│   ├── budget/              # ✅ CREATED
│   ├── rightsizing/         # ✅ CREATED
│   ├── resourcequota/       # ✅ CREATED
│   └── service/             # ✅ CREATED
│
├── pkg/
│   ├── oidc/                # ✅ CREATED
│   ├── opencost/            # ✅ CREATED
│   ├── prometheus/          # ✅ CREATED
│   ├── slack/               # ✅ CREATED
│   ├── webhook/             # ✅ CREATED (admission validator)
│   └── redislock/           # ✅ CREATED (distributed locking)
│
└── mocks/
    ├── user_repository.go       # ✅ CREATED
    ├── team_repository.go       # ✅ CREATED
    ├── role_repository.go       # ✅ CREATED
    ├── permission_repository.go # ✅ CREATED
    ├── apikey_repository.go     # ✅ CREATED
    ├── auditlog_repository.go   # ✅ CREATED
    ├── cost_repository.go       # ✅ CREATED
    ├── budget_repository.go     # ✅ CREATED
    ├── rightsizing_repository.go # ✅ CREATED
    ├── quota_repository.go      # ✅ CREATED
    ├── monitoring_repository.go # ✅ CREATED
    ├── service_repository.go    # ✅ CREATED
    ├── provisioner_repository.go # ✅ CREATED (for rightsizing)
    ├── slack_notifier.go        # ✅ CREATED
    └── opencost_client.go       # ✅ CREATED

deployments/
└── kubernetes/
    ├── base/
    │   ├── cron-deployment.yaml # ✅ CREATED
    │   ├── cron-service.yaml    # ✅ CREATED
    │   └── cron-rbac.yaml       # ✅ CREATED
    └── overlays/production/     # ✅ (updated)

Dockerfile.cron                 # ✅ CREATED
.air.cron.toml                  # ✅ CREATED

migrations/
├── 20260501000000_create_users_table.sql        # ✅ CREATED
├── 20260501000001_create_teams_table.sql        # ✅ CREATED
├── 20260501000002_create_team_members_table.sql # ✅ CREATED
├── 20260502000000_create_roles_table.sql        # ✅ CREATED
├── 20260502000001_create_permissions_table.sql  # ✅ CREATED
├── 20260502000002_create_role_permissions_table.sql # ✅ CREATED
├── 20260502000003_create_user_roles_table.sql   # ✅ CREATED
├── 20260512000000_create_api_keys_table.sql     # ✅ CREATED
├── 20260512000001_create_audit_logs_table.sql   # ✅ CREATED
├── 20260513000000_create_cost_records_table.sql # ✅ CREATED
├── 20260514000000_create_budgets_table.sql      # ✅ CREATED
├── 20260514000001_create_budget_alerts_table.sql # ✅ CREATED
├── 20260515000000_create_rightsizing_recommendations_table.sql # ✅ CREATED
├── 20260516000000_create_resource_quotas_table.sql # ✅ CREATED
├── 20260519000000_create_services_table.sql # ✅ CREATED
├── 20260519000001_create_service_versions_table.sql # ✅ CREATED
├── 20260519000002_create_service_endpoints_table.sql # ✅ CREATED
├── 20260519000003_create_service_dependencies_table.sql # ✅ CREATED
└── 20260519000004_create_service_environments_table.sql # ✅ CREATED
```

***

## ✅ Definition of Done

Each task is considered complete when:

1. ✅ Code is written and follows clean architecture
2. ✅ Unit tests pass with > 80% coverage
3. ✅ Integration tests pass
4. ✅ Swagger documentation is updated
5. ✅ Code passes `golangci-lint`
6. ✅ Code is reviewed and merged

***

## 📝 Notes

- Start with M1 (Auth & RBAC) as it's foundational for other features
- OIDC integration may require external provider setup (Keycloak/Okta)
- OpenCost integration requires OpenCost to be deployed in cluster
- Budget alerts require email/Slack integration setup
- Consider feature flags for gradual rollout

***

## ✅ Completed Work

### M1 Week 1: User & Team Management (May 2026)

**Files Created:**

- `migrations/20260501000000_create_users_table.sql`
- `migrations/20260501000001_create_teams_table.sql`
- `migrations/20260501000002_create_team_members_table.sql`
- `internal/model/user/type.go`
- `internal/model/team/type.go`
- `internal/repository/user/init.go`, `user.go`
- `internal/repository/team/init.go`, `team.go`, `member.go`
- `internal/usecase/user/init.go`, `user.go`
- `internal/usecase/team/init.go`
- `internal/handler/http/user.go`
- `internal/handler/http/team.go`
- `internal/mocks/user_repository.go`
- `internal/mocks/team_repository.go`
- `internal/usecase/user/user_test.go`
- `internal/usecase/team/team_test.go`

**API Endpoints Added:**

| Method | Endpoint                        | Description                |
| ------ | ------------------------------- | -------------------------- |
| GET    | `/v1/users`                     | List users with pagination |
| POST   | `/v1/users`                     | Create a new user          |
| GET    | `/v1/users/:id`                 | Get user details           |
| PATCH  | `/v1/users/:id`                 | Update user                |
| DELETE | `/v1/users/:id`                 | Delete user                |
| PUT    | `/v1/users/:id/status`          | Update user status         |
| GET    | `/v1/teams`                     | List teams with pagination |
| POST   | `/v1/teams`                     | Create a new team          |
| GET    | `/v1/teams/:id`                 | Get team details           |
| PATCH  | `/v1/teams/:id`                 | Update team                |
| DELETE | `/v1/teams/:id`                 | Delete team                |
| GET    | `/v1/teams/:id/members`         | List team members          |
| POST   | `/v1/teams/:id/members`         | Add team member            |
| PATCH  | `/v1/teams/:id/members/:userId` | Update member role         |
| DELETE | `/v1/teams/:id/members/:userId` | Remove team member         |

### M1 Week 2: OIDC Integration & RBAC (May 2026)

**Files Created:**

- `migrations/20260502000000_create_roles_table.sql`
- `migrations/20260502000001_create_permissions_table.sql`
- `migrations/20260502000002_create_role_permissions_table.sql`
- `migrations/20260502000003_create_user_roles_table.sql`
- `internal/model/role/type.go`, `user_role.go`
- `internal/model/permission/type.go`
- `internal/repository/role/init.go`, `role.go`, `member.go`
- `internal/repository/permission/init.go`, `permission.go`
- `internal/usecase/role/init.go`, `role.go`, `role_test.go`
- `internal/usecase/auth/init.go`, `rbac.go`
- `internal/handler/http/role.go`
- `internal/handler/http/middleware/oidc.go`
- `internal/pkg/oidc/client.go`, `verifier.go`
- `internal/pkg/config/config.go` (updated with OIDC settings)
- `internal/mocks/role_repository.go`
- `internal/mocks/permission_repository.go`
- `internal/seed/permission.go`, `role.go`

**API Endpoints Added:**

| Method | Endpoint                                    | Description                |
| ------ | ------------------------------------------- | -------------------------- |
| GET    | `/v1/roles`                                 | List roles with pagination |
| POST   | `/v1/roles`                                 | Create a new role          |
| GET    | `/v1/roles/:id`                             | Get role details           |
| PATCH  | `/v1/roles/:id`                             | Update role                |
| DELETE | `/v1/roles/:id`                             | Delete role                |
| POST   | `/v1/roles/assign`                          | Assign role to user        |
| POST   | `/v1/roles/revoke`                          | Revoke role from user      |
| GET    | `/v1/users/:user_id/roles`                  | Get user's roles           |
| GET    | `/v1/teams/:team_id/members/:user_id/roles` | Get user's team roles      |

### M1 Week 3: API Keys & Audit Logging (May 2026)

**Files Created:**

- `migrations/20260512000000_create_api_keys_table.sql`
- `migrations/20260512000001_create_audit_logs_table.sql`
- `internal/model/auditlog/type.go` (AuditLog, Map, constants, request/response types, converters)
- `internal/repository/auditlog/init.go`, `auditlog.go`
- `internal/usecase/apikey/init.go`, `apikey.go`, `apikey_test.go`
- `internal/usecase/auditlog/init.go`, `auditlog_test.go`
- `internal/handler/http/apikey.go`
- `internal/handler/http/auditlog.go`
- `internal/handler/http/middleware/apikey.go` (API key auth middleware)
- `internal/handler/http/middleware/audit.go` (automatic audit logging middleware)
- `internal/mocks/apikey_repository.go`
- `internal/mocks/auditlog_repository.go`

**Existing Files (already from Phase 1):**

- `internal/model/apikey/type.go`
- `internal/repository/apikey/init.go`

**API Endpoints Added:**

| Method | Endpoint             | Description                |
| ------ | -------------------- | -------------------------- |
| GET    | `/v1/api-keys`       | List API keys              |
| POST   | `/v1/api-keys`       | Create a new API key       |
| GET    | `/v1/api-keys/:id`   | Get API key details        |
| PATCH  | `/v1/api-keys/:id`   | Update API key             |
| DELETE | `/v1/api-keys/:id`   | Delete (revoke) API key    |
| GET    | `/v1/audit-logs`     | List audit logs (filtered) |
| GET    | `/v1/audit-logs/:id` | Get audit log entry        |

**Key Features:**

- API key generation with `idp_` prefix and hex-encoded random suffix
- SHA-256 hashing for storage; plain key returned only on creation
- API key auth middleware supporting both `Authorization: Bearer` and `X-API-Key` headers
- Automatic audit logging middleware on all `/v1` routes (fire-and-forget)
- Audit log filtering by user, team, action, resource type, status, and date range

**Next Steps:**

- Add integration tests for role API, API key auth, and audit log retrieval
- Complete seed script for platform admin user
- Begin M2: FinOps (Week 4)

### M2 Week 4: OpenCost Integration & Cost Tracking (May 2026)

**Files Created:**

- `migrations/20260513000000_create_cost_records_table.sql`
- `internal/model/cost/type.go` (CostRecord, CostRecordResponse, CostListResponse, CostFilter, converters)
- `internal/repository/cost/init.go` (interface + implementation: Create, BatchCreate, List, GetByTeamAndPeriod)
- `internal/usecase/cost/init.go`, `cost.go`, `cost_test.go`
- `internal/handler/http/cost.go`
- `internal/handler/cron/init.go`, `cost.go` — cron job handlers
- `internal/pkg/opencost/client.go`, `client_test.go`, `types.go`
- `internal/pkg/prometheus/client.go` (stub for future queries)
- `internal/pkg/redislock/` — Redis distributed lock (mutex, lock, redis)
- `internal/mocks/cost_repository.go`
- `internal/mocks/opencost_client.go`
- `cmd/cron/main.go`, `server.go` — standalone cron job server
- `Dockerfile.cron` — cron server Docker image
- `.air.cron.toml` — hot-reload config for cron server
- `deployments/kubernetes/base/cron-deployment.yaml`
- `deployments/kubernetes/base/cron-service.yaml`
- `deployments/kubernetes/base/cron-rbac.yaml`

**Existing Files Modified:**

- `internal/pkg/config/config.go` — added FinOpsConfig, OpenCostConfig, PrometheusConfig, CronConfig
- `configs/config.development.yaml`, `configs/config.example.yaml` — added finops, cron, redis sections
- `internal/handler/http/init.go` — added costUseCase
- `cmd/http/server.go` — added CostUseCase, cost routes
- `cmd/http/main.go` — wired opencost/prometheus clients, cost repo/usecase
- `docker-compose.yml` — added cron service, Redis (master/slave/sentinel)
- `Makefile` — added dev-redis-*, dev-cron-*, docker-build-cron targets
- `deployments/kubernetes/base/configmap.yaml`, `secret.yaml`, `kustomization.yaml` — cron + redis config
- `deployments/kubernetes/overlays/production/kustomization.yaml` — cron production overrides

**API Endpoints Added:**

| Method | Endpoint         | Description                         |
| ------ | ---------------- | ----------------------------------- |
| GET    | `/v1/costs`      | List cost records with filtering    |
| GET    | `/v1/costs/team` | Get team cost records by time range |

**Key Features:**

- OpenCost Allocation API client (open source, no API key required) with configurable base URL
- Standalone **idp-core-cron** server using `robfig/cron` for scheduled job execution
- Redis Sentinel distributed locking (`internal/pkg/redislock`) ensuring exactly-once job execution
- Auto-extending lock TTL for long-running sync jobs (120s lock, 1s check interval)
- HTTP server on port 8983 for manual job triggering via URL paths
- Graceful shutdown with configurable timeout
- Namespace-to-team mapping via OpenCost allocation labels
- Prometheus client stub for future rightsizing queries
- Cost records stored with NUMERIC(12,4) precision for all cost columns
- JSONB raw\_data column preserving original OpenCost response
- Indexes on `(team_id, period_start)` and `(namespace, period_start)` for query performance
- Docker Compose with Redis master/slave/sentinel for local development
- Kubernetes manifests (Deployment, Service, ServiceAccount) for production deployment
- Hot-reload via Air for both API and cron servers

**Next Steps:**

- Integration tests: cost API

***

### M2 Week 5: Budget Management & Alerts (May 2026)

**Files Created:**

- `migrations/20260514000000_create_budgets_table.sql`
- `migrations/20260514000001_create_budget_alerts_table.sql`
- `internal/model/budget/type.go` (Budget, BudgetAlert, request/response types, converters, helpers)
- `internal/repository/budget/init.go` (interface + implementation: Create, GetByID, ListByTeam, ListActive, Update, Delete, CreateAlert, GetAlertsByBudget, GetLatestAlertForThreshold)
- `internal/usecase/budget/init.go`, `budget.go`, `budget_test.go`
- `internal/handler/http/budget.go`
- `internal/handler/cron/budget.go`
- `internal/pkg/slack/client.go` (Slack webhook client wrapping slack-go/slack)
- `internal/mocks/budget_repository.go`
- `internal/mocks/slack_notifier.go`

**Existing Files Modified:**

- `internal/pkg/config/config.go` — added SlackConfig
- `configs/config.development.yaml`, `configs/config.example.yaml` — added slack block + budget-alert-check schedule
- `internal/handler/http/init.go` — added budgetUseCase
- `internal/handler/cron/init.go` — added budgetUseCase
- `cmd/http/server.go` — added BudgetUseCase, budget routes
- `cmd/http/main.go` — wired budget repo, usecase, Slack client, AutoMigrate
- `cmd/cron/server.go` — added BudgetUseCase, registered budget-alert-check cron job
- `cmd/cron/main.go` — wired budget repo, usecase, Slack client

**API Endpoints Added:**

| Method | Endpoint                 | Description                  |
| ------ | ------------------------ | ---------------------------- |
| GET    | `/v1/budgets`            | List budgets (team-scoped)   |
| POST   | `/v1/budgets`            | Create a new budget          |
| GET    | `/v1/budgets/:id`        | Get budget details           |
| PATCH  | `/v1/budgets/:id`        | Update budget                |
| DELETE | `/v1/budgets/:id`        | Delete budget                |
| GET    | `/v1/budgets/:id/alerts` | Get alert history for budget |

**Key Features:**

- Budget CRUD with validation (name required, limit > 0, valid period: daily/weekly/monthly)
- Alert thresholds stored as comma-separated TEXT (e.g., `"80,90,100"`), channels as JSON TEXT (e.g., `'["slack"]'`)
- Default thresholds \[80, 90, 100] and default channels \["slack"] when not specified
- Partial updates via pointer fields in UpdateBudgetRequest
- Slack webhook notifications via `github.com/slack-go/slack` with red attachment format
- `SlackNotifier` interface in usecase layer for testability
- Cron job `budget-alert-check` runs every 15 minutes (`*/15 * * * *`) with Redis distributed locking
- Alert deduplication: (budgetID, threshold, periodStart) prevents duplicate alerts in same period
- Period window calculation in UTC: daily (start of day), weekly (Monday 00:00), monthly (1st 00:00)
- Current spend calculated by summing `TotalCost` from cost records filtered by team/environment/period
- Slacks alerts only fire when percentage crosses a threshold for the first time in the period
- Failed Slack sends are logged and alert recorded with `failed` status (does not block processing)

**Next Steps:**

- Integration tests: budget API, alert triggering

***

### M3 Week 6: Rightsizing Recommendations (May 2026)

**Files Created:**

- `migrations/20260515000000_create_rightsizing_recommendations_table.sql`
- `internal/model/rightsizing/type.go` (RightsizingRecommendation, PreviousResourceState, request/response types, converters, helpers)
- `internal/repository/rightsizing/init.go` (interface + implementation: Create, GetByID, List, Update, Delete, DeletePendingByWorkload, ListPendingByWorkload, ExistsPendingForContainer)
- `internal/repository/monitoring/init.go`, `prometheus.go` (Prometheus query wrapper)
- `internal/repository/provisioner/rightsizing.go` (GetDeployment, GetStatefulSet, UpdateDeploymentResources, UpdateStatefulSetResources)
- `internal/usecase/rightsizing/init.go`, `rightsizing.go`, `helpers.go` (GenerateRecommendations, ApplyRecommendation, RollbackRecommendation, DismissRecommendation)
- `internal/handler/http/rightsizing.go`
- `internal/handler/cron/rightsizing.go`

**Existing Files Modified:**

- `internal/pkg/config/config.go` — added RightsizingConfig with thresholds and safety buffers
- `configs/config.development.yaml`, `configs/config.example.yaml` — added rightsizing block + rightsizing-generate schedule
- `internal/pkg/prometheus/client.go` — implemented Query, QueryRange, QuerySingle methods
- `internal/repository/provisioner/init.go` — added K8s update interface methods
- `internal/mocks/provisioner_repository.go` — added mock methods for rightsizing
- `internal/handler/http/init.go` — added rightsizingUseCase
- `internal/handler/cron/init.go` — added rightsizingUseCase
- `cmd/http/server.go` — added RightsizingUseCase, rightsizing routes
- `cmd/http/main.go` — wired rightsizing repo, AutoMigrate
- `cmd/cron/server.go` — registered rightsizing-generate cron job
- `cmd/cron/main.go` — wired rightsizing repo

**API Endpoints Added:**

| Method | Endpoint                                      | Description                        |
| ------ | --------------------------------------------- | ---------------------------------- |
| GET    | `/v1/rightsizing/recommendations`             | List recommendations (filterable)  |
| GET    | `/v1/rightsizing/recommendations/:id`         | Get recommendation details         |
| POST   | `/v1/rightsizing/recommendations/:id/apply`   | Apply recommendation to workload   |
| POST   | `/v1/rightsizing/recommendations/:id/rollback`| Rollback to previous resources     |
| POST   | `/v1/rightsizing/recommendations/:id/dismiss` | Dismiss recommendation             |

**Key Features:**

- Recommendation algorithm: if utilization < 50% of request → scale_down, if > 90% → scale_up
- Prometheus queries for CPU/memory usage: `avg_over_time` and `max_over_time` over configurable lookback period (default 7 days)
- Confidence score (0-100) based on data availability and variance
- Safety buffers: 1.2x for CPU, 1.3x for memory to prevent under-provisioning
- Previous state stored as JSONB for rollback capability
- Manual apply only (no auto-apply) for production safety
- Cron job `rightsizing-generate` runs daily at 6 AM (`0 6 * * *`) with Redis distributed locking
- Supports Deployments and StatefulSets
- Recommendation types: scale_down, scale_up, optimal
- Status flow: pending → applied → (rollback) → pending
- Kubernetes resource updates via client-go with proper resource.Quantity parsing

**Next Steps:**

- Unit tests: usage analyzer, recommendation generator
- Integration tests: rightsizing API
- E2E test: apply recommendation

***

### M3 Week 7: Resource Quotas (May 2026)

**Files Created:**

- `migrations/20260516000000_create_resource_quotas_table.sql`
- `internal/model/resourcequota/type.go` (ResourceQuota, request/response types, converters, helpers)
- `internal/repository/quota/init.go` (interface + implementation: Create, GetByID, GetByNamespace, List, Update, Delete, UpdateUsage, GetActiveByNamespace, ExistsForNamespace)
- `internal/usecase/quota/quota.go` (CreateQuota, GetQuota, ListQuotas, UpdateQuota, DeleteQuota, GetUsage, RefreshUsage, CheckQuota, IsQuotaExceeded)
- `internal/handler/http/quota.go`

**Existing Files Modified:**

- `internal/pkg/webhook/validator.go` — added QuotaRule for admission webhook quota enforcement
- `internal/handler/http/init.go` — added quotaUseCase
- `cmd/http/server.go` — added QuotaUseCase, quota routes
- `cmd/http/main.go` — wired quota repo, usecase, AutoMigrate, webhook with quota

**API Endpoints Added:**

| Method | Endpoint                                        | Description                         |
| ------ | ----------------------------------------------- | ----------------------------------- |
| GET    | `/v1/quotas`                                    | List resource quotas (filterable)   |
| POST   | `/v1/quotas`                                    | Create a new resource quota         |
| GET    | `/v1/quotas/:id`                                | Get quota details                   |
| PATCH  | `/v1/quotas/:id`                                | Update quota                        |
| DELETE | `/v1/quotas/:id`                                | Delete quota                        |
| GET    | `/v1/quotas/namespace/:namespace`               | Get quota by namespace              |
| GET    | `/v1/quotas/namespace/:namespace/usage`         | Get namespace resource usage        |
| POST   | `/v1/quotas/namespace/:namespace/usage/refresh` | Refresh cached usage                |
| POST   | `/v1/quotas/check`                              | Check if request exceeds quota      |

**Key Features:**

- Resource quota model with limits for CPU request/limit, memory request/limit, storage, pod count, configmap count, secret count, PVC count
- Usage calculation from Kubernetes pods via provisioner repository
- Quota enforcement via admission webhook: pods that would exceed quota are rejected
- Current usage tracking with CPU/memory/pod count fields
- Grace period support for soft enforcement
- Fail-open design: if quota check fails, allow the pod through
- QuotaRule added to webhook validator pipeline
- Namespace-scoped quotas with team/environment association
- Utilization calculation helpers for percentage display

**Next Steps:**

- Unit tests: quota repository, quota enforcement
- Integration tests: quota API
- E2E test: quota enforcement

***

### M4 Week 8: Service Registration & Discovery (May 2026)

**Files Created:**

- `migrations/20260519000000_create_services_table.sql`
- `migrations/20260519000001_create_service_versions_table.sql`
- `migrations/20260519000002_create_service_endpoints_table.sql`
- `internal/model/service/type.go` (Service, request/response types, converters, helpers)
- `internal/model/service_version/type.go` (ServiceVersion, request/response types, helpers)
- `internal/model/service_endpoint/type.go` (ServiceEndpoint, request/response types, helpers)
- `internal/repository/service/init.go` (interface + implementation)
- `internal/repository/service/service.go` (Service CRUD methods)
- `internal/repository/service/version.go` (Version & endpoint methods)
- `internal/usecase/service/init.go` (interface + Dependencies)
- `internal/usecase/service/service.go` (Register, Discover, version/endpoint management)
- `internal/handler/http/service.go` (HTTP handlers with Swagger annotations)

**Existing Files Modified:**

- `internal/handler/http/init.go` — added serviceUseCase
- `cmd/http/server.go` — added ServiceUseCase, service routes
- `cmd/http/main.go` — wired service repo, usecase, AutoMigrate

**API Endpoints Added:**

| Method | Endpoint                                        | Description                         |
| ------ | ----------------------------------------------- | ----------------------------------- |
| GET    | `/v1/services`                                  | List services (filterable)          |
| POST   | `/v1/services`                                  | Register a new service              |
| GET    | `/v1/services/discover`                         | Discover services by query          |
| GET    | `/v1/services/:id`                              | Get service details                 |
| PATCH  | `/v1/services/:id`                              | Update service                      |
| DELETE | `/v1/services/:id`                              | Deregister service                  |
| GET    | `/v1/services/:id/versions`                     | List service versions               |
| POST   | `/v1/services/:id/versions`                     | Create service version              |
| GET    | `/v1/services/:id/versions/:versionId`          | Get version details                 |
| PATCH  | `/v1/services/:id/versions/:versionId`          | Update version                      |
| GET    | `/v1/services/:id/versions/:versionId/endpoints`| List endpoints                      |
| POST   | `/v1/services/:id/versions/:versionId/endpoints`| Add endpoint                        |
| PATCH  | `/v1/services/:id/versions/:versionId/endpoints/:endpointId` | Update endpoint       |
| DELETE | `/v1/services/:id/versions/:versionId/endpoints/:endpointId` | Remove endpoint      |

**Key Features:**

- Service model with visibility (public/team/private) for access control
- ServiceVersion tracks git_ref for deployment traceability
- ServiceEndpoint supports HTTP and gRPC protocols
- Unique constraint on (service_id, version) prevents duplicate versions
- Soft delete for services (recoverable), hard delete for endpoints (ephemeral)
- Discover endpoint searches across service names and descriptions
- DiscoverByType filters endpoints by protocol (http/grpc)

**Next Steps:**

- Unit tests: service repository, service usecase
- Integration tests: service API

***

### M4 Week 9: Dependencies & Environments (May 2026)

**Files Created:**

- `migrations/20260519000003_create_service_dependencies_table.sql`
- `migrations/20260519000004_create_service_environments_table.sql`
- `internal/model/service_dependency/type.go` (ServiceDependency, DTOs, helpers, graph types)
- `internal/model/service_environment/type.go` (ServiceEnvironment, DTOs, helpers, deployment types)
- `internal/repository/service/dependency.go` (dependency repository methods)
- `internal/repository/service/environment.go` (deployment repository methods)

**Existing Files Modified:**

- `internal/repository/service/init.go` — added dependency and environment methods to interface
- `internal/usecase/service/init.go` — added methods + EnvironmentRepo dependency
- `internal/usecase/service/service.go` — implemented dependency & deployment logic with circular dependency detection
- `internal/handler/http/service.go` — added dependency & environment handlers
- `cmd/http/server.go` — added dependency, deployment, and environment services routes
- `cmd/http/main.go` — wired EnvironmentRepo, added AutoMigrate for new models

**API Endpoints Added:**

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| GET | `/v1/services/:id/dependencies` | List service dependencies |
| POST | `/v1/services/:id/dependencies` | Add dependency |
| GET | `/v1/services/:id/dependencies/graph` | Get dependency graph |
| GET | `/v1/services/:id/dependencies/:depId` | Get dependency details |
| PATCH | `/v1/services/:id/dependencies/:depId` | Update dependency |
| DELETE | `/v1/services/:id/dependencies/:depId` | Remove dependency |
| GET | `/v1/services/:id/dependents` | List services that depend on this service |
| GET | `/v1/services/:id/environments` | List deployments for service |
| POST | `/v1/services/:id/versions/:versionId/deploy` | Deploy version to environment |
| GET | `/v1/services/:id/versions/:versionId/deployments` | List deployments for version |
| PATCH | `/v1/services/:id/versions/:versionId/deployments/:deploymentId` | Update deployment status |
| GET | `/v1/environments/:id/services` | List services deployed to environment |

**Key Features:**

- Dependency types: runtime, build, data, api
- Circular dependency detection using BFS algorithm
- Dependency graph visualization endpoint returning nodes and edges
- Deployment tracking with status (deployed, deploying, failed, rolled_back)
- Automatic roll-back of existing deployment when deploying new version
- Deployment metadata stored as JSONB for flexibility
- Environment existence validation before deployment

**Next Steps:**

- Unit tests: dependency logic
- Integration tests: dependency API, deployment API

***

### M5 Week 10: Testing & Polish (May 2026)

**Files Created:**

- `internal/mocks/rightsizing_repository.go` — Mock for rightsizing repository
- `internal/mocks/quota_repository.go` — Mock for quota repository
- `internal/mocks/monitoring_repository.go` — Mock for Prometheus monitoring repository
- `internal/mocks/service_repository.go` — Mock for service catalog repository
- `internal/usecase/rightsizing/rightsizing_test.go` — 18 test cases
- `internal/usecase/quota/quota_test.go` — 16 test cases
- `internal/usecase/service/service_test.go` — 14 test cases

**Existing Files Modified:**

- `tests/integration/service_test.go` — Added 19 integration test cases for service catalog

**Test Coverage Added:**

| Package | Tests | Coverage |
| ------- | ----- | -------- |
| `rightsizing` | 18 | List/Get recommendations, Apply/Rollback/Dismiss, Generate, Helper functions, Confidence calculation |
| `quota` | 16 | Create/Get/List/Update/Delete quota, CheckQuota, IsQuotaExceeded, GetUsage/RefreshUsage |
| `service` | 14 | Register/Get service, AddDependency, Circular dependency detection, GetDependencyGraph, DeployToEnvironment, ListDependents, RemoveDependency |

**Key Test Scenarios:**

- Circular dependency detection (direct and indirect cycles)
- Self-dependency prevention
- Kubernetes update failure handling for apply/rollback
- Confidence score calculation with missing data
- Quota enforcement logic (allow/reject based on limits)
- Pod count and resource limit checking
- Previous state storage for rollback capability

**Next Steps:**

- Integration tests: budget API, rightsizing API, quota API
- Documentation updates
- Security review
- Test coverage report

***

## 📎 References

- [PRD Phase 2](./prd/PRD_PHASE_2.md)
- [PRD Overview](./prd/PRD.md)
- [Development Guidelines](./DEV_GUIDELINE.md)
- [Test Documentation](./TEST.md)

